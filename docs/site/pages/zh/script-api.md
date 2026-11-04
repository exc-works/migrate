# 脚本数据库 API

同步入口参数 `db` 提供规范方言名 `dialect`、`query(sql, params?)` 和
`exec(sql, params?)`。编辑器声明可下载 [migration-api.d.ts](../assets/migration-api.d.ts)。

## 查询与写入

下面的例子使用 SQLite 占位符：

```js
export function up(db) {
  const rows = db.query("SELECT id, name FROM users WHERE name=?", ["Alice"]);
  for (const row of rows) {
    const result = db.exec("UPDATE users SET name=? WHERE id=?", ["alice", row.id]);
    // result.rowsAffected 是十进制字符串，例如 "1"。
  }
}
```

`query` 将行完整读入内存，只返回第一个结果集；重复列名应使用别名。
`exec` 返回 `{ rowsAffected: string }`，保留驱动原始计数，包括 `"0"`，
不会根据预期变更数推断替代值。

## 绑定参数与值类型

参数使用数组，SQL 使用方言原生占位符：PostgreSQL/Redshift 为 `$1`，
MySQL/MariaDB/SQLite/TiDB/ClickHouse 为 `?`，Oracle 为 `:1`，MSSQL 为 `@p1`。
即使配置使用别名，`db.dialect` 也始终是规范方言名。

整数和精确小数结果是字符串，NULL 为 `null`，二进制值是复制的字节数组，
Go 时间戳转为精确的 UTC 字符串。不要通过 JavaScript number 转换大整数或精确小数。
`JSON.parse` 也可能对大数值字面量舍入，需要时应选择无损表示。

SQLite NUMERIC 亲和性可能把小数存为浮点数，运行时会拒绝这类值。
应选择能保持所需精度的存储类型。数据库特有类型可能需要显式 SQL 转换，
见[数据库能力](./databases#value-conversion)。

## 原生 SQL 行为

桥接层原样转发 SQL，不拆分或验证批处理、存储过程、管理语句，也不解释 `GO`
这样的客户端分隔符。保留原生终止符，包括 MERGE 必需的分号。
MySQL 批处理可能需要驱动 DSN 选项；实际执行仍受权限和驱动支持限制。

## 处理数据库错误

驱动错误转为可捕获的脱敏异常，不暴露原始厂商细节或错误链。
捕获可恢复错误后可以继续调用数据库，但捕获本身不会修复事务。
应按方言使用保存点或回滚；托管事务无法修复时应重新抛出异常。
捕获后正常返回表示作者接受该结果。

执行 context 取消、整体脚本超时、桥接值校验和资源限制错误，即使捕获也仍是致命错误，
并阻止后续 DB 调用。单独由驱动报告的超时可以捕获。
编写恢复逻辑前请阅读[事务与执行限制](./script-transactions)。
