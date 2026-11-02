package dialect

// Canonical dialect identifiers; input aliases remain accepted by FromName.
const (
	NamePostgres   = "postgres"
	NameMySQL      = "mysql"
	NameMariaDB    = "mariadb"
	NameMSSQL      = "mssql"
	NameOracle     = "oracle"
	NameClickHouse = "clickhouse"
	NameSQLite     = "sqlite"
	NameTiDB       = "tidb"
	NameRedshift   = "redshift"
)

// Driver identifiers are separate from dialect names (for example, PostgreSQL uses pgx).
const (
	DriverPostgres   = "pgx"
	DriverMySQL      = "mysql"
	DriverMSSQL      = "sqlserver"
	DriverOracle     = "oracle"
	DriverClickHouse = "clickhouse"
	DriverSQLite     = "sqlite"
)
