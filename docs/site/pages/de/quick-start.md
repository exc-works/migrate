# Schnellstart mit SQLite

Eine SQL-Migration ohne Datenbankserver erstellen, ausführen und zurücknehmen.

## Installieren

```sh
go install github.com/exc-works/migrate/cmd/migrate@latest
migrate --help
migrate version
```

## Konfigurieren

```sh
mkdir migrate-demo
cd migrate-demo
migrate new config
```

Ersetzen Sie `migration_config.yml` durch:

```yaml
# Tabelle für den Migrationsverlauf
schema_name: migration_schema
dialect: sqlite
data_source_name: ./demo.sqlite # Lokale SQLite-Datei
working_directory: .
migrate_out_of_order: false
logger_level: info
migration_source: migrations # Migrationsverzeichnis
```

## Migration erstellen

```sh
migrate create
migrate new version "init users" -v 202611030001
```

Ersetzen Sie `migrations/V202611030001__init_users.sql` durch:

```sql
-- +migrate Up
CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);

-- +migrate Down
DROP TABLE users;
```

## Ausführen und prüfen

```sh
migrate up --dry-run
migrate up
migrate status
```

Version `202611030001` sollte `applied` anzeigen.

## Zurücknehmen

```sh
migrate down --all --dry-run
migrate down --all
migrate status
```

Die Tabelle wird entfernt, der Status ist wieder `pending`. Weiter zu [SQL](./sql-migrations), [Konfiguration](./configuration) oder [Skripten](./scripting).
