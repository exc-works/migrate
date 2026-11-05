// Each topic has the same route in both languages, making the reading order explicit.
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

export function sidebar(locale) {
  const zh = locale === 'zh'
  const route = path => `/${path}`
  const prefix = zh ? 'zh/' : ''
  return [
    ...sections.map(([en, cn, pages]) => ({
      text: zh ? cn : en,
      items: pages.map(([slug, title, translation]) => ({
        text: zh ? translation : title,
        link: route(`${prefix}${slug}`)
      }))
    })),
    {
      text: zh ? '项目' : 'Project',
      items: [
        { text: zh ? '发布说明' : 'Release notes', link: route('releases') },
        { text: zh ? '变更记录' : 'Changelog', link: route('changelog') },
        { text: zh ? '其他语言' : 'Other languages', link: route('languages') }
      ]
    }
  ]
}
