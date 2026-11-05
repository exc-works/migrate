# サンプルの実行

リポジトリのチェックアウトから実行します。すべて SQLite を使い、移行内でデータを検証します。

## 一時 CLI を作成

同じシェルを開いたままにしてください。

```sh
example_workspace=$(mktemp -d)
go build -o "$example_workspace/migrate" ./cmd/migrate
```

## ローカルの共有ファイル {#local}

[ソース](https://github.com/exc-works/migrate/tree/main/examples/script-imports/local)には共有 TypeScript を使う移行が含まれます。

```sh
export MIGRATE_EXAMPLE_DB="$example_workspace/local.sqlite"
"$example_workspace/migrate" -w examples/script-imports/local create
"$example_workspace/migrate" -w examples/script-imports/local up --dry-run
"$example_workspace/migrate" -w examples/script-imports/local up
"$example_workspace/migrate" -w examples/script-imports/local status
"$example_workspace/migrate" -w examples/script-imports/local down 2
"$example_workspace/migrate" -w examples/script-imports/local down --all
```

up 後は 3 バージョンが適用され、名前は `alice` と `bob` です。`down 2` は 1 と 2 を保持し、`down --all` はテーブルを削除します。

## npm パッケージ

インストールにだけ Node/npm が必要です。lockfile は同梱のローカルパッケージを参照し、オフラインで実行できます。

```sh
npm --prefix examples/script-imports/npm ci --ignore-scripts --no-audit --no-fund --offline
export MIGRATE_EXAMPLE_DB="$example_workspace/npm.sqlite"
"$example_workspace/migrate" -w examples/script-imports/npm create
"$example_workspace/migrate" -w examples/script-imports/npm up
"$example_workspace/migrate" -w examples/script-imports/npm status
"$example_workspace/migrate" -w examples/script-imports/npm down --all
```

2 番目の移行は `@migrate-example/normalizer/names` を読み込み、`alice` を確認します。up 後は 2 バージョンが適用済みです。

## Go への埋め込み

```sh
go run ./examples/script-imports/embedded
```

`embed.FS` と専用のメモリ DB を使い、npm は不要です。出力は次のとおりです。

```text
after up: alice
after down: all migrations pending
```

## 後片付け

```sh
rm -r "$example_workspace"
```

npm の `node_modules` は残ります。自動検証にはルートから `./scripts/test-script-examples.sh` を実行します。[依存の再現性](./modules#checksums)も参照してください。
