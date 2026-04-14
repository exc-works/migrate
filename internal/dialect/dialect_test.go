package dialect

import (
	"strings"
	"testing"
)

func TestFromName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		input      string
		wantName   string
		wantDriver string
	}{
		{
			name:       "default postgres",
			input:      "",
			wantName:   "postgres",
			wantDriver: "pgx",
		},
		{
			name:       "postgres alias",
			input:      "PgSQL",
			wantName:   "postgres",
			wantDriver: "pgx",
		},
		{
			name:       "postgresql alias",
			input:      "postgresql",
			wantName:   "postgres",
			wantDriver: "pgx",
		},
		{
			name:       "mysql",
			input:      "mysql",
			wantName:   "mysql",
			wantDriver: "mysql",
		},
		{
			name:       "mariadb",
			input:      "mariadb",
			wantName:   "mariadb",
			wantDriver: "mysql",
		},
		{
			name:       "oracle",
			input:      "oracle",
			wantName:   "oracle",
			wantDriver: "oracle",
		},
		{
			name:       "oracle alias with spaces",
			input:      " ORCL ",
			wantName:   "oracle",
			wantDriver: "oracle",
		},
		{
			name:       "oracle godror alias",
			input:      "godror",
			wantName:   "oracle",
			wantDriver: "oracle",
		},
		{
			name:       "sqlite",
			input:      "sqlite",
			wantName:   "sqlite",
			wantDriver: "sqlite",
		},
		{
			name:       "sqlite alias",
			input:      "sqlite3",
			wantName:   "sqlite",
			wantDriver: "sqlite",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			d, err := FromName(tc.input)
			if err != nil {
				t.Fatalf("FromName(%q) error: %v", tc.input, err)
			}
			if got := d.Name(); got != tc.wantName {
				t.Fatalf("Name() = %q, want %q", got, tc.wantName)
			}
			if got := d.DriverName(); got != tc.wantDriver {
				t.Fatalf("DriverName() = %q, want %q", got, tc.wantDriver)
			}
		})
	}
}

func TestFromNameUnsupported(t *testing.T) {
	t.Parallel()

	_, err := FromName("sqlserver")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "unsupported dialect") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPostgresAndMySQLPlaceholdersUnchanged(t *testing.T) {
	t.Parallel()

	postgresInsert, err := PostgresDialect{}.InsertSchemaSQL("migration_schema")
	if err != nil {
		t.Fatalf("postgres insert sql error: %v", err)
	}
	if !strings.Contains(postgresInsert, "VALUES ($1, $2, $3, $4, $5, $6)") {
		t.Fatalf("unexpected postgres insert sql: %s", postgresInsert)
	}

	mysqlInsert, err := MySQLDialect{}.InsertSchemaSQL("migration_schema")
	if err != nil {
		t.Fatalf("mysql insert sql error: %v", err)
	}
	if !strings.Contains(mysqlInsert, "VALUES (?, ?, ?, ?, ?, ?)") {
		t.Fatalf("unexpected mysql insert sql: %s", mysqlInsert)
	}

	mariadbInsert, err := MariaDBDialect{}.InsertSchemaSQL("migration_schema")
	if err != nil {
		t.Fatalf("mariadb insert sql error: %v", err)
	}
	if mariadbInsert != mysqlInsert {
		t.Fatalf("mariadb insert should match mysql insert, got %q vs %q", mariadbInsert, mysqlInsert)
	}
}

func TestOracleDialectSQL(t *testing.T) {
	t.Parallel()

	const schema = "migration_schema"
	d := OracleDialect{}

	createSQL, err := d.CreateSchemaSQL(schema)
	if err != nil {
		t.Fatalf("create sql error: %v", err)
	}
	if !strings.Contains(createSQL, "EXECUTE IMMEDIATE 'CREATE TABLE migration_schema") {
		t.Fatalf("unexpected oracle create sql: %s", createSQL)
	}
	if !strings.Contains(createSQL, "SQLCODE != -955") {
		t.Fatalf("oracle create sql should handle existing table, got: %s", createSQL)
	}

	insertSQL, err := d.InsertSchemaSQL(schema)
	if err != nil {
		t.Fatalf("insert sql error: %v", err)
	}
	if !strings.Contains(insertSQL, "VALUES (:1, :2, :3, :4, :5, :6)") {
		t.Fatalf("unexpected oracle insert sql: %s", insertSQL)
	}

	deleteSQL, err := d.DeleteSchemaSQL(schema)
	if err != nil {
		t.Fatalf("delete sql error: %v", err)
	}
	if deleteSQL != "DELETE FROM migration_schema WHERE filename = :1" {
		t.Fatalf("unexpected oracle delete sql: %s", deleteSQL)
	}

	selectSQL, err := d.SelectSchemaSQL(schema)
	if err != nil {
		t.Fatalf("select sql error: %v", err)
	}
	if selectSQL != "SELECT id, version, filename, hash, status, created_at FROM migration_schema ORDER BY id ASC" {
		t.Fatalf("unexpected oracle select sql: %s", selectSQL)
	}
}

func TestSQLiteDialectSQL(t *testing.T) {
	t.Parallel()

	const schema = "migration_schema"
	d := SQLiteDialect{}

	createSQL, err := d.CreateSchemaSQL(schema)
	if err != nil {
		t.Fatalf("create sql error: %v", err)
	}
	if !strings.Contains(createSQL, "AUTOINCREMENT") {
		t.Fatalf("sqlite create sql should use AUTOINCREMENT, got: %s", createSQL)
	}

	insertSQL, err := d.InsertSchemaSQL(schema)
	if err != nil {
		t.Fatalf("insert sql error: %v", err)
	}
	if insertSQL != "INSERT INTO migration_schema (id, version, filename, hash, status, created_at) VALUES (?, ?, ?, ?, ?, ?)" {
		t.Fatalf("unexpected sqlite insert sql: %s", insertSQL)
	}

	deleteSQL, err := d.DeleteSchemaSQL(schema)
	if err != nil {
		t.Fatalf("delete sql error: %v", err)
	}
	if deleteSQL != "DELETE FROM migration_schema WHERE filename = ?" {
		t.Fatalf("unexpected sqlite delete sql: %s", deleteSQL)
	}

	selectSQL, err := d.SelectSchemaSQL(schema)
	if err != nil {
		t.Fatalf("select sql error: %v", err)
	}
	if selectSQL != "SELECT id, version, filename, hash, status, created_at FROM migration_schema ORDER BY id ASC" {
		t.Fatalf("unexpected sqlite select sql: %s", selectSQL)
	}
}

func TestDialectIdentifierValidation(t *testing.T) {
	t.Parallel()

	dialects := []Dialect{
		PostgresDialect{},
		MySQLDialect{},
		MariaDBDialect{},
		OracleDialect{},
		SQLiteDialect{},
	}
	for _, d := range dialects {
		_, err := d.CreateSchemaSQL("invalid-name")
		if err == nil {
			t.Fatalf("%T should reject invalid identifier", d)
		}
	}
}
