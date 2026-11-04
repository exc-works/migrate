# 安装与数据库选择

根据迁移文件选择构建：

| 使用的文件 | 安装命令 |
| --- | --- |
| SQL | `go install github.com/exc-works/migrate/cmd/migrate@latest` |
| SQL、JavaScript 和 TypeScript | `go install -tags=full github.com/exc-works/migrate/cmd/migrate@latest` |

[下载二进制](https://github.com/exc-works/migrate/releases)时，选择带 `_full` 后缀的文件即可运行脚本。

将 `dialect` 设置为 PostgreSQL、MySQL、MariaDB、SQLite、Oracle、MSSQL、ClickHouse、TiDB 或 Redshift。
查看当前构建的文件格式、数据库别名和事务方式：

```sh
migrate capabilities
migrate capabilities --output json
```

脚本支持情况和事务说明见[数据库表](./scripting#配置和执行边界)。
