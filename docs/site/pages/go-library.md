# Use migrate from Go

Import the root package to run migrations during tests, application startup or an
administrative operation. The public API follows SemVer; `internal/*` is not public API.

## Install and register a driver

```sh
go get github.com/exc-works/migrate
```

Your application registers the driver for a caller-created `*sql.DB`. This complete
example uses SQLite and reads SQL files from `./migrations`:

```go
package main

import (
    "context"
    "database/sql"
    "log"

    "github.com/exc-works/migrate"
    _ "modernc.org/sqlite"
)

func main() {
    if err := run(context.Background()); err != nil {
        log.Fatal(err)
    }
}

func run(ctx context.Context) error {
    db, err := sql.Open("sqlite", "./app.sqlite")
    if err != nil {
        return err
    }
    defer db.Close()

    service, err := migrate.NewService(ctx, migrate.Config{
        Dialect: migrate.NewSQLiteDialect(),
        DB:      db,
        MigrationSource: migrate.DirectorySource{
            Directory: "./migrations",
        },
    })
    if err != nil {
        return err
    }
    if err := service.Create(); err != nil {
        return err
    }
    return service.Up()
}
```

The caller owns the database pool and its lifetime. Use a context appropriate for
your deployment timeout. Serialize migrations against the same database.

## Choose a source and dialect

| Source | Use case |
| --- | --- |
| `DirectorySource` | Version files on disk |
| `StringSource` | In-memory migration text, especially tests |
| `FSSource` | Any `fs.FS`, including `embed.FS` and synthetic test filesystems |
| `CombinedSource` | Merge multiple sources |

Use dialect constructors such as `NewPostgresDialect()` or `NewSQLiteDialect()`,
or `DialectFromName("postgres")` for name-based lookup.

Embed SQL files with `//go:embed migrations/*.sql`, then pass
`migrate.FSSource{FS: files, Root: "migrations"}` as the source. For script helpers
and installed packages, see [Shared modules](./modules).

## Service operations

| Method | Behavior |
| --- | --- |
| `Create()` | Idempotently create the history table |
| `Up()` | Apply pending versions |
| `Down(toVersion, all)` | Reverse versions above a target, or all |
| `Status()` | Return `[]migrate.MigrationStatus` |
| `Baseline()` | Record pending versions without executing them |

Set `Config.DryRun` and `Config.DryRunOutput` to preview user migrations. `Create()`
is not affected by dry-run and still creates the history table.
The default logger is `NoopLogger`; use `NewStdLogger("info", os.Stdout)` or your
own `Logger` implementation when needed.

## Scripts and owned connections

Configure script execution limits through
`Config.ScriptOptions`. The [embedded example](./examples) demonstrates
scripts and helpers shipped inside a Go binary.

For a connection managed by migrate, including an optional SSH tunnel, use
`OpenConnection`. Ping `Connection.DB` with `Connection.Context()`, pass that DB
and context to `NewService`, and close the connection after execution. The helper
registers network drivers; SQLite still requires caller registration. Library
secret prompting is opt-in through `ConnectionOptions.Prompt`; host-key approval uses
the separate `HostKeyPrompt` callback. Nil callbacks never prompt or trust unknown
keys. Both must honor cancellation. See [SSH tunnels](./ssh-tunnels).

More runnable API examples are in
[example_test.go](https://github.com/exc-works/migrate/blob/main/example_test.go).
