# Bases compatibles

SQL admite PostgreSQL, MySQL, MariaDB, SQLite, Oracle, MSSQL, ClickHouse, TiDB y Redshift. Selecciona mediante `dialect`.

## Scripts y transacciones {#script-support}

| Base | Estado | Parámetro | Transacción |
| --- | --- | --- | --- |
| PostgreSQL | Verificado | `$1` | Gestionada por migrate |
| MySQL / MariaDB | Verificado | `?` | Tablas transaccionales; DDL puede hacer commit implícito |
| SQLite | Verificado | `?` | Gestionada por migrate |
| TiDB | Experimental | `?` | TiKV distribuido no verificado; DDL puede hacer commit implícito |
| Redshift | Experimental, sin instancia real verificada | `$1` | Según sentencia; no equivale a PostgreSQL completo |
| Oracle | Experimental, cobertura Oracle Free limitada | `:1` | DML transaccional; DDL hace commit implícito |
| ClickHouse | Experimental | `?` | Sin transacción ni reversión automática |
| MSSQL | Experimental, Azure SQL no verificado | `@p1` | Gestionada por migrate |

Los scripts experimentales avisan también en dry-run. Verifica en tu entorno y consulta [transacciones](./script-transactions).

## Nombres y alias {#aliases}

No distingue mayúsculas ni espacios exteriores. `db.dialect` devuelve el canónico. MySQL, MariaDB y TiDB son tipos distintos.

| Nombre | Alias |
| --- | --- |
| `postgres` | `pg`, `pgsql`, `postgresql`, `pgx` |
| `mssql` | `sqlserver`, `azuresql` |
| `oracle` | `orcl`, `ora`, `go-ora`, `goora`, `godror` |
| `sqlite` | `sqlite3` |
| `mysql`, `mariadb`, `tidb`, `redshift`, `clickhouse` | Ninguno más |

## Tipos particulares {#value-conversion}

Enteros y decimales exactos se conservan como cadenas.

- MSSQL: MONEY exacto, UNIQUEIDENTIFIER binario; usa `CAST(id AS VARCHAR(36))` para texto. rowversion/timestamp es binario.
- Oracle: NUMBER texto, RAW bytes, CLOB texto. Para BLOB sin metadatos usa `DBMS_LOB.SUBSTR` limitado o `RAWTOHEX`. En conexiones propias, `FAST+LOGIN=false` facilita la reconexión; la CLI lo usa si no hay opción explícita.
- Redshift: convierte tipos desconocidos de pgx. VARBYTE se escribe con `FROM_HEX($1)` y se lee con `FROM_VARBYTE(payload, 'hex')`.
- ClickHouse: enteros grandes y decimales como texto; convierte Array/Tuple/Map/UUID/IP a escalares. String requiere UTF-8; binarios arbitrarios en hexadecimal. Espera mutaciones con `SETTINGS mutations_sync=2`. `rowsAffected` puede ser `"0"`.

```sh
migrate capabilities
```

Consulta la versión instalada sin conectar; usa `--output json` para automatizar.
