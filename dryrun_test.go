package migrate_test

import (
	"bytes"
	"context"
	"database/sql"
	"strings"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/exc-works/migrate"
)

// TestDryRunWritesSQLOnly pins the DryRun contract: user migration SQL is
// written to DryRunOutput and no user table is created in the database.
// Asserts substring presence rather than exact output to avoid brittleness
// against dry-run formatting changes.
func TestDryRunWritesSQLOnly(t *testing.T) {
	ctx := context.Background()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	src := migrate.StringSource{Migrations: []migrate.SourceFile{{
		Filename: "V1__init.sql",
		Source: "-- +migrate Up\n" +
			"CREATE TABLE t (id INTEGER);\n" +
			"-- +migrate Down\n" +
			"DROP TABLE t;\n",
	}}}

	var buf bytes.Buffer
	svc, err := migrate.NewService(ctx, migrate.Config{
		Dialect:         migrate.NewSQLiteDialect(),
		DB:              db,
		MigrationSource: src,
		DryRun:          true,
		DryRunOutput:    &buf,
	})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	if err := svc.Create(); err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := svc.Up(); err != nil {
		t.Fatalf("up: %v", err)
	}

	if !strings.Contains(buf.String(), "CREATE TABLE t") {
		t.Fatalf("dry-run output missing expected SQL; got %q", buf.String())
	}

	var count int
	if err := db.QueryRow("SELECT count(*) FROM sqlite_master WHERE type='table' AND name='t'").Scan(&count); err != nil {
		t.Fatalf("inspect sqlite_master: %v", err)
	}
	if count != 0 {
		t.Fatalf("dry-run created real table; expected 0, got %d", count)
	}

	statuses, err := svc.Status()
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if len(statuses) != 1 || statuses[0].Status != migrate.StatusPending {
		t.Fatalf("expected single pending migration after dry-run, got %+v", statuses)
	}
}
