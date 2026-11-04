import { defineConfig } from 'vitepress'
import { repository } from '../content.mjs'
import { languages } from '../locales.mjs'
import { repositoryLinks, sourceForRoute } from '../links.mjs'

const siteURL = new URL('https://exc-works.github.io/migrate/')
const route = path => `/${path}`
const editPattern = page => `https://github.com/exc-works/migrate/edit/main/${page.frontmatter.sourcePath}`

function sidebar(language) {
  const { key, prefix, ui, guideRoute } = language
  if (key !== 'root' && key !== 'zh') {
    return [{
      text: ui.guide,
      items: [
        { text: ui.guide, link: route(prefix) },
        { text: 'JavaScript / TypeScript · English', link: route('scripting') },
        { text: `${ui.releases} · English`, link: route('releases') }
      ]
    }]
  }
  const zh = key === 'zh'
  return [{
    text: zh ? '开始使用' : 'Get started',
    items: [
      { text: zh ? '快速开始' : 'Quick start', link: route(`${prefix}quick-start`) },
      { text: ui.guide, link: route(guideRoute.slice(0, -3)) },
      { text: zh ? '命令参考' : 'Commands', link: route(`${prefix}commands`) },
      { text: zh ? '数据库' : 'Databases', link: route(`${prefix}databases`) },
      { text: 'JavaScript / TypeScript', link: route(`${prefix}scripting`) },
      { text: zh ? '示例' : 'Examples', link: route(`${prefix}examples`) },
      { text: ui.releases, link: route('releases') }
    ]
  }]
}

export default defineConfig({
  title: 'migrate',
  description: 'SQL, JavaScript and TypeScript database migrations.',
  base: siteURL.pathname,
  srcDir: '.generated',
  cleanUrls: true,
  lastUpdated: false,
  locales: Object.fromEntries(languages.map(language => [language.key, {
    label: language.label,
    lang: language.lang,
    dir: language.key === 'ar' ? 'rtl' : 'ltr',
    link: route(language.prefix),
    themeConfig: {
      sidebar: sidebar(language),
      outlineTitle: language.ui.outline,
      editLink: { text: language.ui.edit, pattern: editPattern },
      ...(language.key === 'zh' ? { docFooter: { prev: '上一页', next: '下一页' } } : {}),
      nav: [
        { text: language.ui.guide, link: route(language.guideRoute.endsWith('index.md') ? language.prefix : language.guideRoute.slice(0, -3)) },
        { text: language.ui.releases, link: route('releases') }
      ]
    }
  }])),
  themeConfig: {
    i18nRouting: false,
    search: {
      provider: 'local',
      options: {
        miniSearch: {
          options: {
            tokenize: text => Array.from(new Intl.Segmenter(undefined, { granularity: 'word' }).segment(text))
              .filter(part => part.isWordLike).map(part => part.segment)
          }
        },
        locales: Object.fromEntries(languages.map(language => [language.key, {
          translations: {
            button: { buttonText: language.ui.search, buttonAriaLabel: language.ui.search }
          }
        }]))
      }
    },
    outline: [2, 3],
    socialLinks: [{ icon: 'github', link: repository }]
  },
  transformPageData(page) {
    page.frontmatter.sourcePath = sourceForRoute(page.relativePath)
  },
  markdown: { config: repositoryLinks }
})
