//go:build full

package migrate

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/exc-works/migrate/internal/source"
)

func writeMigrationFiles(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for name, text := range files {
		file := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(file), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file, []byte(text), 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

func TestMissingImportPreventsAllSelectedWrites(t *testing.T) {
	for _, dry := range []bool{false, true} {
		root := t.TempDir()
		writeMigrationFiles(t, root, map[string]string{
			"V1__create.sql": migrationSQL("CREATE TABLE users(id INTEGER);", "DROP TABLE users;"),
			"V2__import.js":  `import {up} from './missing.js';export {up};`,
		})
		svc, db, _ := mixedService(t, nil, dry)
		svc.cfg.MigrationSource = source.DirectorySource{Directory: root}
		if _, err := svc.Status(); err != nil {
			t.Fatal(err)
		}
		if err := svc.Up(); err == nil {
			t.Fatal("missing import passed preflight")
		}
		var count int
		if err := db.QueryRow("SELECT count(*) FROM sqlite_master WHERE name='users'").Scan(&count); err != nil || count != 0 {
			t.Fatalf("partially executed plan: %d %v", count, err)
		}
	}
}

func TestMissingImportPreventsPartialRollback(t *testing.T) {
	root := t.TempDir()
	writeMigrationFiles(t, root, map[string]string{
		"helper.js":     `export function create(db){db.exec('CREATE TABLE users(id INTEGER)')}export function drop(db){db.exec('DROP TABLE users')}`,
		"V1__create.js": `import {create,drop} from './helper.js';export const up=create;export const down=drop;`,
		"V2__insert.js": `export function up(db){db.exec('INSERT INTO users VALUES (1)')}export function down(db){db.exec('DELETE FROM users')}`,
	})
	svc, db, _ := mixedService(t, nil, false)
	svc.cfg.MigrationSource = source.DirectorySource{Directory: root}
	if err := svc.Up(); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(root, "helper.js")); err != nil {
		t.Fatal(err)
	}
	if err := svc.Down("", true); err == nil {
		t.Fatal("missing rollback dependency passed preflight")
	}
	var count int
	if err := db.QueryRow("SELECT count(*) FROM users").Scan(&count); err != nil || count != 1 {
		t.Fatalf("partially rolled back plan: %d %v", count, err)
	}
}
