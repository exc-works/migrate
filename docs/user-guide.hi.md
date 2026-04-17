# migrate यूज़र गाइड

यह गाइड पहली बार उपयोग करने वाले यूज़र्स के लिए है। कमांड और फ्लैग वर्तमान इम्प्लीमेंटेशन (`cmd/migrate`) पर आधारित हैं।

## 1. इंस्टॉलेशन

### 1.1 मॉड्यूल से इंस्टॉल करें

```bash
go install github.com/exc-works/migrate/cmd/migrate@latest
```

किसी विशेष संस्करण को इंस्टॉल करें:

```bash
go install github.com/exc-works/migrate/cmd/migrate@vX.Y.Z
```

`vX.Y.Z` को वास्तविक संस्करण से बदलें, उदाहरण के लिए `v0.2.3`।

### 1.2 लोकल सोर्स से इंस्टॉल करें (private repo या internal network)

रिपॉज़िटरी रूट में चलाएँ:

```bash
go install ./cmd/migrate
```

### 1.3 इंस्टॉलेशन सत्यापित करें

```bash
migrate --help
```

अगर कमांड नहीं मिलती है:

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

अगर आपको `Repository not found` दिखे, तो ऊपर दिया गया लोकल सोर्स इंस्टॉल पाथ इस्तेमाल करें।

## 2. इनिशियलाइज़ेशन

### 2.1 कॉन्फ़िग फ़ाइल बनाएँ

```bash
migrate new config
```

वैकल्पिक:

```bash
migrate new config dev.json
migrate new config --force
```

डिफ़ॉल्ट कॉन्फ़िग टेम्पलेट:

```json
{
  "schema_name": "migration_schema",
  "dialect": "postgres",
  "data_source_name": "host=127.0.0.1 port=5432 user=${DB_USER:postgres} password=${DB_PASSWORD:change_me} dbname=${DB_NAME:postgres} sslmode=disable",
  "working_directory": "",
  "migrate_out_of_order": false,
  "logger_level": "info",
  "migration_source": "migrations"
}
```

### 2.2 मुख्य कॉन्फ़िग फ़ील्ड अपडेट करें

- `dialect`: `postgres`, `mysql`, `mariadb`, `oracle`, `sqlite`, `mssql`, `clickhouse`, `tidb`, `redshift`
- `data_source_name`: DB कनेक्शन स्ट्रिंग
- `migration_source`: माइग्रेशन डायरेक्टरी (डिफ़ॉल्ट: `migrations`)

### 2.3 माइग्रेशन हिस्ट्री टेबल इनिशियलाइज़ करें

```bash
migrate create
```

`create` बिना आउटपुट के सफल हो सकता है। इसकी पुष्टि करें:

```bash
migrate status
```

अगर आपके पास मौजूदा स्कीमा है और आप पुरानी SQL को दोबारा चलाना नहीं चाहते, तो इस्तेमाल करें:

```bash
migrate baseline
```

## 3. माइग्रेशन वर्शन फ़ाइलें बनाएँ

### 3.1 स्वतः-जनरेटेड वर्शन

```bash
migrate new version init_users
```

### 3.2 स्पष्ट वर्शन

```bash
migrate new version add_email -v 202604140002
```

जनरेटेड फ़ाइलनाम फ़ॉर्मेट:

```text
V<version>__<description>.sql
```

डिफ़ॉल्ट फ़ाइल टेम्पलेट:

```sql
-- +migrate Up

-- +migrate Down
```

## 4. अपग्रेड (migrations लागू करें)

पहले dry run करें:

```bash
migrate up --dry-run
```

वास्तव में लागू करें:

```bash
migrate up
```

फिर status जाँचें:

```bash
migrate status
```

`up` बिना आउटपुट के सफल हो सकता है। सत्य के स्रोत के रूप में `status` का उपयोग करें।

## 5. रोलबैक

### 5.1 लक्ष्य वर्शन तक रोलबैक करें (लक्ष्य वर्शन बना रहता है)

```bash
migrate down 202604140001
```

अर्थ: केवल `202604140001` से बड़े लागू किए गए वर्शन रोलबैक किए जाते हैं।

### 5.2 सभी लागू किए गए वर्शन रोलबैक करें

```bash
migrate down --all
```

नोट: `migrate down <to-version>` और `migrate down --all` परस्पर अनन्य हैं।

### 5.3 Dry-run रोलबैक

```bash
migrate down 202604140001 --dry-run
migrate down --all --dry-run
```

`down` बिना आउटपुट के सफल हो सकता है। सत्यापन के लिए `migrate status` चलाएँ।

## 6. स्थिति जाँचें

```bash
migrate status
```

मशीन-पठनीय आउटपुट (स्क्रिप्ट्स और AI एजेंट्स के लिए अनुशंसित):

```bash
migrate status --output json
```

आउटपुट कॉलम: `Version`, `Filename`, `Hash`, `Status`।

सामान्य statuses:

- `pending`
- `applied`
- `baseline`
- `outOfOrder`
- `hashMismatch`
- `filenameMismatch`

## 7. टूल को स्वयं अपग्रेड या डाउनग्रेड करें

### 7.1 टूल वर्शन अपग्रेड करें

```bash
go install github.com/exc-works/migrate/cmd/migrate@latest
```

### 7.2 टूल वर्शन डाउनग्रेड करें

```bash
go install github.com/exc-works/migrate/cmd/migrate@vX.Y.Z
```

उदाहरण:

```bash
go install github.com/exc-works/migrate/cmd/migrate@v0.2.3
```

अगर repo private है और `go install github.com/...@...` उपलब्ध नहीं है, तो सोर्स कोड में लक्ष्य वर्शन checkout करें
और चलाएँ:

```bash
go install ./cmd/migrate
```

### 7.3 वर्तमान टूल वर्शन जाँचें

```bash
migrate version
```

नोट: release artifacts release वर्शन प्रिंट करते हैं; `go install ./cmd/migrate` से बने local source builds आमतौर पर `dev` प्रिंट करते हैं।

## 8. Environment variable templates

`data_source_name` में समर्थन है:

- `${KEY}`: आवश्यक, मौजूद होना चाहिए
- `${KEY:default}`: अगर `KEY` नहीं है, तो `default` उपयोग करें

उदाहरण:

```json
{
  "dialect": "postgres",
  "data_source_name": "host=127.0.0.1 port=5432 user=${DB_USER:postgres} password=${DB_PASSWORD} dbname=${DB_NAME:app} sslmode=disable"
}
```

सुनिश्चित करें कि `DB_PASSWORD` आपके environment में पहले से सेट है, फिर चलाएँ:

```bash
migrate status
```

## 9. 10-मिनट का first-run डेमो (SQLite)

### 9.1 डायरेक्टरी और कॉन्फ़िग तैयार करें

पहले कमांड उपलब्धता सत्यापित करें:

```bash
migrate --help
```

डेमो डायरेक्टरी बनाएँ (macOS/Linux):

```bash
mkdir -p ./migrate-demo
cd ./migrate-demo
migrate new config
```

Windows PowerShell समकक्ष:

```powershell
mkdir .\migrate-demo
cd .\migrate-demo
migrate new config
```

`migration_config.json` को इस प्रकार अपडेट करें:

```json
{
  "schema_name": "migration_schema",
  "dialect": "sqlite",
  "data_source_name": "./demo.sqlite",
  "working_directory": ".",
  "migrate_out_of_order": false,
  "logger_level": "info",
  "migration_source": "migrations"
}
```

### 9.2 इनिशियलाइज़ करें और माइग्रेशन फ़ाइलें बनाएँ

```bash
migrate create
migrate new version init_users -v 202604140001
migrate new version add_email -v 202604140002
```

`migrations/V202604140001__init_users.sql` संपादित करें:

```sql
-- +migrate Up
CREATE TABLE users
(
    id   INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);

-- +migrate Down
DROP TABLE IF EXISTS users;
```

`migrations/V202604140002__add_email.sql` संपादित करें:

```sql
-- +migrate Up
ALTER TABLE users
    ADD COLUMN email TEXT;

-- +migrate Down
CREATE TABLE users_tmp
(
    id   INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);
INSERT INTO users_tmp (id, name)
SELECT id, name
FROM users;
DROP TABLE users;
ALTER TABLE users_tmp
    RENAME TO users;
```

### 9.3 लागू करें, status जाँचें, और रोलबैक करें

```bash
migrate up --dry-run
migrate up
migrate status
migrate down 202604140001 --dry-run
migrate down 202604140001
migrate status
migrate down --all
migrate status
```

अपेक्षित:

- `up` के बाद: दोनों वर्शन `applied` हैं
- `down 202604140001` के बाद: `202604140001=applied`, `202604140002=pending`
- `down --all` के बाद: दोनों वर्शन `pending` हैं

## 10. ग्लोबल फ्लैग्स

विशिष्ट कॉन्फ़िग फ़ाइल का उपयोग करें:

```bash
migrate -c ./configs/dev.json status
```

विशिष्ट working directory का उपयोग करें:

```bash
migrate -w ./deploy create
migrate -w ./deploy up
```

## 11. सामान्य त्रुटियाँ और समस्या-निवारण

### 11.1 कॉन्फ़िग फ़ाइल नहीं मिली

त्रुटि: `config file ... no such file or directory`

समाधान:

- सुनिश्चित करें कि `migration_config.json` वर्तमान डायरेक्टरी में मौजूद है
- या `-c` के साथ कॉन्फ़िग पाथ पास करें

### 11.2 Environment variable गायब है

त्रुटि: `can't find env value for XXX`

समाधान:

- `export XXX=...`
- या `${XXX:default}` का उपयोग करें

### 11.3 अधूरे `down` arguments

त्रुटि: `to-version must be set, or use --all`

समाधान:

- `migrate down <version>` का उपयोग करें
- या `migrate down --all` का उपयोग करें

### 11.4 असमर्थित dialect

त्रुटि: `unsupported dialect: xxx`

समाधान: इनमें से किसी एक का उपयोग करें:

- `postgres` `mysql` `mariadb` `oracle` `sqlite` `mssql` `clickhouse` `tidb` `redshift`

### 11.5 माइग्रेशन मेटाडेटा mismatch

त्रुटि: `hash mismatch` या `filename mismatch`

समाधान:

- पहले से लागू माइग्रेशन फ़ाइलों को संपादित न करें
- बदलावों के लिए नया, बड़ा वर्शन माइग्रेशन बनाएँ
