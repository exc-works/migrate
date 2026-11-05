# 設定

在項目目錄中生成設定檔案：

```sh
migrate new config
```

打開生成的 `migration_config.yml`，填寫資料庫類型和連接信息，再執行遷移命令。

## 連接資料庫

下面的設定連接 PostgreSQL，並從環境變數 `DATABASE_DSN` 獲取連接字符串：

```yaml
# 資料庫類型
dialect: postgres
# 執行命令前，在 shell 或部署環境中設置 DATABASE_DSN
data_source_name: "${DATABASE_DSN}"
# 遷移檔案目錄
migration_source: migrations
# 用於記錄已執行版本的歷史表
schema_name: migration_schema
```

| 字段 | 用途 | 省略時的默認值 |
| --- | --- | --- |
| `dialect` | 資料庫類型，見[支持的名稱](./databases#aliases) | `postgres` |
| `data_source_name` | 資料庫連接字符串；與 `data_sources` 選擇其一 | 單庫時必填 |
| `data_sources` | 多個資料庫的名稱和連接設置，見[多資料庫設定](#multiple-databases) | 不啓用 |
| `migration_source` | 遷移檔案目錄 | `migrations` |
| `schema_name` | 遷移歷史表名稱 | `migration_schema` |
| `working_directory` | 查找遷移檔案的起始目錄 | 當前目錄 |
| `migrate_out_of_order` | 是否執行低於已應用版本的待執行版本 | `false` |
| `logger_level` | 記錄級別：`debug`、`info`、`warn`、`error`，見[記錄級別](#logging) | `info` |

### 各資料庫的連接示例

選擇對應資料庫，將下面的 `dialect` 和 `data_source_name` 填入設定。
把示例中的主機、端口、資料庫名或 Oracle 服務名替換為實際值，
並在運行命令前設置 `DB_USER` 和 `DB_PASSWORD` 環境變數。SQLite 直接填寫資料庫檔案路徑。

::: code-group

```yaml [PostgreSQL]
dialect: postgres
data_source_name: "postgres://${DB_USER}:${DB_PASSWORD}@db.example:5432/app?sslmode=require"
```

```yaml [MySQL]
dialect: mysql
data_source_name: "${DB_USER}:${DB_PASSWORD}@tcp(db.example:3306)/app?parseTime=true"
```

```yaml [MariaDB]
dialect: mariadb
data_source_name: "${DB_USER}:${DB_PASSWORD}@tcp(db.example:3306)/app?parseTime=true"
```

```yaml [SQLite]
dialect: sqlite
data_source_name: "./demo.sqlite"
```

```yaml [Oracle]
dialect: oracle
data_source_name: "oracle://${DB_USER}:${DB_PASSWORD}@db.example:1521/FREEPDB1"
```

```yaml [SQL Server]
dialect: mssql
data_source_name: "sqlserver://${DB_USER}:${DB_PASSWORD}@db.example:1433?database=app"
```

```yaml [ClickHouse]
dialect: clickhouse
data_source_name: "clickhouse://${DB_USER}:${DB_PASSWORD}@db.example:9000/app"
```

```yaml [TiDB]
dialect: tidb
data_source_name: "${DB_USER}:${DB_PASSWORD}@tcp(db.example:4000)/app?parseTime=true"
```

```yaml [Redshift]
dialect: redshift
data_source_name: "postgres://${DB_USER}:${DB_PASSWORD}@cluster.example:5439/app?sslmode=require"
```

:::

URL 形式的連接字符串中，使用者名或密碼包含 `@`、`:`、`/` 等特殊字符時，需要先做 URL 編碼；
環境變數替換不會自動編碼。MySQL、MariaDB 和 TiDB 使用的 `user:password@tcp(...)` 格式不按 URL 編碼。
TLS/SSL 參數請按資料庫服務的連接要求設置。

### 記錄級別 {#logging}

`logger_level` 控制顯示哪些級別的記錄，默認為 `info`：

| 值 | 顯示內容 |
| --- | --- |
| `debug` | 除錯信息、普通信息、警告和錯誤 |
| `info` | 普通信息、警告和錯誤 |
| `warn` | 警告和錯誤，也可寫為 `warning` |
| `error` | 僅錯誤 |

```yaml
logger_level: debug # 排查問題時顯示除錯信息
```

## 使用環境變數 {#environment-variables}

連接字符串支持 `${KEY}` 和 `${KEY:default}`：

- `${DATABASE_DSN}`：讀取 `DATABASE_DSN`；未設置時命令會報錯。
- `${DB_NAME:app}`：讀取 `DB_NAME`；未設置時使用 `app`。

同一份設定可以用於開發、測試和生產環境，只需在運行命令的環境中提供不同的變數值。
資料庫憑據通過環境變數或部署平台的密鑰管理功能提供。
SSH 地址、使用者名和Proxy地址也支持這些寫法，詳見 [SSH 隧道與Proxy](./ssh-tunnels)。

## 設定檔案的查找順序

未使用 `-c` 指定檔案時，migrate 在當前工作目錄按以下順序查找，使用找到的第一份設定：

1. `migration_config.yml`
2. `migration_config.yaml`
3. `migration_config.json`

多份設定不會合併。找到的檔案若有錯誤或無法讀取，命令會報錯，需要修正該檔案後再運行。

使用 `-c` 可以直接選擇設定檔案；指定檔案不存在時會報錯，不再查找其他檔案：

```sh
migrate -c ./configs/dev.yml status
```

## 為不同場景生成設定

根據項目需要選擇一條命令，生成後填寫其中的連接信息：

| 場景 | 命令 |
| --- | --- |
| 單個資料庫 | `migrate new config` |
| 使用自定義檔案名 | `migrate new config dev.yml` |
| 多個資料庫 | `migrate new config --multi` |
| 四個資料庫 | `migrate new config --multi --shards 4` |
| 通過 SSH 連接，使用 SSH Agent | `migrate new config --ssh-tunnel` |
| 通過 SSH 連接，使用私鑰 | `migrate new config --ssh-tunnel --ssh-auth private_key` |
| 通過 SSH 連接，使用密碼 | `migrate new config --ssh-tunnel --ssh-auth password` |
| 經 SOCKS5 Proxy連接 SSH 堡壘機 | `migrate new config --ssh-tunnel --proxy socks5` |
| 經 HTTP CONNECT Proxy連接 SSH 堡壘機 | `migrate new config --ssh-tunnel --proxy http` |

這些選項可以組合使用，例如為四個資料庫生成共享 SSH 與Proxy設置：

```sh
migrate new config shards.yml --multi --shards 4 --ssh-tunnel --proxy socks5
```

已有檔案不會被覆蓋；需要重新生成時加上 `--force`。
使用 `migrate new config --help` 查看全部選項。

## 將同一套遷移應用到多個資料庫 {#multiple-databases}

分庫或多個租戶資料庫需要相同變更時，使用 `data_sources` 列出各個資料庫，
替代頂層的 `data_source_name`。所有資料庫使用相同的資料庫類型、遷移目錄和歷史表名稱。

```yaml
dialect: postgres
migration_source: migrations
schema_name: migration_schema
data_sources:
  shard_0:
    data_source_name: "${SHARD_0_DSN}" # 第一個資料庫的連接字符串
  shard_1:
    data_source_name: "${SHARD_1_DSN}" # 第二個資料庫的連接字符串
```

運行前設置 `SHARD_0_DSN` 和 `SHARD_1_DSN`。如果通過 `--multi --shards N`
生成設定，資料庫從 `shard_0` 開始編號，分別引用 `SHARD_0_DSN`、`SHARD_1_DSN` 等變數。

![同一目錄中的遷移分別應用到 shard_0 和 shard_1，各庫獨立記錄歷史，最後匯總結果。](../assets/diagrams/multiple-databases-zh-Hant.svg)

```sh
migrate create
migrate up --dry-run
migrate up
migrate status
```

這些命令會自動處理全部設定的資料庫。每個資料庫根據自己的歷史選擇待執行版本，
並按版本順序執行；已經成功的版本不會重復執行。`new version` 只需運行一次，
生成的檔案供所有資料庫共用。

默認最多同時處理 8 個資料庫。使用 `--parallel` 調整數量，設為 1 時逐庫執行：

```sh
migrate up --parallel 4
migrate status --output json
```

某個資料庫失敗時，其他資料庫會繼續執行；結束後按資料庫匯總結果，任一失敗則命令返回非零退出碼。
這不是跨庫事務，某個資料庫失敗不會撤銷其他資料庫已經完成的變更。
按 Ctrl+C 會停止後續資料庫的執行，並取消正在運行的操作。

如果多個資料庫經過同一台堡壘機，可以在頂層設定 `ssh_tunnel`。
某個資料庫需要不同的 SSH 設置時，在該項下填寫完整的 `ssh_tunnel`，它會替換頂層設置。
不填寫或設為 `null` 時使用頂層設置。具體字段見 [SSH 隧道與Proxy](./ssh-tunnels)。

每個實際資料庫只設定一次；不要同時啓動多個命令修改同一個資料庫。
執行期間保持遷移檔案與指令碼依賴不變。失敗後先檢查該庫的資料和歷史，再決定是否重試。

## 工作目錄與相對路徑

```sh
migrate -w ./deploy up
migrate -w ./deploy -c configs/dev.yml status
```

`-w` 指定工作目錄。設定檔案從該目錄查找，設定中的 `working_directory` 也會被該選項覆蓋。
相對的 `migration_source` 從工作目錄查找；只傳入 `-c` 不會把工作目錄改成設定檔案所在目錄。

例如在項目根目錄運行 `migrate -c configs/dev.yml up`，設定中寫了
`migration_source: migrations`，使用的就是項目根目錄下的 `migrations`。
SSH 私鑰等檔案的相對路徑則從設定檔案所在目錄查找。

設定完成後，可以[編寫第一條 SQL 遷移](./sql-migrations)，或設置[指令碼執行限制](./script-transactions)。
