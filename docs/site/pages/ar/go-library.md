# الاستخدام من Go

واجهة الحزمة الجذرية تتبع SemVer؛ `internal/*` ليست عامة. يدير المستدعي إغلاق المجمع وcontext ويجعل ترحيلات القاعدة الواحدة متسلسلة.

تعرض `Config.DryRun` و`DryRunOutput` الخطة، لكن `Create()` ينشئ السجل دائمًا. المسجل الافتراضي `NoopLogger`. استخدم `FSSource` وتبعياته [لتضمين السكربتات](./modules).

مع `OpenConnection` استخدم `Connection.Context()` في Ping والخدمة، ثم `Close()`. السائقون الشبكيون مسجلون، أما SQLite فعليك تسجيله. `ConnectionOptions.Prompt` و`HostKeyPrompt` اختياريان ويجب أن يحترما الإلغاء. دون callbacks لا طلب إدخال ولا ثقة بمضيف غير معروف.

بالإضافة إلى واجهة CLI، يمكن استيراد `github.com/exc-works/migrate` مباشرةً من كود خدمتك لتشغيل عمليات الترحيل — وهو مفيد لاختبارات الوحدة، وخطوات بدء التشغيل، أو لوحات الإدارة.

## التثبيت

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

## مثال مختصر

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

## واجهات API الرئيسية

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

[Go API](https://github.com/exc-works/migrate/blob/main/migrate.go) · [Go example](https://github.com/exc-works/migrate/blob/main/example_test.go)
