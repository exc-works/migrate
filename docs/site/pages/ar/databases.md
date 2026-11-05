# قواعد البيانات المدعومة

يدعم SQL ‏PostgreSQL وMySQL وMariaDB وSQLite وOracle وMSSQL وClickHouse وTiDB وRedshift. الاختيار عبر `dialect`.

## السكربتات والمعاملات {#script-support}

| القاعدة | الحالة | المعامل | المعاملة |
| --- | --- | --- | --- |
| PostgreSQL | متحقق | `$1` | يديرها migrate |
| MySQL / MariaDB | متحقق | `?` | جداول تدعم المعاملات؛ DDL قد يلتزم ضمنيًا |
| SQLite | متحقق | `?` | يديرها migrate |
| TiDB | تجريبي | `?` | TiKV الموزع غير متحقق؛ التزام DDL ممكن |
| Redshift | تجريبي، دون مثيل حقيقي متحقق | `$1` | حسب العبارة، لا يطابق PostgreSQL كاملًا |
| Oracle | تجريبي، تغطية Oracle Free محدودة | `:1` | DML بمعاملة، DDL يلتزم ضمنيًا |
| ClickHouse | تجريبي | `?` | دون معاملات أو تراجع تلقائي |
| MSSQL | تجريبي، Azure SQL غير متحقق | `@p1` | يديرها migrate |

تظهر تحذيرات التجريبي حتى في المعاينة. تحقق في بيئتك وراجع [المعاملات](./script-transactions).

## الأسماء والبدائل {#aliases}

لا فرق بين الأحرف الكبيرة والصغيرة وتُحذف المسافات الطرفية. يعيد `db.dialect` الاسم القياسي. MySQL وMariaDB وTiDB أنواع مستقلة.

| الاسم | البدائل |
| --- | --- |
| `postgres` | `pg`, `pgsql`, `postgresql`, `pgx` |
| `mssql` | `sqlserver`, `azuresql` |
| `oracle` | `orcl`, `ora`, `go-ora`, `goora`, `godror` |
| `sqlite` | `sqlite3` |
| `mysql`, `mariadb`, `tidb`, `redshift`, `clickhouse` | لا بدائل إضافية |

## قيم خاصة {#value-conversion}

تبقى الأعداد الصحيحة والعشرية الدقيقة نصوصًا.

- MSSQL: ‏MONEY دقيق وUNIQUEIDENTIFIER ثنائي؛ للنص `CAST(id AS VARCHAR(36))`. ‏rowversion/timestamp ثنائي.
- Oracle: ‏NUMBER نص وRAW بايتات وCLOB نص. BLOB دون بيانات وصفية يحتاج `DBMS_LOB.SUBSTR` محدودًا أو `RAWTOHEX`. الاتصال الخاص يستخدم `FAST+LOGIN=false` لإعادة اتصال موثوقة؛ CLI يضبطه ما لم يُحدد صراحة.
- Redshift: حوّل الأنواع غير المعروفة لـpgx. ‏VARBYTE يكتب بـ`FROM_HEX($1)` ويقرأ بـ`FROM_VARBYTE(payload, 'hex')`.
- ClickHouse: أعداد كبيرة وعشرية كنصوص. حوّل Array/Tuple/Map/UUID/IP إلى قيم بسيطة. String يتطلب UTF-8، والثنائي العشوائي hex. انتظار mutation بـ`SETTINGS mutations_sync=2`. قد يكون `rowsAffected` هو `"0"`.

```sh
migrate capabilities
```

يعرض الدعم في النسخة المثبتة دون اتصال؛ استخدم `--output json` للأتمتة.
