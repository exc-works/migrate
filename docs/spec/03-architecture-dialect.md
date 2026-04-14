# 03 Architecture & Dialect

## 推荐目录

```text
cmd/migrate/main.go
internal/config/
internal/migrate/
internal/parser/
internal/dialect/
internal/source/
integrationtest/
```

## 职责边界

- `internal/migrate`：流程编排与状态机。
- `internal/parser`：语法切分与校验。
- `internal/dialect`：DDL/DML 与占位符差异。
- `internal/source`：目录/内存等 migration 来源。
- `cmd`：参数解析与配置装配。

## 事务与错误处理

- 所有 `Exec/Query/Commit/Rollback` 错误必须回传。
- 禁止通过 `panic` 处理正常错误分支。
- 记录错误时不输出敏感信息。

## 数据库支持

V1 必选：

- PostgreSQL（14+）
- MySQL（8.0+）
- MariaDB（10.11+）

V1.1 可选：

- SQLite（3.x）

实现扩展：

- Oracle（通过 `INTEGRATION_ORACLE_DSN` 执行集成测试）

## Schema 约束

- `version` 统一文本语义。
- schema/table 标识符需白名单校验后再拼 SQL。
- 统一迁移记录表字段（id/version/filename/hash/status/created_at）。
