package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/exc-works/migrate/internal/config"
	"github.com/exc-works/migrate/internal/testssh"
	"github.com/exc-works/migrate/internal/tunnel"
)

func TestCommandClosesDBOnSuccessAndFailure(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(strconv.FormatBool(fail), func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			previousOpen, previousPing := openDB, pingDB
			defer func() { openDB = previousOpen; pingDB = previousPing }()
			openDB = func(string, string) (*sql.DB, error) { return db, nil }
			pingDB = func(context.Context, *sql.DB) error { return nil }
			query := mock.ExpectExec("CREATE TABLE")
			if fail {
				query.WillReturnError(errors.New("example execution failure"))
			} else {
				query.WillReturnResult(sqlmock.NewResult(0, 0))
			}
			mock.ExpectClose()
			path := filepath.Join(t.TempDir(), "config.json")
			if err := os.WriteFile(path, []byte(`{"data_source_name":"example"}`), 0600); err != nil {
				t.Fatal(err)
			}
			var out, stderr bytes.Buffer
			cmd := newRootCommand(&out, &stderr)
			cmd.SetArgs([]string{"create", "--config", path})
			err = cmd.Execute()
			if (err != nil) != fail {
				t.Fatalf("unexpected command result: %v", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestSSHNoninteractiveSecret(t *testing.T) {
	s := testssh.New(t)
	host, port, _ := net.SplitHostPort(s.Address)
	p, _ := strconv.Atoi(port)
	const variable = "MIGRATE_CLI_EXAMPLE_SECRET"
	t.Setenv(variable, "")
	if err := os.Unsetenv(variable); err != nil {
		t.Fatal(err)
	}
	cfg := config.DefaultTemplate()
	cfg.DataSourceName = "host=db.internal.example sslmode=disable"
	cfg.SSHTunnel = &tunnel.Options{Host: host, Port: p, User: "example", Auth: tunnel.AuthOptions{Type: tunnel.AuthPasswordMode, PasswordEnv: variable}, KnownHostsFile: s.KnownHosts}
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	var out, stderr bytes.Buffer
	cmd := newRootCommand(&out, &stderr)
	cmd.SetArgs([]string{"create", "--config", path, "--non-interactive"})
	if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), "secret unavailable") {
		t.Fatalf("unexpected secret failure: %v", err)
	}
	if out.Len() != 0 || stderr.Len() != 0 {
		t.Fatalf("noninteractive command prompted: %q %q", out.String(), stderr.String())
	}
}

func TestSSHNoninteractiveRejectsUnknownHost(t *testing.T) {
	s := testssh.New(t)
	s.Serve()
	host, port, _ := net.SplitHostPort(s.Address)
	p, _ := strconv.Atoi(port)
	t.Setenv("MIGRATE_CLI_HOSTKEY_EXAMPLE_SECRET", "example")
	known := filepath.Join(t.TempDir(), "known_hosts")
	if err := os.WriteFile(known, nil, 0600); err != nil {
		t.Fatal(err)
	}
	cfg := config.DefaultTemplate()
	cfg.DataSourceName = "host=db.internal.example sslmode=disable"
	cfg.SSHTunnel = &tunnel.Options{Host: host, Port: p, User: "example", Auth: tunnel.AuthOptions{Type: tunnel.AuthPasswordMode, PasswordEnv: "MIGRATE_CLI_HOSTKEY_EXAMPLE_SECRET"}, KnownHostsFile: known}
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	var out, stderr bytes.Buffer
	cmd := newRootCommand(&out, &stderr)
	cmd.SetArgs([]string{"create", "--config", path, "--non-interactive"})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := cmd.ExecuteContext(ctx); err == nil || !strings.Contains(err.Error(), "host key is unknown") {
		t.Fatalf("unknown host: %v", err)
	}
	if out.Len() != 0 || stderr.Len() != 0 {
		t.Fatal("noninteractive host trust prompted")
	}
	after, err := os.ReadFile(known)
	if err != nil || len(after) != 0 {
		t.Fatalf("noninteractive trust changed: %v", err)
	}
}
