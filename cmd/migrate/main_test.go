package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/exc-works/migrate/internal/config"
	"github.com/exc-works/migrate/internal/migrate"
)

func TestEffectiveBuildVersion(t *testing.T) {
	origBuildVersion := buildVersion
	defer func() {
		buildVersion = origBuildVersion
	}()

	buildVersion = ""
	if got := effectiveBuildVersion(); got != "dev" {
		t.Fatalf("expected dev fallback, got %q", got)
	}

	buildVersion = "v1.2.3"
	if got := effectiveBuildVersion(); got != "v1.2.3" {
		t.Fatalf("expected explicit version, got %q", got)
	}
}

func TestVersionCommand(t *testing.T) {
	origBuildVersion := buildVersion
	origOpenDB := openDB
	defer func() {
		buildVersion = origBuildVersion
		openDB = origOpenDB
	}()

	buildVersion = "v9.9.9"
	openDB = func(driverName, dsn string) (*sql.DB, error) {
		t.Fatalf("version command should not open a database")
		return nil, nil
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := newRootCommand(&stdout, &stderr)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"version"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := stdout.String(); got != "v9.9.9\n" {
		t.Fatalf("unexpected stdout: %q", got)
	}
	if got := stderr.String(); got != "" {
		t.Fatalf("expected empty stderr, got %q", got)
	}
}

func TestVersionCommandRejectsArgs(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := newRootCommand(&stdout, &stderr)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"version", "extra"})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error for extra args")
	}
	if !strings.Contains(err.Error(), "unknown command") {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := stdout.String(); got != "" {
		t.Fatalf("expected empty stdout, got %q", got)
	}
}

func TestRootCommandSilenceSettings(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := newRootCommand(&stdout, &stderr)
	if !cmd.SilenceErrors {
		t.Fatalf("expected SilenceErrors enabled")
	}
	if !cmd.SilenceUsage {
		t.Fatalf("expected SilenceUsage enabled")
	}
}

func TestCreateCommandDoesNotAutoPrintErrorOrUsage(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := newRootCommand(&stdout, &stderr)
	cmd.SetArgs([]string{"create"})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "migration_config.json") {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := stdout.String(); got != "" {
		t.Fatalf("expected empty stdout, got %q", got)
	}
	if got := stderr.String(); got != "" {
		t.Fatalf("expected empty stderr, got %q", got)
	}
}

func TestStatusWorkingDirReportsInvalidConfigJSON(t *testing.T) {
	dir := t.TempDir()
	repoDir := filepath.Join(dir, "repo")
	if err := os.Mkdir(repoDir, 0o755); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}
	content := []byte("{\"data_source_name\":\"ignored\"}\nLog line that is not JSON\n")
	if err := os.WriteFile(filepath.Join(repoDir, "migration_config.json"), content, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := newRootCommand(&stdout, &stderr)
	cmd.SetArgs([]string{"status", "-w", repoDir})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	got := err.Error()
	for _, want := range []string{
		filepath.Join("repo", "migration_config.json"),
		"line 2, column 1",
		"extra content after the JSON object",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected %q in error %q", want, got)
		}
	}
	if stdout.String() != "" || stderr.String() != "" {
		t.Fatalf("expected empty command output before main handles error, stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
}

func TestDownCommandRejectsMissingTargetBeforeConfigRead(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := newRootCommand(&stdout, &stderr)
	cmd.SetArgs([]string{"down"})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if got := err.Error(); got != "to-version must be set, or use --all" {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(err.Error(), "migration_config.json") {
		t.Fatalf("expected argument error before config read, got: %v", err)
	}
}

func TestDownCommandRejectsToVersionWithAll(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := newRootCommand(&stdout, &stderr)
	cmd.SetArgs([]string{"down", "20260417090100", "--all"})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "mutually exclusive") {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(err.Error(), "migration_config.json") {
		t.Fatalf("expected argument validation error before config read, got: %v", err)
	}
}

func TestPrintStatusJSON(t *testing.T) {
	items := []migrate.MigrationStatus{
		{
			Migration: migrate.Migration{
				Version:  "20260417090000",
				Filename: "V20260417090000__init.sql",
				Hash:     "abc123",
			},
			Status: migrate.StatusPending,
		},
	}

	var out bytes.Buffer
	if err := printStatus(&out, items, "json"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var got []statusOutputItem
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("decode json: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 row, got %d", len(got))
	}
	if got[0].Version != "20260417090000" || got[0].Filename != "V20260417090000__init.sql" || got[0].Hash != "abc123" || got[0].Status != migrate.StatusPending {
		t.Fatalf("unexpected row: %#v", got[0])
	}
}

func TestPrintStatusRejectsUnknownOutputFormat(t *testing.T) {
	err := printStatus(io.Discard, nil, "yaml")
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "unsupported output format") {
		t.Fatalf("unexpected error: %v", err)
	}
}

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
