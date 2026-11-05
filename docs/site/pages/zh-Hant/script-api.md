# 指令碼資料庫 API

同步入口參數 `db` 提供規範方言名 `dialect`、`query(sql, params?)` 和
`exec(sql, params?)`。編輯器聲明可下載 [migration-api.d.ts](../assets/migration-api.d.ts)。

## 查詢與寫入

下面的例子使用 SQLite 佔位符：

```js
export function up(db) {
  const rows = db.query("SELECT id, name FROM users WHERE name=?", ["Alice"]);
  for (const row of rows) {
    const result = db.exec("UPDATE users SET name=? WHERE id=?", ["alice", row.id]);
    // result.rowsAffected 是十進制字符串，例如 "1"。
  }
}
```

`query` 將行完整讀入內存，只返回第一個結果集；重復列名應使用別名。
`exec` 返回 `{ rowsAffected: string }`，保留驅動原始計數，包括 `"0"`，
不會根據預期變更數推斷替代值。

## 綁定參數與值類型

參數使用數組，SQL 使用方言原生佔位符：PostgreSQL/Redshift 為 `$1`，
MySQL/MariaDB/SQLite/TiDB/ClickHouse 為 `?`，Oracle 為 `:1`，MSSQL 為 `@p1`。
即使設定使用別名，`db.dialect` 也始終是規範方言名。

整數和精確小數結果是字符串，NULL 為 `null`，二進制值是複製的字節數組，
Go 時間戳轉為精確的 UTC 字符串。不要通過 JavaScript number 轉換大整數或精確小數。
`JSON.parse` 也可能對大數值字面量捨入，需要時應選擇無損表示。

SQLite NUMERIC 親和性可能把小數存為浮點數，運行時會拒絕這類值。
應選擇能保持所需精度的存儲類型。資料庫特有類型可能需要顯式 SQL 轉換，
見[資料庫能力](./databases#value-conversion)。

## 原生 SQL 行為

`db.exec` 和 `db.query` 將 SQL 直接交給資料庫，不自動拆分批處理，
也不處理 `GO` 這樣的客戶端分隔符。保留原生終止符，包括 MERGE 必需的分號。
MySQL 批處理可能需要驅動 DSN 選項；實際執行仍受權限和驅動支持限制。

## 處理資料庫錯誤

驅動錯誤轉為可捕獲的脫敏異常，不暴露原始廠商細節或錯誤鏈。
捕獲可恢復錯誤後可以繼續調用資料庫，但捕獲本身不會修復事務。
應按方言使用保存點或回滾；托管事務無法修復時應重新拋出異常。
捕獲後正常返回表示作者接受該結果。

命令被取消、指令碼超時、返回值無法轉換或查詢結果超限後，遷移會失敗；
即使捕獲這些錯誤也不能繼續操作資料庫。單次資料庫操作的超時錯誤可以捕獲。
編寫恢復邏輯前請閱讀[事務與執行限制](./script-transactions)。
