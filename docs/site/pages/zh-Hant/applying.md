# 執行遷移

執行前，確認設定指向預期資料庫，遷移目錄中也只包含本次需要部署的版本。

## 初始化歷史表

```sh
migrate create
```

該命令冪等創建遷移歷史表，不執行版本檔案。
如果資料庫已包含這些檔案描述的變更，繼續前先閱讀[為已有資料庫建立基線](./history#baseline)。

## 預覽並執行

```sh
migrate up --dry-run
migrate up
migrate status
```

`up` 按版本順序執行待應用遷移，並記錄成功版本。重復執行不會重放已應用版本。
命令成功時可能沒有輸出，請通過 `status` 驗證結果。

Dry-run 輸出 SQL 或指令碼計劃，不執行使用者遷移，但可能讀取資料庫歷史、建立 SSH 隧道。
對於指令碼，它會編譯依賴而不求值，因此無法預測動態生成的 SQL。

## 處理較早的待執行版本

默認情況下，低於已應用版本的待執行檔案會被視為亂序並跳過。
確實需要補執行時，先確認其與當前資料庫結構兼容，再在[設定](./configuration)中
啓用 `migrate_out_of_order`。

## 在部署流程中執行

```sh
migrate --non-interactive up
migrate --non-interactive status --output json
```

提前注入所需環境變數以及 SSH、Proxy憑據。同一資料庫的遷移應串行執行。
在併發、進程崩潰或隱式提交的情況下，歷史記錄不能保證變更恰好執行一次。

執行失敗後，先檢查資料庫實際狀態和歷史，再決定是否重試；DDL 和非事務引擎尤其如此。
參見[問題排查](./troubleshooting)及[資料庫事務說明](./databases)。
