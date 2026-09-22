---
type: init
story: S2-06
---

## S2-06/T1 · attempt 1 · red-worker · 2026-09-22T03:11:06Z

### Mandate
Write every T1 test bullet in `plan.md` § T1 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. Wave 3 parallel REDs on the same base; each owns its own test file + shim zz_agentic_shim_<task>.go; declare ONLY your task's symbols. Pure unit tests (no mount, no restic run, no network) unless plan.md says otherwise.

### Scope
#### May
- Create the test files named for T1 in `tasks.md`: internal/compat/fidelity/compare_test.go (+ internal/compat/fidelity/zz_agentic_shim_t1.go)
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

## S2-06/T2 · attempt 1 · red-worker · 2026-09-22T03:11:07Z

### Mandate
Write every T2 test bullet in `plan.md` § T2 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. Wave 3 parallel REDs on the same base; each owns its own test file + shim zz_agentic_shim_<task>.go; declare ONLY your task's symbols. Pure unit tests (no mount, no restic run, no network) unless plan.md says otherwise. Do not declare T1 symbols (Meta, Result, compare…).

### Scope
#### May
- Create the test files named for T2 in `tasks.md`: internal/compat/fidelity/resticls_test.go (+ internal/compat/fidelity/zz_agentic_shim_t2.go)
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

## S2-06/T3 · attempt 1 · red-worker · 2026-09-22T03:16:13Z

### Mandate
Write every T3 test bullet in `plan.md` § T3 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. Parallel with a scaffolder converting the T1/T2 shims into production files and with the other RED; declare ONLY your task's symbols in your own shim; use existing chain symbols (Meta, Result, …) as declared by the T1/T2 shims.

### Scope
#### May
- Create the test files named for T3 in `tasks.md`: internal/compat/fidelity/fixtures_test.go and internal/compat/fidelity/observe_test.go (+ zz_agentic_shim_t3.go)
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

## S2-06/T4 · attempt 1 · red-worker · 2026-09-22T03:16:13Z

### Mandate
Write every T4 test bullet in `plan.md` § T4 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. Parallel with a scaffolder converting the T1/T2 shims into production files and with the other RED; declare ONLY your task's symbols in your own shim; use existing chain symbols (Meta, Result, …) as declared by the T1/T2 shims.

### Scope
#### May
- Create the test files named for T4 in `tasks.md`: internal/compat/fidelity/evidence_test.go and internal/compat/fidelity/prereq_test.go (+ zz_agentic_shim_t4.go)
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

## S2-06/T1+T2 · attempt 1 · green-worker · 2026-09-22T03:19:33Z

### Mandate
Implement S2-06 T1+T2 per `tasks.md` § T1+T2 with the least change that makes exactly the 10 T1 comparator tests and 4 T2 parser tests pass. Run only T1+T2 tests by exact name (T3 shim present; T4 RED in parallel). MTimeTolerance stays exactly 'const MTimeTolerance time.Duration = 0'.

### Scope
#### May
- internal/compat/fidelity/compare.go and internal/compat/fidelity/resticls.go only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T1+T2, `plan-ready.md` § T1+T2, `validate.md` § T1+T2. Chain base `chain/s2-06` @ 96c5c26.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`internal/compat/fidelity/compare.go internal/compat/fidelity/resticls.go`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S2-06/T3 · attempt 1 · green-worker · 2026-09-22T03:24:10Z

### Mandate
Implement S2-06 T3 per `tasks.md` § T3 with the least change that makes exactly the 5 T3 tests pass. Parallel GREENs: T3 owns fixtures.go/observe*.go, T4 owns evidence.go/prereq.go. Run only your tests by exact name. Observe uses os.Lstat (never os.Stat in observe*.go) and calls statTimes for ctime/birth (darwin: birth from Birthtimespec; linux: nil birth). GOOS=linux go vet must pass.

### Scope
#### May
- internal/compat/fidelity/fixtures.go, observe.go, observe_darwin.go, observe_linux.go only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T3, `plan-ready.md` § T3, `validate.md` § T3. Chain base `chain/s2-06` @ 53c16ae.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`internal/compat/fidelity/fixtures.go internal/compat/fidelity/observe.go internal/compat/fidelity/observe_darwin.go internal/compat/fidelity/observe_linux.go`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S2-06/T4 · attempt 1 · green-worker · 2026-09-22T03:24:10Z

### Mandate
Implement S2-06 T4 per `tasks.md` § T4 with the least change that makes exactly the 6 T4 tests pass. Parallel GREENs: T3 owns fixtures.go/observe*.go, T4 owns evidence.go/prereq.go. Run only your tests by exact name. Unclaimed always marshals claimed:false; no ctime_ok/birth_time_ok anywhere in production code.

### Scope
#### May
- internal/compat/fidelity/evidence.go and prereq.go only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T4, `plan-ready.md` § T4, `validate.md` § T4. Chain base `chain/s2-06` @ 53c16ae.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`internal/compat/fidelity/evidence.go internal/compat/fidelity/prereq.go`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S2-06/T5 · attempt 1 · red-worker · 2026-09-22T03:27:04Z

### Mandate
Write every T5 test bullet in `plan.md` § T5 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. Last task of S2-06. //go:build integration; skip via MissingPrereq exactly as tasks.md § T5. T1-T4 are implemented and S2-03/S2-04 are merged, so after writing it run it for real on this Mac (restic 0.19 + macFUSE): SNAPBACK_FUSE_TESTS=1 SNAPBACK_EVIDENCE_DIR=$(mktemp -d) GOTOOLCHAIN=auto go test -race -tags=integration -run <name> -v ./internal/compat/fidelity/. Disposable repo in t.TempDir, generated data only, no real remote. If it PASSES report PASS-ON-RED (do not weaken; mtime tolerance stays 0 — a real mismatch is a legitimate FAIL to report, not to hide); if it FAILS report the exact assertion and the observed deltas. Always Stop the restic mount and Unmount the catalog in t.Cleanup; verify no mount is left. Do not commit evidence JSON.

### Scope
#### May
- Create the test files named for T5 in `tasks.md`: internal/compat/fidelity/fidelity_integration_test.go only
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

## S2-06/T5b · attempt 1 · red-worker · 2026-09-22T03:30:46Z

### Mandate
Write every T5b test bullet in `plan.md` § T5b at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. CONTRACT RULING (orchestrator, auto-approve): restic 0.19 'ls --json' omits size and linktarget for symlinks (verified by the T5 RED). For symlink entries: compare ONLY mode and mtime against ls --json (tolerance stays 0); check the link target observed through the catalog alias equals the generator's target (small.txt); do NOT compare symlink size against ls --json. Regular files unchanged (size, mode, mtime vs both generator and ls --json). The evidence must record, for the symlink, that size/linktarget are 'not reported by restic ls --json 0.19' (honest, not a pass). Amend TestFidelityThroughSnapshotAlias accordingly, run it for real (restic + macFUSE); expected PASS-ON-RED — report the full observed table. No mount left behind; do not commit evidence.

### Scope
#### May
- Create the test files named for T5b in `tasks.md`: internal/compat/fidelity/fidelity_integration_test.go only
- Test helpers that a T5b test itself exercises (e.g. `repoRoot`, `readRepoFile`, `yamlBlock`, `runInstaller`) go in a marked shim `zz_agentic_shim_test.go` in the same test package (first line `// agentic:shim`) with deliberately WRONG bodies so those tests fail by assertion; helpers NOT under test may be real and live in the named helpers file.
#### May Not
- Create or edit any non-test artifact (workflow YAML, JSON config, install.sh, .goreleaser.yaml, SPEC.md/README.md, go.mod) — those are GREEN's job; touch other stories' files; suppress/skip tests (except the tool-absent `t.Skip` cases plan.md explicitly allows).

### Inputs
- `plan.md` § T5b, `tasks.md` § T5b (contracts), `validate.md` § T5b. Chain base: the story chain branch named in your prompt (stage-1 @ 2158ade = master + sprint2 plan).

### Acceptance
`go test ./test/...` for this story's package compiles; every new test FAILS by assertion; `go vet ./...` clean; diff vs BASE_REF = only this task's test files (+ shim); output.md block appended; selfcheck PASS.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md in first); orchestrator relays output.md back.
- red-worker: make shim bodies differ so comparison tests cannot pass by accident.
- all: errcheck flags unchecked writes — use explicit `_, _ =` discard, never nolint.

## S2-06/T5 · attempt 1 · green-worker · 2026-09-22T03:33:44Z

### Mandate
Implement S2-06 T5 per `tasks.md` § T5 with the least change that makes exactly TestFidelityThroughSnapshotAlias (integration) + full default-build matrix pass. LAST task of S2-06. RED (T5 + T5b contract ruling) accepted; no production change expected. mkdir -p docs/reports/stage1, then SNAPBACK_FUSE_TESTS=1 SNAPBACK_EVIDENCE_DIR=$(pwd)/docs/reports/stage1 GOTOOLCHAIN=auto go test -race -count=1 -tags=integration -run TestFidelityThroughSnapshotAlias -v ./internal/compat/fidelity/; confirm pass and no leftover mount; commit ONLY the generated fidelity-darwin.json. Then full standards matrix incl. coverage >=80%.

### Scope
#### May
- docs/reports/stage1/fidelity-darwin.json (generated) only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T5, `plan-ready.md` § T5, `validate.md` § T5. Chain base `chain/s2-06` @ 5675db1.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`docs/reports/stage1/fidelity-darwin.json`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S2-06/fix-gate · attempt 1 · red-worker · 2026-09-22T04:01:46Z

### Mandate
Write every fix-gate test bullet in `plan.md` § fix-gate at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. Structural review finding: fidelity.MissingPrereq treats any non-empty SNAPBACK_FUSE_TESTS (e.g. "0") as enabled, while resticfx.MissingPrerequisite and the crawler test require exactly "1". Add a table row/test: SNAPBACK_FUSE_TESTS="0" (and "yes") must return the same skip message as unset. Must FAIL BY ASSERTION on current prereq.go. Tests only.

### Scope
#### May
- Create the test files named for fix-gate in `tasks.md`: internal/compat/fidelity/prereq_test.go only
- Test helpers that a fix-gate test itself exercises (e.g. `repoRoot`, `readRepoFile`, `yamlBlock`, `runInstaller`) go in a marked shim `zz_agentic_shim_test.go` in the same test package (first line `// agentic:shim`) with deliberately WRONG bodies so those tests fail by assertion; helpers NOT under test may be real and live in the named helpers file.
#### May Not
- Create or edit any non-test artifact (workflow YAML, JSON config, install.sh, .goreleaser.yaml, SPEC.md/README.md, go.mod) — those are GREEN's job; touch other stories' files; suppress/skip tests (except the tool-absent `t.Skip` cases plan.md explicitly allows).

### Inputs
- `plan.md` § fix-gate, `tasks.md` § fix-gate (contracts), `validate.md` § fix-gate. Chain base: the story chain branch named in your prompt (stage-1 @ 2158ade = master + sprint2 plan).

### Acceptance
`go test ./test/...` for this story's package compiles; every new test FAILS by assertion; `go vet ./...` clean; diff vs BASE_REF = only this task's test files (+ shim); output.md block appended; selfcheck PASS.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md in first); orchestrator relays output.md back.
- red-worker: make shim bodies differ so comparison tests cannot pass by accident.
- all: errcheck flags unchecked writes — use explicit `_, _ =` discard, never nolint.
