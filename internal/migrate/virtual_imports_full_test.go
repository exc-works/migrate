//go:build full

package migrate

import (
	"testing"
	"testing/fstest"

	"github.com/exc-works/migrate/internal/source"
)

func TestCombinedVirtualSourcesKeepIndependentHelpers(t *testing.T) {
	first := fstest.MapFS{
		"migrations/V1__create.ts": {Data: []byte(`import {value} from '../shared/helper.ts';export function up(db){db.exec('CREATE TABLE users(id INTEGER,profile TEXT)');db.exec('INSERT INTO users VALUES (1,?)',[value])}export function down(db){db.exec('DROP TABLE users')}`)},
		"shared/helper.ts":         {Data: []byte(`export const value='first';`)},
	}
	second := fstest.MapFS{"shared/helper.ts": {Data: []byte(`export const value='second';`)}}
	memory := source.StringSource{Migrations: []source.Migration{{Filename: "V2__insert.ts", Source: `import {value} from '../shared/helper.ts';export function up(db){db.exec('INSERT INTO users VALUES (2,?)',[value])}export function down(db){db.exec('DELETE FROM users WHERE id=2')}`}}}
	svc, db, _ := mixedService(t, nil, false)
	svc.cfg.MigrationSource = source.CombinedSource{Sources: []source.MigrationSource{
		source.FSSource{FS: first, Root: "migrations"},
		source.WithModules{Source: memory, FS: second, Root: "migrations"},
	}}
	if err := svc.Up(); err != nil {
		t.Fatal(err)
	}
	var values string
	if err := db.QueryRow("SELECT group_concat(profile,',') FROM (SELECT profile FROM users ORDER BY id)").Scan(&values); err != nil || values != "first,second" {
		t.Fatalf("source contexts crossed: %s %v", values, err)
	}
	if err := svc.Down("", true); err != nil {
		t.Fatal(err)
	}
}

func TestVirtualSourceWithExplicitPackageDirectory(t *testing.T) {
	root := t.TempDir()
	writeMigrationFiles(t, root, map[string]string{
		"node_modules/helper/package.json": `{"main":"index.js"}`,
		"node_modules/helper/index.js":     `export const value='installed';`,
	})
	files := fstest.MapFS{
		"V1__virtual.js": {Data: []byte(`import {value} from './shared.js';export function up(db){db.exec('CREATE TABLE users(profile TEXT)');db.exec('INSERT INTO users VALUES (?)',[value])}`)},
		"shared.js":      {Data: []byte(`export {value} from 'helper';`)},
	}
	svc, db, _ := mixedService(t, nil, false)
	svc.cfg.MigrationSource = source.WithModules{Source: source.FSSource{FS: files}, DependencyDir: root}
	if err := svc.Up(); err != nil {
		t.Fatal(err)
	}
	var value string
	if err := db.QueryRow("SELECT profile FROM users").Scan(&value); err != nil || value != "installed" {
		t.Fatalf("virtual package lookup: %s %v", value, err)
	}
}
