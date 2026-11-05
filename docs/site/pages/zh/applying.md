# 执行迁移

执行前，确认配置指向预期数据库，迁移目录中也只包含本次需要部署的版本。

## 初始化历史表

```sh
migrate create
```

该命令幂等创建迁移历史表，不执行版本文件。
如果数据库已包含这些文件描述的变更，继续前先阅读[为已有数据库建立基线](./history#baseline)。

## 预览并执行

```sh
migrate up --dry-run
migrate up
migrate status
```

`up` 按版本顺序执行待应用迁移，并记录成功版本。重复执行不会重放已应用版本。
命令成功时可能没有输出，请通过 `status` 验证结果。

Dry-run 输出 SQL 或脚本计划，不执行用户迁移，但可能读取数据库历史、建立 SSH 隧道。
对于脚本，它会编译依赖而不求值，因此无法预测动态生成的 SQL。

## 处理较早的待执行版本

默认情况下，低于已应用版本的待执行文件会被视为乱序并跳过。
确实需要补执行时，先确认其与当前数据库结构兼容，再在[配置](./configuration)中
启用 `migrate_out_of_order`。

## 在部署流程中执行

```sh
migrate --non-interactive up
migrate --non-interactive status --output json
```

提前注入所需环境变量以及 SSH、代理凭据。同一数据库的迁移应串行执行。
在并发、进程崩溃或隐式提交的情况下，历史记录不能保证变更恰好执行一次。

执行失败后，先检查数据库实际状态和历史，再决定是否重试；DDL 和非事务引擎尤其如此。
参见[问题排查](./troubleshooting)及[数据库事务说明](./databases)。
