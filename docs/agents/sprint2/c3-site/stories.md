---
type: stories
sprint: 2
---

# C3 — snapback.run landing site in the snapback design system

Story id: **C3** (schema id `S2-12` in the tasks/plan/validate front matter, because the planning schema requires `S<n>-<n>`; C1 and C2 are taken as S2-10 and S2-11). Source: `intake.md` (human request 2026-09-22T03:24:37Z). Depends on C1 merged green into the working branch (snapback.run domain, `docs.yml` copies `install.sh`, C1 tests pin `site_url` and the README docs link). C3 moves MkDocs to `/docs/` and updates those C1 pins in the same task as the files they pin.

## Sprint goal

Add the snapback.run landing page, built on the snapback design system, to Sprint 2 without disturbing the Stage 1 work: React at `/`, the MkDocs docs at `/docs/`, `install.sh` at `/install.sh`, all built and tested in CI.

## Sprint demo

`cd web && npm ci && npm run build && npx vite preview` shows the page in light and dark (toggle the OS setting); `go test ./test/site/...` and `npm test` pass; the docs workflow's `_site/` has `index.html`, `docs/index.html` and `install.sh`.

## Definition of Done

Every acceptance item below is covered by a test in `plan.md` that passes; the standards matrix plus the Node gates in `validate.md` are green; no suppressions; no AI trailers in commits.

## Out of scope

See Non-goals at the end of this file.

## Story dependency graph

C1 (merged) -> C3. C3 is independent of S2-01..S2-09 (no shared files).

## User stories

### Story C3

As a skeptical, technical user who already trusts Restic, I want https://snapback.run/ to tell me in one page what snapback is, the one command to try, what it does not do and exactly how far it has got, so I can judge it without being sold to.

### Owned files

`web/**` (new React site source), `test/site/**` (new Go tests), `.github/workflows/docs.yml`, `.github/workflows/ci.yml` (one new job only), `mkdocs.yml` (`site_url` only), `README.md` (Documentation links only), `NOTICE` (new font and npm entries), `.gitignore` (`web/node_modules/`, `_site/`), and the C1-pinned tests `test/docs/mkdocs_config_test.go`, `test/docs/workflow_test.go`, `test/projectdocs/readme_links_test.go`. Nothing under `cmd/`, `internal/`, `tools/`, `docs-site/`, `install.sh` or `Makefile`.

### Decisions (fixed here; workers do not re-decide)

- **Framework: Vite + React + TypeScript, single static page.** Why: the site is one page with no routing, data or server; Vite emits plain static files for GitHub Pages with no export adapter, fewer dependencies and no framework conventions that fight the `/docs/` subpath (Next.js static export adds `next/image` and `basePath` workarounds for no gain here).
- **No Tailwind.** Why: the design system is already a token set of CSS variables; one generated `tokens.css` plus plain CSS keeps a single vocabulary and one fewer pinned dependency.
- **Directory: `web/`, not `site/`.** Why: the standards matrix runs `mkdocs build --strict --site-dir site`, which cleans `site/`; a React source tree there would be deleted by every gate run.
- **Tests: Go tests under `test/site/` plus vitest under `web/`.** Why: Go tests join the existing `go test -race ./...` matrix and pin repo-level contracts (token contrast, files, workflows, meta) with no Node needed; vitest is the only way to render the React tree and lint the words a visitor actually reads. Rendering uses `react-dom/server` `renderToStaticMarkup` in the Node environment, so no jsdom or browser dependency.
- **Lint for `web/`: `tsc --noEmit`.** Why: strict type checking catches the real defects in a copy-heavy page; ESLint would add a plugin tree for little signal. Revisit if the site grows.
- **Node pinned exactly in `web/.nvmrc`** to the locally installed `26.0.0`, read by CI through `actions/setup-node` `node-version-file`. Why: one version for local gates and CI; `latest` and ranges are banned by standards.md. Every npm dependency exact (`.npmrc` `save-exact=true`), `package-lock.json` committed.
- **Product name is lowercase in every visible string, including headings and meta.** Why: design-system README ("Lowercase product name everywhere"); a blanket ban on `Snapback` is simpler to test than "mid-sentence" and stricter.
- **Links and body copy use `accent-text` / `ink` / `ink-body` / `ink-muted`, never `blue` or `yellow-text` as text on `canvas`.** Why: computed from tokens.json, light `blue` on `canvas` is 4.27:1 and `yellow-text` on `canvas` is 4.41:1, both below AA.
- **No icon library.** Why: the page needs only the brand marks; Lucide would be one more dependency and a route to banned glyphs.

### Acceptance (testable)

1. `npm ci && npm run lint && npm test && npm run build` in `web/` succeed with the pinned Node; `web/dist/index.html` exists.
2. `web/src/styles/tokens.css` is exactly the output of `web/scripts/gen-tokens.mjs` over `web/tokens.json` (a byte copy of the design system's `tokens.json`): every colour token appears as `--<name>` for light in `:root` and for dark under `prefers-color-scheme: dark`; spacing, radius, shadow and `@font-face` rules come from the same file.
3. Contrast computed from `web/tokens.json` is at least 4.5:1 for every text pair the page uses and at least 3:1 for `line-strong` on `canvas`, in both themes.
4. The rendered page has, in order: hero (pitch plus one command), how it works (`.snapshot` in every directory, `cp` to restore), what it does not do, install, status, footer (docs and GitHub links).
5. The status section lists SPEC §22 stages 0-7 by their SPEC names; only stage 0 is marked done and stage 1 in progress; every later stage is planned.
6. Rendered text contains no emoji, no `!`, no `Snapback`, no banned honesty word, no other backend name, and no present-tense claim that the `.snapshot` view works today; it contains an explicit "not yet" statement for the `.snapshot` view.
7. Fonts are self-hosted from `web/public/fonts/`; `web/index.html` and `web/dist/` reference no third-party origin for fonts or scripts.
8. `web/index.html` carries title, description, canonical `https://snapback.run/`, Open Graph and Twitter card tags pointing at `og-card-light.png`, `favicon.svg`, PNG favicons, `apple-touch-icon` and a `color-scheme` meta.
9. The Pages artifact is `_site/` with the React build at `/`, MkDocs at `/docs/`, and `install.sh` at `/install.sh`; `mkdocs.yml` `site_url` is `https://snapback.run/docs/`; README's Documentation section links `https://snapback.run/docs/`.
10. `ci.yml` has a `site` job that runs `npm ci`, lint, test and build in `web/` with the pinned Node.

### Failure scenario

A worker writes hero copy "Browse every Snapback snapshot in Finder today!" with a padlock icon. The build still succeeds, but `web/src/copy.test.tsx::TestVoiceLint` fails on `!` and `Snapback`, `::TestHonesty` fails on "today" next to a feature SPEC §22 places in stage 2+ and on "Finder" (stage 5), and `test/site/icons_test.go::TestNoBannedIconNames` fails on the `Lock`/padlock identifier. The site CI job goes red and deploy never runs, so the claim never reaches snapback.run.

### Non-goals

No docs content changes, no blog, no analytics, no cookie banner, no web UI for the product, no Next.js, no runtime CDN, no new claims beyond SPEC §20 differentiators framed as planned.
