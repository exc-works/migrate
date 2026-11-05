# समस्या निवारण

`migrate version`, `migrate capabilities --output json`, `migrate status --output json` से शुरू करें। केवल status को DB चाहिए।

[इंस्टॉलेशन](./installation) में PATH जाँचें। प्रोजेक्ट में कॉन्फ़िगरेशन बनाएँ या -c/-w दें। YAML त्रुटि स्थान, मान और space indentation जाँचें, Tab नहीं। [खोज क्रम](./configuration#configuration-file-search-order) देखें।

can't find env value for XXX में `${XXX}` सेट करें। `${XXX:default}` उचित गैर-सीक्रेट मानों के लिए है। SSH password_env में वेरिएबल नाम दें।

capabilities जाँचकर पुराने संस्करण अपडेट करें। unsupported dialect में [उपनाम](./databases#aliases) जाँचें। मॉड्यूल के पथ, npm इंस्टॉल और embedded source का DependencyDir देखें। Node/async पैकेज नहीं चलते।

लागू फ़ाइल और नाम Git से वापस लाएँ, नया उच्च संस्करण जोड़ें। अनजान स्थिति पर baseline से त्रुटि न छिपाएँ। शांत कमांड सफल हो सकता है; applied, baseline, outOfOrder देखें।

to-version must be set, or use --all में लक्ष्य या --all दें। लक्ष्य सुरक्षित है; down बिना स्क्रिप्ट वापस नहीं जाती।

दोहराने से पहले डेटा/इतिहास देखें: DDL commit, गैर-ट्रांज़ैक्शन और COMMIT के पास disconnect बदलाव छोड़ सकते हैं। baseline नहीं सुधारता। SSH सीक्रेट, known_hosts, proxy और बास्टियन से DB पहुँच जाँचें। स्वतः reconnect या SQL repeat नहीं। नियंत्रित server logs देखें, सीक्रेट या डेटा साझा न करें।
