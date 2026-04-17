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
