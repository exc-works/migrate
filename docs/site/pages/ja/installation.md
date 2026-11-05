# インストール

Go または[リリースバイナリ](https://github.com/exc-works/migrate/releases)を使用します。OS と CPU に合うファイルを選び、公開された SHA-256 を確認して PATH に配置してください。Windows の実行ファイル名は `migrate.exe` です。

## モジュールからインストール

```bash
go install github.com/exc-works/migrate/cmd/migrate@latest
```

特定バージョンをインストールする場合:

```bash
go install github.com/exc-works/migrate/cmd/migrate@vX.Y.Z
```

## ローカルソースからインストール（プライベートリポジトリまたは社内ネットワーク）

リポジトリのルートで実行します:

```bash
go install ./cmd/migrate
```

## インストール確認

```sh
migrate --help
migrate version
migrate capabilities --output json
```

コマンドが見つからない場合:

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

`version` はインストール済みのリリースを表示します。ローカルビルドは `dev` です。`capabilities` は接続せずに対応状況を表示します。

続いて[クイックスタート](./quick-start)または[設定](./configuration)に進みます。
