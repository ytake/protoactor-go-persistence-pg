package persistencepg

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/oklog/ulid/v2"
	"github.com/ytake/protoactor-go-persistence-pg/testdata"
)

func pgxConfig() *pgxpool.Config {
	conf, _ := pgxpool.ParseConfig("postgres://postgres:postgres@localhost:5432/sample?sslmode=disable&pool_max_conns=10")
	return conf
}

func TestProvider_PersistEvent(t *testing.T) {
	ctx := context.Background()
	conn, err := pgxpool.NewWithConfig(ctx, pgxConfig())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	t.Cleanup(func() {
		conn.Exec(ctx, "TRUNCATE journals")
		conn.Close()
	})
	provider, _ := New(ctx, 3, NewTable(), conn, nil)
	evt := &testdata.UserCreated{
		UserID:   ulid.Make().String(),
		UserName: "test",
		Email:    "",
	}
	provider.PersistEvent("user", 1, evt)
	var evv *testdata.UserCreated
	provider.GetEvents("user", 1, 4, func(e interface{}) {
		ev, ok := e.(*testdata.UserCreated)
		if !ok {
			t.Error("unexpected type")
		}
		evv = ev
	})
	if !reflect.DeepEqual(evt, evv) {
		t.Errorf("unexpected event %v", evv)
	}
}

func TestProvider_PersistSnapshot(t *testing.T) {
	ctx := context.Background()
	conn, err := pgxpool.NewWithConfig(ctx, pgxConfig())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	t.Cleanup(func() {
		conn.Exec(ctx, "TRUNCATE snapshots")
		conn.Close()
	})
	provider, _ := New(ctx, 3, NewTable(), conn, nil)
	evt := &testdata.UserCreated{
		UserID:   ulid.Make().String(),
		UserName: "test",
		Email:    "",
	}
	provider.PersistSnapshot("user", 1, evt)
	snapshot, idx, ok := provider.GetSnapshot("user")
	if !ok {
		t.Error("snapshot not found")
	}
	if idx != 1 {
		t.Errorf("unexpected index %d", idx)
	}
	if !reflect.DeepEqual(snapshot, evt) {
		t.Errorf("unexpected snapshot %v", snapshot)
	}
}
