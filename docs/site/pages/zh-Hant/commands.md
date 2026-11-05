# 命令行參考

通過 `migrate --help` 或 `migrate <command> --help` 確認已安裝版本的參數，
通過 `migrate version` 查看二進製版本。

## 全局選項

| 選項 | 默認值 | 含義 |
| --- | --- | --- |
| `--config`、`-c` | 自動查找 | 指定設定檔案；省略時見[查找順序](./configuration#設定檔案的查找順序) |
| `--working-dir`、`-w` | 當前目錄 | 從該目錄解析設定，並覆蓋遷移工作目錄 |
| `--non-interactive` | `false` | 缺少憑據或遇到未知 SSH 主機密鑰時失敗，不請求交互輸入或批准 |

## new config

```sh
migrate new config [filename] [--format yaml|json] [--multi] [--shards N] [--ssh-tunnel] [--force]
migrate new config shards.yml --multi --shards 4
migrate new config shards-ssh.yml --multi --ssh-tunnel --proxy socks5
```

生成 `migration_config.yml`。用 `--multi` 生成多資料庫設定，`--shards N` 指定資料庫數量；
`--ssh-tunnel` 添加 SSH 設置，`--proxy socks5|http` 添加 SSH 上游Proxy。
`--ssh-auth agent|private_key|password` 選擇 SSH 認證方式。
`--shards N` 需要同時使用 `--multi`，數量必須是正整數，編號從 0 開始。
`--proxy` 需要同時使用 `--ssh-tunnel`。
`--force` 允許覆蓋已有檔案。可通過 `.json` 檔案名或 `--format json` 選擇 JSON 輸出。
全部選項見 `migrate new config --help`，常見用法見[設定](./configuration)。

## new version

```sh
migrate new version <description> [-v <version>] [--format sql|js|ts|mjs]
```

在遷移目錄創建版本檔案。`--version` / `-v` 覆蓋自動生成的版本，
`--format` 默認為 `sql`。描述包含空格時使用引號，例如 `migrate new version "init users"`；
檔案名中的空格會替換為下划線。
參見 [SQL 遷移](./sql-migrations)或[指令碼遷移](./scripting)。

## 多資料庫

設定 `data_sources` 後，create/baseline/status/up/down 自動遍歷全部資料來源，
均支持 `--parallel N`（默認 8）。單庫失敗不取消其他目標，結果按資料來源匯總；
任一失敗返回非零退出碼。共享目錄與 SSH 設置見[多資料庫設定](./configuration#multiple-databases)。
`status --output json` 按資料來源名稱排序返回結果，每項包含 `data_source`、`status`，
成功時附帶 `migrations`，失敗時附帶 `error`。

## create

```sh
migrate create
```

冪等創建歷史表，不執行使用者遷移檔案。

## baseline

```sh
migrate baseline
```

將待執行遷移登記為基線，不執行內容。先確認現有資料庫已與這些檔案一致。
參見[歷史與基線](./history)。

## up

```sh
migrate up [--dry-run]
```

按順序應用待執行版本。Dry-run 輸出 SQL 或指令碼計劃，不執行使用者遷移，
但可能連接資料庫讀取歷史。參見[執行遷移](./applying)。

## down

```sh
migrate down <to-version> [--dry-run]
migrate down --all [--dry-run]
```

回退高於目標的版本並保留目標，或回退全部已應用版本。
目標和 `--all` 互斥，必須選擇其一。參見[回退遷移](./rollback)。

## status

```sh
migrate status [--output table|json]
```

對比檔案和資料庫歷史，`--output` / `-o` 默認為 `table`。
狀態含義見[歷史與基線](./history)。

## capabilities

```sh
migrate capabilities [--output text|json]
```

列出構建格式、別名、指令碼驗證狀態及事務限制，無須讀取設定或連接資料庫。
`--output` / `-o` 默認為 `text`。參見[支持的資料庫](./databases)。

## version

```sh
migrate version
```

輸出 CLI 版本，無須讀取設定或連接資料庫。這裡顯示的是工具版本，不是資料庫最新遷移版本。
