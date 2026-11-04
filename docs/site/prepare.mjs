import { cp, mkdir, readFile, rm, writeFile } from 'node:fs/promises'
import { repositoryRoot, sourceRoutes } from './content.mjs'

// Only the explicit public content map is copied; specs and validation notes stay out.
const output = new URL('.generated/', import.meta.url)
await rm(output, { recursive: true, force: true })
await cp(new URL('pages/', import.meta.url), output, { recursive: true })
for (const [source, route] of sourceRoutes) {
  const target = new URL(route, output)
  await mkdir(new URL('./', target), { recursive: true })
  await writeFile(target, await readFile(new URL(source, repositoryRoot)))
}
await mkdir(new URL('public/assets/', output), { recursive: true })
await cp(new URL('docs/migration-api.d.ts', repositoryRoot), new URL('public/assets/migration-api.d.ts', output))
const releases = [...sourceRoutes]
  .filter(([source]) => source.startsWith('docs/releases/'))
  .map(([, route]) => route)
  .sort((a, b) => b.localeCompare(a, 'en', { numeric: true }))
await writeFile(new URL('releases.md', output), [
  '# Release notes / 发布说明',
  '',
  'This website follows main. Select the release matching `migrate version` for shipped behavior.',
  '',
  ...releases.map(route => `- [${route.split('/').pop().slice(0, -3)}](./${route.slice(0, -3)})`),
  '',
  '[Changelog](./changelog) · [Downloads](https://github.com/exc-works/migrate/releases)',
  ''
].join('\n'))
console.log(`Prepared ${sourceRoutes.size} repository documents.`)
