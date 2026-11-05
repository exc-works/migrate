# CLI संदर्भ

`migrate version`, `migrate --help`, `migrate <command> --help` देखें।

| वैश्विक विकल्प | काम |
| --- | --- |
| `-c`, `--config` | चुनी फ़ाइल, अन्यथा [स्वतः खोज](./configuration#configuration-file-search-order) |
| `-w`, `--working-dir` | कार्य डायरेक्टरी बदलें |
| `--non-interactive` | सीक्रेट या SSH विश्वास पूछने की जगह विफल |

## बनाएँ

```sh
migrate new config [filename] [--format yaml|json] [--multi] [--shards N] [--ssh-tunnel] [--force]
migrate new config shards.yml --multi --shards 4
migrate new config shards-ssh.yml --multi --ssh-tunnel --proxy socks5
```

डिफ़ॉल्ट migration_config.yml। --shards N के लिए --multi और धनात्मक संख्या, गिनती 0 से। --proxy के लिए --ssh-tunnel। प्रमाणीकरण --ssh-auth agent|private_key|password। --force ओवरराइट; .json या --format json JSON चुनता है।

```sh
migrate new version <description> [-v <version>] [--format sql|js|ts|mjs]
```

डिफ़ॉल्ट SQL और स्वतः संस्करण या -v। `"init users"` जैसे विवरण उद्धरण में दें; स्पेस नाम में `_` बनते हैं।

## DB ऑपरेशन

```sh
migrate create
```

केवल इतिहास तालिका बनाता है, मौजूदा को दोहराता नहीं।

```sh
migrate baseline
```

केवल DB पहले से मेल खाता हो तो बिना चलाए दर्ज। [बेसलाइन](./history#baseline)।

```sh
migrate up [--dry-run]
```

लंबित क्रम से लागू। dry-run इतिहास पढ़ सकता है, बदलाव नहीं करता।

```sh
migrate down <to-version> [--dry-run]
migrate down --all [--dry-run]
```

लक्ष्य ऊपर वापस या सभी। लक्ष्य रहता है, लक्ष्य और --all साथ नहीं।

```sh
migrate status [--output table|json]
```

डिफ़ॉल्ट तालिका। कई DB का JSON नाम क्रम में data_source, status, सफलता में migrations, विफलता में error देता है। create/baseline/status/up/down सभी कनेक्शन पर चलते हैं; --parallel N डिफ़ॉल्ट 8। एक विफलता अन्य नहीं रोकती और निकास कोड गैर-शून्य है।

## बिना कनेक्शन जानकारी

```sh
migrate capabilities [--output text|json]
```

प्रारूप, उपनाम, स्क्रिप्ट और ट्रांज़ैक्शन; डिफ़ॉल्ट text।

```sh
migrate version
```

टूल संस्करण, माइग्रेशन संस्करण नहीं।
