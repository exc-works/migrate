# Início rápido com SQLite

Crie, aplique e reverta uma migração sem servidor de banco de dados.

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

Substitua `migration_config.yml` por:

```yaml
# Tabela do histórico
schema_name: migration_schema
dialect: sqlite
data_source_name: ./demo.sqlite # Arquivo SQLite local
working_directory: .
migrate_out_of_order: false
logger_level: info
migration_source: migrations # Diretório de migrações
```

## Criar a migração

```sh
migrate create
migrate new version "init users" -v 202611030001
```

Substitua `migrations/V202611030001__init_users.sql` por:

```sql
-- +migrate Up
CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);

-- +migrate Down
DROP TABLE users;
```

## Aplicar e verificar

```sh
migrate up --dry-run
migrate up
migrate status
```

A versão `202611030001` deve ficar `applied`.

## Reverter

```sh
migrate down --all --dry-run
migrate down --all
migrate status
```

A tabela é removida e a versão volta a `pending`. Continue com [SQL](./sql-migrations), [configuração](./configuration) e [scripts](./scripting).
