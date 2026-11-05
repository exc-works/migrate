# كتابة السكربتات

حوّل البيانات باستخدام JS أو TS أو MJS دون Node.js وقت التنفيذ.

```sh
migrate new version "normalize names" --format ts
```

النمط `V<version>__<description>.<format>` يشارك إصدارات SQL، وحدود MJS مثل JS.

```ts
export function up(db) {
  for (const row of db.query("SELECT id, name FROM users")) {
    db.exec("UPDATE users SET name=? WHERE id=?", [row.name.trim(), row.id]);
  }
}
```

مثال SQLite على `users.name`. يرمي up المولد خطأ حتى تنفيذه. الدوال متزامنة: لا async ولا مولدات ولا Promise؛ تُهمل النتائج المتزامنة العادية. [معاملات SQL](./script-api) تختلف حسب القاعدة.

حذف المسافات يفقد المعلومات، لذلك لا يوجد down هنا. صدّر `down(db)` بعد تنفيذ العكس فقط، واحفظ الأصل إذا احتجت الاسترجاع.

```sh
migrate up --dry-run
migrate up
migrate status
```

dry-run لا يشغّل السكربت. ضع عمليات DB داخل up/down لا أعلى الوحدة. لا تعتمد على حالة عامة بين الترحيلات. الأنواع في القالب و[migration-api.d.ts](../assets/migration-api.d.ts)، لكن التحويل لا يفحص الأنواع. لا تتوفر API Node أو الشبكة أو الملفات أو البيئة أو المؤقتات أو التحميل الديناميكي؛ تحقق من حزم npm كذلك. انظر [الوحدات](./modules) و[المعاملات](./script-transactions).
