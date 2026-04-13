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

func FromName(name string) (Dialect, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "", "postgres", "pgsql", "postgresql":
		return PostgresDialect{}, nil
	case "mysql":
		return MySQLDialect{}, nil
	case "mariadb":
		return MariaDBDialect{}, nil
	default:
		return nil, fmt.Errorf("unsupported dialect: %s", name)
	}
}
