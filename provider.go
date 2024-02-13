package persistencepg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/asynkron/protoactor-go/persistence"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/oklog/ulid/v2"
	"google.golang.org/protobuf/proto"
)

// Provider is the abstraction used for persistence
type Provider struct {
	context          context.Context
	tableSchema      Schemaer
	snapshotInterval int
	db               *pgxpool.Pool
	logger           *slog.Logger
}

// New creates a new pg provider
func New(ctx context.Context, snapshotInterval int, table Schemaer, db *pgxpool.Pool, logger *slog.Logger) (*Provider, error) {
	return &Provider{
		context:          ctx,
		tableSchema:      table,
		snapshotInterval: snapshotInterval,
		db:               db,
		logger:           logger,
	}, nil
}

// DeleteEvents removes all events from the provider
func (provider *Provider) DeleteEvents(_ string, _ int) {
}

func (provider *Provider) Restart() {
	_ = provider.db.Ping(provider.context)
}

func (provider *Provider) GetSnapshotInterval() int {
	return provider.snapshotInterval
}

func (provider *Provider) selectColumns() string {
	return strings.Join([]string{
		provider.tableSchema.ID(),
		provider.tableSchema.Payload(),
		provider.tableSchema.SequenceNumber(),
		provider.tableSchema.ActorName(),
	}, ",")
}

func (provider *Provider) GetEvents(actorName string, eventIndexStart int, eventIndexEnd int, callback func(e interface{})) {
	tx, _ := provider.db.BeginTx(provider.context, pgx.TxOptions{IsoLevel: pgx.Serializable})
	defer tx.Commit(provider.context)
	rows, err := tx.Query(
		provider.context,
		fmt.Sprintf(
			"SELECT %s FROM %s WHERE %s = $1 AND %s BETWEEN $2 AND $3 ORDER BY %s ASC",
			provider.selectColumns(),
			provider.tableSchema.JournalTableName(),
			provider.tableSchema.ActorName(),
			provider.tableSchema.SequenceNumber(),
			provider.tableSchema.SequenceNumber(),
		),
		actorName, eventIndexEnd, eventIndexStart)
	if !errors.Is(err, sql.ErrNoRows) && err != nil {
		provider.logger.Error(err.Error(), slog.String("actor_name", actorName))
		return
	}
	for rows.Next() {
		env := &envelope{}
		if err := rows.Scan(&env.ID, &env.Message, &env.EventIndex, &env.ActorName); err != nil {
			return
		}
		m, err := env.message()
		if err != nil {
			provider.logger.Error(err.Error(), slog.String("actor_name", actorName))
			return
		}
		callback(m)
	}
}

// 'executeTx' is a function that manages the lifecycle of a DB transaction.
// It takes a function 'op' that contains DB transaction operation to be executed.
func (provider *Provider) executeTx(op func(tx pgx.Tx) error) (err error) {
	tx, err := provider.db.BeginTx(provider.context, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return err
	}
	defer tx.Rollback(provider.context)
	if err != nil {
		return err
	}
	// Execute operation
	err = op(tx)
	if err != nil {
		return err
	}
	// Everything went fine
	return tx.Commit(provider.context)
}

func (provider *Provider) PersistEvent(actorName string, eventIndex int, snapshot proto.Message) {
	envelope, err := newEnvelope(snapshot)
	if err != nil {
		provider.logger.Error(
			fmt.Sprintf("persistence error: %s", err), slog.String("actor_name", actorName))
		return
	}
	err = provider.executeTx(func(tx pgx.Tx) error {
		_, err := tx.Exec(
			provider.context,
			fmt.Sprintf(
				"INSERT INTO %s (%s) VALUES ($1, $2, $3, $4)",
				provider.tableSchema.JournalTableName(), provider.selectColumns()),
			ulid.Make().String(), string(envelope), eventIndex, actorName)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		provider.logger.Error(
			fmt.Sprintf("persistence event / sql error: %s", err.Error()),
			slog.String("actor_name", actorName))
		return
	}
}

func (provider *Provider) PersistSnapshot(actorName string, eventIndex int, snapshot proto.Message) {
	envelope, err := newEnvelope(snapshot)
	if err != nil {
		provider.logger.Error(
			fmt.Sprintf("persistence error: %s", err), slog.String("actor_name", actorName))
		return
	}
	err = provider.executeTx(func(tx pgx.Tx) error {
		_, err := tx.Exec(
			provider.context,
			fmt.Sprintf(
				"INSERT INTO %s (%s) VALUES ($1, $2, $3, $4)",
				provider.tableSchema.SnapshotTableName(), provider.selectColumns()),
			ulid.Make().String(), string(envelope), eventIndex, actorName)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		provider.logger.Error(
			fmt.Sprintf("persistence snapshot / sql error: %s", err.Error()),
			slog.String("actor_name", actorName))
		return
	}
}

func (provider *Provider) GetSnapshot(actorName string) (snapshot interface{}, eventIndex int, ok bool) {
	tx, _ := provider.db.BeginTx(provider.context, pgx.TxOptions{IsoLevel: pgx.Serializable})
	defer tx.Commit(provider.context)
	rows, err := tx.Query(
		provider.context,
		fmt.Sprintf(
			"SELECT %s FROM %s WHERE %s = $1 ORDER BY %s DESC",
			provider.selectColumns(),
			provider.tableSchema.SnapshotTableName(),
			provider.tableSchema.ActorName(),
			provider.tableSchema.SequenceNumber(),
		),
		actorName)
	defer rows.Close()
	if !errors.Is(err, sql.ErrNoRows) && err != nil {
		provider.logger.Error(err.Error(), slog.String("actor_name", actorName))
		return nil, 0, false
	}
	for rows.Next() {
		env := envelope{}
		if err := rows.Scan(&env.ID, &env.Message, &env.EventIndex, &env.ActorName); err != nil {
			return nil, 0, false
		}
		m, err := env.message()
		if err != nil {
			provider.logger.Error(err.Error(), slog.String("actor_name", actorName))
			return nil, 0, false
		}
		return m, env.EventIndex, true
	}
	return nil, 0, false
}

func (provider *Provider) DeleteSnapshots(_ string, _ int) {
}

func (provider *Provider) GetState() persistence.ProviderState {
	return provider
}
