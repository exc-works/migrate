# Go から利用する

ルートパッケージの公開 API は SemVer に従います。`internal/*` は公開 API ではありません。

接続プールの終了処理は呼び出し側の責任です。同じ DB の移行は直列実行し、適切な context の期限を使います。`Config.DryRun` と `DryRunOutput` は変更をプレビューしますが、`Create()` は履歴テーブルを作成します。既定のログは `NoopLogger` です。

`FSSource` は `embed.FS` を使えます。スクリプトの依存は[共有モジュール](./modules)を参照してください。migrate に接続を管理させる場合は `OpenConnection` を使い、`Connection.Context()` で Ping とサービスを呼び、最後に `Close()` します。SQLite のドライバーは呼び出し側で登録します。`ConnectionOptions.Prompt` と `HostKeyPrompt` は明示的に設定し、キャンセルに対応させます。未指定なら入力や未知のホストへの信頼は行いません。

CLI に加えて、`github.com/exc-works/migrate` はサービスコードから直接 import してマイグレーションを実行できます。ユニットテスト、起動フック、管理画面への組み込みに便利です。

## インストール

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

## 最小例

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

## 主要 API

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

[Go API](https://github.com/exc-works/migrate/blob/main/migrate.go) · [Go example](https://github.com/exc-works/migrate/blob/main/example_test.go)
