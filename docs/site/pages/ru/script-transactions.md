# Транзакции и ограничения

Если поддерживается, скрипт и запись успеха выполняются в одной транзакции: commit при успехе, rollback при ошибке. Не выполняйте BEGIN/COMMIT/ROLLBACK вручную. Неявный commit DDL и нетранзакционные базы могут оставить изменения.

```js
export const config = { transaction: false };

export function up(db) {
  // PostgreSQL: эта команда должна выполняться вне транзакции.
  db.exec("CREATE INDEX CONCURRENTLY users_name_idx ON users (name)");
}

export function down(db) {
  db.exec("DROP INDEX CONCURRENTLY users_name_idx");
}
```

`transaction` — boolean, по умолчанию `true`, действует на up/down. status, baseline, dry-run не вычисляют экспортированную настройку. Вызовы одной миграции используют выделенное соединение. Завершайте свои транзакции и восстанавливайте настройки сессии до возврата. Сессию между миграциями не считайте общей. После нетранзакционного сбоя возможны изменения без успешной истории.

## Лимиты ресурсов

```yaml
script_timeout: 5m # Время выполнения одного скрипта
script_max_rows: 1000 # Максимум строк одного запроса
script_max_bytes: 16777216 # Максимум байтов запроса (16 MiB)
```

Превышение — ошибка, не усечение. Большие выборки делите на страницы, без автоматического частичного commit или возобновления. Отмена DB может занять время. В Go — `Config.ScriptOptions`. См. [поддержку баз](./databases#script-support).
