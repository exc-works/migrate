package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/exc-works/migrate/internal/batch"
	"github.com/exc-works/migrate/internal/config"
	"github.com/exc-works/migrate/internal/migrate"
	"go.yaml.in/yaml/v3"
)

func writeBatchConfiguration(t *testing.T, dir string, sources map[string]config.DataSourceConfig) {
	t.Helper()
	data, err := yaml.Marshal(config.FileConfig{Dialect: "sqlite", DataSources: sources})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, config.DefaultYAMLFilename), data, 0600); err != nil {
		t.Fatal(err)
	}
}

func runBatchTestCommand(ctx context.Context, dir string, args ...string) (string, string, error) {
	var out, stderr bytes.Buffer
	command := newRootCommand(&out, &stderr)
	command.SetArgs(append([]string{"--working-dir", dir, "--non-interactive"}, args...))
	err := command.ExecuteContext(ctx)
	return out.String(), stderr.String(), err
}

func TestMultiDatabaseSQLiteWorkflow(t *testing.T) {
	dir := t.TempDir()
	migrations := filepath.Join(dir, "migrations")
	if err := os.Mkdir(migrations, 0700); err != nil {
		t.Fatal(err)
	}
	writeMigration := func(name, text string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(migrations, name), []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	writeMigration("V1__create.sql", "-- +migrate Up\nCREATE TABLE items(value TEXT);\n-- +migrate Down\nDROP TABLE items;\n")
	sources := map[string]config.DataSourceConfig{
		"a": {DataSourceName: filepath.Join(dir, "a.db")},
		"b": {DataSourceName: filepath.Join(dir, "b.db")},
	}
	writeBatchConfiguration(t, dir, sources)
	run := func(args ...string) (string, string) {
		t.Helper()
		out, stderr, err := runBatchTestCommand(context.Background(), dir, args...)
		if err != nil {
			t.Fatalf("command %v: %v; %s", args, err, out)
		}
		return out, stderr
	}
	run("create", "--parallel", "2")
	// Give a a head start, then use the same directory for both targets.
	single, _ := json.Marshal(config.FileConfig{Dialect: "sqlite", DataSourceName: sources["a"].DataSourceName})
	if err := os.WriteFile(filepath.Join(dir, "single.json"), single, 0600); err != nil {
		t.Fatal(err)
	}
	run("up", "--config", "single.json")
	writeMigration("V2__seed.ts", `export const config={transaction:false};export function up(db){db.exec('INSERT INTO items VALUES (?)',['seed'])}export function down(db){db.exec('DELETE FROM items')}`)
	writeMigration("V3__change.sql", "-- +migrate Up\nUPDATE items SET value='updated';\n-- +migrate Down\nUPDATE items SET value='seed';\n")
	run("up", "--dry-run", "--parallel", "2")
	readStatus := func() []batchOutputItem {
		t.Helper()
		out, stderr := run("status", "--output", "json", "--parallel", "2")
		if stderr != "" {
			t.Fatalf("metadata status emitted logs: %s", stderr)
		}
		var results []batchOutputItem
		if err := json.Unmarshal([]byte(out), &results); err != nil || len(results) != 2 {
			t.Fatalf("invalid multi JSON: %v %s", err, out)
		}
		return results
	}
	before := readStatus()
	if (*before[0].Migrations)[0].Status != migrate.StatusApplied || (*before[1].Migrations)[0].Status != migrate.StatusPending {
		t.Fatal("independent pending plans or dry-run history changed")
	}
	out, stderr := run("up", "--parallel", "2")
	if strings.Contains(out, "WARN") || !strings.Contains(stderr, `[data_source="a"]`) || !strings.Contains(stderr, `[data_source="b"]`) {
		t.Fatalf("warnings were not isolated by target: %q %q", out, stderr)
	}
	run("up", "--parallel", "2")
	for _, item := range readStatus() {
		if item.Status != batch.StatusSucceeded || item.Migrations == nil || len(*item.Migrations) != 3 {
			t.Fatalf("incomplete migration status: %+v", item)
		}
		for _, migration := range *item.Migrations {
			if migration.Status != migrate.StatusApplied {
				t.Fatalf("not applied: %+v", migration)
			}
		}
	}
	for _, source := range sources {
		db, err := sql.Open("sqlite", source.DataSourceName)
		if err != nil {
			t.Fatal(err)
		}
		var count int
		var value string
		err = db.QueryRow("SELECT count(*), max(value) FROM items").Scan(&count, &value)
		_ = db.Close()
		if err != nil || count != 1 || value != "updated" {
			t.Fatalf("idempotent mixed migration: %d %s %v", count, value, err)
		}
	}
	run("down", "1", "--parallel", "2")
	run("down", "--all", "--parallel", "2")
	run("baseline", "--parallel", "2")
	for _, item := range readStatus() {
		for _, migration := range *item.Migrations {
			if migration.Status != migrate.StatusBaseline {
				t.Fatalf("baseline not recorded: %+v", migration)
			}
		}
	}
	run("new", "version", "one_file", "--version", "4")
	entries, err := os.ReadDir(migrations)
	if err != nil || len(entries) != 4 {
		t.Fatalf("new version was not generated exactly once: %v", err)
	}
}

func TestMultiDatabaseFailuresCleanUpAndRemainSanitized(t *testing.T) {
	for _, failure := range []string{"ping", "service", "panic", "close"} {
		t.Run(failure, func(t *testing.T) {
			dir := t.TempDir()
			writeBatchConfiguration(t, dir, map[string]config.DataSourceConfig{"a": {DataSourceName: "a"}, "b": {DataSourceName: "b"}})
			badDB, bad, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			goodDB, good, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			originalOpen, originalPing, originalNew := openDB, pingDB, newMigrateService
			t.Cleanup(func() { openDB = originalOpen; pingDB = originalPing; newMigrateService = originalNew })
			openDB = func(_, dsn string) (*sql.DB, error) {
				if dsn == "a" {
					return badDB, nil
				}
				return goodDB, nil
			}
			privateError := errors.New("private-example-value")
			pingDB = func(_ context.Context, db *sql.DB) error {
				if db == badDB && failure == "ping" {
					return privateError
				}
				return nil
			}
			newMigrateService = func(ctx context.Context, cfg migrate.Config) (*migrate.Service, error) {
				if cfg.DB == badDB {
					if failure == "service" {
						return nil, privateError
					}
					if failure == "panic" {
						panic(privateError)
					}
				}
				return migrate.NewService(ctx, cfg)
			}
			if failure == "close" {
				bad.ExpectExec("CREATE TABLE").WillReturnResult(sqlmock.NewResult(0, 0))
				bad.ExpectClose().WillReturnError(privateError)
			} else {
				bad.ExpectClose()
			}
			good.ExpectExec("CREATE TABLE").WillReturnResult(sqlmock.NewResult(0, 0))
			good.ExpectClose()
			out, stderr, err := runBatchTestCommand(context.Background(), dir, "create", "--parallel", "2")
			if err == nil || !strings.Contains(out, "failed") || !strings.Contains(out, "succeeded") {
				t.Fatalf("failure did not preserve other target: %v %s", err, out)
			}
			if strings.Contains(out+stderr+err.Error(), "private-example-value") {
				t.Fatal("batch exposed raw driver/panic/cleanup error")
			}
			for _, mock := range []sqlmock.Sqlmock{bad, good} {
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func TestMultiCancellationCleansUpBeforeReturning(t *testing.T) {
	dir := t.TempDir()
	writeBatchConfiguration(t, dir, map[string]config.DataSourceConfig{"a": {DataSourceName: "a"}, "b": {DataSourceName: "b"}})
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	originalOpen, originalPing := openDB, pingDB
	t.Cleanup(func() { openDB = originalOpen; pingDB = originalPing })
	openDB = func(string, string) (*sql.DB, error) { return db, nil }
	started := make(chan struct{})
	pingDB = func(ctx context.Context, _ *sql.DB) error {
		close(started)
		<-ctx.Done()
		return ctx.Err()
	}
	mock.ExpectClose()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		_, _, err := runBatchTestCommand(ctx, dir, "create", "--parallel", "1")
		done <- err
	}()
	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("first target did not start")
	}
	cancel()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("canceled command succeeded")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("canceled command did not stop")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestBatchRejectsInvalidOptionsBeforeConnecting(t *testing.T) {
	dir := t.TempDir()
	writeBatchConfiguration(t, dir, map[string]config.DataSourceConfig{"a": {DataSourceName: "a"}})
	original := openDB
	t.Cleanup(func() { openDB = original })
	openDB = func(string, string) (*sql.DB, error) { t.Fatal("invalid command opened DB"); return nil, nil }
	for _, args := range [][]string{{"up", "--parallel", "0"}, {"status", "--output", "unknown"}} {
		if _, _, err := runBatchTestCommand(context.Background(), dir, args...); err == nil {
			t.Fatalf("accepted invalid options: %v", args)
		}
	}
}

func TestCLIParallelizesConnectionInitialization(t *testing.T) {
	dir := t.TempDir()
	writeBatchConfiguration(t, dir, map[string]config.DataSourceConfig{"a": {DataSourceName: "a"}, "b": {DataSourceName: "b"}})
	databases := make(map[string]*sql.DB)
	var mocks []sqlmock.Sqlmock
	for _, name := range []string{"a", "b"} {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		databases[name] = db
		mock.ExpectExec("CREATE TABLE").WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectClose()
		mocks = append(mocks, mock)
	}
	originalOpen, originalPing := openDB, pingDB
	t.Cleanup(func() { openDB = originalOpen; pingDB = originalPing })
	openDB = func(_, dsn string) (*sql.DB, error) { return databases[dsn], nil }
	started := make(chan struct{}, 2)
	release := make(chan struct{})
	pingDB = func(ctx context.Context, _ *sql.DB) error {
		started <- struct{}{}
		select {
		case <-release:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() {
		_, _, err := runBatchTestCommand(ctx, dir, "create", "--parallel", "2")
		done <- err
	}()
	for range 2 {
		select {
		case <-started:
		case <-ctx.Done():
			t.Fatal("CLI did not initialize databases concurrently")
		}
	}
	close(release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	for _, mock := range mocks {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestNewVersionUsesSharedSettingsWithoutDatabaseCredentials(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("MIGRATE_NEW_VERSION_EXAMPLE_DSN", "")
	if err := os.Unsetenv("MIGRATE_NEW_VERSION_EXAMPLE_DSN"); err != nil {
		t.Fatal(err)
	}
	writeBatchConfiguration(t, dir, map[string]config.DataSourceConfig{"a": {DataSourceName: "${MIGRATE_NEW_VERSION_EXAMPLE_DSN}"}, "b": {DataSourceName: "${MIGRATE_NEW_VERSION_EXAMPLE_DSN}"}})
	if _, _, err := runBatchTestCommand(context.Background(), dir, "new", "version", "example", "--version", "1"); err != nil {
		t.Fatalf("file generation required database credentials: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "migrations", "V1__example.sql")); err != nil {
		t.Fatal(err)
	}
}

func TestPromptQueueHonorsCancellation(t *testing.T) {
	gate := make(chan struct{}, 1)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for range 100 {
		if err := acquirePrompt(ctx, gate); !errors.Is(err, context.Canceled) || len(gate) != 0 {
			t.Fatalf("canceled prompt acquired or leaked gate: %v", err)
		}
	}
	gate <- struct{}{}
	if err := acquirePrompt(ctx, gate); !errors.Is(err, context.Canceled) || len(gate) != 1 {
		t.Fatal("canceled waiter changed another prompt's ownership")
	}
}
