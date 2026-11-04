//go:build integration

package integrationtest

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/exc-works/migrate/internal/config"
	"github.com/exc-works/migrate/internal/dialect"
	"github.com/exc-works/migrate/internal/testssh"
)

func TestBatchCLIInFlightCancellationAndTransportLoss(t *testing.T) {
	if !slices.Contains(selectedFlavors(), dialect.NamePostgres) {
		t.Skip("in-flight batch workflows use PostgreSQL")
	}
	d, _ := dialect.FromName(dialect.NamePostgres)
	admin, baseDSN, cleanup := startContainerDBWithDSN(t, dialect.NamePostgres, d)
	defer cleanup()
	dir := t.TempDir()
	binary := filepath.Join(dir, "migrate")
	buildCtx, cancelBuild := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancelBuild()
	build := exec.CommandContext(buildCtx, "go", "build", "-o", binary, "./cmd/migrate")
	build.Dir = ".."
	if data, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build batch CLI: %v %s", err, data)
	}
	for _, mode := range []string{"cancel", "transport"} {
		t.Run(mode, func(t *testing.T) {
			work := t.TempDir()
			cfg := config.FileConfig{Dialect: dialect.NamePostgres, DataSources: map[string]config.DataSourceConfig{}}
			var databases []*sql.DB
			var locks []*sql.Conn
			var names []string
			var applications []string
			var servers []*testssh.Server
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			for _, name := range []string{"shard_0", "shard_1"} {
				database := randomName("batch_fault")
				if _, err := admin.ExecContext(ctx, "CREATE DATABASE "+database); err != nil {
					t.Fatal(err)
				}
				names = append(names, database)
				dsn := strings.Replace(baseDSN, "dbname=testdb", "dbname="+database, 1)
				db, err := sql.Open(d.DriverName(), dsn)
				if err != nil {
					t.Fatal(err)
				}
				databases = append(databases, db)
				t.Cleanup(func() { _ = db.Close() })
				lock, err := db.Conn(ctx)
				if err != nil {
					t.Fatal(err)
				}
				locks = append(locks, lock)
				t.Cleanup(func() { _ = lock.Close() })
				if _, err := lock.ExecContext(ctx, "SELECT pg_advisory_lock(314159)"); err != nil {
					t.Fatal(err)
				}
				remote, tunneledDSN := sshDatabaseDSN(t, dialect.NamePostgres, dsn)
				application := randomName("batch_client")
				applications = append(applications, application)
				address, err := url.Parse(tunneledDSN)
				if err != nil {
					t.Fatal(err)
				}
				query := address.Query()
				query.Set("application_name", application)
				address.RawQuery = query.Encode()
				tunneledDSN = address.String()
				server := testssh.New(t)
				server.Forward = func(string) (net.Conn, error) { return net.Dial("tcp", remote) }
				server.Serve()
				servers = append(servers, server)
				ssh := workflowSSHOptions(t, server)
				cfg.DataSources[name] = config.DataSourceConfig{DataSourceName: tunneledDSN, SSHTunnel: &ssh}
				if name == "shard_1" {
					cfg.SSHTunnel = &ssh
				}
			}
			writeConfig := func() {
				t.Helper()
				data, err := json.Marshal(cfg)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(work, "config.json"), data, 0600); err != nil {
					t.Fatal(err)
				}
			}
			writeConfig()
			command := func(args ...string) *exec.Cmd {
				return exec.CommandContext(ctx, binary, append([]string{"--working-dir", work, "--config", "config.json", "--non-interactive"}, args...)...)
			}
			if data, err := command("create").CombinedOutput(); err != nil {
				t.Fatalf("create: %v %s", err, data)
			}
			if err := os.Mkdir(filepath.Join(work, "migrations"), 0700); err != nil {
				t.Fatal(err)
			}
			migration := "-- +migrate Up\nCREATE TABLE batch_items(id INT);\nSELECT pg_advisory_xact_lock(314159);\nINSERT INTO batch_items VALUES(1);\n-- +migrate Down\nDROP TABLE batch_items;\n"
			if err := os.WriteFile(filepath.Join(work, "migrations", "V1__items.sql"), []byte(migration), 0600); err != nil {
				t.Fatal(err)
			}
			cmd := command("up", "--parallel", "2")
			var stdout, stderr bytes.Buffer
			cmd.Stdout, cmd.Stderr = &stdout, &stderr
			if err := cmd.Start(); err != nil {
				t.Fatal(err)
			}
			done := make(chan error, 1)
			go func() { done <- cmd.Wait() }()
			// Observe both transactions blocked in SQL, rather than guessing with a sleep.
			ticker := time.NewTicker(20 * time.Millisecond)
			defer ticker.Stop()
			for {
				var blocked int
				err := admin.QueryRowContext(ctx, "SELECT count(*) FROM pg_stat_activity WHERE datname IN ($1,$2) AND state='active' AND wait_event='advisory'", names[0], names[1]).Scan(&blocked)
				if err != nil {
					t.Fatal(err)
				}
				if blocked == 2 {
					break
				}
				select {
				case err := <-done:
					t.Fatalf("command exited before both migrations blocked: %v %s %s", err, stdout.String(), stderr.String())
				case <-ctx.Done():
					t.Fatal("migrations did not reach the SQL barrier")
				case <-ticker.C:
				}
			}
			if mode == "cancel" {
				if err := cmd.Process.Signal(os.Interrupt); err != nil {
					t.Fatal(err)
				}
			} else {
				servers[0].Close()
				// Only the healthy SSH transport remains. Release its SQL barrier.
				if _, err := locks[1].ExecContext(ctx, "SELECT pg_advisory_unlock(314159)"); err != nil {
					t.Fatal(err)
				}
			}
			select {
			case err := <-done:
				if err == nil {
					t.Fatal("interrupted batch succeeded")
				}
			case <-ctx.Done():
				t.Fatal("batch did not finish after interruption")
			}
			text := strings.Join(strings.Fields(stdout.String()), " ")
			// PostgreSQL may not observe a closed client socket while waiting on
			// our advisory lock. Release the test barriers before checking that
			// the disconnected backends have rolled back and exited.
			for _, lock := range locks {
				if _, err := lock.ExecContext(ctx, "SELECT pg_advisory_unlock_all()"); err != nil {
					t.Fatal(err)
				}
			}
			for {
				var connected int
				if err := admin.QueryRowContext(ctx, "SELECT count(*) FROM pg_stat_activity WHERE application_name IN ($1,$2)", applications[0], applications[1]).Scan(&connected); err != nil {
					t.Fatal(err)
				}
				if connected == 0 {
					break
				}
				select {
				case <-ctx.Done():
					t.Fatal("batch returned without releasing its database sessions")
				case <-ticker.C:
				}
			}
			if mode == "cancel" {
				if !strings.Contains(text, `"shard_0" canceled`) || !strings.Contains(text, `"shard_1" canceled`) {
					t.Fatalf("cancellation was not propagated: %s %s", stdout.String(), stderr.String())
				}
			} else if !strings.Contains(text, `"shard_0" failed`) || !strings.Contains(text, `"shard_1" succeeded`) {
				t.Fatalf("transport failure affected healthy source: %s %s", stdout.String(), stderr.String())
			}
			for i, db := range databases {
				var history int
				if err := db.QueryRowContext(ctx, "SELECT count(*) FROM migration_schema").Scan(&history); err != nil {
					t.Fatal(err)
				}
				want := 0
				if mode == "transport" && i == 1 {
					want = 1
				}
				if history != want {
					t.Fatalf("shard_%d history=%d, want %d", i, history, want)
				}
				var exists bool
				if err := db.QueryRowContext(ctx, "SELECT to_regclass('batch_items') IS NOT NULL").Scan(&exists); err != nil {
					t.Fatal(err)
				}
				if exists != (want == 1) {
					t.Fatalf("shard_%d failed transaction left DDL behind", i)
				}
			}
			if mode == "transport" {
				source := cfg.DataSources["shard_0"]
				source.SSHTunnel = nil
				cfg.DataSources["shard_0"] = source
				writeConfig()
			}
			if data, err := command("up", "--parallel", "2").CombinedOutput(); err != nil {
				t.Fatalf("retry: %v %s", err, data)
			}
			for _, db := range databases {
				var count int
				if err := db.QueryRowContext(ctx, "SELECT count(*) FROM batch_items").Scan(&count); err != nil || count != 1 {
					t.Fatalf("retry replayed or lost migration: count=%d err=%v", count, err)
				}
			}
		})
	}
}
