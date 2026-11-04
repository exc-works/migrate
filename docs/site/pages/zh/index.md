---
layout: home
hero:
  name: migrate
  text: 管理数据库变更。
  tagline: 用 SQL 维护迁移版本，用 JavaScript 或 TypeScript 处理数据。
  actions:
    - theme: brand
      text: 开始使用
      link: ./quick-start
    - theme: alt
      text: 浏览文档
      link: ./guide
features:
  - title: 从 SQL 开始
    details: 版本化迁移、执行预览、历史校验，以及明确的回退操作。
    link: ./sql-migrations
    linkText: 编写迁移
  - title: 按需使用脚本
    details: 完整构建增加 JS、TS 和 MJS，支持本地模块与已安装 npm 包；执行迁移不需要 Node.js 运行时。
    link: ./scripting
    linkText: 使用 JavaScript 或 TypeScript
  - title: 明确能力边界
    details: 使用 migrate capabilities 查看验证状态、数据库别名和事务限制。
    link: ./databases
    linkText: 检查数据库支持
---

## 你想完成什么？

| 任务 | 阅读入口 |
| --- | --- |
| 第一次尝试 migrate | [SQLite 快速开始](./quick-start) |
| 配置新项目 | [安装](./installation) · [配置](./configuration) |
| 部署数据库变更 | [执行迁移](./applying) · [回退迁移](./rollback) |
| 接管已有数据库 | [历史与基线](./history) |
| 复用脚本逻辑 | [共享模块](./modules) · [运行示例](./examples) |
| 接入应用或网络环境 | [Go 库](./go-library) · [SSH 隧道](./ssh-tunnels) |
| 查参数或解决错误 | [命令行参考](./commands) · [问题排查](./troubleshooting) |

::: info 适用版本
本站跟随仓库 main 分支，可能包含尚未发布的变更。请通过 `migrate version`
查看已安装版本，并对照该版本的发布说明。
:::
