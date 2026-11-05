# Миграции SQL

Одно логическое изменение на файл; SQL и скрипты используют общие версии.

```sh
migrate new version "init users"
migrate new version "add email" -v 202611040002
```

Описание с пробелами заключайте в кавычки: в имени они станут `_`. `-v` задаёт версию, иначе она генерируется. Формат: `V<version>__<description>.sql`. Применённые файлы не меняйте и не переименовывайте — добавляйте версии.

```sql
-- +migrate Up
CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);

-- +migrate Down
DROP TABLE users;
```

Up применяет, Down отменяет. Здесь Down удалит и данные; заранее определите их восстановление.

Внутренние точки с запятой защищайте `StatementBegin` / `StatementEnd`:

```sql
-- +migrate Up
-- +migrate StatementBegin
CREATE FUNCTION migration_answer() RETURNS integer AS $$
BEGIN
    RETURN 42;
END;
$$ LANGUAGE plpgsql;
-- +migrate StatementEnd

-- +migrate Down
DROP FUNCTION migration_answer();
```

Синтаксис и транзакции DDL зависят от базы. Далее — [выполнение](./applying).
