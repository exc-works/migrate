# واجهة قاعدة البيانات

يوفر `db` المتزامن `dialect` و`query(sql, params?)` و`exec(sql, params?)`. [تعريفات الأنواع](../assets/migration-api.d.ts).

```js
export function up(db) {
  const rows = db.query("SELECT id, name FROM users WHERE name=?", ["Alice"]);
  for (const row of rows) {
    const result = db.exec("UPDATE users SET name=? WHERE id=?", ["alice", row.id]);
    // result.rowsAffected سلسلة عشرية مثل "1".
  }
}
```

تحمل query مجموعة النتائج الأولى كاملة في الذاكرة؛ استخدم أسماء بديلة للأعمدة المتكررة. تعيد exec ‏`{ rowsAffected: string }` وتحفظ عدد السائق حتى لو `"0"`.

المعاملات مصفوفة: PostgreSQL/Redshift ‏`$1`، MySQL/MariaDB/SQLite/TiDB/ClickHouse ‏`?`، Oracle ‏`:1`، MSSQL ‏`@p1`. يعيد `db.dialect` الاسم القياسي.

الأعداد الصحيحة والعشرية الدقيقة نصوص، NULL يصبح `null`، الثنائي نسخ لمصفوفات بايت، والأوقات نصوص UTC دقيقة. number و`JSON.parse` قد يفقدان الدقة. تُرفض أعداد SQLite العشرية المخزنة float. راجع [الأنواع الخاصة](./databases#value-conversion).

يُرسل SQL دون تقسيم أو تفسير `GO`؛ احتفظ بالفواصل المطلوبة مثل MERGE. دفعات MySQL قد تحتاج خيار DSN. أخطاء السائق المنقحة قابلة للإمساك، لكن catch لا يصلح المعاملة؛ استخدم savepoint/rollback أو أعد رمي الخطأ. العودة الطبيعية تقبل النتيجة. الإلغاء والمهلة الكلية والقيم غير الصالحة وتجاوز الموارد تبقى قاتلة بعد catch وتمنع DB التالية. مهلة السائق وحدها قابلة للإمساك. راجع [المعاملات](./script-transactions).
