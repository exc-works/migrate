//go:build integration

package integrationtest

import (
	"slices"
	"testing"
)

func TestIntegrationSelectionCanonicalizesAliases(t *testing.T) {
	t.Setenv(envIntegrationDB, " SQLSERVER,azuresql,mssql,PgSQL,postgresql,pg,pgx,MySQL,mariadb,sqlite3 ")
	got := selectedFlavors()
	want := []string{"mssql", "postgres", "mysql", "mariadb", "sqlite"}
	if !slices.Equal(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
	for _, name := range want {
		if !isFlavorExplicitlySelected(name) {
			t.Fatalf("%s not explicitly selected", name)
		}
	}
	if isFlavorExplicitlySelected("oracle") {
		t.Fatal("unselected Oracle treated as explicit")
	}
	t.Setenv(envIntegrationDB, "")
	if isFlavorExplicitlySelected("mssql") {
		t.Fatal("default optional MSSQL selection became explicit")
	}
}
