# Поддерживаемые базы

SQL: PostgreSQL, MySQL, MariaDB, SQLite, Oracle, MSSQL, ClickHouse, TiDB, Redshift. Выбор через `dialect`.

## Скрипты и транзакции {#script-support}

| База | Статус | Параметр | Транзакция |
| --- | --- | --- | --- |
| PostgreSQL | Проверено | `$1` | Управляет migrate |
| MySQL / MariaDB | Проверено | `?` | Нужны транзакционные таблицы; DDL может делать неявный commit |
| SQLite | Проверено | `?` | Управляет migrate |
| TiDB | Экспериментально | `?` | Распределённый TiKV не проверен; возможен commit DDL |
| Redshift | Экспериментально, реальный экземпляр не проверен | `$1` | Зависит от SQL, не полный PostgreSQL |
| Oracle | Экспериментально, ограниченный Oracle Free | `:1` | DML управляется, DDL делает commit |
| ClickHouse | Экспериментально | `?` | Без транзакций и автоматического отката |
| MSSQL | Экспериментально, Azure SQL не проверен | `@p1` | Управляет migrate |

Предупреждение об экспериментальном скрипте выводится и в dry-run. Проверяйте целевую среду. См. [транзакции](./script-transactions).

## Имена и псевдонимы {#aliases}

Регистр и крайние пробелы игнорируются. `db.dialect` канонический. MySQL, MariaDB, TiDB различаются.

| Имя | Псевдонимы |
| --- | --- |
| `postgres` | `pg`, `pgsql`, `postgresql`, `pgx` |
| `mssql` | `sqlserver`, `azuresql` |
| `oracle` | `orcl`, `ora`, `go-ora`, `goora`, `godror` |
| `sqlite` | `sqlite3` |
| `mysql`, `mariadb`, `tidb`, `redshift`, `clickhouse` | Нет других |

## Особые значения {#value-conversion}

Точные целые и десятичные сохраняются строками.

- MSSQL: MONEY точный, UNIQUEIDENTIFIER бинарный; текст через `CAST(id AS VARCHAR(36))`. rowversion/timestamp бинарный.
- Oracle: NUMBER — текст, RAW — байты, CLOB — текст. BLOB без метаданных преобразуйте ограниченным `DBMS_LOB.SUBSTR` или `RAWTOHEX`. Свои соединения: `FAST+LOGIN=false` для устойчивого переподключения; CLI использует это без явной настройки.
- Redshift: неизвестные pgx типы преобразуйте. VARBYTE: запись `FROM_HEX($1)`, чтение `FROM_VARBYTE(payload, 'hex')`.
- ClickHouse: большие числа и десятичные — строки; Array/Tuple/Map/UUID/IP преобразуйте в скаляры. String требует UTF-8, произвольные байты — hex. Ожидание mutation: `SETTINGS mutations_sync=2`. `rowsAffected` может быть `"0"`.

```sh
migrate capabilities
```

Сведения об установленной версии без подключения; `--output json` для автоматизации.
