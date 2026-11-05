# SQL マイグレーションの作成

ひとつのファイルにひとつの論理的な変更を記述します。SQL とスクリプトは同じバージョン列を共有します。

```sh
migrate new version "init users"
migrate new version "add email" -v 202611040002
```

説明に空白がある場合は引用符を付けます。空白はファイル名では `_` になります。`-v` を省略するとバージョンを自動生成します。ファイル名は `V<version>__<description>.sql` です。適用済みファイルは編集や改名をせず、新しいバージョンを追加します。

## Up と Down

```sql
-- +migrate Up
CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);

-- +migrate Down
DROP TABLE users;
```

Up は変更、Down は逆操作です。この Down はテーブルとデータを削除します。失われるデータを復元できるかも確認してください。

## 文内のセミコロン

関数など、文内のセミコロンで分割してはいけない場合は、文全体を `StatementBegin` と `StatementEnd` で囲みます。

```sql
-- +migrate Up
-- +migrate StatementBegin
CREATE FUNCTION migration_answer() RETURNS integer AS $$
BEGIN
    RETURN 42;
END;
$$ LANGUAGE plpgsql;
-- +migrate StatementEnd

-- +migrate Down
DROP FUNCTION migration_answer();
```

SQL 文法と DDL のトランザクション動作はデータベースに依存します。続いて[実行](./applying)します。
