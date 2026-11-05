# كتابة ترحيلات SQL

تغيير منطقي واحد لكل ملف. SQL والسكربتات تشترك في تسلسل الإصدارات.

```sh
migrate new version "init users"
migrate new version "add email" -v 202611040002
```

ضع الوصف ذي المسافات بين علامتي اقتباس؛ تتحول المسافات إلى `_` في الاسم. يحدد `-v` الإصدار، وإلا يُولّد. النمط `V<version>__<description>.sql`. لا تعدّل ملفًا مطبقًا أو تعيد تسميته؛ أضف إصدارًا.

```sql
-- +migrate Up
CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);

-- +migrate Down
DROP TABLE users;
```

يطبق Up ويعكس Down. يحذف Down هنا البيانات أيضًا؛ خطط للاستعادة مسبقًا.

لحماية الفواصل المنقوطة الداخلية، استخدم `StatementBegin` و`StatementEnd`:

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

الصياغة ومعاملات DDL تعتمد على القاعدة. تابع [التطبيق](./applying).
