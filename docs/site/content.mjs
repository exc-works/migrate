import { readdirSync } from 'node:fs'

export const repository = 'https://github.com/exc-works/migrate'
export const repositoryRoot = new URL('../../', import.meta.url)
export const sourceRoutes = new Map([
  // Only release records are shared with the repository. Product documentation
  // is authored in pages/, including topic pages and language entry pages.
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
