# 支持的数据库

SQL 方言包括 PostgreSQL、MySQL、MariaDB、SQLite、Oracle、MSSQL、ClickHouse、TiDB 和 Redshift。
在配置中通过 `dialect` 选择数据库类型。不同数据库对脚本迁移和事务的支持有所不同，见下表。

## 脚本数据库支持 {#script-support}

实验性脚本支持尚未覆盖所有数据库版本和部署方式，使用前请在目标环境验证迁移。
事务列说明脚本失败时能否自动撤销变更，不代表所有 SQL 都可回退。

| 数据库 | 脚本状态 | 占位符 | 事务行为 |
| --- | --- | --- | --- |
| PostgreSQL | 已验证 | `$1` | migrate 管理事务 |
| MySQL / MariaDB | 已验证 | `?` | 需要事务表，DDL 可能隐式提交 |
| SQLite | 已验证 | `?` | migrate 管理事务 |
| TiDB | 实验性 | `?` | DDL 可能隐式提交；分布式 TiKV 部署尚未验证 |
| Redshift | 实验性，未完成真实实例验证 | `$1` | 事务取决于语句，不等同于完整 PostgreSQL |
| Oracle | 实验性，有限 Oracle Free 覆盖 | `:1` | DML 使用 migrate 管理事务，DDL 隐式提交 |
| ClickHouse | 实验性 | `?` | 非事务，无自动回退 |
| MSSQL / SQL Server | 实验性 | `@p1` | migrate 管理事务，Azure SQL 实例未验证 |

执行或预览实验性数据库的脚本迁移时，命令会显示提示。

## 可接受的名称 {#aliases}

将下表中的名称填入 `dialect`，例如 `dialect: postgres`。名称不区分大小写；
使用别名时，脚本中的 `db.dialect` 返回对应的规范名称。
MySQL、MariaDB 和 TiDB 仍是不同方言。

| 规范名称 | 别名 |
| --- | --- |
| `postgres` | `pg`、`pgsql`、`postgresql`、`pgx` |
| `mssql` | `sqlserver`、`azuresql` |
| `oracle` | `orcl`、`ora`、`go-ora`、`goora`、`godror` |
| `sqlite` | `sqlite3` |
| `mysql`、`mariadb`、`tidb`、`redshift`、`clickhouse` | 无额外别名 |

## 事务与回退

不同数据库的事务行为不同：部分 DDL 会隐式提交，ClickHouse 不提供框架自动回退。
关闭自动事务的脚本失败后，可能留下已提交的数据变更却没有成功历史；
重试前应先检查数据和迁移记录。

使用前请阅读[事务模式](./script-transactions)和能力表，
再通过 `migrate capabilities` 确认执行限制。

## 数据库特有的脚本值 {#value-conversion}

[数据库 API](./script-api)通过字符串保留精确整数与小数。部分类型还需要额外处理：

- **MSSQL：** MONEY/SMALLMONEY 保持精确字符串。UNIQUEIDENTIFIER 保留驱动的二进制表示，
  格式化文本使用 `CAST(id AS VARCHAR(36))`。Rowversion/timestamp 是二进制，时间值遵循通用转换规则。
- **Oracle：** NUMBER 为精确文本，RAW 为字节，CLOB 为文本。缺少类型元数据的 BLOB
  应通过有界 `DBMS_LOB.SUBSTR` 投影为 RAW，或通过 `RAWTOHEX` 转为文本。
  自建 DB 应设置 `FAST+LOGIN=false` 以可靠重连；CLI 默认关闭 go-ora v3 快速登录 cookie，保留显式配置。
- **Redshift：** pgx 元数据未知的类型应显式转换为支持的标量。
  VARBYTE 通过 `FROM_HEX($1)` 写入十六进制文本，使用 `FROM_VARBYTE(payload, 'hex')` 读取。
- **ClickHouse：** 128/256 位整数及小数保持精确字符串。Array/Tuple/Map/UUID/IP
  需转为支持的标量。String 必须为有效 UTF-8，任意二进制应转十六进制。
  需要等待 mutation 完成时使用语句级 `SETTINGS mutations_sync=2`；驱动的 `rowsAffected` 可能为 `"0"`。

## 查看已安装版本的支持情况

```sh
migrate capabilities
```

该命令列出支持的文件格式、数据库名称与脚本事务说明，不需要数据库连接。
自动化流程可以使用 `--output json`。
