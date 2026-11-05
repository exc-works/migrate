# 編寫 SQL 遷移

每個版本檔案保存一次邏輯變更。migrate 從設定中的 `migration_source` 讀取檔案，
根據版本生成有序的遷移計劃。

## 創建版本檔案

```sh
migrate new version "init users"
migrate new version "add email" -v 202611040002
```

描述中有空格時用引號包住，生成檔案名時空格會替換為下划線。
第一條命令自動生成版本，`-v` 可以指定版本。檔案名採用
`V<version>__<description>.sql`。SQL 與指令碼檔案共用一個版本序列。
已應用的檔案不要修改或重命名，應新增版本完成後續變更。

## 定義 Up 和 Down

```sql
-- +migrate Up
CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);

-- +migrate Down
DROP TABLE users;
```

Up 執行變更，Down 實現逆操作。這個例子的 Down 會刪除表及其中的資料。
編寫遷移時就應確定回退行為，包括丟失的資料能否恢復。

## 保留語句內部的分號

當一條語句內部包含分號，不能按分號拆開執行時，使用 `StatementBegin` 和
`StatementEnd` 包住整條語句。例如 PostgreSQL 函數：

```sql
-- +migrate Up
-- +migrate StatementBegin
CREATE FUNCTION migration_answer() RETURNS integer AS $$
BEGIN
    RETURN 42;
END;
$$ LANGUAGE plpgsql;
-- +migrate StatementEnd

-- +migrate Down
DROP FUNCTION migration_answer();
```

這些標記控制 SQL 解析；實際語法和 DDL 事務行為仍取決於資料庫引擎。
接下來閱讀[執行遷移](./applying)。
