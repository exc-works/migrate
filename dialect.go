package migrate

import (
	"github.com/exc-works/migrate/internal/dialect"
)

// Dialect abstracts the per-database SQL needed to manage the history table.
type Dialect = dialect.Dialect

// PostgresDialect targets PostgreSQL via the pgx driver.
type PostgresDialect = dialect.PostgresDialect

// MySQLDialect targets MySQL via the go-sql-driver/mysql driver.
type MySQLDialect = dialect.MySQLDialect

// MariaDBDialect targets MariaDB via the MySQL driver.
type MariaDBDialect = dialect.MariaDBDialect

// MSSQLDialect targets SQL Server / Azure SQL via the microsoft/go-mssqldb driver.
type MSSQLDialect = dialect.MSSQLDialect

// OracleDialect targets Oracle via the sijms/go-ora driver.
type OracleDialect = dialect.OracleDialect

// ClickHouseDialect targets ClickHouse via the clickhouse-go driver.
type ClickHouseDialect = dialect.ClickHouseDialect

// SQLiteDialect targets SQLite via the modernc.org/sqlite driver.
type SQLiteDialect = dialect.SQLiteDialect

// TiDBDialect targets TiDB via the MySQL driver.
type TiDBDialect = dialect.TiDBDialect

// RedshiftDialect targets Amazon Redshift via the pgx driver.
type RedshiftDialect = dialect.RedshiftDialect

// Dialect constructors. Prefer these over zero-value struct literals
// (e.g. migrate.NewPostgresDialect() rather than migrate.PostgresDialect{}):
// they return the Dialect interface, so the facade can grow options or
// swap concrete implementations without breaking callers.

// NewPostgresDialect returns a Dialect targeting PostgreSQL.
func NewPostgresDialect() Dialect { return dialect.PostgresDialect{} }

// NewMySQLDialect returns a Dialect targeting MySQL.
func NewMySQLDialect() Dialect { return dialect.MySQLDialect{} }

// NewMariaDBDialect returns a Dialect targeting MariaDB.
func NewMariaDBDialect() Dialect { return dialect.MariaDBDialect{} }

// NewMSSQLDialect returns a Dialect targeting SQL Server / Azure SQL.
func NewMSSQLDialect() Dialect { return dialect.MSSQLDialect{} }

// NewOracleDialect returns a Dialect targeting Oracle.
func NewOracleDialect() Dialect { return dialect.OracleDialect{} }

// NewClickHouseDialect returns a Dialect targeting ClickHouse.
func NewClickHouseDialect() Dialect { return dialect.ClickHouseDialect{} }

// NewSQLiteDialect returns a Dialect targeting SQLite.
func NewSQLiteDialect() Dialect { return dialect.SQLiteDialect{} }

// NewTiDBDialect returns a Dialect targeting TiDB.
func NewTiDBDialect() Dialect { return dialect.TiDBDialect{} }

// NewRedshiftDialect returns a Dialect targeting Amazon Redshift.
func NewRedshiftDialect() Dialect { return dialect.RedshiftDialect{} }

// DialectFromName resolves a canonical or aliased dialect name to a Dialect implementation.
func DialectFromName(name string) (Dialect, error) {
	return dialect.FromName(name)
}
