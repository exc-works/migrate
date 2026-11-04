import { readdirSync } from 'node:fs'
import { languages } from './locales.mjs'

export const repository = 'https://github.com/exc-works/migrate'
export const repositoryRoot = new URL('../../', import.meta.url)
export const sourceRoutes = new Map([
  ...languages.map(language => [language.source, language.guideRoute]),
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

export const pageRoutes = new Map(
  readdirSync(new URL('pages/', import.meta.url), { recursive: true })
    .filter(file => file.endsWith('.md'))
    .map(file => {
      const route = file.replaceAll('\\', '/')
      return [`docs/site/pages/${route}`, route]
    })
)
