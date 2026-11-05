# 安裝

通過 Go 安裝 migrate，或下載發佈的二進制。

## 通過 Go 安裝

```sh
go install github.com/exc-works/migrate/cmd/migrate@latest
```

固定或降級工具版本時，將 `@latest` 替換成已發佈的版本標籤。
從源碼安裝時，在倉庫根目錄執行 `go install ./cmd/migrate`。

## 使用發佈的二進制

從 [GitHub Releases](https://github.com/exc-works/migrate/releases) 下載符合操作系統
和架構的檔案。
校驗發佈的 SHA-256 後，將可執行檔案改名為 `migrate`（Windows 為
`migrate.exe`），放入 PATH 包含的目錄。

## 驗證安裝結果

```sh
migrate --help
migrate version
migrate capabilities --output json
```

`version` 顯示發佈二進制的版本號；通過 `go install ...@版本` 安裝時顯示實際模塊版本。
本地開發構建顯示 `dev`。
舊版本通過 `go install` 安裝後可能顯示 `dev`，可用 `go version -m <binary>`
查看實際安裝的模塊版本。
`capabilities` 無須連接資料庫，就能列出遷移格式、資料庫驗證狀態和事務限制。
具體含義見[支持的資料庫](./databases)。

macOS/Linux 上通過 Go 安裝後仍找不到命令時，將 Go 的二進制目錄加入 PATH：

```sh
export PATH="$(go env GOPATH)/bin:$PATH"
```

接下來可以完成[快速開始](./quick-start)，或直接[設定項目](./configuration)。
