package source

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func TestDirectorySourceLoadMigrations(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "V1__init.sql"), []byte("-- +migrate Up\nSELECT 1;\n-- +migrate Down\nSELECT 1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("ignored"), 0o644); err != nil {
		t.Fatal(err)
	}
	src := DirectorySource{Directory: dir}
	migrations, err := src.LoadMigrations()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(migrations) != 1 {
		t.Fatalf("expected 1 migration, got %d", len(migrations))
	}
}

func TestFSSourceLoadMigrations(t *testing.T) {
	fsys := fstest.MapFS{
		"migrations/V1__init.sql":  {Data: []byte("-- +migrate Up\nCREATE TABLE a (id INT);\n")},
		"migrations/V2__add.sql":   {Data: []byte("-- +migrate Up\nCREATE TABLE b (id INT);\n")},
		"migrations/README.md":     {Data: []byte("ignored")},
		"migrations/sub/V9__x.sql": {Data: []byte("nested and ignored")},
		"other/V3__other.sql":      {Data: []byte("outside root")},
	}

	got, err := FSSource{FS: fsys, Root: "migrations"}.LoadMigrations()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 migrations, got %d: %+v", len(got), got)
	}
	if got[0].Filename != "V1__init.sql" || got[1].Filename != "V2__add.sql" {
		t.Fatalf("unexpected order: %+v", got)
	}
	if got[0].Source != "-- +migrate Up\nCREATE TABLE a (id INT);\n" {
		t.Fatalf("unexpected content: %q", got[0].Source)
	}
}

func TestFSSourceEmptyRootDefaultsToFSRoot(t *testing.T) {
	fsys := fstest.MapFS{
		"V1__one.sql": {Data: []byte("x")},
		"notes.txt":   {Data: []byte("ignored")},
	}
	got, err := FSSource{FS: fsys}.LoadMigrations()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(got) != 1 || got[0].Filename != "V1__one.sql" {
		t.Fatalf("expected single V1__one.sql, got %+v", got)
	}
}
