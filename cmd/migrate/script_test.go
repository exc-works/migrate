package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/exc-works/migrate/internal/script"
)

func TestCapabilitiesNeverReadsConfigOrConnects(t *testing.T) {
	original := openDB
	t.Cleanup(func() { openDB = original })
	openDB = func(string, string) (*sql.DB, error) { t.Fatal("capabilities connected to database"); return nil, nil }
	var out bytes.Buffer
	cmd := newRootCommand(&out, &out)
	cmd.SetArgs([]string{"--config", filepath.Join(t.TempDir(), "missing.json"), "capabilities", "--output", "json"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	var got struct {
		Build   string   `json:"build"`
		Formats []string `json:"formats"`
	}
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Formats) != len(script.Formats()) || got.Formats[0] != "sql" {
		t.Fatalf("%s", out.String())
	}
}

func TestNewVersionFormats(t *testing.T) {
	for _, format := range []string{"sql", "js", "ts", "unknown"} {
		t.Run(format, func(t *testing.T) {
			root := t.TempDir()
			var out bytes.Buffer
			cmd := newRootCommand(&out, &out)
			cmd.SetArgs([]string{"--working-dir", root, "new", "version", "example", "--version", "1", "--format", format})
			err := cmd.Execute()
			unsupported := format == "unknown" || format != "sql" && len(script.Formats()) == 1
			if unsupported {
				if err == nil {
					t.Fatal("expected unsupported format")
				}
				if _, err := os.Stat(filepath.Join(root, "migrations")); !os.IsNotExist(err) {
					t.Fatal("unsupported format created directory")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			content, err := os.ReadFile(filepath.Join(root, "migrations", "V1__example."+format))
			if err != nil {
				t.Fatal(err)
			}
			if format == "sql" && !strings.Contains(string(content), "-- +migrate Up") {
				t.Fatal("SQL annotation missing")
			}
			if format != "sql" && (!strings.Contains(string(content), "export function up(") || strings.Contains(string(content), "export function down(")) {
				t.Fatalf("unsafe script template: %s", content)
			}
			if format != "sql" {
				program, err := script.Compile("V1__example."+format, string(content), "up", script.Options{})
				if err != nil {
					t.Fatalf("generated template does not compile: %v", err)
				}
				invocation, err := program.Instantiate(context.Background())
				if err != nil {
					t.Fatalf("generated entrypoint is invalid: %v", err)
				}
				if err = invocation.Run(context.Background(), nil); err == nil {
					t.Fatal("unimplemented migration template must fail")
				}
			}
		})
	}
}
