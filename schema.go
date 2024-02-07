package persistencepg

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

func NewTable() *DefaultSchema {
	return &DefaultSchema{
		journalTable:  "journals",
		snapshotTable: "snapshots",
	}
}

// WithJournalTable sets the name of the journal table.
func (d *DefaultSchema) WithJournalTable(name string) *DefaultSchema {
	d.journalTable = name
	return d
}

// WithSnapshotTable sets the name of the snapshot table.
func (d *DefaultSchema) WithSnapshotTable(name string) *DefaultSchema {
	d.snapshotTable = name
	return d
}

// DefaultSchema is the default implementation of the Schemaer interface.
type DefaultSchema struct {
	journalTable  string
	snapshotTable string
}

// JournalTableName returns the name of the journal table.
func (d *DefaultSchema) JournalTableName() string {
	return d.journalTable
}

// SnapshotTableName returns the name of the snapshot table.
func (d *DefaultSchema) SnapshotTableName() string {
	return d.snapshotTable
}

// ID returns the name of the id column.
func (d *DefaultSchema) ID() string {
	return "id"
}

// Payload returns the name of the payload column.
func (d *DefaultSchema) Payload() string {
	return "payload"
}

// ActorName returns the name of the actor name column.
func (d *DefaultSchema) ActorName() string {
	return "actor_name"
}

// SequenceNumber returns the name of the sequence number column.
func (d *DefaultSchema) SequenceNumber() string {
	return "sequence_number"
}

// Created returns the name of the created at column.
func (d *DefaultSchema) Created() string {
	return "created_at"
}

// CreateTable returns the sql statement to create the table.
func (d *DefaultSchema) CreateTable() []string {
	tables := []string{
		d.JournalTableName(),
		d.SnapshotTableName(),
	}
	createTables := make([]string, 0, len(tables))
	for _, table := range tables {
		createTables = append(createTables, "CREATE TABLE "+table+" ("+
			d.ID()+" VARCHAR(26) NOT NULL,"+
			d.Payload()+" JSONB NOT NULL,"+
			d.SequenceNumber()+" BIGINT,"+
			d.ActorName()+" VARCHAR(255),"+
			d.Created()+" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,"+
			"PRIMARY KEY ("+d.ID()+"),"+
			"UNIQUE ("+d.ID()+"),"+
			"UNIQUE ("+d.ActorName()+", "+d.SequenceNumber()+");")
	}
	return createTables
}
