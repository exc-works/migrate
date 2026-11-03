package dialect

import "strings"

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

// Names lists canonical dialect identifiers, independent of migration format.
func Names() []string {
	return []string{NamePostgres, NameMySQL, NameMariaDB, NameMSSQL, NameOracle, NameClickHouse, NameSQLite, NameTiDB, NameRedshift}
}

// Aliases lists alternate inputs for a canonical name. MySQL, MariaDB and TiDB
// remain distinct dialects even though they share a driver.
func Aliases(canonical string) []string {
	switch canonical {
	case NamePostgres:
		return []string{"pg", "pgsql", "postgresql", DriverPostgres}
	case NameMSSQL:
		return []string{DriverMSSQL, "azuresql"}
	case NameOracle:
		return []string{"orcl", "ora", "go-ora", "goora", "godror"}
	case NameSQLite:
		return []string{"sqlite3"}
	}
	return nil
}

// CanonicalName shares normalization between SQL resolution and script support.
// Empty SQL configuration retains its historical PostgreSQL default.
func CanonicalName(name string) (string, bool) {
	normalized := strings.ToLower(strings.TrimSpace(name))
	if normalized == "" {
		return NamePostgres, true
	}
	for _, canonical := range Names() {
		if normalized == canonical {
			return canonical, true
		}
		for _, alias := range Aliases(canonical) {
			if normalized == alias {
				return canonical, true
			}
		}
	}
	return "", false
}
