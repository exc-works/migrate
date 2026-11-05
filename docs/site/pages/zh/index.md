---
layout: home
hero:
  name: migrate
  text: 管理数据库变更。
  tagline: 从单库到多库，让数据库变更有序可控。
  actions:
    - theme: brand
      text: 开始使用
      link: ./quick-start
    - theme: alt
      text: 浏览文档
      link: ./guide
features:
  - title: 多种数据库，一套工具
    details: 支持 PostgreSQL、MySQL、MariaDB、SQLite、Oracle、SQL Server、ClickHouse、TiDB 和 Redshift。
    link: ./databases
    linkText: 查看数据库支持
  - title: SQL 与脚本迁移
    details: 支持编写 SQL、TS、JS 或 MJS 文件执行数据库迁移。
    link: ./scripting
    linkText: 编写迁移脚本
  - title: 一次执行，更新多个数据库
    details: 多个数据库共享一套迁移文件，并发执行、独立记录历史，按数据库汇总结果。
    link: ./configuration#multiple-databases
    linkText: 配置多个数据库
  - title: 环境变量配置
    details: 通过环境变量切换数据库连接与部署环境，在 YAML 中维护项目设置。
    link: ./configuration#environment-variables
    linkText: 配置连接与环境变量
  - title: SSH 隧道与代理
    details: 通过 SSH 堡垒机访问数据库，支持密码、私钥、SSH Agent，以及 SOCKS5 和 HTTP CONNECT 代理。
    link: ./ssh-tunnels
    linkText: 连接远程数据库
  - title: 预览、校验与回退
    details: 执行前预览变更，检查已应用文件是否被修改，并用 Down 操作回退版本。
    link: ./applying
    linkText: 执行数据库变更
---
