# 构建与数据库能力

## SQL 版与完整构建

```sh
# SQL 版
go install github.com/exc-works/migrate/cmd/migrate@latest

# 完整构建：SQL、JS、TS 和 MJS
go install -tags=full github.com/exc-works/migrate/cmd/migrate@latest
migrate capabilities
migrate capabilities --output json
```

Release 中不带后缀的文件是 SQL 版，带 `_full` 后缀的文件包含脚本能力。
两者执行时都不会下载解释器。使用 npm 导入前需自行安装依赖，脚本运行时不提供 Node.js API。

## 查询当前二进制

SQL 方言包括 PostgreSQL、MySQL、MariaDB、SQLite、Oracle、MSSQL、ClickHouse、TiDB 和 Redshift。
SQL 可用性和脚本验证状态是不同的能力，选择构建和数据库前应检查当前二进制输出：

- `formats`：当前构建包含的迁移格式。
- `script_dialect_capabilities`：脚本验证状态、事务模式与限制。
- `aliases`：可接受的名称，配置最终归一化为规范方言名。

[脚本数据库能力表](./scripting#配置和执行边界)统一维护在脚本指南中。
实验性支持的验证覆盖有限；外部数据库测试跳过不代表已完成真实实例验证。

## 事务与回退

不同引擎的边界不同：部分 DDL 会隐式提交，ClickHouse 不提供框架自动回退。
显式关闭框架事务的脚本失败后，可能留下已提交的数据变更却没有成功历史；
重试前应先检查数据和迁移记录。

使用前请阅读[数据与 SQL 接口](./scripting#数据与-sql-接口)和能力表，
再通过 `migrate capabilities` 确认已安装构建的实际限制。
