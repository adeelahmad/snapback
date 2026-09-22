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
