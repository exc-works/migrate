# migrate ユーザーガイド

## 1. インストール

### 1.1 モジュールからインストール

```bash
go install github.com/exc-works/migrate/cmd/migrate@latest
```

特定バージョンをインストールする場合:

```bash
go install github.com/exc-works/migrate/cmd/migrate@vX.Y.Z
```

`vX.Y.Z` を実際のバージョン（例: `v0.2.3`）に置き換えてください。

### 1.2 ローカルソースからインストール（プライベートリポジトリまたは社内ネットワーク）

リポジトリのルートで実行します:

```bash
go install ./cmd/migrate
```

### 1.3 インストール確認

```bash
migrate --help
```

コマンドが見つからない場合:

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

`Repository not found` と表示される場合は、上記のローカルソースからのインストール手順を使用してください。

## 2. 初期化

### 2.1 設定ファイルを生成

```bash
migrate new config
```

任意:

```bash
migrate new config dev.yml
migrate new config --force
```

デフォルト設定テンプレート:

```yaml
schema_name: migration_schema
dialect: postgres
data_source_name: "host=127.0.0.1 port=5432 user=${DB_USER:postgres} password=${DB_PASSWORD} dbname=${DB_NAME:postgres} sslmode=disable"
working_directory: ""
migrate_out_of_order: false
logger_level: info
migration_source: migrations
```

### 2.2 主要な設定項目を更新

- `dialect`: `postgres`, `mysql`, `mariadb`, `oracle`, `sqlite`, `mssql`, `clickhouse`, `tidb`, `redshift`
- `data_source_name`: DB接続文字列
- `migration_source`: マイグレーションディレクトリ（デフォルト: `migrations`）

### 2.3 マイグレーション履歴テーブルを初期化

```bash
migrate create
```

`create` は出力なしで成功する場合があります。次で確認してください:

```bash
migrate status
```

既存スキーマがあり、古いSQLを再実行したくない場合は、次を使用します:

```bash
migrate baseline
```

## 3. マイグレーションのバージョンファイルを作成

### 3.1 自動生成バージョン

```bash
migrate new version "init users"
```

### 3.2 明示的バージョン

```bash
migrate new version "add email" -v 202604140002
```

生成されるファイル名形式:

```text
V<version>__<description>.sql
```

デフォルトのファイルテンプレート:

```sql
-- +migrate Up

-- +migrate Down
```

## 4. アップグレード（マイグレーション適用）

まずドライラン:

```bash
migrate up --dry-run
```

実際に適用:

```bash
migrate up
```

その後ステータスを確認:

```bash
migrate status
```

`up` は出力なしで成功する場合があります。`status` を信頼できる基準として使用してください。

## 5. ロールバック

### 5.1 対象バージョンまでロールバック（対象バージョン自体は維持）

```bash
migrate down 202604140001
```

意味: `202604140001` より大きい適用済みバージョンのみがロールバックされます。

### 5.2 適用済みバージョンをすべてロールバック

```bash
migrate down --all
```


### 5.3 ドライランでロールバック

```bash
migrate down 202604140001 --dry-run
migrate down --all --dry-run
```

`down` は出力なしで成功する場合があります。確認のため `migrate status` を実行してください。

## 6. ステータス確認

```bash
migrate status
```

機械可読な出力（スクリプトおよびAIエージェントに推奨）:

```bash
migrate status --output json
```

出力列: `Version`, `Filename`, `Hash`, `Status`。

よくあるステータス:

- `pending`
- `applied`
- `baseline`
- `outOfOrder`
- `hashMismatch`
- `filenameMismatch`

## 7. ツールのバージョンを確認

```bash
migrate version
```

## 8. 環境変数テンプレート

`data_source_name` では次をサポートします:

- `${KEY}`: 必須。存在している必要があります
- `${KEY:default}`: `default` は `KEY` が未設定の場合に使用

例:

```yaml
dialect: postgres
data_source_name: "host=127.0.0.1 port=5432 user=${DB_USER:postgres} password=${DB_PASSWORD} dbname=${DB_NAME:app} sslmode=disable"
```

`DB_PASSWORD` が環境変数に設定済みであることを確認してから、次を実行します:

```bash
migrate status
```

## 9. 10分でできる初回デモ（SQLite）

[10分でできる初回デモ（SQLite） · English](site/pages/quick-start.md)

## 10. グローバルフラグ

特定の設定ファイルを使用:

```bash
migrate -c ./configs/dev.yml status
```

特定の作業ディレクトリを使用:

```bash
migrate -w ./deploy create
migrate -w ./deploy up
```

## 11. よくあるエラーとトラブルシューティング

### 11.1 設定ファイルが見つからない

Error: `config file ... no such file or directory`

対処:

- 現在のディレクトリに `migration_config.yml` が存在することを確認
- または `-c` で設定パスを渡す

### 11.2 環境変数が未設定

Error: `can't find env value for XXX`

対処:

- `export XXX=...`
- または `${XXX:default}` を使用

### 11.3 `down` 引数が不完全

Error: `to-version must be set, or use --all`

対処:

- `migrate down <version>` を使用
- または `migrate down --all` を使用

### 11.4 未サポートの dialect

Error: `unsupported dialect: xxx`

対処: 次のいずれかを使用:

- `postgres` `mysql` `mariadb` `oracle` `sqlite` `mssql` `clickhouse` `tidb` `redshift`

### 11.5 マイグレーションメタデータ不一致

Error: `hash mismatch` or `filename mismatch`

対処:

- バージョン管理から元のマイグレーションファイルを復元
- 変更にはより高いバージョンの新しいマイグレーションを作成

## 12. migrate を Go ライブラリとして組み込む

CLI に加えて、`github.com/exc-works/migrate` はサービスコードから直接 import してマイグレーションを実行できます。ユニットテスト、起動フック、管理画面への組み込みに便利です。

### 12.1 インストール

```bash
go get github.com/exc-works/migrate
```

必要なデータベースドライバを import してください（ライブラリは特定のドライバに依存しません）:

```go
import (
    _ "modernc.org/sqlite"             // sqlite
    _ "github.com/jackc/pgx/v5/stdlib" // postgres
    _ "github.com/go-sql-driver/mysql" // mysql / mariadb / tidb
    // ...
)
```

### 12.2 最小例

```go
package main

import (
    "context"
    "database/sql"

    _ "modernc.org/sqlite"

    "github.com/exc-works/migrate"
)

func main() {
    db, err := sql.Open("sqlite", "./app.sqlite")
    if err != nil {
        panic(err)
    }
    defer db.Close()

    svc, err := migrate.NewService(context.Background(), migrate.Config{
        Dialect:         migrate.NewSQLiteDialect(),
        DB:              db,
        MigrationSource: migrate.DirectorySource{Directory: "./migrations"},
    })
    if err != nil {
        panic(err)
    }

    if err := svc.Create(); err != nil { // idempotent: 履歴テーブルが無ければ作成
        panic(err)
    }
    if err := svc.Up(); err != nil {
        panic(err)
    }
}
```

### 12.3 主要 API

- `migrate.NewService(ctx, migrate.Config)` はマイグレーション実行器を構築します
- `svc.Create()` は履歴テーブル `migration_schema` を作成します（idempotent）
- `svc.Up()` は未適用のマイグレーションをすべて適用します
- `svc.Down(toVersion, all)` は指定バージョンまたはすべてをロールバックします
- `svc.Status()` は `[]migrate.MigrationStatus` を返します
- `svc.Baseline()` は既存の pending ファイルを `baseline` としてマークします

よく使う型:

- Dialect（コンストラクタを推奨 — `Dialect` インターフェースを返します）: `migrate.NewPostgresDialect()`, `NewMySQLDialect()`, `NewSQLiteDialect()`, `NewMSSQLDialect()`, `NewOracleDialect()`, `NewClickHouseDialect()`, `NewMariaDBDialect()`, `NewTiDBDialect()`, `NewRedshiftDialect()`、または名前解決用の `migrate.DialectFromName("postgres")`
- Source: `DirectorySource`（ファイルシステム）、`StringSource`（インメモリ）、`FSSource`（任意の `fs.FS`、例えば `//go:embed` や `os.DirFS`）、`CombinedSource`（複数ソースの結合）
- Logger: `migrate.NoopLogger{}`（デフォルト）、`migrate.NewStdLogger("info", os.Stdout)`、または `migrate.Logger` を実装した任意の型

[Go API](../migrate.go) · [Go example](../example_test.go)
