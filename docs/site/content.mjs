import { readdirSync } from 'node:fs'

export const repository = 'https://github.com/exc-works/migrate'
export const repositoryRoot = new URL('../../', import.meta.url)
export const sourceRoutes = new Map([
  ['docs/user-guide.en.md', 'guide.md'],
  ['docs/user-guide.md', 'zh/guide.md'],
  ['docs/script-migrations.en.md', 'scripting.md'],
  ['docs/script-migrations.md', 'zh/scripting.md'],
  ['examples/script-imports/README.en.md', 'examples.md'],
  ['examples/script-imports/README.md', 'zh/examples.md'],
  ['CHANGELOG.md', 'changelog.md'],
  ...readdirSync(new URL('docs/releases/', repositoryRoot))
    .filter(name => name.endsWith('.md'))
    .sort()
    .map(name => [`docs/releases/${name}`, `releases/${name}`])
])
