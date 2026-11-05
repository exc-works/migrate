import assert from 'node:assert/strict'
import test from 'node:test'
import { rewriteLink, sourceForRoute } from './links.mjs'
import { sourceRoutes, pageRoutes, repositoryRoot } from './content.mjs'
import { sidebar, topicSlugs, navigationLabels } from './navigation.mjs'
import { languages } from './locales.mjs'
import { existsSync, readFileSync } from 'node:fs'

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

test('every language has a complete sidebar with local topic pages', () => {
  for (const language of languages) {
    const labels = navigationLabels(language.key)
    assert.equal(labels.titles.length, topicSlugs.length, language.key)
    assert.equal(labels.groups.length, 6, language.key)
    const topicLinks = sidebar(language.key).slice(0, -1).flatMap(section => section.items.map(item => item.link))
    assert.deepEqual(topicLinks, topicSlugs.map(slug => `/${language.prefix}${slug}`))
    for (const { items } of sidebar(language.key)) {
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

test('release records can link to authored website pages', () => {
  assert.equal(rewriteLink('../site/pages/quick-start.md', 'releases/v1.5.0.md'), '../quick-start.md')
  assert.equal(rewriteLink('../site/pages/zh/quick-start.md', 'releases/v1.5.0.md'), '../zh/quick-start.md')
})

test('every existing language has an authored website entry and remains selectable', async () => {
  const { readdirSync } = await import('node:fs')
  const { languages } = await import('./locales.mjs')
  const { default: config } = await import('./.vitepress/config.mjs')
  const translated = readdirSync(new URL('../', import.meta.url))
    .filter(file => file === 'user-guide.md' || /^user-guide\.[^.]+\.md$/.test(file))
    .map(file => `docs/${file}`)
  assert.deepEqual(new Set(languages.map(language => language.legacySource)), new Set(translated))
  for (const language of languages) {
    assert.equal(pageRoutes.get(language.source), language.guideRoute)
    assert.equal(existsSync(new URL(language.source, repositoryRoot)), true, language.source)
    assert.equal(config.locales[language.key].link, `/${language.prefix}`)
    assert.equal(config.locales[language.key].lang, language.lang)
    assert.equal(sourceForRoute(language.guideRoute), language.source)
  }
  assert.equal(config.locales.ar.dir, 'rtl')
})

test('localized homepages and onboarding examples match the complete site', () => {
  const read = route => readFileSync(new URL(`pages/${route}`, import.meta.url), 'utf8')
  const blocks = (text, language) => text.split('```')
    .filter((block, index) => index % 2 === 1 && block.split(String.fromCharCode(10))[0] === language)
    .map(block => block.slice(language.length + 1).trimEnd())
  const english = read('quick-start.md')
  assert.equal(blocks(english, 'sh').length, 5)
  assert.equal(blocks(english, 'sql').length, 1)
  for (const language of languages) {
    const home = read(`${language.prefix}index.md`)
    assert.ok(home.split(String.fromCharCode(10)).includes('layout: home'))
    assert.equal(home.split(String.fromCharCode(10)).filter(line => line.startsWith('  - title:')).length, 6, language.key)
    for (const slug of topicSlugs) {
      const content = read(`${language.prefix}${slug}.md`)
      assert.match(content, /^# /, `${language.key}/${slug}`)
      assert.doesNotMatch(content, /\{\{(?:code|connections|aliases)[:}]/, `${language.key}/${slug}`)
    }
    const quickStart = read(`${language.prefix}quick-start.md`)
    assert.deepEqual(blocks(quickStart, 'sh'), blocks(english, 'sh'), language.key)
    assert.deepEqual(blocks(quickStart, 'sql'), blocks(english, 'sql'), language.key)
  }
})
