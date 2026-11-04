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
	"sync/atomic"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/exc-works/migrate/internal/batch"
	"github.com/exc-works/migrate/internal/config"
	"github.com/exc-works/migrate/internal/migrate"
)

func TestBatchMigrationFailureRollsBackAndRetriesOnlyFailedDatabase(t *testing.T) {
	for _, format := range []string{"sql", "ts"} {
		t.Run(format, func(t *testing.T) {
			dir := t.TempDir()
			sources := map[string]config.DataSourceConfig{
				"shard_0": {DataSourceName: filepath.Join(dir, "0.db")},
				"shard_1": {DataSourceName: filepath.Join(dir, "1.db")},
			}
			writeBatchConfiguration(t, dir, sources)
			migrations := filepath.Join(dir, "migrations")
			if err := os.Mkdir(migrations, 0700); err != nil {
				t.Fatal(err)
			}
			initial := "-- +migrate Up\nCREATE TABLE items(id INT PRIMARY KEY, value TEXT);\nINSERT INTO items VALUES(1,'before');\n-- +migrate Down\nDROP TABLE items;\n"
			if err := os.WriteFile(filepath.Join(migrations, "V1__items.sql"), []byte(initial), 0600); err != nil {
				t.Fatal(err)
			}
			run := func(args ...string) string {
				t.Helper()
				out, stderr, err := runBatchTestCommand(context.Background(), dir, args...)
				if err != nil {
					t.Fatalf("%v: %v; %s %s", args, err, out, stderr)
				}
				return out
			}
			run("create")
			run("up")
			bad, err := sql.Open("sqlite", sources["shard_0"].DataSourceName)
			if err != nil {
				t.Fatal(err)
			}
			defer bad.Close()
			if _, err := bad.Exec("INSERT INTO items VALUES(2,'conflict')"); err != nil {
				t.Fatal(err)
			}
			change := "-- +migrate Up\nUPDATE items SET value='changed' WHERE id=1;\nINSERT INTO items VALUES(2,'new');\n-- +migrate Down\nDELETE FROM items WHERE id=2;\n"
			if format == "ts" {
				change = `export function up(db){db.exec("UPDATE items SET value='changed' WHERE id=1");db.exec("INSERT INTO items VALUES(2,'new')")} export function down(db){db.exec('DELETE FROM items WHERE id=2')}`
			}
			if err := os.WriteFile(filepath.Join(migrations, "V2__change."+format), []byte(change), 0600); err != nil {
				t.Fatal(err)
			}
			out, _, err := runBatchTestCommand(context.Background(), dir, "up", "--parallel", "2")
			if err == nil || !strings.Contains(strings.Join(strings.Fields(out), " "), `"shard_0" failed`) || !strings.Contains(strings.Join(strings.Fields(out), " "), `"shard_1" succeeded`) {
				t.Fatalf("migration failure was not isolated: %v %s", err, out)
			}
			var statuses []batchOutputItem
			if err := json.Unmarshal([]byte(run("status", "--output", "json")), &statuses); err != nil {
				t.Fatal(err)
			}
			if len(statuses) != 2 || statuses[0].Migrations == nil || statuses[1].Migrations == nil || len(*statuses[0].Migrations) != 2 || len(*statuses[1].Migrations) != 2 {
				t.Fatalf("invalid histories: %+v", statuses)
			}
			if (*statuses[0].Migrations)[1].Status != migrate.StatusPending || (*statuses[1].Migrations)[1].Status != migrate.StatusApplied {
				t.Fatal("failed migration recorded history or healthy migration was not applied")
			}
			var value string
			if err := bad.QueryRow("SELECT value FROM items WHERE id=1").Scan(&value); err != nil || value != "before" {
				t.Fatalf("failed migration did not roll back earlier writes: %q %v", value, err)
			}
			if _, err := bad.Exec("DELETE FROM items WHERE id=2"); err != nil {
				t.Fatal(err)
			}
			run("up", "--parallel", "2")
			// Replaying V2 on the healthy shard would fail its primary-key insert.
			for _, source := range sources {
				db, err := sql.Open("sqlite", source.DataSourceName)
				if err != nil {
					t.Fatal(err)
				}
				var count int
				err = db.QueryRow("SELECT count(*) FROM items WHERE (id=1 AND value='changed') OR (id=2 AND value='new')").Scan(&count)
				_ = db.Close()
				if err != nil || count != 2 {
					t.Fatalf("retry did not converge: count=%d err=%v", count, err)
				}
			}
		})
	}
}

func TestBatchCancellationDuringSQLExecution(t *testing.T) {
	for _, mode := range []string{"cancel", "deadline"} {
		t.Run(mode, func(t *testing.T) {
			dir := t.TempDir()
			writeBatchConfiguration(t, dir, map[string]config.DataSourceConfig{"shard_0": {DataSourceName: "0"}, "shard_1": {DataSourceName: "1"}})
			started := make(chan struct{}, 1)
			matcher := sqlmock.QueryMatcherFunc(func(_, _ string) error { started <- struct{}{}; return nil })
			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(matcher))
			if err != nil {
				t.Fatal(err)
			}
			mock.ExpectExec("CREATE TABLE").WillDelayFor(time.Hour).WillReturnResult(sqlmock.NewResult(0, 0))
			mock.ExpectClose()
			originalOpen, originalPing := openDB, pingDB
			t.Cleanup(func() { openDB = originalOpen; pingDB = originalPing })
			var opens atomic.Int32
			openDB = func(string, string) (*sql.DB, error) { opens.Add(1); return db, nil }
			pingDB = func(context.Context, *sql.DB) error { return nil }
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			if mode == "deadline" {
				cancel()
				ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
			}
			defer cancel()
			type response struct {
				out string
				err error
			}
			done := make(chan response, 1)
			go func() {
				out, _, err := runBatchTestCommand(ctx, dir, "create", "--parallel", "1")
				done <- response{out, err}
			}()
			select {
			case <-started:
			case <-time.After(5 * time.Second):
				t.Fatal("SQL execution did not start")
			}
			if mode == "cancel" {
				cancel()
			}
			select {
			case result := <-done:
				if result.err == nil || strings.Count(result.out, "canceled") != 4 || opens.Load() != 1 {
					t.Fatalf("incomplete cancellation or dispatched a queued source: %v opens=%d %s", result.err, opens.Load(), result.out)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("SQL cancellation did not finish")
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestBatchJSONMixedResultsAndGroupedDryRun(t *testing.T) {
	names := []string{"shard_0\n\t", "shard_1", "shard_2"}
	results := []batch.Result{
		{Name: names[0], Status: batch.StatusSucceeded},
		{Name: names[1], Status: batch.StatusFailed, Err: errors.New("private-example-value")},
		{Name: names[2], Status: batch.StatusCanceled, Err: context.Canceled},
	}
	payloads := map[string]*targetOutput{}
	for _, name := range names {
		payloads[name] = &targetOutput{}
	}
	var out bytes.Buffer
	if err := printBatchResults(&out, results, payloads, commandOperation{name: commandStatus}, "json"); err == nil {
		t.Fatal("mixed results succeeded")
	}
	var items []batchOutputItem
	if err := json.Unmarshal(out.Bytes(), &items); err != nil {
		t.Fatal(err)
	}
	if len(items) != 3 || items[0].DataSource != names[0] || items[0].Migrations == nil || len(*items[0].Migrations) != 0 || items[1].Migrations != nil || items[2].Migrations != nil || items[1].Status != batch.StatusFailed || items[2].Status != batch.StatusCanceled {
		t.Fatalf("invalid mixed JSON: %s", out.String())
	}
	if strings.Contains(out.String(), "private-example-value") || items[1].Error == "" || items[2].Error == "" {
		t.Fatal("unsafe or missing error summaries")
	}
	out.Reset()
	for i, name := range names {
		payloads[name].plan.WriteString("-- plan " + string(rune('0'+i)) + "\n")
	}
	if err := printBatchResults(&out, results, payloads, commandOperation{name: "up", dryRun: true}, "table"); err == nil {
		t.Fatal("failed dry run succeeded")
	}
	text := out.String()
	if strings.Contains(text, names[0]) || !strings.Contains(text, `Data source "shard_0\n\t"`+"\n-- plan 0\n") || !strings.Contains(text, "Data source \"shard_1\"\n-- plan 1\n") || !strings.Contains(text, "Data source \"shard_2\"\n-- plan 2\n") {
		t.Fatalf("ungrouped or unescaped dry run: %q", text)
	}
}

func TestPromptQueueSerializesCallbacks(t *testing.T) {
	gate := make(chan struct{}, 1)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var active atomic.Int32
	entered := make(chan int, 2)
	release := make(chan struct{})
	done := make(chan error, 2)
	for range 2 {
		go func() {
			if err := acquirePrompt(ctx, gate); err != nil {
				done <- err
				return
			}
			entered <- int(active.Add(1))
			select {
			case <-release:
			case <-ctx.Done():
			}
			active.Add(-1)
			<-gate
			done <- nil
		}()
	}
	select {
	case n := <-entered:
		if n != 1 {
			t.Fatal("overlapping prompts")
		}
	case <-ctx.Done():
		t.Fatal("prompt never acquired gate")
	}
	close(release)
	for range 2 {
		if err := <-done; err != nil {
			t.Fatal(err)
		}
	}
	if n := <-entered; n != 1 || len(gate) != 0 || active.Load() != 0 {
		t.Fatal("overlapping prompts or leaked gate")
	}
}
