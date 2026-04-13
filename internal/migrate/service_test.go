package migrate

import "testing"

func TestBuildStatusesOutOfOrder(t *testing.T) {
	migrations := []Migration{
		{Version: "1", Filename: "V1__a.sql", Hash: "h1"},
		{Version: "3", Filename: "V3__b.sql", Hash: "h3"},
	}
	schemas := []SchemaRecord{{ID: 1, Version: "3", Filename: "V3__b.sql", Hash: "h3", Status: StatusApplied}}
	statuses, _ := buildStatuses(migrations, schemas)
	if len(statuses) != 2 {
		t.Fatalf("unexpected status count: %d", len(statuses))
	}
	if statuses[0].Status != StatusOutOfOrder {
		t.Fatalf("expected outOfOrder, got %s", statuses[0].Status)
	}
	if statuses[1].Status != StatusApplied {
		t.Fatalf("expected applied, got %s", statuses[1].Status)
	}
}

func TestValidateStatuses(t *testing.T) {
	statuses := []MigrationStatus{{Migration: Migration{Filename: "V1__a.sql", Version: "1"}, Status: StatusHashMismatch}}
	if err := validateStatuses(statuses); err == nil {
		t.Fatalf("expected error")
	}
}
