package migrate_test

import (
	"context"
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"

	"github.com/exc-works/migrate"
)

// Example demonstrates embedding migrate into another Go service:
// build an in-memory SourceFile list, open a SQL database, construct
// a Service, and drive the Create/Up/Status lifecycle.
func Example() {
	ctx := context.Background()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	src := migrate.StringSource{Migrations: []migrate.SourceFile{
		{
			Filename: "V1__init_users.sql",
			Source: "-- +migrate Up\n" +
				"CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT);\n" +
				"-- +migrate Down\n" +
				"DROP TABLE users;\n",
		},
		{
			Filename: "V2__add_email.sql",
			Source: "-- +migrate Up\n" +
				"ALTER TABLE users ADD COLUMN email TEXT;\n" +
				"-- +migrate Down\n" +
				"ALTER TABLE users DROP COLUMN email;\n",
		},
	}}

	svc, err := migrate.NewService(ctx, migrate.Config{
		Dialect:         migrate.NewSQLiteDialect(),
		DB:              db,
		MigrationSource: src,
		SchemaName:      "migration_schema",
		Logger:          migrate.NoopLogger{},
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

	statuses, err := svc.Status()
	if err != nil {
		panic(err)
	}
	for _, st := range statuses {
		fmt.Printf("%s %s\n", st.Migration.Version, st.Status)
	}

	// Output:
	// 1 applied
	// 2 applied
}
