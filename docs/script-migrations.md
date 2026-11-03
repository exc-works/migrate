# SQL + JS/TS 迁移使用指南

默认 `migrate` 保持仅 SQL。完整版本内嵌 JS 引擎和 TS 转译器，不需要 Node.js，
执行时也不会下载运行时。详细行为以 [脚本迁移规范](spec/07-script-migrations.md) 为准。

## 安装和能力查询

```bash
go install -tags=full github.com/exc-works/migrate/cmd/migrate@latest
migrate capabilities
migrate capabilities --output json
```

从 Releases 下载时，原文件名对应 SQL 版，带 `_full` 后缀的文件对应完整版本；
Windows 后缀位于 `.exe` 前。两者都可以命名为 `migrate` 使用。校验发布的 SHA-256。
切换版本不会重建历史表。默认版遇到已知脚本文件会拒绝执行，不会跳过版本。

## 创建和执行

```bash
migrate new version normalize_profile --format js
migrate new version backfill_settings --format ts
migrate up --dry-run
migrate up
```

文件名沿用 `V<version>__<description>`，SQL、JS（含 `.mjs`）和 TS 共享同一版本序列，版本号不能重复。
`.mjs` 是 `.js` 的兼容后缀，执行方式和限制相同；可用 `--format mjs` 生成，默认仍生成 SQL。
不增加 Node.js 运行时。
已有 SQL 文件及已应用脚本保持不变；修改内容或改名会被历史校验阻断。

```js
export function up(db) {
  const sql = db.dialect === "postgres"
    ? "UPDATE users SET profile=$1 WHERE id=$2"
    : "UPDATE users SET profile=? WHERE id=?";

  for (const row of db.query("SELECT id, profile FROM users")) {
    const profile = JSON.parse(row.profile);
    profile.displayName = profile.name;
    delete profile.name;
    db.exec(sql, [JSON.stringify(profile), row.id]);
  }
}
```

仅当能实现真实逆操作时才导出同步 `down(db)`。缺少 down 会阻止整个选定回滚范围；
不要用空 down 假装可以回滚。生成的 up 默认抛错，必须先实现后执行。
普通同步返回值会被忽略，支持 `export const up = db => db.exec(...)` 等写法。
Promise 返回值、异步和生成器入口仍不支持；工具不会等待或执行返回的任务。

JS 模板通过 JSDoc 声明 `MigrationDB` 并标注 `db`，TS 模板提供等价的类型声明；
两者都包含数据库方法、参数和结果类型，便于编辑器提示。TS 执行时去除类型，不进行完整类型检查。
需要共享声明时，将 [migration-api.d.ts](migration-api.d.ts) 放在项目中，使用
`/// <reference path="./migration-api.d.ts" />` 引用，并移除模板内同名的局部类型。

## 共享模块

目录来源的迁移可以静态导入或再导出 `.js`、`.ts`、`.mjs` 文件。路径相对于每个
导入文件解析，允许引用迁移目录外的共享代码；辅助文件仅作为依赖，不作为独立迁移执行。

```ts
import { normalize } from "../shared/normalize.ts";

export function up(db) {
  db.exec("UPDATE users SET name=? WHERE id=?", [normalize(" Alice "), "1"]);
}
```

工具在迁移写入前打包依赖；每条迁移保持独立的 VM 和模块状态。后续选中迁移的依赖错误
会阻止前面的迁移执行；回滚也先准备整个选定范围。dry-run 编译但不执行模块，
status/baseline 只读取元数据。

checksum 只覆盖入口原始源码。修改辅助文件不会导致历史校验失败，也不会重新执行已应用
迁移。需要保留历史可重放性时，保留旧辅助实现，新增版本供新迁移使用。

也支持已安装 npm 包的包名导入，例如
`import { normalize } from "@company/migration-utils"`。运行迁移前，使用项目的包管理器
和 lockfile 安装依赖。esbuild 处理包导出、子路径、嵌套依赖及静态 CommonJS 导入；
使用 neutral 条件，入口字段依次为 `module`、`main`。工具不安装包，也不增加 Node API；
包的运行要求必须兼容 Goja 同步运行时。虚拟来源的配置方式见下文。
Node 内置模块、原生扩展、动态 import、动态 require 和异步入口仍不支持；
运行时不提供通用模块加载器。

## 嵌入与内存模块

`FSSource` 从提供的 `fs.FS` 解析相对模块，支持 `embed.FS`、`fstest.MapFS`；
每次导入相对于当前模块位置查找。支持扩展名查找和目录 `index` 文件，暂不解析虚拟
package.json 目录入口或虚拟 node_modules。需要已安装 npm 包时，显式提供磁盘依赖目录：

```go
source := migrate.WithModules{
    Source: migrate.FSSource{
        FS:   embeddedFiles,
        Root: "migrations",
    },
    DependencyDir: projectDirectory,
}
```

`embeddedFiles` 是调用者提供的文件系统，`projectDirectory` 是已安装依赖的本地项目目录。
相对 helper 仍从虚拟文件系统读取，包及包内依赖使用磁盘解析。未指定 DependencyDir 时，
虚拟来源不会查找进程当前目录。

`StringSource` 可以通过 `WithModules` 的 `FS` 提供辅助模块、通过可选的 `Root` 定位入口；
也可以设置 `Directory` 使用本地辅助文件和包。现有来源接口及结构体字段布局保持兼容。
自定义来源可显式实现 `MigrationEntriesSource` 提供上下文；已有 LoadMigrations 覆盖方法
（包括嵌入内置来源的类型）继续按原有方式执行。

## 数据与 SQL 接口

`db.query(sql, params?)` 返回行数组，`db.exec(sql, params?)` 返回字符串形式的
`rowsAffected`（所有数据库均返回驱动计数的字符串形式）。使用数据库原生占位符，参数必须是数组。整数、精确 decimal 结果和
受影响行数都是字符串；日期时间中的 Go 时间值输出为保留精度的 UTC 字符串。
NULL 对应 null，二进制结果为复制的字节数组。

不要把超出安全范围的整数或 decimal 转成 JS number 后再写入。JSON.parse 也可能
损失 JSON 大整数精度；这类内容需采用数字字符串等无损表示。SQLite NUMERIC 亲和性
可能将 decimal 存为浮点，工具会拒绝这种不精确结果；需要精确保留时使用合适的存储类型。

SQL 会原样交给驱动。原生注释、分号、批处理、过程定义/调用、管理命令的支持由数据库
权限和驱动决定；工具不再做 SQL 关键字过滤、语法检查或语句拆分，也不解释 GO 等客户端
分隔符。保留原生语法要求的结束符，例如 MERGE 的分号。MySQL 批处理可能需要 DSN 驱动
选项。查询列名必须唯一；`query` 只返回第一个结果集，exec 计数沿用驱动的批处理语义。

支持事务的数据库默认使用框架事务。需要自行管理事务、调整会话，或执行不能在事务中运行
的 SQL 时，在迁移模块导出：

```js
export const config = { transaction: false }; // 同时作用于 up 和 down
export function up(db) {
  db.exec("BEGIN TRANSACTION"); // 使用所选数据库的原生语法
  db.exec("UPDATE users SET profile=@p1", ["changed"]);
  db.exec("COMMIT TRANSACTION");
}
```

省略 config/transaction 默认 true；提供的 transaction 值必须是 boolean。实际执行前初始化
所有选中模块时验证并固定该配置，status、baseline 和 dry-run 不执行配置表达式。
显式关闭事务后，一条迁移的所有调用使用同一独占连接；成功后归还连接池复用，
由 database/sql 和驱动的有效性/重置机制处理不健康连接。不同驱动不一定清除会话设置，
作者应恢复修改过的设置，不要依赖跨迁移的会话状态。
作者应在正常返回前结束所有手动事务；成功返回却遗留未结束事务属于无效用法。
数据库驱动返回的执行、受影响行数、查询元数据、扫描、迭代和结果关闭错误，统一作为
脱敏异常抛给 JS；脚本可以捕获后继续调用数据库，或重新抛出使迁移失败。工具不按数据库
错误码分类，也不保证捕获错误后的连接仍然可用。捕获错误并正常返回，表示作者接受该错误。
数据库事务不会因此自动恢复，需按方言使用保存点或手动回滚。如果整个框架事务已被
数据库终止且无法用保存点恢复，应重新抛错；需要重试整个事务时使用显式手动模式。
未捕获的异常、提交失败和历史写入失败仍使迁移失败。
实际执行上下文取消、脚本总超时、参数/结果校验及资源超限即使被捕获仍会终止迁移，
阻止后续数据库调用；驱动单独报告的超时或取消可以由脚本处理。
驱动异常不向 JS 暴露原始详情或错误链；未捕获或致命错误会丢弃独占连接，清理尚未提交的事务。
非事务迁移失败时，已提交的数据可能保留而没有成功历史；错误会提示部分变更，重试前
检查数据和历史。默认框架事务中不要自行 BEGIN/COMMIT/ROLLBACK；工具会转发这些 SQL，
但它们可能破坏回滚和历史一致性保障。迁移作者负责所执行的 SQL。

## 配置和执行边界

可在现有配置文件中加入以下可选字段：

```json
{
  "script_timeout": "5m",
  "script_max_rows": 1000,
  "script_max_bytes": 16777216
}
```

Go API 对应 `migrate.Config.ScriptOptions` 和 `migrate.ScriptOptions`；业务程序
使用脚本时，也需要以 `-tags=full` 构建。结果超限会失败，不会静默截断。大表采用
有序分批查询；分批读取降低结果内存，不意味着分批提交或支持断点恢复。

脚本支持以下数据库；“实验性”表示允许执行，但数据库版本、类型和部署形态的验证范围有限。
`migrate capabilities` 的文本和 JSON 输出会显示验证状态、事务模式及限制；JSON 保留
`script_dialects`，并新增 `script_dialect_capabilities`。选定计划含实验性脚本时会警告一次，
无需额外确认；status、baseline 和纯 SQL 计划不产生该警告。

| 数据库 | 状态 | 占位符 | 执行边界 |
| --- | --- | --- | --- |
| PostgreSQL | 已验证 | `$1` | 使用框架事务 |
| MySQL / MariaDB | 已验证 | `?` | 需要事务表；DDL 可能隐式提交 |
| SQLite | 已验证 | `?` | 使用框架事务 |
| TiDB | 实验性 | `?` | CI 使用 Unistore；分布式 TiKV 部署未验证；DDL 可能隐式提交 |
| Redshift | 实验性，真实实例未验证 | `$1` | 事务能力取决于语句；不能视为完全等同 PostgreSQL |
| Oracle | 实验性，仅 Oracle Free 验证 | `:1` | DML 使用框架事务；DDL 隐式提交 |
| ClickHouse | 实验性 | `?` | 非事务执行；无自动回滚 |
| MSSQL / SQL Server | 实验性 | `@p1` | 使用框架事务；Azure SQL 实例未验证 |

数据库名称忽略大小写及首尾空白；配置、CLI、Go facade 和脚本能力检查共用同一套别名：

| 规范名称 | 接受的别名 |
| --- | --- |
| `postgres` | `pg`、`pgsql`、`postgresql`、`pgx` |
| `mssql` | `sqlserver`、`azuresql` |
| `oracle` | `orcl`、`ora`、`go-ora`、`goora`、`godror` |
| `sqlite` | `sqlite3` |
| `mysql`、`mariadb`、`tidb`、`redshift`、`clickhouse` | 无额外别名 |

MySQL、MariaDB、TiDB 仍是独立方言；不要把共享驱动当成方言别名。
`db.dialect` 始终返回规范名称，例如 `azuresql` 配置会得到 `mssql`。
`capabilities` 的文本输出和 JSON `aliases` 字段会列出接受的别名。

MSSQL 金额类型 MONEY/SMALLMONEY 返回精确字符串。UNIQUEIDENTIFIER 保留驱动的二进制
表示并复制为字节数组；需要标准 GUID 文本时使用 `CAST(id AS VARCHAR(36))`，写入可绑定
GUID 字符串。rowversion/timestamp 是二进制值；DATETIMEOFFSET 等日期时间沿用通用时间
转换约定。原生 T-SQL 批处理和过程调用由驱动执行。

框架事务模式下，数据操作与成功记录共用事务。已处理的 SQL 错误允许恢复；未捕获的
错误和致命桥接错误使迁移失败。手动事务控制、过程副作用、DDL 隐式提交等都可能破坏
回滚保障。非事务模式不承诺数据和历史的原子性。

Oracle 使用固定的 go-ora v3 驱动，NUMBER 以精确字符串返回；RAW 返回字节数组，CLOB 返回文本。
驱动未提供 BLOB 类型元数据时需显式投影，例如用有界的 `DBMS_LOB.SUBSTR` 读成 RAW，
或用 `RAWTOHEX` 传递 hex 文本；不要依赖不明确的字节类型推断。
CLI 默认关闭 v3 的快速登录缓存以保证重连，保留显式 `FAST LOGIN` 配置。
Go facade 调用方自行创建数据库连接时，建议在 DSN 中设置 `FAST+LOGIN=false`。
原生 PL/SQL 块、过程调用按驱动能力执行。Redshift 特有类型可能无法被 pgx 识别，需显式
转换为已知标量；例如使用 `FROM_HEX($1)` 写入 VARBYTE，使用
`FROM_VARBYTE(payload, 'hex')` 读成十六进制文本，不要假设 BYTEA 完全兼容。

ClickHouse 的整型（含 128/256 位）和 decimal 保持精确字符串。Array、Tuple、Map、
UUID、IP 等未适配驱动值需在 SQL 中显式转换为字符串或其他已支持标量；String 作为文本
必须是有效 UTF-8，任意二进制内容可用 hex 文本传递。`rowsAffected` 直接采用驱动返回值；ClickHouse 驱动可能返回 `"0"`，框架不另行替换或推断。
需要等待 mutation 完成时，在相应语句中使用 `SETTINGS mutations_sync=2`；会话控制按驱动能力执行，需显式关闭框架事务。

ClickHouse 脚本失败、超时或历史写入失败后，先前成功的写入可能已经生效，却没有成功历史；
错误会提示可能存在部分变更。down 也没有自动回滚保障。重试前检查数据与迁移历史，按实际
状态恢复；不要自动重跑，也不要为了消除错误盲目 baseline。

可使用以下命令运行新增的容器验证；TiDB/ClickHouse/MSSQL 的外部 DSN 优先于临时容器：

```bash
INTEGRATION_DB=clickhouse,tidb,mssql go test -tags=integration,full ./integrationtest/...
```

Oracle/Redshift 通过 `INTEGRATION_ORACLE_DSN` / `INTEGRATION_REDSHIFT_DSN` 运行可选验证。
未配置时跳过不等于通过真实实例验证。

运行时不提供 Node.js、npm、运行时模块导入、网络、文件系统、环境变量或定时器接口。
入口必须同步；迁移脚本应由团队维护和审查。超时会中断 JS 并取消数据库操作，但不承诺
原生调用的硬实时终止或每个 VM 的硬内存隔离。

status、baseline 和 dry-run 不执行脚本。实际执行前，工具会准备整个选定范围并验证
入口；模块初始化时不提供 DB，且每个模块只初始化一次。dry-run 展示脚本计划，不能
预测动态 SQL。错误保留文件和行号，但不会回显原始异常内容、查询值和凭据。

同一数据库上的迁移操作需要串行运行。历史表记录成功版本，不构成跨崩溃、隐式提交或
并发执行的 exactly-once 保证。
