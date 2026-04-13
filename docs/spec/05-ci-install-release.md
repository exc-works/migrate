# 05 CI, Install, Release

## GitHub Actions

新增：`.github/workflows/ci.yml`。

至少包含 3 个 job：

1. `unit`
- `go test ./...`
- `go vet ./...`

2. `integration`（matrix: postgres/mysql/mariadb）
- `go test -tags=integration ./...`

3. `build`
- `go build ./...`
- `go build ./cmd/sql-migrate`

触发：

- `pull_request`
- `push` 到默认分支

门禁：

- 任一 job 失败即阻断合并。

## go install 支持

- CLI 入口：`cmd/sql-migrate`。
- 安装命令必须可用：
  - `go install <module>/cmd/sql-migrate@latest`
  - `go install <module>/cmd/sql-migrate@vX.Y.Z`

## 发布约束

- 使用语义化 tag（例如 `v0.1.0`）。
- 变更日志明确区分“行为修复”与“功能新增”。
