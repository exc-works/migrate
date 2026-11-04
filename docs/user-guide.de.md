# migrate Benutzerhandbuch

## 1. Installation

### 1.1 Aus Modul installieren

```bash
go install github.com/exc-works/migrate/cmd/migrate@latest
```

Eine bestimmte Version installieren:

```bash
go install github.com/exc-works/migrate/cmd/migrate@vX.Y.Z
```

Ersetzen Sie `vX.Y.Z` durch eine reale Version, zum Beispiel `v0.2.3`.

### 1.2 Aus lokalem Quellcode installieren (privates Repo oder internes Netzwerk)

Im Repository-Root ausführen:

```bash
go install ./cmd/migrate
```

### 1.3 Installation prüfen

```bash
migrate --help
```

Wenn der Befehl nicht gefunden wird:

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

Wenn `Repository not found` angezeigt wird, verwenden Sie den obigen Installationsweg aus lokalem Quellcode.

## 2. Initialisierung

### 2.1 Konfigurationsdatei erzeugen

```bash
migrate new config
```

Optional:

```bash
migrate new config dev.json
migrate new config --force
```

Standard-Konfigurationsvorlage:

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

### 2.2 Wichtige Konfigurationsfelder aktualisieren

- `dialect`: `postgres`, `mysql`, `mariadb`, `oracle`, `sqlite`, `mssql`, `clickhouse`, `tidb`, `redshift`
- `data_source_name`: DB-Verbindungszeichenfolge
- `migration_source`: Migrationsverzeichnis (Standard: `migrations`)

### 2.3 Migrationsverlaufstabelle initialisieren

```bash
migrate create
```

`create` kann ohne Ausgabe erfolgreich sein. Prüfen mit:

```bash
migrate status
```

Wenn bereits ein Schema existiert und alte SQL-Dateien nicht erneut ausgeführt werden sollen, verwenden Sie:

```bash
migrate baseline
```

## 3. Migrationsversionsdateien erstellen

### 3.1 Automatisch generierte Version

```bash
migrate new version init_users
```

### 3.2 Explizite Version

```bash
migrate new version add_email -v 202604140002
```

Format für generierte Dateinamen:

```text
V<version>__<description>.sql
```

Standard-Dateivorlage:

```sql
-- +migrate Up

-- +migrate Down
```

## 4. Upgrade (Migrationen anwenden)

Zuerst Dry-Run:

```bash
migrate up --dry-run
```

Dann tatsächlich anwenden:

```bash
migrate up
```

Danach Status prüfen:

```bash
migrate status
```

`up` kann ohne Ausgabe erfolgreich sein. Verwenden Sie `status` als maßgebliche Quelle.

## 5. Rollback

### 5.1 Auf eine Zielversion zurückrollen (Zielversion bleibt erhalten)

```bash
migrate down 202604140001
```

Semantik: Nur angewendete Versionen größer als `202604140001` werden zurückgerollt.

### 5.2 Alle angewendeten Versionen zurückrollen

```bash
migrate down --all
```


### 5.3 Dry-Run-Rollback

```bash
migrate down 202604140001 --dry-run
migrate down --all --dry-run
```

`down` kann ohne Ausgabe erfolgreich sein. Führen Sie `migrate status` zur Überprüfung aus.

## 6. Status prüfen

```bash
migrate status
```

Maschinenlesbare Ausgabe (empfohlen für Skripte und KI-Agenten):

```bash
migrate status --output json
```

Ausgabespalten: `Version`, `Filename`, `Hash`, `Status`.

Häufige Statuswerte:

- `pending`
- `applied`
- `baseline`
- `outOfOrder`
- `hashMismatch`
- `filenameMismatch`

## 7. Tool-Version prüfen

```bash
migrate version
```

## 8. Umgebungsvariablen-Templates

`data_source_name` unterstützt:

- `${KEY}`: erforderlich, muss vorhanden sein
- `${KEY:default}`: verwendet `default`, wenn `KEY` fehlt

Beispiel:

```json
{
  "dialect": "postgres",
  "data_source_name": "host=127.0.0.1 port=5432 user=${DB_USER:postgres} password=${DB_PASSWORD} dbname=${DB_NAME:app} sslmode=disable"
}
```

Stellen Sie sicher, dass `DB_PASSWORD` bereits in Ihrer Umgebung gesetzt ist, und führen Sie dann aus:

```bash
migrate status
```

## 9. 10-Minuten-Erstdemo (SQLite)

[10-Minuten-Erstdemo (SQLite) · English](site/pages/quick-start.md)

## 10. Globale Flags

Bestimmte Konfigurationsdatei verwenden:

```bash
migrate -c ./configs/dev.json status
```

Bestimmtes Arbeitsverzeichnis verwenden:

```bash
migrate -w ./deploy create
migrate -w ./deploy up
```

## 11. Häufige Fehler und Fehlerbehebung

### 11.1 Konfigurationsdatei nicht gefunden

Fehler: `config file ... no such file or directory`

Lösung:

- sicherstellen, dass `migration_config.json` im aktuellen Verzeichnis vorhanden ist
- oder den Konfigurationspfad mit `-c` übergeben

### 11.2 Fehlende Umgebungsvariable

Fehler: `can't find env value for XXX`

Lösung:

- `export XXX=...`
- oder `${XXX:default}` verwenden

### 11.3 Unvollständige `down`-Argumente

Fehler: `to-version must be set, or use --all`

Lösung:

- `migrate down <version>` verwenden
- oder `migrate down --all` verwenden

### 11.4 Nicht unterstützter Dialekt

Fehler: `unsupported dialect: xxx`

Lösung: Einen der folgenden verwenden:

- `postgres` `mysql` `mariadb` `oracle` `sqlite` `mssql` `clickhouse` `tidb` `redshift`

### 11.5 Inkonsistente Migrationsmetadaten

Fehler: `hash mismatch` oder `filename mismatch`

Lösung:

- die ursprüngliche Migrationsdatei aus der Versionsverwaltung wiederherstellen
- für Änderungen eine neue Migration mit höherer Version erstellen

## 12. migrate als Go-Bibliothek einbetten

Zusätzlich zur CLI kann `github.com/exc-works/migrate` direkt aus deinem Servicecode importiert werden, um Migrationen auszuführen — nützlich für Unit-Tests, Startup-Hooks oder Admin-Panels.

### 12.1 Installation

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

### 12.2 Minimales Beispiel

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

### 12.3 Zentrale API

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

[Go API](../migrate.go) · [Go example](../example_test.go)
