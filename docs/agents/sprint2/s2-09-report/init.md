---
type: init
story: S2-09
---

## S2-09/T1 · attempt 1 · red-worker · 2026-09-22T03:38:38Z

### Mandate
Write every T1 test bullet in `plan.md` § T1 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. Evidence on stage-1 now: docs/reports/stage1/{pathtemplate,catalog,fidelity,crawler}-darwin.json and latency.json. The *-linux.json files will be copied in by the orchestrator from the PR #1 fuse-linux CI artifact; inventory tests may legitimately FAIL on their absence now. Report must stay honest: no go/no-go verdict (human decides).

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

## S2-09/T1 · attempt 1 · green-worker · 2026-09-22T03:42:32Z

### Mandate
Implement S2-09 T1 per `tasks.md` § T1 with the least change that makes exactly TestReportHasAllSectionsInOrder, TestEvidenceInventory, TestEvidenceFilesParse pass. All 9 evidence files are now committed in docs/reports/stage1/ (4 darwin, 4 linux from CI run 35683903691, latency.json). Create the report skeleton with the 9 section headings in order and the evidence inventory only (later tasks fill sections). No verdict/go-no-go. Run go test ./test/reports/.

### Scope
#### May
- docs/reports/stage1-measurements.md (create: skeleton + evidence inventory only) only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T1, `plan-ready.md` § T1, `validate.md` § T1. Chain base `chain/s2-09` @ 4564a77.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`docs/reports/stage1-measurements.md`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S2-09/T2 · attempt 1 · red-worker · 2026-09-22T03:44:57Z

### Mandate
Write every T2 test bullet in `plan.md` § T2 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. Sequential story (all tasks edit the same report). Report exists with 9 headings; evidence files all present.

### Scope
#### May
- Create the test files named for T2 in `tasks.md`: the T2 test file(s) named in plan.md § T2
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

## S2-09/T2 · attempt 1 · green-worker · 2026-09-22T03:47:18Z

### Mandate
Implement S2-09 T2 per `tasks.md` § T2 with the least change that makes exactly TestGoFuseVersionMatchesGoMod, TestPinnedVersionsTable (+ T1 tests) pass. Fill only the Pinned versions table from go.mod and the evidence JSON (go-fuse, restic 0.19.0, rclone v1.75.0, Go toolchain). Plain factual.

### Scope
#### May
- docs/reports/stage1-measurements.md (## Pinned versions only) only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T2, `plan-ready.md` § T2, `validate.md` § T2. Chain base `chain/s2-09` @ 5584da3.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`docs/reports/stage1-measurements.md`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S2-09/T3 · attempt 1 · red-worker · 2026-09-22T03:48:47Z

### Mandate
Write every T3 test bullet in `plan.md` § T3 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. Sequential story. Evidence: docs/reports/stage1/{pathtemplate,catalog,fidelity}-{darwin,linux}.json. Note fidelity symlink contract ruling (T5b): symlink size/linktarget not reported by restic ls --json 0.19 — the report must say so.

### Scope
#### May
- Create the test files named for T3 in `tasks.md`: the T3 test file(s) named in plan.md § T3
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

## S2-09/T3 · attempt 1 · green-worker · 2026-09-22T03:51:44Z

### Mandate
Implement S2-09 T3 per `tasks.md` § T3 with the least change that makes exactly the 5 T3 tests (+ T1, T2) pass. Fill Path-template, Catalog, Metadata fidelity sections from the darwin+linux JSON with the row formats in tasks.md. In Metadata fidelity also state plainly: restic ls --json 0.19 does not report symlink size or link target; the symlink was compared on mode and mtime, and its target was checked through the catalog alias. ctime/birth time recorded, not claimed.

### Scope
#### May
- docs/reports/stage1-measurements.md (three platform sections only) only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T3, `plan-ready.md` § T3, `validate.md` § T3. Chain base `chain/s2-09` @ ba7eb45.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`docs/reports/stage1-measurements.md`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S2-09/T4 · attempt 1 · red-worker · 2026-09-22T03:53:45Z

### Mandate
Write every T4 test bullet in `plan.md` § T4 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. Sequential story. latency.json values: cold_listing 32.66 ms n=1, warm_prewarmed_listing median 1.436 n=3, cold_first_file_read 3012.38 n=1, warm_listing_after_restart median 6.385 n=3, remote_deleted true. No verdict/threshold words (human decides).

### Scope
#### May
- Create the test files named for T4 in `tasks.md`: the T4 test file(s) named in plan.md § T4
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
