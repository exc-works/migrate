# migrate 使用指南

本指南面向首次使用者。命令與旗標以目前實作（`cmd/migrate`）為準。

## 1. 安裝

### 1.1 從 module 安裝

```bash
go install github.com/exc-works/migrate/cmd/migrate@latest
```

安裝指定版本：

```bash
go install github.com/exc-works/migrate/cmd/migrate@vX.Y.Z
```

將 `vX.Y.Z` 替換為實際版本，例如 `v0.2.3`。

### 1.2 從本機原始碼安裝（私有 repo 或內網）

在儲存庫根目錄執行：

```bash
go install ./cmd/migrate
```

### 1.3 驗證安裝

```bash
migrate --help
```

如果找不到命令：

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

如果看到 `Repository not found`，請使用上方的本機原始碼安裝方式。

## 2. 初始化

### 2.1 產生設定檔

```bash
migrate new config
```

可選：

```bash
migrate new config dev.json
migrate new config --force
```

預設設定範本：

```json
{
  "schema_name": "migration_schema",
  "dialect": "postgres",
  "data_source_name": "host=127.0.0.1 port=5432 user=${DB_USER:postgres} password=${DB_PASSWORD:change_me} dbname=${DB_NAME:postgres} sslmode=disable",
  "working_directory": "",
  "migrate_out_of_order": false,
  "logger_level": "info",
  "migration_source": "migrations"
}
```

### 2.2 更新關鍵設定欄位

- `dialect`: `postgres`, `mysql`, `mariadb`, `oracle`, `sqlite`, `mssql`, `clickhouse`, `tidb`, `redshift`
- `data_source_name`: DB 連線字串
- `migration_source`: migration 目錄（預設：`migrations`）

### 2.3 初始化 migration 歷史表

```bash
migrate create
```

`create` 可能會在無輸出的情況下成功。可用以下命令確認：

```bash
migrate status
```

如果你已有既有 schema，且不想重播舊 SQL，請使用：

```bash
migrate baseline
```

## 3. 建立 migration 版本檔案

### 3.1 自動產生版本號

```bash
migrate new version init_users
```

### 3.2 明確指定版本號

```bash
migrate new version add_email -v 202604140002
```

產生的檔名格式：

```text
V<version>__<description>.sql
```

預設檔案範本：

```sql
-- +migrate Up

-- +migrate Down
```

## 4. 升級（套用 migrations）

先做 dry run：

```bash
migrate up --dry-run
```

正式套用：

```bash
migrate up
```

接著檢查狀態：

```bash
migrate status
```

`up` 可能會在無輸出的情況下成功。請以 `status` 作為最終依據。

## 5. 回滾

### 5.1 回滾到目標版本（保留目標版本）

```bash
migrate down 202604140001
```

語義：只會回滾大於 `202604140001` 且已套用的版本。

### 5.2 回滾所有已套用版本

```bash
migrate down --all
```

注意：`migrate down <to-version>` 與 `migrate down --all` 互斥。

### 5.3 Dry-run 回滾

```bash
migrate down 202604140001 --dry-run
migrate down --all --dry-run
```

`down` 可能會在無輸出的情況下成功。請執行 `migrate status` 驗證。

## 6. 檢查狀態

```bash
migrate status
```

機器可讀輸出（建議用於腳本與 AI agents）：

```bash
migrate status --output json
```

輸出欄位：`Version`、`Filename`、`Hash`、`Status`。

常見狀態：

- `pending`
- `applied`
- `baseline`
- `outOfOrder`
- `hashMismatch`
- `filenameMismatch`

## 7. 升級或降級工具本身

### 7.1 升級工具版本

```bash
go install github.com/exc-works/migrate/cmd/migrate@latest
```

### 7.2 降級工具版本

```bash
go install github.com/exc-works/migrate/cmd/migrate@vX.Y.Z
```

範例：

```bash
go install github.com/exc-works/migrate/cmd/migrate@v0.2.3
```

若 repo 為私有，且無法使用 `go install github.com/...@...`，請在原始碼中 checkout 目標版本
並執行：

```bash
go install ./cmd/migrate
```

### 7.3 檢查目前工具版本

```bash
migrate version
```

注意：release artifacts 會印出 release 版本；透過 `go install ./cmd/migrate` 的本機原始碼建置通常會印出 `dev`。

## 8. 環境變數範本

`data_source_name` 支援：

- `${KEY}`: 必填，必須存在
- `${KEY:default}`: 若 `KEY` 缺失則使用 `default`

範例：

```json
{
  "dialect": "postgres",
  "data_source_name": "host=127.0.0.1 port=5432 user=${DB_USER:postgres} password=${DB_PASSWORD} dbname=${DB_NAME:app} sslmode=disable"
}
```

請先確認你的環境中已設定 `DB_PASSWORD`，然後執行：

```bash
migrate status
```

## 9. 10 分鐘首次執行示範（SQLite）

### 9.1 準備目錄與設定

先確認命令可用：

```bash
migrate --help
```

建立示範目錄（macOS/Linux）：

```bash
mkdir -p ./migrate-demo
cd ./migrate-demo
migrate new config
```

Windows PowerShell 對應：

```powershell
mkdir .\migrate-demo
cd .\migrate-demo
migrate new config
```

將 `migration_config.json` 更新為：

```json
{
  "schema_name": "migration_schema",
  "dialect": "sqlite",
  "data_source_name": "./demo.sqlite",
  "working_directory": ".",
  "migrate_out_of_order": false,
  "logger_level": "info",
  "migration_source": "migrations"
}
```

### 9.2 初始化並建立 migration 檔案

```bash
migrate create
migrate new version init_users -v 202604140001
migrate new version add_email -v 202604140002
```

編輯 `migrations/V202604140001__init_users.sql`：

```sql
-- +migrate Up
CREATE TABLE users
(
    id   INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);

-- +migrate Down
DROP TABLE IF EXISTS users;
```

編輯 `migrations/V202604140002__add_email.sql`：

```sql
-- +migrate Up
ALTER TABLE users
    ADD COLUMN email TEXT;

-- +migrate Down
CREATE TABLE users_tmp
(
    id   INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);
INSERT INTO users_tmp (id, name)
SELECT id, name
FROM users;
DROP TABLE users;
ALTER TABLE users_tmp
    RENAME TO users;
```

### 9.3 套用、檢查狀態與回滾

```bash
migrate up --dry-run
migrate up
migrate status
migrate down 202604140001 --dry-run
migrate down 202604140001
migrate status
migrate down --all
migrate status
```

預期：

- `up` 之後：兩個版本都是 `applied`
- `down 202604140001` 之後：`202604140001=applied`、`202604140002=pending`
- `down --all` 之後：兩個版本都是 `pending`

## 10. 全域旗標

使用指定設定檔：

```bash
migrate -c ./configs/dev.json status
```

使用指定工作目錄：

```bash
migrate -w ./deploy create
migrate -w ./deploy up
```

## 11. 常見錯誤與疑難排解

### 11.1 找不到設定檔

錯誤：`config file ... no such file or directory`

修復：

- 確認目前目錄存在 `migration_config.json`
- 或使用 `-c` 傳入設定檔路徑

### 11.2 缺少環境變數

錯誤：`can't find env value for XXX`

修復：

- `export XXX=...`
- 或使用 `${XXX:default}`

### 11.3 `down` 參數不完整

錯誤：`to-version must be set, or use --all`

修復：

- 使用 `migrate down <version>`
- 或使用 `migrate down --all`

### 11.4 不支援的 dialect

錯誤：`unsupported dialect: xxx`

修復：請使用以下其中之一：

- `postgres` `mysql` `mariadb` `oracle` `sqlite` `mssql` `clickhouse` `tidb` `redshift`

### 11.5 Migration 中繼資料不一致

錯誤：`hash mismatch` 或 `filename mismatch`

修復：

- 不要編輯已套用的 migration 檔案
- 變更請建立新且更高版本的 migration

## 12. 作為 Go 函式庫嵌入使用

除了 CLI，`github.com/exc-works/migrate` 也可以作為函式庫直接在你的服務程式碼中觸發 migration，方便撰寫單元測試、接進啟動流程或嵌入管理後台。

### 12.1 安裝

```bash
go get github.com/exc-works/migrate
```

依需要引入資料庫驅動（本函式庫不綁定特定驅動）：

```go
import (
    _ "modernc.org/sqlite"             // sqlite
    _ "github.com/jackc/pgx/v5/stdlib" // postgres
    _ "github.com/go-sql-driver/mysql" // mysql / mariadb / tidb
    // ...
)
```

### 12.2 最小範例

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

    if err := svc.Create(); err != nil { // 冪等：歷史表不存在時建立
        panic(err)
    }
    if err := svc.Up(); err != nil {
        panic(err)
    }
}
```

### 12.3 關鍵 API

- `migrate.NewService(ctx, migrate.Config)` 建立 migration 執行器
- `svc.Create()` 建立 `migration_schema` 歷史表（冪等）
- `svc.Up()` 套用所有未執行的 migration
- `svc.Down(toVersion, all)` 回退至指定版本或全部
- `svc.Status()` 回傳 `[]migrate.MigrationStatus`
- `svc.Baseline()` 將目前的 pending 檔案標記為 `baseline`

常用型別：

- 方言（建議使用建構函式 — 回傳 `Dialect` 介面）：`migrate.NewPostgresDialect()`、`NewMySQLDialect()`、`NewSQLiteDialect()`、`NewMSSQLDialect()`、`NewOracleDialect()`、`NewClickHouseDialect()`、`NewMariaDBDialect()`、`NewTiDBDialect()`、`NewRedshiftDialect()`，或依名稱查找用的 `migrate.DialectFromName("postgres")`
- 來源：`DirectorySource`（讀取檔案系統）、`StringSource`（記憶體陣列）、`FSSource`（任意 `fs.FS`，例如 `//go:embed` 或 `os.DirFS`）、`CombinedSource`（合併多個來源）
- 日誌：`migrate.NoopLogger{}`（預設）、`migrate.NewStdLogger("info", os.Stdout)`，或自行實作 `migrate.Logger` 介面

### 12.4 測試友善：StringSource + 記憶體版 SQLite

```go
src := migrate.StringSource{Migrations: []migrate.SourceFile{{
    Filename: "V1__init.sql",
    Source:   "-- +migrate Up\nCREATE TABLE t(id INT);\n-- +migrate Down\nDROP TABLE t;\n",
}}}

db, _ := sql.Open("sqlite", ":memory:")
svc, _ := migrate.NewService(ctx, migrate.Config{
    Dialect:         migrate.NewSQLiteDialect(),
    DB:              db,
    MigrationSource: src,
})
```

不依賴檔案系統，可直接在單元測試中執行。

### 12.5 用 //go:embed 將 migration 打包進二進位

利用 Go 的 embed 能力，把 migration SQL 直接打進程式的二進位檔：

```go
import "embed"

//go:embed migrations/*.sql
var migrations embed.FS

// 在服務中這樣接入：
// MigrationSource: migrate.FSSource{FS: migrations, Root: "migrations"},
```

`FSSource` 接受任意 `fs.FS`，所以 `os.DirFS` 和 `fstest.MapFS` 也都能用 — 測試時可換成合成檔案系統。

### 12.6 預覽 SQL（DryRun）

```go
var buf bytes.Buffer
svc, _ := migrate.NewService(ctx, migrate.Config{
    Dialect:         migrate.NewPostgresDialect(),
    DB:              db,
    MigrationSource: src,
    DryRun:          true,
    DryRunOutput:    &buf,
})
_ = svc.Create() // Create() 不受 DryRun 影響，用於建立歷史表
_ = svc.Up()     // 使用者 migration SQL 僅寫入 buf，不會實際建表
```

### 12.7 穩定性承諾

- `github.com/exc-works/migrate`（根套件）為對外公開 API，依 SemVer 維護
- `internal/*` 不在穩定性承諾內，請勿直接 import
- 完整可執行範例位於 repository 根目錄下的 `example_test.go`
