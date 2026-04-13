# 04 Testing with testcontainers-go

## 单元测试

至少覆盖：

- 文件名规则与版本比较。
- SQL 分段与注释/分号边界。
- out-of-order 开关行为。
- 配置读取与环境变量替换。
- mismatch 与非法输入错误路径。

建议目标：

- `internal/migrate` + `internal/parser` 行覆盖率 >= 80%。

## 集成测试（必须）

- 使用 `testcontainers-go`，禁止依赖本地固定数据库。
- 每个测试使用独立 schema 或数据库名。
- 统一命令：`go test -tags=integration ./...`。

## 必测场景（每种 DB 都执行）

- `create -> status(empty) -> up -> status(applied)`
- 二次 `up` 幂等
- `down` 到指定版本
- `baseline` 行为
- `hash mismatch` 阻断

## 建议镜像

- PostgreSQL: `postgres:16-alpine`
- MySQL: `mysql:8.0`
- MariaDB: `mariadb:11`
