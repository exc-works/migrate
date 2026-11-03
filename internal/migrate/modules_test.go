package migrate

import (
	"crypto/sha256"
	"fmt"
	"testing"

	"github.com/exc-works/migrate/internal/source"
)

type moduleEntriesSource struct {
	loads   int
	entries []source.MigrationEntry
}

func (s *moduleEntriesSource) LoadMigrations() ([]source.Migration, error) {
	panic("extended source should be loaded only once")
}

func (s *moduleEntriesSource) LoadMigrationEntries() ([]source.MigrationEntry, error) {
	s.loads++
	return s.entries, nil
}

func TestModuleContextsFollowVersionSelection(t *testing.T) {
	first, second := &source.ModuleContext{EntryPath: "first"}, &source.ModuleContext{EntryPath: "second"}
	src := &moduleEntriesSource{entries: []source.MigrationEntry{
		{Migration: source.Migration{Filename: "V2__second.sql", Source: migrationSQL("SELECT 2;", "SELECT 2;")}, Modules: second},
		{Migration: source.Migration{Filename: "V1__first.sql", Source: migrationSQL("SELECT 1;", "SELECT 1;")}, Modules: first},
	}}
	svc := &Service{cfg: Config{MigrationSource: src}}
	loaded, err := svc.loadMigrationSet(true, true)
	if err != nil || src.loads != 1 || len(loaded.migrations) != 2 {
		t.Fatalf("%+v loads=%d err=%v", loaded, src.loads, err)
	}
	if loaded.migrations[0].Version != "1" || loaded.modules["1"] != first || loaded.modules["2"] != second {
		t.Fatal("sorting detached module context from the entry")
	}
	selected := []Migration{loaded.migrations[1], loaded.migrations[0]}
	plan, err := svc.preparePlanWithModules(selected, "down", loaded.modules)
	if err != nil || len(plan) != 2 || plan[0].modules != second || plan[1].modules != first {
		t.Fatalf("plan=%+v err=%v", plan, err)
	}
	src.entries = append(src.entries, src.entries[0])
	if _, err := svc.loadMigrationSet(false, false); err == nil {
		t.Fatal("module contexts bypassed duplicate version detection")
	}
}

func TestModuleContextDoesNotChangeScriptHash(t *testing.T) {
	text := "export function up(){}"
	src := &moduleEntriesSource{entries: []source.MigrationEntry{{Migration: source.Migration{Filename: "V1__one.js", Source: text}, Modules: &source.ModuleContext{EntryPath: "one"}}}}
	svc := &Service{cfg: Config{MigrationSource: src}}
	loaded, err := svc.loadMigrationSet(false, false)
	if err != nil || loaded.migrations[0].Hash != fmt.Sprintf("%x", sha256.Sum256([]byte(text))) {
		t.Fatalf("%+v %v", loaded, err)
	}
}
