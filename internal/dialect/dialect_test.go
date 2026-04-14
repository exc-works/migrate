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
			name:       "mssql",
			input:      "mssql",
			wantName:   "mssql",
			wantDriver: "sqlserver",
		},
		{
			name:       "mssql alias",
			input:      "sqlserver",
			wantName:   "mssql",
			wantDriver: "sqlserver",
		},
		{
			name:       "clickhouse",
			input:      "clickhouse",
			wantName:   "clickhouse",
			wantDriver: "clickhouse",
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
		{
			name:       "tidb",
			input:      "tidb",
			wantName:   "tidb",
			wantDriver: "mysql",
		},
		{
			name:       "redshift",
			input:      "redshift",
			wantName:   "redshift",
			wantDriver: "pgx",
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

	_, err := FromName("db2")
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

func TestMSSQLDialectSQL(t *testing.T) {
	t.Parallel()

	const schema = "migration_schema"
	d := MSSQLDialect{}

	createSQL, err := d.CreateSchemaSQL(schema)
	if err != nil {
		t.Fatalf("create sql error: %v", err)
	}
	if !strings.Contains(createSQL, "OBJECT_ID") || !strings.Contains(createSQL, "SYSUTCDATETIME()") {
		t.Fatalf("unexpected mssql create sql: %s", createSQL)
	}

	insertSQL, err := d.InsertSchemaSQL(schema)
	if err != nil {
		t.Fatalf("insert sql error: %v", err)
	}
	if !strings.Contains(insertSQL, "VALUES (@p1, @p2, @p3, @p4, @p5, @p6)") {
		t.Fatalf("unexpected mssql insert sql: %s", insertSQL)
	}

	deleteSQL, err := d.DeleteSchemaSQL(schema)
	if err != nil {
		t.Fatalf("delete sql error: %v", err)
	}
	if deleteSQL != "DELETE FROM migration_schema WHERE filename = @p1" {
		t.Fatalf("unexpected mssql delete sql: %s", deleteSQL)
	}
}

func TestClickHouseDialectSQL(t *testing.T) {
	t.Parallel()

	const schema = "migration_schema"
	d := ClickHouseDialect{}

	createSQL, err := d.CreateSchemaSQL(schema)
	if err != nil {
		t.Fatalf("create sql error: %v", err)
	}
	if !strings.Contains(createSQL, "ENGINE = MergeTree()") || !strings.Contains(createSQL, "DateTime64") {
		t.Fatalf("unexpected clickhouse create sql: %s", createSQL)
	}

	insertSQL, err := d.InsertSchemaSQL(schema)
	if err != nil {
		t.Fatalf("insert sql error: %v", err)
	}
	if !strings.Contains(insertSQL, "VALUES (?, ?, ?, ?, ?, ?)") {
		t.Fatalf("unexpected clickhouse insert sql: %s", insertSQL)
	}

	deleteSQL, err := d.DeleteSchemaSQL(schema)
	if err != nil {
		t.Fatalf("delete sql error: %v", err)
	}
	if deleteSQL != "DELETE FROM migration_schema WHERE filename = ? SETTINGS mutations_sync = 2" {
		t.Fatalf("unexpected clickhouse delete sql: %s", deleteSQL)
	}
}

func TestTiDBAndRedshiftDialectSQL(t *testing.T) {
	t.Parallel()

	tidbInsert, err := TiDBDialect{}.InsertSchemaSQL("migration_schema")
	if err != nil {
		t.Fatalf("tidb insert sql error: %v", err)
	}
	mysqlInsert, err := MySQLDialect{}.InsertSchemaSQL("migration_schema")
	if err != nil {
		t.Fatalf("mysql insert sql error: %v", err)
	}
	if tidbInsert != mysqlInsert {
		t.Fatalf("tidb insert should match mysql insert, got %q vs %q", tidbInsert, mysqlInsert)
	}

	redshiftCreate, err := RedshiftDialect{}.CreateSchemaSQL("migration_schema")
	if err != nil {
		t.Fatalf("redshift create sql error: %v", err)
	}
	if !strings.Contains(redshiftCreate, "DEFAULT GETDATE()") {
		t.Fatalf("unexpected redshift create sql: %s", redshiftCreate)
	}

	redshiftInsert, err := RedshiftDialect{}.InsertSchemaSQL("migration_schema")
	if err != nil {
		t.Fatalf("redshift insert sql error: %v", err)
	}
	if !strings.Contains(redshiftInsert, "VALUES ($1, $2, $3, $4, $5, $6)") {
		t.Fatalf("unexpected redshift insert sql: %s", redshiftInsert)
	}
}

func TestDialectIdentifierValidation(t *testing.T) {
	t.Parallel()

	dialects := []Dialect{
		PostgresDialect{},
		MySQLDialect{},
		MariaDBDialect{},
		MSSQLDialect{},
		OracleDialect{},
		ClickHouseDialect{},
		SQLiteDialect{},
		TiDBDialect{},
		RedshiftDialect{},
	}
	for _, d := range dialects {
		_, err := d.CreateSchemaSQL("invalid-name")
		if err == nil {
			t.Fatalf("%T should reject invalid identifier", d)
		}
	}
}
