# 指令碼事務與執行限制

## 默認事務行為

在支持事務的資料庫上，migrate 默認把一條指令碼及其成功記錄放在同一個事務中。
指令碼成功後提交；執行失敗時嘗試回滾。默認模式下不要自行執行 `BEGIN`、`COMMIT` 或 `ROLLBACK`。

建表、改表等 DDL 在部分資料庫上會隱式提交，ClickHouse 等資料庫不支持事務，
因此失敗時可能仍有變更保留。請先查看[資料庫事務說明](./databases#script-support)。

## 需要自己管理事務時

對於必須在事務之外執行的 SQL，導出以下設定：

```js
export const config = { transaction: false };

export function up(db) {
  // PostgreSQL：該語句不能在事務塊內執行。
  db.exec("CREATE INDEX CONCURRENTLY users_name_idx ON users (name)");
}

export function down(db) {
  db.exec("DROP INDEX CONCURRENTLY users_name_idx");
}
```

設定同時作用於 up 和 down。省略 config/transaction 時默認為 true；
`transaction` 只接受 `true` 或 `false`。`status`、`baseline` 和 dry-run 不運行指令碼，
因此不會檢查運行時導出的設定是否有效。

關閉自動事務後，同一條遷移中的資料庫操作使用同一個連接。
如果自行開啓事務，必須在函數返回前提交或回滾；修改的會話設置也應恢復。
不要依賴下一條遷移保留上一條遷移的會話狀態。

非事務失敗可能留下已提交變更，卻沒有成功歷史。重試前先檢查資料和歷史；
非事務引擎上的 Down 同樣沒有自動回退保證。

## 限制執行時間與結果大小

在項目設定中增加以下可選字段：

```yaml
script_timeout: 5m # 每條指令碼的執行超時
script_max_rows: 1000 # 單次查詢最多返回的行數
script_max_bytes: 16777216 # 單次查詢結果大小上限，單位為字節（16 MiB）
```

超限會失敗，不會截斷結果。大量讀取應顯式分頁，但分頁不會自動產生分批提交或可恢復遷移。
取消命令會中斷指令碼並請求取消資料庫操作；資料庫可能需要一些時間才能結束正在執行的語句。

Go 調用方通過 `migrate.ScriptOptions` 設置 `Config.ScriptOptions`。參見[在 Go 中使用](./go-library)。
