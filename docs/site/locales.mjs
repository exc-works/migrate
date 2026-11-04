const definitions = [
  ['root', 'en', 'English', 'user-guide.en.md', ['User guide', 'Release notes', 'Search docs', 'On this page', 'Edit this page']],
  ['zh', 'zh-CN', '简体中文', 'user-guide.md', ['用户指南', '发布说明', '搜索文档', '本页内容', '编辑此页']],
  ['zh-Hant', 'zh-Hant', '繁體中文', 'user-guide.zh-Hant.md', ['使用指南', '版本說明', '搜尋文件', '本頁內容', '編輯此頁']],
  ['ja', 'ja', '日本語', 'user-guide.ja.md', ['ユーザーガイド', 'リリースノート', 'ドキュメントを検索', '目次', 'このページを編集']],
  ['ko', 'ko', '한국어', 'user-guide.ko.md', ['사용자 가이드', '릴리스 노트', '문서 검색', '목차', '이 페이지 편집']],
  ['es', 'es', 'Español', 'user-guide.es.md', ['Guía de usuario', 'Notas de la versión', 'Buscar', 'En esta página', 'Editar esta página']],
  ['fr', 'fr', 'Français', 'user-guide.fr.md', ['Guide utilisateur', 'Notes de version', 'Rechercher', 'Sur cette page', 'Modifier cette page']],
  ['de', 'de', 'Deutsch', 'user-guide.de.md', ['Benutzerhandbuch', 'Versionshinweise', 'Suchen', 'Auf dieser Seite', 'Diese Seite bearbeiten']],
  ['pt-BR', 'pt-BR', 'Português do Brasil', 'user-guide.pt-BR.md', ['Guia do usuário', 'Notas da versão', 'Pesquisar', 'Nesta página', 'Editar esta página']],
  ['ru', 'ru', 'Русский', 'user-guide.ru.md', ['Руководство', 'Примечания к выпуску', 'Поиск', 'На этой странице', 'Редактировать страницу']],
  ['ar', 'ar', 'العربية', 'user-guide.ar.md', ['دليل المستخدم', 'ملاحظات الإصدار', 'بحث', 'في هذه الصفحة', 'تحرير هذه الصفحة']],
  ['hi', 'hi', 'हिन्दी', 'user-guide.hi.md', ['उपयोगकर्ता गाइड', 'रिलीज़ नोट्स', 'खोजें', 'इस पृष्ठ पर', 'पृष्ठ संपादित करें']]
]

export const languages = definitions.map(([key, lang, label, file, [guide, releases, search, outline, edit]]) => {
  const prefix = key === 'root' ? '' : `${key}/`
  return {
    key,
    lang,
    label,
    source: `docs/${file}`,
    home: `${prefix}index.md`,
    guideRoute: `${prefix}${key === 'root' || key === 'zh' ? 'guide' : 'index'}.md`,
    prefix,
    ui: { guide, releases, search, outline, edit }
  }
})
