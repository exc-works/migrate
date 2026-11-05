# المعاملات والحدود

حيث يتوفر الدعم، يشترك السكربت وسجل النجاح بمعاملة: commit للنجاح وrollback للفشل. لا تنفذ BEGIN/COMMIT/ROLLBACK يدويًا. قد تبقى تغييرات DDL ذات الالتزام الضمني أو قواعد دون معاملات.

```js
export const config = { transaction: false };

export function up(db) {
  // PostgreSQL: يجب تنفيذ هذه العبارة خارج المعاملة.
  db.exec("CREATE INDEX CONCURRENTLY users_name_idx ON users (name)");
}

export function down(db) {
  db.exec("DROP INDEX CONCURRENTLY users_name_idx");
}
```

`transaction` منطقي، افتراضيًا `true` ويؤثر على up/down. لا يقيّم status/baseline/dry-run الإعدادات المصدرة عند التنفيذ. تشترك عمليات الترحيل باتصال مخصص. أنهِ معاملاتك وأعد إعدادات الجلسة قبل العودة، ولا تعتمد على استمرارها بين الترحيلات. الفشل غير المعاملاتي قد يترك تغييرات دون سجل نجاح.

## حدود الموارد

```yaml
script_timeout: 5m # مهلة كل سكربت
script_max_rows: 1000 # أقصى عدد صفوف لكل استعلام
script_max_bytes: 16777216 # أقصى حجم لكل استعلام بالبايت (16 MiB)
```

التجاوز يفشل دون اقتطاع. قسّم القراءات الكبيرة إلى صفحات؛ لا ينتج عن ذلك commit جزئي أو استئناف تلقائي. قد يتأخر توقف DB بعد طلب الإلغاء. Go يستخدم `Config.ScriptOptions`. انظر [دعم القواعد](./databases#script-support).
