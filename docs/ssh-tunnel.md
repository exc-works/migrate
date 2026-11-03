# SSH 隧道与代理

SQL 和 full 构建均支持通过单台 SSH 堡垒机连接 PostgreSQL、MySQL、MariaDB。
可选 SOCKS5 或 HTTP CONNECT 代理用于连接堡垒机。所有连接由 Go 管理，无需安装 ssh。
配置中的数据库地址始终填写从堡垒机访问的原始数据库地址，TLS 参数保持原样。

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
    "known_hosts_file": "~/.ssh/known_hosts",
    "proxy": {
      "type": "socks5",
      "address": "127.0.0.1:1080"
    }
  }
}
```

不需要代理时删除 `proxy`；使用 HTTP CONNECT 时将 `type` 改为 `http`。
代理地址是 host:port，不是 URL。代理必须允许访问 SSH 端口。工具不会读取系统代理
环境变量或在代理失败后回退直连。

需要代理认证时，在 `proxy` 中增加 `username` 和 `password_env`（环境变量名）。
SOCKS5 支持用户名密码，HTTP CONNECT 支持 Basic；首期不支持 HTTPS 代理端点、
NTLM/Kerberos、MFA、代理链、ProxyJump 或 ProxyCommand。HTTP Basic 代理认证
发生在 SSH 握手之前；连接 HTTP 代理的这一段没有 TLS 保护，凭据仅应在可信网络使用。

## 密码、私钥和 agent

- `password`：从 `password_env` 指定的环境变量获取密码；未设置时在终端无回显输入。
- `private_key`：用 `key_file` 指定私钥文件。加密私钥通过 `passphrase_env` 获取口令；
  未设置时在终端输入。不要将私钥或口令写进配置。
- `agent`：不指定密码或私钥字段，使用 `SSH_AUTH_SOCK`。当前 agent 接入使用 Unix
  socket，Windows 用户请使用密码或私钥文件。

环境变量已设置但为空会报错，密码不会被去除首尾空格。登录失败不会自动重复提示。
代理密码与 SSH 密码独立。提示写入 stderr，状态 JSON 的 stdout 不包含密码提示。
终端输入支持退格、Ctrl+U 清空、Ctrl+C/Ctrl+D 取消；取消后恢复终端回显。
CI 中注入 secret 并显式禁止交互：

```bash
migrate --non-interactive up
```

缺失密码且无法使用终端时会报错。`password_env`、`passphrase_env` 的值是变量名，
不能写成 `${...}`；其余字符串配置支持已有环境变量模板。`port` 是 JSON 数字。
数据库密码继续通过 `DATABASE_DSN` 等现有配置传入，不会复用 SSH 密码。

## 路径、超时与失败处理

SSH 文件路径支持 `~` 和 `~/...`，指向运行工具的用户主目录；不支持 `~otheruser`。
相对 SSH 文件路径以迁移配置文件所在目录为基准。工具不读取 `~/.ssh/config`。
`known_hosts_file` 默认是 `~/.ssh/known_hosts`，必须事先准备好可信主机记录；未知或
变更的主机密钥直接失败，不会自动修改该文件。

可设置正数 duration：`connect_timeout` 默认 `10s`，`keepalive_interval` 默认 `15s`，
`keepalive_timeout` 默认 `5s`。数据库驱动自己的连接超时同样覆盖远端 SSH channel 建立。
查询超时继续使用驱动/context 机制。一个 channel 建立超时会使整个隧道失效，避免遗留
无法单独取消的 SSH channel 请求。

只支持一个 TCP 数据库目标；多主机 DSN、Unix socket 和其他数据库的 SSH 配置会明确
报错。未配置 SSH 时继续使用原来的直连功能。SSH 中断后停止后续操作，不重连、不重放 SQL。
断线可能发生在数据库已经完成 DDL/COMMIT 之后，请检查数据库和迁移记录再决定如何恢复。

`version`、`capabilities`、`new` 不启动隧道。`status`、`baseline`、`--dry-run` 保留原有
数据库读取需求，因此可能建立连接。命令成功、失败或取消时都会清理自建的连接和隧道。

## Go API

用 `migrate.OpenConnection(ctx, migrate.ConnectionOptions{...})` 创建连接，先调用
`connection.DB.PingContext(connection.Context())` 验证可用，再将 DB 和
`connection.Context()` 传给 `migrate.NewService`。调用方负责最后执行
`connection.Close()`，它先关闭连接池，再释放隧道。

库默认不读终端。需要交互时显式提供 `Prompt` 回调；已有调用方自行传入 `*sql.DB` 的
API 保持不变，连接生命周期仍由调用方管理。参见 [English guide](ssh-tunnel.en.md)。
