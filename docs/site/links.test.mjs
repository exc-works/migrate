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
  assert.equal(sourceForRoute('index.md'), 'docs/site/pages/index.md')
  assert.equal(sourceForRoute('releases.md'), 'docs/site/prepare.mjs')
})

test('shared guides can link to the website quick start', () => {
  assert.equal(rewriteLink('site/pages/quick-start.md', 'guide.md'), 'quick-start.md')
  assert.equal(rewriteLink('site/pages/zh/quick-start.md', 'zh/guide.md'), 'quick-start.md')
})

test('every existing translation is published and selectable', async () => {
  const { readdirSync } = await import('node:fs')
  const { sourceRoutes } = await import('./content.mjs')
  const { languages } = await import('./locales.mjs')
  const { default: config } = await import('./.vitepress/config.mjs')
  const translated = readdirSync(new URL('../', import.meta.url))
    .filter(file => file === 'user-guide.md' || /^user-guide\.[^.]+\.md$/.test(file))
    .map(file => `docs/${file}`)
  assert.deepEqual(new Set(languages.map(language => language.source)), new Set(translated))
  for (const language of languages) {
    assert.equal(sourceRoutes.get(language.source), language.guideRoute)
    assert.equal(config.locales[language.key].link, `/${language.prefix}`)
    assert.equal(config.locales[language.key].lang, language.lang)
    assert.equal(sourceForRoute(language.guideRoute), language.source)
  }
  assert.equal(config.locales.ar.dir, 'rtl')
})
