---
type: init
story: S1-07
---

## S1-07/T1 · attempt 1 · red-worker · 2026-09-22T00:45:58Z

### Mandate
Write every T1 test bullet in `plan.md` § T1 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. Both T1 tests exercise helpers (repoRoot, yamlScalar) — put them in the marked shim with wrong bodies.

### Scope
#### May
- Create the test files named for T1 in `tasks.md`: `test/docs/helpers_test.go`
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

## S1-07/T2 · attempt 1 · red-worker · 2026-09-22T00:48:39Z

### Mandate
Write every T2 test bullet in `plan.md` § T2 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. requirements-docs.txt does not exist yet; reuse helpers (readRepoFile real, shim repoRoot/yamlScalar untouched).

### Scope
#### May
- Create the test files named for T2 in `tasks.md`: `test/docs/requirements_test.go`
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

## S1-07/T3 · attempt 1 · red-worker · 2026-09-22T00:50:21Z

### Mandate
Write every T3 test bullet in `plan.md` § T3 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. mkdocs.yml does not exist yet; reuse helpers; do not edit the shim.

### Scope
#### May
- Create the test files named for T3 in `tasks.md`: `test/docs/mkdocs_config_test.go`
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

## S1-07/T4 · attempt 1 · red-worker · 2026-09-22T00:53:56Z

### Mandate
Write every T4 test bullet in `plan.md` § T4 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. docs-site/index.md absent; reuse helpers; do not edit the shim.

### Scope
#### May
- Create the test files named for T4 in `tasks.md`: `test/docs/landing_test.go`
- Test helpers that a T4 test itself exercises (e.g. `repoRoot`, `readRepoFile`, `yamlBlock`, `runInstaller`) go in a marked shim `zz_agentic_shim_test.go` in the same test package (first line `// agentic:shim`) with deliberately WRONG bodies so those tests fail by assertion; helpers NOT under test may be real and live in the named helpers file.
#### May Not
- Create or edit any non-test artifact (workflow YAML, JSON config, install.sh, .goreleaser.yaml, SPEC.md/README.md, go.mod) — those are GREEN's job; touch other stories' files; suppress/skip tests (except the tool-absent `t.Skip` cases plan.md explicitly allows).

### Inputs
- `plan.md` § T4, `tasks.md` § T4 (contracts), `validate.md` § T4. Chain base: the story chain branch named in your prompt (T1 tests + shim present; stage-0 S1-01 merged underneath).

### Acceptance
`go test ./test/...` for this story's package compiles; every new test FAILS by assertion; `go vet ./...` clean; diff vs BASE_REF = only this task's test files (+ shim); output.md block appended; selfcheck PASS.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md in first); orchestrator relays output.md back.
- red-worker: make shim bodies differ so comparison tests cannot pass by accident.
- all: errcheck flags unchecked writes — use explicit `_, _ =` discard, never nolint.

## S1-07/T5 · attempt 1 · red-worker · 2026-09-22T00:56:06Z

### Mandate
Write every T5 test bullet in `plan.md` § T5 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. docs.yml absent; reuse helpers; do not edit the shim.

### Scope
#### May
- Create the test files named for T5 in `tasks.md`: `test/docs/workflow_test.go`
- Test helpers that a T5 test itself exercises (e.g. `repoRoot`, `readRepoFile`, `yamlBlock`, `runInstaller`) go in a marked shim `zz_agentic_shim_test.go` in the same test package (first line `// agentic:shim`) with deliberately WRONG bodies so those tests fail by assertion; helpers NOT under test may be real and live in the named helpers file.
#### May Not
- Create or edit any non-test artifact (workflow YAML, JSON config, install.sh, .goreleaser.yaml, SPEC.md/README.md, go.mod) — those are GREEN's job; touch other stories' files; suppress/skip tests (except the tool-absent `t.Skip` cases plan.md explicitly allows).

### Inputs
- `plan.md` § T5, `tasks.md` § T5 (contracts), `validate.md` § T5. Chain base: the story chain branch named in your prompt (T1 tests + shim present; stage-0 S1-01 merged underneath).

### Acceptance
`go test ./test/...` for this story's package compiles; every new test FAILS by assertion; `go vet ./...` clean; diff vs BASE_REF = only this task's test files (+ shim); output.md block appended; selfcheck PASS.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md in first); orchestrator relays output.md back.
- red-worker: make shim bodies differ so comparison tests cannot pass by accident.
- all: errcheck flags unchecked writes — use explicit `_, _ =` discard, never nolint.

## S1-07/T6 · attempt 1 · red-worker · 2026-09-22T00:58:19Z

### Mandate
Write every T6 test bullet in `plan.md` § T6 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. Scans docs-site/** and mkdocs.yml; banned-backend list excludes rclone (plan fixed). Negative scans that pass on empty input must be reported PASS-ON-RED, not weakened.

### Scope
#### May
- Create the test files named for T6 in `tasks.md`: `test/docs/honesty_test.go`
- Test helpers that a T6 test itself exercises (e.g. `repoRoot`, `readRepoFile`, `yamlBlock`, `runInstaller`) go in a marked shim `zz_agentic_shim_test.go` in the same test package (first line `// agentic:shim`) with deliberately WRONG bodies so those tests fail by assertion; helpers NOT under test may be real and live in the named helpers file.
#### May Not
- Create or edit any non-test artifact (workflow YAML, JSON config, install.sh, .goreleaser.yaml, SPEC.md/README.md, go.mod) — those are GREEN's job; touch other stories' files; suppress/skip tests (except the tool-absent `t.Skip` cases plan.md explicitly allows).

### Inputs
- `plan.md` § T6, `tasks.md` § T6 (contracts), `validate.md` § T6. Chain base: the story chain branch named in your prompt (T1 tests + shim present; stage-0 S1-01 merged underneath).

### Acceptance
`go test ./test/...` for this story's package compiles; every new test FAILS by assertion; `go vet ./...` clean; diff vs BASE_REF = only this task's test files (+ shim); output.md block appended; selfcheck PASS.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md in first); orchestrator relays output.md back.
- red-worker: make shim bodies differ so comparison tests cannot pass by accident.
- all: errcheck flags unchecked writes — use explicit `_, _ =` discard, never nolint.

## S1-07/T7 · attempt 1 · red-worker · 2026-09-22T01:00:49Z

### Mandate
Write every T7 test bullet in `plan.md` § T7 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. mkdocs build --strict: skip only when mkdocs is absent; otherwise fails because mkdocs.yml is missing.

### Scope
#### May
- Create the test files named for T7 in `tasks.md`: `test/docs/build_test.go`
- Test helpers that a T7 test itself exercises (e.g. `repoRoot`, `readRepoFile`, `yamlBlock`, `runInstaller`) go in a marked shim `zz_agentic_shim_test.go` in the same test package (first line `// agentic:shim`) with deliberately WRONG bodies so those tests fail by assertion; helpers NOT under test may be real and live in the named helpers file.
#### May Not
- Create or edit any non-test artifact (workflow YAML, JSON config, install.sh, .goreleaser.yaml, SPEC.md/README.md, go.mod) — those are GREEN's job; touch other stories' files; suppress/skip tests (except the tool-absent `t.Skip` cases plan.md explicitly allows).

### Inputs
- `plan.md` § T7, `tasks.md` § T7 (contracts), `validate.md` § T7. Chain base: the story chain branch named in your prompt (T1 tests + shim present; stage-0 S1-01 merged underneath).

### Acceptance
`go test ./test/...` for this story's package compiles; every new test FAILS by assertion; `go vet ./...` clean; diff vs BASE_REF = only this task's test files (+ shim); output.md block appended; selfcheck PASS.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md in first); orchestrator relays output.md back.
- red-worker: make shim bodies differ so comparison tests cannot pass by accident.
- all: errcheck flags unchecked writes — use explicit `_, _ =` discard, never nolint.

## S1-07/scaffold · attempt 1 · scaffolder · 2026-09-22T01:03:44Z

### Mandate
S1-07 RED is complete (T1–T7, verified). The only symbols living in the marked shim `test/docs/zz_agentic_shim_test.go` are test helpers the T1 tests exercise (`repoRoot`, `yamlScalar`, and any other shim symbol). Move each into its canonical file `test/docs/helpers_test.go` as a stub with body `panic("SUB-AGENT-TODO: <recipe from tasks.md T1>")` (preserving the exact signature the tests call), define each exactly once, then delete the shim. Write the moved symbol names one per line to `.agentic/scaffold-symbols`.

### Scope
#### May
- Edit `test/docs/helpers_test.go` (add stubs only), delete `test/docs/zz_agentic_shim_test.go`.
#### May Not
- Implement any body; touch other test files or non-test files.

### Inputs
- `plan-ready.md`, `tasks.md` § T1, the shim file. Chain base `chain/s1-07` @ 664b5ce.

### Acceptance
`go vet ./test/docs/` compiles; `go test ./test/docs/` has no passing test that was failing before except PASS-ON-RED guards; no `agentic:shim` left; each symbol defined once; output.md block with `### Scaffold`; selfcheck PASS.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.

## S1-07/T1 · attempt 1 · green-worker · 2026-09-22T01:05:31Z

### Mandate
Implement S1-07 T1 per `tasks.md` § T1 with the least change that makes exactly the 2 T1 tests (TestRepoRootHasGoMod, TestYAMLScalar) pass. This fills scaffolded test-helper stubs (the only Go symbols in S1-07). Other S1-07 tests will then fail on missing docs files — expected until T2-T7.

### Scope
#### May
- `test/docs/helpers_test.go` (fill the repoRoot and yamlScalar stubs only) only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T1, `plan-ready.md` § T1, `validate.md` § T1. Chain base `chain/s1-07` @ 5713cd6.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`test/docs/helpers_test.go`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S1-07/T2 · attempt 1 · green-worker · 2026-09-22T01:07:24Z

### Mandate
Implement S1-07 T2 per `tasks.md` § T2 with the least change that makes exactly the 2 T2 tests in test/docs/requirements_test.go pass. Pin mkdocs==1.6.1 and mkdocs-material to a version that really exists on PyPI (9.6.14 is the planned pick; confirm with `pip3 index versions mkdocs-material`). T3/T4 run in parallel on the same base — touch only your file.

### Scope
#### May
- `requirements-docs.txt` (create) only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T2, `plan-ready.md` § T2, `validate.md` § T2. Chain base `chain/s1-07` @ 92645e7.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`requirements-docs.txt`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S1-07/T3 · attempt 1 · green-worker · 2026-09-22T01:07:24Z

### Mandate
Implement S1-07 T3 per `tasks.md` § T3 with the least change that makes exactly the 5 T3 tests in test/docs/mkdocs_config_test.go pass. T4 (docs-site/index.md) runs in parallel on the same base — do NOT create docs-site files; TestNavEntriesExist may only pass after the orchestrator merges T4 — report its status honestly.

### Scope
#### May
- `mkdocs.yml` (create; docs_dir: docs-site; theme material; nav pointing at index.md) only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T3, `plan-ready.md` § T3, `validate.md` § T3. Chain base `chain/s1-07` @ 92645e7.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`mkdocs.yml`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S1-07/T4 · attempt 1 · green-worker · 2026-09-22T01:07:24Z

### Mandate
Implement S1-07 T4 per `tasks.md` § T4 with the least change that makes exactly the 6 T4 tests in test/docs/landing_test.go pass. Lead with the restore story; install section says Coming soon; states pre-release; per-release versioning not yet available; names Restic. No banned honesty words (substring "static" is banned — avoid "static site"), no other backends except rclone is allowed. T3 (mkdocs.yml) runs in parallel — touch only index.md.

### Scope
#### May
- `docs-site/index.md` (create) only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T4, `plan-ready.md` § T4, `validate.md` § T4. Chain base `chain/s1-07` @ 92645e7.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`docs-site/index.md`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S1-07/T5 · attempt 1 · green-worker · 2026-09-22T01:10:17Z

### Mandate
Implement S1-07 T5 per `tasks.md` § T5 with the least change that makes exactly the 6 T5 tests in test/docs/workflow_test.go pass. This is S1-07's LAST task: FULL standards matrix runs. All 26 S1-07 tests must pass. Build mkdocs --strict on every PR; deploy via upload-pages-artifact + deploy-pages on master and v* tags only; deploy job pages: write + id-token: write; top-level contents: read; pin actions; install docs deps from requirements-docs.txt.

### Scope
#### May
- `.github/workflows/docs.yml` (create) only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T5, `plan-ready.md` § T5, `validate.md` § T5. Chain base `chain/s1-07` @ c90c644.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`.github/workflows/docs.yml`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S1-07/T5 · attempt 2 · red-worker · 2026-09-22T01:13:54Z

### Mandate
Lint-only, behaviour-preserving fix to the RED test file written in S1-07/T5 attempt 1. The T5 GREEN (09e0892, docs.yml) passes all 26 S1-07 tests but the full matrix fails: `golangci-lint run` → staticcheck QF1001 "could apply De Morgan's law" at `test/docs/workflow_test.go:119:61`. Rewrite ONLY that expression into its logically equivalent De Morgan form so staticcheck is clean. Do not change what the test asserts.

### Scope
#### May
- Edit `test/docs/workflow_test.go` line ~119 (that one condition).
#### May Not
- Change any other line, weaken/skip/delete an assertion, touch docs.yml or other files.

### Feedback
From green-worker S1-07/T5 attempt 1 (status: escalate): suggested equivalent form `if !hasLineWith(deploy, "environment:", "github-pages") && (!strings.Contains(deploy, "environment:") || !hasLineWith(deploy, "name:", "github-pages")) {` — verify equivalence yourself.

### Inputs
- Chain base `chain/s1-07` @ 09e0892 (GREEN T5 included).

### Acceptance
`golangci-lint run` = 0 issues; all 26 S1-07 tests still PASS; diff = 1 line in workflow_test.go; output.md block states the before/after expression and why they are equivalent; selfcheck PASS.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree.
- red-worker: run golangci-lint on test files you write — the full matrix lints tests too.
