---
type: init
story: S1-06
---

## S1-06/T1 · attempt 1 · red-worker · 2026-09-22T00:45:58Z

### Mandate
Write every T1 test bullet in `plan.md` § T1 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. .goreleaser.yaml does not exist yet. TestGoreleaserVersionVarsCompile references internal/version vars that already exist and would pass — make it fail by assertion only if plan.md asserts something not yet true; otherwise report it in output.md as a test that cannot be RED (do not weaken it).

### Scope
#### May
- Create the test files named for T1 in `tasks.md`: `test/release/helpers_test.go`, `test/release/goreleaser_build_test.go`
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

## S1-06/T2 · attempt 1 · red-worker · 2026-09-22T00:48:39Z

### Mandate
Write every T2 test bullet in `plan.md` § T2 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. .goreleaser.yaml does not exist yet; reuse helpers. Archive naming must match the S1-03 contract in s1-03-installer/tasks.md.

### Scope
#### May
- Create the test files named for T2 in `tasks.md`: `test/release/goreleaser_artifacts_test.go`
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

## S1-06/T3 · attempt 1 · red-worker · 2026-09-22T00:51:16Z

### Mandate
Write every T3 test bullet in `plan.md` § T3 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. .releaserc.json and CHANGELOG.md do not exist yet; reuse helpers.

### Scope
#### May
- Create the test files named for T3 in `tasks.md`: `test/release/releaserc_test.go`
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

## S1-06/T4 · attempt 1 · red-worker · 2026-09-22T00:51:16Z

### Mandate
Write every T4 test bullet in `plan.md` § T4 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. release.yml does not exist yet; the actionlint test may t.Skip only when actionlint is absent. T3 RED runs in parallel in another worktree on the same base; touch only your file.

### Scope
#### May
- Create the test files named for T4 in `tasks.md`: `test/release/workflow_test.go`
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

## S1-06/T1 · attempt 1 · green-worker · 2026-09-22T00:55:42Z

### Mandate
Implement S1-06 T1 per `tasks.md` § T1 with the least change that makes exactly T1's tests in test/release/goreleaser_build_test.go pass. T2-T4 tests may still fail. SCAFFOLD vacuous for S1-06 (no Go symbols, no shims).

### Scope
#### May
- `.goreleaser.yaml` (create: version 2, project_name, builds only) only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T1, `plan-ready.md` § T1, `validate.md` § T1. Chain base `chain/s1-06` @ 45866ed.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`.goreleaser.yaml`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S1-06/T2 · attempt 1 · green-worker · 2026-09-22T00:57:44Z

### Mandate
Implement S1-06 T2 per `tasks.md` § T2 with the least change that makes exactly the T2 tests in test/release/goreleaser_artifacts_test.go pass. T1 tests must stay passing; T3/T4 tests may still fail. Archive names must match the S1-03 contract (s1-03-installer/tasks.md top); goarm 7 / gomips softfloat keep .Arch bare; cosign keyless signs checksums.txt; release mode append; no Homebrew/apt/OpenWrt publishers.

### Scope
#### May
- `.goreleaser.yaml` (edit: add archives, checksum, signs, release, changelog) only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T2, `plan-ready.md` § T2, `validate.md` § T2. Chain base `chain/s1-06` @ 8b81adb.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`.goreleaser.yaml`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S1-06/T3 · attempt 1 · green-worker · 2026-09-22T00:59:52Z

### Mandate
Implement S1-06 T3 per `tasks.md` § T3 with the least change that makes exactly the 7 T3 tests in test/release/releaserc_test.go pass. T4 GREEN runs in parallel in another worktree on the same base (release.yml) — touch only your two files. T1/T2 must stay passing.

### Scope
#### May
- `.releaserc.json` and `CHANGELOG.md` (create) only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T3, `plan-ready.md` § T3, `validate.md` § T3. Chain base `chain/s1-06` @ c1d1499.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`.releaserc.json CHANGELOG.md`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S1-06/T4 · attempt 1 · green-worker · 2026-09-22T00:59:52Z

### Mandate
Implement S1-06 T4 per `tasks.md` § T4 with the least change that makes exactly the 11 T4 tests in test/release/workflow_test.go pass. T3 GREEN runs in parallel on the same base (.releaserc.json/CHANGELOG.md) — touch only release.yml. semantic-release then GoReleaser gated on new_release_published; GoReleaser job alone has id-token: write; only GITHUB_TOKEN; no "tap" substring anywhere. The orchestrator merges T3+T4 and runs the full matrix.

### Scope
#### May
- `.github/workflows/release.yml` (create) only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T4, `plan-ready.md` § T4, `validate.md` § T4. Chain base `chain/s1-06` @ c1d1499.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`.github/workflows/release.yml`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S1-06/review · attempt 1 · structural-reviewer · 2026-09-22T01:03:44Z

### Mandate
Read-only structural review of S1-06 as merged on stage-0 @ 57f1518 (.goreleaser.yaml, .releaserc.json, CHANGELOG.md, .github/workflows/release.yml, test/release/**). Check duplicate/parallel helpers within test/release (e.g. topLevelBlock/yamlBlock/findKeys overlap), drift between release.yml and ci.yml/commitlint.yml (pinning, permissions), cross-story contracts (S1-01 ldflags paths, S1-03 asset names), leftover markers.

### Scope
#### May
- Read; ctx-symbols/grep/go/actionlint read-only; append review block.
#### May Not
- Edit anything.

### Acceptance
verdict clean|isolated|foundation-poisoning; findings file:line; selfcheck (known _test.go false positive may be reported, not fixed).

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree.
- structural-reviewer: gate-structural-integrity misreports Go _test.go function-local duplicates as HIGH — verify scope first.

## S1-06/fix2 · attempt 1 · green-worker · 2026-09-22T01:39:59Z

### Mandate
Implement S1-06 fix2 per `tasks.md` § fix2 with the least change that makes exactly all S1-06 tests; TestGoreleaserCheck now RUNS (goreleaser 2.18.2 installed locally) and must PASS pass. Release run 35676441842 failed: goreleaser v2.5.1 (pinned in release.yml) rejects .goreleaser.yaml "line 29: field formats not found in type config.Archive" — `formats` needs a newer GoReleaser. Local `goreleaser check` with v2.18.2 validates the config. Bump the pin to v2.18.2 (a real release; matches local), nothing else. Commit type must be `fix:` so semantic-release publishes a patch release whose GoReleaser job uses the fixed pin. Full matrix runs.

### Scope
#### May
- `.github/workflows/release.yml` (edit: bump the pinned GoReleaser version) only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § fix2, `plan-ready.md` § fix2, `validate.md` § fix2. Chain base `chain/s1-06` @ 4d44590.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`.github/workflows/release.yml`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S1-06/fix3 · attempt 1 · green-worker · 2026-09-22T01:42:11Z

### Mandate
Implement S1-06 fix3 per `tasks.md` § fix3 with the least change that makes exactly all S1-06 tests, specifically TestChangelogSeededHeader and every TestReleaserc* test pass. master is red: the v1.0.0 release commit 4d44590 (semantic-release-bot, [skip ci]) regenerated CHANGELOG.md without a title, so TestChangelogSeededHeader fails. Restore the header now and configure changelogTitle so future releases keep it. Check the TestReleaserc* tests still accept the changelog plugin options. Commit type `fix:`. Full matrix runs. Base includes fix2 (goreleaser v2.18.2 pin).

### Scope
#### May
- `.releaserc.json` (edit: add `changelogTitle: "# Changelog"` to the @semantic-release/changelog options) and `CHANGELOG.md` (edit: restore the `# Changelog` header line at the very top, keeping the existing 1.0.0 release notes below it) only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § fix3, `plan-ready.md` § fix3, `validate.md` § fix3. Chain base `chain/s1-06` @ 61a1906.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`.releaserc.json CHANGELOG.md`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S1-06/author-fix · attempt 1 · red-worker · 2026-09-22T02:39:43Z

### Mandate
Human decision (2026-09-22): release commits must be authored by the user, not semantic-release-bot. Write ONE failing test in `test/release/workflow_test.go` (append; reuse helpers) named `TestWorkflowReleaseCommitsAuthoredByUser`: in `.github/workflows/release.yml`'s semantic-release step env, assert `GIT_AUTHOR_NAME` and `GIT_COMMITTER_NAME` equal "Adeel Ahmad" and `GIT_AUTHOR_EMAIL` and `GIT_COMMITTER_EMAIL` equal "adeelahmad99@gmail.com". Must fail by assertion now (release.yml has none of these). No shim needed.

### Acceptance
New test FAILs by assertion; the other test/release tests still pass (TestGoreleaserCheck runs — goreleaser installed); lint clean; commit has NO AI attribution trailers.

### Memory
- all: commit only as the configured git user; no Co-Authored-By / Claude-Session trailers.

## S1-06/author-fix · attempt 1 · green-worker · 2026-09-22T02:41:49Z

### Mandate
Add to the `semrel` step's `env:` in `.github/workflows/release.yml`: GIT_AUTHOR_NAME and GIT_COMMITTER_NAME = "Adeel Ahmad", GIT_AUTHOR_EMAIL and GIT_COMMITTER_EMAIL = "adeelahmad99@gmail.com", so semantic-release commits are authored by the user. Nothing else.

### Acceptance
All test/release tests PASS (incl. TestWorkflowReleaseCommitsAuthoredByUser); actionlint clean; full standards matrix green; commit has NO AI attribution trailers.
