# Bases compatibles

SQL prend en charge PostgreSQL, MySQL, MariaDB, SQLite, Oracle, MSSQL, ClickHouse, TiDB et Redshift. Choisissez avec `dialect`.

## Scripts et transactions {#script-support}

| Base | État | Paramètre | Transaction |
| --- | --- | --- | --- |
| PostgreSQL | Vérifié | `$1` | Gérée par migrate |
| MySQL / MariaDB | Vérifié | `?` | Tables transactionnelles ; DDL peut committer implicitement |
| SQLite | Vérifié | `?` | Gérée par migrate |
| TiDB | Expérimental | `?` | TiKV distribué non vérifié ; commits DDL implicites possibles |
| Redshift | Expérimental, sans vérification d’instance réelle | `$1` | Dépend des instructions ; différent de PostgreSQL |
| Oracle | Expérimental, couverture Oracle Free limitée | `:1` | DML transactionnel, DDL avec commit implicite |
| ClickHouse | Expérimental | `?` | Sans transaction ni annulation automatique |
| MSSQL | Expérimental, Azure SQL non vérifié | `@p1` | Gérée par migrate |

Une migration expérimentale affiche un avertissement, même en aperçu. Vérifiez dans votre environnement. Voir [transactions](./script-transactions).

## Noms et alias {#aliases}

Insensibles à la casse et aux espaces extérieurs. `db.dialect` renvoie le nom canonique. MySQL, MariaDB et TiDB restent distincts.

| Nom | Alias |
| --- | --- |
| `postgres` | `pg`, `pgsql`, `postgresql`, `pgx` |
| `mssql` | `sqlserver`, `azuresql` |
| `oracle` | `orcl`, `ora`, `go-ora`, `goora`, `godror` |
| `sqlite` | `sqlite3` |
| `mysql`, `mariadb`, `tidb`, `redshift`, `clickhouse` | Aucun autre |

## Types particuliers {#value-conversion}

Les entiers et décimaux exacts restent des chaînes.

- MSSQL : MONEY est exact, UNIQUEIDENTIFIER binaire ; utilisez `CAST(id AS VARCHAR(36))` pour du texte. rowversion/timestamp est binaire.
- Oracle : NUMBER texte, RAW octets, CLOB texte. Pour BLOB sans métadonnées, projetez `DBMS_LOB.SUBSTR` borné ou `RAWTOHEX`. Pour une connexion propre, `FAST+LOGIN=false` améliore la reconnexion ; la CLI l’emploie sauf option explicite.
- Redshift : convertissez les types inconnus de pgx. VARBYTE s’écrit avec `FROM_HEX($1)` et se lit avec `FROM_VARBYTE(payload, 'hex')`.
- ClickHouse : grands entiers et décimaux en texte ; convertissez Array/Tuple/Map/UUID/IP en scalaires. String exige UTF-8 ; les binaires arbitraires passent en hexadécimal. Pour attendre une mutation : `SETTINGS mutations_sync=2`. `rowsAffected` peut être `"0"`.

```sh
migrate capabilities
```

Affiche la prise en charge de la version installée sans connexion ; JSON via `--output json`.
