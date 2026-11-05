# Bancos compatíveis

SQL oferece PostgreSQL, MySQL, MariaDB, SQLite, Oracle, MSSQL, ClickHouse, TiDB, Redshift. Selecione com `dialect`.

## Scripts e transações {#script-support}

| Banco | Estado | Parâmetro | Transação |
| --- | --- | --- | --- |
| PostgreSQL | Verificado | `$1` | Gerenciada pelo migrate |
| MySQL / MariaDB | Verificado | `?` | Tabelas transacionais; DDL pode fazer commit implícito |
| SQLite | Verificado | `?` | Gerenciada pelo migrate |
| TiDB | Experimental | `?` | TiKV distribuído não verificado; DDL pode fazer commit implícito |
| Redshift | Experimental, sem instância real verificada | `$1` | Depende da instrução, não equivale a PostgreSQL completo |
| Oracle | Experimental, cobertura Oracle Free limitada | `:1` | DML gerenciado; DDL com commit implícito |
| ClickHouse | Experimental | `?` | Sem transação nem reversão automática |
| MSSQL | Experimental, Azure SQL não verificado | `@p1` | Gerenciada pelo migrate |

Scripts experimentais avisam também na prévia. Valide no ambiente alvo. Veja [transações](./script-transactions).

## Nomes e aliases {#aliases}

Ignora maiúsculas e espaços externos. `db.dialect` retorna o nome canônico. MySQL, MariaDB e TiDB são distintos.

| Nome | Aliases |
| --- | --- |
| `postgres` | `pg`, `pgsql`, `postgresql`, `pgx` |
| `mssql` | `sqlserver`, `azuresql` |
| `oracle` | `orcl`, `ora`, `go-ora`, `goora`, `godror` |
| `sqlite` | `sqlite3` |
| `mysql`, `mariadb`, `tidb`, `redshift`, `clickhouse` | Nenhum outro |

## Tipos específicos {#value-conversion}

Inteiros e decimais exatos permanecem strings.

- MSSQL: MONEY exato, UNIQUEIDENTIFIER binário; use `CAST(id AS VARCHAR(36))` para texto. rowversion/timestamp é binário.
- Oracle: NUMBER texto, RAW bytes, CLOB texto. BLOB sem metadados exige `DBMS_LOB.SUBSTR` limitado ou `RAWTOHEX`. Conexões próprias devem usar `FAST+LOGIN=false` para reconectar; a CLI usa isso sem opção explícita.
- Redshift: converta tipos desconhecidos de pgx. VARBYTE usa `FROM_HEX($1)` para escrever e `FROM_VARBYTE(payload, 'hex')` para ler.
- ClickHouse: números grandes e decimais como texto; converta Array/Tuple/Map/UUID/IP em escalares. String exige UTF-8, binários em hexadecimal. Aguarde mutations com `SETTINGS mutations_sync=2`. `rowsAffected` pode ser `"0"`.

```sh
migrate capabilities
```

Consulta suporte instalado sem conexão; use `--output json` para automação.
