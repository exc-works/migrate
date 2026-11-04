# Руководство пользователя migrate

## 1. Установка

### 1.1 Установка из модуля

```bash
go install github.com/exc-works/migrate/cmd/migrate@latest
```

Установка конкретной версии:

```bash
go install github.com/exc-works/migrate/cmd/migrate@vX.Y.Z
```

Замените `vX.Y.Z` на реальную версию, например `v0.2.3`.

### 1.2 Установка из локального исходного кода (приватный репозиторий или внутренняя сеть)

Выполните в корне репозитория:

```bash
go install ./cmd/migrate
```

### 1.3 Проверка установки

```bash
migrate --help
```

Если команда не найдена:

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

Если вы видите `Repository not found`, используйте путь установки из локального исходного кода выше.

## 2. Инициализация

### 2.1 Создание файла конфигурации

```bash
migrate new config
```

Опционально:

```bash
migrate new config dev.json
migrate new config --force
```

Шаблон конфигурации по умолчанию:

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

### 2.2 Обновите ключевые поля конфигурации

- `dialect`: `postgres`, `mysql`, `mariadb`, `oracle`, `sqlite`, `mssql`, `clickhouse`, `tidb`, `redshift`
- `data_source_name`: строка подключения к БД
- `migration_source`: директория миграций (по умолчанию: `migrations`)

### 2.3 Инициализация таблицы истории миграций

```bash
migrate create
```

`create` может завершиться успешно без вывода. Подтвердите с помощью:

```bash
migrate status
```

Если у вас уже есть существующая схема и вы не хотите повторно применять старый SQL, используйте:

```bash
migrate baseline
```

## 3. Создание файлов версий миграций

### 3.1 Автоматически сгенерированная версия

```bash
migrate new version init_users
```

### 3.2 Явно заданная версия

```bash
migrate new version add_email -v 202604140002
```

Формат имени сгенерированного файла:

```text
V<version>__<description>.sql
```

Шаблон файла по умолчанию:

```sql
-- +migrate Up

-- +migrate Down
```

## 4. Обновление (применение миграций)

Сначала dry run:

```bash
migrate up --dry-run
```

Реальное применение:

```bash
migrate up
```

Затем проверьте статус:

```bash
migrate status
```

`up` может завершиться успешно без вывода. Используйте `status` как источник истины.

## 5. Откат

### 5.1 Откат к целевой версии (целевая версия сохраняется)

```bash
migrate down 202604140001
```

Семантика: откатываются только примененные версии больше `202604140001`.

### 5.2 Откат всех примененных версий

```bash
migrate down --all
```


### 5.3 Dry-run откат

```bash
migrate down 202604140001 --dry-run
migrate down --all --dry-run
```

`down` может завершиться успешно без вывода. Выполните `migrate status` для проверки.

## 6. Проверка статуса

```bash
migrate status
```

Машиночитаемый вывод (рекомендуется для скриптов и AI-агентов):

```bash
migrate status --output json
```

Колонки вывода: `Version`, `Filename`, `Hash`, `Status`.

Распространенные статусы:

- `pending`
- `applied`
- `baseline`
- `outOfOrder`
- `hashMismatch`
- `filenameMismatch`

## 7. Проверка версии

```bash
migrate version
```

## 8. Шаблоны переменных окружения

`data_source_name` поддерживает:

- `${KEY}`: обязательно, должно существовать
- `${KEY:default}`: использовать `default`, если `KEY` отсутствует

Пример:

```json
{
  "dialect": "postgres",
  "data_source_name": "host=127.0.0.1 port=5432 user=${DB_USER:postgres} password=${DB_PASSWORD} dbname=${DB_NAME:app} sslmode=disable"
}
```

Убедитесь, что `DB_PASSWORD` уже задана в окружении, затем выполните:

```bash
migrate status
```

## 9. 10-минутное демо первого запуска (SQLite)

[10-минутное демо первого запуска (SQLite) · English](site/pages/quick-start.md)

## 10. Глобальные флаги

Использование конкретного файла конфигурации:

```bash
migrate -c ./configs/dev.json status
```

Использование конкретной рабочей директории:

```bash
migrate -w ./deploy create
migrate -w ./deploy up
```

## 11. Частые ошибки и устранение неполадок

### 11.1 Файл конфигурации не найден

Ошибка: `config file ... no such file or directory`

Исправление:

- убедитесь, что `migration_config.json` существует в текущей директории
- или передайте путь к конфигурации через `-c`

### 11.2 Отсутствует переменная окружения

Ошибка: `can't find env value for XXX`

Исправление:

- `export XXX=...`
- или используйте `${XXX:default}`

### 11.3 Неполные аргументы `down`

Ошибка: `to-version must be set, or use --all`

Исправление:

- используйте `migrate down <version>`
- или используйте `migrate down --all`

### 11.4 Неподдерживаемый dialect

Ошибка: `unsupported dialect: xxx`

Исправление: используйте один из:

- `postgres` `mysql` `mariadb` `oracle` `sqlite` `mssql` `clickhouse` `tidb` `redshift`

### 11.5 Несоответствие метаданных миграции

Ошибка: `hash mismatch` или `filename mismatch`

Исправление:

- восстановите исходный файл миграции из системы контроля версий
- создайте новую миграцию с более высокой версией для изменений

## 12. Встраивание migrate как Go-библиотеки

Помимо CLI, `github.com/exc-works/migrate` можно импортировать прямо из кода вашего сервиса для запуска миграций — удобно для модульных тестов, хуков запуска или админ-панелей.

### 12.1 Установка

```bash
go get github.com/exc-works/migrate
```

Импортируйте нужный драйвер базы данных (библиотека не привязана к конкретному):

```go
import (
    _ "modernc.org/sqlite"             // sqlite
    _ "github.com/jackc/pgx/v5/stdlib" // postgres
    _ "github.com/go-sql-driver/mysql" // mysql / mariadb / tidb
    // ...
)
```

### 12.2 Минимальный пример

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

    if err := svc.Create(); err != nil { // идемпотентно: создаёт таблицу истории, если её нет
        panic(err)
    }
    if err := svc.Up(); err != nil {
        panic(err)
    }
}
```

### 12.3 Ключевой API

- `migrate.NewService(ctx, migrate.Config)` создаёт исполнитель миграций
- `svc.Create()` создаёт таблицу истории `migration_schema` (идемпотентно)
- `svc.Up()` применяет все ожидающие миграции
- `svc.Down(toVersion, all)` откатывает до целевой версии или полностью
- `svc.Status()` возвращает `[]migrate.MigrationStatus`
- `svc.Baseline()` помечает существующие pending-файлы как `baseline`

Часто используемые типы:

- Диалекты (предпочитайте конструкторы — возвращают интерфейс `Dialect`): `migrate.NewPostgresDialect()`, `NewMySQLDialect()`, `NewSQLiteDialect()`, `NewMSSQLDialect()`, `NewOracleDialect()`, `NewClickHouseDialect()`, `NewMariaDBDialect()`, `NewTiDBDialect()`, `NewRedshiftDialect()`, либо `migrate.DialectFromName("postgres")` для поиска по имени
- Источники: `DirectorySource` (файловая система), `StringSource` (срез в памяти), `FSSource` (любой `fs.FS`, например `//go:embed` или `os.DirFS`), `CombinedSource` (объединяет несколько источников)
- Логгеры: `migrate.NoopLogger{}` (по умолчанию), `migrate.NewStdLogger("info", os.Stdout)` или любой тип, реализующий `migrate.Logger`

[Go API](../migrate.go) · [Go example](../example_test.go)
