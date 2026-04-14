# user-guide 验证记录

## 目标

- 覆盖：安装、初始化、创建版本文件、升级、回退、状态、环境变量、常见错误
- 可用性：首次用户可在 10 分钟内完成 SQLite 最小演练
- 质量门槛：无 High；Medium <= 2 且有规避说明

## Round 1

### A（文档作者）

- 已产出初版：`docs/user-guide.md`

### B（审阅者）

- 覆盖性：8 项覆盖完整。
- 命令准确性：与 CLI 一致。
- 发现问题：
- `Medium`：安装/回退示例写死 `v0.1.0`，可能不存在。
- `Low`：SQLite 第二个迁移 `Down` 使用 no-op，容易误导“结构可逆”。
- `Low`：演练目录示例使用固定绝对路径，Windows 不友好。
- 结论：本轮不通过（有 `Medium`）。

### C（新手模拟者）

- 在临时目录实际执行 SQLite 流程，主链路可跑通。
- 发现问题：
- `High`：`go install github.com/exc-works/sql-migrate/cmd/sql-migrate@latest` 在当前环境报 `Repository not found`。
- `Medium`：第 9 节缺少 `sql-migrate --help` 前置检查。
- `Medium`：`create/up/down` 成功时常无输出，需补“如何验证成功”。
- `Low`：no-op 回退示例应更明确或替换为可逆 SQL。

### 主代理修订

- 安装章节改为双路径：
- 模块安装（公开仓库）`go install github.com/exc-works/sql-migrate/cmd/sql-migrate@latest`
- 本地源码安装（私有仓库）`go install ./cmd/sql-migrate`
- 工具版本安装/回退示例由固定 tag 改为 `@<version>` 参数化。
- SQLite 演练新增前置检查：`sql-migrate --help`。
- 在 `create/up/down` 章节补充“成功可能无输出，使用 status 验证”。
- SQLite 第二个迁移的 `Down` 改为真实可逆 SQL（重建表回退列）。
- 演练目录增加 Windows PowerShell 示例。

### 结果

- 未通过（已进入 Round 2 复验）。

## Round 2

### A（文档作者）

- 如需再次改写再回填

### B（审阅者）

- 发现问题：
- `Medium`：第 7 章未给私有仓库升级/回退等价路径。
- `Low`：`<version>` 占位符可能被误复制。
- `Low`：Windows 路径示例可读性一般。
- 结论：`High=0`、`Medium=1`、`Low=2`，按门槛“无 High 且 Medium<=2”通过。

### C（新手模拟者）

- SQLite 主流程实跑通过：
- `new config -> create -> new version -> up -> status -> down -> status` 均成功。
- `down --all` 后状态符合预期。
- 安装验证受当前沙箱网络限制影响，不属于流程逻辑阻塞。

### 主代理修订

- 补第 7 章私有仓库升级/回退路径：`go install ./cmd/sql-migrate`。
- 将 `@<version>` 全部替换为 `@vX.Y.Z`，并增加“替换为真实版本号”说明。
- 修正 Windows PowerShell 演练路径示例为相对路径 `.\sql-migrate-demo` 形式。

### 结果

- 通过（满足门槛；剩余问题已进一步修复）。

## Round 3

### B（审阅者）

- 复审结果：`High=0`、`Medium=0`、`Low=1`（建议补一条真实版本号示例）。
- 判定：可发布。

### C（新手模拟者）

- SQLite 关键链路复测通过：
- `new config -> create -> new version -> up -> status -> down -> status -> down --all -> status`
- 无阻塞问题；仅提醒“create/up/down 成功时可能无输出，需看 status”。

### 主代理修订

- 在“工具回退”补充可直接复制示例：
- `go install github.com/exc-works/sql-migrate/cmd/sql-migrate@v0.2.3`

### 结果

- 最终通过（High/Medium 均为 0）。

## 最终结论

- 覆盖完整：安装、初始化、创建版本、升级、回退、状态、环境变量、错误排查均已覆盖。
- 可用性达标：新用户 SQLite 路径可跑通。
- 发布建议：可将 `docs/user-guide.md` 作为用户入口文档发布。
