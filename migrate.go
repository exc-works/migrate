// Package migrate is the public, library-facing API.
//
// It re-exports the subset of types and constructors that external services
// need to embed migration execution into their own processes. The underlying
// implementation lives under internal/ and is not part of the public API.
package migrate

import (
	"context"

	internalmigrate "github.com/exc-works/migrate/internal/migrate"
)

// Service executes migrations and inspects schema history against a database.
type Service = internalmigrate.Service

// Config holds the inputs required to construct a Service.
type Config = internalmigrate.Config

// Status is the lifecycle state of a single migration file.
type Status = internalmigrate.Status

// Migration is a parsed migration: filename, version, hash, and split SQL statements.
type Migration = internalmigrate.Migration

// SchemaRecord is a row in the migration history table.
type SchemaRecord = internalmigrate.SchemaRecord

// MigrationStatus pairs a parsed Migration with its recorded SchemaRecord (if any).
type MigrationStatus = internalmigrate.MigrationStatus

// Status values reported by Service.Status and persisted in the history table.
const (
	StatusApplied          = internalmigrate.StatusApplied
	StatusPending          = internalmigrate.StatusPending
	StatusBaseline         = internalmigrate.StatusBaseline
	StatusOutOfOrder       = internalmigrate.StatusOutOfOrder
	StatusHashMismatch     = internalmigrate.StatusHashMismatch
	StatusFilenameMismatch = internalmigrate.StatusFilenameMismatch
)

// NewService validates the Config and returns a ready-to-use Service.
func NewService(ctx context.Context, cfg Config) (*Service, error) {
	return internalmigrate.NewService(ctx, cfg)
}

// SplitFilename extracts the version portion from a "V<version>__<desc>.sql" filename.
func SplitFilename(filename string) (string, error) {
	return internalmigrate.SplitFilename(filename)
}

// CompareVersion reports whether version vi is ordered before vj (numeric when possible, lexicographic otherwise).
func CompareVersion(vi, vj string) bool {
	return internalmigrate.CompareVersion(vi, vj)
}
