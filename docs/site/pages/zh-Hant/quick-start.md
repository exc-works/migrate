# SQLite 快速開始

在新的 SQLite 項目中完成一次遷移的創建、執行和回退。

## 安裝

```sh
go install github.com/exc-works/migrate/cmd/migrate@latest
migrate --help
migrate version
```

源碼安裝和 PATH 設置見[安裝](./installation)，
支持的遷移格式見[支持的資料庫](./databases)。

## 設定

```sh
mkdir migrate-demo
cd migrate-demo
migrate new config
```

將生成的 `migration_config.yml` 替換為：

```yaml
# 記錄遷移歷史的表名
schema_name: migration_schema
dialect: sqlite
data_source_name: ./demo.sqlite # 本地 SQLite 檔案
working_directory: .
migrate_out_of_order: false
logger_level: info
migration_source: migrations # 遷移檔案目錄
```

## 創建遷移

```sh
migrate create
migrate new version "init users" -v 202611030001
```

將 `migrations/V202611030001__init_users.sql` 替換為：

```sql
-- +migrate Up
CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);

-- +migrate Down
DROP TABLE users;
```

## 執行並驗證

```sh
migrate up --dry-run
migrate up
migrate status
```

確認版本 `202611030001` 顯示為 `applied`。

## 回退

```sh
migrate down --all --dry-run
migrate down --all
migrate status
```

確認示例表已刪除，版本恢復為 `pending`。

繼續閱讀[編寫 SQL 遷移](./sql-migrations)、[設定說明](./configuration)
或 [JavaScript 和 TypeScript 遷移](./scripting)。
