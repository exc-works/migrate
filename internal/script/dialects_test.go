package script

import (
	"testing"

	"github.com/exc-works/migrate/internal/dialect"
)

func TestDialectCapabilities(t *testing.T) {
	for _, name := range []string{"postgres", "mysql", "mariadb", "sqlite"} {
		c, ok := Capability(name)
		if !ok || c.Verification != VerificationVerified || c.TransactionMode != TransactionFramework {
			t.Fatalf("%s: %+v", name, c)
		}
	}
	for _, name := range []string{"tidb", "redshift", "oracle", "clickhouse", "mssql"} {
		c, ok := Capability(name)
		if !ok || !SupportedDialect(name) || c.Verification != VerificationExperimental || c.Notes == "" {
			t.Fatalf("%s: %+v", name, c)
		}
	}
	for _, name := range []string{"", " ", "unknown"} {
		if SupportedDialect(name) {
			t.Fatalf("accepted %q", name)
		}
	}
	c, _ := Capability("clickhouse")
	if c.TransactionMode != TransactionNone {
		t.Fatalf("%+v", c)
	}
	list := Dialects()
	list[0].Name = "changed"
	if !SupportedDialect("postgres") {
		t.Fatal("capabilities shared mutable state")
	}
}

func TestCapabilitiesCoverSQLDialectsAndAliases(t *testing.T) {
	if len(Dialects()) != len(dialect.Names()) {
		t.Fatal("SQL and script dialect coverage differ")
	}
	for _, canonical := range dialect.Names() {
		for _, alias := range append([]string{canonical}, dialect.Aliases(canonical)...) {
			c, ok := Capability(alias)
			if !ok || !SupportedDialect(alias) || c.Name != canonical {
				t.Fatalf("%s: %+v", alias, c)
			}
		}
	}
	for _, input := range []string{" PGSQL ", " SQLSERVER ", " AZURESQL ", " MySQL "} {
		if !SupportedDialect(input) {
			t.Fatalf("rejected %q", input)
		}
	}
	c, _ := Capability("pgsql")
	c.Aliases[0] = "changed"
	again, _ := Capability("postgres")
	if again.Aliases[0] == "changed" {
		t.Fatal("shared mutable capability aliases")
	}
}
