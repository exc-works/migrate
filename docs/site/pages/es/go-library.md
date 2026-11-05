# Usar desde Go

La API pública del paquete raíz sigue SemVer; `internal/*` no forma parte de ella. El llamador cierra el pool, usa un context con tiempo adecuado y serializa migraciones en la misma base.

`Config.DryRun` y `DryRunOutput` previsualizan cambios, pero `Create()` crea siempre el historial. El registro predeterminado es `NoopLogger`. Para [scripts embebidos](./modules) usa `FSSource` y configura sus dependencias.

Con `OpenConnection`, usa `Connection.Context()` para Ping y el servicio, y termina con `Close()`. Registra SQLite tú mismo; los controladores de red se registran automáticamente. `ConnectionOptions.Prompt` y `HostKeyPrompt` son opcionales y deben atender cancelaciones. Sin ellos no se pide entrada ni se confía en hosts desconocidos.

Además del CLI, `github.com/exc-works/migrate` puede importarse directamente desde el código de tu servicio para ejecutar migraciones — útil para pruebas unitarias, hooks de arranque o paneles de administración.

## Instalación

```bash
go get github.com/exc-works/migrate
```

Importa el driver de base de datos que necesites (la librería no fija ninguno):

```go
import (
    _ "modernc.org/sqlite"             // sqlite
    _ "github.com/jackc/pgx/v5/stdlib" // postgres
    _ "github.com/go-sql-driver/mysql" // mysql / mariadb / tidb
    // ...
)
```

## Ejemplo mínimo

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

    if err := svc.Create(); err != nil { // idempotente: crea la tabla de historial si no existe
        panic(err)
    }
    if err := svc.Up(); err != nil {
        panic(err)
    }
}
```

## API principal

- `migrate.NewService(ctx, migrate.Config)` construye un ejecutor de migraciones
- `svc.Create()` crea la tabla de historial `migration_schema` (idempotente)
- `svc.Up()` aplica todas las migraciones pendientes
- `svc.Down(toVersion, all)` revierte hasta una versión objetivo o todo
- `svc.Status()` devuelve `[]migrate.MigrationStatus`
- `svc.Baseline()` marca los archivos pendientes existentes como `baseline`

Tipos comunes:

- Dialectos (prefiere los constructores — devuelven la interfaz `Dialect`): `migrate.NewPostgresDialect()`, `NewMySQLDialect()`, `NewSQLiteDialect()`, `NewMSSQLDialect()`, `NewOracleDialect()`, `NewClickHouseDialect()`, `NewMariaDBDialect()`, `NewTiDBDialect()`, `NewRedshiftDialect()`, o `migrate.DialectFromName("postgres")` para resolución por nombre
- Fuentes: `DirectorySource` (sistema de archivos), `StringSource` (slice en memoria), `FSSource` (cualquier `fs.FS`, p. ej. `//go:embed` o `os.DirFS`), `CombinedSource` (combina varias fuentes)
- Loggers: `migrate.NoopLogger{}` (por defecto), `migrate.NewStdLogger("info", os.Stdout)` o cualquier tipo que implemente `migrate.Logger`

[Go API](https://github.com/exc-works/migrate/blob/main/migrate.go) · [Go example](https://github.com/exc-works/migrate/blob/main/example_test.go)
