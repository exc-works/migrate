//go:build !full

package migrate

import (
	"errors"
	"testing"

	"github.com/exc-works/migrate/internal/script"
	"github.com/exc-works/migrate/internal/source"
)

func TestSQLOnlyRefusesScriptsBeforeDatabaseWrites(t *testing.T) {
	for _, extension := range []string{"js", "ts", "mjs"} {
		t.Run(extension, func(t *testing.T) { testSQLOnlyRefusesScript(t, extension) })
	}
}

func testSQLOnlyRefusesScript(t *testing.T, extension string) {
	t.Helper()
	for _, operation := range []string{"up", "down", "baseline"} {
		t.Run(operation, func(t *testing.T) {
			src := []source.Migration{{Filename: "V1__sql.sql", Source: migrationSQL("SELECT 1;", "SELECT 1;")}, {Filename: "V2__script." + extension, Source: `export function up(){}`}}
			svc, mock, cleanup := newServiceWithMock(t, src, false, false, nil)
			defer cleanup()
			var err error
			switch operation {
			case "up":
				err = svc.Up()
			case "down":
				err = svc.Down("", true)
			case "baseline":
				err = svc.Baseline()
			}
			if !errors.Is(err, script.ErrRequiresFull) {
				t.Fatalf("%v", err)
			}
			if err = mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
