import { readdirSync } from 'node:fs'

export const repository = 'https://github.com/exc-works/migrate'
export const repositoryRoot = new URL('../../', import.meta.url)
export const sourceRoutes = new Map([
  // Only release records are shared with the repository. Product documentation
  // is authored as topic pages in pages/, in both supported website languages.
  ['CHANGELOG.md', 'changelog.md'],
  ...readdirSync(new URL('docs/releases/', repositoryRoot))
    .filter(name => name.endsWith('.md'))
    .sort()
    .map(name => [`docs/releases/${name}`, `releases/${name}`])
])
