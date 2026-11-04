package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewVersionConfigurationFallback(t *testing.T) {
	dir := t.TempDir()
	if _, err := readConfigOrDefault("", dir); err != nil {
		t.Fatalf("absent configuration should use defaults: %v", err)
	}
	if _, err := readConfigOrDefault("missing.json", dir); err == nil {
		t.Fatal("explicit missing config should not use defaults")
	}
	if err := os.WriteFile(filepath.Join(dir, "migration_config.yml"), []byte("data_source_name: [broken\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := readConfigOrDefault("", dir); err == nil {
		t.Fatal("invalid YAML should not use defaults")
	}
}

func TestMultiDatabaseCommandIsExplicitlyDeferred(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "migration_config.yml"), []byte(`data_sources:
  a:
    data_source_name: example
`), 0600); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	command := newRootCommand(&out, &out)
	command.SetArgs([]string{"create", "-w", dir})
	if err := command.Execute(); err == nil || !strings.Contains(err.Error(), "multi-database execution is not available yet") {
		t.Fatalf("expected explicit deferral before connection: %v", err)
	}
}
