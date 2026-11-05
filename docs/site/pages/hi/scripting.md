# स्क्रिप्ट लिखें

JS, TS, MJS से डेटा बदलें; चलाने के लिए Node.js नहीं चाहिए।

```sh
migrate new version "normalize names" --format ts
```

`V<version>__<description>.<format>` SQL के साथ क्रम साझा करता है। MJS की सीमाएँ JS जैसी हैं।

```ts
export function up(db) {
  for (const row of db.query("SELECT id, name FROM users")) {
    db.exec("UPDATE users SET name=? WHERE id=?", [row.name.trim(), row.id]);
  }
}
```

SQLite में `users.name` का उदाहरण। बनाया up लागू होने तक त्रुटि फेंकता है। केवल सिंक्रोनस फ़ंक्शन: async, generator, Promise नहीं; सामान्य सिंक्रोनस वापसी मान अनदेखा होता है। [SQL पैरामीटर](./script-api) DB पर निर्भर हैं।

स्पेस हटाने से मूल जानकारी जाती है, इसलिए down नहीं है। उलटा ऑपरेशन लिखने के बाद ही `down(db)` निर्यात करें; वापसी चाहिए तो मूल मान रखें।

```sh
migrate up --dry-run
migrate up
migrate status
```

Dry-run स्क्रिप्ट नहीं चलाता। DB काम up/down में रखें, मॉड्यूल के ऊपरी स्तर पर नहीं। माइग्रेशन के बीच साझा global पर निर्भर न हों। टेम्पलेट और [migration-api.d.ts](../assets/migration-api.d.ts) में प्रकार हैं, लेकिन रूपांतरण प्रकार नहीं जाँचता। Node API, नेटवर्क, फ़ाइल, वातावरण, टाइमर, गतिशील लोडिंग उपलब्ध नहीं; npm भी जाँचें। [मॉड्यूल](./modules) और [ट्रांज़ैक्शन](./script-transactions) देखें।
