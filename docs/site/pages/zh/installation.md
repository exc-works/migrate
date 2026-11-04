# 安装

默认构建包含 SQL、JavaScript、TypeScript 和 MJS 迁移。
内嵌脚本引擎不依赖 Node.js，也不会下载解释器。

## 通过 Go 安装

```sh
go install github.com/exc-works/migrate/cmd/migrate@latest
```

固定或降级工具版本时，将 `@latest` 替换成已发布的版本标签。
从源码安装时，在仓库根目录执行 `go install ./cmd/migrate`，无需额外构建标签。

## 使用发布的二进制

从 [GitHub Releases](https://github.com/exc-works/migrate/releases) 下载符合操作系统
和架构的文件。每个平台只发布一种包含全部迁移格式的二进制，无构建后缀。
校验发布的 SHA-256 后，将可执行文件改名为 `migrate`（Windows 为
`migrate.exe`），放入 PATH 包含的目录。

## 验证安装结果

```sh
migrate --help
migrate version
migrate capabilities --output json
```

`version` 显示二进制版本，源码安装通常显示 `dev`。
`capabilities` 无须连接数据库，就能列出迁移格式、数据库验证状态和事务限制。
具体含义见[数据库能力](./databases)。

macOS/Linux 上通过 Go 安装后仍找不到命令时，将 Go 的二进制目录加入 PATH：

```sh
export PATH="$(go env GOPATH)/bin:$PATH"
```

接下来可以完成[快速开始](./quick-start)，或直接[配置项目](./configuration)。
