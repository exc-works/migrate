# حل المشكلات

ابدأ بـ`migrate version` و`migrate capabilities --output json` و`migrate status --output json`. الأخير فقط يحتاج اتصالًا.

تحقق من PATH في [التثبيت](./installation)، ولّد الإعدادات في المشروع أو حدد `-c` / `-w`. في YAML راجع الموقع والقيم واستخدم مسافات لا Tab. انظر [ترتيب البحث](./configuration#configuration-file-search-order).

`can't find env value for XXX` يتطلب `${XXX}`، أو `${XXX:default}` لقيمة غير سرية مناسبة. حقل SSH password_env اسم متغير.

تحقق من capabilities وحدّث النسخ القديمة. unsupported dialect يحتاج مراجعة [البدائل](./databases#aliases). للوحدات تحقق من المسار وتثبيت npm و`DependencyDir` للمصادر المضمّنة. الاعتماد على Node أو async غير متوافق.

استعد الملفات المطبقة وأسماءها من Git وأضف إصدارًا أعلى. لا تتجاوز حالة غير معروفة بـbaseline. قد تنجح الأوامر الصامتة؛ تحقق من applied وbaseline وoutOfOrder.

`to-version must be set, or use --all` يحتاج هدفًا أو `--all`. الهدف يبقى، والسكربت دون down لا يرجع.

قبل التكرار افحص البيانات والسجل: DDL الضمني والعمليات دون معاملات والانقطاع عند COMMIT قد تترك تغييرات، وbaseline لا يصلحها. راجع أسرار SSH وknown_hosts والوكيل ووصول الخادم الوسيط إلى DB. لا اتصال أو إعادة SQL تلقائيًا. استخدم سجلات خادم مضبوطة دون كشف أسرار أو بيانات.
