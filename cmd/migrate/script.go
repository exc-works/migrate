package main

import (
	"encoding/json"
	"fmt"
	"strings"

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
			capabilities := []script.DialectCapability{}
			if len(formats) > 1 {
				build = script.BuildFull
				capabilities = script.Dialects()
				for _, capability := range capabilities {
					dialects = append(dialects, capability.Name)
				}
			}
			switch output {
			case "text":
				_, err := fmt.Fprintf(cmd.OutOrStdout(), "build: %s\nformats: %s\n", build, strings.Join(formats, ", "))
				if err != nil {
					return err
				}
				for _, capability := range capabilities {
					if _, err := fmt.Fprintf(cmd.OutOrStdout(), "script dialect: %s (%s, transaction=%s) %s\n", capability.Name, capability.Verification, capability.TransactionMode, capability.Notes); err != nil {
						return err
					}
				}
				return nil
			case "json":
				return json.NewEncoder(cmd.OutOrStdout()).Encode(struct {
					Version                   string                     `json:"version"`
					Build                     string                     `json:"build"`
					Formats                   []string                   `json:"formats"`
					ScriptDialects            []string                   `json:"script_dialects"`
					ScriptDialectCapabilities []script.DialectCapability `json:"script_dialect_capabilities"`
				}{
					Version:                   effectiveBuildVersion(),
					Build:                     build,
					Formats:                   formats,
					ScriptDialects:            dialects,
					ScriptDialectCapabilities: capabilities,
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
		return `-- +migrate Up

-- +migrate Down
`, nil
	case source.FormatJS, source.FormatMJS:
		if len(script.Formats()) == 1 {
			return "", script.ErrRequiresFull
		}
		return `/**
 * @typedef {string | number | boolean | null | readonly number[] | Uint8Array} SQLParameter
 * @typedef {string | number | boolean | null | number[]} SQLValue
 * @typedef {{
 *   readonly dialect: "postgres" | "mysql" | "mariadb" | "sqlite" | "tidb" | "redshift" | "oracle",
 *   query(sql: string, params?: readonly SQLParameter[]): Record<string, SQLValue>[],
 *   exec(sql: string, params?: readonly SQLParameter[]): { rowsAffected: string }
 * }} MigrationDB
 */

/**
 * @param {MigrationDB} db
 * @returns {void}
 */
export function up(db) {
  throw new Error("Implement this migration");
}

// Export down(db) only when reversal is implemented.
`, nil
	case source.FormatTS:
		if len(script.Formats()) == 1 {
			return "", script.ErrRequiresFull
		}
		return `type SQLParameter = string | number | boolean | null | readonly number[] | Uint8Array;
type SQLValue = string | number | boolean | null | number[];
type MigrationDB = {
  readonly dialect: "postgres" | "mysql" | "mariadb" | "sqlite" | "tidb" | "redshift" | "oracle";
  query(sql: string, params?: readonly SQLParameter[]): Record<string, SQLValue>[];
  exec(sql: string, params?: readonly SQLParameter[]): { rowsAffected: string };
};

export function up(db: MigrationDB): void {
  throw new Error("Implement this migration");
}

// Export down(db) only when reversal is implemented.
`, nil
	default:
		return "", fmt.Errorf("unsupported migration format; use sql, js, ts or mjs")
	}
}
