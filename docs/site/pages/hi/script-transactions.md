# ट्रांज़ैक्शन और सीमाएँ

समर्थित DB में स्क्रिप्ट और सफल इतिहास एक ट्रांज़ैक्शन में हैं: सफलता commit, विफलता rollback। खुद BEGIN/COMMIT/ROLLBACK न चलाएँ। अप्रत्यक्ष DDL commit और गैर-ट्रांज़ैक्शन DB बदलाव छोड़ सकते हैं।

```js
export const config = { transaction: false };

export function up(db) {
  // PostgreSQL: यह कथन ट्रांज़ैक्शन के बाहर चलना चाहिए।
  db.exec("CREATE INDEX CONCURRENTLY users_name_idx ON users (name)");
}

export function down(db) {
  db.exec("DROP INDEX CONCURRENTLY users_name_idx");
}
```

transaction boolean है, डिफ़ॉल्ट true, up/down दोनों पर लागू। status/baseline/dry-run निर्यातित रनटाइम कॉन्फ़िगरेशन नहीं आँकते। एक माइग्रेशन समर्पित कनेक्शन साझा करता है। अपनी ट्रांज़ैक्शन समाप्त करें और session सेटिंग लौटाकर ही वापस आएँ; अगले माइग्रेशन तक session रहने पर निर्भर न हों। गैर-ट्रांज़ैक्शन विफलता सफल इतिहास बिना बदलाव छोड़ सकती है।

## संसाधन सीमाएँ

```yaml
script_timeout: 5m # हर स्क्रिप्ट की समय सीमा
script_max_rows: 1000 # एक क्वेरी में अधिकतम पंक्तियाँ
script_max_bytes: 16777216 # एक क्वेरी में अधिकतम बाइट (16 MiB)
```

सीमा पार होने पर काटना नहीं, विफलता है। बड़ी पढ़ाई पृष्ठों में करें; उससे batch commit या स्वतः पुनः शुरू नहीं होता। cancel के बाद DB रुकने में समय लग सकता है। Go में `Config.ScriptOptions` दें। [DB व्यवहार](./databases#script-support) देखें।
