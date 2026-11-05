# उदाहरण चलाएँ

रिपोजिटरी की कॉपी में चलाएँ। सभी SQLite उपयोग कर अपने डेटा जाँचते हैं।

## अस्थायी CLI

उसी shell में आगे चलें।

```sh
example_workspace=$(mktemp -d)
go build -o "$example_workspace/migrate" ./cmd/migrate
```

## स्थानीय मॉड्यूल {#local}

[स्रोत](https://github.com/exc-works/migrate/tree/main/examples/script-imports/local): दो माइग्रेशन TypeScript सहायक साझा करते हैं।

```sh
export MIGRATE_EXAMPLE_DB="$example_workspace/local.sqlite"
"$example_workspace/migrate" -w examples/script-imports/local create
"$example_workspace/migrate" -w examples/script-imports/local up --dry-run
"$example_workspace/migrate" -w examples/script-imports/local up
"$example_workspace/migrate" -w examples/script-imports/local status
"$example_workspace/migrate" -w examples/script-imports/local down 2
"$example_workspace/migrate" -w examples/script-imports/local down --all
```

up के बाद तीन संस्करण और नाम alice, bob हैं। down 2 संस्करण 1, 2 रखता है; down --all तालिका हटाता है।

## npm पैकेज

Node/npm सिर्फ इंस्टॉल के लिए। lockfile स्थानीय पैकेज का है, registry नहीं चाहिए।

```sh
npm --prefix examples/script-imports/npm ci --ignore-scripts --no-audit --no-fund --offline
export MIGRATE_EXAMPLE_DB="$example_workspace/npm.sqlite"
"$example_workspace/migrate" -w examples/script-imports/npm create
"$example_workspace/migrate" -w examples/script-imports/npm up
"$example_workspace/migrate" -w examples/script-imports/npm status
"$example_workspace/migrate" -w examples/script-imports/npm down --all
```

दूसरा माइग्रेशन `@migrate-example/normalizer/names` लाकर alice जाँचता है। दो संस्करण लागू हैं।

## Go में एम्बेड

```sh
go run ./examples/script-imports/embedded
```

embed.FS और निजी मेमोरी DB, npm बिना। आउटपुट:

```text
after up: alice
after down: all migrations pending
```

## सफ़ाई

```sh
rm -r "$example_workspace"
```

npm का अनदेखा node_modules बचता है। root से `./scripts/test-script-examples.sh` अस्थायी कॉपी जाँचता है। [पुनरुत्पादन](./modules#checksums) देखें।
