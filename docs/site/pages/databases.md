# Supported databases

SQL dialects include PostgreSQL, MySQL, MariaDB, SQLite, Oracle, MSSQL, ClickHouse,
TiDB and Redshift. Set `dialect` in your configuration to choose a database.
Script and transaction support varies by database, as shown below.

## Script database support {#script-support}

Experimental script support does not cover every database version or deployment.
Validate migrations in your target environment before use. The transaction column describes
automatic recovery from script failures, not a guarantee that every SQL statement can be rolled back.

| Database | Script status | Placeholder | Transaction behavior |
| --- | --- | --- | --- |
| PostgreSQL | Verified | `$1` | migrate-managed transaction |
| MySQL / MariaDB | Verified | `?` | Requires transactional tables; DDL may implicitly commit |
| SQLite | Verified | `?` | migrate-managed transaction |
| TiDB | Experimental | `?` | DDL may implicitly commit; distributed TiKV deployments unverified |
| Redshift | Experimental; no real-instance verification | `$1` | Statement-dependent transactions; not full PostgreSQL equivalence |
| Oracle | Experimental; limited Oracle Free coverage | `:1` | migrate-managed transaction for DML; DDL implicitly commits |
| ClickHouse | Experimental | `?` | Nontransactional; no automatic rollback |
| MSSQL / SQL Server | Experimental | `@p1` | migrate-managed transaction; Azure SQL instances unverified |

Executing or previewing scripts for an experimental database displays a warning.

## Accepted names {#aliases}

Names are trimmed and case-insensitive. Configuration and `db.dialect` normalize
aliases to the canonical name. MySQL, MariaDB and TiDB remain distinct dialects.

| Canonical name | Aliases |
| --- | --- |
| `postgres` | `pg`, `pgsql`, `postgresql`, `pgx` |
| `mssql` | `sqlserver`, `azuresql` |
| `oracle` | `orcl`, `ora`, `go-ora`, `goora`, `godror` |
| `sqlite` | `sqlite3` |
| `mysql`, `mariadb`, `tidb`, `redshift`, `clickhouse` | No extra aliases |

## Transactions and rollback

Database engines differ: some DDL implicitly commits, and ClickHouse does not offer
the framework's automatic rollback. An explicitly nontransactional script can leave
partial changes without successful history if it fails. Inspect data and history before retrying.

Review [transaction modes](./script-transactions) and the database matrix,
then use `migrate capabilities` to check execution limits.

## Vendor-specific script values {#value-conversion}

The [database API](./script-api) preserves exact integers and decimals as strings.
Some vendor types need additional handling:

- **MSSQL:** MONEY/SMALLMONEY stay exact strings. UNIQUEIDENTIFIER keeps the driver's
  copied binary representation; use `CAST(id AS VARCHAR(36))` for formatted text.
  Rowversion/timestamp is binary; time values follow the common conversion contract.
- **Oracle:** NUMBER is exact text, RAW is bytes and CLOB is text. For BLOB without
  type metadata, project bounded `DBMS_LOB.SUBSTR` to RAW or `RAWTOHEX` to text.
  Caller-created DBs should set `FAST+LOGIN=false` for reliable go-ora v3 reconnection;
  CLI connections disable fast-login cookies by default unless explicitly configured.
- **Redshift:** explicitly cast types with unknown pgx metadata to supported scalars.
  For VARBYTE, write hex using `FROM_HEX($1)` and read with
  `FROM_VARBYTE(payload, 'hex')`.
- **ClickHouse:** 128/256-bit integers and decimals stay exact strings. Convert
  Array/Tuple/Map/UUID/IP to supported scalars. String must be valid UTF-8; use hex
  for arbitrary binary data. Use statement-local `SETTINGS mutations_sync=2` when
  mutation completion is required. Driver-reported `rowsAffected` may be `"0"`.

## Check support in your installed version

```sh
migrate capabilities
```

This lists supported file formats, database names and script transaction details without a database connection.
Use `--output json` for automation.
