# इंस्टॉलेशन

Go से इंस्टॉल करें या [Releases](https://github.com/exc-works/migrate/releases) से अपने OS और आर्किटेक्चर का बाइनरी लें। SHA-256 जाँचें और `migrate` (Windows में `migrate.exe`) को PATH में रखें।

## मॉड्यूल से इंस्टॉल करें

```bash
go install github.com/exc-works/migrate/cmd/migrate@latest
```

किसी विशेष संस्करण को इंस्टॉल करें:

```bash
go install github.com/exc-works/migrate/cmd/migrate@vX.Y.Z
```

## लोकल सोर्स से इंस्टॉल करें (private repo या internal network)

रिपॉज़िटरी रूट में चलाएँ:

```bash
go install ./cmd/migrate
```

## इंस्टॉलेशन सत्यापित करें

```sh
migrate --help
migrate version
migrate capabilities --output json
```

अगर कमांड नहीं मिलती है:

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

स्थानीय बिल्ड `dev` दिखाते हैं। पुराने Go इंस्टॉल के लिए `go version -m <binary>` देखें। capabilities डेटाबेस से नहीं जुड़ता। आगे [त्वरित शुरुआत](./quick-start) देखें।
