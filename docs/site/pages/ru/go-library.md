# Использование из Go

Публичный API корневого пакета следует SemVer; `internal/*` не публичен. Вызывающий закрывает пул, задаёт context и последовательно выполняет миграции одной базы.

`Config.DryRun` и `DryRunOutput` показывают план, но `Create()` всё равно создаёт историю. Логгер по умолчанию — `NoopLogger`. Для [встроенных скриптов](./modules) настройте `FSSource` и зависимости.

При `OpenConnection` используйте `Connection.Context()` для Ping и сервиса, затем `Close()`. Сетевые драйверы регистрируются, SQLite регистрируйте сами. `ConnectionOptions.Prompt` и `HostKeyPrompt` необязательны, должны поддерживать отмену. Без них нет запросов ввода и доверия неизвестным хостам.

Помимо CLI, `github.com/exc-works/migrate` можно импортировать прямо из кода вашего сервиса для запуска миграций — удобно для модульных тестов, хуков запуска или админ-панелей.

## Установка

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

## Минимальный пример

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

## Ключевой API

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

[Go API](https://github.com/exc-works/migrate/blob/main/migrate.go) · [Go example](https://github.com/exc-works/migrate/blob/main/example_test.go)
