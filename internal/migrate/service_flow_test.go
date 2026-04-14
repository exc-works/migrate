package migrate

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/exc-works/sql-migrate/internal/dialect"
	"github.com/exc-works/sql-migrate/internal/logger"
	"github.com/exc-works/sql-migrate/internal/source"
)

func TestUpSkipsOutOfOrderWhenDisabled(t *testing.T) {
	mig1 := source.Migration{Filename: "V1__init.sql", Source: migrationSQL("CREATE TABLE t1 (id INT);", "DROP TABLE t1;")}
	mig3 := source.Migration{Filename: "V3__seed.sql", Source: migrationSQL("SELECT 3;", "SELECT 3;")}
	svc, mock, cleanup := newServiceWithMock(t, []source.Migration{mig1, mig3}, false, false, nil)
	defer cleanup()

	selectSQL, _ := dialect.PostgresDialect{}.SelectSchemaSQL("migration_schema")
	rows := sqlmock.NewRows([]string{"id", "version", "filename", "hash", "status", "created_at"}).
		AddRow(1, "3", "V3__seed.sql", hashString(mig3.Source), string(StatusApplied), time.Now().UTC())
	mock.ExpectQuery(regexp.QuoteMeta(selectSQL)).WillReturnRows(rows)

	if err := svc.Up(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestUpFailsOnHashMismatch(t *testing.T) {
	mig := source.Migration{Filename: "V1__init.sql", Source: migrationSQL("SELECT 1;", "SELECT 1;")}
	svc, mock, cleanup := newServiceWithMock(t, []source.Migration{mig}, false, false, nil)
	defer cleanup()

	selectSQL, _ := dialect.PostgresDialect{}.SelectSchemaSQL("migration_schema")
	rows := sqlmock.NewRows([]string{"id", "version", "filename", "hash", "status", "created_at"}).
		AddRow(1, "1", "V1__init.sql", "different_hash", string(StatusApplied), time.Now().UTC())
	mock.ExpectQuery(regexp.QuoteMeta(selectSQL)).WillReturnRows(rows)

	err := svc.Up()
	if err == nil {
		t.Fatalf("expected hash mismatch error")
	}
	if !strings.Contains(err.Error(), "hash mismatch") {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestDownToVersionRollsBackOnlyHigherVersions(t *testing.T) {
	mig1 := source.Migration{Filename: "V1__init.sql", Source: migrationSQL("SELECT 1;", "DELETE FROM accounts WHERE id = 1;")}
	mig2 := source.Migration{Filename: "V2__seed.sql", Source: migrationSQL("SELECT 2;", "DELETE FROM accounts WHERE id = 2;")}
	mig3 := source.Migration{Filename: "V3__seed_more.sql", Source: migrationSQL("SELECT 3;", "DELETE FROM accounts WHERE id = 3;")}
	svc, mock, cleanup := newServiceWithMock(t, []source.Migration{mig1, mig2, mig3}, false, false, nil)
	defer cleanup()

	selectSQL, _ := dialect.PostgresDialect{}.SelectSchemaSQL("migration_schema")
	rows := sqlmock.NewRows([]string{"id", "version", "filename", "hash", "status", "created_at"}).
		AddRow(1, "1", "V1__init.sql", hashString(mig1.Source), string(StatusApplied), time.Now().UTC()).
		AddRow(2, "2", "V2__seed.sql", hashString(mig2.Source), string(StatusApplied), time.Now().UTC()).
		AddRow(3, "3", "V3__seed_more.sql", hashString(mig3.Source), string(StatusApplied), time.Now().UTC())
	mock.ExpectQuery(regexp.QuoteMeta(selectSQL)).WillReturnRows(rows)

	deleteSQL, _ := dialect.PostgresDialect{}.DeleteSchemaSQL("migration_schema")
	mock.ExpectBegin()
	mock.ExpectExec(`(?s).*DELETE FROM accounts WHERE id = 3;.*`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(deleteSQL)).WithArgs("V3__seed_more.sql").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := svc.Down("2", false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestUpDryRunPrintsSQLWithoutExecution(t *testing.T) {
	mig := source.Migration{Filename: "V1__init.sql", Source: migrationSQL("CREATE TABLE dry_run_demo (id INT);", "DROP TABLE dry_run_demo;")}
	buf := &bytes.Buffer{}
	svc, mock, cleanup := newServiceWithMock(t, []source.Migration{mig}, true, false, buf)
	defer cleanup()

	selectSQL, _ := dialect.PostgresDialect{}.SelectSchemaSQL("migration_schema")
	emptyRows := sqlmock.NewRows([]string{"id", "version", "filename", "hash", "status", "created_at"})
	mock.ExpectQuery(regexp.QuoteMeta(selectSQL)).WillReturnRows(emptyRows)
	mock.ExpectBegin()
	mock.ExpectCommit()

	if err := svc.Up(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := buf.String(); !strings.Contains(got, "CREATE TABLE dry_run_demo") {
		t.Fatalf("dry-run output missing SQL, got: %s", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestUpClickHouseRunsWithoutTransaction(t *testing.T) {
	mig := source.Migration{Filename: "V1__init.sql", Source: migrationSQL("SELECT 1;", "SELECT 1;")}

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer db.Close()

	cfg := Config{
		Dialect:         dialect.ClickHouseDialect{},
		DB:              db,
		Logger:          logger.NoopLogger{},
		SchemaName:      "migration_schema",
		MigrationSource: source.StringSource{Migrations: []source.Migration{mig}},
	}
	svc, err := NewService(context.Background(), cfg)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	selectSQL, _ := dialect.ClickHouseDialect{}.SelectSchemaSQL("migration_schema")
	mock.ExpectQuery(regexp.QuoteMeta(selectSQL)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "version", "filename", "hash", "status", "created_at"}))

	insertSQL, _ := dialect.ClickHouseDialect{}.InsertSchemaSQL("migration_schema")
	mock.ExpectExec(regexp.QuoteMeta("SELECT 1;")).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(insertSQL)).
		WithArgs(sqlmock.AnyArg(), "1", "V1__init.sql", hashString(mig.Source), string(StatusApplied), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := svc.Up(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func newServiceWithMock(t *testing.T, migrations []source.Migration, dryRun bool, outOfOrder bool, dryOut *bytes.Buffer) (*Service, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	cfg := Config{
		Dialect:           dialect.PostgresDialect{},
		DB:                db,
		Logger:            logger.NoopLogger{},
		SchemaName:        "migration_schema",
		MigrateOutOfOrder: outOfOrder,
		MigrationSource:   source.StringSource{Migrations: migrations},
		DryRun:            dryRun,
	}
	if dryOut != nil {
		cfg.DryRunOutput = dryOut
	}
	svc, err := NewService(context.Background(), cfg)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	cleanup := func() {
		_ = db.Close()
	}
	return svc, mock, cleanup
}

func migrationSQL(up, down string) string {
	return fmt.Sprintf(`-- +migrate Up
%s

-- +migrate Down
%s
`, up, down)
}

func hashString(input string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(input)))
}
