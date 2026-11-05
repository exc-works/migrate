# 編寫指令碼遷移

當資料轉換更適合通過代碼表達時，可以使用 JavaScript 或 TypeScript。
指令碼通過內嵌引擎同步執行，無須 Node.js 運行時。
開始前先[安裝工具](./installation)。

## 創建指令碼

```sh
migrate new version "normalize names" --format ts
```

支持 `js`、`ts` 和 `mjs`；`.mjs` 與 `.js` 的運行時限制相同。
檔案名採用 `V<version>__<description>.<format>`，與 SQL 檔案共用版本序列及歷史。

## 實現入口函數

假設 SQLite 的 `users` 表包含文本字段 `name`：

```ts
export function up(db) {
  for (const row of db.query("SELECT id, name FROM users")) {
    db.exec("UPDATE users SET name=? WHERE id=?", [row.name.trim(), row.id]);
  }
}
```

生成模板中的 up 會拋錯，需先完成實現。入口必須同步執行，不支持 async 函數、
generator 或 Promise 返回值；普通同步返回值會被忽略。
SQL 佔位符隨方言變化，參見[資料庫 API](./script-api)。

只有實現逆操作後才應導出 `down(db)`。去掉首尾空白會丟失信息，因此這個例子沒有
down 函數。包含缺失 down 指令碼的回退範圍會被阻斷；需要可逆轉換時，應先保存原值。

## 預覽並執行

```sh
migrate up --dry-run
migrate up
migrate status
```

Dry-run 檢查指令碼及其導入，並列出待執行檔案，不運行指令碼，因此不會展示動態生成的 SQL。
把資料庫操作放在 `up(db)` 或 `down(db)` 中，不要寫在模塊頂層；不同遷移之間不要依賴全局變數傳遞資料。

## 編輯器提示與指令碼要求

生成的 JS 模板通過 JSDoc 標注 `db`，TS 模板提供等價聲明。轉譯不執行類型檢查。
也可以使用共享的 [migration-api.d.ts](../assets/migration-api.d.ts) 替換局部聲明，
通過 triple-slash 路徑注釋引用。

指令碼可以使用 JavaScript 語法和上面的資料庫 API，但不能調用 Node 內置模塊、
網絡、檔案系統、環境變數或定時器，也不能動態加載模塊。導入 npm 包前，先確認它不依賴這些功能。

接下來閱讀[共享模塊](./modules)、[事務與執行限制](./script-transactions)，
或嘗試[可運行示例](./examples)。
