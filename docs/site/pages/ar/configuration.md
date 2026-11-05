# الإعدادات

أنشئ `migration_config.yml` وعدّله داخل مجلد المشروع:

```sh
migrate new config
```

## الاتصال بقاعدة البيانات

مثال PostgreSQL يقرأ الاتصال من `DATABASE_DSN`:

```yaml
# نوع قاعدة البيانات
dialect: postgres
# اضبط DATABASE_DSN قبل التشغيل
data_source_name: "${DATABASE_DSN}"
# مجلد ملفات الترحيل
migration_source: migrations
# جدول الإصدارات المطبقة
schema_name: migration_schema
```

| الحقل | الغرض | الافتراضي |
| --- | --- | --- |
| `dialect` | نوع قاعدة البيانات | `postgres` |
| `data_source_name` | اتصال واحد؛ لا يجتمع مع `data_sources` | مطلوب لقاعدة واحدة |
| `data_sources` | اتصالات مسماة | غير مفعّل |
| `migration_source` | مجلد الملفات | `migrations` |
| `schema_name` | جدول السجل | `migration_schema` |
| `working_directory` | أساس البحث عن الملفات | المجلد الحالي |
| `migrate_out_of_order` | تطبيق الإصدارات المعلقة الأقدم | `false` |
| `logger_level` | `debug` أو `info` أو `warn` أو `error` | `info` |

### أمثلة الاتصال

غيّر المضيف والمنفذ واسم القاعدة أو خدمة Oracle واضبط `DB_USER` و`DB_PASSWORD`. يستخدم SQLite مسار ملف.

::: code-group

```yaml [PostgreSQL]
dialect: postgres
data_source_name: "postgres://${DB_USER}:${DB_PASSWORD}@db.example:5432/app?sslmode=require"
```

```yaml [MySQL]
dialect: mysql
data_source_name: "${DB_USER}:${DB_PASSWORD}@tcp(db.example:3306)/app?parseTime=true"
```

```yaml [MariaDB]
dialect: mariadb
data_source_name: "${DB_USER}:${DB_PASSWORD}@tcp(db.example:3306)/app?parseTime=true"
```

```yaml [SQLite]
dialect: sqlite
data_source_name: "./demo.sqlite"
```

```yaml [Oracle]
dialect: oracle
data_source_name: "oracle://${DB_USER}:${DB_PASSWORD}@db.example:1521/FREEPDB1"
```

```yaml [SQL Server]
dialect: mssql
data_source_name: "sqlserver://${DB_USER}:${DB_PASSWORD}@db.example:1433?database=app"
```

```yaml [ClickHouse]
dialect: clickhouse
data_source_name: "clickhouse://${DB_USER}:${DB_PASSWORD}@db.example:9000/app"
```

```yaml [TiDB]
dialect: tidb
data_source_name: "${DB_USER}:${DB_PASSWORD}@tcp(db.example:4000)/app?parseTime=true"
```

```yaml [Redshift]
dialect: redshift
data_source_name: "postgres://${DB_USER}:${DB_PASSWORD}@cluster.example:5439/app?sslmode=require"
```

:::

في روابط URL، رمّز المحارف الخاصة في المستخدم وكلمة المرور مثل `@` و`:` و`/`. استبدال المتغيرات لا يرمّزها تلقائيًا. صيغة MySQL/MariaDB/TiDB ‏`user:password@tcp(...)` لا تستخدم ترميز URL. اضبط TLS/SSL حسب متطلبات الخادم.

### مستويات السجل {#logging}

| القيمة | الرسائل |
| --- | --- |
| `debug` | تنقيح ومعلومات وتحذيرات وأخطاء |
| `info` | معلومات وتحذيرات وأخطاء |
| `warn` | تحذيرات وأخطاء؛ يقبل `warning` أيضًا |
| `error` | أخطاء فقط |

```yaml
logger_level: debug # إظهار رسائل التنقيح
```

## متغيرات البيئة {#environment-variables}

يتطلب `${DATABASE_DSN}` متغيرًا مضبوطًا، بينما يستخدم `${DB_NAME:app}` القيمة `app` عند غيابه. شارك YAML بين البيئات وقدّم الأسرار عبر متغيرات أو مخزن أسرار. مضيف SSH والمستخدم وعنوان الوكيل يقبلون المتغيرات أيضًا.

## ترتيب البحث {#configuration-file-search-order}

يُستخدم أول ملف موجود في مجلد العمل:

1. `migration_config.yml`
2. `migration_config.yaml`
3. `migration_config.json`

لا تُدمج الملفات. الملف المختار غير الصالح أو غير المقروء يسبب خطأ. مع `-c` يُستخدم الملف المحدد فقط، حتى إن كان مفقودًا.

```sh
migrate -c ./configs/dev.yml status
```

## إنشاء إعدادات حسب الحالة

| الحالة | الأمر |
| --- | --- |
| قاعدة واحدة | `migrate new config` |
| اسم مخصص | `migrate new config dev.yml` |
| عدة قواعد | `migrate new config --multi` |
| أربع قواعد | `migrate new config --multi --shards 4` |
| SSH Agent | `migrate new config --ssh-tunnel` |
| مفتاح خاص | `migrate new config --ssh-tunnel --ssh-auth private_key` |
| كلمة مرور SSH | `migrate new config --ssh-tunnel --ssh-auth password` |
| SOCKS5 | `migrate new config --ssh-tunnel --proxy socks5` |
| HTTP CONNECT | `migrate new config --ssh-tunnel --proxy http` |

يمكن جمع الخيارات. استخدم `--force` للكتابة فوق ملف، و`migrate new config --help` لجميع الخيارات.

```sh
migrate new config shards.yml --multi --shards 4 --ssh-tunnel --proxy socks5
```

## عدة قواعد بيانات {#multiple-databases}

استخدم `data_sources` بدل `data_source_name`. النوع والملفات واسم جدول السجل مشتركة. تبدأ الأسماء المولدة من `shard_0` والمتغيرات من `SHARD_0_DSN`.

```yaml
dialect: postgres
migration_source: migrations
schema_name: migration_schema
data_sources:
  shard_0:
    data_source_name: "${SHARD_0_DSN}" # اتصال قاعدة البيانات الأولى
  shard_1:
    data_source_name: "${SHARD_1_DSN}" # اتصال قاعدة البيانات الثانية
```

![ملفات الترحيل → shard_0 / shard_1 → النتائج](../assets/diagrams/multiple-databases-ar.svg)

```sh
migrate create
migrate up --dry-run
migrate up
migrate status
```

تختار كل قاعدة الإصدارات المعلقة بحسب سجلها وتنفذها بالترتيب دون تكرار الناجح. ينشئ `new version` ملفًا مشتركًا واحدًا. التوازي الافتراضي 8؛ يستخدم `--parallel 1` التنفيذ المتسلسل.

```sh
migrate up --parallel 4
migrate status --output json
```

فشل قاعدة لا يوقف البقية. تُجمع النتائج بالاسم وأي فشل يعيد رمز خروج غير صفري. ليست معاملة بين القواعد؛ لا تُلغى نجاحات القواعد الأخرى. يوقف Ctrl+C المهام الجديدة ويلغي النشطة.

يمكن مشاركة `ssh_tunnel` في المستوى الأعلى. كائن محلي يستبدله كاملًا، أما غيابه أو `null` فيرثه. انظر [SSH](./ssh-tunnels). لا تسجل القاعدة الفعلية مرتين ولا تشغّل عليها أوامر متداخلة. ثبّت الملفات والتبعيات أثناء التشغيل وافحص البيانات والسجل قبل إعادة المحاولة.

## مجلد العمل

```sh
migrate -w ./deploy up
migrate -w ./deploy -c configs/dev.yml status
```

يحدد `-w` البحث ويستبدل `working_directory`. المسار النسبي `migration_source` يبدأ منه؛ `-c` وحده لا يغيره. مسارات ملفات SSH النسبية تبدأ من مجلد ملف الإعدادات.
