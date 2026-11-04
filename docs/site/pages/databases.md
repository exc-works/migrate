# Database capabilities

## SQL and full builds

```sh
# SQL build
go install github.com/exc-works/migrate/cmd/migrate@latest

# Full build: SQL, JS, TS and MJS
go install -tags=full github.com/exc-works/migrate/cmd/migrate@latest
migrate capabilities
migrate capabilities --output json
```

Release downloads without a suffix are SQL-only; `_full` assets include scripts.
Both builds run without downloading an interpreter. npm imports require packages
to be installed beforehand; script execution does not provide Node.js APIs.

## Inspect your binary

SQL dialects include PostgreSQL, MySQL, MariaDB, SQLite, Oracle, MSSQL, ClickHouse,
TiDB and Redshift. SQL availability and script verification are separate capabilities.
Inspect the output of your binary before choosing a build or database:

- `formats` lists the migration formats present in the build.
- `script_dialect_capabilities` reports script verification, transaction modes and limitations.
- `aliases` lists accepted names; configurations normalize to the canonical dialect.

## Script database support {#script-support}

This table describes the default script execution boundary, not a guarantee that
every statement can be rolled back. Experimental support has limited coverage of
database versions, types and deployment configurations. Skipped external checks do
not establish real-instance verification.

| Database | Script status | Placeholder | Execution boundary |
| --- | --- | --- | --- |
| PostgreSQL | Verified | `$1` | Framework transaction |
| MySQL / MariaDB | Verified | `?` | Requires transactional tables; DDL may implicitly commit |
| SQLite | Verified | `?` | Framework transaction |
| TiDB | Experimental | `?` | CI uses Unistore; distributed TiKV unverified; DDL may implicitly commit |
| Redshift | Experimental; no real-instance verification | `$1` | Statement-dependent transactions; not full PostgreSQL equivalence |
| Oracle | Experimental; limited Oracle Free coverage | `:1` | Framework transaction for DML; DDL implicitly commits |
| ClickHouse | Experimental | `?` | Nontransactional; no automatic rollback |
| MSSQL / SQL Server | Experimental | `@p1` | Framework transaction; Azure SQL instances unverified |

Selected experimental script plans warn once, including dry-run. Status, baseline
and SQL-only plans do not emit that warning.

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
then use `migrate capabilities` to check your installed build's limits.

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
