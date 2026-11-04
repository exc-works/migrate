# 安装与数据库选择

默认安装包含 SQL、JavaScript、TypeScript 和 MJS 迁移：

```sh
go install github.com/exc-works/migrate/cmd/migrate@latest
```

[发布二进制](https://github.com/exc-works/migrate/releases)包含相同的迁移格式，无需额外构建标签或构建后缀。

将 `dialect` 设置为 PostgreSQL、MySQL、MariaDB、SQLite、Oracle、MSSQL、ClickHouse、TiDB 或 Redshift。
查看当前构建的文件格式、数据库别名和事务方式：

```sh
migrate capabilities
migrate capabilities --output json
```

脚本支持情况和事务说明见[数据库表](./scripting#配置和执行边界)。
