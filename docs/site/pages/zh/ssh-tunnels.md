# SSH 隧道与代理

通过单台 SSH 堡垒机连接数据库，也可先经一个 SOCKS5 或 HTTP CONNECT 代理连接堡垒机。
连接由 Go 管理，无须外部 ssh 进程，也不读取 OpenSSH 配置。

## 配置连接路径

`DATABASE_DSN` 保留原始数据库地址与 TLS 设置，堡垒机需要能访问该地址。
在项目配置中添加：

```json
{
  "dialect": "postgres",
  "data_source_name": "${DATABASE_DSN}",
  "ssh_tunnel": {
    "host": "${SSH_HOST}",
    "port": 22,
    "user": "${SSH_USER}",
    "auth": {
      "type": "password",
      "password_env": "MIGRATE_SSH_PASSWORD"
    },
    "known_hosts_file": "~/.ssh/known_hosts"
  }
}
```

交互式 CLI 遇到未知主机密钥时，会显示 SSH 地址、算法和 SHA256 指纹。
独立核对指纹后，显式选择“信任并保存”；默认选项为拒绝。
只有批准后才追加记录，并以 0600/0700 权限创建缺失的文件或目录，保留已有记录。
变更或撤销的密钥、无效文件直接失败，不提供信任提示。
非交互模式不会批准或写入未知密钥，CI 执行前应预先准备可信记录。

## 选择认证方式

| `auth.type` | 字段 | 行为 |
| --- | --- | --- |
| `password` | `password_env` | 从指定环境变量读取密码 |
| `private_key` | `key_file`，可选 `passphrase_env` | 读取本地私钥；加密私钥的口令来自指定变量 |
| `agent` | 不设置密码或私钥字段 | 使用 Unix `SSH_AUTH_SOCK` socket |

Secret 字段填写 `MIGRATE_SSH_PASSWORD` 这样的变量名，不能填写凭据或 `${...}` 模板。
变量未设置且存在控制终端时，可以无回显输入；显式空值会失败。
认证失败不会重复提示。Windows 用户应选择密码或私钥文件。

CI 中提前注入凭据，再运行 `migrate --non-interactive up`。
提示写入 stderr，不影响 stdout JSON。库调用方通过 `ConnectionOptions.Prompt` 显式启用凭据输入。
独立的 `HostKeyPrompt` 接收 `SSHHostKeyInfo`，应在核对指纹后才返回 true，批准保存新密钥。
空回调不提示、不信任未知密钥；回调必须响应取消。见[在 Go 中使用](./go-library)。

## 添加可选代理

在 `ssh_tunnel` 内放入 `proxy` 对象：

```json
{
  "type": "socks5",
  "address": "127.0.0.1:1080"
}
```

HTTP CONNECT 使用 `"type": "http"`；address 是 host:port，不是 URL。
可选认证要求同时配置 `username` 和 `password_env`。SOCKS5 支持用户名密码，
HTTP CONNECT 支持 Basic。HTTP 代理认证发生在 SSH 加密之前，没有 TLS 保护，
只应在可信网络使用。

代理失败不会回退直连。不支持系统代理变量、HTTPS 代理端点、NTLM/Kerberos、MFA、
代理链或 SSH 跳板链。

## 路径与超时

SSH 文件路径支持 `~` 和 `~/...`，相对路径从迁移配置文件所在目录解析。
默认 known-hosts 为 `~/.ssh/known_hosts`。其他非敏感字符串支持环境模板，`port` 为 JSON 数字。

| 设置 | 默认值 |
| --- | --- |
| `connect_timeout` | `10s` |
| `keepalive_interval` | `15s` |
| `keepalive_timeout` | `5s` |

Duration 必须为正。等待主机密钥确认时暂停网络握手超时，调用方 context 仍可取消。
驱动连接期限也覆盖远端 channel 建立；一个 channel 建立超时会使整个路径失效。
传输中断会取消执行，不重连、不重放 SQL。DDL 或 COMMIT 可能已经完成，恢复前先检查数据和历史。

## 数据库连接支持

PostgreSQL、MySQL、MariaDB、MSSQL、Oracle、ClickHouse、TiDB 与 Redshift 支持单个 TCP
目标。多主机 DSN、Unix socket、跳转到其他地址以及 SQLite 隧道配置会明确失败。

- TiDB 使用 MySQL 接线，Redshift 使用 PostgreSQL 接线。
- MSSQL 要求 TCP host/port（默认 1433），不支持命名实例发现、failover 或显式 MultiSubnetFailover。
- Oracle 要求带端口的单主机 `oracle://` URL，支持 TCP/TCPS；不支持描述符、额外 SERVER、OOB 或外部认证。
- ClickHouse 支持 Native 与 HTTP(S)，拒绝数据库 `http_proxy`，数据库请求禁用
  `HTTP_PROXY`/`HTTPS_PROXY`，原 DSN 中的 TLS 身份保持不变。

`version`、`capabilities`、`new` 不启动 SSH；`status`、`baseline` 和 dry-run 可能为读取历史连接数据库。
CLI 创建的连接池和隧道在成功、失败或取消时均会关闭。
