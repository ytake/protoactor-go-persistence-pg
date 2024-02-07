# protoactor-go-persistence-pg

Go package with persistence provider for Proto Actor (Go) based on PostgreSQL.

using [pgx v5](https://github.com/jackc/pgx)

# Usage

```go
package main

import (
	"context"
	
	"github.com/asynkron/protoactor-go/actor"
	"github.com/asynkron/protoactor-go/persistence"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ytake/protoactor-go-persistence-pg"
)

type Actor struct {
	persistence.Mixin
}

func (a *Actor) Receive(ctx actor.Context) {
	// example
}

func main() {

	conf, _ := pgxpool.ParseConfig("postgres://postgres:postgres@localhost:5432/sample?sslmode=disable&pool_max_conns=10")

	system := actor.NewActorSystem()
	ctx := context.Background()
	conn, _ := pgxpool.NewWithConfig(ctx, conf)
	provider, _ := persistencepg.New(ctx, 3, persistencepg.NewTable(), conn, system.Logger())

	props := actor.PropsFromProducer(func() actor.Actor { return &Actor{} },
		actor.WithReceiverMiddleware(persistence.Using(provider)))

	pid, _ := system.Root.SpawnNamed(props, "persistent")
}

```

# Default table schema

use ulid as id(varchar(26)) and json as payload

```sql
CREATE TABLE journals
(
    id              VARCHAR(26) NOT NULL,
    payload         JSONB NOT NULL,
    sequence_number BIGINT,
    actor_name      VARCHAR(255),
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE (id),
    UNIQUE (actor_name, sequence_number)
);

CREATE TABLE snapshots
(
    id              VARCHAR(26) NOT NULL,
    payload         JSONB NOT NULL,
    sequence_number BIGINT,
    actor_name      VARCHAR(255),
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE (id),
    UNIQUE (actor_name, sequence_number)
);

```

## change table name

use the interface to change the table name.

for journal table and snapshot table.

```go 
// Schemaer is the interface that wraps the basic methods for a schema.
type Schemaer interface {
    // JournalTableName returns the name of the journal table.
    JournalTableName() string
    // SnapshotTableName returns the name of the snapshot table.
    SnapshotTableName() string
    // ID returns the name of the id column.
    ID() string
    // Payload returns the name of the payload column.
    Payload() string
    // ActorName returns the name of the actor name column.
    ActorName() string
    // SequenceNumber returns the name of the sequence number column.
    SequenceNumber() string
    // Created returns the name of the created at column.
    Created() string
    // CreateTable returns the sql statement to create the table.
    CreateTable() []string
}
```
