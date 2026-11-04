package main

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"os"

	"github.com/exc-works/migrate"
	_ "modernc.org/sqlite"
)

//go:embed migrations shared
var files embed.FS

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return err
	}
	defer db.Close()
	// SQLite's in-memory database belongs to a connection.
	db.SetMaxOpenConns(1)
	svc, err := migrate.NewService(ctx, migrate.Config{
		DB:      db,
		Dialect: migrate.NewSQLiteDialect(),
		MigrationSource: migrate.FSSource{
			FS:   files,
			Root: "migrations",
		},
	})
	if err != nil {
		return err
	}
	if err := svc.Create(); err != nil {
		return err
	}
	if err := svc.Up(); err != nil {
		return err
	}
	var name string
	if err := db.QueryRowContext(ctx, "SELECT display_name FROM users WHERE id=1").Scan(&name); err != nil {
		return err
	}
	if name != "alice" {
		return fmt.Errorf("unexpected normalized name")
	}
	fmt.Printf("after up: %s\n", name)
	if err := svc.Down("", true); err != nil {
		return err
	}
	statuses, err := svc.Status()
	if err != nil {
		return err
	}
	for _, status := range statuses {
		if status.Status != migrate.StatusPending {
			return fmt.Errorf("rollback left an applied migration")
		}
	}
	fmt.Println("after down: all migrations pending")
	return nil
}
