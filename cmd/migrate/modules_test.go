package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/exc-works/migrate/internal/script"
)

func TestCLIDirectoryModulesWorkflow(t *testing.T) {
	root := t.TempDir()
	database := filepath.Join(root, "modules.db")
	files := map[string]string{
		"migrations/V1__create.sql":           "-- +migrate Up\nCREATE TABLE users(profile TEXT);\n-- +migrate Down\nDROP TABLE users;\n",
		"migrations/V2__seed.ts":              `import {normalize} from '../shared/helper.ts';export const config={transaction:false};export function up(db){db.exec('BEGIN');try{db.exec('INSERT INTO users VALUES (?)',[normalize(' Alice ')]);db.exec('COMMIT')}catch(e){db.exec('ROLLBACK');throw e}}export function down(db){db.exec('DELETE FROM users')}`,
		"migrations/V3__change.mjs":           `import {normalize} from '../shared/helper.ts';export function up(db){db.exec('UPDATE users SET profile=?',[normalize(' Bob ')])}export function down(db){db.exec('UPDATE users SET profile=?',[normalize(' Alice ')])}`,
		"shared/helper.ts":                    `throw Error('dry-run must not evaluate modules');export function normalize(value){return value}`,
		"node_modules/normalize/package.json": `{"main":"index.js"}`,
		"node_modules/normalize/index.js":     `export function normalize(value){return value.trim().toLowerCase()}`,
	}
	for name, text := range files {
		file := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(file), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file, []byte(text), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	config, err := json.Marshal(map[string]string{"dialect": "sqlite", "data_source_name": database})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "migration_config.json"), config, 0o600); err != nil {
		t.Fatal(err)
	}
	run := func(arguments ...string) error {
		t.Helper()
		var output bytes.Buffer
		command := newRootCommand(&output, &output)
		command.SetArgs(append([]string{"--working-dir", root}, arguments...))
		return command.Execute()
	}
	if err := run("create"); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", database)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if len(script.Formats()) == 1 {
		if err := run("up"); err == nil {
			t.Fatal("SQL build accepted module migrations")
		}
		var count int
		if err := db.QueryRow("SELECT count(*) FROM sqlite_master WHERE name='users'").Scan(&count); err != nil || count != 0 {
			t.Fatalf("default build partially executed: %d %v", count, err)
		}
		return
	}
	if err := run("up", "--dry-run"); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRow("SELECT count(*) FROM sqlite_master WHERE name='users'").Scan(&count); err != nil || count != 0 {
		t.Fatalf("dry-run wrote data: %d %v", count, err)
	}
	if err := os.WriteFile(filepath.Join(root, "shared", "helper.ts"), []byte(`export {normalize} from 'normalize';`), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, command := range []string{"status", "up"} {
		if err := run(command); err != nil {
			t.Fatal(err)
		}
	}
	var profile string
	if err := db.QueryRow("SELECT profile FROM users").Scan(&profile); err != nil || profile != "bob" {
		t.Fatalf("module up: %s %v", profile, err)
	}
	if err := run("down", "2"); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow("SELECT profile FROM users").Scan(&profile); err != nil || profile != "alice" {
		t.Fatalf("module down: %s %v", profile, err)
	}
	if err := run("down", "--all"); err != nil {
		t.Fatal(err)
	}
}
