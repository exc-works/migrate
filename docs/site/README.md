# Documentation website

Use Node.js 22+ and the pnpm version pinned in `package.json`.

## Run locally

```sh
pnpm --dir docs/site install --frozen-lockfile
pnpm --dir docs/site run dev
```

## Edit

- Edit user guides and script guides in `docs/`, and runnable examples in `examples/`.
- Add website pages in `docs/site/pages/`.
- Register a language in `locales.mjs`; its guide is published from the existing translation.
- Register other shared documents in `content.mjs`.
- Add release notes in `docs/releases/`; the index is generated automatically.
- Restart `dev` after editing shared documents to refresh generated content.
- Commit `pnpm-lock.yaml` when changing dependencies. `pnpm-workspace.yaml` pins Vite and permits esbuild's build script.

## Verify

```sh
pnpm --dir docs/site test
pnpm --dir docs/site run build
python3 scripts/check-docs-site.py
python3 scripts/test-docs-quick-start.py
pnpm --dir docs/site run preview
```

## Publish

Set **Settings → Pages → Source: GitHub Actions** and allow `main` in the
`github-pages` environment. Push to main, or run the `docs` workflow on main.
The site is published at https://exc-works.github.io/migrate/.

For a deployment error, check the `docs` workflow logs, Pages source and environment settings.
When changing the domain or repository name, update the VitePress base and the URL in `scripts/check-docs-site.py`.
