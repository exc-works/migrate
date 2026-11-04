package main

import (
	"os"
	"path/filepath"
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
