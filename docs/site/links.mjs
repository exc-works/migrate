import { existsSync } from 'node:fs'
import { posix } from 'node:path'
import { repository, repositoryRoot, sourceRoutes } from './content.mjs'

const routeSources = new Map([...sourceRoutes].map(([source, route]) => [route, source]))

export function sourceForRoute(route) {
  return routeSources.get(route) ?? `docs/site/pages/${route}`
}

// Resolve links from the original file, so GitHub Markdown stays readable too.
export function rewriteLink(href, route) {
  const source = routeSources.get(route)
  if (!source || /^(?:[a-z][a-z\d+.-]*:|\/|#)/i.test(href)) return href
  const [, pathname, suffix] = href.match(/^([^?#]*)(.*)$/)
  if (!pathname) return href
  const resolved = posix.normalize(posix.join(posix.dirname(source), decodeURIComponent(pathname)))
  if (resolved.startsWith('..') || !existsSync(new URL(resolved, repositoryRoot))) {
    throw new Error(`Missing repository link in ${source}: ${href}`)
  }
  const target = sourceRoutes.get(resolved)
    ?? (resolved === 'docs/migration-api.d.ts' ? 'assets/migration-api.d.ts' : undefined)
  if (target) return `${posix.relative(posix.dirname(route), target)}${suffix}`
  return `${repository}/blob/main/${resolved}${suffix}`
}

export function repositoryLinks(md) {
  md.core.ruler.after('inline', 'repository-links', state => {
    const visit = tokens => {
      for (const token of tokens) {
        if (token.type === 'link_open') {
          token.attrSet('href', rewriteLink(token.attrGet('href'), state.env.relativePath))
        }
        if (token.children) visit(token.children)
      }
    }
    visit(state.tokens)
  })
}
