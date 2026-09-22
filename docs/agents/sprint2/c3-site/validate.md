---
type: validate
story: S2-12
---

# C3 validate (schema id S2-12) — PASS/FAIL rubric

Run from the repo root `/Users/adeelahmad/work/snapback` unless a row says `web/`. A row PASSES only when the literal expected output matches; anything else is FAIL and cites the rule it breaks. T1-T7 GREENs run task-scoped gates with `GATE_RUN_MATRIX=0`; T5 (last wave) and final sign-off run the full matrix plus the Node gates (M-005).

## Pre-flight

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| C1 merged | `grep -c 'snapback.run' mkdocs.yml README.md` | both counts >= 1 | stories.md dependency graph (C1 before C3) |
| Go toolchain | `go version` | contains `go1.27.1` | standards.md toolchain pinning |
| Node toolchain | `node --version` | `v26.0.0` | stories.md decision (Node pinned in `web/.nvmrc`) |
| npm present | `npm --version` | exit 0 | Node gates below |
| Inputs present | `ls "$DS/design-system/project/fonts"/*.woff2 \| wc -l` | `6` | tasks.md fixed inputs |
| Scope | `git diff --name-only HEAD -- . ':!docs/agents'` | only paths listed in stories.md "Owned files" | stories.md owned files; M-012 |

## Node gates (join the standards matrix)

These lines are added to the fenced block under `## Cross-cutting gates` in `docs/agents/sprint2/standards.md` by the orchestrator (the standards file's owner) when T1 merges, after `mkdocs build --strict --site-dir site`. Each is a single command so `run_standards_matrix` runs it verbatim:

```bash
npm --prefix web ci
npm --prefix web run lint
npm --prefix web test
npm --prefix web run build
test "$(node --version)" = "v$(cat web/.nvmrc)"
! grep -rEl 'https?://(fonts\.googleapis|fonts\.gstatic|unpkg|cdn\.jsdelivr|cdnjs)' web/dist
```

Pinned versions: Node `26.0.0` (`web/.nvmrc`, `engines.node`, `engine-strict=true`); every npm package exact in `web/package.json` and resolved in the committed `web/package-lock.json` (`npm ci` fails on any drift); `actions/setup-node@v4` in CI. `go test -race ./...` already covers `test/site/`.

## T1 — Scaffold

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| Go scaffold tests | `go test -race ./test/site/ -run 'TestPackageJSON\|TestNodePinned\|TestLockfile\|TestGitignore\|TestNoticeListsRuntimeDeps'` | `ok` | standards.md pinning; CLAUDE.md (NOTICE per dependency) |
| Install from lock | `npm --prefix web ci` | exit 0 | standards.md pinning |
| Lint, test, build | `npm --prefix web run lint && npm --prefix web test && npm --prefix web run build` | exit 0; `web/dist/index.html` exists | stories.md acceptance 1 |

## T2 — Tokens

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| Token source unchanged | `cmp web/tokens.json "$DS/artifact-files/e75f5781-2fc6-4585-b3b2-227c107dbcbb/project/tokens.json"` | no output, exit 0 | intake Must (use the design system exactly) |
| Generator tests | `npm --prefix web test -- scripts/gen-tokens.test.ts` | all pass | stories.md acceptance 2 |
| Contrast | `go test -race ./test/site/ -run 'TestTextContrastAA\|TestNonTextContrast\|TestContrastRatio\|TestTokenFileShape'` | `ok` | WCAG 2.x 1.4.3 / 1.4.11; stories.md acceptance 3 |

## T3 — Fonts and assets

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| Fonts are byte copies | `for f in "$DS"/design-system/project/fonts/*.woff2; do cmp "$f" "web/public/fonts/$(basename "$f")" \|\| echo DIFF; done` | no output | intake Must (self-host design-system fonts) |
| Brand assets are byte copies | `for f in web/public/brand/* web/public/favicon* web/public/app-icon* web/public/og-card*; do cmp "$f" "$BRAND_DIR/$(basename "$f")" \|\| echo DIFF; done` | no output | intake Must (real SVG assets) |
| Asset and icon tests | `go test -race ./test/site/ -run 'TestFonts\|TestBrand\|TestOGCards\|TestNoticeListsFonts\|TestNoBannedIconNames\|TestNoRuntimeCDN'` | `ok` | intake Must not (banned icons, CDN fonts) |

## T4 — Content, header, hero, how it works

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| Section tests | `npm --prefix web test -- src/sections.test.tsx` | all pass | stories.md acceptance 4 |
| Lint | `npm --prefix web run lint` | exit 0 | Node gates |

## T5 — Remaining sections, copy lint (runs full matrix)

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| Copy and status tests | `npm --prefix web test -- src/copy.test.tsx src/status.test.ts` | all pass | SPEC wording rules, §20 claims, §22 stages; stories.md acceptance 5, 6 |
| Human read-through | orchestrator reads `renderToStaticMarkup` output once | no claim that the `.snapshot` view works today | intake Must (honest content) |
| Full matrix | standards.md matrix plus Node gates above | every line exit 0 | M-005 |

## T6 — Meta

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| Meta tests | `go test -race ./test/site/ -run 'TestHeadMeta\|TestThemeColorPerScheme\|TestMetaVoice\|TestMetaIconsExist\|TestNoExternalOrigins'` | `ok` | stories.md acceptance 8 |

## T7 — Pages restructure

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| Docs and README tests | `go test -race ./test/docs/... ./test/projectdocs/...` | `ok` for both | stories.md acceptance 9; C1 pins updated, not deleted |
| actionlint | `actionlint` | no output | standards.md gate |
| Local assemble dry run | `npm --prefix web run build && mkdocs build --strict --site-dir site && rm -rf _site && mkdir -p _site/docs && cp -R web/dist/. _site/ && cp -R site/. _site/docs/ && cp install.sh _site/install.sh && ls _site/index.html _site/docs/index.html _site/install.sh` | the three paths listed | stories.md acceptance 9 |
| MkDocs links under /docs/ | `grep -c 'https://snapback.run/docs/' site/sitemap.xml` | >= 1 | site_url moved to /docs/ |

## T8 — CI job

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| CI tests | `go test -race ./test/ci/...` | `ok` | stories.md acceptance 10 |
| actionlint | `actionlint` | no output | standards.md gate |

## Final sign-off

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| Full standards matrix | every fenced line in `docs/agents/sprint2/standards.md` `## Cross-cutting gates`, now including the Node gates | every line exit 0 | standards.md |
| Suppressions | `grep -rnE 'nolint\|@ts-ignore\|@ts-expect-error\|eslint-disable\|\.skip\(\|it\.only\|t\.Skip' web/src web/scripts test/site test/ci/site_job_test.go` | no output | standards.md (zero suppressions) |
| Every plan item ticked | `grep -c '\- \[ \]' docs/agents/sprint2/c3-site/plan-ready.md` | `0` | agentic-agile final gate |
| Scope | `git diff --name-only master -- . ':!docs/agents'` | only stories.md owned files (plus C1/earlier sprint files already on the branch) | M-012 |
| Deploy evidence (after merge to master) | `curl -fsS -o /dev/null -w '%{http_code}' https://snapback.run/ https://snapback.run/docs/ https://snapback.run/install.sh` | `200` three times | intake Done-when; orchestrator records, not a test |
| Commit hygiene | `git log master..HEAD --format=%B \| grep -ciE 'co-authored-by: claude\|claude-session'` | `0` | user memory: no AI trailers |
