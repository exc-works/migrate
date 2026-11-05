# 運行指令碼示例

以下示例在倉庫檢出目錄中運行，均使用 SQLite，無須資料庫服務。
它們演示模塊解析，並在遷移內部校驗資料變更。

## 構建臨時 CLI

在倉庫根目錄執行；本地導入和 npm 示例繼續使用同一個 shell：

```sh
example_workspace=$(mktemp -d)
go build -o "$example_workspace/migrate" ./cmd/migrate
```

## 共享本地 helper {#local}

[瀏覽本地示例](https://github.com/exc-works/migrate/tree/main/examples/script-imports/local)。
其中兩條遷移導入遷移目錄之外的同一個 TypeScript helper。

```sh
export MIGRATE_EXAMPLE_DB="$example_workspace/local.sqlite"
"$example_workspace/migrate" -w examples/script-imports/local create
"$example_workspace/migrate" -w examples/script-imports/local up --dry-run
"$example_workspace/migrate" -w examples/script-imports/local up
"$example_workspace/migrate" -w examples/script-imports/local status
"$example_workspace/migrate" -w examples/script-imports/local down 2
"$example_workspace/migrate" -w examples/script-imports/local down --all
```

up 後三個版本均已應用，使用者名為 `alice` 和 `bob`。
`down 2` 保留版本 1、2；`down --all` 刪除示例表。

## 解析已安裝的包

[瀏覽 npm 示例](https://github.com/exc-works/migrate/tree/main/examples/script-imports/npm)。
Node/npm 僅用於安裝。Lockfile 引用倉庫內的本地包，沒有 registry 依賴，
因此下面的安裝可以離線完成：

```sh
npm --prefix examples/script-imports/npm ci --ignore-scripts --no-audit --no-fund --offline
export MIGRATE_EXAMPLE_DB="$example_workspace/npm.sqlite"
"$example_workspace/migrate" -w examples/script-imports/npm create
"$example_workspace/migrate" -w examples/script-imports/npm up
"$example_workspace/migrate" -w examples/script-imports/npm status
"$example_workspace/migrate" -w examples/script-imports/npm down --all
```

第二條遷移通過 exports 子路徑導入 `@migrate-example/normalizer/names`，
並校驗歸一化後的名稱 `alice`。up 後兩個版本均已應用。

## 將 helper 嵌入 Go 二進制

[嵌入式應用](https://github.com/exc-works/migrate/tree/main/examples/script-imports/embedded)
將版本檔案和共享模塊一起放入 `embed.FS`。它使用私有內存資料庫，無須安裝 npm 依賴：

```sh
go run ./examples/script-imports/embedded
```

預期輸出：

```text
after up: alice
after down: all migrations pending
```

## 清理與驗證

手動完成 CLI 示例後，刪除臨時資料庫與可執行檔案：

```sh
rm -r "$example_workspace"
```

npm 安裝還會留下已忽略的 `examples/script-imports/npm/node_modules` 目錄。
需要自動檢查並自行管理臨時副本時，在倉庫根目錄執行 `./scripts/test-script-examples.sh`。

共享代碼的導入方式，以及修改依賴對舊遷移的影響，見[共享模塊](./modules)。
