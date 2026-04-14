package dialect

import (
	"fmt"
	"regexp"
	"strings"
)

var identifierPattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

type Dialect interface {
	Name() string
	DriverName() string
	CreateSchemaSQL(schemaName string) (string, error)
	InsertSchemaSQL(schemaName string) (string, error)
	DeleteSchemaSQL(schemaName string) (string, error)
	SelectSchemaSQL(schemaName string) (string, error)
}

func ValidateIdentifier(name string) error {
	if !identifierPattern.MatchString(name) {
		return fmt.Errorf("invalid identifier: %q", name)
	}
	return nil
}

type PostgresDialect struct{}

func (PostgresDialect) Name() string {
	return "postgres"
}

func (PostgresDialect) DriverName() string {
	return "pgx"
}

func (PostgresDialect) CreateSchemaSQL(schemaName string) (string, error) {
	if err := ValidateIdentifier(schemaName); err != nil {
		return "", err
	}
	return strings.TrimSpace(fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s (
    id BIGSERIAL PRIMARY KEY,
    version TEXT NOT NULL,
    filename TEXT NOT NULL UNIQUE,
    hash TEXT NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
`, schemaName)), nil
}

func (PostgresDialect) InsertSchemaSQL(schemaName string) (string, error) {
	if err := ValidateIdentifier(schemaName); err != nil {
		return "", err
	}
	return fmt.Sprintf(`INSERT INTO %s (id, version, filename, hash, status, created_at) VALUES ($1, $2, $3, $4, $5, $6)`, schemaName), nil
}

func (PostgresDialect) DeleteSchemaSQL(schemaName string) (string, error) {
	if err := ValidateIdentifier(schemaName); err != nil {
		return "", err
	}
	return fmt.Sprintf(`DELETE FROM %s WHERE filename = $1`, schemaName), nil
}

func (PostgresDialect) SelectSchemaSQL(schemaName string) (string, error) {
	if err := ValidateIdentifier(schemaName); err != nil {
		return "", err
	}
	return fmt.Sprintf(`SELECT id, version, filename, hash, status, created_at FROM %s ORDER BY id ASC`, schemaName), nil
}

type MySQLDialect struct{}

func (MySQLDialect) Name() string {
	return "mysql"
}

func (MySQLDialect) DriverName() string {
	return "mysql"
}

func (MySQLDialect) CreateSchemaSQL(schemaName string) (string, error) {
	if err := ValidateIdentifier(schemaName); err != nil {
		return "", err
	}
	return strings.TrimSpace(fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s (
    id BIGINT NOT NULL PRIMARY KEY AUTO_INCREMENT,
    version VARCHAR(128) NOT NULL,
    filename VARCHAR(200) NOT NULL UNIQUE,
    hash VARCHAR(128) NOT NULL,
    status VARCHAR(32) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
) CHARACTER SET utf8mb4;
`, schemaName)), nil
}

func (MySQLDialect) InsertSchemaSQL(schemaName string) (string, error) {
	if err := ValidateIdentifier(schemaName); err != nil {
		return "", err
	}
	return fmt.Sprintf(`INSERT INTO %s (id, version, filename, hash, status, created_at) VALUES (?, ?, ?, ?, ?, ?)`, schemaName), nil
}

func (MySQLDialect) DeleteSchemaSQL(schemaName string) (string, error) {
	if err := ValidateIdentifier(schemaName); err != nil {
		return "", err
	}
	return fmt.Sprintf(`DELETE FROM %s WHERE filename = ?`, schemaName), nil
}

func (MySQLDialect) SelectSchemaSQL(schemaName string) (string, error) {
	if err := ValidateIdentifier(schemaName); err != nil {
		return "", err
	}
	return fmt.Sprintf(`SELECT id, version, filename, hash, status, created_at FROM %s ORDER BY id ASC`, schemaName), nil
}

type MariaDBDialect struct{}

func (MariaDBDialect) Name() string {
	return "mariadb"
}

func (MariaDBDialect) DriverName() string {
	return "mysql"
}

func (MariaDBDialect) CreateSchemaSQL(schemaName string) (string, error) {
	return MySQLDialect{}.CreateSchemaSQL(schemaName)
}

func (MariaDBDialect) InsertSchemaSQL(schemaName string) (string, error) {
	return MySQLDialect{}.InsertSchemaSQL(schemaName)
}

func (MariaDBDialect) DeleteSchemaSQL(schemaName string) (string, error) {
	return MySQLDialect{}.DeleteSchemaSQL(schemaName)
}

func (MariaDBDialect) SelectSchemaSQL(schemaName string) (string, error) {
	return MySQLDialect{}.SelectSchemaSQL(schemaName)
}

type MSSQLDialect struct{}

func (MSSQLDialect) Name() string {
	return "mssql"
}

func (MSSQLDialect) DriverName() string {
	return "sqlserver"
}

func (MSSQLDialect) CreateSchemaSQL(schemaName string) (string, error) {
	if err := ValidateIdentifier(schemaName); err != nil {
		return "", err
	}
	return strings.TrimSpace(fmt.Sprintf(`
IF OBJECT_ID(N'%s', N'U') IS NULL
BEGIN
    CREATE TABLE %s (
        id BIGINT NOT NULL PRIMARY KEY,
        version NVARCHAR(128) NOT NULL,
        filename NVARCHAR(200) NOT NULL UNIQUE,
        hash NVARCHAR(128) NOT NULL,
        status NVARCHAR(32) NOT NULL,
        created_at DATETIME2 NOT NULL DEFAULT SYSUTCDATETIME()
    );
END
`, schemaName, schemaName)), nil
}

func (MSSQLDialect) InsertSchemaSQL(schemaName string) (string, error) {
	if err := ValidateIdentifier(schemaName); err != nil {
		return "", err
	}
	return fmt.Sprintf(`INSERT INTO %s (id, version, filename, hash, status, created_at) VALUES (@p1, @p2, @p3, @p4, @p5, @p6)`, schemaName), nil
}

func (MSSQLDialect) DeleteSchemaSQL(schemaName string) (string, error) {
	if err := ValidateIdentifier(schemaName); err != nil {
		return "", err
	}
	return fmt.Sprintf(`DELETE FROM %s WHERE filename = @p1`, schemaName), nil
}

func (MSSQLDialect) SelectSchemaSQL(schemaName string) (string, error) {
	if err := ValidateIdentifier(schemaName); err != nil {
		return "", err
	}
	return fmt.Sprintf(`SELECT id, version, filename, hash, status, created_at FROM %s ORDER BY id ASC`, schemaName), nil
}

type OracleDialect struct{}

func (OracleDialect) Name() string {
	return "oracle"
}

func (OracleDialect) DriverName() string {
	return "oracle"
}

func (OracleDialect) CreateSchemaSQL(schemaName string) (string, error) {
	if err := ValidateIdentifier(schemaName); err != nil {
		return "", err
	}
	return strings.TrimSpace(fmt.Sprintf(`
BEGIN
    EXECUTE IMMEDIATE 'CREATE TABLE %s (
        id NUMBER GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
        version VARCHAR2(128 CHAR) NOT NULL,
        filename VARCHAR2(200 CHAR) NOT NULL UNIQUE,
        hash VARCHAR2(128 CHAR) NOT NULL,
        status VARCHAR2(32 CHAR) NOT NULL,
        created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT SYSTIMESTAMP
    )';
EXCEPTION
    WHEN OTHERS THEN
        IF SQLCODE != -955 THEN
            RAISE;
        END IF;
END;
`, schemaName)), nil
}

func (OracleDialect) InsertSchemaSQL(schemaName string) (string, error) {
	if err := ValidateIdentifier(schemaName); err != nil {
		return "", err
	}
	return fmt.Sprintf(`INSERT INTO %s (id, version, filename, hash, status, created_at) VALUES (:1, :2, :3, :4, :5, :6)`, schemaName), nil
}

func (OracleDialect) DeleteSchemaSQL(schemaName string) (string, error) {
	if err := ValidateIdentifier(schemaName); err != nil {
		return "", err
	}
	return fmt.Sprintf(`DELETE FROM %s WHERE filename = :1`, schemaName), nil
}

func (OracleDialect) SelectSchemaSQL(schemaName string) (string, error) {
	if err := ValidateIdentifier(schemaName); err != nil {
		return "", err
	}
	return fmt.Sprintf(`SELECT id, version, filename, hash, status, created_at FROM %s ORDER BY id ASC`, schemaName), nil
}

type ClickHouseDialect struct{}

func (ClickHouseDialect) Name() string {
	return "clickhouse"
}

func (ClickHouseDialect) DriverName() string {
	return "clickhouse"
}

func (ClickHouseDialect) CreateSchemaSQL(schemaName string) (string, error) {
	if err := ValidateIdentifier(schemaName); err != nil {
		return "", err
	}
	return strings.TrimSpace(fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s (
    id UInt64,
    version String,
    filename String,
    hash String,
    status String,
    created_at DateTime64(3, 'UTC')
) ENGINE = MergeTree()
ORDER BY id;
`, schemaName)), nil
}

func (ClickHouseDialect) InsertSchemaSQL(schemaName string) (string, error) {
	if err := ValidateIdentifier(schemaName); err != nil {
		return "", err
	}
	return fmt.Sprintf(`INSERT INTO %s (id, version, filename, hash, status, created_at) VALUES (?, ?, ?, ?, ?, ?)`, schemaName), nil
}

func (ClickHouseDialect) DeleteSchemaSQL(schemaName string) (string, error) {
	if err := ValidateIdentifier(schemaName); err != nil {
		return "", err
	}
	return fmt.Sprintf(`DELETE FROM %s WHERE filename = ? SETTINGS mutations_sync = 2`, schemaName), nil
}

func (ClickHouseDialect) SelectSchemaSQL(schemaName string) (string, error) {
	if err := ValidateIdentifier(schemaName); err != nil {
		return "", err
	}
	return fmt.Sprintf(`SELECT id, version, filename, hash, status, created_at FROM %s ORDER BY id ASC`, schemaName), nil
}

type SQLiteDialect struct{}

func (SQLiteDialect) Name() string {
	return "sqlite"
}

func (SQLiteDialect) DriverName() string {
	return "sqlite"
}

func (SQLiteDialect) CreateSchemaSQL(schemaName string) (string, error) {
	if err := ValidateIdentifier(schemaName); err != nil {
		return "", err
	}
	return strings.TrimSpace(fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    version TEXT NOT NULL,
    filename TEXT NOT NULL UNIQUE,
    hash TEXT NOT NULL,
    status TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
`, schemaName)), nil
}

func (SQLiteDialect) InsertSchemaSQL(schemaName string) (string, error) {
	if err := ValidateIdentifier(schemaName); err != nil {
		return "", err
	}
	return fmt.Sprintf(`INSERT INTO %s (id, version, filename, hash, status, created_at) VALUES (?, ?, ?, ?, ?, ?)`, schemaName), nil
}

func (SQLiteDialect) DeleteSchemaSQL(schemaName string) (string, error) {
	if err := ValidateIdentifier(schemaName); err != nil {
		return "", err
	}
	return fmt.Sprintf(`DELETE FROM %s WHERE filename = ?`, schemaName), nil
}

func (SQLiteDialect) SelectSchemaSQL(schemaName string) (string, error) {
	if err := ValidateIdentifier(schemaName); err != nil {
		return "", err
	}
	return fmt.Sprintf(`SELECT id, version, filename, hash, status, created_at FROM %s ORDER BY id ASC`, schemaName), nil
}

type TiDBDialect struct{}

func (TiDBDialect) Name() string {
	return "tidb"
}

func (TiDBDialect) DriverName() string {
	return "mysql"
}

func (TiDBDialect) CreateSchemaSQL(schemaName string) (string, error) {
	return MySQLDialect{}.CreateSchemaSQL(schemaName)
}

func (TiDBDialect) InsertSchemaSQL(schemaName string) (string, error) {
	return MySQLDialect{}.InsertSchemaSQL(schemaName)
}

func (TiDBDialect) DeleteSchemaSQL(schemaName string) (string, error) {
	return MySQLDialect{}.DeleteSchemaSQL(schemaName)
}

func (TiDBDialect) SelectSchemaSQL(schemaName string) (string, error) {
	return MySQLDialect{}.SelectSchemaSQL(schemaName)
}

type RedshiftDialect struct{}

func (RedshiftDialect) Name() string {
	return "redshift"
}

func (RedshiftDialect) DriverName() string {
	return "pgx"
}

func (RedshiftDialect) CreateSchemaSQL(schemaName string) (string, error) {
	if err := ValidateIdentifier(schemaName); err != nil {
		return "", err
	}
	return strings.TrimSpace(fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s (
    id BIGINT NOT NULL PRIMARY KEY,
    version VARCHAR(128) NOT NULL,
    filename VARCHAR(200) NOT NULL UNIQUE,
    hash VARCHAR(128) NOT NULL,
    status VARCHAR(32) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT GETDATE()
);
`, schemaName)), nil
}

func (RedshiftDialect) InsertSchemaSQL(schemaName string) (string, error) {
	if err := ValidateIdentifier(schemaName); err != nil {
		return "", err
	}
	return fmt.Sprintf(`INSERT INTO %s (id, version, filename, hash, status, created_at) VALUES ($1, $2, $3, $4, $5, $6)`, schemaName), nil
}

func (RedshiftDialect) DeleteSchemaSQL(schemaName string) (string, error) {
	if err := ValidateIdentifier(schemaName); err != nil {
		return "", err
	}
	return fmt.Sprintf(`DELETE FROM %s WHERE filename = $1`, schemaName), nil
}

func (RedshiftDialect) SelectSchemaSQL(schemaName string) (string, error) {
	if err := ValidateIdentifier(schemaName); err != nil {
		return "", err
	}
	return fmt.Sprintf(`SELECT id, version, filename, hash, status, created_at FROM %s ORDER BY id ASC`, schemaName), nil
}

func FromName(name string) (Dialect, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "", "postgres", "pgsql", "postgresql":
		return PostgresDialect{}, nil
	case "mysql":
		return MySQLDialect{}, nil
	case "mariadb":
		return MariaDBDialect{}, nil
	case "mssql", "sqlserver", "azuresql":
		return MSSQLDialect{}, nil
	case "oracle", "orcl", "ora", "go-ora", "goora", "godror":
		return OracleDialect{}, nil
	case "clickhouse":
		return ClickHouseDialect{}, nil
	case "sqlite", "sqlite3":
		return SQLiteDialect{}, nil
	case "tidb":
		return TiDBDialect{}, nil
	case "redshift":
		return RedshiftDialect{}, nil
	default:
		return nil, fmt.Errorf("unsupported dialect: %s", name)
	}
}
