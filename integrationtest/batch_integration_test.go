//go:build integration

package integrationtest

import (
	"context"
	"encoding/json"
	"net"
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
	"github.com/exc-works/migrate/internal/tunnel"
)

func TestBatchCLISharedSSHAndIndependentDatabases(t *testing.T) {
	if !slices.Contains(selectedFlavors(), dialect.NamePostgres) {
		t.Skip("shared SSH batch workflow uses PostgreSQL")
	}
	d, _ := dialect.FromName(dialect.NamePostgres)
	admin, baseDSN, cleanup := startContainerDBWithDSN(t, dialect.NamePostgres, d)
	defer cleanup()
	dir := t.TempDir()
	binary := filepath.Join(dir, "migrate")
	buildCtx, stopBuild := context.WithTimeout(context.Background(), 2*time.Minute)
	defer stopBuild()
	build := exec.CommandContext(buildCtx, "go", "build", "-o", binary, "./cmd/migrate")
	build.Dir = ".."
	if data, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build batch CLI: %v %s", err, data)
	}
	sources := make(map[string]config.DataSourceConfig)
	var remoteTarget string
	for _, name := range []string{"a", "b"} {
		database := randomName("batch_database")
		if _, err := admin.Exec("CREATE DATABASE " + database); err != nil {
			t.Fatal(err)
		}
		dsn := strings.Replace(baseDSN, "dbname=testdb", "dbname="+database, 1)
		target, tunneledDSN := sshDatabaseDSN(t, dialect.NamePostgres, dsn)
		remoteTarget = target
		sources[name] = config.DataSourceConfig{DataSourceName: tunneledDSN}
	}
	server := testssh.New(t)
	server.Forward = func(string) (net.Conn, error) { return net.Dial("tcp", remoteTarget) }
	server.Serve()
	ssh := workflowSSHOptions(t, server)
	cfg := config.FileConfig{Dialect: dialect.NamePostgres, SSHTunnel: &ssh, DataSources: sources}
	writeConfig := func() {
		t.Helper()
		data, err := json.Marshal(cfg)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "config.json"), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	writeConfig()
	migrations := filepath.Join(dir, "migrations")
	if err := os.Mkdir(migrations, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(migrations, "V1__items.sql"), []byte("-- +migrate Up\nCREATE TABLE batch_items(id INT);\n-- +migrate Down\nDROP TABLE batch_items;\n"), 0600); err != nil {
		t.Fatal(err)
	}
	run := func(wantFailure bool, args ...string) []byte {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, binary, append([]string{"--working-dir", dir, "--config", "config.json", "--non-interactive"}, args...)...)
		data, err := cmd.CombinedOutput()
		if (err != nil) != wantFailure {
			t.Fatalf("batch %v: %v %s", args, err, data)
		}
		return data
	}
	run(false, "create", "--parallel", "2")
	run(false, "up", "--parallel", "2")
	if err := os.WriteFile(filepath.Join(migrations, "V2__audit.sql"), []byte("-- +migrate Up\nCREATE TABLE batch_audit(id INT);\n-- +migrate Down\nDROP TABLE batch_audit;\n"), 0600); err != nil {
		t.Fatal(err)
	}
	const missingVariable = "MIGRATE_BATCH_MISSING_EXAMPLE_SECRET"
	t.Setenv(missingVariable, "")
	if err := os.Unsetenv(missingVariable); err != nil {
		t.Fatal(err)
	}
	badSSH := ssh
	badSSH.Auth = tunnel.AuthOptions{Type: tunnel.AuthPasswordMode, PasswordEnv: missingVariable}
	source := cfg.DataSources["a"]
	source.SSHTunnel = &badSSH
	cfg.DataSources["a"] = source
	writeConfig()
	failed := run(true, "up", "--parallel", "2")
	if !strings.Contains(string(failed), "failed") || !strings.Contains(string(failed), "succeeded") {
		t.Fatal("SSH failure canceled the healthy database")
	}
	source.SSHTunnel = nil
	cfg.DataSources["a"] = source
	writeConfig()
	// a still needs V2; b already applied it through the shared SSH defaults.
	before := run(false, "status", "--output", "json", "--parallel", "2")
	var statuses []struct {
		DataSource string `json:"data_source"`
		Migrations []struct {
			Status string `json:"status"`
		} `json:"migrations"`
	}
	if err := json.Unmarshal(before, &statuses); err != nil || len(statuses) != 2 || len(statuses[0].Migrations) != 2 || len(statuses[1].Migrations) != 2 {
		t.Fatalf("batch JSON: %v %s", err, before)
	}
	if statuses[0].Migrations[1].Status != "pending" || statuses[1].Migrations[1].Status != "applied" {
		t.Fatal("independent history or failure isolation was lost")
	}
	run(false, "up", "--parallel", "2")
	run(false, "down", "--all", "--parallel", "2")
}
