---
type: init
story: S2-12
---

## S2-12/T1 · attempt 1 · red-worker · 2026-09-22T03:32:34Z

### Mandate
Write every T1 test bullet in `plan.md` § T1 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. C3 site (human request). BRAND_DIR=/private/tmp/claude-501/-Users-adeelahmad-work-snapback/3a548722-26e5-4d22-8251-dbadece5e939/scratchpad/brand holds the design system (tokens.json, README.md, 10-logo.md, 20-assets.md, 30-terminal.md, Logos/, Icons/, Social/, fonts/) — read-only reference; T1 does not copy it. Node v26.0.0 is installed locally; npm available. Tests must be offline and deterministic.

### Scope
#### May
- Create the test files named for T1 in `tasks.md`: the T1 test files named in plan.md § T1
- Test helpers that a T1 test itself exercises (e.g. `repoRoot`, `readRepoFile`, `yamlBlock`, `runInstaller`) go in a marked shim `zz_agentic_shim_test.go` in the same test package (first line `// agentic:shim`) with deliberately WRONG bodies so those tests fail by assertion; helpers NOT under test may be real and live in the named helpers file.
#### May Not
- Create or edit any non-test artifact (workflow YAML, JSON config, install.sh, .goreleaser.yaml, SPEC.md/README.md, go.mod) — those are GREEN's job; touch other stories' files; suppress/skip tests (except the tool-absent `t.Skip` cases plan.md explicitly allows).

### Inputs
- `plan.md` § T1, `tasks.md` § T1 (contracts), `validate.md` § T1. Chain base: the story chain branch named in your prompt (stage-1 @ 2158ade = master + sprint2 plan).

### Acceptance
`go test ./test/...` for this story's package compiles; every new test FAILS by assertion; `go vet ./...` clean; diff vs BASE_REF = only this task's test files (+ shim); output.md block appended; selfcheck PASS.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md in first); orchestrator relays output.md back.
- red-worker: make shim bodies differ so comparison tests cannot pass by accident.
- all: errcheck flags unchecked writes — use explicit `_, _ =` discard, never nolint.

## S2-12/T1 · attempt 1 · green-worker · 2026-09-22T03:36:18Z

### Mandate
Implement S2-12 T1 per `tasks.md` § T1 with the least change that makes exactly the 7 Go tests in test/site/ and web/src/App.test.tsx (vitest) pass. Vite+React+TS per tasks.md § T1; Node pinned 26.0.0 in .nvmrc and engines; exact versions only (npm install --save-exact), commit package-lock.json; no network at test time. Run: GOTOOLCHAIN=auto go test ./test/site/ and (cd web && npm ci && npm test && npm run lint && npm run build). Never commit node_modules or dist. NOTICE lines for react and react-dom (MIT). TEST_GLOBS must include **/*.test.tsx.

### Scope
#### May
- web/ scaffold files, .gitignore, NOTICE (exactly the T1 file list in tasks.md) only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T1, `plan-ready.md` § T1, `validate.md` § T1. Chain base `chain/s2-12` @ 3512f58.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`web/** .gitignore NOTICE`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S2-12/T2 · attempt 1 · red-worker · 2026-09-22T03:40:10Z

### Mandate
Write every T2 test bullet in `plan.md` § T2 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. C3 wave 2 parallel RED (T2,T3,T4,T6,T8 run concurrently; T7 later). Own only your task's test files (per plan.md) — Go tests in test/site/<your>_test.go and/or web/src/**/*.test.tsx; shared helpers in test/site/helpers_test.go are read-only (add task-local helpers in your own file). No shim files in web/ unless the test cannot compile otherwise (then web/src/zz_shim_<task>.ts, marked). BRAND_DIR=/private/tmp/claude-501/-Users-adeelahmad-work-snapback/3a548722-26e5-4d22-8251-dbadece5e939/scratchpad/brand (tokens.json, fonts/, Logos/, Icons/, Social/, brand docs). TEST_GLOBS must include **/*.test.tsx and **/*.test.ts. Tests offline, deterministic.

### Scope
#### May
- Create the test files named for T2 in `tasks.md`: the T2 test files named in plan.md § T2
- Test helpers that a T2 test itself exercises (e.g. `repoRoot`, `readRepoFile`, `yamlBlock`, `runInstaller`) go in a marked shim `zz_agentic_shim_test.go` in the same test package (first line `// agentic:shim`) with deliberately WRONG bodies so those tests fail by assertion; helpers NOT under test may be real and live in the named helpers file.
#### May Not
- Create or edit any non-test artifact (workflow YAML, JSON config, install.sh, .goreleaser.yaml, SPEC.md/README.md, go.mod) — those are GREEN's job; touch other stories' files; suppress/skip tests (except the tool-absent `t.Skip` cases plan.md explicitly allows).

### Inputs
- `plan.md` § T2, `tasks.md` § T2 (contracts), `validate.md` § T2. Chain base: the story chain branch named in your prompt (stage-1 @ 2158ade = master + sprint2 plan).

### Acceptance
`go test ./test/...` for this story's package compiles; every new test FAILS by assertion; `go vet ./...` clean; diff vs BASE_REF = only this task's test files (+ shim); output.md block appended; selfcheck PASS.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md in first); orchestrator relays output.md back.
- red-worker: make shim bodies differ so comparison tests cannot pass by accident.
- all: errcheck flags unchecked writes — use explicit `_, _ =` discard, never nolint.

## S2-12/T3 · attempt 1 · red-worker · 2026-09-22T03:40:10Z

### Mandate
Write every T3 test bullet in `plan.md` § T3 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. C3 wave 2 parallel RED (T2,T3,T4,T6,T8 run concurrently; T7 later). Own only your task's test files (per plan.md) — Go tests in test/site/<your>_test.go and/or web/src/**/*.test.tsx; shared helpers in test/site/helpers_test.go are read-only (add task-local helpers in your own file). No shim files in web/ unless the test cannot compile otherwise (then web/src/zz_shim_<task>.ts, marked). BRAND_DIR=/private/tmp/claude-501/-Users-adeelahmad-work-snapback/3a548722-26e5-4d22-8251-dbadece5e939/scratchpad/brand (tokens.json, fonts/, Logos/, Icons/, Social/, brand docs). TEST_GLOBS must include **/*.test.tsx and **/*.test.ts. Tests offline, deterministic.

### Scope
#### May
- Create the test files named for T3 in `tasks.md`: the T3 test files named in plan.md § T3
- Test helpers that a T3 test itself exercises (e.g. `repoRoot`, `readRepoFile`, `yamlBlock`, `runInstaller`) go in a marked shim `zz_agentic_shim_test.go` in the same test package (first line `// agentic:shim`) with deliberately WRONG bodies so those tests fail by assertion; helpers NOT under test may be real and live in the named helpers file.
#### May Not
- Create or edit any non-test artifact (workflow YAML, JSON config, install.sh, .goreleaser.yaml, SPEC.md/README.md, go.mod) — those are GREEN's job; touch other stories' files; suppress/skip tests (except the tool-absent `t.Skip` cases plan.md explicitly allows).

### Inputs
- `plan.md` § T3, `tasks.md` § T3 (contracts), `validate.md` § T3. Chain base: the story chain branch named in your prompt (stage-1 @ 2158ade = master + sprint2 plan).

### Acceptance
`go test ./test/...` for this story's package compiles; every new test FAILS by assertion; `go vet ./...` clean; diff vs BASE_REF = only this task's test files (+ shim); output.md block appended; selfcheck PASS.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md in first); orchestrator relays output.md back.
- red-worker: make shim bodies differ so comparison tests cannot pass by accident.
- all: errcheck flags unchecked writes — use explicit `_, _ =` discard, never nolint.

## S2-12/T4 · attempt 1 · red-worker · 2026-09-22T03:40:10Z

### Mandate
Write every T4 test bullet in `plan.md` § T4 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. C3 wave 2 parallel RED (T2,T3,T4,T6,T8 run concurrently; T7 later). Own only your task's test files (per plan.md) — Go tests in test/site/<your>_test.go and/or web/src/**/*.test.tsx; shared helpers in test/site/helpers_test.go are read-only (add task-local helpers in your own file). No shim files in web/ unless the test cannot compile otherwise (then web/src/zz_shim_<task>.ts, marked). BRAND_DIR=/private/tmp/claude-501/-Users-adeelahmad-work-snapback/3a548722-26e5-4d22-8251-dbadece5e939/scratchpad/brand (tokens.json, fonts/, Logos/, Icons/, Social/, brand docs). TEST_GLOBS must include **/*.test.tsx and **/*.test.ts. Tests offline, deterministic.

### Scope
#### May
- Create the test files named for T4 in `tasks.md`: the T4 test files named in plan.md § T4
- Test helpers that a T4 test itself exercises (e.g. `repoRoot`, `readRepoFile`, `yamlBlock`, `runInstaller`) go in a marked shim `zz_agentic_shim_test.go` in the same test package (first line `// agentic:shim`) with deliberately WRONG bodies so those tests fail by assertion; helpers NOT under test may be real and live in the named helpers file.
#### May Not
- Create or edit any non-test artifact (workflow YAML, JSON config, install.sh, .goreleaser.yaml, SPEC.md/README.md, go.mod) — those are GREEN's job; touch other stories' files; suppress/skip tests (except the tool-absent `t.Skip` cases plan.md explicitly allows).

### Inputs
- `plan.md` § T4, `tasks.md` § T4 (contracts), `validate.md` § T4. Chain base: the story chain branch named in your prompt (stage-1 @ 2158ade = master + sprint2 plan).

### Acceptance
`go test ./test/...` for this story's package compiles; every new test FAILS by assertion; `go vet ./...` clean; diff vs BASE_REF = only this task's test files (+ shim); output.md block appended; selfcheck PASS.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md in first); orchestrator relays output.md back.
- red-worker: make shim bodies differ so comparison tests cannot pass by accident.
- all: errcheck flags unchecked writes — use explicit `_, _ =` discard, never nolint.

## S2-12/T6 · attempt 1 · red-worker · 2026-09-22T03:40:10Z

### Mandate
Write every T6 test bullet in `plan.md` § T6 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. C3 wave 2 parallel RED (T2,T3,T4,T6,T8 run concurrently; T7 later). Own only your task's test files (per plan.md) — Go tests in test/site/<your>_test.go and/or web/src/**/*.test.tsx; shared helpers in test/site/helpers_test.go are read-only (add task-local helpers in your own file). No shim files in web/ unless the test cannot compile otherwise (then web/src/zz_shim_<task>.ts, marked). BRAND_DIR=/private/tmp/claude-501/-Users-adeelahmad-work-snapback/3a548722-26e5-4d22-8251-dbadece5e939/scratchpad/brand (tokens.json, fonts/, Logos/, Icons/, Social/, brand docs). TEST_GLOBS must include **/*.test.tsx and **/*.test.ts. Tests offline, deterministic.

### Scope
#### May
- Create the test files named for T6 in `tasks.md`: the T6 test files named in plan.md § T6
- Test helpers that a T6 test itself exercises (e.g. `repoRoot`, `readRepoFile`, `yamlBlock`, `runInstaller`) go in a marked shim `zz_agentic_shim_test.go` in the same test package (first line `// agentic:shim`) with deliberately WRONG bodies so those tests fail by assertion; helpers NOT under test may be real and live in the named helpers file.
#### May Not
- Create or edit any non-test artifact (workflow YAML, JSON config, install.sh, .goreleaser.yaml, SPEC.md/README.md, go.mod) — those are GREEN's job; touch other stories' files; suppress/skip tests (except the tool-absent `t.Skip` cases plan.md explicitly allows).

### Inputs
- `plan.md` § T6, `tasks.md` § T6 (contracts), `validate.md` § T6. Chain base: the story chain branch named in your prompt (stage-1 @ 2158ade = master + sprint2 plan).

### Acceptance
`go test ./test/...` for this story's package compiles; every new test FAILS by assertion; `go vet ./...` clean; diff vs BASE_REF = only this task's test files (+ shim); output.md block appended; selfcheck PASS.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md in first); orchestrator relays output.md back.
- red-worker: make shim bodies differ so comparison tests cannot pass by accident.
- all: errcheck flags unchecked writes — use explicit `_, _ =` discard, never nolint.

## S2-12/T8 · attempt 1 · red-worker · 2026-09-22T03:40:10Z

### Mandate
Write every T8 test bullet in `plan.md` § T8 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. C3 wave 2 parallel RED (T2,T3,T4,T6,T8 run concurrently; T7 later). Own only your task's test files (per plan.md) — Go tests in test/site/<your>_test.go and/or web/src/**/*.test.tsx; shared helpers in test/site/helpers_test.go are read-only (add task-local helpers in your own file). No shim files in web/ unless the test cannot compile otherwise (then web/src/zz_shim_<task>.ts, marked). BRAND_DIR=/private/tmp/claude-501/-Users-adeelahmad-work-snapback/3a548722-26e5-4d22-8251-dbadece5e939/scratchpad/brand (tokens.json, fonts/, Logos/, Icons/, Social/, brand docs). TEST_GLOBS must include **/*.test.tsx and **/*.test.ts. Tests offline, deterministic.

### Scope
#### May
- Create the test files named for T8 in `tasks.md`: the T8 test files named in plan.md § T8
- Test helpers that a T8 test itself exercises (e.g. `repoRoot`, `readRepoFile`, `yamlBlock`, `runInstaller`) go in a marked shim `zz_agentic_shim_test.go` in the same test package (first line `// agentic:shim`) with deliberately WRONG bodies so those tests fail by assertion; helpers NOT under test may be real and live in the named helpers file.
#### May Not
- Create or edit any non-test artifact (workflow YAML, JSON config, install.sh, .goreleaser.yaml, SPEC.md/README.md, go.mod) — those are GREEN's job; touch other stories' files; suppress/skip tests (except the tool-absent `t.Skip` cases plan.md explicitly allows).

### Inputs
- `plan.md` § T8, `tasks.md` § T8 (contracts), `validate.md` § T8. Chain base: the story chain branch named in your prompt (stage-1 @ 2158ade = master + sprint2 plan).

### Acceptance
`go test ./test/...` for this story's package compiles; every new test FAILS by assertion; `go vet ./...` clean; diff vs BASE_REF = only this task's test files (+ shim); output.md block appended; selfcheck PASS.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md in first); orchestrator relays output.md back.
- red-worker: make shim bodies differ so comparison tests cannot pass by accident.
- all: errcheck flags unchecked writes — use explicit `_, _ =` discard, never nolint.

## S2-12/T2 · attempt 1 · green-worker · 2026-09-22T03:44:57Z

### Mandate
Implement S2-12 T2 per `tasks.md` § T2 with the least change that makes exactly the 6 vitest cases in web/scripts/gen-tokens.test.ts and TestTextContrastAA, TestNonTextContrast, TestTokenFileShape, TestContrastRatioKnownValues pass. BRAND_DIR=/private/tmp/claude-501/-Users-adeelahmad-work-snapback/3a548722-26e5-4d22-8251-dbadece5e939/scratchpad/brand. focus token emitted as --focus: var(--blue); in both themes. Deterministic output. Run go test ./test/site/ -run 'Contrast|TokenFile' and (cd web && npm ci && npm test && npm run lint && npm run build). T3 GREEN (fonts/assets) runs after you.

### Scope
#### May
- web/tokens.json (copied verbatim from BRAND_DIR), web/scripts/gen-tokens.mjs, web/src/styles/tokens.css (generated, committed), web/src/main.tsx (import line), web/package.json (gen:tokens → node scripts/gen-tokens.mjs), and test/site/zz_agentic_shim_t2_test.go (ORCHESTRATOR EXCEPTION: replace the shim with a real WCAG 2.x contrastRatio helper and drop the agentic:shim marker — it is test-support code the RED stubbed) only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T2, `plan-ready.md` § T2, `validate.md` § T2. Chain base `chain2/c3-site` @ 723b061.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`web/tokens.json web/scripts/gen-tokens.mjs web/src/styles/tokens.css web/src/main.tsx web/package.json test/site/zz_agentic_shim_t2_test.go`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S2-12/T8 · attempt 1 · green-worker · 2026-09-22T03:45:33Z

### Mandate
Implement S2-12 T8 per `tasks.md` § T8 with the least change that makes exactly the 4 T8 tests in test/ci/site_job_test.go pass. Side chain chain2/c3-t8 (parallel with T2). Job per tasks.md § T8: actions/setup-node pinned to a released version with node-version-file web/.nvmrc, working-directory web, steps exactly npm ci, npm run lint, npm test, npm run build (in that order), no install fallbacks. Pin all actions to released versions. actionlint clean; all test/ci pass.

### Scope
#### May
- .github/workflows/ci.yml (add the site job only) only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T8, `plan-ready.md` § T8, `validate.md` § T8. Chain base `chain/s2-12` @ cef15cc.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`.github/workflows/ci.yml`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S2-12/T3 · attempt 1 · green-worker · 2026-09-22T03:48:47Z

### Mandate
Implement S2-12 T3 per `tasks.md` § T3 with the least change that makes exactly the T3 tests in test/site/assets_test.go and icons_test.go pass. BRAND_DIR=/private/tmp/claude-501/-Users-adeelahmad-work-snapback/3a548722-26e5-4d22-8251-dbadece5e939/scratchpad/brand (Logos/, Icons/, Social/, fonts/). Copy byte-for-byte; filenames per plan.md § T3. No CDN, no banned icon names.

### Scope
#### May
- web/public/fonts/*.woff2, web/public/** brand files (copied from BRAND_DIR, never redrawn), NOTICE (font licences: IBM Plex Sans/Mono, JetBrains Mono — SIL OFL 1.1) only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T3, `plan-ready.md` § T3, `validate.md` § T3. Chain base `chain/s2-12` @ 8335795.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`web/public/** NOTICE`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S2-12/T4 · attempt 1 · scaffolder · 2026-09-22T03:48:48Z

### Mandate
Replace the RED shim `web/src/zz_shim_t4.ts` with production stubs: `web/src/sections/Header.tsx`, `Hero.tsx`, `HowItWorks.tsx` (each a component returning `null` with a `// SUB-AGENT-TODO: <recipe from tasks.md § T4>` comment), `web/src/content.ts` (typed content model per tasks.md § T4 with empty/placeholder values marked SUB-AGENT-TODO), and an empty `web/src/styles/site.css`. Change ONLY the import line in `web/src/sections.test.tsx` that points at `./zz_shim_t4` to import from `./sections/Header`, `./sections/Hero`, `./sections/HowItWorks` (scaffolder is allowed this one-line test import change). Delete the shim.

### Acceptance
`npm --prefix web run lint` clean; T4 vitest cases fail by assertion (not import errors); other tests unchanged; no `agentic:shim` left; commit `chore: scaffold site sections (S2-12 T4)`, no AI trailers.

## S2-12/T7 · attempt 1 · red-worker · 2026-09-22T03:49:13Z

### Mandate
Write every T7 test bullet in `plan.md` § T7 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. Side chain chain2/c3-t7 from stage-1 (includes C1 domain + C4 launch README, but NOT web/). Tests are static file checks: docs.yml builds MkDocs into _site/docs/, the React build into _site/, copies install.sh to _site/install.sh, uploads _site; mkdocs.yml site_url https://snapback.run/docs/; README docs link https://snapback.run/docs/. Update the existing pins (test/docs/mkdocs_config_test.go site_url, test/projectdocs readme_links docsSiteURL, test/docs/workflow_test.go install.sh step) — note each in output.md.

### Scope
#### May
- Create the test files named for T7 in `tasks.md`: the T7 test files named in plan.md § T7 (and updates to the C1/C4 tests that pin site_url and the README docs link)
- Test helpers that a T7 test itself exercises (e.g. `repoRoot`, `readRepoFile`, `yamlBlock`, `runInstaller`) go in a marked shim `zz_agentic_shim_test.go` in the same test package (first line `// agentic:shim`) with deliberately WRONG bodies so those tests fail by assertion; helpers NOT under test may be real and live in the named helpers file.
#### May Not
- Create or edit any non-test artifact (workflow YAML, JSON config, install.sh, .goreleaser.yaml, SPEC.md/README.md, go.mod) — those are GREEN's job; touch other stories' files; suppress/skip tests (except the tool-absent `t.Skip` cases plan.md explicitly allows).

### Inputs
- `plan.md` § T7, `tasks.md` § T7 (contracts), `validate.md` § T7. Chain base: the story chain branch named in your prompt (stage-1 @ 2158ade = master + sprint2 plan).

### Acceptance
`go test ./test/...` for this story's package compiles; every new test FAILS by assertion; `go vet ./...` clean; diff vs BASE_REF = only this task's test files (+ shim); output.md block appended; selfcheck PASS.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md in first); orchestrator relays output.md back.
- red-worker: make shim bodies differ so comparison tests cannot pass by accident.
- all: errcheck flags unchecked writes — use explicit `_, _ =` discard, never nolint.

## S2-12/T4 · attempt 1 · green-worker · 2026-09-22T03:50:39Z

### Mandate
Implement S2-12 T4 per `tasks.md` § T4 with the least change that makes exactly the 6 T4 vitest cases in web/src/sections.test.tsx pass. Side chain chain2/c3-t4 (parallel with T3 GREEN in web/public). Design system: Liquid Glass panels over tinted radial grounds, tokens only (var(--…)), IBM Plex, man-page voice (lowercase snapback, sentence case, no emoji, no '!'), command before description. Honest: Stage 0 skeleton + Stage 1 compatibility evidence; no working .snapshot mount claimed — the cp .snapshot example is labelled as the design goal. No blue/yellow as text colour (tests ban it); links use accent-text. Run (cd web && npm ci && npm test && npm run lint && npm run build).

### Scope
#### May
- web/src/sections/{Header,Hero,HowItWorks}.tsx, web/src/content.ts, web/src/styles/site.css, web/src/App.tsx, web/src/main.tsx (site.css import) only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T4, `plan-ready.md` § T4, `validate.md` § T4. Chain base `chain/s2-12` @ 2772af5.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`web/src/sections/** web/src/content.ts web/src/styles/site.css web/src/App.tsx web/src/main.tsx`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S2-12/T6 · attempt 1 · green-worker · 2026-09-22T03:50:53Z

### Mandate
Implement S2-12 T6 per `tasks.md` § T6 with the least change that makes exactly the 5 T6 tests in test/site/meta_test.go pass. Head per 20-assets.md: canonical https://snapback.run/, color-scheme light dark, theme-color per scheme from tokens canvas, description 50-160 chars in the brand voice (lowercase snapback, no '!'), og:title/description/url/type/image (og-card-dark.png, absolute https://snapback.run/…), twitter:card summary_large_image, icon svg + 16/32 png + apple-touch-icon 180. No external origins. A parallel T4 GREEN edits web/src — do not touch it.

### Scope
#### May
- web/index.html (head meta only) only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T6, `plan-ready.md` § T6, `validate.md` § T6. Chain base `chain/s2-12` @ 63b8f8c.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`web/index.html`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S2-12/T7 · attempt 1 · green-worker · 2026-09-22T03:51:32Z

### Mandate
Implement S2-12 T7 per `tasks.md` § T7 with the least change that makes exactly the T7 tests in test/docs/ and test/projectdocs/readme_links_test.go pass. Side chain chain2/c3-t7. docs.yml build job: checkout, setup-python (as now) + mkdocs build --strict --site-dir site, actions/setup-node pinned with node-version-file web/.nvmrc, npm ci + npm run build in web, then exactly: cp -R web/dist/. _site/ ; cp -R site/. _site/docs/ ; cp install.sh _site/install.sh ; upload-pages-artifact path: _site. mkdocs site_url https://snapback.run/docs/. README Documentation link https://snapback.run/docs/ (keep everything else in README unchanged). Note: web/ is NOT on this side chain — tests are static; do not try to build web here. actionlint clean; all test/... pass; mkdocs --strict passes.

### Scope
#### May
- .github/workflows/docs.yml, mkdocs.yml (site_url), README.md (Documentation docs link only) only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T7, `plan-ready.md` § T7, `validate.md` § T7. Chain base `chain/s2-12` @ 55ed3d2.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`.github/workflows/docs.yml mkdocs.yml README.md`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S2-12/T6fix · attempt 1 · red-worker · 2026-09-22T03:52:45Z

### Mandate
Write every T6fix test bullet in `plan.md` § T6fix at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. Test bug found by T6 GREEN: canvasValues decodes every colour token value as map[string]string, but tokens.json has single-string values (e.g. accent "#1e7e43" applies to all themes) → json unmarshal error before index.html is checked. Fix the helper (json.RawMessage; accept string = same for all themes, or object {light,dark}) and look up only the canvas token. After the fix TestThemeColorPerScheme is expected to PASS against the committed index.html (#f8f9fa light, #202124 dark) — PASS-ON-RED waived by orchestrator; report before/after.

### Scope
#### May
- Create the test files named for T6fix in `tasks.md`: test/site/meta_test.go only
- Test helpers that a T6fix test itself exercises (e.g. `repoRoot`, `readRepoFile`, `yamlBlock`, `runInstaller`) go in a marked shim `zz_agentic_shim_test.go` in the same test package (first line `// agentic:shim`) with deliberately WRONG bodies so those tests fail by assertion; helpers NOT under test may be real and live in the named helpers file.
#### May Not
- Create or edit any non-test artifact (workflow YAML, JSON config, install.sh, .goreleaser.yaml, SPEC.md/README.md, go.mod) — those are GREEN's job; touch other stories' files; suppress/skip tests (except the tool-absent `t.Skip` cases plan.md explicitly allows).

### Inputs
- `plan.md` § T6fix, `tasks.md` § T6fix (contracts), `validate.md` § T6fix. Chain base: the story chain branch named in your prompt (stage-1 @ 2158ade = master + sprint2 plan).

### Acceptance
`go test ./test/...` for this story's package compiles; every new test FAILS by assertion; `go vet ./...` clean; diff vs BASE_REF = only this task's test files (+ shim); output.md block appended; selfcheck PASS.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md in first); orchestrator relays output.md back.
- red-worker: make shim bodies differ so comparison tests cannot pass by accident.
- all: errcheck flags unchecked writes — use explicit `_, _ =` discard, never nolint.

## S2-12/T5 · attempt 1 · red-worker · 2026-09-22T03:53:37Z

### Mandate
Write every T5 test bullet in `plan.md` § T5 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. Last C3 task. Chain has T1-T4, T6 (T6fix pending in test/site/meta_test.go — do not touch it). Voice lint, honesty and section-order tests over the rendered page; status stages read from SPEC.md §22. Vitest under web/src/. TEST_GLOBS **/*.test.ts **/*.test.tsx.

### Scope
#### May
- Create the test files named for T5 in `tasks.md`: the T5 test files named in plan.md § T5
- Test helpers that a T5 test itself exercises (e.g. `repoRoot`, `readRepoFile`, `yamlBlock`, `runInstaller`) go in a marked shim `zz_agentic_shim_test.go` in the same test package (first line `// agentic:shim`) with deliberately WRONG bodies so those tests fail by assertion; helpers NOT under test may be real and live in the named helpers file.
#### May Not
- Create or edit any non-test artifact (workflow YAML, JSON config, install.sh, .goreleaser.yaml, SPEC.md/README.md, go.mod) — those are GREEN's job; touch other stories' files; suppress/skip tests (except the tool-absent `t.Skip` cases plan.md explicitly allows).

### Inputs
- `plan.md` § T5, `tasks.md` § T5 (contracts), `validate.md` § T5. Chain base: the story chain branch named in your prompt (stage-1 @ 2158ade = master + sprint2 plan).

### Acceptance
`go test ./test/...` for this story's package compiles; every new test FAILS by assertion; `go vet ./...` clean; diff vs BASE_REF = only this task's test files (+ shim); output.md block appended; selfcheck PASS.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md in first); orchestrator relays output.md back.
- red-worker: make shim bodies differ so comparison tests cannot pass by accident.
- all: errcheck flags unchecked writes — use explicit `_, _ =` discard, never nolint.

## S2-12/T5 · attempt 1 · scaffolder · 2026-09-22T03:56:40Z

### Mandate
Replace `web/src/zz_agentic_shim_t5.ts` with production stubs: `web/src/sections/Limits.tsx`, `Install.tsx`, `Status.tsx`, `Footer.tsx` (each returns null with a `// SUB-AGENT-TODO: <recipe from tasks.md § T5>` comment) and a `stages` export added to `web/src/content.ts` (empty array, SUB-AGENT-TODO). Change ONLY the import lines in `web/src/copy.test.tsx` and `web/src/status.test.ts` that point at the shim so they import from `./sections/*` and `./content`. Delete the shim.

### Acceptance
`npm --prefix web run lint` clean; T5 tests fail by assertion (TestVoiceLint/TestHonesty may pass); earlier tests unchanged; no shim marker; commit `chore: scaffold site limits, install, status, footer (S2-12 T5)`, no AI trailers.

## S2-12/T5 · attempt 1 · green-worker · 2026-09-22T03:57:51Z

### Mandate
Implement S2-12 T5 per `tasks.md` § T5 with the least change that makes exactly the T5 vitest tests (copy.test.tsx, status.test.ts) + all earlier tests pass. LAST C3 task on this chain: full matrix + Node gates (npm ci, lint, test, build). Section ids hero, how-it-works, limits, install, status; stages from SPEC.md §22 (only stage 0 done). Limits: what snapback does NOT do (no scheduling, no retention, no file-content cache, no union/overlay FS, no Windows, Restic only). Install: exactly one 'curl -fsSL https://snapback.run/install.sh | sh'. Footer: /docs/, GitHub repo, LICENSE. Voice: lowercase snapback, no emoji, no '!', no banned icons; honest (skeleton + compatibility evidence). Tokens only; no blue/yellow text.

### Scope
#### May
- web/src/sections/{Limits,Install,Status,Footer}.tsx, web/src/content.ts, web/src/App.tsx, web/src/styles/site.css only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T5, `plan-ready.md` § T5, `validate.md` § T5. Chain base `chain/s2-12` @ 0961bed.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`web/src/sections/** web/src/content.ts web/src/App.tsx web/src/styles/site.css`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S2-12/integrate · attempt 1 · green-worker · 2026-09-22T04:01:55Z

### Mandate
Integrate C3 into the stage-1 line on branch `chain2/c3-int` (= stage-1): `git merge --no-ff chain2/c3-site -m "chore: merge S2-12 snapback.run site into stage-1"`, resolve the ONE conflict in `.github/workflows/ci.yml` by keeping BOTH sides (stage-1's fuse-linux upload `if-no-files-found: error` AND the new `site` job), commit; then `git merge --no-ff chain2/c3-t7 -m "chore: merge S2-12 pages layout into stage-1"` (resolve any conflict keeping both intents). Do not change anything else.

### Acceptance
Full standards matrix green; `GOTOOLCHAIN=auto go test ./...` (incl. test/site, test/ci, test/docs, test/projectdocs); in web/: npm ci, npm test, npm run lint, npm run build; actionlint clean; local dry run of the Pages assembly: `mkdocs build --strict --site-dir site && (cd web && npm run build) && rm -rf _site && mkdir -p _site/docs && cp -R web/dist/. _site/ && cp -R site/. _site/docs/ && cp install.sh _site/install.sh && test -f _site/index.html && test -f _site/docs/index.html && test -f _site/install.sh`, then delete _site/ site/ web/dist web/node_modules. Merge commits authored by the repo user, no AI trailers.

## S2-12/rename · attempt 1 · red-worker · 2026-09-22T04:13:50Z

### Mandate
Write every rename test bullet in `plan.md` § rename at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. Final-gate finding: the file name matches the banned zz_agentic_shim pattern although its body is real test-support code (WCAG contrast helpers). git mv it to test/site/contrast_helpers_test.go; no content change. All test/site tests must still PASS (PASS-ON-RED waived). Confirm git grep -l zz_agentic_shim returns nothing.

### Scope
#### May
- Create the test files named for rename in `tasks.md`: test/site/zz_agentic_shim_t2_test.go → test/site/contrast_helpers_test.go (git mv only)
- Test helpers that a rename test itself exercises (e.g. `repoRoot`, `readRepoFile`, `yamlBlock`, `runInstaller`) go in a marked shim `zz_agentic_shim_test.go` in the same test package (first line `// agentic:shim`) with deliberately WRONG bodies so those tests fail by assertion; helpers NOT under test may be real and live in the named helpers file.
#### May Not
- Create or edit any non-test artifact (workflow YAML, JSON config, install.sh, .goreleaser.yaml, SPEC.md/README.md, go.mod) — those are GREEN's job; touch other stories' files; suppress/skip tests (except the tool-absent `t.Skip` cases plan.md explicitly allows).

### Inputs
- `plan.md` § rename, `tasks.md` § rename (contracts), `validate.md` § rename. Chain base: the story chain branch named in your prompt (stage-1 @ 2158ade = master + sprint2 plan).

### Acceptance
`go test ./test/...` for this story's package compiles; every new test FAILS by assertion; `go vet ./...` clean; diff vs BASE_REF = only this task's test files (+ shim); output.md block appended; selfcheck PASS.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md in first); orchestrator relays output.md back.
- red-worker: make shim bodies differ so comparison tests cannot pass by accident.
- all: errcheck flags unchecked writes — use explicit `_, _ =` discard, never nolint.
