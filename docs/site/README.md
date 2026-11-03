# Documentation website

The site uses VitePress and requires Node.js 22 or newer and pnpm 10.21.0.
Use the version pinned in `package.json` (`corepack enable` if pnpm is not installed). From the repository root:

```sh
pnpm --dir docs/site install --frozen-lockfile
pnpm --dir docs/site run dev
pnpm --dir docs/site test
pnpm --dir docs/site run build
pnpm --dir docs/site run preview
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

VitePress is pinned to its stable release. The override in `pnpm-workspace.yaml` selects the patched
6.4 maintenance line because VitePress 1.x otherwise installs Vite 5. Validate
site builds, development and preview when updating either dependency.

Commit `pnpm-lock.yaml` when dependencies change. CI installs with `--frozen-lockfile`.
Only esbuild is allowed to run its dependency build script.
