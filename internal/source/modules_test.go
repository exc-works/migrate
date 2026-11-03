package source

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

type countingEntriesSource struct {
	legacy, extended int
	entries          []MigrationEntry
}

func (s *countingEntriesSource) LoadMigrations() ([]Migration, error) {
	s.legacy++
	return nil, nil
}

func (s *countingEntriesSource) LoadMigrationEntries() ([]MigrationEntry, error) {
	s.extended++
	return s.entries, nil
}

func TestLoadEntriesUsesOneSnapshot(t *testing.T) {
	src := &countingEntriesSource{entries: []MigrationEntry{{Migration: Migration{"V1__one.js", "entry"}}}}
	entries, err := LoadEntries(src)
	if err != nil || len(entries) != 1 || src.extended != 1 || src.legacy != 0 {
		t.Fatalf("entries=%+v source=%+v err=%v", entries, src, err)
	}
	legacy := StringSource{[]Migration{{"V2__two.js", "standalone"}}}
	entries, err = LoadEntries(legacy)
	if err != nil || len(entries) != 1 || entries[0].Modules != nil || entries[0].Migration.Source != "standalone" {
		t.Fatalf("legacy entries=%+v err=%v", entries, err)
	}
}

func TestCombinedEntriesKeepSourceContext(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "V2__disk.js"), []byte("disk"), 0o600); err != nil {
		t.Fatal(err)
	}
	virtual := fstest.MapFS{"migrations/V1__virtual.ts": {Data: []byte("virtual")}}
	src := CombinedSource{[]MigrationSource{DirectorySource{dir}, FSSource{virtual, "migrations"}}}
	entries, err := LoadEntries(src)
	if err != nil || len(entries) != 2 {
		t.Fatalf("%+v %v", entries, err)
	}
	if entries[0].Modules.Directory != dir || entries[0].Modules.EntryPath != "V2__disk.js" {
		t.Fatalf("disk context: %+v", entries[0].Modules)
	}
	context := entries[1].Modules
	if context.Directory != "" || context.EntryPath != "migrations/V1__virtual.ts" {
		t.Fatalf("virtual context: %+v", context)
	}
	data, err := fs.ReadFile(context.FS, context.EntryPath)
	if err != nil || string(data) != "virtual" {
		t.Fatalf("virtual content: %s %v", data, err)
	}
}

func TestWithModulesPreservesUnderlyingContext(t *testing.T) {
	original := &ModuleContext{FS: fstest.MapFS{}, EntryPath: "migrations/V1__one.js"}
	src := &countingEntriesSource{entries: []MigrationEntry{{Migration: Migration{"V1__one.js", "original bytes"}, Modules: original}}}
	dependencyDir := t.TempDir()
	wrapped := WithModules{Source: src, DependencyDir: dependencyDir}
	entries, err := wrapped.LoadMigrationEntries()
	if err != nil || len(entries) != 1 {
		t.Fatalf("%+v %v", entries, err)
	}
	context := entries[0].Modules
	if context == original || original.DependencyDir != "" || context.DependencyDir != dependencyDir || context.EntryPath != original.EntryPath || context.FS == nil {
		t.Fatalf("original=%+v wrapped=%+v", original, context)
	}
	if entries[0].Migration.Source != "original bytes" || src.extended != 1 || src.legacy != 0 {
		t.Fatal("wrapper reloaded or replaced source bytes")
	}
}

func TestWithModulesConfiguresLegacyEntries(t *testing.T) {
	legacy := StringSource{[]Migration{{"V1__one.js", "entry"}}}
	virtual := fstest.MapFS{"helpers/one.js": {Data: []byte("helper")}}
	wrapped := WithModules{Source: legacy, FS: virtual, Root: "migrations"}
	entries, err := LoadEntries(wrapped)
	if err != nil || entries[0].Modules.EntryPath != "migrations/V1__one.js" || entries[0].Modules.FS == nil {
		t.Fatalf("%+v %v", entries, err)
	}
	wrapped = WithModules{Source: legacy, Directory: "."}
	entries, err = LoadEntries(wrapped)
	if err != nil || !filepath.IsAbs(entries[0].Modules.Directory) || entries[0].Modules.EntryPath != "V1__one.js" {
		t.Fatalf("%+v %v", entries, err)
	}
	for _, invalid := range []WithModules{
		{},
		{Source: legacy, FS: virtual, Directory: "."},
		{Source: legacy, Root: "migrations"},
		{Source: legacy, FS: virtual, Root: "../outside"},
	} {
		if _, err := invalid.LoadMigrationEntries(); err == nil {
			t.Fatalf("accepted invalid wrapper: %+v", invalid)
		}
	}
}
