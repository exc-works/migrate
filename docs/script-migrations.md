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
不增加 Node.js、npm 或模块导入能力。
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

JS 模板通过 JSDoc 声明 `MigrationDB` 并标注 `db`，TS 模板提供等价的类型声明；
两者都包含数据库方法、参数和结果类型，便于编辑器提示。TS 执行时去除类型，不进行完整类型检查。
需要共享声明时，将 [migration-api.d.ts](migration-api.d.ts) 放在项目中，使用
`/// <reference path="./migration-api.d.ts" />` 引用，并移除模板内同名的局部类型。

## 数据与 SQL 接口

`db.query(sql, params?)` 返回行数组，`db.exec(sql, params?)` 返回字符串形式的
`rowsAffected`（ClickHouse 为 `null`，表示未知）。使用数据库原生占位符，参数必须是数组。整数、精确 decimal 结果和
受影响行数都是字符串；日期时间中的 Go 时间值输出为保留精度的 UTC 字符串。
NULL 对应 null，二进制结果为复制的字节数组。

不要把超出安全范围的整数或 decimal 转成 JS number 后再写入。JSON.parse 也可能
损失 JSON 大整数精度；这类内容需采用数字字符串等无损表示。SQLite NUMERIC 亲和性
可能将 decimal 存为浮点，工具会拒绝这种不精确结果；需要精确保留时使用合适的存储类型。

SQL 接口一次执行一条语句，可带尾部分号，不允许内部出现分号；包含分号的值使用绑定
参数传入。只接受规范列出的直接数据/结构语句；手动事务、会话控制、动态 SQL、过程调用、
管理命令和 MySQL/MariaDB 可执行注释会被拒绝。存储过程定义等复合 SQL
继续放在 SQL 迁移文件中。查询列名必须唯一，必要时使用别名。语句在空白后必须直接以支持的关键字开头；
不支持前置 SQL 注释，以避免不同数据库的嵌套和换行规则改变命令识别。说明请放在 JS/TS 注释中。

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

支持事务的路径中，数据操作与成功记录共用框架事务。数据库异常即使被脚本捕获仍会使
迁移失败；DDL 隐式提交等限制仍然存在，不能承诺所有操作都可完整回滚。

Oracle 使用固定的 go-ora v3 驱动，NUMBER 以精确字符串返回；RAW 返回字节数组，CLOB 返回文本。
驱动未提供 BLOB 类型元数据时需显式投影，例如用有界的 `DBMS_LOB.SUBSTR` 读成 RAW，
或用 `RAWTOHEX` 传递 hex 文本；不要依赖不明确的字节类型推断。
CLI 默认关闭 v3 的快速登录缓存以保证重连，保留显式 `FAST LOGIN` 配置。
Go facade 调用方自行创建数据库连接时，建议在 DSN 中设置 `FAST+LOGIN=false`。
PL/SQL 块、过程调用仍需使用 SQL 文件。Redshift 特有类型可能无法被 pgx 识别，需显式
转换为已知标量；例如使用 `FROM_HEX($1)` 写入 VARBYTE，使用
`FROM_VARBYTE(payload, 'hex')` 读成十六进制文本，不要假设 BYTEA 完全兼容。

ClickHouse 的整型（含 128/256 位）和 decimal 保持精确字符串。Array、Tuple、Map、
UUID、IP 等未适配驱动值需在 SQL 中显式转换为字符串或其他已支持标量；String 作为文本
必须是有效 UTF-8，任意二进制内容可用 hex 文本传递。`rowsAffected: null` 不能当成零行。
需要等待 mutation 完成时，在相应语句中使用 `SETTINGS mutations_sync=2`；不提供会话设置接口。

ClickHouse 脚本失败、超时或历史写入失败后，先前成功的写入可能已经生效，却没有成功历史；
错误会提示可能存在部分变更。down 也没有自动回滚保障。重试前检查数据与迁移历史，按实际
状态恢复；不要自动重跑，也不要为了消除错误盲目 baseline。

可使用以下命令运行新增的容器验证；TiDB/ClickHouse 的外部 DSN 优先于临时容器：

```bash
INTEGRATION_DB=clickhouse,tidb go test -tags=integration,full ./integrationtest/...
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
