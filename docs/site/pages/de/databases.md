# Unterstützte Datenbanken

SQL unterstützt PostgreSQL, MySQL, MariaDB, SQLite, Oracle, MSSQL, ClickHouse, TiDB und Redshift. Auswahl über `dialect`.

## Skripte und Transaktionen {#script-support}

| Datenbank | Status | Parameter | Transaktion |
| --- | --- | --- | --- |
| PostgreSQL | Verifiziert | `$1` | Von migrate verwaltet |
| MySQL / MariaDB | Verifiziert | `?` | Transaktionale Tabellen; implizite DDL-Commits möglich |
| SQLite | Verifiziert | `?` | Von migrate verwaltet |
| TiDB | Experimentell | `?` | Verteiltes TiKV nicht verifiziert; implizite DDL-Commits möglich |
| Redshift | Experimentell, keine reale Instanz verifiziert | `$1` | Anweisungsabhängig, nicht vollständig PostgreSQL |
| Oracle | Experimentell, begrenzte Oracle-Free-Prüfung | `:1` | DML verwaltet, DDL mit implizitem Commit |
| ClickHouse | Experimentell | `?` | Keine Transaktion oder automatische Rücknahme |
| MSSQL | Experimentell, Azure SQL ungeprüft | `@p1` | Von migrate verwaltet |

Experimentelle Skripte warnen auch im Dry-run. Im Zielsystem prüfen; siehe [Transaktionen](./script-transactions).

## Namen und Aliase {#aliases}

Groß-/Kleinschreibung und äußere Leerzeichen werden ignoriert. `db.dialect` ist kanonisch. MySQL, MariaDB und TiDB bleiben verschieden.

| Name | Aliase |
| --- | --- |
| `postgres` | `pg`, `pgsql`, `postgresql`, `pgx` |
| `mssql` | `sqlserver`, `azuresql` |
| `oracle` | `orcl`, `ora`, `go-ora`, `goora`, `godror` |
| `sqlite` | `sqlite3` |
| `mysql`, `mariadb`, `tidb`, `redshift`, `clickhouse` | Keine weiteren |

## Besondere Werte {#value-conversion}

Exakte Ganz- und Dezimalzahlen bleiben Zeichenfolgen.

- MSSQL: MONEY exakt, UNIQUEIDENTIFIER binär; für Text `CAST(id AS VARCHAR(36))`. rowversion/timestamp ist binär.
- Oracle: NUMBER Text, RAW Bytes, CLOB Text. BLOB ohne Metadaten mit begrenztem `DBMS_LOB.SUBSTR` oder `RAWTOHEX` projizieren. Eigene Verbindungen sollten `FAST+LOGIN=false` verwenden; CLI nutzt dies ohne explizite Gegenoption.
- Redshift: Unbekannte pgx-Typen umwandeln. VARBYTE mit `FROM_HEX($1)` schreiben und `FROM_VARBYTE(payload, 'hex')` lesen.
- ClickHouse: Große Zahlen und Dezimalwerte als Text. Array/Tuple/Map/UUID/IP in Skalare umwandeln. String braucht UTF-8, beliebige Binärdaten Hex. Für Mutation-Abschluss `SETTINGS mutations_sync=2`; `rowsAffected` kann `"0"` sein.

```sh
migrate capabilities
```

Unterstützung der Installation ohne Verbindung prüfen; JSON mit `--output json`.
