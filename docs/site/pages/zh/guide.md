# 文档导览

migrate 将数据库变更保存为有序的版本文件。你可以在部署流程中使用 CLI，
也可以把同一个迁移服务嵌入 Go 应用。

## 完成第一次迁移

从 [SQLite 快速开始](./quick-start)入手，无须启动数据库服务，就能完成一条 SQL
迁移的创建、执行和回退。已有项目可以先在[安装](./installation)中选择构建，
再按照[配置](./configuration)连接数据库。

## 管理数据库变更

1. [编写 SQL 迁移](./sql-migrations)：版本命名、Up/Down 分区与语句块。
2. [执行迁移](./applying)：预览计划、执行并验证结果。
3. [回退迁移](./rollback)：保留目标版本，或回退全部已应用版本。
4. [历史与基线](./history)：理解状态、校验和，以及如何接管已有数据库。

## 使用脚本和集成能力

需要代码处理的数据转换可以使用 [JavaScript 或 TypeScript](./scripting)。
[共享模块](./modules)、[数据库 API](./script-api)和[事务边界](./script-transactions)
分别介绍运行时的不同方面。[示例](./examples)提供本地导入、npm 包和嵌入文件的演练。

应用集成见[在 Go 中使用](./go-library)；数据库位于堡垒机之后时，
见 [SSH 隧道与代理](./ssh-tunnels)。

## 按需查阅

- [命令行参考](./commands)：命令、参数与选项。
- [数据库能力](./databases)：验证状态、占位符与引擎限制。
- [问题排查](./troubleshooting)：配置、历史与执行错误。
- [发布说明](../releases)：特定版本实际交付的功能。

::: info 适用版本
本站跟随 main 分支。请通过 `migrate version` 与对应发布说明，确认已安装版本是否包含所需功能。
:::
