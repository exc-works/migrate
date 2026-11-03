//go:build full

package migrate

import (
	"testing"

	"github.com/exc-works/migrate/internal/source"
)

func TestBaselineAndOutOfOrderDoNotResolveSkippedImports(t *testing.T) {
	for _, baseline := range []bool{true, false} {
		root := t.TempDir()
		writeMigrationFiles(t, root, map[string]string{"V2__standalone.js": `export function up(db){db.exec('CREATE TABLE users(profile TEXT)')}`})
		svc, _, _ := mixedService(t, nil, false)
		svc.cfg.MigrationSource = source.DirectorySource{Directory: root}
		if baseline {
			writeMigrationFiles(t, root, map[string]string{"V1__missing.ts": `export {up} from './missing.ts';`})
			if err := svc.Baseline(); err != nil {
				t.Fatal(err)
			}
		} else {
			if err := svc.Up(); err != nil {
				t.Fatal(err)
			}
			writeMigrationFiles(t, root, map[string]string{"V1__missing.ts": `export {up} from './missing.ts';`})
		}
		if err := svc.Up(); err != nil {
			t.Fatalf("resolved skipped import: %v", err)
		}
		if !baseline {
			svc.cfg.MigrateOutOfOrder = true
			if err := svc.Up(); err == nil {
				t.Fatal("selected out-of-order migration bypassed dependency validation")
			}
		}
	}
}

func TestImportedEntryChangesStillFailHistoryValidation(t *testing.T) {
	root := t.TempDir()
	entry := `export {up} from './helper.ts';`
	writeMigrationFiles(t, root, map[string]string{
		"V1__entry.ts": entry,
		"helper.ts":    `export function up(db){db.exec('CREATE TABLE users(id INTEGER)')}`,
	})
	svc, _, _ := mixedService(t, nil, false)
	svc.cfg.MigrationSource = source.DirectorySource{Directory: root}
	if err := svc.Up(); err != nil {
		t.Fatal(err)
	}
	writeMigrationFiles(t, root, map[string]string{"V1__entry.ts": entry + "\n// changed entry"})
	if err := svc.Up(); err == nil {
		t.Fatal("entry mutation bypassed history checksum validation")
	}
}
