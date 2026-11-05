import { defineConfig } from 'vitepress'
import { repository } from '../content.mjs'
import { languages } from '../locales.mjs'
import { repositoryLinks, sourceForRoute } from '../links.mjs'
import { sidebar, navigationLabels } from '../navigation.mjs'

const siteURL = new URL('https://exc-works.github.io/migrate/')
const route = path => `/${path}`
const editPattern = page => `https://github.com/exc-works/migrate/edit/main/${page.frontmatter.sourcePath}`

export default defineConfig({
  title: 'migrate',
  description: 'SQL, JavaScript and TypeScript database migrations.',
  base: siteURL.pathname,
  head: [
    ['link', { rel: 'icon', type: 'image/svg+xml', href: `${siteURL.pathname}favicon.svg` }]
  ],
  srcDir: '.generated',
  cleanUrls: true,
  lastUpdated: false,
  locales: Object.fromEntries(languages.map(language => [language.key, {
    label: language.label,
    lang: language.lang,
    dir: language.key === 'ar' ? 'rtl' : 'ltr',
    link: route(language.prefix),
    themeConfig: {
      sidebar: sidebar(language.key),
      outlineTitle: language.ui.outline,
      editLink: { text: language.ui.edit, pattern: editPattern },
      docFooter: { prev: navigationLabels(language.key).footer[0], next: navigationLabels(language.key).footer[1] },
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
