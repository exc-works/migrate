import { navigationTranslations } from './navigation-translations.mjs'

// Every language uses the same topic filenames and reading order.
const sections = [
  ['Start here', '开始使用', [
    ['guide', 'Documentation map', '文档导览'],
    ['quick-start', 'Quick start', '快速开始'],
    ['installation', 'Installation', '安装'],
    ['configuration', 'Configuration', '配置']
  ]],
  ['Manage migrations', '管理迁移', [
    ['sql-migrations', 'Write SQL migrations', '编写 SQL 迁移'],
    ['applying', 'Apply migrations', '执行迁移'],
    ['rollback', 'Roll back migrations', '回退迁移'],
    ['history', 'History and baseline', '历史与基线']
  ]],
  ['JavaScript & TypeScript', 'JavaScript 与 TypeScript', [
    ['scripting', 'Write script migrations', '编写脚本迁移'],
    ['modules', 'Import shared modules', '导入共享模块'],
    ['script-api', 'Database API', '数据库 API'],
    ['script-transactions', 'Transactions and limits', '事务与执行限制'],
    ['examples', 'Run the examples', '运行示例']
  ]],
  ['Integrations', '集成', [
    ['go-library', 'Use from Go', '在 Go 中使用'],
    ['ssh-tunnels', 'SSH tunnels and proxies', 'SSH 隧道与代理']
  ]],
  ['Reference', '参考', [
    ['commands', 'CLI reference', '命令行参考'],
    ['databases', 'Supported databases', '支持的数据库'],
    ['troubleshooting', 'Troubleshooting', '问题排查']
  ]]
]

export const topicSlugs = sections.flatMap(([, , pages]) => pages.map(([slug]) => slug))

export function navigationLabels(locale) {
  if (locale === 'root' || locale === 'en' || locale === 'zh') {
    const zh = locale === 'zh'
    return {
      groups: sections.map(section => section[zh ? 1 : 0]).concat(zh ? '项目' : 'Project'),
      titles: sections.flatMap(([, , pages]) => pages.map(page => page[zh ? 2 : 1])),
      extras: zh ? ['发布说明', '变更记录', '语言'] : ['Release notes', 'Changelog', 'Languages'],
      footer: zh ? ['上一页', '下一页'] : ['Previous page', 'Next page']
    }
  }
  const labels = navigationTranslations[locale]
  if (!labels) throw new Error(`Unknown documentation locale: ${locale}`)
  return labels
}

export function sidebar(locale) {
  const labels = navigationLabels(locale)
  const prefix = locale === 'root' || locale === 'en' ? '' : `${locale}/`
  return [
    ...sections.map(([, , pages], index) => ({
      text: labels.groups[index],
      items: pages.map(([slug]) => ({
        text: labels.titles[topicSlugs.indexOf(slug)],
        link: `/${prefix}${slug}`
      }))
    })),
    {
      text: labels.groups[5],
      items: ['releases', 'changelog', 'languages'].map((slug, index) => ({
        text: labels.extras[index],
        link: `/${slug}`
      }))
    }
  ]
}
