---
type: init
story: S1-08
---

## S1-08/T1 · attempt 1 · red-worker · 2026-09-22T00:45:58Z

### Mandate
Write every T1 test bullet in `plan.md` § T1 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. SPEC.md does not exist yet (the git mv is GREEN work). Helper tests (repoRoot, section extraction) use shim helpers with wrong bodies.

### Scope
#### May
- Create the test files named for T1 in `tasks.md`: `test/projectdocs/helpers_test.go`, `test/projectdocs/spec_test.go`
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

## S1-08/T2 · attempt 1 · red-worker · 2026-09-22T00:48:39Z

### Mandate
Write every T2 test bullet in `plan.md` § T2 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. The new user-facing README does not exist yet (README.md is still the spec). Reuse shim helpers; tests must fail by assertion against the current state.

### Scope
#### May
- Create the test files named for T2 in `tasks.md`: `test/projectdocs/readme_test.go`
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

## S1-08/T3 · attempt 1 · red-worker · 2026-09-22T00:51:16Z

### Mandate
Write every T3 test bullet in `plan.md` § T3 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. New README does not exist yet; reuse shim helpers as-is.

### Scope
#### May
- Create the test files named for T3 in `tasks.md`: `test/projectdocs/readme_links_test.go`
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

## S1-08/T4 · attempt 1 · red-worker · 2026-09-22T00:53:56Z

### Mandate
Write every T4 test bullet in `plan.md` § T4 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. Scope honesty/backend scans to README.md and ARCHITECTURE.md only (never SPEC.md); whole-word matching per plan.md. Reuse shim helpers.

### Scope
#### May
- Create the test files named for T4 in `tasks.md`: `test/projectdocs/honesty_test.go`
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

## S1-08/T5 · attempt 1 · red-worker · 2026-09-22T00:56:06Z

### Mandate
Write every T5 test bullet in `plan.md` § T5 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. ARCHITECTURE.md absent; reuse shim helpers as-is.

### Scope
#### May
- Create the test files named for T5 in `tasks.md`: `test/projectdocs/architecture_test.go`
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

## S1-08/T6 · attempt 1 · red-worker · 2026-09-22T00:58:19Z

### Mandate
Write every T6 test bullet in `plan.md` § T6 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. NOTICE absent; reuse shim helpers as-is.

### Scope
#### May
- Create the test files named for T6 in `tasks.md`: `test/projectdocs/notice_test.go`
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

## S1-08/T7 · attempt 1 · red-worker · 2026-09-22T01:00:49Z

### Mandate
Write every T7 test bullet in `plan.md` § T7 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. CLAUDE.md absent; reuse shim helpers.

### Scope
#### May
- Create the test files named for T7 in `tasks.md`: `test/projectdocs/claudemd_test.go` (use the exact name from tasks.md/plan.md)
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

## S1-08/T8 · attempt 1 · red-worker · 2026-09-22T01:03:44Z

### Mandate
Write every T8 test bullet in `plan.md` § T8 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. DEVLOG.md absent; reuse shim helpers.

### Scope
#### May
- Create the test files named for T8 in `tasks.md`: `test/projectdocs/devlog_test.go` (use the exact name from tasks.md)
- Test helpers that a T8 test itself exercises (e.g. `repoRoot`, `readRepoFile`, `yamlBlock`, `runInstaller`) go in a marked shim `zz_agentic_shim_test.go` in the same test package (first line `// agentic:shim`) with deliberately WRONG bodies so those tests fail by assertion; helpers NOT under test may be real and live in the named helpers file.
#### May Not
- Create or edit any non-test artifact (workflow YAML, JSON config, install.sh, .goreleaser.yaml, SPEC.md/README.md, go.mod) — those are GREEN's job; touch other stories' files; suppress/skip tests (except the tool-absent `t.Skip` cases plan.md explicitly allows).

### Inputs
- `plan.md` § T8, `tasks.md` § T8 (contracts), `validate.md` § T8. Chain base: the story chain branch named in your prompt (T1 tests + shim present; stage-0 S1-01 merged underneath).

### Acceptance
`go test ./test/...` for this story's package compiles; every new test FAILS by assertion; `go vet ./...` clean; diff vs BASE_REF = only this task's test files (+ shim); output.md block appended; selfcheck PASS.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md in first); orchestrator relays output.md back.
- red-worker: make shim bodies differ so comparison tests cannot pass by accident.
- all: errcheck flags unchecked writes — use explicit `_, _ =` discard, never nolint.

## S1-08/scaffold · attempt 1 · scaffolder · 2026-09-22T01:05:46Z

### Mandate
S1-08 RED is complete (verified). Move every symbol in the marked shim `test/projectdocs/zz_agentic_shim_test.go` into its canonical file `test/projectdocs/helpers_test.go` as a stub whose body is `panic("SUB-AGENT-TODO: <recipe from tasks.md T1>")` with the exact signature the tests call (package-level vars/slices: declare without initializer + `// SUB-AGENT-TODO:` comment). Define each exactly once, delete the shim, and write the names to `.agentic/scaffold-symbols`.

### Scope
#### May
- Edit `test/projectdocs/helpers_test.go` (stubs only); delete the shim.
#### May Not
- Implement bodies; touch other files.

### Inputs
- `plan-ready.md`, `tasks.md` § T1, the shim. Chain base `chain/s1-08` @ 3276a14.

### Acceptance
`go vet ./test/projectdocs/` compiles; no new passing tests (PASS-ON-RED static guards excepted); no `agentic:shim` left; each symbol once; output.md block with `### Scaffold`; selfcheck PASS.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.

## S1-08/T1 · attempt 1 · green-worker · 2026-09-22T01:07:37Z

### Mandate
Implement S1-08 T1 per `tasks.md` § T1 with the least change that makes exactly the T1 tests (TestRepoRootHasGoMod, TestSectionExtractsBody, TestSpecHashMatchesStage0Readme, TestSpecSizeAndFirstLine; TestReadmeIsNotTheSpec passes when README.md is absent and SPEC.md exists) pass. SPEC.md sha256 must be 6bc35dadf698271b23790db43d2b521859d724081fae39dd1e281726e0fafe28. Do NOT write a new README (T2). Other S1-08 tests will fail until T2-T8.

### Scope
#### May
- `git mv README.md SPEC.md` (byte-for-byte, no content change) and `test/projectdocs/helpers_test.go` (fill repoRoot and section stubs only) only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T1, `plan-ready.md` § T1, `validate.md` § T1. Chain base `chain/s1-08` @ 3b6de11.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`SPEC.md README.md test/projectdocs/helpers_test.go`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S1-08/T2 · attempt 1 · green-worker · 2026-09-22T01:10:17Z

### Mandate
Implement S1-08 T2 per `tasks.md` § T2 with the least change that makes exactly the 6 T2 tests in test/projectdocs/readme_test.go (plus TestReadmeIsNotTheSpec) pass. Title Snapback; lead (before first ##) with the restore story incl. "restoring a file should be as easy as it was in 2008", tagline, a cp .snapshot/... example and the <!-- TODO: restore GIF marker; ## Status says pre-release; ## Install with `curl -fsSL https://snapback.example.com/install.sh | sh` marked placeholder; no working-release claims. Avoid banned honesty words and any backend other than Restic (T4 scans README). T3 adds credits/links later.

### Scope
#### May
- `README.md` (create the new user-facing README opening) only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T2, `plan-ready.md` § T2, `validate.md` § T2. Chain base `chain/s1-08` @ 246c947.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`README.md`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S1-08/T3 · attempt 1 · green-worker · 2026-09-22T01:12:33Z

### Mandate
Implement S1-08 T3 per `tasks.md` § T3 with the least change that makes exactly the 3 T3 tests in test/projectdocs/readme_links_test.go pass. T2 README tests must stay passing. Credit httm without naming the filesystems/tools it supports (README may name only Restic). T5/T6 run in parallel on other files.

### Scope
#### May
- `README.md` (edit: add ## Prior art and credit crediting httm — link, kimono-koans, MPL-2.0 — within the first four ## headings, and ## Documentation linking SPEC.md, ARCHITECTURE.md, CONTRIBUTING.md, SECURITY.md and the docs site) only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T3, `plan-ready.md` § T3, `validate.md` § T3. Chain base `chain/s1-08` @ 37477de.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`README.md`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S1-08/T5 · attempt 1 · green-worker · 2026-09-22T01:12:33Z

### Mandate
Implement S1-08 T5 per `tasks.md` § T5 with the least change that makes exactly the 4 T5 tests in test/projectdocs/architecture_test.go (plus TestArchitectureNamesNoOtherBackend and the ARCHITECTURE.md subtests of the honesty scans) pass. Current state (only cmd/snapback + internal/version exist), a table of all 15 SPEC §4 modules, the SnapshotProvider seam. May name the seam; must name NO backend other than Restic; no banned honesty words or first-of-kind claims. T3/T6 run in parallel on other files.

### Scope
#### May
- `ARCHITECTURE.md` (create) only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T5, `plan-ready.md` § T5, `validate.md` § T5. Chain base `chain/s1-08` @ 37477de.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`ARCHITECTURE.md`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S1-08/T6 · attempt 1 · green-worker · 2026-09-22T01:12:33Z

### Mandate
Implement S1-08 T6 per `tasks.md` § T6 with the least change that makes exactly the 3 T6 tests in test/projectdocs/notice_test.go pass. Go standard library — BSD-3-Clause (https://go.dev/LICENSE) and the sentence "This list is updated as dependencies are added." T3/T5 run in parallel on other files.

### Scope
#### May
- `NOTICE` (create) only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T6, `plan-ready.md` § T6, `validate.md` § T6. Chain base `chain/s1-08` @ 37477de.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`NOTICE`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S1-08/T7 · attempt 1 · green-worker · 2026-09-22T01:14:04Z

### Mandate
Implement S1-08 T7 per `tasks.md` § T7 with the least change that makes exactly the 4 T7 tests in test/projectdocs/claude_md_test.go pass. Follow the user's CLAUDE.md template (~/.claude/rules/project-docs.md), <=80 lines, Go commands (go build/test/vet, golangci-lint), link SPEC.md, DEVLOG.md, ARCHITECTURE.md. T8 (DEVLOG.md) runs in parallel — touch only CLAUDE.md.

### Scope
#### May
- `CLAUDE.md` (create) only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T7, `plan-ready.md` § T7, `validate.md` § T7. Chain base `chain/s1-08` @ 37477de.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`CLAUDE.md`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S1-08/T8 · attempt 1 · green-worker · 2026-09-22T01:14:04Z

### Mandate
Implement S1-08 T8 per `tasks.md` § T8 with the least change that makes exactly the 4 T8 tests in test/projectdocs/devlog_test.go pass. Follow ~/.claude/rules/workflow.md DEVLOG template: Working State (<=80 lines) with ### Key Files (current shape) (<=5 entries), Session Archive, Milestones, Mistakes & Lessons, Technical Debt. Content: honest record of Sprint 1 / Stage 0 so far (use ORCHESTRATOR.md in the main tree as the source; include the plugin gate issues and the test-helper consolidation debt). T7 runs in parallel — touch only DEVLOG.md.

### Scope
#### May
- `DEVLOG.md` (create) only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T8, `plan-ready.md` § T8, `validate.md` § T8. Chain base `chain/s1-08` @ 37477de.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`DEVLOG.md`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.
