package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/exc-works/migrate/internal/dialect"
	"github.com/exc-works/migrate/internal/script"
	"github.com/exc-works/migrate/internal/source"
	"github.com/spf13/cobra"
)

func newCapabilitiesCommand() *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "capabilities",
		Short: "Show compiled migration formats without connecting to a database",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			formats := script.Formats()
			build := script.BuildSQL
			dialects := []string{}
			if len(formats) > 1 {
				build = script.BuildFull
				dialects = []string{dialect.NamePostgres, dialect.NameMySQL, dialect.NameMariaDB, dialect.NameSQLite}
			}
			switch output {
			case "text":
				_, err := fmt.Fprintf(cmd.OutOrStdout(), "build: %s\nformats: %s\n", build, strings.Join(formats, ", "))
				return err
			case "json":
				return json.NewEncoder(cmd.OutOrStdout()).Encode(struct {
					Version        string   `json:"version"`
					Build          string   `json:"build"`
					Formats        []string `json:"formats"`
					ScriptDialects []string `json:"script_dialects"`
				}{
					Version:        effectiveBuildVersion(),
					Build:          build,
					Formats:        formats,
					ScriptDialects: dialects,
				})
			default:
				return fmt.Errorf("unsupported capabilities output; use text or json")
			}
		},
	}
	cmd.Flags().StringVarP(&output, "output", "o", "text", "output format: text|json")
	return cmd
}

func migrationTemplate(format string) (string, error) {
	switch format {
	case source.FormatSQL:
		return "-- +migrate Up\n\n-- +migrate Down\n", nil
	case source.FormatJS, source.FormatTS:
		if len(script.Formats()) == 1 {
			return "", script.ErrRequiresFull
		}
	default:
		return "", fmt.Errorf("unsupported migration format; use sql, js or ts")
	}
	declaration := ""
	parameter := "db"
	if format == source.FormatTS {
		declaration = `type SQLValue = string | number | boolean | null | number[];
type MigrationDB = {
  readonly dialect: string;
  query(sql: string, params?: SQLValue[]): Record<string, SQLValue>[];
  exec(sql: string, params?: SQLValue[]): { rowsAffected: string };
};

`
		parameter = "db: MigrationDB"
	}
	return declaration + "export function up(" + parameter + ") {\n  throw new Error(\"Implement this migration\");\n}\n\n// Export down(db) only when reversal is implemented.\n", nil
}
