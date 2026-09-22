---
type: init
story: S1-04
---

## S1-04/T1 · attempt 1 · red-worker · 2026-09-22T00:45:58Z

### Mandate
Write every T1 test bullet in `plan.md` § T1 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. .commitlintrc.json does not exist yet. TestCommitlintCLIVerdicts may t.Skip only when the commitlint CLI is absent, as plan.md allows.

### Scope
#### May
- Create the test files named for T1 in `tasks.md`: `test/commitlint/helpers_test.go`, `test/commitlint/config_test.go`
- Test helpers that a T1 test itself exercises (e.g. `repoRoot`, `readRepoFile`, `yamlBlock`, `runInstaller`) go in a marked shim `zz_agentic_shim_test.go` in the same test package (first line `// agentic:shim`) with deliberately WRONG bodies so those tests fail by assertion; helpers NOT under test may be real and live in the named helpers file.
#### May Not
- Create or edit any non-test artifact (workflow YAML, JSON config, install.sh, .goreleaser.yaml, SPEC.md/README.md, go.mod) — those are GREEN's job; touch other stories' files; suppress/skip tests (except the tool-absent `t.Skip` cases plan.md explicitly allows).

### Inputs
- `plan.md` § T1, `tasks.md` § T1 (contracts), `validate.md` § T1. Chain base: `stage-0` @ b0f2304 (S1-01 merged: module `github.com/adeelahmad/snapback`, `internal/version`).

### Acceptance
`go test ./test/...` for this story's package compiles; every new test FAILS by assertion; `go vet ./...` clean; diff vs BASE_REF = only this task's test files (+ shim); output.md block appended; selfcheck PASS.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md in first); orchestrator relays output.md back.
- red-worker: make shim bodies differ so comparison tests cannot pass by accident.
- all: errcheck flags unchecked writes — use explicit `_, _ =` discard, never nolint.

## S1-04/T2 · attempt 1 · red-worker · 2026-09-22T00:48:05Z

### Mandate
Write every T2 test bullet in `plan.md` § T2 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. commitlint.yml does not exist yet. blockChildKeys helper (already in helpers_test.go) is test infra here.

### Scope
#### May
- Create the test files named for T2 in `tasks.md`: `test/commitlint/workflow_test.go`
- Test helpers that a T2 test itself exercises (e.g. `repoRoot`, `readRepoFile`, `yamlBlock`, `runInstaller`) go in a marked shim `zz_agentic_shim_test.go` in the same test package (first line `// agentic:shim`) with deliberately WRONG bodies so those tests fail by assertion; helpers NOT under test may be real and live in the named helpers file.
#### May Not
- Create or edit any non-test artifact (workflow YAML, JSON config, install.sh, .goreleaser.yaml, SPEC.md/README.md, go.mod) — those are GREEN's job; touch other stories' files; suppress/skip tests (except the tool-absent `t.Skip` cases plan.md explicitly allows).

### Inputs
- `plan.md` § T2, `tasks.md` § T2 (contracts), `validate.md` § T2. Chain base: the story chain branch named in your prompt (T1 tests + shim present; stage-0 S1-01 merged underneath).

### Acceptance
`go test ./test/...` for this story's package compiles; every new test FAILS by assertion; `go vet ./...` clean; diff vs BASE_REF = only this task's test files (+ shim); output.md block appended; selfcheck PASS.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md in first); orchestrator relays output.md back.
- red-worker: make shim bodies differ so comparison tests cannot pass by accident.
- all: errcheck flags unchecked writes — use explicit `_, _ =` discard, never nolint.

## S1-04/T3 · attempt 1 · red-worker · 2026-09-22T00:51:15Z

### Mandate
Write every T3 test bullet in `plan.md` § T3 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. commitlint.yml does not exist yet; reuse helpers.

### Scope
#### May
- Create the test files named for T3 in `tasks.md`: `test/commitlint/workflow_pins_test.go`
- Test helpers that a T3 test itself exercises (e.g. `repoRoot`, `readRepoFile`, `yamlBlock`, `runInstaller`) go in a marked shim `zz_agentic_shim_test.go` in the same test package (first line `// agentic:shim`) with deliberately WRONG bodies so those tests fail by assertion; helpers NOT under test may be real and live in the named helpers file.
#### May Not
- Create or edit any non-test artifact (workflow YAML, JSON config, install.sh, .goreleaser.yaml, SPEC.md/README.md, go.mod) — those are GREEN's job; touch other stories' files; suppress/skip tests (except the tool-absent `t.Skip` cases plan.md explicitly allows).

### Inputs
- `plan.md` § T3, `tasks.md` § T3 (contracts), `validate.md` § T3. Chain base: the story chain branch named in your prompt (T1 tests + shim present; stage-0 S1-01 merged underneath).

### Acceptance
`go test ./test/...` for this story's package compiles; every new test FAILS by assertion; `go vet ./...` clean; diff vs BASE_REF = only this task's test files (+ shim); output.md block appended; selfcheck PASS.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md in first); orchestrator relays output.md back.
- red-worker: make shim bodies differ so comparison tests cannot pass by accident.
- all: errcheck flags unchecked writes — use explicit `_, _ =` discard, never nolint.

## S1-04/T1 · attempt 1 · green-worker · 2026-09-22T00:53:56Z

### Mandate
Create `.commitlintrc.json` with the least content that makes T1's config tests pass: extends `@commitlint/config-conventional`; `rules.type-enum` = `[2, "always", ["feat","fix","refactor","docs","test","chore","perf","ci"]]` exactly. SCAFFOLD was vacuous for S1-04 (no Go production symbols, no shims).

### Scope
#### May
- Create `.commitlintrc.json` only.
#### May Not
- Edit tests; create the workflow (T2/T3); touch other files.

### Inputs
- `tasks.md` § T1, `plan-ready.md` § T1, `validate.md` § T1. Chain base `chain/s1-04` @ 4fb073b.

### Acceptance
The 4 T1 config tests PASS (CLI verdict test may SKIP when commitlint is absent); T2/T3 workflow tests may still fail; diff = .commitlintrc.json only; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 (intermediate GREEN).

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.

## S1-04/T2 · attempt 1 · green-worker · 2026-09-22T00:55:42Z

### Mandate
Implement S1-04 T2 per `tasks.md` § T2 with the least change that makes exactly the 5 T2 tests in test/commitlint/workflow_test.go pass. T1 config tests must stay passing; T3 pinning tests may still fail.

### Scope
#### May
- `.github/workflows/commitlint.yml` (create) only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T2, `plan-ready.md` § T2, `validate.md` § T2. Chain base `chain/s1-04` @ 127cbd0.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`.github/workflows/commitlint.yml`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S1-04/T3 · attempt 1 · green-worker · 2026-09-22T00:57:44Z

### Mandate
Implement S1-04 T3 per `tasks.md` § T3 with the least change that makes exactly the 4 T3 tests in test/commitlint/workflow_pins_test.go pass. This is S1-04's LAST task: the FULL standards matrix runs (no GATE_RUN_MATRIX=0). All S1-04 tests must pass (CLI verdict test may SKIP if commitlint is absent).

### Scope
#### May
- `.github/workflows/commitlint.yml` (edit: pin Node + commitlint packages, add least-privilege permissions) only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T3, `plan-ready.md` § T3, `validate.md` § T3. Chain base `chain/s1-04` @ cc88587.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`.github/workflows/commitlint.yml`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.
