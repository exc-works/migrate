# Скриптовые миграции

JS, TS, MJS преобразуют данные без Node.js при выполнении.

```sh
migrate new version "normalize names" --format ts
```

`V<version>__<description>.<format>` использует общую с SQL последовательность. MJS ограничен так же, как JS.

```ts
export function up(db) {
  for (const row of db.query("SELECT id, name FROM users")) {
    db.exec("UPDATE users SET name=? WHERE id=?", [row.name.trim(), row.id]);
  }
}
```

Пример для SQLite `users.name`. Созданный up бросает ошибку до реализации. Функции синхронные: без async, генераторов, Promise. Обычный результат игнорируется. [Параметры SQL](./script-api) зависят от базы.

Обрезка пробелов теряет исходные данные, поэтому down здесь нет. Экспортируйте `down(db)` только с реализованной обратной операцией и сохраняйте оригиналы, если нужен возврат.

```sh
migrate up --dry-run
migrate up
migrate status
```

Dry-run не выполняет скрипт. DB-вызовы размещайте в up/down, не на уровне модуля. Не разделяйте глобальное состояние между миграциями. Типы есть в шаблоне и [migration-api.d.ts](../assets/migration-api.d.ts), но транспиляция типы не проверяет. Нет Node API, сети, файлов, окружения, таймеров или динамических импортов; проверяйте и npm. См. [модули](./modules), [транзакции](./script-transactions).
