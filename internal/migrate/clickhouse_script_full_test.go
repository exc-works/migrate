//go:build full

package migrate

import (
	"bytes"
	"context"
	"errors"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/exc-works/migrate/internal/dialect"
	"github.com/exc-works/migrate/internal/logger"
	"github.com/exc-works/migrate/internal/script"
	"github.com/exc-works/migrate/internal/source"
)

func TestClickHouseScriptFailureBoundary(t *testing.T) {
	for _, failure := range []string{"", "script", "database", "history", "panic"} {
		t.Run(failure, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			d := dialect.ClickHouseDialect{}
			text := `export function up(db){db.exec('INSERT INTO users VALUES (?)',['first']);`
			switch failure {
			case "script":
				text += `throw Error('private detail');`
			case "database":
				text += `try{db.query('SELECT broken')}catch(e){}db.exec('INSERT INTO users VALUES (?)',['must not run']);`
			}
			text += `}`
			var logs bytes.Buffer
			svc, err := NewService(context.Background(), Config{DB: db, Dialect: d, Logger: logger.NewStd("warn", &logs), MigrationSource: source.StringSource{Migrations: []source.Migration{{Filename: "V1__clickhouse.js", Source: text}}}})
			if err != nil {
				t.Fatal(err)
			}
			selectSQL, _ := d.SelectSchemaSQL("migration_schema")
			mock.ExpectQuery(regexp.QuoteMeta(selectSQL)).WillReturnRows(sqlmock.NewRows([]string{"id", "version", "filename", "hash", "status", "created_at"}))
			mock.ExpectExec(regexp.QuoteMeta("INSERT INTO users VALUES (?)")).WithArgs("first").WillReturnResult(sqlmock.NewResult(0, 0))
			if failure == "database" {
				mock.ExpectQuery("SELECT broken").WillReturnError(errors.New("private detail"))
			}
			if failure == "" || failure == "history" || failure == "panic" {
				insertSQL, _ := d.InsertSchemaSQL("migration_schema")
				expect := mock.ExpectExec(regexp.QuoteMeta(insertSQL))
				if failure == "history" {
					expect.WillReturnError(errors.New("history failed"))
				} else {
					expect.WillReturnResult(sqlmock.NewResult(0, 1))
				}
			}
			// Unexpected native panics must receive the same partial-state error.
			if failure == "panic" {
				err = svc.withStepTransaction(scriptStep{}, func(execContextRunner) error { panic("private detail") })
				if err == nil || !strings.Contains(err.Error(), "partial changes") || strings.Contains(err.Error(), "private detail") {
					t.Fatalf("%v", err)
				}
			}
			err = svc.Up()
			bad := failure != "" && failure != "panic"
			if (err != nil) != bad {
				t.Fatalf("%v", err)
			}
			if bad && (!strings.Contains(err.Error(), "partial changes") || strings.Contains(err.Error(), "private detail")) {
				t.Fatalf("%v", err)
			}
			if !strings.Contains(logs.String(), "no migration transaction") {
				t.Fatalf("%s", logs.String())
			}
			if err = mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestClickHouseDownHistoryFailureIsExplicit(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	svc, err := NewService(context.Background(), Config{DB: db, Dialect: dialect.ClickHouseDialect{}, MigrationSource: source.StringSource{}})
	if err != nil {
		t.Fatal(err)
	}
	p, err := script.Compile("V1__clickhouse.js", `export function down(db){db.exec('DELETE FROM users')}`, "down", script.Options{})
	if err != nil {
		t.Fatal(err)
	}
	i, err := p.Instantiate(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectExec("DELETE FROM users").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("DELETE FROM migration_schema").WithArgs("V1__clickhouse.js").WillReturnError(errors.New("history failed"))
	err = svc.applyDown(Migration{Filename: "V1__clickhouse.js"}, scriptStep{dialect: "clickhouse", invocation: i}, "DELETE FROM migration_schema WHERE filename=?")
	if err == nil || !strings.Contains(err.Error(), "partial changes") {
		t.Fatalf("%v", err)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
