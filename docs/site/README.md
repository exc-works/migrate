# Documentation website

The site uses VitePress and requires Node.js 22 or newer. From the repository root:

```sh
npm --prefix docs/site ci
npm --prefix docs/site run dev
npm --prefix docs/site test
npm --prefix docs/site run build
npm --prefix docs/site run preview
```

Edit existing guides, examples and release notes at their original repository paths.
`content.mjs` is the explicit map of public source documents to website routes.
`prepare.mjs` recreates ignored `.generated` content before development and builds.
Website-only pages live in `pages/`; do not edit generated files.

Relative links in shared documents are resolved from the canonical source path.
Published documents become website links; specs and source files link to GitHub.
The type declaration is copied as a downloadable asset. Specs and validation notes
are not included in the website build or search index.

Documents follow main, rather than claiming all described features exist in every
release. Check release notes when using an older binary. Other translations remain
available from the repository README until they are ported to the website.

VitePress is pinned to its stable release. The Vite override selects the patched
6.4 maintenance line because VitePress 1.x otherwise installs Vite 5. Validate
site builds, development and preview when updating either dependency.

The language menu opens each locale's homepage, because shared release/changelog
pages do not have translated counterparts. Restart `dev` after changing a shared
repository document so the generated copy is refreshed.

Run the actual bilingual SQLite onboarding examples with:

```sh
python3 scripts/test-docs-quick-start.py
```

The release index is generated from `docs/releases/`; add new notes there rather
than editing generated pages.

## GitHub Pages

After the PR stack is merged, enable **Settings → Pages → Build and deployment →
Source: GitHub Actions** in the repository. The workflow publishes to
https://exc-works.github.io/migrate/ after pushes to main. Run the `docs` workflow
manually on main to retry deployment after enabling Pages. Other branches and PRs
build and validate only. No personal access token or committed build output is needed.

The deployment job requires both the site build/link check and the bilingual
SQLite examples to pass. It uses the `github-pages` environment and grants Pages
and OIDC write permissions only to the deployment job. Configure that environment
to allow deployment from main. Existing Go CI gates remain in place.

The artifact check can also be run locally after building:

```sh
python3 scripts/check-docs-site.py
```

It validates rendered pages, links, fragments and local resources under the
configured project base. External sites are not fetched by this check. If the
repository name or domain changes, update both the VitePress base and checker URL.

If a page has a broken link or fragment, fix its canonical Markdown source. If a
shared guide edit does not appear in preview, rebuild to refresh generated content.
For a Pages 404, confirm the workflow source setting, deployment environment and
project base, then check the `docs` workflow's deployment result.
