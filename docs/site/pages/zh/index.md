---
layout: home
hero:
  name: migrate
  text: 有序管理数据库变更。
  tagline: Go 编写的数据库迁移 CLI 与库，支持 SQL，可选内嵌 JavaScript 和 TypeScript 运行时。
  actions:
    - theme: brand
      text: 从 SQLite 开始
      link: ./quick-start
    - theme: alt
      text: 阅读用户指南
      link: ./guide
features:
  - title: 从 SQL 开始
    details: 版本化迁移、执行预览、历史校验，以及明确的回退操作。
  - title: 按需使用脚本
    details: 完整构建增加 JS、TS 和 MJS，支持本地模块与已安装 npm 包；执行迁移不需要 Node.js 运行时。
  - title: 明确能力边界
    details: 使用 migrate capabilities 查看验证状态、数据库别名和事务限制。
---

## 选择下一步

- [快速开始](./quick-start)：使用 SQLite 完成可回退的迁移，无须数据库服务。
- [用户指南](./guide)：安装、配置、迁移历史和 Go 库集成。
- [命令参考](./commands)：按任务查找命令及参数说明。
- [数据库能力](./databases)：SQL/full 构建差异、验证状态和事务限制。
- [脚本迁移](./scripting)与[可运行示例](./examples)。
- [发布说明](../releases)、[变更记录](../changelog)与[其他语言](../languages)。

::: info 适用版本
本站跟随仓库 main 分支，可能包含尚未发布的变更。请通过 `migrate version`
查看已安装版本，并对照该版本的发布说明。
:::
