# التثبيت

استخدم Go أو نزّل الملف المناسب لنظامك ومعماريتك من [الإصدارات](https://github.com/exc-works/migrate/releases). تحقق من SHA-256 وضع `migrate` (أو `migrate.exe` في Windows) ضمن PATH.

## التثبيت من الوحدة

```bash
go install github.com/exc-works/migrate/cmd/migrate@latest
```

تثبيت إصدار محدد:

```bash
go install github.com/exc-works/migrate/cmd/migrate@vX.Y.Z
```

## التثبيت من المصدر المحلي (مستودع خاص أو شبكة داخلية)

شغّل في جذر المستودع:

```bash
go install ./cmd/migrate
```

## التحقق من التثبيت

```sh
migrate --help
migrate version
migrate capabilities --output json
```

إذا لم يتم العثور على الأمر:

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

تعرض البنيات المحلية `dev`. للنسخ القديمة المثبتة عبر Go استخدم `go version -m <binary>`. لا يتصل capabilities بقاعدة البيانات. تابع إلى [البدء السريع](./quick-start).
