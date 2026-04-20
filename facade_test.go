package migrate

import (
	"github.com/exc-works/migrate/internal/dialect"
	"github.com/exc-works/migrate/internal/logger"
	internalmigrate "github.com/exc-works/migrate/internal/migrate"
	"github.com/exc-works/migrate/internal/source"
)

// Compile-time guards that the public facade is alias-identical to the
// internal implementation types. Pointer assignments without conversion
// compile iff the two types are the same (Go type aliases), so if anyone
// converts an alias into a defined wrapper type, these lines stop compiling
// and the break is surfaced by `go test`.
var (
	_ *internalmigrate.Service         = (*Service)(nil)
	_ *internalmigrate.Config          = (*Config)(nil)
	_ *internalmigrate.Migration       = (*Migration)(nil)
	_ *internalmigrate.SchemaRecord    = (*SchemaRecord)(nil)
	_ *internalmigrate.MigrationStatus = (*MigrationStatus)(nil)

	_ *source.Migration       = (*SourceFile)(nil)
	_ *source.DirectorySource = (*DirectorySource)(nil)
	_ *source.StringSource    = (*StringSource)(nil)
	_ *source.FSSource        = (*FSSource)(nil)
	_ *source.CombinedSource  = (*CombinedSource)(nil)

	_ *dialect.PostgresDialect   = (*PostgresDialect)(nil)
	_ *dialect.MySQLDialect      = (*MySQLDialect)(nil)
	_ *dialect.MariaDBDialect    = (*MariaDBDialect)(nil)
	_ *dialect.MSSQLDialect      = (*MSSQLDialect)(nil)
	_ *dialect.OracleDialect     = (*OracleDialect)(nil)
	_ *dialect.ClickHouseDialect = (*ClickHouseDialect)(nil)
	_ *dialect.SQLiteDialect     = (*SQLiteDialect)(nil)
	_ *dialect.TiDBDialect       = (*TiDBDialect)(nil)
	_ *dialect.RedshiftDialect   = (*RedshiftDialect)(nil)

	_ *logger.StdLogger  = (*StdLogger)(nil)
	_ *logger.NoopLogger = (*NoopLogger)(nil)

	_ internalmigrate.Status = Status("")
	_ logger.Level           = LogLevel(0)
)
