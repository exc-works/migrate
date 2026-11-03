//go:build full

package migrate

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/exc-works/migrate/internal/dialect"
	"github.com/exc-works/migrate/internal/logger"
	"github.com/exc-works/migrate/internal/source"
)

func TestExperimentalScriptsUseFrameworkTransaction(t *testing.T) {
	for _, d := range []dialect.Dialect{dialect.TiDBDialect{}, dialect.RedshiftDialect{}, dialect.OracleDialect{}} {
		for _, fail := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/fail=%t", d.Name(), fail), func(t *testing.T) {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatal(err)
				}
				defer db.Close()
				placeholder := "?"
				if d.Name() == dialect.NameRedshift {
					placeholder = "$1"
				}
				if d.Name() == dialect.NameOracle {
					placeholder = ":1"
				}
				query := "UPDATE users SET profile=" + placeholder
				body := fmt.Sprintf(`export function up(db){if(db.dialect!==%q)throw Error('dialect');db.exec(%q,['changed']);`, d.Name(), query)
				if fail {
					body += `try{db.query('SELECT broken')}catch(e){}`
				}
				body += `}`
				var logs bytes.Buffer
				svc, err := NewService(context.Background(), Config{DB: db, Dialect: d, Logger: logger.NewStd("warn", &logs), MigrationSource: source.StringSource{Migrations: []source.Migration{{Filename: "V1__experimental.js", Source: body}}}})
				if err != nil {
					t.Fatal(err)
				}
				selectSQL, _ := d.SelectSchemaSQL("migration_schema")
				mock.ExpectQuery(regexp.QuoteMeta(selectSQL)).WillReturnRows(sqlmock.NewRows([]string{"id", "version", "filename", "hash", "status", "created_at"}))
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(query)).WithArgs("changed").WillReturnResult(sqlmock.NewResult(0, 1))
				if fail {
					mock.ExpectQuery("SELECT broken").WillReturnError(fmt.Errorf("private driver detail"))
					mock.ExpectRollback()
				} else {
					insertSQL, _ := d.InsertSchemaSQL("migration_schema")
					mock.ExpectExec(regexp.QuoteMeta(insertSQL)).WillReturnResult(sqlmock.NewResult(0, 1))
					mock.ExpectCommit()
				}
				err = svc.Up()
				if (err != nil) != fail || err != nil && strings.Contains(err.Error(), "private driver detail") {
					t.Fatalf("%v", err)
				}
				if strings.Count(logs.String(), "experimental script dialect=") != 1 {
					t.Fatalf("%s", logs.String())
				}
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}

func TestExperimentalWarningOnlyForSelectedScripts(t *testing.T) {
	var logs bytes.Buffer
	svc := &Service{ctx: context.Background(), cfg: Config{Dialect: dialect.TiDBDialect{}, Logger: logger.NewStd("warn", &logs), DryRun: true}}
	scripts := []Migration{{Filename: "V1__one.js", Source: `throw Error('dry-run must not evaluate');export function up(){}`}, {Filename: "V2__two.ts", Source: `export function up():void{}`}}
	if _, err := svc.preparePlan(scripts, "up"); err != nil {
		t.Fatal(err)
	}
	if strings.Count(logs.String(), "experimental script dialect=") != 1 {
		t.Fatalf("%s", logs.String())
	}
	logs.Reset()
	if _, err := svc.preparePlan([]Migration{{Filename: "V1__only.sql", UpStatements: []string{"SELECT 1"}}}, "up"); err != nil {
		t.Fatal(err)
	}
	if logs.Len() != 0 {
		t.Fatalf("SQL-only warning: %s", logs.String())
	}
}
