//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/exc-works/migrate/internal/dialect"
	"github.com/exc-works/migrate/internal/migrate"
	"github.com/exc-works/migrate/internal/source"
)

// Reuse the existing native-driver fixtures; module resolution itself is dialect independent.
func runScriptModuleScenario(t *testing.T, db *sql.DB, d dialect.Dialect) {
	t.Helper()
	switch d.Name() {
	case dialect.NamePostgres, dialect.NameMySQL, dialect.NameMariaDB, dialect.NameSQLite, dialect.NameMSSQL:
	default:
		return
	}
	ctx := context.Background()
	root := t.TempDir()
	table, history := randomName("module_users"), randomName("module_history")
	placeholder := "?"
	if d.Name() == dialect.NamePostgres {
		placeholder = "$1"
	} else if d.Name() == dialect.NameMSSQL {
		placeholder = "@p1"
	}
	update := "UPDATE " + table + " SET profile=" + placeholder
	files := map[string]string{
		"migrations/V1__create.sql":           "-- +migrate Up\nCREATE TABLE " + table + " (profile VARCHAR(80));\nINSERT INTO " + table + " VALUES ('seed');\n-- +migrate Down\nDROP TABLE " + table + ";\n",
		"migrations/V2__import.ts":            `export {up,down} from '../shared/steps.ts';`,
		"shared/steps.ts":                     `import {normalize} from 'normalize';export function up(db){db.exec(` + strconv.Quote(update) + `,[normalize(' IMPORTED ')])}export function down(db){db.exec(` + strconv.Quote(update) + `,['seed'])}`,
		"node_modules/normalize/package.json": `{"main":"index.js"}`,
		"node_modules/normalize/index.js":     `export function normalize(value){return value.trim().toLowerCase()}`,
	}
	write := func(name, text string) {
		t.Helper()
		file := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(file), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file, []byte(text), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	for name, text := range files {
		write(name, text)
	}
	svc, err := migrate.NewService(ctx, migrate.Config{DB: db, Dialect: d, SchemaName: history, MigrationSource: source.DirectorySource{Directory: filepath.Join(root, "migrations")}})
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.Create(); err != nil {
		t.Fatal(err)
	}
	defer db.ExecContext(ctx, "DROP TABLE "+history)
	defer db.ExecContext(ctx, "DROP TABLE "+table)
	if err := svc.Up(); err != nil {
		t.Fatal(err)
	}
	assertScriptProfile(t, db, table, "imported")
	if err := svc.Down("1", false); err != nil {
		t.Fatal(err)
	}
	assertScriptProfile(t, db, table, "seed")
	write("shared/fail.ts", `export function up(db){db.exec(`+strconv.Quote(update)+`,['partial']);throw Error('private imported failure')}`)
	write("migrations/V3__failure.ts", `export {up} from '../shared/fail.ts';`)
	if err := svc.Up(); err == nil {
		t.Fatal("imported migration failure accepted")
	}
	assertScriptProfile(t, db, table, "imported")
	statuses, err := svc.Status()
	if err != nil || len(statuses) != 3 || statuses[2].Status != migrate.StatusPending {
		t.Fatalf("failed imported migration history: %+v %v", statuses, err)
	}
	if err := os.Remove(filepath.Join(root, "migrations", "V3__failure.ts")); err != nil {
		t.Fatal(err)
	}
	if err := svc.Down("", true); err != nil {
		t.Fatal(err)
	}
}
