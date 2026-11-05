# Aus Go verwenden

Die öffentliche API des Root-Pakets folgt SemVer, `internal/*` gehört nicht dazu. Der Aufrufer verwaltet den Pool, einen geeigneten Context und serialisiert Migrationen derselben Datenbank.

`Config.DryRun` und `DryRunOutput` zeigen die Vorschau, aber `Create()` erstellt weiterhin die Verlaufstabelle. Standardlogger: `NoopLogger`. Für [eingebettete Skripte](./modules) `FSSource` mit Abhängigkeiten verwenden.

Bei `OpenConnection` mit `Connection.Context()` pingen und den Dienst aufrufen, anschließend `Close()`. Netzwerktreiber werden registriert; SQLite muss der Aufrufer registrieren. `ConnectionOptions.Prompt` und `HostKeyPrompt` sind optional und müssen Abbrüche beachten. Ohne Callbacks keine Eingabe und kein Vertrauen unbekannter Hosts.

Zusätzlich zur CLI kann `github.com/exc-works/migrate` direkt aus deinem Servicecode importiert werden, um Migrationen auszuführen — nützlich für Unit-Tests, Startup-Hooks oder Admin-Panels.

## Installation

```bash
go get github.com/exc-works/migrate
```

Importiere den benötigten Datenbanktreiber (die Bibliothek bindet keinen spezifischen Treiber):

```go
import (
    _ "modernc.org/sqlite"             // sqlite
    _ "github.com/jackc/pgx/v5/stdlib" // postgres
    _ "github.com/go-sql-driver/mysql" // mysql / mariadb / tidb
    // ...
)
```

## Minimales Beispiel

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

    if err := svc.Create(); err != nil { // idempotent: legt die History-Tabelle an, falls sie fehlt
        panic(err)
    }
    if err := svc.Up(); err != nil {
        panic(err)
    }
}
```

## Zentrale API

- `migrate.NewService(ctx, migrate.Config)` baut einen Migration-Executor
- `svc.Create()` legt die History-Tabelle `migration_schema` an (idempotent)
- `svc.Up()` wendet alle ausstehenden Migrationen an
- `svc.Down(toVersion, all)` rollt auf eine Zielversion oder komplett zurück
- `svc.Status()` liefert `[]migrate.MigrationStatus`
- `svc.Baseline()` markiert bestehende ausstehende Dateien als `baseline`

Häufige Typen:

- Dialekte (Konstruktoren bevorzugen — sie liefern das `Dialect`-Interface): `migrate.NewPostgresDialect()`, `NewMySQLDialect()`, `NewSQLiteDialect()`, `NewMSSQLDialect()`, `NewOracleDialect()`, `NewClickHouseDialect()`, `NewMariaDBDialect()`, `NewTiDBDialect()`, `NewRedshiftDialect()`, oder `migrate.DialectFromName("postgres")` für namensbasierte Auflösung
- Quellen: `DirectorySource` (Dateisystem), `StringSource` (In-Memory-Slice), `FSSource` (beliebiges `fs.FS`, z. B. `//go:embed` oder `os.DirFS`), `CombinedSource` (fasst mehrere Quellen zusammen)
- Logger: `migrate.NoopLogger{}` (Standard), `migrate.NewStdLogger("info", os.Stdout)` oder ein beliebiger Typ, der `migrate.Logger` implementiert

[Go API](https://github.com/exc-works/migrate/blob/main/migrate.go) · [Go example](https://github.com/exc-works/migrate/blob/main/example_test.go)
