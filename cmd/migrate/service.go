package main

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/exc-works/migrate/internal/config"
	"github.com/exc-works/migrate/internal/connection"
	"github.com/exc-works/migrate/internal/dialect"
	"github.com/exc-works/migrate/internal/logger"
	"github.com/exc-works/migrate/internal/migrate"
	"github.com/exc-works/migrate/internal/source"
	"github.com/exc-works/migrate/internal/tunnel"
)

func newSSHService(ctx context.Context, c *config.FileConfig, prompt tunnel.Prompt, hostKeyPrompt tunnel.HostKeyPrompt) (*migrate.Service, func() error, error) {
	return newCommandService(ctx, c, prompt, hostKeyPrompt, os.Stdout, os.Stdout)
}

func newServiceFromConfig(c *config.FileConfig) (*migrate.Service, error) {
	return newServiceFromConfigContext(context.Background(), c)
}

func newServiceFromConfigContext(ctx context.Context, c *config.FileConfig) (*migrate.Service, error) {
	svc, _, err := newCommandService(ctx, c, nil, nil, os.Stdout, os.Stdout)
	return svc, err
}

// Each invocation owns its DB/transport. Writers and callbacks are supplied by
// the caller so concurrent invocations never depend on Cobra's mutable state.
func newCommandService(ctx context.Context, c *config.FileConfig, prompt tunnel.Prompt, hostKeyPrompt tunnel.HostKeyPrompt, logs, plan io.Writer) (svc *migrate.Service, cleanup func() error, err error) {
	options, err := c.ScriptOptions()
	if err != nil {
		return nil, nil, err
	}
	d, err := dialect.FromName(c.Dialect)
	if err != nil {
		return nil, nil, err
	}
	var db *sql.DB
	var closeConnection func() error
	ping := pingDB
	if c.SSHTunnel != nil {
		conn, err := connection.Open(ctx, connection.Options{
			Dialect:        c.Dialect,
			DataSourceName: c.DataSourceName,
			SSHTunnel:      c.SSHTunnel,
			Prompt:         prompt,
			HostKeyPrompt:  hostKeyPrompt,
		})
		if err != nil {
			return nil, nil, err
		}
		db = conn.DB
		ctx = conn.Context()
		closeConnection = conn.Close
		ping = func(ctx context.Context, db *sql.DB) error { return db.PingContext(ctx) }
	} else {
		dsn, err := config.DriverDSN(d.DriverName(), c.DataSourceName)
		if err != nil {
			return nil, nil, err
		}
		db, err = openDB(d.DriverName(), dsn)
		if err != nil {
			return nil, nil, err
		}
		closeConnection = db.Close
	}
	owned := true
	defer func() {
		if owned {
			err = errors.Join(err, closeConnection())
		}
	}()
	pingCtx, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()
	if err := ping(pingCtx, db); err != nil {
		if c.SSHTunnel != nil {
			err = errors.New("database connection through SSH failed")
		}
		return nil, nil, err
	}
	svc, err = newMigrateService(ctx, migrate.Config{
		Dialect:           d,
		DB:                db,
		Logger:            logger.NewStd(c.LoggerLevel, logs),
		SchemaName:        c.SchemaName,
		MigrateOutOfOrder: c.MigrateOutOfOrder,
		MigrationSource: source.DirectorySource{
			Directory: filepath.Join(c.WorkingDirectory, c.MigrationSource),
		},
		DryRunOutput:  plan,
		ScriptOptions: options,
	})
	if err != nil {
		return nil, nil, err
	}
	owned = false
	return svc, closeConnection, nil
}
