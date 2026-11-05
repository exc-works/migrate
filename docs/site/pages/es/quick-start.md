# Inicio rápido con SQLite

Crea, aplica y revierte una migración sin servidor de base de datos.

## Instalar

```sh
go install github.com/exc-works/migrate/cmd/migrate@latest
migrate --help
migrate version
```

## Configurar

```sh
mkdir migrate-demo
cd migrate-demo
migrate new config
```

Sustituye `migration_config.yml` por:

```yaml
# Tabla del historial
schema_name: migration_schema
dialect: sqlite
data_source_name: ./demo.sqlite # Archivo SQLite local
working_directory: .
migrate_out_of_order: false
logger_level: info
migration_source: migrations # Directorio de migraciones
```

## Crear una migración

```sh
migrate create
migrate new version "init users" -v 202611030001
```

Sustituye `migrations/V202611030001__init_users.sql` por:

```sql
-- +migrate Up
CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);

-- +migrate Down
DROP TABLE users;
```

## Aplicar y comprobar

```sh
migrate up --dry-run
migrate up
migrate status
```

La versión `202611030001` debe aparecer como `applied`.

## Revertir

```sh
migrate down --all --dry-run
migrate down --all
migrate status
```

La tabla desaparece y el estado vuelve a `pending`. Consulta [SQL](./sql-migrations), [configuración](./configuration) y [scripts](./scripting).
