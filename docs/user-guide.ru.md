# Руководство пользователя migrate

Это руководство предназначено для пользователей, которые запускают инструмент впервые. Команды и флаги основаны на текущей реализации (`cmd/migrate`).

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

Примечание: `migrate down <to-version>` и `migrate down --all` взаимоисключающие.

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

## 7. Обновление или откат версии самого инструмента

### 7.1 Обновление версии инструмента

```bash
go install github.com/exc-works/migrate/cmd/migrate@latest
```

### 7.2 Откат версии инструмента

```bash
go install github.com/exc-works/migrate/cmd/migrate@vX.Y.Z
```

Пример:

```bash
go install github.com/exc-works/migrate/cmd/migrate@v0.2.3
```

Если репозиторий приватный и `go install github.com/...@...` недоступен, переключитесь на целевую версию в исходном коде
и выполните:

```bash
go install ./cmd/migrate
```

### 7.3 Проверка текущей версии инструмента

```bash
migrate version
```

Примечание: релизные артефакты выводят версию релиза; локальные сборки из `go install ./cmd/migrate` обычно выводят `dev`.

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

### 9.1 Подготовка директории и конфигурации

Сначала проверьте доступность команды:

```bash
migrate --help
```

Создайте demo-директорию (macOS/Linux):

```bash
mkdir -p ./migrate-demo
cd ./migrate-demo
migrate new config
```

Эквивалент для Windows PowerShell:

```powershell
mkdir .\migrate-demo
cd .\migrate-demo
migrate new config
```

Обновите `migration_config.json` до:

```json
{
  "schema_name": "migration_schema",
  "dialect": "sqlite",
  "data_source_name": "./demo.sqlite",
  "working_directory": ".",
  "migrate_out_of_order": false,
  "logger_level": "info",
  "migration_source": "migrations"
}
```

### 9.2 Инициализация и создание файлов миграций

```bash
migrate create
migrate new version init_users -v 202604140001
migrate new version add_email -v 202604140002
```

Отредактируйте `migrations/V202604140001__init_users.sql`:

```sql
-- +migrate Up
CREATE TABLE users
(
    id   INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);

-- +migrate Down
DROP TABLE IF EXISTS users;
```

Отредактируйте `migrations/V202604140002__add_email.sql`:

```sql
-- +migrate Up
ALTER TABLE users
    ADD COLUMN email TEXT;

-- +migrate Down
CREATE TABLE users_tmp
(
    id   INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);
INSERT INTO users_tmp (id, name)
SELECT id, name
FROM users;
DROP TABLE users;
ALTER TABLE users_tmp
    RENAME TO users;
```

### 9.3 Применение, проверка статуса и откат

```bash
migrate up --dry-run
migrate up
migrate status
migrate down 202604140001 --dry-run
migrate down 202604140001
migrate status
migrate down --all
migrate status
```

Ожидается:

- после `up`: обе версии имеют статус `applied`
- после `down 202604140001`: `202604140001=applied`, `202604140002=pending`
- после `down --all`: обе версии имеют статус `pending`

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

- не редактируйте уже примененные файлы миграций
- создайте новую миграцию с более высокой версией для изменений
