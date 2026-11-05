# SQL माइग्रेशन लिखें

एक फ़ाइल में एक तार्किक बदलाव रखें। SQL और स्क्रिप्ट संस्करण साझा करते हैं।

```sh
migrate new version "init users"
migrate new version "add email" -v 202611040002
```

स्पेस वाला विवरण उद्धरण में दें; फ़ाइल नाम में स्पेस `_` बनता है। `-v` संस्करण चुनता है, नहीं तो स्वतः बनता है। प्रारूप `V<version>__<description>.sql` है। लागू फ़ाइल न बदलें/नाम न बदलें; नया संस्करण जोड़ें।

```sql
-- +migrate Up
CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);

-- +migrate Down
DROP TABLE users;
```

Up लागू करता है, Down उलटता है। यहाँ Down डेटा भी हटाता है; उसकी वापसी की योजना बनाएँ।

अंदरूनी सेमीकोलन वाले कथन को `StatementBegin` / `StatementEnd` में रखें:

```sql
-- +migrate Up
-- +migrate StatementBegin
CREATE FUNCTION migration_answer() RETURNS integer AS $$
BEGIN
    RETURN 42;
END;
$$ LANGUAGE plpgsql;
-- +migrate StatementEnd

-- +migrate Down
DROP FUNCTION migration_answer();
```

SQL व्याकरण और DDL ट्रांज़ैक्शन डेटाबेस पर निर्भर हैं। आगे [चलाएँ](./applying)।
