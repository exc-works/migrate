# 编写脚本迁移

当数据转换更适合通过代码表达时，可以使用 JavaScript 或 TypeScript。
完整构建内嵌 esbuild 和 Goja，同步执行脚本，无须 Node.js 运行时。
开始前先[安装完整构建](./installation)。

## 创建脚本

```sh
migrate new version normalize_names --format ts
```

支持 `js`、`ts` 和 `mjs`；`.mjs` 与 `.js` 的运行时限制相同。
文件名采用 `V<version>__<description>.<format>`，与 SQL 文件共用版本序列及历史。
SQL 版遇到已知脚本文件会报错，不会静默跳过。

## 实现入口函数

假设 SQLite 的 `users` 表包含文本字段 `name`：

```ts
export function up(db) {
  for (const row of db.query("SELECT id, name FROM users")) {
    db.exec("UPDATE users SET name=? WHERE id=?", [row.name.trim(), row.id]);
  }
}
```

生成模板中的 up 会抛错，需先完成实现。入口必须同步执行，不支持 async 函数、
generator 或 Promise 返回值；普通同步返回值会被忽略。
SQL 占位符随方言变化，参见[数据库 API](./script-api)。

只有实现逆操作后才应导出 `down(db)`。去掉首尾空白会丢失信息，因此这个例子没有
down 函数。包含缺失 down 脚本的回退范围会被阻断；需要可逆转换时，应先保存原值。

## 预览并执行

```sh
migrate up --dry-run
migrate up
migrate status
```

Dry-run 编译依赖并输出计划，不对脚本求值。实际执行会在写入前准备整个选中范围，
模块初始化时不能访问数据库。每条迁移拥有独立的 VM 和模块状态。

## 编辑器支持与运行时边界

生成的 JS 模板通过 JSDoc 标注 `db`，TS 模板提供等价声明。转译不执行类型检查。
也可以使用共享的 [migration-api.d.ts](../assets/migration-api.d.ts) 替换局部声明，
通过 triple-slash 路径注释引用。

脚本应视为受信任的仓库代码。不提供 Node 内置模块、运行时模块加载、网络、文件系统、
环境变量或定时器 API。取消会中断 JavaScript 并取消数据库操作，但不保证原生调用
硬实时终止或每个 VM 的堆隔离。

接下来阅读[共享模块](./modules)、[事务与执行限制](./script-transactions)，
或尝试[可运行示例](./examples)。
