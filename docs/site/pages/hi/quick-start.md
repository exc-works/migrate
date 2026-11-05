# SQLite के साथ त्वरित शुरुआत

अलग सर्वर के बिना SQL माइग्रेशन बनाएँ, चलाएँ और वापस लें।

## इंस्टॉल करें

```sh
go install github.com/exc-works/migrate/cmd/migrate@latest
migrate --help
migrate version
```

## सेटअप

```sh
mkdir migrate-demo
cd migrate-demo
migrate new config
```

`migration_config.yml` को इससे बदलें:

```yaml
# माइग्रेशन इतिहास की तालिका
schema_name: migration_schema
dialect: sqlite
data_source_name: ./demo.sqlite # स्थानीय SQLite फ़ाइल
working_directory: .
migrate_out_of_order: false
logger_level: info
migration_source: migrations # माइग्रेशन डायरेक्टरी
```

## माइग्रेशन बनाएँ

```sh
migrate create
migrate new version "init users" -v 202611030001
```

`migrations/V202611030001__init_users.sql` को इससे बदलें:

```sql
-- +migrate Up
CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);

-- +migrate Down
DROP TABLE users;
```

## चलाएँ और जाँचें

```sh
migrate up --dry-run
migrate up
migrate status
```

संस्करण `202611030001` का स्टेटस `applied` होना चाहिए।

## वापस लें

```sh
migrate down --all --dry-run
migrate down --all
migrate status
```

तालिका हटती है और स्टेटस `pending` होता है। आगे [SQL](./sql-migrations), [कॉन्फ़िगरेशन](./configuration), [स्क्रिप्ट](./scripting) देखें।
