# Быстрый старт с SQLite

Создайте, примените и откатите SQL-миграцию без отдельного сервера.

## Установка

```sh
go install github.com/exc-works/migrate/cmd/migrate@latest
migrate --help
migrate version
```

## Настройка

```sh
mkdir migrate-demo
cd migrate-demo
migrate new config
```

Замените `migration_config.yml`:

```yaml
# Таблица истории
schema_name: migration_schema
dialect: sqlite
data_source_name: ./demo.sqlite # Локальный файл SQLite
working_directory: .
migrate_out_of_order: false
logger_level: info
migration_source: migrations # Каталог миграций
```

## Создание миграции

```sh
migrate create
migrate new version "init users" -v 202611030001
```

Замените `migrations/V202611030001__init_users.sql`:

```sql
-- +migrate Up
CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);

-- +migrate Down
DROP TABLE users;
```

## Выполнение и проверка

```sh
migrate up --dry-run
migrate up
migrate status
```

У версии `202611030001` должен быть статус `applied`.

## Откат

```sh
migrate down --all --dry-run
migrate down --all
migrate status
```

Таблица удаляется, статус снова `pending`. Далее: [SQL](./sql-migrations), [настройка](./configuration), [скрипты](./scripting).
