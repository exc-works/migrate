# 编写脚本迁移

当数据转换更适合通过代码表达时，可以使用 JavaScript 或 TypeScript。
脚本通过内嵌引擎同步执行，无须 Node.js 运行时。
开始前先[安装工具](./installation)。

## 创建脚本

```sh
migrate new version "normalize names" --format ts
```

支持 `js`、`ts` 和 `mjs`；`.mjs` 与 `.js` 的运行时限制相同。
文件名采用 `V<version>__<description>.<format>`，与 SQL 文件共用版本序列及历史。

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

Dry-run 检查脚本及其导入，并列出待执行文件，不运行脚本，因此不会展示动态生成的 SQL。
把数据库操作放在 `up(db)` 或 `down(db)` 中，不要写在模块顶层；不同迁移之间不要依赖全局变量传递数据。

## 编辑器提示与脚本要求

生成的 JS 模板通过 JSDoc 标注 `db`，TS 模板提供等价声明。转译不执行类型检查。
也可以使用共享的 [migration-api.d.ts](../assets/migration-api.d.ts) 替换局部声明，
通过 triple-slash 路径注释引用。

脚本可以使用 JavaScript 语法和上面的数据库 API，但不能调用 Node 内置模块、
网络、文件系统、环境变量或定时器，也不能动态加载模块。导入 npm 包前，先确认它不依赖这些功能。

接下来阅读[共享模块](./modules)、[事务与执行限制](./script-transactions)，
或尝试[可运行示例](./examples)。
