---
type: plan-ready
story: S2-12
from_red_at: 2026-09-22T03:40:16Z
---

# S2-12 plan-ready (RED verified by orchestrator at chain @ eb2da75; every box currently FAILs by assertion unless noted)


# C3 test plan (schema id S2-12, tests only)

Contracts under test: `tasks.md`. Vitest test names below are the `it()` titles in camelCase (the worker uses them verbatim as titles). Tooling: Go tests in `test/site/` (package `site_test`, stdlib only: `encoding/json`, `encoding/xml`, `image/png`, `regexp`) for repo-level contracts; vitest in `web/` (Node environment, `react-dom/server` `renderToStaticMarkup`, `node:fs`) for rendered copy and the token generator. Why both: Go tests run in the existing `go test -race ./...` gate with no Node; only vitest can render React. Every test is offline and deterministic: no network, no browser, no clock, no scratchpad paths. Negative checks first assert the expected content is non-empty (M-002). "Rendered text" means `renderToStaticMarkup(<App/>)` with tags stripped, entities decoded and whitespace collapsed.

## T1 — Scaffold

- [x] `test/site/scaffold_test.go::TestPackageJSONPinsExactVersions` — input `web/package.json`; action decode `dependencies` and `devDependencies`; assert both maps non-empty and every value matches `^\d+\.\d+\.\d+$` (no `^`, `~`, `*`, `latest`, ranges, URLs).
- [x] `test/site/scaffold_test.go::TestPackageJSONAllowedDependencies` — assert `dependencies` keys are exactly {react, react-dom} and devDependencies keys are a subset of {vite, @vitejs/plugin-react, typescript, vitest, @types/react, @types/react-dom, @types/node}; `tailwindcss`, `next`, `lucide-react` absent.
- [x] `test/site/scaffold_test.go::TestPackageJSONScripts` — assert `scripts.build == "vite build"`, `scripts.test == "vitest run"`, `scripts.lint == "tsc --noEmit"`, `scripts.prebuild` contains `gen:tokens`.
- [x] `test/site/scaffold_test.go::TestNodePinned` — read `web/.nvmrc`; assert trimmed content matches `^\d+\.\d+\.\d+$` and equals `26.0.0`; assert `engines.node` in package.json equals it; assert `web/.npmrc` contains lines `save-exact=true` and `engine-strict=true`.
- [x] `test/site/scaffold_test.go::TestLockfileCommittedAndConsistent` — decode `web/package-lock.json`; assert `lockfileVersion >= 3` and `packages[""].dependencies` equals package.json `dependencies`.
- [x] `test/site/scaffold_test.go::TestGitignoreCoversSiteOutputs` — assert `.gitignore` has lines `web/node_modules/` and `_site/`, and `dist/` still present.
- [x] `test/site/scaffold_test.go::TestNoticeListsRuntimeDeps` — assert `NOTICE` mentions `react` and `react-dom` with `MIT`.
- [x] `web/src/App.test.tsx::rendersAMainLandmarkWithTheLowercaseName` — render `<App/>`; assert markup contains `<main` and an `<h1>` whose text contains `snapback`.

## T2 — Tokens

- [x] `web/scripts/gen-tokens.test.ts::emitsEveryColourTokenForBothThemes` — load `web/tokens.json`; `generate(tokens)`; split at `@media (prefers-color-scheme: dark)`; for each of the 30 colour tokens assert the light block has `--<name>: <light>;` and the dark block has `--<name>: <dark>;` (exact strings; count asserted as 30 first).
- [x] `web/scripts/gen-tokens.test.ts::emitsSpacingRadiusAndShadowTokens` — assert every `spacing.tokens[].name`, `radius.tokens[].name` appears as `--<name>: <value>;` and each shadow token's light value in `:root`, dark value in the dark block.
- [x] `web/scripts/gen-tokens.test.ts::emitsOneFontFacePerFontFile` — assert the count of `@font-face` equals `type.fonts.length` (asserted > 0) and each has `url('/fonts/<basename>') format('woff2')`, its `font-weight` and `font-display: swap`.
- [x] `web/scripts/gen-tokens.test.ts::declaresColorSchemeLightDark` — assert `:root` block contains `color-scheme: light dark;`.
- [x] `web/scripts/gen-tokens.test.ts::committedTokensCssMatchesTheGenerator` — read `web/src/styles/tokens.css`; assert it equals `generate(tokens)` byte for byte (drift guard: no hand edits).
- [x] `web/scripts/gen-tokens.test.ts::isDeterministic` — call `generate` twice on a deep clone; assert identical output and that the input object is unchanged.
- [x] `test/site/contrast_test.go::TestTextContrastAA` — input `web/tokens.json`; action WCAG 2.x relative luminance and ratio per theme; table (fg, bg) = (ink, canvas), (ink, surface), (ink-body, canvas), (ink-body, surface), (ink-body, canvas-sunken), (ink-muted, canvas), (ink-muted, surface), (accent-text, canvas), (accent-text, surface), (on-accent, accent); assert ratio >= 4.5 in `light` and `dark`, error message names pair, theme and ratio to two decimals.
- [x] `test/site/contrast_test.go::TestNonTextContrast` — assert (line-strong, canvas) >= 3.0 in both themes.
- [x] `test/site/contrast_test.go::TestContrastRatioKnownValues` — unit check of the helper: #000000 on #ffffff = 21.00, #ffffff on #ffffff = 1.00, #767676 on #ffffff >= 4.54 and < 4.55.
- [x] `test/site/contrast_test.go::TestTokenFileShape` — assert `name == "snapback"`, themes are exactly [light, dark], and every colour token referenced by the tables above exists with a `#rrggbb` value per theme (so a renamed token fails loudly instead of skipping).

## T3 — Fonts and assets

- [x] `test/site/assets_test.go::TestFontsSelfHosted` — for each `type.fonts[].file` in `web/tokens.json` (count > 0), assert `web/public/fonts/<basename>` exists and its first four bytes are `wOF2`.
- [x] `test/site/assets_test.go::TestBrandSVGs` — table of the six `web/public/brand/*.svg` names plus `web/public/favicon.svg`; assert each parses with `encoding/xml` and its root element is `svg`, and contains no `<linearGradient` or `<radialGradient` (the mark is flat; gradients are app-icon only).
- [x] `test/site/assets_test.go::TestBrandPNGs` — table favicon-16 (16x16), favicon-32 (32x32), app-icon-180 (180x180), app-icon-512 (512x512); `png.DecodeConfig`; assert exact width and height.
- [x] `test/site/assets_test.go::TestOGCardsAreWide` — `og-card-light.png`, `og-card-dark.png`; assert width >= 1200 and width/height within 0.02 of 1.91.
- [x] `test/site/assets_test.go::TestNoticeListsFonts` — assert `NOTICE` contains `IBM Plex Sans`, `IBM Plex Mono`, `JetBrains Mono` and `SIL Open Font License`.
- [x] `test/site/icons_test.go::TestNoBannedIconNames` — walk `web/src` (`.ts`, `.tsx`, `.css`; assert >= 1 file) and `web/public` file names; assert no match of `(?i)\b(hat|cap|shield|lock|padlock|cloud|hard-?drive|harddrive|clock|refresh|rotate-?c?c?w|history-icon|camera|shutter|aperture)\b` in identifiers, import specifiers, class names or file names; `lucide` absent.
- [x] `test/site/icons_test.go::TestNoRuntimeCDNInSource` — scan `web/index.html`, `web/src/**`; assert no `fonts.googleapis`, `fonts.gstatic`, `unpkg`, `jsdelivr`, `cdnjs`, `<script src="http`.

## T4 — Content model, header, hero, how it works

- [x] `web/src/sections.test.tsx::heroShowsTheOneCommand` — render `<Hero/>`; assert text contains exactly one occurrence of `curl -fsSL https://snapback.run/install.sh | sh` inside a `<code>` or `<pre>`.
- [x] `web/src/sections.test.tsx::heroStatesWhatExistsToday` — assert hero text matches `/stage 0/i` and `/not yet/i`.
- [x] `web/src/sections.test.tsx::howItWorksShowsSnapshotAndCp` — render `<HowItWorks/>`; assert text contains `.snapshot`, a line starting `cp .snapshot/`, and `/planned/i`.
- [x] `web/src/sections.test.tsx::headerLinksDocsAndGitHub` — render `<Header/>`; assert an `<a href="/docs/">` and an `<a href="https://github.com/adeelahmad/snapback">`, and an `<img>` whose `src` is `/brand/lockup-horizontal.svg` with non-empty `alt` equal to `snapback`.
- [x] `web/src/sections.test.tsx::stylesheetUsesTokensOnly` — read `web/src/styles/site.css` (non-empty); assert no `#[0-9a-fA-F]{3,8}\b`, no `rgb(`/`hsl(` literals, no `linear-gradient`; assert it contains `radial-gradient` using `var(--glass-tint-` and `var(--shadow-glass)`.
- [x] `web/src/sections.test.tsx::stylesheetSetsNoBlueOrYellowText` — assert no `color:\s*var\(--(blue|blue-alt|yellow|yellow-text|red)\)` in site.css.

## T5 — Remaining sections and copy lint

- [ ] `web/src/copy.test.tsx::TestSectionOrder` — render `<App/>`; collect `<section id>` values; assert exactly `["hero","how-it-works","limits","install","status"]` followed by a `<footer>`.
- [ ] `web/src/copy.test.tsx::TestVoiceLint` — rendered text (asserted > 200 chars) plus `content.ts` string values; assert no match of `\p{Extended_Pictographic}` (u flag), no `!`, no `Snapback`, no `snap back`, no Title Case heading (each `<h1>`-`<h3>` text: no word after the first starts uppercase except `Restic`, `Time`, `Machine`, `GitHub`, `Linux`, `macOS`, `FUSE`, `SPEC`).
- [ ] `web/src/copy.test.tsx::TestHonesty` — rendered text; assert no case-insensitive match of `production-ready`, `cross-platform`, `static`, `finder-integrated`, `available now`, `works today`, `now supports`, `first-of-its-kind`, `borg`, `kopia`, `duplicity`, `duplicati`, `tarsnap`, `rustic`, `multi-backend`; assert no sentence containing `.snapshot` also contains `today` or `now` unless it contains `not yet`; assert text contains `not yet` and `Restic`.
- [ ] `web/src/copy.test.tsx::TestLimitsSection` — render `<Limits/>`; assert text mentions `schedul`, `retention`, `Windows`, and `never writes` near `Restic repository`.
- [ ] `web/src/copy.test.tsx::TestInstallSection` — render `<Install/>`; assert the command appears once, and text contains `snapback version` and `checksum`.
- [ ] `web/src/copy.test.tsx::TestFooterLinks` — render `<Footer/>`; assert hrefs `/docs/` and `https://github.com/adeelahmad/snapback`; every `<a>` with `http` href has `rel` containing `noopener`.
- [ ] `web/src/status.test.ts::stageNamesMatchSPECSection22` — read `../SPEC.md` via `node:fs`; parse the §22 table rows `| <n>. <name> |`; assert 8 rows parsed and `stages.map(s => s.name)` equals them in order.
- [ ] `web/src/status.test.ts::onlyStage0IsDone` — assert `stages[0].state === "done"`, `stages[1].state === "in progress"`, and `stages.slice(2).every(s => s.state === "planned")`.

## T6 — Meta

- [ ] `test/site/meta_test.go::TestHeadMeta` — parse `web/index.html` head with regexp over `<meta|<link|<title|<html`; table of required (selector, value): `html[lang]=en`; `title` starts `snapback`; `link[rel=canonical]=https://snapback.run/`; `meta[name=color-scheme]=light dark`; `og:url=https://snapback.run/`; `og:type=website`; `og:image=https://snapback.run/og-card-light.png`; `twitter:card=summary_large_image`; `link[rel=icon][type=image/svg+xml]=/favicon.svg`; `link[rel=apple-touch-icon]=/app-icon-180.png`; assert each present with the exact value.
- [ ] `test/site/meta_test.go::TestThemeColorPerScheme` — assert two `theme-color` metas with `media` `(prefers-color-scheme: light)` / `dark` whose `content` equals the canvas light and dark values read from `web/tokens.json`.
- [ ] `test/site/meta_test.go::TestMetaVoice` — collect title, description, og:title, og:description (all non-empty); assert no `!`, no `Snapback`, no emoji (rune ranges U+1F000-U+1FAFF, U+2600-U+27BF), description length 50-160.
- [ ] `test/site/meta_test.go::TestMetaIconsExist` — for every local `href`/`content` path starting `/` in the head, assert `web/public/<path>` exists.
- [ ] `test/site/meta_test.go::TestNoExternalOrigins` — assert no `src=` or `href=` with `http` other than the canonical, og and twitter values on `https://snapback.run/`.

## T7 — Pages restructure

- [ ] `test/docs/mkdocs_config_test.go::TestSiteURLAndName` — updated: `site_url` = `https://snapback.run/docs/`.
- [ ] `test/docs/workflow_test.go::TestWorkflowBuildsSite` — build job has a `setup-node` step pinned `@v<major>` (not `latest`/`main`) with `node-version-file: web/.nvmrc`, and `npm ci` then `npm run build` steps with `working-directory: web`, both before the assemble step.
- [ ] `test/docs/workflow_test.go::TestWorkflowAssemblesPagesTree` — an assemble step after both `npm run build` and `mkdocs build` copies `web/dist/.` to `_site/`, `site/.` to `_site/docs/`, and `install.sh` to `_site/install.sh`; it precedes `upload-pages-artifact`.
- [ ] `test/docs/workflow_test.go::TestWorkflowUploadsPagesArtifact` — updated: `path: _site`.
- [ ] `test/docs/workflow_test.go::TestWorkflowPublishesInstallScript` — updated (C1): target is `_site/install.sh`.
- [ ] `test/docs/workflow_test.go::TestWorkflowBuildsStrict` — unchanged: `mkdocs build --strict --site-dir site` still present (the gate matrix command stays valid).
- [ ] `test/projectdocs/readme_links_test.go::TestReadmeLinksDocs` — updated: `docsSiteURL = "https://snapback.run/docs/"`; README still has no `adeelahmad.github.io/snapback`.

## T8 — CI job

- [x] `test/ci/site_job_test.go::TestCISiteJobExists` — decode `ci.yml`; assert job `site` exists and existing jobs `test`, `cross-compile`, `fuse-linux` are still present.
- [x] `test/ci/site_job_test.go::TestCISiteJobPinsNode` — `actions/setup-node@v<major>`, `node-version-file: web/.nvmrc`, `cache-dependency-path: web/package-lock.json`; no `node-version: latest`, `lts/*` or `current`.
- [x] `test/ci/site_job_test.go::TestCISiteJobSteps` — with `working-directory: web` (step or `defaults.run`), assert run steps in order: `npm ci`, `npm run lint`, `npm test`, `npm run build`, then a step whose run contains `grep` over `dist` for `fonts.googleapis`.
- [x] `test/ci/site_job_test.go::TestCISiteJobNoInstallFallback` — assert no step runs `npm install` (only `npm ci`, so the lockfile is authoritative).
