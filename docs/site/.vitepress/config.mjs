import { defineConfig } from 'vitepress'
import { repository } from '../content.mjs'
import { repositoryLinks, sourceForRoute } from '../links.mjs'

const siteURL = new URL('https://exc-works.github.io/migrate/')
const route = path => `/${path}`
const sidebar = locale => {
  const prefix = locale === 'zh' ? 'zh/' : ''
  return [
    {
      text: locale === 'zh' ? '文档' : 'Documentation',
      items: [
        { text: locale === 'zh' ? '概览' : 'Overview', link: route(prefix) },
        { text: locale === 'zh' ? '用户指南' : 'User guide', link: route(`${prefix}guide`) },
        { text: locale === 'zh' ? 'JS / TS 迁移' : 'JS / TS migrations', link: route(`${prefix}scripting`) },
        { text: locale === 'zh' ? '可运行示例' : 'Runnable examples', link: route(`${prefix}examples`) },
        { text: 'Changelog', link: route('changelog') }
      ]
    }
  ]
}

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
      themeConfig: { sidebar: sidebar('zh'), outlineTitle: '本页内容' }
    }
  },
  themeConfig: {
    nav: [
      { text: 'Guide', link: route('guide') },
      { text: 'Releases', link: `${repository}/releases` }
    ],
    search: { provider: 'local' },
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
