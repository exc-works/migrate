import { defineConfig } from 'vitepress'
import { repository } from '../content.mjs'
import { repositoryLinks, sourceForRoute } from '../links.mjs'
import { sidebar } from '../navigation.mjs'

const siteURL = new URL('https://exc-works.github.io/migrate/')
const route = path => `/${path}`

export default defineConfig({
  title: 'migrate',
  description: 'SQL, JavaScript and TypeScript database migrations in one Go tool.',
  base: siteURL.pathname,
  srcDir: '.generated',
  cleanUrls: true,
  lastUpdated: false,
  locales: {
    root: { label: 'English', lang: 'en', themeConfig: { sidebar: sidebar('en') } },
    zh: {
      label: '简体中文',
      lang: 'zh-CN',
      link: route('zh/'),
      themeConfig: {
        sidebar: sidebar('zh'),
        outlineTitle: '本页内容',
        docFooter: { prev: '上一页', next: '下一页' },
        editLink: {
          text: '编辑此页',
          pattern: page => `https://github.com/exc-works/migrate/edit/main/${page.frontmatter.sourcePath}`
        },
        footer: { message: '文档跟随 main 分支；版本可用性请查阅对应发布说明。' },
        nav: [
          { text: '指南', link: route('zh/guide') },
          { text: '发布说明', link: route('releases') }
        ]
      }
    }
  },
  themeConfig: {
    // Shared release/changelog pages have no translated counterpart.
    i18nRouting: false,
    nav: [
      { text: 'Guide', link: route('guide') },
      { text: 'Releases', link: route('releases') }
    ],
    search: {
      provider: 'local',
      options: {
        miniSearch: {
          options: {
            tokenize: text => Array.from(new Intl.Segmenter(undefined, { granularity: 'word' }).segment(text))
              .filter(part => part.isWordLike).map(part => part.segment)
          }
        },
        locales: {
          zh: {
            translations: {
              button: { buttonText: '搜索文档', buttonAriaLabel: '搜索文档' },
              modal: { noResultsText: '没有找到结果', resetButtonTitle: '清空搜索', backButtonTitle: '关闭搜索' }
            }
          }
        }
      }
    },
    outline: [2, 3],
    socialLinks: [{ icon: 'github', link: repository }],
    editLink: {
      pattern: page => `https://github.com/exc-works/migrate/edit/main/${page.frontmatter.sourcePath}`
    },
    footer: { message: 'Documentation follows main. Check release notes for version availability.' }
  },
  transformPageData(page) {
    page.frontmatter.sourcePath = sourceForRoute(page.relativePath)
  },
  markdown: { config: repositoryLinks }
})
