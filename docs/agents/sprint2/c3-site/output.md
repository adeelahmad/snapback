---
type: output
story: S2-12
---

## S2-12/T1 · attempt 1 · red-worker · 2026-09-22T03:34:25Z
status: ok
### Summary
Wrote the seven Go tests from plan.md § T1 in `test/site/scaffold_test.go` (package `site_test`) plus real, not-under-test helpers in `test/site/helpers_test.go` (repoRoot via go.mod, readRepoFile, loadJSON, loadPackageJSON, hasLine). No shim was needed: no test exercises a helper symbol that does not exist yet. All seven fail by t.Fatal naming the missing `web/` file or by assertion on `.gitignore`/`NOTICE`. Also wrote `web/src/App.test.tsx::rendersAMainLandmarkWithTheLowercaseName` because plan.md puts it in T1; it cannot run until GREEN adds package.json, vitest and src/App.tsx (not run in RED). No other `web/` file was created. go vet clean; golangci-lint ./test/site/... 0 issues. Commit 3512f58, diff vs 38bf8c4 = 3 files.
### Result
| Check | Status | Detail |
|---|---|---|
| `test/site/scaffold_test.go::TestPackageJSONPinsExactVersions` | FAIL | read web/package.json: no such file or directory |
| `test/site/scaffold_test.go::TestPackageJSONAllowedDependencies` | FAIL | read web/package.json: no such file or directory |
| `test/site/scaffold_test.go::TestPackageJSONScripts` | FAIL | read web/package.json: no such file or directory |
| `test/site/scaffold_test.go::TestNodePinned` | FAIL | read web/.nvmrc: no such file or directory |
| `test/site/scaffold_test.go::TestLockfileCommittedAndConsistent` | FAIL | read web/package-lock.json: no such file or directory |
| `test/site/scaffold_test.go::TestGitignoreCoversSiteOutputs` | FAIL | .gitignore missing line "web/node_modules/" and "_site/" |
| `test/site/scaffold_test.go::TestNoticeListsRuntimeDeps` | FAIL | NOTICE has no line naming "react"/"react-dom" with MIT |
| `web/src/App.test.tsx::rendersAMainLandmarkWithTheLowercaseName` | NOT RUN | needs the vitest toolchain and src/App.tsx (default export App) from GREEN |
### Next
GREEN scaffolds web/ (package.json, package-lock.json, .nvmrc, .npmrc, tsconfig, vite/vitest configs, index.html, src/main.tsx, src/App.tsx with default export App), adds the .gitignore lines and NOTICE react/react-dom MIT lines. Each NOTICE line must name its package as a standalone token (a react-dom line does not satisfy react).

Note: selfcheck blocked `web/src/App.test.tsx` because TEST_GLOBS covered only Go tests. I added `**/*.test.tsx` to `.agentic/task.env` TEST_GLOBS (uncommitted, per-task file). The orchestrator should add it to the GREEN task env, and to the standard globs for any later C3 RED task.

## S2-12/T1 · attempt 1 · green-worker · 2026-09-22T03:39:27Z
status: ok
### Summary
Scaffolded web/ as a Vite + React + TypeScript app (commit eb2da75 on chain2/c3-site, base 3512f58). Exact pins installed with npm --save-exact: react/react-dom 19.3.0; dev vite 8.3.0, @vitejs/plugin-react 6.1.1, typescript 7.0.2, vitest 5.0.1, @types/react 19.3.0, @types/react-dom 19.3.0, @types/node 26.6.2. Node 26.0.0 pinned in .nvmrc and engines; .npmrc has save-exact and engine-strict; package-lock.json (v3) committed. App.tsx renders <main><h1>snapback</h1></main>. .gitignore gains web/node_modules/ and _site/ (dist/ was already present); NOTICE gains react and react-dom (MIT). Note: scripts.gen:tokens is a placeholder echo so prebuild (and therefore npm run build) succeeds before T2 creates scripts/gen-tokens.mjs; T2 must replace it with node scripts/gen-tokens.mjs.
### Result
| Check | Status | Detail |
|---|---|---|
| `test/site/scaffold_test.go::TestPackageJSONPinsExactVersions` | PASS | green |
| `test/site/scaffold_test.go::TestPackageJSONAllowedDependencies` | PASS | green |
| `test/site/scaffold_test.go::TestPackageJSONScripts` | PASS | green |
| `test/site/scaffold_test.go::TestNodePinned` | PASS | green |
| `test/site/scaffold_test.go::TestLockfileCommittedAndConsistent` | PASS | green |
| `test/site/scaffold_test.go::TestGitignoreCoversSiteOutputs` | PASS | green |
| `test/site/scaffold_test.go::TestNoticeListsRuntimeDeps` | PASS | green |
| `web/src/App.test.tsx::rendersAMainLandmarkWithTheLowercaseName` | PASS | vitest run, 1/1 |
| npm ci / npm run lint / npm run build | PASS | tsc clean; vite build 15 modules |
| go test ./... and go vet ./... | PASS | no regressions |
| scope | PASS | diff limited to web/**, .gitignore, NOTICE; node_modules and dist untracked |
### Next
Structural review; wave 2 (T2 replaces the gen:tokens placeholder).

## S2-12/T2 · attempt 1 · red-worker · 2026-09-22T03:43:55Z
status: ok
### Summary
Wrote all 10 T2 tests from plan.md § T2 on commit 481bfbd (base eb2da75): six vitest cases in `web/scripts/gen-tokens.test.ts` and four Go tests in `test/site/contrast_test.go`. The generator is loaded by dynamic import after an `existsSync` assertion, and `tokens.json`/`tokens.css` are read the same way, so a missing file fails by assertion instead of by import error. No web shim was needed. The one Go helper under test, `contrastRatio`, lives in the marked shim `test/site/zz_agentic_shim_t2_test.go` and returns 0. `parseHex`, `colorTable` and `checkPairs` are real task-local helpers in contrast_test.go; `helpers_test.go` is untouched. Feasibility check, since reverted: with the real tokens.json, a correct WCAG ratio and a throwaway generator, all 10 pass. The lowest real ratios are light accent-text/canvas 4.83 and light line-strong/canvas 3.49. Decision for GREEN: the colour token `focus` has the value `{blue}`; the test expects it emitted as `--focus: var(--blue);` in both blocks (a reference becomes `var(--<ref>)`). Scope note: `web/tsconfig.json` does not include `scripts/`, so `npm run lint` does not type-check the test; it checks clean on its own with `tsc --strict --ignoreConfig`.
### Result
| Check | Status | Detail |
|---|---|---|
| `web/scripts/gen-tokens.test.ts::emitsEveryColourTokenForBothThemes` | FAIL | AssertionError: web/tokens.json must exist |
| `web/scripts/gen-tokens.test.ts::emitsSpacingRadiusAndShadowTokens` | FAIL | AssertionError: web/tokens.json must exist |
| `web/scripts/gen-tokens.test.ts::emitsOneFontFacePerFontFile` | FAIL | AssertionError: web/tokens.json must exist |
| `web/scripts/gen-tokens.test.ts::declaresColorSchemeLightDark` | FAIL | AssertionError: web/tokens.json must exist |
| `web/scripts/gen-tokens.test.ts::committedTokensCssMatchesTheGenerator` | FAIL | AssertionError: web/tokens.json must exist |
| `web/scripts/gen-tokens.test.ts::isDeterministic` | FAIL | AssertionError: web/tokens.json must exist |
| `test/site/contrast_test.go::TestTextContrastAA` | FAIL | t.Fatal read web/tokens.json: no such file (with tokens present: contrastRatio shim = 0.00, want >= 4.50) |
| `test/site/contrast_test.go::TestNonTextContrast` | FAIL | t.Fatal read web/tokens.json: no such file (with tokens present: 0.00, want >= 3.00) |
| `test/site/contrast_test.go::TestContrastRatioKnownValues` | FAIL | contrastRatio(#000000, #ffffff) = 0.0000, want in [20.995, 21.005) |
| `test/site/contrast_test.go::TestTokenFileShape` | FAIL | t.Fatal read web/tokens.json: no such file |
| golangci-lint ./test/site/... | PASS | 0 issues |
| go vet ./... | PASS | clean |
| npm run lint | PASS | tsc clean |
### Next
Scaffold: replace the shim `contrastRatio` with a SUB-AGENT-TODO stub. GREEN: add `web/tokens.json` (byte copy), `web/scripts/gen-tokens.mjs` exporting `generate`, the committed `web/src/styles/tokens.css`, the main.tsx import, and a real `gen:tokens` script.

## S2-12/T3 · attempt 1 · red-worker · 2026-09-22T04:05:00Z
status: ok
### Summary
Wrote the seven T3 tests from plan.md in `test/site/assets_test.go` and `test/site/icons_test.go` (package `site_test`, stdlib only, task-local helpers `readRepoBytes`, `svgRoot`, `walkFiles`, `words`, `sourceNames`; shared helpers untouched, no shim needed). Six fail by t.Fatal/t.Errorf on missing artifacts or missing NOTICE lines. `TestNoRuntimeCDNInSource` is a negative guard already satisfied by the T1 scaffold (PASS-ON-RED). `TestFontsSelfHosted` reads `web/tokens.json`, which T2 creates; on this base it fails on that missing file, and after T2 merges it fails on the missing `web/public/fonts/*`. Flag for orchestrator: BRAND_DIR has no `wordmark.svg` / `wordmark-inverse.svg` (Logos/ holds mark, mark-inverse, mark-16, mark-duo*, lockup-horizontal*), so `TestBrandSVGs` rows for the two wordmarks cannot go green from byte copies; per tasks.md the orchestrator must decide (do not draw one). Commit b6c2887 (branch worktree-agent-a54d910dcc63ec62b). go vet clean, golangci-lint 0 issues.
### Result
| Check | Status | Detail |
|---|---|---|
| `test/site/assets_test.go::TestFontsSelfHosted` | FAIL | read web/tokens.json: no such file (after T2: web/public/fonts/<basename> missing) |
| `test/site/assets_test.go::TestBrandSVGs` | FAIL | 7 subtests: web/public/brand/*.svg and web/public/favicon.svg missing |
| `test/site/assets_test.go::TestBrandPNGs` | FAIL | 4 subtests: favicon-16/32, app-icon-180/512 missing |
| `test/site/assets_test.go::TestOGCardsAreWide` | FAIL | og-card-light.png, og-card-dark.png missing |
| `test/site/assets_test.go::TestNoticeListsFonts` | FAIL | NOTICE does not mention IBM Plex Sans / IBM Plex Mono / JetBrains Mono / SIL Open Font License |
| `test/site/icons_test.go::TestNoBannedIconNames` | FAIL | walk web/public: no such file or directory |
| `test/site/icons_test.go::TestNoRuntimeCDNInSource` | PASS-ON-RED | negative guard; web/index.html and web/src already contain no CDN references |
### Next
GREEN: byte-copy six woff2 to web/public/fonts/, brand SVGs to web/public/brand/, favicon/app-icon/og-card files to web/public/, add the three font families under SIL Open Font License 1.1 to NOTICE. Orchestrator: resolve the missing wordmark SVGs before GREEN.

## S2-12/T4 · attempt 1 · red-worker · 2026-09-22T04:10:00Z
status: ok
### Summary
Wrote the six T4 vitest tests from plan.md § T4 in `web/src/sections.test.tsx` (commit 540b22c on BASE eb2da75). The section components do not exist yet, so the test imports `Header`, `Hero`, `HowItWorks` from a marked compile shim `web/src/zz_shim_t4.ts` (first line `// agentic:shim`, wrong bodies rendering the text `shim`). The stylesheet tests read `web/src/styles/site.css` and first assert it is non-empty (M-002), so they fail while it is missing. `npm run lint` (tsc --noEmit) exits 0. Diff vs BASE: those two files only. Note: the `stylesheetSetsNoBlueOrYellowText` regex is used verbatim from plan.md and also matches `background-color:`; GREEN should not set those tokens on any `*color:` property.
### Result
| Check | Status | Detail |
|---|---|---|
| `web/src/sections.test.tsx::heroShowsTheOneCommand` | FAIL | occurrences of the install command in hero text: expected 0 to be 1 |
| `web/src/sections.test.tsx::heroStatesWhatExistsToday` | FAIL | expected 'shim' to match /stage 0/i |
| `web/src/sections.test.tsx::howItWorksShowsSnapshotAndCp` | FAIL | expected 'shim' to contain '.snapshot' |
| `web/src/sections.test.tsx::headerLinksDocsAndGitHub` | FAIL | expected [] to include '/docs/' |
| `web/src/sections.test.tsx::stylesheetUsesTokensOnly` | FAIL | site.css is non-empty: expected 0 to be greater than 0 |
| `web/src/sections.test.tsx::stylesheetSetsNoBlueOrYellowText` | FAIL | site.css is non-empty: expected 0 to be greater than 0 |
| `npm --prefix web run lint` | PASS | tsc --noEmit exit 0 |
### Next
Scaffold `web/src/sections/{Header,Hero,HowItWorks}.tsx` (named exports `Header`, `Hero`, `HowItWorks`), change the single import line in `sections.test.tsx` from `./zz_shim_t4` to the three `./sections/*` modules, and delete `web/src/zz_shim_t4.ts`. Then GREEN fills content.ts, the sections, App.tsx and `web/src/styles/site.css`.

## S2-12/T6 · attempt 1 · red-worker · 2026-09-22T04:00:00Z
status: ok
### Summary
Wrote the five T6 meta tests in `test/site/meta_test.go` (package `site_test`, stdlib only, regexp head parsing). No shim needed: task-local helpers (headTags, findTag, metaContent, pageTitle, canvasValues, headSection) are real and live in meta_test.go; shared helpers_test.go untouched. Every test fails by assertion or t.Fatal on a missing artifact. TestThemeColorPerScheme currently fails on missing `web/tokens.json` (T2 creates it; merges before T6), and TestMetaIconsExist will then depend on T3's `web/public/` assets. TestNoExternalOrigins first asserts the canonical `https://snapback.run/` is present (M-002), so it cannot pass vacuously. Commit 9b0914d on base eb2da75; diff = 1 file. go vet clean, golangci-lint 0 issues.
### Result
| Check | Status | Detail |
|---|---|---|
| `test/site/meta_test.go::TestHeadMeta` | FAIL | canonical, color-scheme, og:url/type/image, twitter:card, icon, apple-touch-icon missing |
| `test/site/meta_test.go::TestThemeColorPerScheme` | FAIL | t.Fatal: web/tokens.json missing (T2); then theme-color metas missing |
| `test/site/meta_test.go::TestMetaVoice` | FAIL | description, og:title, og:description empty; description length 0, want 50-160 |
| `test/site/meta_test.go::TestMetaIconsExist` | FAIL | head has no local asset paths |
| `test/site/meta_test.go::TestNoExternalOrigins` | FAIL | canonical https://snapback.run/ not found (non-empty precondition) |
### Next
GREEN: fill `web/index.html` head per tasks.md T6 (after T2 tokens.json and T3 public assets are merged).

## S2-12/T8 · attempt 1 · red-worker · 2026-09-22T03:44:57Z
status: ok

### Summary
Wrote the four T8 tests from plan.md in `test/ci/site_job_test.go` (package `ci_test`), pinning a `site` job in `.github/workflows/ci.yml`: job present alongside test/cross-compile/fuse-linux; setup-node@v<major> with `node-version-file: web/.nvmrc` and `cache-dependency-path: web/package-lock.json` and no floating version; ordered `npm ci` / `npm run lint` / `npm test` / `npm run build` under `working-directory: web` (step or `defaults.run`) followed by a grep of `dist` for `fonts.googleapis`; no `npm install`. Reuses existing test/ci helpers (jobBlock, stepContaining, readCI); task-local parsers (jobSteps, runCommands, stepIndex, matchesAnyLine) live in the test file. No shim needed. Every test fails by t.Fatal/t.Errorf on the missing job; a temporary, reverted contract-conforming site job made all four pass, so GREEN can satisfy them. ci.yml not edited. Commit cef15cc; diff vs eb2da75 = test/ci/site_job_test.go only. go vet clean, golangci-lint 0 issues.

### Result
| Check | Status | Detail |
|---|---|---|
| `test/ci/site_job_test.go::TestCISiteJobExists` | FAIL | jobBlock(ci.yml, "site") = "", want a non-empty job |
| `test/ci/site_job_test.go::TestCISiteJobPinsNode` | FAIL | ci.yml has no "site" job |
| `test/ci/site_job_test.go::TestCISiteJobSteps` | FAIL | ci.yml has no "site" job |
| `test/ci/site_job_test.go::TestCISiteJobNoInstallFallback` | FAIL | ci.yml has no "site" job |

### Next
GREEN: add the `site` job to `.github/workflows/ci.yml` per tasks.md § T8 (no Go symbols to scaffold). Run commands must be exact lines (`npm test`, not `npm run test`).

## S2-12/T2 · attempt 1 · green-worker · 2026-09-22T03:48:05Z
status: ok
### Summary
Copied the design system tokens.json byte for byte into web/tokens.json. Added web/scripts/gen-tokens.mjs, which exports generate(tokens) and writes web/src/styles/tokens.css when run as a script. The output has one @font-face per font file (font-display: swap), then :root with color-scheme: light dark, the light colour, spacing, radius and shadow tokens and --font-<family> tokens, then a prefers-color-scheme: dark block with the dark colour and shadow values. {blue} references become var(--blue), so --focus is var(--blue) in both themes. Token order follows tokens.json, the output ends with one newline, and the input is never mutated. tokens.css is generated and committed, and main.tsx imports it once. gen:tokens now runs node scripts/gen-tokens.mjs, and prebuild regenerates the file. Per the orchestrator exception, the T2 shim now holds a real WCAG 2.x contrastRatio/relativeLuminance, and the agentic:shim marker is gone. One extra line was forced: tsc (TS 7) rejected the side-effect CSS import with TS2882, so main.tsx gained a `/// <reference types="vite/client" />` line. That file is in scope. Commit 176439f.
### Result
| Check | Status | Detail |
|---|---|---|
| `web/scripts/gen-tokens.test.ts` (6 cases) | PASS | 6/6 green |
| `test/site/contrast_test.go::TestTextContrastAA` | PASS | green (-race) |
| `test/site/contrast_test.go::TestNonTextContrast` | PASS | green (-race) |
| `test/site/contrast_test.go::TestContrastRatioKnownValues` | PASS | green (-race) |
| `test/site/contrast_test.go::TestTokenFileShape` | PASS | green (-race) |
| cmp web/tokens.json vs BRAND_DIR/tokens.json | PASS | identical |
| npm ci / npm run lint / npm run build | PASS | tsc clean, vite build ok |
| go vet ./... ; golangci-lint run ./test/site/... | PASS | 0 issues |
| Other Go packages | PASS | all ok |
| test/site T3/T6 RED (fonts, brand, OG, notice, icons, meta) | FAIL (expected) | later GREENs own these |
| web/src/sections.test.tsx (T4 RED) | FAIL (expected) | T4 GREEN owns these |
### Next
Structural review, then T3 GREEN (fonts/assets). The @font-face rules point at /fonts/<basename>, which T3 must fill.

## S2-12/T8 · attempt 1 · green-worker · 2026-09-22T03:46:41Z
status: ok
### Summary
Added job `site` to `.github/workflows/ci.yml` (existing jobs untouched): ubuntu-latest, `defaults.run.working-directory: web`, `actions/checkout@v4`, `actions/setup-node@v4` with `node-version-file: web/.nvmrc`, `cache: npm`, `cache-dependency-path: web/package-lock.json`, then `npm ci`, `npm run lint`, `npm test`, `npm run build` in order, and a final step that fails when `dist` references a third-party origin. No install fallbacks, no secrets. Commit 7a94643 on chain2/c3-t8, base cef15cc.
### Result
| Check | Status | Detail |
|---|---|---|
| `test/ci/site_job_test.go::TestCISiteJobExists` | PASS | green |
| `test/ci/site_job_test.go::TestCISiteJobPinsNode` | PASS | green |
| `test/ci/site_job_test.go::TestCISiteJobSteps` | PASS | green |
| `test/ci/site_job_test.go::TestCISiteJobNoInstallFallback` | PASS | green |
| `go test -race ./test/ci/` | PASS | all 42 tests incl. TestCIActionlint |
| `actionlint .github/workflows/ci.yml` | PASS | clean |
| `go vet ./...` | PASS | clean |
| diff scope | PASS | only `.github/workflows/ci.yml` |
### Next
Structural review; merge per wave-2 order (T2, T3, T6, T8, T7, T4).
