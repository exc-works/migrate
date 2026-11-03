import assert from 'node:assert/strict'
import test from 'node:test'
import { rewriteLink, sourceForRoute } from './links.mjs'

test('shared repository pages link to published routes', () => {
  assert.equal(rewriteLink('script-migrations.en.md', 'guide.md'), 'scripting.md')
  assert.equal(rewriteLink('../../docs/script-migrations.md#共享模块', 'zh/examples.md'), 'scripting.md#共享模块')
  assert.equal(rewriteLink('README.en.md', 'zh/examples.md'), '../examples.md')
})

test('specs and source directories link back to GitHub', () => {
  assert.equal(rewriteLink('spec/07-script-migrations.md', 'zh/scripting.md'),
    'https://github.com/exc-works/migrate/blob/main/docs/spec/07-script-migrations.md')
  assert.equal(rewriteLink('local', 'examples.md'),
    'https://github.com/exc-works/migrate/blob/main/examples/script-imports/local')
})

test('downloadable declarations keep a local asset link', () => {
  assert.equal(rewriteLink('migration-api.d.ts', 'scripting.md'), 'assets/migration-api.d.ts')
})

test('external links, fragments and handwritten pages stay intact', () => {
  for (const href of ['https://example.com/a', '#section', 'mailto:example@example.com']) {
    assert.equal(rewriteLink(href, 'zh/guide.md'), href)
  }
  assert.equal(rewriteLink('./guide', 'zh/index.md'), './guide')
})

test('missing repository targets fail instead of producing broken links', () => {
  assert.throws(() => rewriteLink('missing.md', 'guide.md'), new RegExp('Missing repository link'))
})

test('edit links address canonical source files', () => {
  assert.equal(sourceForRoute('zh/guide.md'), 'docs/user-guide.md')
  assert.equal(sourceForRoute('en/index.md'), 'docs/site/pages/en/index.md')
})
