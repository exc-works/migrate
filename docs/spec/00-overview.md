# 00 Overview

## 目标

在当前项目实现独立的 `sql-migrate` 工具，并满足：

- 保留核心命令能力（create/baseline/status/up/down/new）。
- 支持常见数据库（V1: PostgreSQL/MySQL/MariaDB）。
- 使用 `testcontainers-go` 完成集成测试。
- 支持 `go install` 安装。
- 增加 GitHub Actions CI。

## 交付物

- 可构建的 CLI：`cmd/sql-migrate`。
- 结构化实现：`internal/*`。
- 单测与集成测试。
- CI 工作流。
- 安装与发布说明。

## Definition of Done

全部满足才算完成：

- `go build ./...` 通过。
- 单元测试通过。
- `go test -tags=integration ./...` 通过。
- PR 上 CI 全绿。
- 已修复 `01` 分册中的关键问题。

## 推荐执行顺序

1. 读 `01`，先确认风险与兼容边界。
2. 按 `02 + 03` 完成功能与架构。
3. 按 `04` 完善测试。
4. 按 `05` 接入 CI 与安装方式。
5. 按 `06` 做安全与审查收口。
