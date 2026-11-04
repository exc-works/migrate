import assert from 'node:assert/strict'
import test from 'node:test'
import { rewriteLink, sourceForRoute } from './links.mjs'
import { sourceRoutes, repositoryRoot } from './content.mjs'
import { sidebar } from './navigation.mjs'
import { existsSync } from 'node:fs'

test('release records link to published release routes', () => {
  assert.equal(rewriteLink('docs/releases/v1.5.0.md', 'changelog.md'), 'releases/v1.5.0.md')
  assert.equal(rewriteLink('../../CHANGELOG.md', 'releases/v1.5.0.md'), '../changelog.md')
})

test('specs and source directories link back to GitHub', () => {
  assert.equal(rewriteLink('../spec/07-script-migrations.md', 'releases/v1.5.0.md'),
    'https://github.com/exc-works/migrate/blob/main/docs/spec/07-script-migrations.md')
  assert.equal(rewriteLink('../../cmd/migrate', 'releases/v1.5.0.md'),
    'https://github.com/exc-works/migrate/blob/main/cmd/migrate')
})

test('downloadable declarations keep a local asset link', () => {
  assert.equal(rewriteLink('../migration-api.d.ts', 'releases/v1.5.0.md'), '../assets/migration-api.d.ts')
})

test('external links, fragments and handwritten pages stay intact', () => {
  for (const href of ['https://example.com/a', '#section', 'mailto:example@example.com']) {
    assert.equal(rewriteLink(href, 'zh/guide.md'), href)
  }
  assert.equal(rewriteLink('./guide', 'zh/index.md'), './guide')
})

test('missing repository targets fail instead of producing broken links', () => {
  assert.throws(() => rewriteLink('missing.md', 'changelog.md'), new RegExp('Missing repository link'))
})

test('edit links address canonical source files', () => {
  assert.equal(sourceForRoute('zh/guide.md'), 'docs/site/pages/zh/guide.md')
  assert.equal(sourceForRoute('index.md'), 'docs/site/pages/index.md')
  assert.equal(sourceForRoute('releases.md'), 'docs/site/prepare.mjs')
})

test('both language sidebars resolve to independently authored topic pages', () => {
  for (const locale of ['en', 'zh']) {
    for (const { items } of sidebar(locale)) {
      for (const { link } of items) {
        const route = `${link.slice(1)}.md`
        if (['releases.md', 'changelog.md', 'languages.md'].includes(route)) continue
        assert.equal(existsSync(new URL(sourceForRoute(route), repositoryRoot)), true, route)
      }
    }
  }
  assert.ok([...sourceRoutes.keys()].every(source =>
    source === 'CHANGELOG.md' || source.startsWith('docs/releases/')))
})
