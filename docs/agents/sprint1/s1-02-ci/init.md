---
type: init
story: S1-02
---

## S1-02/T1 · attempt 1 · red-worker · 2026-09-22T00:45:58Z

### Mandate
Write every T1 test bullet in `plan.md` § T1 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. ci.yml does not exist yet, so tests reading it fail via t.Fatal.

### Scope
#### May
- Create the test files named for T1 in `tasks.md`: `test/ci/ci_test.go`, `test/ci/helpers_test.go`
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

## S1-02/T2 · attempt 1 · red-worker · 2026-09-22T00:48:39Z

### Mandate
Write every T2 test bullet in `plan.md` § T2 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. ci.yml does not exist yet; reuse the real helpers in helpers_test.go.

### Scope
#### May
- Create the test files named for T2 in `tasks.md`: `test/ci/matrix_test.go`
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

## S1-02/T3 · attempt 1 · red-worker · 2026-09-22T00:51:15Z

### Mandate
Write every T3 test bullet in `plan.md` § T3 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. .golangci.yml does not exist yet; the actionlint test may t.Skip only when actionlint is absent.

### Scope
#### May
- Create the test files named for T3 in `tasks.md`: `test/ci/lint_test.go`
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

## S1-02/T1 · attempt 1 · green-worker · 2026-09-22T00:52:57Z

### Mandate
Create `.github/workflows/ci.yml` with the least content that makes exactly T1's 14 tests in `test/ci/ci_test.go` pass, per `tasks.md` § T1 (push+pull_request triggers; setup-go from go.mod; gofmt/goimports, build, vet; `go test -race` with coverage + an 80% threshold step that fails the job; coverage artifact upload; govulncheck; shellcheck tolerating zero .sh files; no continue-on-error; every action pinned; no mount jobs; no references to other stories' files). SCAFFOLD was vacuous for S1-02 (no Go production symbols, no shims).

### Scope
#### May
- Create `.github/workflows/ci.yml` only.
#### May Not
- Edit tests or helpers; add the cross-compile job (T2) or golangci config (T3); touch other files.

### Inputs
- `tasks.md` § T1, `plan-ready.md` § T1, `validate.md` § T1. Chain base `chain/s1-02` @ c8000dd.

### Acceptance
The 14 T1 tests PASS (`go test ./test/ci/ -run 'TestCI' ` minus T2/T3 names — T2 matrix and T3 lint tests may still fail); `actionlint .github/workflows/ci.yml` clean if installed; diff = ci.yml only; output.md block; selfcheck PASS. Whole-repo matrix disabled for this intermediate GREEN (GATE_RUN_MATRIX=0); full matrix runs at the story's last GREEN.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: T3's TestCIActionsArePinned-style checks and T2/T3 tests share ci.yml — keep T1's YAML minimal and structured so T2 (cross-compile job) and T3 (golangci-lint step) can be added without rewriting.

## S1-02/T2 · attempt 1 · green-worker · 2026-09-22T00:55:42Z

### Mandate
Implement S1-02 T2 per `tasks.md` § T2 with the least change that makes exactly the 5 T2 tests in test/ci/matrix_test.go pass. T1's 14 tests and TestCIActionlint must stay passing; T3 lint tests may still fail.

### Scope
#### May
- `.github/workflows/ci.yml` (edit: add the cross-compile job) only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T2, `plan-ready.md` § T2, `validate.md` § T2. Chain base `chain/s1-02` @ b3a273a.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`.github/workflows/ci.yml`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S1-02/T3 · attempt 1 · green-worker · 2026-09-22T00:57:56Z

### Mandate
Implement S1-02 T3 per `tasks.md` § T3 with the least change that makes exactly the T3 tests in test/ci/lint_test.go pass. This is S1-02's LAST task: the FULL standards matrix runs. All 22 S1-02 tests must pass. golangci-lint v2 config format (installed locally: 2.13.2); pin the action's version: to a real release, not latest.

### Scope
#### May
- `.golangci.yml` (create) and `.github/workflows/ci.yml` (edit: add pinned golangci-lint-action step) only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T3, `plan-ready.md` § T3, `validate.md` § T3. Chain base `chain/s1-02` @ e13ee42.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`.golangci.yml .github/workflows/ci.yml`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S1-02+S1-04/review · attempt 1 · structural-reviewer · 2026-09-22T01:00:49Z

### Mandate
Read-only structural review of stage-0 @ 3bd0642 (S1-02 CI + S1-04 commitlint merged on top of S1-01). Detect orphan modules, parallel implementations, duplicate helpers across `test/ci` and `test/commitlint` (e.g. near-identical repoRoot/readRepoFile/indent helpers — note: each test package legitimately needs its own helpers since Go test packages cannot share unexported test code; only flag real duplication within one package or cross-package duplication that a shared non-test helper package would clearly fix), workflow drift between ci.yml and commitlint.yml (pinning style, permissions), leftover markers.

### Scope
#### May
- Read, run ctx-symbols/grep/go tooling/actionlint read-only; append review block to output.md.
#### May Not
- Edit anything.

### Acceptance
verdict: clean | isolated (name the story/task) | foundation-poisoning; findings with file:line; selfcheck PASS (known gate false positive: `_test.go` locals — report, don't fix).

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first).
- structural-reviewer: gate-structural-integrity misreports Go _test.go function-local duplicates as HIGH (no _test.go carve-out) — verify scope before calling poisoning.

## S1-02/fix1 · attempt 1 · green-worker · 2026-09-22T01:07:24Z

### Mandate
Implement S1-02 fix1 per `tasks.md` § fix1 with the least change that makes exactly all 22 existing S1-02 tests (they must stay passing; no new tests) pass. Structural review (S1-02+S1-04, verdict isolated) found ci.yml:24 and :59 install goimports/govulncheck with @latest, violating the pinned-toolchain rule (SPEC §18, standards.md). Fix only that. Base: branch chain/s1-02-fix = stage-0 @ 57f1518. Full matrix runs.

### Scope
#### May
- `.github/workflows/ci.yml` (edit: pin goimports and govulncheck `go install` versions to real released versions instead of @latest) only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § fix1, `plan-ready.md` § fix1, `validate.md` § fix1. Chain base `chain/s1-02` @ 57f1518.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`.github/workflows/ci.yml`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.
