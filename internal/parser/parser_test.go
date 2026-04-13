package parser

import (
	"strings"
	"testing"
)

func TestSplitSQLStatementsUpDown(t *testing.T) {
	sqlText := `-- +migrate Up
CREATE TABLE demo (id INT);
INSERT INTO demo (id) VALUES (1);

-- +migrate Down
DELETE FROM demo WHERE id = 1;
DROP TABLE demo;`

	up, err := SplitSQLStatements(strings.NewReader(sqlText), DirectionUp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(up) != 2 {
		t.Fatalf("expected 2 up statements, got %d", len(up))
	}

	down, err := SplitSQLStatements(strings.NewReader(sqlText), DirectionDown)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(down) != 2 {
		t.Fatalf("expected 2 down statements, got %d", len(down))
	}
}

func TestSplitSQLStatementsNoAnnotation(t *testing.T) {
	_, err := SplitSQLStatements(strings.NewReader("SELECT 1;"), DirectionUp)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestSplitSQLStatementsStatementBlock(t *testing.T) {
	sqlText := `-- +migrate Up
-- +migrate StatementBegin
CREATE FUNCTION x() RETURNS VOID AS $$
BEGIN
  PERFORM 1;
END;
$$ LANGUAGE plpgsql;
-- +migrate StatementEnd

-- +migrate Down
SELECT 1;`
	up, err := SplitSQLStatements(strings.NewReader(sqlText), DirectionUp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(up) != 1 {
		t.Fatalf("expected 1 up statement, got %d", len(up))
	}
}
