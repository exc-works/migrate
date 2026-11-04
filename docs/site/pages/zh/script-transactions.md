# 脚本事务与执行限制

## 框架托管事务

在支持的数据库上，每条脚本默认使用框架事务，成功历史写入也在同一事务中。
托管模式下不要自行执行 BEGIN、COMMIT 或 ROLLBACK；这些 SQL 会原样转发，
可能破坏框架的回退保证。

正确处理的数据库错误可以恢复，但未捕获异常、致命桥接错误、提交失败和历史写入失败
仍会导致迁移失败。DDL 隐式提交、过程副作用和引擎限制意味着无法对所有操作保证回退。
请先检查[数据库执行边界](./databases#script-support)。

## 显式关闭框架事务

对于必须在框架事务之外执行的 SQL，导出以下配置：

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
显式 transaction 值必须是布尔值。配置在实际执行预检时求值并快照，
status、baseline 和 dry-run 均不会对它求值。

显式非事务步骤中的所有调用共用一个专用连接。自行管理事务时，必须在成功返回前结束事务。
修改的会话设置应恢复，驱动重置行为有所不同，不能依赖跨迁移保留会话状态。
失败时专用连接会被丢弃。

非事务失败可能留下已提交变更，却没有成功历史。重试前先检查数据和历史；
非事务引擎上的 Down 同样没有自动回退保证。

## 限制执行时间与结果大小

在项目配置中增加以下可选字段：

```json
{
  "script_timeout": "5m",
  "script_max_rows": 1000,
  "script_max_bytes": 16777216
}
```

超限会失败，不会截断结果。大量读取应显式分页，但分页不会自动产生分批提交或可恢复迁移。
取消会中断 JS 并取消数据库操作，不保证原生调用硬实时终止。

Go 调用方通过 `migrate.ScriptOptions` 设置 `Config.ScriptOptions`，
并以 `-tags=full` 构建应用。参见[在 Go 中使用](./go-library)。
