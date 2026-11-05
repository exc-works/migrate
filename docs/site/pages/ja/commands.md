# CLI リファレンス

インストール済みの引数は `migrate --help`、`migrate <command> --help`、バージョンは `migrate version` で確認します。

## 共通オプション

| オプション | 動作 |
| --- | --- |
| `-c`、`--config` | 指定ファイルを使用。省略時は[検索順序](./configuration#configuration-file-search-order)に従う |
| `-w`、`--working-dir` | 作業ディレクトリを指定し、設定内の値を上書き |
| `--non-interactive` | 認証情報の入力や未知の SSH 鍵の承認を求めずに失敗 |

## 設定とファイルの生成

```sh
migrate new config [filename] [--format yaml|json] [--multi] [--shards N] [--ssh-tunnel] [--force]
migrate new config shards.yml --multi --shards 4
migrate new config shards-ssh.yml --multi --ssh-tunnel --proxy socks5
```

既定は `migration_config.yml`。`--shards N` は `--multi` と正の数が必要です。`--proxy` は `--ssh-tunnel` と併用します。`--ssh-auth agent|private_key|password` で認証を選択。上書きは `--force`、JSON は `--format json` または `.json` を使用します。

```sh
migrate new version <description> [-v <version>] [--format sql|js|ts|mjs]
```

既定は SQL、自動バージョンです。空白を含む説明は `"init users"` のように引用し、ファイル名では `_` になります。

## データベース操作

```sh
migrate create
```

履歴テーブルを作成するだけで、ファイルは実行しません。

```sh
migrate baseline
```

実 DB がファイルの内容と一致するときだけ、実行せず[ベースライン](./history#baseline)を記録します。

```sh
migrate up [--dry-run]
```

未適用バージョンを順に実行。dry-run は履歴への接続を行う場合がありますが変更は適用しません。

```sh
migrate down <to-version> [--dry-run]
migrate down --all [--dry-run]
```

対象より高いバージョンだけを戻すか、すべて戻します。対象と `--all` は排他的で、どちらかが必要です。

```sh
migrate status [--output table|json]
```

既定は表形式。状態は[履歴](./history)を参照してください。複数接続の JSON は名前順で `data_source`、`status` と、成功時の `migrations` または失敗時の `error` を返します。

`create`、`baseline`、`status`、`up`、`down` は全接続を処理し、`--parallel N`（既定 8）を使用できます。失敗は他を中止せず、いずれかが失敗すると終了コードは非ゼロです。

## 接続不要の情報

```sh
migrate capabilities [--output text|json]
```

ファイル形式、名前、スクリプト対応、トランザクション制限を表示します。既定は `text`。

```sh
migrate version
```

DB の移行バージョンではなくツールのバージョンを表示します。
