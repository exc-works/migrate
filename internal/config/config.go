package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

type FileConfig struct {
	SchemaName        string `json:"schema_name"`
	Dialect           string `json:"dialect"`
	DataSourceName    string `json:"data_source_name"`
	WorkingDirectory  string `json:"working_directory"`
	MigrateOutOfOrder bool   `json:"migrate_out_of_order"`
	LoggerLevel       string `json:"logger_level"`
	MigrationSource   string `json:"migration_source"`
}

func Read(path, workingDir string) (*FileConfig, error) {
	if path == "" {
		path = "migration_config.json"
	}
	if workingDir != "" {
		path = filepath.Join(workingDir, path)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg FileConfig
	if err := json.Unmarshal(content, &cfg); err != nil {
		return nil, err
	}
	if workingDir != "" {
		cfg.WorkingDirectory = workingDir
	}
	applyDefaults(&cfg)
	cfg.DataSourceName, err = ReplaceEnvVar(cfg.DataSourceName)
	if err != nil {
		return nil, err
	}
	if cfg.DataSourceName == "" {
		return nil, fmt.Errorf("data_source_name must be set")
	}
	return &cfg, nil
}

func applyDefaults(cfg *FileConfig) {
	if cfg.SchemaName == "" {
		cfg.SchemaName = "migration_schema"
	}
	if cfg.Dialect == "" {
		cfg.Dialect = "postgres"
	}
	if cfg.LoggerLevel == "" {
		cfg.LoggerLevel = "info"
	}
	if cfg.MigrationSource == "" {
		cfg.MigrationSource = "migrations"
	}
}

var envPattern = regexp.MustCompile(`\$\{([^:}]+)(?::([^}]*))?\}`)

func ReplaceEnvVar(in string) (string, error) {
	var replaceErr error
	out := envPattern.ReplaceAllStringFunc(in, func(match string) string {
		parts := envPattern.FindStringSubmatch(match)
		if len(parts) < 2 {
			return match
		}
		key := parts[1]
		val, ok := os.LookupEnv(key)
		if ok {
			return val
		}
		if len(parts) > 2 && parts[2] != "" {
			return parts[2]
		}
		replaceErr = fmt.Errorf("can't find env value for %s", key)
		return match
	})
	if replaceErr != nil {
		return "", replaceErr
	}
	return out, nil
}

func DefaultTemplate() FileConfig {
	return FileConfig{
		SchemaName:        "migration_schema",
		Dialect:           "postgres",
		DataSourceName:    "host=127.0.0.1 port=5432 user=${DB_USER:postgres} password=${DB_PASSWORD:change_me} dbname=${DB_NAME:postgres} sslmode=disable",
		MigrateOutOfOrder: false,
		LoggerLevel:       "info",
		MigrationSource:   "migrations",
	}
}
