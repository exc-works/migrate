# migrate ユーザーガイド

本ガイドは初めて利用するユーザー向けです。コマンドとフラグは現在の実装（`cmd/migrate`）に基づいています。

## 1. インストール

### 1.1 モジュールからインストール

```bash
go install github.com/exc-works/migrate/cmd/migrate@latest
```

特定バージョンをインストールする場合:

```bash
go install github.com/exc-works/migrate/cmd/migrate@vX.Y.Z
```

`vX.Y.Z` を実際のバージョン（例: `v0.2.3`）に置き換えてください。

### 1.2 ローカルソースからインストール（プライベートリポジトリまたは社内ネットワーク）

リポジトリのルートで実行します:

```bash
go install ./cmd/migrate
```

### 1.3 インストール確認

```bash
migrate --help
```

コマンドが見つからない場合:

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

`Repository not found` と表示される場合は、上記のローカルソースからのインストール手順を使用してください。

## 2. 初期化

### 2.1 設定ファイルを生成

```bash
migrate new config
```

任意:

```bash
migrate new config dev.json
migrate new config --force
```

デフォルト設定テンプレート:

```json
{
  "schema_name": "migration_schema",
  "dialect": "postgres",
  "data_source_name": "host=127.0.0.1 port=5432 user=${DB_USER:postgres} password=${DB_PASSWORD:change_me} dbname=${DB_NAME:postgres} sslmode=disable",
  "working_directory": "",
  "migrate_out_of_order": false,
  "logger_level": "info",
  "migration_source": "migrations"
}
```

### 2.2 主要な設定項目を更新

- `dialect`: `postgres`, `mysql`, `mariadb`, `oracle`, `sqlite`, `mssql`, `clickhouse`, `tidb`, `redshift`
- `data_source_name`: DB接続文字列
- `migration_source`: マイグレーションディレクトリ（デフォルト: `migrations`）

### 2.3 マイグレーション履歴テーブルを初期化

```bash
migrate create
```

`create` は出力なしで成功する場合があります。次で確認してください:

```bash
migrate status
```

既存スキーマがあり、古いSQLを再実行したくない場合は、次を使用します:

```bash
migrate baseline
```

## 3. マイグレーションのバージョンファイルを作成

### 3.1 自動生成バージョン

```bash
migrate new version init_users
```

### 3.2 明示的バージョン

```bash
migrate new version add_email -v 202604140002
```

生成されるファイル名形式:

```text
V<version>__<description>.sql
```

デフォルトのファイルテンプレート:

```sql
-- +migrate Up

-- +migrate Down
```

## 4. アップグレード（マイグレーション適用）

まずドライラン:

```bash
migrate up --dry-run
```

実際に適用:

```bash
migrate up
```

その後ステータスを確認:

```bash
migrate status
```

`up` は出力なしで成功する場合があります。`status` を信頼できる基準として使用してください。

## 5. ロールバック

### 5.1 対象バージョンまでロールバック（対象バージョン自体は維持）

```bash
migrate down 202604140001
```

意味: `202604140001` より大きい適用済みバージョンのみがロールバックされます。

### 5.2 適用済みバージョンをすべてロールバック

```bash
migrate down --all
```

注意: `migrate down <to-version>` と `migrate down --all` は互いに排他です。

### 5.3 ドライランでロールバック

```bash
migrate down 202604140001 --dry-run
migrate down --all --dry-run
```

`down` は出力なしで成功する場合があります。確認のため `migrate status` を実行してください。

## 6. ステータス確認

```bash
migrate status
```

機械可読な出力（スクリプトおよびAIエージェントに推奨）:

```bash
migrate status --output json
```

出力列: `Version`, `Filename`, `Hash`, `Status`。

よくあるステータス:

- `pending`
- `applied`
- `baseline`
- `outOfOrder`
- `hashMismatch`
- `filenameMismatch`

## 7. ツール自体のアップグレード/ダウングレード

### 7.1 ツールのバージョンをアップグレード

```bash
go install github.com/exc-works/migrate/cmd/migrate@latest
```

### 7.2 ツールのバージョンをダウングレード

```bash
go install github.com/exc-works/migrate/cmd/migrate@vX.Y.Z
```

例:

```bash
go install github.com/exc-works/migrate/cmd/migrate@v0.2.3
```

リポジトリがプライベートで `go install github.com/...@...` が使えない場合は、ソースコードで対象バージョンをチェックアウトして
次を実行します:

```bash
go install ./cmd/migrate
```

### 7.3 現在のツールバージョンを確認

```bash
migrate version
```

注: リリースアーティファクトはリリースバージョンを表示します。`go install ./cmd/migrate` によるローカルソースビルドは通常 `dev` を表示します。

## 8. 環境変数テンプレート

`data_source_name` では次をサポートします:

- `${KEY}`: 必須。存在している必要があります
- `${KEY:default}`: `default` は `KEY` が未設定の場合に使用

例:

```json
{
  "dialect": "postgres",
  "data_source_name": "host=127.0.0.1 port=5432 user=${DB_USER:postgres} password=${DB_PASSWORD} dbname=${DB_NAME:app} sslmode=disable"
}
```

`DB_PASSWORD` が環境変数に設定済みであることを確認してから、次を実行します:

```bash
migrate status
```

## 9. 10分でできる初回デモ（SQLite）

### 9.1 ディレクトリと設定を準備

まずコマンドが使えることを確認します:

```bash
migrate --help
```

デモ用ディレクトリを作成（macOS/Linux）:

```bash
mkdir -p ./migrate-demo
cd ./migrate-demo
migrate new config
```

Windows PowerShell の場合:

```powershell
mkdir .\migrate-demo
cd .\migrate-demo
migrate new config
```

`migration_config.json` を次のように更新します:

```json
{
  "schema_name": "migration_schema",
  "dialect": "sqlite",
  "data_source_name": "./demo.sqlite",
  "working_directory": ".",
  "migrate_out_of_order": false,
  "logger_level": "info",
  "migration_source": "migrations"
}
```

### 9.2 初期化してマイグレーションファイルを作成

```bash
migrate create
migrate new version init_users -v 202604140001
migrate new version add_email -v 202604140002
```

`migrations/V202604140001__init_users.sql` を編集:

```sql
-- +migrate Up
CREATE TABLE users
(
    id   INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);

-- +migrate Down
DROP TABLE IF EXISTS users;
```

`migrations/V202604140002__add_email.sql` を編集:

```sql
-- +migrate Up
ALTER TABLE users
    ADD COLUMN email TEXT;

-- +migrate Down
CREATE TABLE users_tmp
(
    id   INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);
INSERT INTO users_tmp (id, name)
SELECT id, name
FROM users;
DROP TABLE users;
ALTER TABLE users_tmp
    RENAME TO users;
```

### 9.3 適用、ステータス確認、ロールバック

```bash
migrate up --dry-run
migrate up
migrate status
migrate down 202604140001 --dry-run
migrate down 202604140001
migrate status
migrate down --all
migrate status
```

期待結果:

- `up` 後: 両バージョンが `applied`
- `down 202604140001` 後: `202604140001=applied`, `202604140002=pending`
- `down --all` 後: 両バージョンが `pending`

## 10. グローバルフラグ

特定の設定ファイルを使用:

```bash
migrate -c ./configs/dev.json status
```

特定の作業ディレクトリを使用:

```bash
migrate -w ./deploy create
migrate -w ./deploy up
```

## 11. よくあるエラーとトラブルシューティング

### 11.1 設定ファイルが見つからない

Error: `config file ... no such file or directory`

対処:

- 現在のディレクトリに `migration_config.json` が存在することを確認
- または `-c` で設定パスを渡す

### 11.2 環境変数が未設定

Error: `can't find env value for XXX`

対処:

- `export XXX=...`
- または `${XXX:default}` を使用

### 11.3 `down` 引数が不完全

Error: `to-version must be set, or use --all`

対処:

- `migrate down <version>` を使用
- または `migrate down --all` を使用

### 11.4 未サポートの dialect

Error: `unsupported dialect: xxx`

対処: 次のいずれかを使用:

- `postgres` `mysql` `mariadb` `oracle` `sqlite` `mssql` `clickhouse` `tidb` `redshift`

### 11.5 マイグレーションメタデータ不一致

Error: `hash mismatch` or `filename mismatch`

対処:

- 既に適用済みのマイグレーションファイルを編集しない
- 変更にはより高いバージョンの新しいマイグレーションを作成
