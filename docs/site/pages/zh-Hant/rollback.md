# 回退遷移

回退執行遷移提供的 Down SQL 或 `down(db)`。
它只能撤銷這些代碼實際實現的操作，無法自行恢復已經刪除的資料。

## 保留目標版本

```sh
migrate down 202611030001 --dry-run
migrate down 202611030001
migrate status
```

只有高於目標的已應用版本會被回退，目標版本保持已應用。
例如 1、2、3 均已應用時，`migrate down 2` 只回退版本 3，保留版本 1 和 2。

## 回退全部已應用版本

```sh
migrate down --all --dry-run
migrate down --all
migrate status
```

目標版本和 `--all` 互斥。成功回退後，對應檔案恢復為待執行狀態，之後可以再次應用。

## 檢查整個回退範圍

檢查所有選中遷移的 Down 實現，以及資料庫的事務限制。
指令碼回退會在寫入前準備整個選中範圍；缺少 `down(db)` 會阻斷該範圍。
空的 down 函數並不能撤銷 up 的變更。

DDL 隱式提交與非事務變更可能在回退失敗後仍然保留。
恢復前應先檢查資料和歷史，參見[事務與執行限制](./script-transactions)及[問題排查](./troubleshooting)。
