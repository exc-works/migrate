# 同一套迁移并发执行到多个数据库

状态：已在堆叠 PR 中实现，等待各层当前 head/base 的 CI 与审核收口。交付记录见
[11-multi-database-pr-stack.md](11-multi-database-pr-stack.md)。

## 实现原则

在现有单库功能外增加一个编排层：遍历配置中的数据源，以有界并发为每个
数据源创建独立 Service，调用现有命令，关闭连接，最后汇总结果。

编排层只负责“在哪些数据库上运行、同时运行多少个、结果是什么”。版本选择、
历史校验、SQL 解析、脚本编译、事务和实际执行继续由现有 Service 负责。
不拆分 Up/Down，不增加共享编译、全局预检、迁移锁或新的计划状态机。

## 配置

在现有 FileConfig 中增加可选 data_sources，共享原有迁移设置：

```yaml
dialect: postgres
migration_source: migrations
schema_name: migration_schema
migrate_out_of_order: false

# 所有数据源默认共用这份 SSH 配置
ssh_tunnel:
  host: "${SSH_HOST}"
  user: "${SSH_USER}"
  auth:
    type: agent

data_sources:
  shard_0:
    data_source_name: "${SHARD_0_DSN}"
  shard_1:
    data_source_name: "${SHARD_1_DSN}"
```

- 原单源配置继续使用顶层 data_source_name 和可选 ssh_tunnel，行为保持兼容。
- 多源配置的每项包含 DSN 和可选 ssh_tunnel；方言、目录、历史表和脚本限制
  均使用顶层设置。仅顶层 data_source_name 与 data_sources 互斥，空集合报错。
- 顶层 ssh_tunnel 是所有数据源共用的默认 SSH 配置，只需填写一次。数据源级
  非 null 的 ssh_tunnel 对象优先，整体替换顶层 SSH 配置，不逐字段合并；
  未提供或为 null 时继承顶层配置，两处都没有时直连。
- 沿用当前默认值、环境变量替换及路径规则，不引入配置版本。
- 配置读取时校验所有数据源，并复用既有 DSN/SSH 解析；配置无效在联网前报错。
  目标名用于日志和结果，必须非空；日志输出需安全转义名称中的控制字符。
- 将共享设置和每项连接设置展开为独立的有效 FileConfig，按名称排序。
  每份配置清除 data_sources，避免递归展开；SSH 可变选项不能跨任务共享修改。

## 配置格式与查找

新版本默认使用 YAML，保留严格 JSON 读取。注释使用 YAML 原生的 # 语法，
不再引入 JSONC 或自定义 JSON 注释预处理器。

- 未显式传 -c/--config 时，按 migration_config.yml、migration_config.yaml、
  migration_config.json 的顺序查找；只在候选文件不存在时尝试下一个。
- 多个候选同时存在时使用优先级最高的一份，不合并配置。找到的文件若解析、
  权限、环境展开或校验失败，直接报错，不回退到另一份配置。
- 显式 -c 指定文件时只读取该文件，即使它不存在也不搜索其他路径。
- .yml/.yaml 使用 YAML 解析；.json 使用既有 encoding/json。显式指定其他
  扩展名或无扩展名的旧配置文件时，仍按原来的 JSON 方式读取。
- --working-dir 的路径语义保持原样，自动查找在相同工作目录内进行；SSH 相对
  路径仍以最终选中的配置文件目录为基准。new version 仅在没有候选配置时使用
  原有默认设置，不能忽略已存在配置的语法或校验错误。

读取端复用仓库已有的 go.yaml.in/yaml/v3，将其固定版本提升为直接依赖。
FileConfig、数据源配置和 tunnel.Options 等嵌套配置补齐与 JSON 同名的 yaml tag，
两种格式解码到同一结构后共享默认值、环境变量替换、SSH 继承和业务校验。
YAML 使用 Decoder 读取一个映射文档并确认后续 EOF，拒绝重复 key 和多文档，
类型或语法错误只返回安全位置与摘要，不输出可能包含凭据的原始错误片段。
纯 JSON 的错误定位和尾随文本检查保留。不要在 YAML 失败后再试 JSON 解析。

CLI 不再用固定的 JSON 默认 flag 值掩盖用户是否显式指定 -c；由配置读取入口
负责默认文件查找，并在 --help 中说明优先级。

## 命令

配置为多源时，现有数据库命令自动遍历全部数据源，只新增 --parallel：

```bash
migrate up --parallel 8
migrate status --parallel 8 --output json
migrate down 202610010001 --parallel 8
migrate down --all --parallel 8
migrate up --parallel 8 --dry-run
```

parallel 默认 8，必须为正整数，设为 1 即串行；单源仍只运行一次。
create/baseline/status/up/down 使用同一编排入口；down 的参数原样传给各 Service。
不增加 --all-targets、目标筛选、失败策略开关或批量超时配置。

new version 仍在共享目录生成一份文件；new config 根据初始化参数生成单源或多源
模板，并可附带顶层共享 SSH 和 proxy 设置。version/capabilities 保持原行为。
无需数据库的命令不遍历数据源或建立连接。

## 配置初始化与注释

用户应通过 new config 直接生成对应场景的可编辑模板，不必从其他文档手写
SSH/proxy 结构。默认生成带注释的单源 YAML，默认文件名为 migration_config.yml。

```bash
migrate new config
migrate new config shards.yml --multi
migrate new config shards-four.yml --multi --shards 4
migrate new config shards-ssh.yml --multi --ssh-tunnel
migrate new config shards-key.yml --multi --ssh-tunnel --ssh-auth private_key
migrate new config shards-proxy.yml --multi --ssh-tunnel --proxy socks5
migrate new config shards-http.yml --multi --ssh-tunnel --proxy http
migrate new config --format json --multi
```

初始化参数如下，均只作用于 new config：

| 参数 | 生成行为 |
| --- | --- |
| --format yaml\|json | 未指定文件名时默认 yaml，生成 migration_config.yml；选择 json 时生成 migration_config.json |
| --multi | 生成命名分片，默认 shard_0、shard_1，使用各自 SHARD_0_DSN、SHARD_1_DSN 占位符 |
| --shards N | 指定分片数量，默认 2，必须为正整数且与 --multi 一起使用；编号从 0 到 N-1 |
| --ssh-tunnel | 生成顶层共享 ssh_tunnel；默认 agent 认证 |
| --ssh-host、--ssh-port、--ssh-user | 可选 SSH 地址参数；未提供 host/user 时生成环境变量占位符，port 默认为 22 |
| --ssh-auth agent\|private_key\|password | 选择认证模板，仅生成该模式需要的字段 |
| --ssh-key-file、--ssh-passphrase-env | private_key 模式的密钥路径与可选口令环境变量名；未提供路径时生成 ${SSH_KEY_FILE} |
| --ssh-password-env | password 模式的密码环境变量名，默认为 SSH_PASSWORD |
| --ssh-known-hosts-file | 可选 known_hosts 路径，默认沿用现有默认值 |
| --proxy socks5\|http | 在 ssh_tunnel 内生成指定类型 proxy；要求同时启用 --ssh-tunnel |
| --proxy-address | 可选代理地址；未填写时生成 ${SSH_PROXY_ADDRESS} 占位符 |
| --proxy-username、--proxy-password-env | 可选代理认证，用户名和密码环境变量名必须成对提供 |
| --force | 保留现有覆盖行为；默认拒绝覆盖已有文件 |

指定文件名时从 .yml/.yaml/.json 推断格式；其他扩展名或无扩展名的旧式文件名
按 JSON 处理，与旧配置读取行为一致。显式 --format 与文件名推断不一致时报错，
不生成扩展名与内容不匹配的配置。已有 JSON 不自动转换或删除；new config 只
创建选定的目标文件，同名文件的覆盖仍须 --force。

显式指定的 SSH 从属参数要求 --ssh-tunnel；proxy 从属参数要求 --proxy。错误认证模式、
代理类型、端口、环境变量名、认证字段组合或参数依赖在创建/覆盖文件前报错。
传入路径和字符串值必须经对应格式的编码器处理，不能直接拼接到模板文本。
new config --help 列出参数用途、认证/代理枚举和典型组合命令。

未填写的连接信息生成环境变量占位符；初始化时不展开环境变量、不读取现有
迁移配置、不验证数据库连通性、不进行 SSH 认证或终端秘密提示。密码和密钥
口令只允许指定环境变量名称，不提供明文秘密参数。生成器不写入私钥内容。
只生成所选模式的有效结构，其他可选字段、默认值和填写方式通过注释说明；
不生成多个互斥 auth 字段让用户自行删除。proxy 在顶层共享 SSH 内只需配置一次。

YAML 模板通过 yaml.Node 的注释字段或等价的结构化生成方式附加字段说明，
注释覆盖共享字段、DSN 环境变量、SSH 继承、认证模式、known_hosts 信任规则、
超时/keepalive 和 proxy 类型/认证；未启用 SSH/proxy 时附简短启用命令提示，
不放入会触发连接的示例配置对象。注释中的示例变量不参与环境变量展开。
字符串正确编码，特别验证冒号、#、引号、换行和反斜杠；DSN 与环境变量引用
作为字符串保存。JSON 模板使用标准编码器输出严格 JSON，不带注释。

例如 --multi --ssh-tunnel --proxy socks5 生成的关键结构如下，实际模板补全
字段说明和既有默认选项：

```yaml
# 所有数据源共用同一份迁移目录和历史表名称
dialect: postgres
migration_source: migrations
schema_name: migration_schema

# SSH 配置只写一次，各数据源默认继承
ssh_tunnel:
  host: "${SSH_HOST}"
  port: 22
  user: "${SSH_USER}"
  # 使用本机 SSH agent；其他认证模板由 --ssh-auth 生成
  auth:
    type: agent
  # proxy 用于连接 SSH 服务器；需要认证时填写用户名和密码环境变量名
  proxy:
    type: socks5
    address: "${SSH_PROXY_ADDRESS}"

# 运行前分别设置两个数据库的 DSN 环境变量
data_sources:
  shard_0:
    data_source_name: "${SHARD_0_DSN}"
  shard_1:
    data_source_name: "${SHARD_1_DSN}"
```

## 编排流程

```text
读取配置，展开数据源
  -> 固定数量 worker 遍历全部数据源
       -> 创建该数据源的连接和 Service
       -> 调用现有 Create/Baseline/Status/Up/Down
       -> 关闭该任务拥有的连接和 SSH transport
       -> 返回该数据源结果
  -> 汇总输出，返回总体退出码
```

新增小型 internal/batch 包，只提供有界并发遍历和结果收集。它接收目标名称列表、
并发数及执行回调，不依赖迁移计划、数据库方言或事务实现。CLI 的回调负责
将目标配置交给现有 Service 构造代码和命令方法。

worker 数量为 min(parallel, 数据源数量)。开始处理任务才打开连接，清理完成才
处理下一个目标，不为所有数据源预先创建 Service 或连接。复用现有 direct/SSH
工厂及 cleanup；工厂初始化失败也要清理已经创建的资源。

共用顶层 SSH 配置不要求调用方重复填写，也不改变现有 Connection 对 transport
的所有权：每个任务使用继承后的配置建立自己的 SSH 连接和转发，结束时自行关闭。
首版不增加跨任务共享运行时 SSH transport 的连接池或引用计数。

每个 Service、context、日志标签和输出缓冲均独立，不能让 worker 共同修改
Cobra command、共享 svc 变量或 cleanup。父 context 取消所有任务；SSH transport
失败仅取消对应 Service，不把它的 context 设置成整个批次的 command context。
终端密码和 host-key 回调串行调用，其他连接和执行仍可并发，保留原交互规则。

同一库内部仍按现有版本顺序执行。各 Service 使用相同目录，但各自按现有逻辑
加载文件、编译脚本、查询自己的历史，得到自己的 pending 列表。不共享 VM 或事务，
运行期间保持迁移目录及依赖不变。

## 失败与结果

每库独立执行，一个数据源连接或迁移失败后记录结果，继续处理其余数据源。
这是统一的遍历语义，不增加 fail-fast/continue 两套模式。全部清理和汇总后，
全部成功返回 0，任一失败返回 1。

父 context 取消后停止派发新任务，并取消活跃任务。每个数据源最终都有
succeeded、failed 或 canceled 结果，尚未开始的目标也标记 canceled。
执行错误与 cleanup 错误一起保留；成功迁移后 cleanup 失败也返回失败。

结果只需目标名、状态、安全错误摘要；status 另附原有迁移状态列表，dry-run
另附原有计划文本。不统计 Service 尚未暴露的已执行步骤数，不新增复杂错误码。
收集器按目标名排序后统一输出，多源 JSON 使用结果数组；原单源 JSON 保持不变。
日志带数据源标签并写 stderr，stdout 只输出结果，各任务不能并发混写计划或 JSON。
原始驱动错误、DSN、SQL 参数、凭据和本地绝对路径不能直接进入结果或日志。

不做全局预检屏障，也不回滚其他库的成功结果。重跑时各 Service 按现有历史
跳过成功迁移；非事务失败或提交结果不明的恢复方式沿用当前约定。

## 兼容与并发边界

迁移核心、source/script/dialect 接口、历史表和公共 NewService API 保持现有行为。
编排层不因方言新增拒绝规则，当前可执行的数据库继续通过原 Service 执行，
同时保留各方言已有的实验性标记和事务限制。

不同任务应对应独立数据库；配置时避免将同一个库以不同名称重复列入。
跨进程或跨命令迁移同一数据库仍遵守现有操作者串行化约定。本次“并发”指一个
命令中的不同数据库并行，不新增对同库重复目标或重叠发布的互斥保证。

首版不增加公共 BatchRunner API。Go 调用方继续使用现有 NewService；
公共批量 API、迁移锁、编译缓存和按实例限流在出现明确需求后独立评估。

## 验收

- 旧配置、旧命令、原单源输出保持兼容；多源配置可正确展开 DSN/SSH 和共享选项。
- 顶层 SSH 可与 data_sources 共存；未配置局部 SSH 时继承，局部配置整体覆盖，
  无 SSH 配置时直连；展开后的可变 SSH 选项相互隔离，不修改顶层模板。
- 初始化覆盖单源/多源、三种 SSH 认证、两种 proxy 及代理认证组合；默认带注释，
  --format json 为严格 JSON；格式与扩展名冲突、非法参数和无 --force 时不覆盖已有文件。
- 所有模板在提供测试环境变量后可被读取，初始化本身无需环境变量或任何网络。
  YAML/JSON 解码结果一致，覆盖注释、特殊字符、嵌套 YAML tag、重复 key 和多文档。
- 查找覆盖 yml/yaml/json 优先级、旧 JSON 回退、显式 -c、--working-dir 与无配置；
  已找到的 YAML 内容无效或无权限时不得回退，new version 不吞掉配置错误。
- 多个数据库复用同一目录，各自历史进度不同，执行和重跑结果仍正确。
- 使用通道屏障验证任务实际重叠，活跃目标数不超过 parallel；大量目标不会
  预先打开全部连接，不依赖固定 sleep 判断并发。
- 一个目标失败，其余目标仍执行；父取消停止剩余任务；连接和 SSH 在成功、
  失败、初始化失败和取消路径均清理，关闭错误可见。
- status JSON 与 dry-run 无输出交错，日志带目标名且安全脱敏；SSH 终端提示不并发。
- 配置和 worker 各自有单测，CLI 使用临时 SQLite 多库验证完整命令；沿用现有
  testcontainers 设施增加代表性网络数据库和 SSH 回归，不重新建设方言测试矩阵。
- 并发路径运行 go test -race，所有 PR 遵守仓库五项 merge gates。
