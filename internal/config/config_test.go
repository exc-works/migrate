package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReplaceEnvVar(t *testing.T) {
	t.Setenv("APP_USER", "alice")
	out, err := ReplaceEnvVar("user=${APP_USER} pass=${APP_PASS:default}")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "user=alice pass=default" {
		t.Fatalf("unexpected output: %s", out)
	}
}

func TestReplaceEnvVarMissing(t *testing.T) {
	_, err := ReplaceEnvVar("x=${NOT_SET}")
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestReadParsesConfigWithSpecialPasswordCharacters(t *testing.T) {
	unsetEnvForTest(t, "HOST")
	unsetEnvForTest(t, "PORT")

	dir := t.TempDir()
	path := filepath.Join(dir, "migration_config.json")
	content := []byte(`{
  "schema_name": "",
  "dialect": "postgres",
  "data_source_name": "host=${HOST:127.0.0.1} port=${PORT:5432} user=postgres password=YOU$,PASSWORD> dbname=test sslmode=disable TimeZone=UTC",
  "migration_source": "migrations",
  "migrate_out_of_order": true,
  "disable_color_output": false
}`)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Read(path, "")
	if err != nil {
		t.Fatalf("read config: %v", err)
	}

	if cfg.SchemaName != "migration_schema" {
		t.Fatalf("expected default schema_name, got %q", cfg.SchemaName)
	}
	if cfg.Dialect != "postgres" {
		t.Fatalf("unexpected dialect: %q", cfg.Dialect)
	}
	wantDSN := "host=127.0.0.1 port=5432 user=postgres password=YOU$,PASSWORD> dbname=test sslmode=disable TimeZone=UTC"
	if cfg.DataSourceName != wantDSN {
		t.Fatalf("unexpected data_source_name: want %q, got %q", wantDSN, cfg.DataSourceName)
	}
	if cfg.MigrationSource != "migrations" {
		t.Fatalf("unexpected migration_source: %q", cfg.MigrationSource)
	}
	if !cfg.MigrateOutOfOrder {
		t.Fatalf("expected migrate_out_of_order to be true")
	}
}

func TestReadReportsInvalidJSONLocation(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "migration_config.json")
	content := []byte("{\"data_source_name\":\"ignored\"}\nLog line that is not JSON\n")
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, err := Read(path, "")
	if err == nil {
		t.Fatalf("expected error")
	}

	got := err.Error()
	for _, want := range []string{
		"invalid JSON in config file",
		"migration_config.json",
		"line 2, column 1",
		"extra content after the JSON object",
		"invalid character 'L' after top-level value",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected %q in error %q", want, got)
		}
	}
}

func unsetEnvForTest(t *testing.T, key string) {
	t.Helper()

	value, ok := os.LookupEnv(key)
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("unset %s: %v", key, err)
	}
	t.Cleanup(func() {
		var err error
		if ok {
			err = os.Setenv(key, value)
		} else {
			err = os.Unsetenv(key)
		}
		if err != nil {
			t.Fatalf("restore %s: %v", key, err)
		}
	})
}
