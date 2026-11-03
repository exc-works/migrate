//go:build full

package migrate_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing/fstest"

	"github.com/exc-works/migrate"
	_ "modernc.org/sqlite"
)

func ExampleWithModules() {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	src := migrate.WithModules{
		Source: migrate.StringSource{Migrations: []migrate.SourceFile{{
			Filename: "V1__shared.ts",
			Source:   `import {answer} from './shared.ts';export function up(db){db.exec('CREATE TABLE result(value TEXT)');db.exec('INSERT INTO result VALUES (?)',[answer])}`,
		}}},
		FS: fstest.MapFS{"shared.ts": {Data: []byte(`export const answer='42';`)}},
	}
	svc, err := migrate.NewService(context.Background(), migrate.Config{
		DB:              db,
		Dialect:         migrate.NewSQLiteDialect(),
		MigrationSource: src,
	})
	if err != nil {
		panic(err)
	}
	if err := svc.Create(); err != nil {
		panic(err)
	}
	if err := svc.Up(); err != nil {
		panic(err)
	}
	var value string
	if err := db.QueryRow("SELECT value FROM result").Scan(&value); err != nil {
		panic(err)
	}
	fmt.Println(value)
	// Output: 42
}
