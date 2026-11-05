# API базы данных

Синхронный `db`: `dialect`, `query(sql, params?)`, `exec(sql, params?)`. [Типы](../assets/migration-api.d.ts).

```js
export function up(db) {
  const rows = db.query("SELECT id, name FROM users WHERE name=?", ["Alice"]);
  for (const row of rows) {
    const result = db.exec("UPDATE users SET name=? WHERE id=?", ["alice", row.id]);
    // result.rowsAffected — десятичная строка, например "1".
  }
}
```

query загружает первый набор целиком в память; повторяющимся столбцам задайте alias. exec возвращает `{ rowsAffected: string }` с исходным счётчиком драйвера, даже `"0"`.

Массив параметров: PostgreSQL/Redshift `$1`; MySQL/MariaDB/SQLite/TiDB/ClickHouse `?`; Oracle `:1`; MSSQL `@p1`. `db.dialect` — каноническое имя.

Целые и точные десятичные — строки, NULL — `null`, бинарные — копии массивов байтов, время — точные строки UTC. JavaScript number и `JSON.parse` могут округлять. Десятичные SQLite, сохранённые float, отклоняются. См. [особые типы](./databases#value-conversion).

SQL передаётся без разделения и обработки `GO`; сохраняйте терминаторы, включая `;` у MERGE. Пакетам MySQL могут требоваться опции DSN. Очищенные ошибки драйвера перехватываются, но catch не чинит транзакцию: используйте savepoint/rollback или повторно бросайте ошибку. Нормальный возврат принимает результат. Отмена, общий таймаут, неверные значения и лимиты остаются фатальными даже после catch и запрещают дальнейшие DB-вызовы. Таймаут только драйвера можно перехватить. См. [транзакции](./script-transactions).
