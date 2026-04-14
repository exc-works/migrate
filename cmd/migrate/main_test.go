package main

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/exc-works/migrate/internal/config"
	"github.com/exc-works/migrate/internal/migrate"
)

func TestSanitizeDescription(t *testing.T) {
	t.Run("accepts safe description", func(t *testing.T) {
		got, err := sanitizeDescription("init users")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "init_users" {
			t.Fatalf("unexpected sanitized value: %s", got)
		}
	})

	t.Run("rejects path traversal and separators", func(t *testing.T) {
		inputs := []string{"../oops", "a/b", `a\b`, "unsafe-flag"}
		for _, in := range inputs {
			_, err := sanitizeDescription(in)
			if err == nil {
				t.Fatalf("expected error for input %q", in)
			}
		}
	})
}

func TestSanitizeVersion(t *testing.T) {
	t.Run("accepts safe version", func(t *testing.T) {
		got, err := sanitizeVersion("20260413153000")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "20260413153000" {
			t.Fatalf("unexpected sanitized value: %s", got)
		}
	})

	t.Run("rejects path traversal and separators", func(t *testing.T) {
		inputs := []string{"../oops", "a/b", `a\b`, "v1.0"}
		for _, in := range inputs {
			_, err := sanitizeVersion(in)
			if err == nil {
				t.Fatalf("expected error for input %q", in)
			}
		}
	})
}

func TestSecureJoinWithin(t *testing.T) {
	base := t.TempDir()

	okPath, err := secureJoinWithin(base, "V1__init.sql")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(okPath, filepath.Clean(base)+string(filepath.Separator)) {
		t.Fatalf("expected path under base dir, got %s", okPath)
	}

	if _, err := secureJoinWithin(base, "../pwn.sql"); err == nil {
		t.Fatalf("expected escape path to fail")
	}
}

func TestNewServiceFromConfigClosesDBOnServiceCreateFailure(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}

	restore := stubDBInitDeps(t, db)
	defer restore()

	pingDB = func(ctx context.Context, gotDB *sql.DB) error {
		if gotDB != db {
			t.Fatalf("unexpected db in ping")
		}
		return nil
	}
	wantErr := errors.New("service create failed")
	newMigrateService = func(ctx context.Context, cfg migrate.Config) (*migrate.Service, error) {
		return nil, wantErr
	}

	_, gotErr := newServiceFromConfig(&config.FileConfig{
		Dialect:         "postgres",
		DataSourceName:  "ignored",
		SchemaName:      "migration_schema",
		MigrationSource: "migrations",
	})
	if !errors.Is(gotErr, wantErr) {
		t.Fatalf("expected error %v, got %v", wantErr, gotErr)
	}

	if pingErr := db.Ping(); pingErr == nil || !strings.Contains(pingErr.Error(), "closed") {
		t.Fatalf("expected db to be closed, ping err: %v", pingErr)
	}
}

func TestNewServiceFromConfigUsesPingTimeout(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}

	restore := stubDBInitDeps(t, db)
	defer restore()

	pingTimeout = 2 * time.Second
	defer func() {
		pingTimeout = 5 * time.Second
	}()

	var timeoutSeen time.Duration
	pingDB = func(ctx context.Context, gotDB *sql.DB) error {
		deadline, ok := ctx.Deadline()
		if !ok {
			t.Fatalf("expected ping context deadline")
		}
		timeoutSeen = time.Until(deadline)
		return nil
	}
	newMigrateService = func(ctx context.Context, cfg migrate.Config) (*migrate.Service, error) {
		return &migrate.Service{}, nil
	}

	_, err = newServiceFromConfig(&config.FileConfig{
		Dialect:         "postgres",
		DataSourceName:  "ignored",
		SchemaName:      "migration_schema",
		MigrationSource: "migrations",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if timeoutSeen <= 0 || timeoutSeen > pingTimeout {
		t.Fatalf("unexpected timeout seen: %v", timeoutSeen)
	}
}

func stubDBInitDeps(t *testing.T, db *sql.DB) func() {
	t.Helper()
	origOpenDB := openDB
	origPingDB := pingDB
	origNewService := newMigrateService
	origPingTimeout := pingTimeout

	openDB = func(driverName, dsn string) (*sql.DB, error) {
		return db, nil
	}

	return func() {
		openDB = origOpenDB
		pingDB = origPingDB
		newMigrateService = origNewService
		pingTimeout = origPingTimeout
		_ = db.Close()
	}
}
