# 在 Go 中使用 migrate

導入根包即可在測試、應用啓動或管理操作中執行遷移。
公開 API 遵循 SemVer；`internal/*` 不屬於公開 API。

## 安裝並註冊驅動

```sh
go get github.com/exc-works/migrate
```

調用方自行創建 `*sql.DB` 時，需要註冊對應驅動。下面的完整示例使用 SQLite，
從 `./migrations` 讀取 SQL 檔案：

```go
package main

import (
    "context"
    "database/sql"
    "log"

    "github.com/exc-works/migrate"
    _ "modernc.org/sqlite"
)

func main() {
    if err := run(context.Background()); err != nil {
        log.Fatal(err)
    }
}

func run(ctx context.Context) error {
    db, err := sql.Open("sqlite", "./app.sqlite")
    if err != nil {
        return err
    }
    defer db.Close()

    service, err := migrate.NewService(ctx, migrate.Config{
        Dialect: migrate.NewSQLiteDialect(),
        DB:      db,
        MigrationSource: migrate.DirectorySource{
            Directory: "./migrations",
        },
    })
    if err != nil {
        return err
    }
    if err := service.Create(); err != nil {
        return err
    }
    return service.Up()
}
```

資料庫連接池及生命週期由調用方管理。Context 應採用符合部署流程的超時，
同一資料庫上的遷移應串行執行。

## 選擇來源與方言

| 來源 | 用途 |
| --- | --- |
| `DirectorySource` | 磁盤版本檔案 |
| `StringSource` | 內存中的遷移文本，尤其適合測試 |
| `FSSource` | 任意 `fs.FS`，包括 `embed.FS` 與測試檔案系統 |
| `CombinedSource` | 合併多個來源 |

使用 `NewPostgresDialect()`、`NewSQLiteDialect()` 等構造函數，
也可用 `DialectFromName("postgres")` 按名稱查找。

通過 `//go:embed migrations/*.sql` 嵌入 SQL，再將
`migrate.FSSource{FS: files, Root: "migrations"}` 傳給來源字段。
指令碼 helper 與 npm 包的接入見[共享模塊](./modules)。

## 服務方法

| 方法 | 行為 |
| --- | --- |
| `Create()` | 冪等創建歷史表 |
| `Up()` | 執行待應用版本 |
| `Down(toVersion, all)` | 回退高於目標的版本，或回退全部 |
| `Status()` | 返回 `[]migrate.MigrationStatus` |
| `Baseline()` | 登記待執行版本，不執行內容 |

設置 `Config.DryRun` 和 `Config.DryRunOutput` 可預覽使用者遷移。
`Create()` 不受 dry-run 影響，仍會創建歷史表。默認記錄器為 `NoopLogger`，
需要時使用 `NewStdLogger("info", os.Stdout)` 或自行實現 `Logger`。

## 指令碼與工具管理的連接

通過 `Config.ScriptOptions` 設置指令碼限制。
[嵌入示例](./examples)演示將指令碼與 helper 一起編入 Go 二進制。

需要 migrate 管理連接（包括可選 SSH 隧道）時，使用 `OpenConnection`。
先通過 `Connection.Context()` 對 `Connection.DB` 執行 Ping，再把 DB 和 context
傳給 `NewService`，最後關閉 connection。該 helper 註冊網絡驅動，SQLite 仍需調用方註冊。
庫憑據輸入通過 `ConnectionOptions.Prompt` 顯式啓用，主機密鑰批准使用獨立的 `HostKeyPrompt`。
空回調不提示、不信任未知密鑰，兩者都必須響應取消。參見 [SSH 隧道](./ssh-tunnels)。

更多可運行 API 示例見
[example_test.go](https://github.com/exc-works/migrate/blob/main/example_test.go)。
