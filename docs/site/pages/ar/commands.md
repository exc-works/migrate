# مرجع الأوامر

راجع `migrate version` و`migrate --help` و`migrate <command> --help`.

| الخيار العام | الوظيفة |
| --- | --- |
| `-c`, `--config` | ملف محدد، وإلا [بحث تلقائي](./configuration#configuration-file-search-order) |
| `-w`, `--working-dir` | استبدال مجلد العمل |
| `--non-interactive` | خطأ بدل طلب سر أو الثقة بمفتاح SSH جديد |

## الإنشاء

```sh
migrate new config [filename] [--format yaml|json] [--multi] [--shards N] [--ssh-tunnel] [--force]
migrate new config shards.yml --multi --shards 4
migrate new config shards-ssh.yml --multi --ssh-tunnel --proxy socks5
```

الافتراضي `migration_config.yml`. يتطلب `--shards N` الخيار `--multi` وعددًا موجبًا، والترقيم من 0. يتطلب `--proxy` الخيار `--ssh-tunnel`. المصادقة `--ssh-auth agent|private_key|password`. الكتابة فوق الملف بـ`--force`، وJSON عبر `.json` أو `--format json`.

```sh
migrate new version <description> [-v <version>] [--format sql|js|ts|mjs]
```

SQL افتراضيًا وإصدار تلقائي أو `-v`. الوصف `"init users"` بين اقتباس، والمسافات تصبح `_` بالاسم.

## عمليات القاعدة

```sh
migrate create
```

ينشئ جدول السجل فقط، ولا يكرر إنشاءه إن وجد.

```sh
migrate baseline
```

يسجل دون تنفيذ فقط عندما تتوافق القاعدة مع الملفات. راجع [baseline](./history#baseline).

```sh
migrate up [--dry-run]
```

يطبق المعلّق بالترتيب؛ dry-run قد يقرأ التاريخ دون تغييرات.

```sh
migrate down <to-version> [--dry-run]
migrate down --all [--dry-run]
```

يلغي ما فوق الهدف مع إبقائه، أو الجميع. اختر الهدف أو `--all` لا كليهما.

```sh
migrate status [--output table|json]
```

جدول افتراضيًا. JSON لعدة قواعد مرتب بالاسم ويحتوي `data_source` و`status`، ثم `migrations` للنجاح أو `error` للفشل. create/baseline/status/up/down تشمل جميع الاتصالات وتقبل `--parallel N` (8). الفشل لا يوقف الآخرين ويعيد خروجًا غير صفري.

## معلومات دون اتصال

```sh
migrate capabilities [--output text|json]
```

الصيغ والأسماء ودعم السكربتات والمعاملات؛ الافتراضي `text`.

```sh
migrate version
```

نسخة الأداة لا نسخة الترحيلات.
