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

Author user documentation directly in `pages/`, with matching topic filenames under
`pages/zh/`. These pages are the maintained website content; the build does not import
the old user guides, script guides or example READMEs. Organize new content around a
reader's task and add it to the appropriate section in `navigation.mjs` in both languages.
Keep tutorials, operational guides and reference material on their own pages.

`content.mjs` maps only CHANGELOG and release records to website routes. These remain
maintained at their repository paths. `prepare.mjs` assembles topic pages and release
records in ignored `.generated` before development and builds; do not edit that output.

Use relative website links between topic pages. Link to GitHub when readers need to
inspect source code or runnable example directories. Links within shared release
records are resolved from their canonical repository paths.
The type declaration is copied as a downloadable asset. Specs and validation notes
are not included in the website build or search index.

Documents follow main, rather than claiming all described features exist in every
release. Check release notes when using an older binary. Other translations remain
available from the website's language directory until they are ported.

VitePress is pinned to its stable release. The override in `pnpm-workspace.yaml` selects the patched
6.4 maintenance line because VitePress 1.x otherwise installs Vite 5. Validate
site builds, development and preview when updating either dependency.

The language menu opens each locale's homepage, because release/changelog pages do
not have translated counterparts. Restart `dev` after editing topic pages or release
records to refresh the assembled content.

Run the actual bilingual SQLite onboarding examples with:

```sh
python3 scripts/test-docs-quick-start.py
```

The release index is generated from `docs/releases/`; add new notes there rather
than editing generated pages.

Commit `pnpm-lock.yaml` when dependencies change. CI installs with `--frozen-lockfile`.
Only esbuild is allowed to run its dependency build script.

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

If a page has a broken link or fragment, fix its authored topic page (or the original
release record). Rebuild before previewing changes with `preview`.
For a Pages 404, confirm the workflow source setting, deployment environment and
project base, then check the `docs` workflow's deployment result.
