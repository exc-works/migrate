# الوحدات المشتركة

تدعم استيرادًا وإعادة تصدير ثابتين لملفات `.js` و`.ts` و`.mjs`. الملفات المساعدة ليست ترحيلات مستقلة.

```text
project/
  migrations/V2__normalize.ts
  shared/names.ts
```

```ts
import { normalize } from "../shared/names.ts";

export function up(db) {
  db.exec("UPDATE users SET name=? WHERE id=?", [normalize(" Alice "), "1"]);
}
```

المسارات من الملف المستورد. عرّف `normalize` في `shared/names.ts` وأنشئ `users` أولًا. راجع [المثال الكامل](./examples#local).

ثبّت npm وفق lockfile مقدمًا. دعم exports والمسارات الفرعية والتبعيات المتداخلة وCommonJS الثابت بشروط `neutral` ثم `module` و`main`. لا API Node أو إضافات أصلية أو import/require ديناميكي. migrate لا يثبت الحزم.

## المصادر المضمّنة

يحل `FSSource` الملفات داخل `fs.FS`، مثل `embed.FS`، مع البحث عن الامتدادات و`index`، لا package.json أو node_modules افتراضية. تحتاج npm إلى مجلد فعلي:

```go
source := migrate.WithModules{
    Source: migrate.FSSource{
        FS:   embeddedFiles,
        Root: "migrations",
    },
    DependencyDir: projectDirectory,
}
```

تبقى الملفات النسبية افتراضية. دون `DependencyDir` لا يُبحث في المضيف. يقبل `StringSource` غلاف `WithModules` مع `FS`/`Root` أو `Directory`. يمكن للمصدر المخصص تنفيذ `MigrationEntriesSource`. انظر [Go](./go-library).

## قابلية الإعادة {#checksums}

البصمة للملف الرئيسي فقط. تعديل مساعد أو حزمة أو lockfile لا يسبب اختلافًا ولا يعيد إصدارًا مطبقًا. ثبّت التبعيات واحتفظ بالسلوك التاريخي. تُفحص تبعيات النطاق كله قبل الكتابة؛ خطأ لاحق يمنع الخطوات السابقة أيضًا.
