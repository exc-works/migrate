# 文檔導覽

migrate 將資料庫變更保存為有序的版本檔案。你可以在部署流程中使用 CLI，
也可以把同一個遷移服務嵌入 Go 應用。

## 完成第一次遷移

從 [SQLite 快速開始](./quick-start)入手，無須啓動資料庫服務，就能完成一條 SQL
遷移的創建、執行和回退。已有項目可以先按照[安裝說明](./installation)安裝工具，
再按照[設定](./configuration)連接資料庫。

## 管理資料庫變更

1. [編寫 SQL 遷移](./sql-migrations)：版本命名、Up/Down 分區與語句塊。
2. [執行遷移](./applying)：預覽計劃、執行並驗證結果。
3. [回退遷移](./rollback)：保留目標版本，或回退全部已應用版本。
4. [歷史與基線](./history)：理解狀態、校驗和，以及如何接管已有資料庫。

## 使用指令碼和集成能力

需要代碼處理的資料轉換可以使用 [JavaScript 或 TypeScript](./scripting)。
[共享模塊](./modules)、[資料庫 API](./script-api)和[事務與執行限制](./script-transactions)
分別介紹代碼復用、資料讀寫，以及失敗時的處理方式。[示例](./examples)提供本地導入、npm 包和嵌入檔案的演練。

應用集成見[在 Go 中使用](./go-library)；資料庫位於堡壘機之後時，
見 [SSH 隧道與Proxy](./ssh-tunnels)。

## 按需查閱

- [命令行參考](./commands)：命令、參數與選項。
- [支持的資料庫](./databases)：驗證狀態、佔位符與引擎限制。
- [問題排查](./troubleshooting)：設定、歷史與執行錯誤。
- [發佈說明](../releases)：特定版本實際交付的功能。

::: info 適用版本
本站跟隨 main 分支。請通過 `migrate version` 與對應發佈說明，確認已安裝版本是否包含所需功能。
:::
