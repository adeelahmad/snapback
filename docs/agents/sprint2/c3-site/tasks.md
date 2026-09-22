---
type: tasks
story: S2-12
---

# C3 tasks (schema id S2-12) — snapback.run landing site

Story, decisions and owned files: `stories.md` (C3). Tests: `plan.md`. Gates: `validate.md`. Go test code follows the Google Go style guide (`docs/agents/go-styleguide/`): table-driven, `t.Errorf("got %v, want %v")`, no assertion libraries, `t.Helper()` in helpers. Go tests live in package `site_test` under `test/site/` and share one `helpers_test.go` (repo root lookup via `go.mod`, file read, JSON load) created in T1.

Fixed inputs (orchestrator supplies; tests never read them):

- `DS=/private/tmp/claude-501/-Users-adeelahmad-work-snapback/3a548722-26e5-4d22-8251-dbadece5e939/scratchpad`
- tokens: `$DS/artifact-files/e75f5781-2fc6-4585-b3b2-227c107dbcbb/project/tokens.json`
- fonts: `$DS/design-system/project/fonts/*.woff2` (six files)
- brand assets: a local folder the orchestrator names in the T3 init as `BRAND_DIR`, holding the files listed in 20-assets.md.

Copy rules for T4/T5 (from intake and SPEC): lowercase `snapback`; sentence case; no emoji; no `!`; command first, then the sentence; state limitations; Restic is the only backend named; today's product is the Stage 0 skeleton (`snapback version` only) plus Stage 1 compatibility evidence; the `.snapshot` view, `link`/`open`/`snap`, web UI, services and macOS are planned, never present tense. Text colours only `ink`, `ink-body`, `ink-muted`, `accent-text` (links). Glass panels (`glass-*`, `shadow-glass`, `radius-md`) over one to three radial `glass-tint-green`/`glass-tint-blue` discs on `canvas`; no gradients elsewhere; the mark flat, from the SVG assets.

## T1 — Scaffold `web/` with the pinned toolchain

Vite + React + TypeScript app in `web/`: `package.json` (scripts `build` = `vite build`, `test` = `vitest run`, `lint` = `tsc --noEmit`, `gen:tokens`, `prebuild` = `npm run gen:tokens`; exact versions only; `engines.node` equal to `.nvmrc`), `package-lock.json`, `.nvmrc` = `26.0.0`, `.npmrc` (`save-exact=true`, `engine-strict=true`), `tsconfig.json` (strict), `vite.config.ts` (`base: '/'`), `vitest.config.ts` (environment `node`), `index.html` (minimal shell, filled in T6), `src/main.tsx`, `src/App.tsx` rendering `<main>` with a placeholder `<h1>snapback</h1>`. Dependencies: `react`, `react-dom`; dev: `vite`, `@vitejs/plugin-react`, `typescript`, `vitest`, `@types/react`, `@types/react-dom`, `@types/node`. Nothing else. `.gitignore` gains `web/node_modules/` and `_site/`. `NOTICE` gains one line per runtime dependency (react, react-dom; MIT). Creates `test/site/helpers_test.go`.

Files: `web/{package.json,package-lock.json,.nvmrc,.npmrc,tsconfig.json,vite.config.ts,vitest.config.ts,index.html}`, `web/src/{main.tsx,App.tsx,App.test.tsx}`, `test/site/{helpers_test.go,scaffold_test.go}`, `.gitignore`, `NOTICE`.

## T2 — Tokens to CSS variables, generated

`web/tokens.json`: byte copy of the design system's tokens.json. `web/scripts/gen-tokens.mjs` exports `generate(tokens) -> string` and, when run as a script, writes `web/src/styles/tokens.css`. Output: `@font-face` per `type.fonts` entry (`url('/fonts/<basename>') format('woff2')`, `font-display: swap`); `:root` with `color-scheme: light dark`, every colour token's light value as `--<name>`, every spacing/radius token, shadow tokens' light values, font-family tokens; `@media (prefers-color-scheme: dark) { :root { ... } }` with every colour and shadow dark value. Token order follows tokens.json; output ends with one newline. `tokens.css` is committed and imported once from `src/main.tsx`. Hand-editing `tokens.css` is a failure (drift test).

Files: `web/tokens.json`, `web/scripts/gen-tokens.mjs`, `web/scripts/gen-tokens.test.ts`, `web/src/styles/tokens.css`, `web/src/main.tsx` (one import line), `test/site/contrast_test.go`.

## T3 — Self-hosted fonts and brand assets

Copy the six `.woff2` files to `web/public/fonts/` under their original names. Copy from `BRAND_DIR` into `web/public/brand/`: `mark.svg`, `mark-inverse.svg`, `lockup-horizontal.svg`, `lockup-horizontal-inverse.svg` (if absent in BRAND_DIR, the orchestrator is told; do not draw one), `wordmark.svg`, `wordmark-inverse.svg`; into `web/public/`: `favicon.svg`, `favicon-16.png`, `favicon-32.png`, `app-icon-180.png`, `app-icon-512.png`, `og-card-light.png`, `og-card-dark.png`. Byte copies only; no re-export, no edits. `NOTICE` gains IBM Plex Sans, IBM Plex Mono and JetBrains Mono under the SIL Open Font License 1.1.

Files: `web/public/fonts/*.woff2`, `web/public/brand/*.svg`, `web/public/{favicon.svg,favicon-16.png,favicon-32.png,app-icon-180.png,app-icon-512.png,og-card-light.png,og-card-dark.png}`, `NOTICE`, `test/site/assets_test.go`, `test/site/icons_test.go`.

## T4 — Content model, header, hero, how it works

`web/src/content.ts` holds every visible string as typed constants (one source for copy, so the lint tests read one place). `web/src/sections/Header.tsx` (horizontal lockup `<img>` with light/dark `<picture>` source, links: docs `/docs/`, GitHub), `Hero.tsx` (one-sentence pitch, the command `curl -fsSL https://snapback.run/install.sh | sh` in a snippet box, a status line saying what exists today), `HowItWorks.tsx` (terminal-style block showing `ls -a` revealing `.snapshot`, `ls .snapshot/`, `cp .snapshot/<timestamp>/file ./file`, labelled as the planned behaviour). `App.tsx` renders Header, Hero, HowItWorks inside `<main>`. `web/src/styles/site.css` holds layout and glass rules using only `var(--…)` token references (no hex literals).

Files: `web/src/content.ts`, `web/src/sections/{Header,Hero,HowItWorks}.tsx`, `web/src/styles/site.css`, `web/src/App.tsx`, `web/src/main.tsx` (one import line for site.css), `web/src/sections.test.tsx`.

## T5 — What it does not do, install, status, footer, and the copy lint

`web/src/sections/Limits.tsx` (SPEC non-goals: no backup scheduling, no retention, no file-content cache, never writes to the Restic repository, no Windows, no live overlay; Restic only), `Install.tsx` (the command, then what it installs today: a binary whose only command is `snapback version`; checksums verified by the script), `Status.tsx` (stages 0-7 from a `stages` array in `content.ts` whose names are SPEC §22's; stage 0 `done`, stage 1 `in progress`, 2-7 `planned`), `Footer.tsx` (docs `/docs/`, GitHub repo, licence). `App.tsx` renders them after HowItWorks in that order. Adds the copy lint tests.

Files: `web/src/sections/{Limits,Install,Status,Footer}.tsx`, `web/src/content.ts` (append), `web/src/App.tsx`, `web/src/copy.test.tsx`, `web/src/status.test.ts`.

## T6 — SEO, Open Graph and favicon meta

`web/index.html` head: `<html lang="en">`, `<title>snapback: Time Machine-style restore for Restic</title>`, `meta description` (sentence case, lowercase name, no `!`), `link rel="canonical" href="https://snapback.run/"`, `meta name="color-scheme" content="light dark"`, two `theme-color` metas (canvas light and dark values with `media`), `og:title`, `og:description`, `og:url` = `https://snapback.run/`, `og:type` = `website`, `og:image` = `https://snapback.run/og-card-light.png` with `og:image:width`/`height`, `twitter:card` = `summary_large_image`, `twitter:image`, `link rel="icon" type="image/svg+xml" href="/favicon.svg"`, PNG icons 16 and 32, `apple-touch-icon` = `/app-icon-180.png`. No external origins anywhere in the file.

Files: `web/index.html`, `test/site/meta_test.go`.

## T7 — Pages restructure: React at `/`, MkDocs at `/docs/`, install.sh at `/install.sh`

`docs.yml` build job: checkout; `actions/setup-node@v4` with `node-version-file: web/.nvmrc`, `cache: npm`, `cache-dependency-path: web/package-lock.json`; `npm ci`, `npm run build` with `working-directory: web`; existing Python setup and `mkdocs build --strict --site-dir site` unchanged; one assemble step: `mkdir -p _site/docs && cp -R web/dist/. _site/ && cp -R site/. _site/docs/ && cp install.sh _site/install.sh`; `upload-pages-artifact` `path: _site`. Deploy job unchanged. `mkdocs.yml` `site_url: https://snapback.run/docs/`. README Documentation section links `https://snapback.run/docs/` (and may link `https://snapback.run/` as the project site). Update the C1 pins in the same task: `test/docs/mkdocs_config_test.go` site_url, `test/docs/workflow_test.go` (install copy now targets `_site/install.sh`, upload path `_site`, new site-build and assemble checks), `test/projectdocs/readme_links_test.go` docsSiteURL.

Files: `.github/workflows/docs.yml`, `mkdocs.yml`, `README.md`, `test/docs/mkdocs_config_test.go`, `test/docs/workflow_test.go`, `test/projectdocs/readme_links_test.go`.

## T8 — CI job that builds and tests the site

`ci.yml` gains job `site` (runs-on `ubuntu-latest`, like the existing jobs): checkout, `actions/setup-node@v4` with `node-version-file: web/.nvmrc` and npm cache on `web/package-lock.json`, then with `working-directory: web`: `npm ci`, `npm run lint`, `npm test`, `npm run build`, and a step that fails if `web/dist` references a third-party origin (`! grep -rEl 'https?://(fonts\.googleapis|fonts\.gstatic|cdn\.|unpkg|jsdelivr)' dist`). Existing jobs untouched.

Files: `.github/workflows/ci.yml`, `test/ci/site_job_test.go`.

## Order and parallelism

| Wave | Tasks | Depends on | Why parallel is safe |
| --- | --- | --- | --- |
| 0 | pre-flight | C1 merged green | C3 edits C1's pins |
| 1 | T1 | — | creates `web/`, helpers and lockfile every later task needs |
| 2 | T2, T3, T4, T6, T7, T8 | T1 | disjoint files, except `NOTICE` (T3 only in this wave) and `web/src/main.tsx` (T2 and T4 each add one import line; merge T2 first, T4 rebases) |
| 3 | T5 | T4 | appends to `content.ts` and `App.tsx`; its copy lint must see every section |
| 4 | final sign-off | all | full matrix plus Node gates on the merged tree |

Merge order inside wave 2: T2, T3, T6, T8, T7, T4. T7 runs the site build in `docs.yml` but its tests read YAML only, so it does not wait for T4.
