# دليل مستخدم migrate

## 1. التثبيت

### 1.1 التثبيت من الوحدة

```bash
go install github.com/exc-works/migrate/cmd/migrate@latest
```

تثبيت إصدار محدد:

```bash
go install github.com/exc-works/migrate/cmd/migrate@vX.Y.Z
```

استبدل `vX.Y.Z` بإصدار فعلي، على سبيل المثال `v0.2.3`.

### 1.2 التثبيت من المصدر المحلي (مستودع خاص أو شبكة داخلية)

شغّل في جذر المستودع:

```bash
go install ./cmd/migrate
```

### 1.3 التحقق من التثبيت

```bash
migrate --help
```

إذا لم يتم العثور على الأمر:

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

إذا ظهر لك `Repository not found`، فاستخدم مسار التثبيت من المصدر المحلي المذكور أعلاه.

## 2. التهيئة

### 2.1 إنشاء ملف الإعداد

```bash
migrate new config
```

اختياري:

```bash
migrate new config dev.yml
migrate new config --force
```

قالب الإعداد الافتراضي:

```yaml
schema_name: migration_schema
dialect: postgres
data_source_name: "host=127.0.0.1 port=5432 user=${DB_USER:postgres} password=${DB_PASSWORD} dbname=${DB_NAME:postgres} sslmode=disable"
working_directory: ""
migrate_out_of_order: false
logger_level: info
migration_source: migrations
```

### 2.2 تحديث حقول الإعداد الأساسية

- `dialect`: `postgres`, `mysql`, `mariadb`, `oracle`, `sqlite`, `mssql`, `clickhouse`, `tidb`, `redshift`
- `data_source_name`: سلسلة اتصال قاعدة البيانات
- `migration_source`: دليل الترحيل (الافتراضي: `migrations`)

### 2.3 تهيئة جدول سجل الترحيل

```bash
migrate create
```

قد ينجح `create` بدون أي مخرجات. أكّد ذلك عبر:

```bash
migrate status
```

إذا كان لديك مخطط قاعدة بيانات موجود مسبقًا ولا تريد إعادة تنفيذ ملفات SQL القديمة، فاستخدم:

```bash
migrate baseline
```

## 3. إنشاء ملفات إصدارات الترحيل

### 3.1 إصدار يتم توليده تلقائيًا

```bash
migrate new version "init users"
```

### 3.2 إصدار صريح

```bash
migrate new version "add email" -v 202604140002
```

صيغة اسم الملف المُنشأ:

```text
V<version>__<description>.sql
```

قالب الملف الافتراضي:

```sql
-- +migrate Up

-- +migrate Down
```

## 4. الترقية (تطبيق ملفات الترحيل)

نفّذ تجربة تشغيل أولًا:

```bash
migrate up --dry-run
```

التطبيق الفعلي:

```bash
migrate up
```

ثم تحقق من الحالة:

```bash
migrate status
```

قد ينجح `up` بدون أي مخرجات. استخدم `status` كمصدر الحقيقة.

## 5. التراجع

### 5.1 التراجع إلى إصدار مستهدف (يتم الإبقاء على الإصدار المستهدف)

```bash
migrate down 202604140001
```

الدلالة: يتم التراجع فقط عن الإصدارات المطبقة الأكبر من `202604140001`.

### 5.2 التراجع عن جميع الإصدارات المطبقة

```bash
migrate down --all
```


### 5.3 تجربة تشغيل للتراجع

```bash
migrate down 202604140001 --dry-run
migrate down --all --dry-run
```

قد ينجح `down` بدون أي مخرجات. شغّل `migrate status` للتحقق.

## 6. التحقق من الحالة

```bash
migrate status
```

مخرجات قابلة للقراءة آليًا (موصى بها للسكربتات ووكلاء الذكاء الاصطناعي):

```bash
migrate status --output json
```

أعمدة المخرجات: `Version`, `Filename`, `Hash`, `Status`.

الحالات الشائعة:

- `pending`
- `applied`
- `baseline`
- `outOfOrder`
- `hashMismatch`
- `filenameMismatch`

## 7. التحقق من الإصدار

```bash
migrate version
```

## 8. قوالب متغيرات البيئة

يدعم `data_source_name`:

- `${KEY}`: مطلوب ويجب أن يكون موجودًا
- `${KEY:default}`: استخدم `default` إذا كان `KEY` مفقودًا

مثال:

```yaml
dialect: postgres
data_source_name: "host=127.0.0.1 port=5432 user=${DB_USER:postgres} password=${DB_PASSWORD} dbname=${DB_NAME:app} sslmode=disable"
```

تأكد من أن `DB_PASSWORD` مضبوط بالفعل في البيئة لديك، ثم شغّل:

```bash
migrate status
```

## 9. عرض تشغيل أول خلال 10 دقائق (SQLite)

[عرض تشغيل أول خلال 10 دقائق (SQLite) · English](site/pages/quick-start.md)

## 10. الأعلام العامة

استخدم ملف إعداد محدد:

```bash
migrate -c ./configs/dev.yml status
```

استخدم دليل عمل محدد:

```bash
migrate -w ./deploy create
migrate -w ./deploy up
```

## 11. الأخطاء الشائعة واستكشاف الأخطاء وإصلاحها

### 11.1 لم يتم العثور على ملف الإعداد

الخطأ: `config file ... no such file or directory`

الحل:

- تأكد من وجود `migration_config.yml` في الدليل الحالي
- أو مرّر مسار ملف الإعداد عبر `-c`

### 11.2 متغير بيئة مفقود

الخطأ: `can't find env value for XXX`

الحل:

- `export XXX=...`
- أو استخدم `${XXX:default}`

### 11.3 معاملات `down` غير مكتملة

الخطأ: `to-version must be set, or use --all`

الحل:

- استخدم `migrate down <version>`
- أو استخدم `migrate down --all`

### 11.4 dialect غير مدعوم

الخطأ: `unsupported dialect: xxx`

الحل: استخدم واحدة من:

- `postgres` `mysql` `mariadb` `oracle` `sqlite` `mssql` `clickhouse` `tidb` `redshift`

### 11.5 عدم تطابق بيانات الترحيل الوصفية

الخطأ: `hash mismatch` أو `filename mismatch`

الحل:

- استعد ملف الترحيل الأصلي من نظام التحكم في الإصدارات
- أنشئ ترحيلًا جديدًا بإصدار أعلى لإجراء التغييرات

## 12. استخدام migrate كمكتبة Go

بالإضافة إلى واجهة CLI، يمكن استيراد `github.com/exc-works/migrate` مباشرةً من كود خدمتك لتشغيل عمليات الترحيل — وهو مفيد لاختبارات الوحدة، وخطوات بدء التشغيل، أو لوحات الإدارة.

### 12.1 التثبيت

```bash
go get github.com/exc-works/migrate
```

استورد مشغّل قاعدة البيانات الذي تحتاجه (المكتبة لا تربط واحدًا بعينه):

```go
import (
    _ "modernc.org/sqlite"             // sqlite
    _ "github.com/jackc/pgx/v5/stdlib" // postgres
    _ "github.com/go-sql-driver/mysql" // mysql / mariadb / tidb
    // ...
)
```

### 12.2 مثال مختصر

```go
package main

import (
    "context"
    "database/sql"

    _ "modernc.org/sqlite"

    "github.com/exc-works/migrate"
)

func main() {
    db, err := sql.Open("sqlite", "./app.sqlite")
    if err != nil {
        panic(err)
    }
    defer db.Close()

    svc, err := migrate.NewService(context.Background(), migrate.Config{
        Dialect:         migrate.NewSQLiteDialect(),
        DB:              db,
        MigrationSource: migrate.DirectorySource{Directory: "./migrations"},
    })
    if err != nil {
        panic(err)
    }

    if err := svc.Create(); err != nil { // idempotent: ينشئ جدول السجل إن لم يكن موجودًا
        panic(err)
    }
    if err := svc.Up(); err != nil {
        panic(err)
    }
}
```

### 12.3 واجهات API الرئيسية

- `migrate.NewService(ctx, migrate.Config)` يبني مشغّل الترحيل
- `svc.Create()` ينشئ جدول السجل `migration_schema` (idempotent)
- `svc.Up()` يطبّق جميع عمليات الترحيل المعلّقة
- `svc.Down(toVersion, all)` يتراجع إلى الإصدار المحدد أو كل شيء
- `svc.Status()` يُعيد `[]migrate.MigrationStatus`
- `svc.Baseline()` يُعلّم الملفات المعلّقة الحالية بأنها `baseline`

الأنواع الشائعة:

- اللهجات (يُفضَّل استخدام الدوال البانية — فهي تُعيد واجهة `Dialect`): `migrate.NewPostgresDialect()`, `NewMySQLDialect()`, `NewSQLiteDialect()`, `NewMSSQLDialect()`, `NewOracleDialect()`, `NewClickHouseDialect()`, `NewMariaDBDialect()`, `NewTiDBDialect()`, `NewRedshiftDialect()`، أو `migrate.DialectFromName("postgres")` للبحث بالاسم
- المصادر: `DirectorySource` (نظام الملفات)، `StringSource` (شريحة في الذاكرة)، `FSSource` (أي `fs.FS`، مثل `//go:embed` أو `os.DirFS`)، `CombinedSource` (يجمع عدة مصادر)
- السجلات: `migrate.NoopLogger{}` (افتراضي)، `migrate.NewStdLogger("info", os.Stdout)`، أو أي نوع يُطبّق `migrate.Logger`

[Go API](../migrate.go) · [Go example](../example_test.go)
