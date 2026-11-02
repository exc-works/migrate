package script

import "testing"

func TestDialectCapabilities(t *testing.T) {
	for _, name := range []string{"postgres", "mysql", "mariadb", "sqlite"} {
		c, ok := Capability(name)
		if !ok || c.Verification != VerificationVerified || c.TransactionMode != TransactionFramework {
			t.Fatalf("%s: %+v", name, c)
		}
	}
	for _, name := range []string{"tidb", "redshift"} {
		c, ok := Capability(name)
		if !ok || !SupportedDialect(name) || c.Verification != VerificationExperimental || c.Notes == "" {
			t.Fatalf("%s: %+v", name, c)
		}
	}
	for _, name := range []string{"", "unknown", "mssql"} {
		if SupportedDialect(name) {
			t.Fatalf("accepted %q", name)
		}
	}
	list := Dialects()
	list[0].Name = "changed"
	if !SupportedDialect("postgres") {
		t.Fatal("capabilities shared mutable state")
	}
}
