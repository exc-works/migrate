# SQLite クイックスタート

データベースサーバーを用意せずに、ひとつの SQL マイグレーションを作成、実行、ロールバックします。

## インストール

```sh
go install github.com/exc-works/migrate/cmd/migrate@latest
migrate --help
migrate version
```

## 設定

```sh
mkdir migrate-demo
cd migrate-demo
migrate new config
```

生成された `migration_config.yml` を次の内容に置き換えます。

```yaml
# 履歴テーブル名
schema_name: migration_schema
dialect: sqlite
data_source_name: ./demo.sqlite # ローカルの SQLite ファイル
working_directory: .
migrate_out_of_order: false
logger_level: info
migration_source: migrations # マイグレーションのディレクトリ
```

## マイグレーションの作成

```sh
migrate create
migrate new version "init users" -v 202611030001
```

`migrations/V202611030001__init_users.sql` を次の内容に置き換えます。

```sql
-- +migrate Up
CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);

-- +migrate Down
DROP TABLE users;
```

## 実行と確認

```sh
migrate up --dry-run
migrate up
migrate status
```

バージョン `202611030001` が `applied` になったことを確認します。

## ロールバック

```sh
migrate down --all --dry-run
migrate down --all
migrate status
```

テーブルが削除され、状態が `pending` に戻ります。[SQL](./sql-migrations)、[設定](./configuration)、[スクリプト](./scripting)へ進んでください。
