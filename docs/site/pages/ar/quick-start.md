# البدء السريع باستخدام SQLite

أنشئ ترحيل SQL وطبّقه ثم تراجع عنه دون خادم قاعدة بيانات منفصل.

## التثبيت

```sh
go install github.com/exc-works/migrate/cmd/migrate@latest
migrate --help
migrate version
```

## الإعداد

```sh
mkdir migrate-demo
cd migrate-demo
migrate new config
```

استبدل محتوى `migration_config.yml` بما يلي:

```yaml
# جدول سجل الترحيلات
schema_name: migration_schema
dialect: sqlite
data_source_name: ./demo.sqlite # ملف SQLite محلي
working_directory: .
migrate_out_of_order: false
logger_level: info
migration_source: migrations # مجلد الترحيلات
```

## إنشاء الترحيل

```sh
migrate create
migrate new version "init users" -v 202611030001
```

استبدل `migrations/V202611030001__init_users.sql` بما يلي:

```sql
-- +migrate Up
CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);

-- +migrate Down
DROP TABLE users;
```

## التطبيق والتحقق

```sh
migrate up --dry-run
migrate up
migrate status
```

يجب أن تكون حالة الإصدار `202611030001` هي `applied`.

## التراجع

```sh
migrate down --all --dry-run
migrate down --all
migrate status
```

يُحذف الجدول وتعود الحالة إلى `pending`. تابع [SQL](./sql-migrations) و[الإعدادات](./configuration) و[السكربتات](./scripting).
