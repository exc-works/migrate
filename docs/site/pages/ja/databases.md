# 対応データベース

SQL は PostgreSQL、MySQL、MariaDB、SQLite、Oracle、MSSQL、ClickHouse、TiDB、Redshift に対応します。`dialect` で種類を選びます。

## スクリプトとトランザクション {#script-support}

| データベース | 対応状況 | パラメーター | トランザクション |
| --- | --- | --- | --- |
| PostgreSQL | 検証済み | `$1` | migrate が管理 |
| MySQL / MariaDB | 検証済み | `?` | トランザクション対応テーブルが必要。DDL は暗黙コミットの場合あり |
| SQLite | 検証済み | `?` | migrate が管理 |
| TiDB | 実験的 | `?` | 分散 TiKV は未検証、DDL は暗黙コミットの場合あり |
| Redshift | 実験的、実インスタンス未検証 | `$1` | 文に依存。PostgreSQL と完全互換ではない |
| Oracle | 実験的、Oracle Free の限定検証 | `:1` | DML は管理、DDL は暗黙コミット |
| ClickHouse | 実験的 | `?` | 非トランザクション、自動ロールバックなし |
| MSSQL | 実験的、Azure SQL 未検証 | `@p1` | migrate が管理 |

実験的なスクリプトはプレビューでも通知します。バージョンや構成の検証範囲は限られるため、対象環境で確認してください。[トランザクション](./script-transactions)も参照してください。

## 名前と別名 {#aliases}

大文字小文字と前後の空白は区別しません。`db.dialect` は正規名を返し、MySQL、MariaDB、TiDB は別の種類です。

| 正規名 | 別名 |
| --- | --- |
| `postgres` | `pg`、`pgsql`、`postgresql`、`pgx` |
| `mssql` | `sqlserver`、`azuresql` |
| `oracle` | `orcl`、`ora`、`go-ora`、`goora`、`godror` |
| `sqlite` | `sqlite3` |
| `mysql`、`mariadb`、`tidb`、`redshift`、`clickhouse` | 追加なし |

## 固有の値 {#value-conversion}

精密な整数と小数は文字列で保持します。

- MSSQL の MONEY は文字列、UNIQUEIDENTIFIER はバイナリです。文字列が必要なら `CAST(id AS VARCHAR(36))` を使います。rowversion/timestamp はバイナリです。
- Oracle の NUMBER は文字列、RAW はバイト、CLOB は文字列。メタデータのない BLOB はサイズを制限した `DBMS_LOB.SUBSTR` または `RAWTOHEX` で変換します。自前接続には `FAST+LOGIN=false` を指定すると再接続が安定します。CLI は明示指定がなければこの設定を使用します。
- Redshift は pgx が型を認識できない場合に明示変換します。VARBYTE は `FROM_HEX($1)` で書き、`FROM_VARBYTE(payload, 'hex')` で読みます。
- ClickHouse の大きい整数と小数は文字列です。Array/Tuple/Map/UUID/IP は対応する単純型に変換。String は有効な UTF-8 が必要で、任意バイナリは hex にします。mutation の完了待ちは `SETTINGS mutations_sync=2`。`rowsAffected` が `"0"` の場合があります。

```sh
migrate capabilities
```

接続せずにインストール済みバージョンの対応状況を表示します。自動処理には `--output json` を使用できます。
