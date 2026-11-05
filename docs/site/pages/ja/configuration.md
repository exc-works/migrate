# 設定

プロジェクトのディレクトリで設定ファイルを生成し、`migration_config.yml` を編集します。

```sh
migrate new config
```

## データベース接続

接続文字列を環境変数で渡す PostgreSQL の設定例です。

```yaml
# データベースの種類
dialect: postgres
# 実行前に DATABASE_DSN を設定
data_source_name: "${DATABASE_DSN}"
# マイグレーションの保存先
migration_source: migrations
# 適用済みバージョンの履歴テーブル
schema_name: migration_schema
```

| 項目 | 用途 | 省略時 |
| --- | --- | --- |
| `dialect` | データベースの種類 | `postgres` |
| `data_source_name` | 単一データベースの接続文字列 | 単一接続では必須 |
| `data_sources` | 名前付き接続。上の項目とは併用不可 | 無効 |
| `migration_source` | ファイルの保存先 | `migrations` |
| `schema_name` | 履歴テーブル名 | `migration_schema` |
| `working_directory` | ファイル検索の基準 | 現在のディレクトリ |
| `migrate_out_of_order` | 適用済みより古い未適用バージョンも実行 | `false` |
| `logger_level` | `debug`、`info`、`warn`、`error` | `info` |

### 接続文字列の例

ホスト、ポート、データベース名、Oracle サービス名を置き換え、`DB_USER` と `DB_PASSWORD` を設定します。SQLite はファイルパスを指定します。

::: code-group

```yaml [PostgreSQL]
dialect: postgres
data_source_name: "postgres://${DB_USER}:${DB_PASSWORD}@db.example:5432/app?sslmode=require"
```

```yaml [MySQL]
dialect: mysql
data_source_name: "${DB_USER}:${DB_PASSWORD}@tcp(db.example:3306)/app?parseTime=true"
```

```yaml [MariaDB]
dialect: mariadb
data_source_name: "${DB_USER}:${DB_PASSWORD}@tcp(db.example:3306)/app?parseTime=true"
```

```yaml [SQLite]
dialect: sqlite
data_source_name: "./demo.sqlite"
```

```yaml [Oracle]
dialect: oracle
data_source_name: "oracle://${DB_USER}:${DB_PASSWORD}@db.example:1521/FREEPDB1"
```

```yaml [SQL Server]
dialect: mssql
data_source_name: "sqlserver://${DB_USER}:${DB_PASSWORD}@db.example:1433?database=app"
```

```yaml [ClickHouse]
dialect: clickhouse
data_source_name: "clickhouse://${DB_USER}:${DB_PASSWORD}@db.example:9000/app"
```

```yaml [TiDB]
dialect: tidb
data_source_name: "${DB_USER}:${DB_PASSWORD}@tcp(db.example:4000)/app?parseTime=true"
```

```yaml [Redshift]
dialect: redshift
data_source_name: "postgres://${DB_USER}:${DB_PASSWORD}@cluster.example:5439/app?sslmode=require"
```

:::

URL 形式ではユーザー名やパスワードの `@`、`:`、`/` などを URL エンコードしてください。環境変数の展開ではエンコードされません。MySQL、MariaDB、TiDB の `user:password@tcp(...)` は URL 形式ではありません。TLS/SSL は接続先の要件に合わせます。

### ログレベル {#logging}

| 値 | 表示内容 |
| --- | --- |
| `debug` | デバッグ、情報、警告、エラー |
| `info` | 情報、警告、エラー |
| `warn` | 警告、エラー。`warning` も使用可能 |
| `error` | エラーのみ |

```yaml
logger_level: debug # 調査時にデバッグ情報を表示
```

## 環境変数 {#environment-variables}

`${DATABASE_DSN}` は必須変数を読み込み、未設定ならエラーになります。`${DB_NAME:app}` は未設定時に `app` を使用します。同じ YAML を各環境で共有し、認証情報は環境変数やデプロイ基盤のシークレット管理で渡してください。SSH のホスト、ユーザー名、プロキシアドレスも環境変数を使用できます。

## 設定ファイルの検索順序 {#configuration-file-search-order}

作業ディレクトリから次の順で探し、最初に見つかったファイルを使います。

1. `migration_config.yml`
2. `migration_config.yaml`
3. `migration_config.json`

設定は結合しません。見つかったファイルが無効、または読み取り不可ならエラーになります。`-c` を指定すると、そのファイルだけを使用します。

```sh
migrate -c ./configs/dev.yml status
```

## 用途別に設定を生成

| 用途 | コマンド |
| --- | --- |
| 単一接続 | `migrate new config` |
| 別のファイル名 | `migrate new config dev.yml` |
| 複数の接続 | `migrate new config --multi` |
| 4 接続 | `migrate new config --multi --shards 4` |
| SSH Agent | `migrate new config --ssh-tunnel` |
| SSH 秘密鍵 | `migrate new config --ssh-tunnel --ssh-auth private_key` |
| SSH パスワード | `migrate new config --ssh-tunnel --ssh-auth password` |
| SOCKS5 | `migrate new config --ssh-tunnel --proxy socks5` |
| HTTP CONNECT | `migrate new config --ssh-tunnel --proxy http` |

オプションは組み合わせられます。既存のファイルを上書きするには `--force` が必要です。全項目は `migrate new config --help` で確認できます。

```sh
migrate new config shards.yml --multi --shards 4 --ssh-tunnel --proxy socks5
```

## 複数のデータベース {#multiple-databases}

`data_source_name` の代わりに `data_sources` を設定します。データベースの種類、ファイル、履歴テーブル名は共通です。生成時の接続名は `shard_0` から始まり、`SHARD_0_DSN` などを参照します。

```yaml
dialect: postgres
migration_source: migrations
schema_name: migration_schema
data_sources:
  shard_0:
    data_source_name: "${SHARD_0_DSN}" # 最初のデータベースの接続文字列
  shard_1:
    data_source_name: "${SHARD_1_DSN}" # 次のデータベースの接続文字列
```

![移行ファイル → shard_0 / shard_1 → 結果を集約](../assets/diagrams/multiple-databases-ja.svg)

```sh
migrate create
migrate up --dry-run
migrate up
migrate status
```

各データベースは独自の履歴に基づき、未適用のバージョンを順番に実行します。`new version` は一度だけ実行し、ファイルを共有します。既定の同時実行数は 8 です。`--parallel 1` は直列実行になります。

```sh
migrate up --parallel 4
migrate status --output json
```

ひとつが失敗しても他は続行し、最後に接続ごとの結果を表示します。失敗があれば終了コードは非ゼロです。データベース間のトランザクションではないため、他のデータベースの変更は取り消されません。Ctrl+C は新しい処理を止め、実行中の操作もキャンセルします。

共通の `ssh_tunnel` を最上位に設定できます。接続ごとに指定した SSH オブジェクトは共通設定全体を置き換えます。未指定や `null` は共通設定を継承します。[SSH](./ssh-tunnels)も参照してください。同じ実データベースを重複登録せず、同時に別の移行コマンドを実行しないでください。実行中はファイルと依存パッケージを変更せず、失敗時はデータと履歴を確認してから再試行します。

## 作業ディレクトリ

```sh
migrate -w ./deploy up
migrate -w ./deploy -c configs/dev.yml status
```

`-w` から設定ファイルを検索し、設定内の `working_directory` も上書きします。相対的な `migration_source` は作業ディレクトリ基準です。`-c configs/dev.yml` だけでは基準は変わりません。SSH ファイルの相対パスは設定ファイルのディレクトリ基準です。
