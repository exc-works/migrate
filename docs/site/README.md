# Documentation website

Use Node.js 22+ and the pnpm version pinned in `package.json`.

## Run locally

```sh
pnpm --dir docs/site install --frozen-lockfile
pnpm --dir docs/site run dev
```

## Edit

- Author user documentation in `pages/`, with matching topic filenames under every registered language directory.
  Organize content around reader tasks in `navigation.mjs`, keeping tutorials,
  operational guides and references on separate pages.
- Register languages in `locales.mjs`. Every language has an authored homepage, all topic pages, localized navigation, and
  previous/next links. Keep `navigation-translations.mjs` aligned with the shared
  topic order in `navigation.mjs`; do not substitute English or repository links
  for a translated topic. Shared release records remain in their original languages.
- The website does not import old user guides, script guides or example READMEs.
  Link to GitHub when readers need source code, original translations or example projects.
- `content.mjs` shares only CHANGELOG and release records. Add release notes in
  `docs/releases/`; the release index is generated automatically.
- `prepare.mjs` assembles content in ignored `.generated`. Do not edit generated
  files. Restart `dev` after editing content to refresh the assembled pages.
- Topic links are relative to website pages. Links in release records are resolved
  from their original repository paths; edit links point to the actual source.
- The API declaration is copied as a download. Specs and validation notes stay out
  of the website and search index.
- Commit `pnpm-lock.yaml` when changing dependencies. `pnpm-workspace.yaml` pins
  Vite and permits esbuild's build script.

Documentation follows main. Use `migrate version` and release notes to identify
features shipped in an installed binary. The language menu opens each language's
homepage because shared release records have no translated counterparts.

## Verify

```sh
pnpm --dir docs/site test
pnpm --dir docs/site run build
python3 scripts/check-docs-site.py
python3 scripts/test-docs-quick-start.py
pnpm --dir docs/site run preview
```

The artifact check validates rendered pages, local resources, links and fragments.
It does not fetch external websites. Rebuild before previewing changed content.
The onboarding runner requires Go, Python 3 and Node.js. It checks all twelve
localized SQLite walkthroughs and repository script-guide examples against
freshly built binaries. It reads the language registry so newly registered
languages cannot silently skip executable verification.

## Publish

Set **Settings → Pages → Source: GitHub Actions** and allow `main` in the
`github-pages` environment. Push to main, or run the `docs` workflow on main.
The site is published at https://exc-works.github.io/migrate/ after the site checks
and executable onboarding examples pass. Other branches and PRs validate only.

For a deployment error, check the `docs` workflow logs, Pages source and environment
settings. When changing the domain or repository name, update the VitePress base
and the URL in `scripts/check-docs-site.py`.
