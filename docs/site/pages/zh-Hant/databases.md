# 支持的資料庫

SQL 方言包括 PostgreSQL、MySQL、MariaDB、SQLite、Oracle、MSSQL、ClickHouse、TiDB 和 Redshift。
在設定中通過 `dialect` 選擇資料庫類型。不同資料庫對指令碼遷移和事務的支持有所不同，見下表。

## 指令碼資料庫支持 {#script-support}

實驗性指令碼支持尚未覆蓋所有資料庫版本和部署方式，使用前請在目標環境驗證遷移。
事務列說明指令碼失敗時能否自動撤銷變更，不代表所有 SQL 都可回退。

| 資料庫 | 指令碼狀態 | 佔位符 | 事務行為 |
| --- | --- | --- | --- |
| PostgreSQL | 已驗證 | `$1` | migrate 管理事務 |
| MySQL / MariaDB | 已驗證 | `?` | 需要事務表，DDL 可能隱式提交 |
| SQLite | 已驗證 | `?` | migrate 管理事務 |
| TiDB | 實驗性 | `?` | DDL 可能隱式提交；分布式 TiKV 部署尚未驗證 |
| Redshift | 實驗性，未完成真實實例驗證 | `$1` | 事務取決於語句，不等同於完整 PostgreSQL |
| Oracle | 實驗性，有限 Oracle Free 覆蓋 | `:1` | DML 使用 migrate 管理事務，DDL 隱式提交 |
| ClickHouse | 實驗性 | `?` | 非事務，無自動回退 |
| MSSQL / SQL Server | 實驗性 | `@p1` | migrate 管理事務，Azure SQL 實例未驗證 |

執行或預覽實驗性資料庫的指令碼遷移時，命令會顯示提示。

## 可接受的名稱 {#aliases}

將下表中的名稱填入 `dialect`，例如 `dialect: postgres`。名稱不區分大小寫；
使用別名時，指令碼中的 `db.dialect` 返回對應的規範名稱。
MySQL、MariaDB 和 TiDB 仍是不同方言。

| 規範名稱 | 別名 |
| --- | --- |
| `postgres` | `pg`、`pgsql`、`postgresql`、`pgx` |
| `mssql` | `sqlserver`、`azuresql` |
| `oracle` | `orcl`、`ora`、`go-ora`、`goora`、`godror` |
| `sqlite` | `sqlite3` |
| `mysql`、`mariadb`、`tidb`、`redshift`、`clickhouse` | 無額外別名 |

## 事務與回退

不同資料庫的事務行為不同：部分 DDL 會隱式提交，ClickHouse 不提供框架自動回退。
關閉自動事務的指令碼失敗後，可能留下已提交的資料變更卻沒有成功歷史；
重試前應先檢查資料和遷移記錄。

使用前請閱讀[事務模式](./script-transactions)和能力表，
再通過 `migrate capabilities` 確認執行限制。

## 資料庫特有的指令碼值 {#value-conversion}

[資料庫 API](./script-api)通過字符串保留精確整數與小數。部分類型還需要額外處理：

- **MSSQL：** MONEY/SMALLMONEY 保持精確字符串。UNIQUEIDENTIFIER 保留驅動的二進制表示，
  格式化文本使用 `CAST(id AS VARCHAR(36))`。Rowversion/timestamp 是二進制，時間值遵循通用轉換規則。
- **Oracle：** NUMBER 為精確文本，RAW 為字節，CLOB 為文本。缺少類型元資料的 BLOB
  應通過有界 `DBMS_LOB.SUBSTR` 投影為 RAW，或通過 `RAWTOHEX` 轉為文本。
  自建 DB 應設置 `FAST+LOGIN=false` 以可靠重連；CLI 默認關閉 go-ora v3 快速登錄 cookie，保留顯式設定。
- **Redshift：** pgx 元資料未知的類型應顯式轉換為支持的標量。
  VARBYTE 通過 `FROM_HEX($1)` 寫入十六進制文本，使用 `FROM_VARBYTE(payload, 'hex')` 讀取。
- **ClickHouse：** 128/256 位整數及小數保持精確字符串。Array/Tuple/Map/UUID/IP
  需轉為支持的標量。String 必須為有效 UTF-8，任意二進制應轉十六進制。
  需要等待 mutation 完成時使用語句級 `SETTINGS mutations_sync=2`；驅動的 `rowsAffected` 可能為 `"0"`。

## 查看已安裝版本的支持情況

```sh
migrate capabilities
```

該命令列出支持的檔案格式、資料庫名稱與指令碼事務說明，不需要資料庫連接。
自動化流程可以使用 `--output json`。
