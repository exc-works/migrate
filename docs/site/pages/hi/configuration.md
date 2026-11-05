# कॉन्फ़िगरेशन

प्रोजेक्ट डायरेक्टरी में `migration_config.yml` बनाएँ और संपादित करें:

```sh
migrate new config
```

## डेटाबेस कनेक्शन

`DATABASE_DSN` वेरिएबल से PostgreSQL कनेक्शन:

```yaml
# डेटाबेस प्रकार
dialect: postgres
# चलाने से पहले DATABASE_DSN सेट करें
data_source_name: "${DATABASE_DSN}"
# माइग्रेशन फ़ाइलों की डायरेक्टरी
migration_source: migrations
# लागू संस्करणों की तालिका
schema_name: migration_schema
```

| फ़ील्ड | उपयोग | डिफ़ॉल्ट |
| --- | --- | --- |
| `dialect` | डेटाबेस प्रकार | `postgres` |
| `data_source_name` | एक कनेक्शन; `data_sources` के साथ नहीं | एकल मोड में आवश्यक |
| `data_sources` | नाम वाले कनेक्शन | निष्क्रिय |
| `migration_source` | फ़ाइल डायरेक्टरी | `migrations` |
| `schema_name` | इतिहास तालिका | `migration_schema` |
| `working_directory` | फ़ाइल खोज का आधार | वर्तमान डायरेक्टरी |
| `migrate_out_of_order` | पुराने लंबित संस्करण चलाएँ | `false` |
| `logger_level` | `debug`, `info`, `warn`, `error` | `info` |

### कनेक्शन उदाहरण

होस्ट, पोर्ट, डेटाबेस या Oracle सर्विस बदलें और `DB_USER`, `DB_PASSWORD` सेट करें। SQLite में फ़ाइल पथ दें।

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

URL में उपयोगकर्ता/पासवर्ड के `@`, `:`, `/` जैसे अक्षर URL-encode करें। वेरिएबल प्रतिस्थापन यह अपने-आप नहीं करता। MySQL/MariaDB/TiDB का `user:password@tcp(...)` URL encoding नहीं उपयोग करता। TLS/SSL सर्वर की आवश्यकता के अनुसार रखें।

### लॉग स्तर {#logging}

| मान | संदेश |
| --- | --- |
| `debug` | डिबग, सूचना, चेतावनी, त्रुटि |
| `info` | सूचना, चेतावनी, त्रुटि |
| `warn` | चेतावनी और त्रुटि; `warning` भी मान्य |
| `error` | केवल त्रुटि |

```yaml
logger_level: debug # समस्या जाँच में डिबग संदेश दिखाएँ
```

## एनवायरनमेंट वेरिएबल {#environment-variables}

`${DATABASE_DSN}` का सेट होना आवश्यक है। `${DB_NAME:app}` अनुपस्थित होने पर `app` लेता है। एक YAML कई परिवेशों में उपयोग करें और क्रेडेंशियल वेरिएबल या सीक्रेट स्टोर से दें। SSH होस्ट/उपयोगकर्ता और प्रॉक्सी पता भी टेम्पलेट स्वीकार करते हैं।

## फ़ाइल खोज क्रम {#configuration-file-search-order}

कार्य डायरेक्टरी में पहली उपलब्ध फ़ाइल उपयोग होती है:

1. `migration_config.yml`
2. `migration_config.yaml`
3. `migration_config.json`

फ़ाइलें मिलाई नहीं जातीं। मिली फ़ाइल गलत या अपठनीय हो तो त्रुटि होगी। `-c` केवल चुनी फ़ाइल पढ़ता है, उसके न होने पर दूसरा विकल्प नहीं खोजता।

```sh
migrate -c ./configs/dev.yml status
```

## परिस्थिति के अनुसार बनाएँ

| परिस्थिति | कमांड |
| --- | --- |
| एक डेटाबेस | `migrate new config` |
| अपना नाम | `migrate new config dev.yml` |
| कई डेटाबेस | `migrate new config --multi` |
| चार डेटाबेस | `migrate new config --multi --shards 4` |
| SSH Agent | `migrate new config --ssh-tunnel` |
| निजी कुंजी | `migrate new config --ssh-tunnel --ssh-auth private_key` |
| SSH पासवर्ड | `migrate new config --ssh-tunnel --ssh-auth password` |
| SOCKS5 | `migrate new config --ssh-tunnel --proxy socks5` |
| HTTP CONNECT | `migrate new config --ssh-tunnel --proxy http` |

विकल्प जोड़ सकते हैं। ओवरराइट के लिए `--force`, सभी विकल्पों के लिए `migrate new config --help` देखें।

```sh
migrate new config shards.yml --multi --shards 4 --ssh-tunnel --proxy socks5
```

## कई डेटाबेस {#multiple-databases}

`data_source_name` की जगह `data_sources` दें। प्रकार, फ़ाइलें और इतिहास तालिका नाम साझा हैं। बनाए नाम `shard_0` और वेरिएबल `SHARD_0_DSN` से शुरू होते हैं।

```yaml
dialect: postgres
migration_source: migrations
schema_name: migration_schema
data_sources:
  shard_0:
    data_source_name: "${SHARD_0_DSN}" # पहले डेटाबेस का कनेक्शन
  shard_1:
    data_source_name: "${SHARD_1_DSN}" # दूसरे डेटाबेस का कनेक्शन
```

![माइग्रेशन फ़ाइलें → shard_0 / shard_1 → परिणाम](../assets/diagrams/multiple-databases-hi.svg)

```sh
migrate create
migrate up --dry-run
migrate up
migrate status
```

हर डेटाबेस अपने इतिहास से लंबित संस्करण चुनकर क्रम में चलाता है। सफल संस्करण दोबारा नहीं चलते। `new version` एक साझा फ़ाइल बनाता है। डिफ़ॉल्ट समानांतर सीमा 8 है; `--parallel 1` क्रमिक है।

```sh
migrate up --parallel 4
migrate status --output json
```

एक विफलता दूसरों को नहीं रोकती। नाम के अनुसार परिणाम मिलते हैं; कोई विफलता हो तो निकास कोड गैर-शून्य है। यह डेटाबेसों के बीच ट्रांज़ैक्शन नहीं है: दूसरे सफल बदलाव वापस नहीं होते। Ctrl+C नए काम रोकता और सक्रिय काम रद्द करता है।

ऊपरी `ssh_tunnel` साझा है। स्थानीय SSH ऑब्जेक्ट उसे पूरा बदलता है; अनुपस्थित या `null` हो तो साझा मान मिलता है। [SSH](./ssh-tunnels) देखें। हर असली DB एक बार सेट करें, समान DB पर टकराते कमांड न चलाएँ, रन में फ़ाइल/निर्भरता न बदलें। विफलता के बाद डेटा और इतिहास जाँचकर दोहराएँ।

## कार्य डायरेक्टरी

```sh
migrate -w ./deploy up
migrate -w ./deploy -c configs/dev.yml status
```

`-w` खोज निर्धारित कर `working_directory` बदलता है। सापेक्ष `migration_source` वहीं से है; केवल `-c` आधार नहीं बदलता। सापेक्ष SSH पथ कॉन्फ़िगरेशन फ़ाइल से शुरू होते हैं।
