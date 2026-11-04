# JavaScript 和 TypeScript 迁移

用脚本处理数据转换，并通过 `db` 接口执行 SQL。以下示例使用 SQLite。

## 安装和能力查询

```bash
go install -tags=full github.com/exc-works/migrate/cmd/migrate@latest
migrate capabilities
```

下载二进制时选择带 `_full` 后缀的文件。完整构建内嵌脚本引擎，可执行 SQL、JS、TS 和 MJS。

## 创建和执行

在已配置的迁移项目中创建文件：

```bash
migrate new version create_projects --format js
```

将生成文件的内容替换为：

```js
export function up(db) {
  db.exec("CREATE TABLE projects (id INTEGER PRIMARY KEY, name TEXT NOT NULL)");
}

export function down(db) {
  db.exec("DROP TABLE projects");
}
```

预览、执行并确认状态：

```bash
migrate up --dry-run
migrate up
migrate status
```

使用同步的 `up(db)` 编写变更，使用 `down(db)` 编写回退操作；回退时工具会检查选定范围内每个脚本的 `down`。
需要 TypeScript 时选择 `--format ts`，需要 MJS 时选择 `--format mjs`。
SQL 和脚本共用一条版本序列，为每个迁移分配独立版本号。

模板提供编辑器类型提示。可下载 [migration-api.d.ts](migration-api.d.ts) 供共享模块使用，
需要类型检查时，在项目中配置 TypeScript 并运行 `pnpm exec tsc --noEmit`。

## 共享模块

在 `shared/normalize.ts` 中保存共享函数：

```ts
export const normalize = (value: string) => value.trim();
```

在迁移文件中通过相对路径导入。以下示例使用已有的 `users(id, name)` 表：

```ts
import { normalize } from "../shared/normalize.ts";

export function up(db) {
  db.exec("UPDATE users SET name=? WHERE id=?", [normalize(" Alice "), "1"]);
}
```

使用 npm 包时，运行 `pnpm add <包名>` 安装依赖，再按包名静态导入。提交 lockfile，部署时运行 `pnpm install --frozen-lockfile`。
选择仅使用 JavaScript 语言和 `db` 接口做同步数据处理的包。完整项目见[可运行示例](../examples/script-imports/README.md)。

工具会在执行迁移前打包选定范围的全部依赖。用 `--dry-run` 检查打包结果和迁移计划。
校验和覆盖迁移入口；为保持历史可重放，保留已发布入口、共享函数及 lockfile，变更时新增版本。

## 嵌入与内存模块

在 Go 应用中使用 `FSSource` 加载 `embed.FS` 或内存文件系统。
通过 `DependencyDir` 指向已安装包的项目目录：

```go
source := migrate.WithModules{
    Source: migrate.FSSource{
        FS:   embeddedFiles,
        Root: "migrations",
    },
    DependencyDir: projectDirectory,
}
```

相对导入从提供的文件系统读取，包名导入从 `DependencyDir` 读取。
`StringSource` 可通过 `WithModules.FS` 提供辅助文件，或通过 `Directory` 使用本地目录。
[嵌入示例](../examples/script-imports/embedded/main.go)展示完整配置。

## 数据与 SQL 接口

| 操作 | 调用 |
| --- | --- |
| 查询行 | `db.query(sql, [参数])`，返回行数组 |
| 执行 SQL | `db.exec(sql, [参数])`，返回字符串形式的受影响行数 |
| 选择方言语法 | `db.dialect`，返回规范数据库名称 |

使用数据库原生占位符，给查询列设置唯一名称。SQL 按驱动原生语法执行，`query` 读取第一个结果集。

整数和精确 decimal 结果以字符串返回，计算和存储时保留精确表示。
时间值使用 UTC 字符串，NULL 使用 `null`，二进制结果使用字节数组。
SQLite 中需要精确保留 decimal 时，选择能保留精度的存储类型。

支持事务的数据库默认由工具管理事务，数据变更与成功历史共用事务。
需要执行 SQLite `VACUUM` 等事务外语句，或自行控制事务时，在模块中导出：

```js
export const config = { transaction: false };
```

该配置同时作用于 up 和 down。自行控制事务时，在函数返回前完成提交或回退，并恢复修改过的会话设置。
处理 SQL 异常时，先用保存点或回滚恢复事务；无法恢复时重新抛出异常。

## 配置和执行边界

在配置文件中设置执行时间和单次查询结果上限：

```json
{
  "script_timeout": "5m",
  "script_max_rows": 1000,
  "script_max_bytes": 16777216
}
```

超时或结果超限会终止迁移；调整查询或限额后再重试。
大表使用有序分页查询。分页读取仍遵循迁移本身的事务范围。
Go 应用通过 `migrate.Config.ScriptOptions` 设置这些值，并使用 `-tags=full` 构建。

运行 `migrate capabilities --output json` 查看当前构建的数据库状态、别名和事务模式。

| 数据库 | 脚本状态 | 占位符 | 事务方式 |
| --- | --- | --- | --- |
| PostgreSQL | 已验证 | `$1` | 框架事务 |
| MySQL / MariaDB | 已验证 | `?` | 使用事务表；DDL 可能隐式提交 |
| SQLite | 已验证 | `?` | 框架事务 |
| TiDB | 实验性，CI 使用 Unistore | `?` | 在实际 TiKV 部署中验证；DDL 可能隐式提交 |
| Redshift | 实验性，待真实实例验证 | `$1` | 按语句确认事务支持 |
| Oracle | 实验性，Oracle Free 已验证 | `:1` | DML 使用事务；DDL 隐式提交 |
| ClickHouse | 实验性 | `?` | 非事务执行 |
| MSSQL / SQL Server | 实验性 | `@p1` | 框架事务；在目标 Azure SQL 实例中验证 |

遇到数据库特有类型时，先在 SQL 中转换成工具支持的标量：

| 数据 | 做法 |
| --- | --- |
| MSSQL GUID | 使用 `CAST(id AS VARCHAR(36))` 读取文本 |
| Oracle BLOB | 用有界的 `DBMS_LOB.SUBSTR` 读成 RAW，或用 `RAWTOHEX` 读取十六进制文本 |
| Redshift VARBYTE | 用 `FROM_HEX($1)` 写入，用 `FROM_VARBYTE(payload, 'hex')` 读取 |
| ClickHouse Array / Map / Tuple | 转成字符串；等待 mutation 完成时使用 `SETTINGS mutations_sync=2` |

非事务迁移失败时，先检查数据与 `migrate status`，修复已提交的部分变更，再重试。
在部署流水线中为同一数据库设置单一迁移任务。

详细接口见[类型声明](migration-api.d.ts)，完整用法见[脚本示例](../examples/script-imports/README.md)。
