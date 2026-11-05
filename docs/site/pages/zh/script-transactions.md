# 脚本事务与执行限制

## 默认事务行为

在支持事务的数据库上，migrate 默认把一条脚本及其成功记录放在同一个事务中。
脚本成功后提交；执行失败时尝试回滚。默认模式下不要自行执行 `BEGIN`、`COMMIT` 或 `ROLLBACK`。

建表、改表等 DDL 在部分数据库上会隐式提交，ClickHouse 等数据库不支持事务，
因此失败时可能仍有变更保留。请先查看[数据库事务说明](./databases#script-support)。

## 需要自己管理事务时

对于必须在事务之外执行的 SQL，导出以下配置：

```js
export const config = { transaction: false };

export function up(db) {
  // PostgreSQL：该语句不能在事务块内执行。
  db.exec("CREATE INDEX CONCURRENTLY users_name_idx ON users (name)");
}

export function down(db) {
  db.exec("DROP INDEX CONCURRENTLY users_name_idx");
}
```

配置同时作用于 up 和 down。省略 config/transaction 时默认为 true；
`transaction` 只接受 `true` 或 `false`。`status`、`baseline` 和 dry-run 不运行脚本，
因此不会检查运行时导出的配置是否有效。

关闭自动事务后，同一条迁移中的数据库操作使用同一个连接。
如果自行开启事务，必须在函数返回前提交或回滚；修改的会话设置也应恢复。
不要依赖下一条迁移保留上一条迁移的会话状态。

非事务失败可能留下已提交变更，却没有成功历史。重试前先检查数据和历史；
非事务引擎上的 Down 同样没有自动回退保证。

## 限制执行时间与结果大小

在项目配置中增加以下可选字段：

```yaml
script_timeout: 5m # 每条脚本的执行超时
script_max_rows: 1000 # 单次查询最多返回的行数
script_max_bytes: 16777216 # 单次查询结果大小上限，单位为字节（16 MiB）
```

超限会失败，不会截断结果。大量读取应显式分页，但分页不会自动产生分批提交或可恢复迁移。
取消命令会中断脚本并请求取消数据库操作；数据库可能需要一些时间才能结束正在执行的语句。

Go 调用方通过 `migrate.ScriptOptions` 设置 `Config.ScriptOptions`。参见[在 Go 中使用](./go-library)。
