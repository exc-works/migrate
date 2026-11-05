# साझा मॉड्यूल

`.js`, `.ts`, `.mjs` का स्थिर import और पुनः निर्यात समर्थित है। सहायक फ़ाइल स्वतंत्र माइग्रेशन नहीं है।

```text
project/
  migrations/V2__normalize.ts
  shared/names.ts
```

```ts
import { normalize } from "../shared/names.ts";

export function up(db) {
  db.exec("UPDATE users SET name=? WHERE id=?", [normalize(" Alice "), "1"]);
}
```

पथ आयात करने वाली फ़ाइल से है। `shared/names.ts` में normalize दें और users पहले बनाएँ। [पूरा उदाहरण](./examples#local)।

npm lockfile से पहले इंस्टॉल करें। exports, उपपथ, भीतर की निर्भरता और स्थिर CommonJS समर्थित हैं; `neutral`, फिर `module`, `main` चुने जाते हैं। Node API, native extension, गतिशील import/require नहीं। migrate पैकेज इंस्टॉल नहीं करता।

## एम्बेड किए स्रोत

`FSSource` अपने `fs.FS` जैसे `embed.FS` में एक्सटेंशन और index ढूँढता है, लेकिन आभासी package.json/node_modules नहीं। npm के लिए असली डायरेक्टरी दें:

```go
source := migrate.WithModules{
    Source: migrate.FSSource{
        FS:   embeddedFiles,
        Root: "migrations",
    },
    DependencyDir: projectDirectory,
}
```

सापेक्ष फ़ाइलें आभासी ही रहती हैं। `DependencyDir` बिना host नहीं खोजता। `StringSource` में `WithModules` के `FS`/`Root` या `Directory` दें। अपना स्रोत `MigrationEntriesSource` लागू कर सकता है। [Go](./go-library) देखें।

## पुनरुत्पादन {#checksums}

हैश केवल प्रवेश फ़ाइल का है। सहायक, पैकेज, lockfile बदलने से असंगति या लागू माइग्रेशन का पुनः रन नहीं होता। निर्भरताएँ स्थिर रखें और पुराना व्यवहार बचाएँ। चुने दायरे की सभी निर्भरताएँ लिखने से पहले संकलित होती हैं; बाद की त्रुटि पहले चरण भी रोकती है।
