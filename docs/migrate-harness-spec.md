# Migrate Harness 规范索引

该文件作为总入口，详细规范已拆分到 `docs/spec/`，用于降低单文档上下文负担，方便 AI Agent 分步执行。

## 阅读顺序

1. [00-overview.md](./spec/00-overview.md)
2. [01-quality-gates.md](./spec/01-quality-gates.md)
3. [02-functional-cli-spec.md](./spec/02-functional-cli-spec.md)
4. [03-architecture-dialect.md](./spec/03-architecture-dialect.md)
5. [04-testing-testcontainers.md](./spec/04-testing-testcontainers.md)
6. [05-ci-install-release.md](./spec/05-ci-install-release.md)
7. [06-reviewer-security-rules.md](./spec/06-reviewer-security-rules.md)
8. [07-script-migrations.md](./spec/07-script-migrations.md)
9. [08-script-modules-plan.md](./spec/08-script-modules-plan.md)（模块导入 PR stack 与验收记录）
10. [09-ssh-tunnel.md](./spec/09-ssh-tunnel.md)（SSH 隧道、代理与交付约束）
11. [10-multi-database-migrations.md](./spec/10-multi-database-migrations.md)（同一套迁移并发执行到多库的技术方案）
12. [11-multi-database-pr-stack.md](./spec/11-multi-database-pr-stack.md)（多库编排 PR stack 与分层验收计划）

## 使用方式

- 实施时一次只加载 1-2 个分册。
- 先完成 `00 + 01`，再进入实现与测试。
- Reviewer 审核必须额外加载 `06`。
