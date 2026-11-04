# 编写 SQL 迁移

每个版本文件保存一次逻辑变更。migrate 从配置中的 `migration_source` 读取文件，
根据版本生成有序的迁移计划。

## 创建版本文件

```sh
migrate new version init_users
migrate new version add_email -v 202611040002
```

第一条命令自动生成版本，`-v` 可以显式指定版本。文件名采用
`V<version>__<description>.sql`。SQL 与脚本文件共用一个版本序列。
已应用的文件不要修改或重命名，应新增版本完成后续变更。

## 定义 Up 和 Down

```sql
-- +migrate Up
CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);

-- +migrate Down
DROP TABLE users;
```

Up 执行变更，Down 实现逆操作。这个例子的 Down 会删除表及其中的数据。
编写迁移时就应确定回退行为，包括丢失的数据能否恢复。

## 保留语句内部的分号

当一条语句内部包含分号，不能按分号拆开执行时，使用 `StatementBegin` 和
`StatementEnd` 包住整条语句。例如 PostgreSQL 函数：

```sql
-- +migrate Up
-- +migrate StatementBegin
CREATE FUNCTION migration_answer() RETURNS integer AS $$
BEGIN
    RETURN 42;
END;
$$ LANGUAGE plpgsql;
-- +migrate StatementEnd

-- +migrate Down
DROP FUNCTION migration_answer();
```

这些标记控制 SQL 解析；实际语法和 DDL 事务行为仍取决于数据库引擎。
接下来阅读[执行迁移](./applying)。
