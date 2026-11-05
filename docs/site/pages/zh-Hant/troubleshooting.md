# 問題排查

先查看 `migrate version`、`migrate capabilities --output json` 和
`migrate status --output json`。Status 需要資料庫連接，version 和 capabilities 不需要。

## 找不到命令或設定

PATH 中找不到 `migrate` 時，按[安裝](./installation)檢查二進制位置。
找不到設定檔案時，在項目目錄執行 `migrate new config`，或使用 `-c` / `-w` 指定已有設定的位置。
路徑解析見[設定](./configuration)。

默認設定檔案是 `migration_config.yml`。設定報錯時，根據提示的位置檢查縮進和字段值；
YAML 縮進使用空格，不使用 Tab。設定檔案的[查找順序](./configuration#設定檔案的查找順序)決定實際使用哪一份設定。

## 缺少環境變數

`can't find env value for XXX` 表示必需的 `${XXX}` 模板無法解析。
在命令環境中提供變數，或對合適的非敏感字段使用 `${XXX:default}`。
SSH 的 `password_env` 等字段應直接填寫變數名，參見 [SSH 認證](./ssh-tunnels)。

## 不支持的格式、方言或模塊

- 通過 `migrate capabilities` 確認 SQL、JS、TS 和 MJS 支持。舊版 SQL 二進制
  應按[安裝說明](./installation)升級。
- `unsupported dialect` 應檢查[規範名稱與別名](./databases#aliases)。
- 模塊無法解析時，檢查相對導入檔案的路徑，並提前安裝 npm 依賴。
  在 Go 中嵌入指令碼時，npm 包需通過 `DependencyDir` 指定安裝目錄。
- 依賴 Node API 或異步執行的包，即使成功打包也不能直接運行。參見[共享模塊](./modules)。

## 哈希或檔案名不一致

從版本控制恢復已應用檔案的原始內容和檔案名，後續變更新增更高版本。
不要通過對未知資料庫狀態建立基線來繞過錯誤。參見[歷史與基線](./history)。

## up 後沒有變化

成功命令可能沒有輸出。查看 status：版本可能已經應用或登記為基線，
也可能因低於已應用版本而被視為亂序並跳過。修改設定前先檢查[版本順序行為](./applying)。

## 回退被拒絕

`to-version must be set, or use --all` 表示必須指定目標或 `--all`，兩者不能同時使用。
目標版本會保留。指令碼缺少 down 函數時，對應選中範圍不可回退。參見[回退遷移](./rollback)。

## 遷移、超時或連接失敗

重試前檢查實際資料和歷史。非事務操作、DDL 隱式提交、COMMIT 前後斷線，
均可能導致命令失敗但變更已經生效。Baseline 不能修復部分執行。

指令碼請檢查[事務與資源限制](./script-transactions)。SSH 請檢查可信主機記錄、憑據、
Proxy以及堡壘機對原始資料庫地址的訪問。傳輸丟失不會自動重連或重放 SQL。
診斷信息經過脫敏；調查廠商錯誤時可檢查受控的資料庫或服務端記錄，避免分享憑據或行資料。
