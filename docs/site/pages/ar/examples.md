# تشغيل الأمثلة

من نسخة المستودع، تستخدم الأمثلة SQLite وتتحقق من بياناتها داخليًا.

## CLI مؤقت

تابع في الصدفة نفسها.

```sh
example_workspace=$(mktemp -d)
go build -o "$example_workspace/migrate" ./cmd/migrate
```

## وحدة محلية {#local}

[المصادر](https://github.com/exc-works/migrate/tree/main/examples/script-imports/local): ترحيلان يتشاركان TypeScript.

```sh
export MIGRATE_EXAMPLE_DB="$example_workspace/local.sqlite"
"$example_workspace/migrate" -w examples/script-imports/local create
"$example_workspace/migrate" -w examples/script-imports/local up --dry-run
"$example_workspace/migrate" -w examples/script-imports/local up
"$example_workspace/migrate" -w examples/script-imports/local status
"$example_workspace/migrate" -w examples/script-imports/local down 2
"$example_workspace/migrate" -w examples/script-imports/local down --all
```

بعد up ثلاثة إصدارات والاسمان `alice` و`bob`. يحتفظ `down 2` بـ1 و2، ويحذف `down --all` الجدول.

## حزمة npm

Node/npm للتثبيت فقط. lockfile يشير إلى حزمة محلية دون سجل خارجي.

```sh
npm --prefix examples/script-imports/npm ci --ignore-scripts --no-audit --no-fund --offline
export MIGRATE_EXAMPLE_DB="$example_workspace/npm.sqlite"
"$example_workspace/migrate" -w examples/script-imports/npm create
"$example_workspace/migrate" -w examples/script-imports/npm up
"$example_workspace/migrate" -w examples/script-imports/npm status
"$example_workspace/migrate" -w examples/script-imports/npm down --all
```

يستورد الترحيل الثاني `@migrate-example/normalizer/names` ويتحقق من `alice`. يطبق إصداران.

## تضمين Go

```sh
go run ./examples/script-imports/embedded
```

يستخدم `embed.FS` وقاعدة ذاكرة خاصة دون npm. الخرج:

```text
after up: alice
after down: all migrations pending
```

## التنظيف

```sh
rm -r "$example_workspace"
```

يبقى node_modules المتجاهل. للتحقق بنسخ مؤقتة استخدم `./scripts/test-script-examples.sh` من الجذر. انظر [قابلية الإعادة](./modules#checksums).
