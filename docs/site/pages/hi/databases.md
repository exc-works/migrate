# समर्थित डेटाबेस

SQL: PostgreSQL, MySQL, MariaDB, SQLite, Oracle, MSSQL, ClickHouse, TiDB, Redshift। dialect से चुनें।

## स्क्रिप्ट और ट्रांज़ैक्शन {#script-support}

| DB | स्थिति | पैरामीटर | ट्रांज़ैक्शन |
| --- | --- | --- | --- |
| PostgreSQL | सत्यापित | `$1` | migrate प्रबंधित |
| MySQL / MariaDB | सत्यापित | `?` | ट्रांज़ैक्शन तालिका; DDL अप्रत्यक्ष commit कर सकता है |
| SQLite | सत्यापित | `?` | migrate प्रबंधित |
| TiDB | प्रयोगात्मक | `?` | वितरित TiKV असत्यापित; DDL commit संभव |
| Redshift | प्रयोगात्मक, असली instance असत्यापित | `$1` | कथन पर निर्भर, पूरा PostgreSQL नहीं |
| Oracle | प्रयोगात्मक, सीमित Oracle Free | `:1` | DML प्रबंधित, DDL अप्रत्यक्ष commit |
| ClickHouse | प्रयोगात्मक | `?` | ट्रांज़ैक्शन/स्वतः rollback नहीं |
| MSSQL | प्रयोगात्मक, Azure SQL असत्यापित | `@p1` | migrate प्रबंधित |

प्रयोगात्मक स्क्रिप्ट की पूर्वावलोकन में भी चेतावनी आती है। अपने परिवेश में जाँचें। [ट्रांज़ैक्शन](./script-transactions) देखें।

## नाम और उपनाम {#aliases}

बड़े/छोटे अक्षर और किनारे के स्पेस अनदेखे। db.dialect मानक नाम है। MySQL, MariaDB, TiDB अलग हैं।

| नाम | उपनाम |
| --- | --- |
| `postgres` | `pg`, `pgsql`, `postgresql`, `pgx` |
| `mssql` | `sqlserver`, `azuresql` |
| `oracle` | `orcl`, `ora`, `go-ora`, `goora`, `godror` |
| `sqlite` | `sqlite3` |
| `mysql`, `mariadb`, `tidb`, `redshift`, `clickhouse` | और नहीं |

## विशेष मान {#value-conversion}

सटीक पूर्णांक और दशमलव स्ट्रिंग रहते हैं।

- MSSQL: MONEY सटीक, UNIQUEIDENTIFIER बाइनरी; पाठ के लिए `CAST(id AS VARCHAR(36))`। rowversion/timestamp बाइनरी।
- Oracle: NUMBER पाठ, RAW बाइट, CLOB पाठ। metadata बिना BLOB को सीमित `DBMS_LOB.SUBSTR` या RAWTOHEX से लें। अपने कनेक्शन पर `FAST+LOGIN=false` दोबारा कनेक्शन के लिए दें; CLI स्पष्ट विकल्प न होने पर यही रखता है।
- Redshift: pgx के अनजान प्रकार बदलें। VARBYTE लिखने में `FROM_HEX($1)`, पढ़ने में `FROM_VARBYTE(payload, 'hex')`।
- ClickHouse: बड़े अंक/दशमलव पाठ; Array/Tuple/Map/UUID/IP को scalar बनाएँ। String UTF-8 है, मनमाने बाइनरी hex। mutation प्रतीक्षा `SETTINGS mutations_sync=2`। rowsAffected `"0"` हो सकता है।

```sh
migrate capabilities
```

बिना कनेक्शन स्थापित समर्थन देखें; स्वचालन में --output json।
