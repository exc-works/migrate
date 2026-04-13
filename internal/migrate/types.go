package migrate

import (
	"time"
)

type Status string

const (
	StatusApplied          Status = "applied"
	StatusPending          Status = "pending"
	StatusBaseline         Status = "baseline"
	StatusOutOfOrder       Status = "outOfOrder"
	StatusHashMismatch     Status = "hashMismatch"
	StatusFilenameMismatch Status = "filenameMismatch"
)

type Migration struct {
	Filename string
	Source   string
	Version  string
	Hash     string

	UpStatements   []string
	DownStatements []string
}

type SchemaRecord struct {
	ID        uint64
	Version   string
	Filename  string
	Hash      string
	Status    Status
	CreatedAt time.Time
}

type MigrationStatus struct {
	Migration Migration
	Schema    *SchemaRecord
	Status    Status
}
