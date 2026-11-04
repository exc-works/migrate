# migrate User Guide

## 1. Installation

### 1.1 Install from module

```bash
go install github.com/exc-works/migrate/cmd/migrate@latest
```

Install a specific version:

```bash
go install github.com/exc-works/migrate/cmd/migrate@vX.Y.Z
```

Replace `vX.Y.Z` with a version from [Releases](https://github.com/exc-works/migrate/releases).

### 1.2 Install from local source (private repo or internal network)

Run in the repository root:

```bash
go install ./cmd/migrate
```

### 1.3 Verify installation

```bash
migrate --help
```

If the command is not found:

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

If you see `Repository not found`, use the local source install path above.

## 2. Initialization

### 2.1 Generate config file

```bash
migrate new config
```

Optional:

```bash
migrate new config dev.json
migrate new config --force
```

Default config template:

```json
{
  "schema_name": "migration_schema",
  "dialect": "postgres",
  "data_source_name": "host=127.0.0.1 port=5432 user=${DB_USER:postgres} password=${DB_PASSWORD:change_me} dbname=${DB_NAME:postgres} sslmode=disable",
  "working_directory": "",
  "migrate_out_of_order": false,
  "logger_level": "info",
  "migration_source": "migrations"
}
```

### 2.2 Update key config fields

- `dialect`: `postgres`, `mysql`, `mariadb`, `oracle`, `sqlite`, `mssql`, `clickhouse`, `tidb`, `redshift`
- `data_source_name`: DB connection string
- `migration_source`: migration directory (default: `migrations`)

### 2.3 Initialize migration history table

```bash
migrate create
```

Check migration status:

```bash
migrate status
```

To adopt an existing database, match its schema to the migration files, then record a baseline:

```bash
migrate baseline
```

## 3. Create migration version files

### 3.1 Auto-generated version

```bash
migrate new version init_users
```

### 3.2 Explicit version

```bash
migrate new version add_email -v 202604140002
```

Generated filename format:

```text
V<version>__<description>.sql
```

Default file template:

```sql
-- +migrate Up

-- +migrate Down
```

## 4. Upgrade (apply migrations)

Dry run first:

```bash
migrate up --dry-run
```

Apply for real:

```bash
migrate up
```

Then check status:

```bash
migrate status
```

Applied versions appear as `applied` in the status output.

## 5. Rollback

### 5.1 Roll back to a target version (target version is kept)

```bash
migrate down 202604140001
```

Semantics: only applied versions greater than `202604140001` are rolled back.

### 5.2 Roll back all applied versions

```bash
migrate down --all
```


### 5.3 Dry-run rollback

```bash
migrate down 202604140001 --dry-run
migrate down --all --dry-run
```

Run `migrate status` and confirm the rolled-back versions are `pending`.

## 6. Check status

```bash
migrate status
```

Read JSON from a script:

```bash
migrate status --output json
```

Output columns: `Version`, `Filename`, `Hash`, `Status`.

Common statuses:

- `pending`
- `applied`
- `baseline`
- `outOfOrder`
- `hashMismatch`
- `filenameMismatch`

## 7. Check tool version

```bash
migrate version
```

## 8. Environment variable templates

`data_source_name` supports:

- `${KEY}`: required, must exist
- `${KEY:default}`: use `default` if `KEY` is missing

Example:

```json
{
  "dialect": "postgres",
  "data_source_name": "host=127.0.0.1 port=5432 user=${DB_USER:postgres} password=${DB_PASSWORD} dbname=${DB_NAME:app} sslmode=disable"
}
```

Make sure `DB_PASSWORD` is already set in your environment, then run:

```bash
migrate status
```

## 9. 10-minute first-run demo (SQLite)

[10-minute first-run demo (SQLite)](site/pages/quick-start.md)

## 10. Global flags

Use specific config file:

```bash
migrate -c ./configs/dev.json status
```

Use specific working directory:

```bash
migrate -w ./deploy create
migrate -w ./deploy up
```

## 11. Common errors and troubleshooting

| Message | Action |
| --- | --- |
| `config file ... no such file or directory` | Run `migrate new config`, or select an existing file with `-c`. |
| `can't find env value for XXX` | Set environment variable `XXX`, or use `${XXX:default}` in the configuration. |
| `to-version must be set, or use --all` | Choose `migrate down <version>` or `migrate down --all`. |
| `unsupported dialect: xxx` | Set `dialect` to a database name listed in section 2. |
| `hash mismatch` / `filename mismatch` | Restore the original migration from version control, then add a new version for your change. |

## 12. Embedding migrate as a Go library

Run migrations from a Go application:

### 12.1 Installation

```bash
go get github.com/exc-works/migrate
```

Import your database driver:

```go
import (
    _ "modernc.org/sqlite"             // sqlite
    _ "github.com/jackc/pgx/v5/stdlib" // postgres
    _ "github.com/go-sql-driver/mysql" // mysql / mariadb / tidb
    // ...
)
```

### 12.2 Minimal example

```go
package main

import (
    "context"
    "database/sql"

    _ "modernc.org/sqlite"

    "github.com/exc-works/migrate"
)

func main() {
    db, err := sql.Open("sqlite", "./app.sqlite")
    if err != nil {
        panic(err)
    }
    defer db.Close()

    svc, err := migrate.NewService(context.Background(), migrate.Config{
        Dialect:         migrate.NewSQLiteDialect(),
        DB:              db,
        MigrationSource: migrate.DirectorySource{Directory: "./migrations"},
    })
    if err != nil {
        panic(err)
    }

    if err := svc.Create(); err != nil { // idempotent: creates the history table if missing
        panic(err)
    }
    if err := svc.Up(); err != nil {
        panic(err)
    }
}
```

### 12.3 Key API

- `migrate.NewService(ctx, migrate.Config)` builds a migration executor
- `svc.Create()` creates the `migration_schema` history table (idempotent)
- `svc.Up()` applies all pending migrations
- `svc.Down(toVersion, all)` rolls back to a target version or everything
- `svc.Status()` returns `[]migrate.MigrationStatus`
- `svc.Baseline()` marks existing pending files as `baseline`

Common types:

- Dialects (prefer the constructors — they return the `Dialect` interface): `migrate.NewPostgresDialect()`, `NewMySQLDialect()`, `NewSQLiteDialect()`, `NewMSSQLDialect()`, `NewOracleDialect()`, `NewClickHouseDialect()`, `NewMariaDBDialect()`, `NewTiDBDialect()`, `NewRedshiftDialect()`, or `migrate.DialectFromName("postgres")` for name-based lookup
- Sources: `DirectorySource` (filesystem), `StringSource` (in-memory slice), `FSSource` (any `fs.FS`, e.g. `//go:embed` or `os.DirFS`), `CombinedSource` (merges several sources)
- Loggers: `migrate.NoopLogger{}` (default), `migrate.NewStdLogger("info", os.Stdout)`, or any type implementing `migrate.Logger`

[Go API](../migrate.go) · [Go example](../example_test.go)

## SQL, JS and TS migrations

Migrate supports SQL and JS/TS/MJS scripts, including local and installed-package imports. See [SQL, JS and TS migrations](script-migrations.en.md).
