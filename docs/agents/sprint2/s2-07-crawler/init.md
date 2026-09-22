---
type: init
story: S2-07
---

## S2-07/T1 · attempt 1 · red-worker · 2026-09-22T03:11:07Z

### Mandate
Write every T1 test bullet in `plan.md` § T1 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. Wave 3 parallel REDs on the same base; each owns its own test file + shim zz_agentic_shim_<task>.go; declare ONLY your task's symbols. Pure unit tests (no mount, no restic run, no network) unless plan.md says otherwise. Use the merged mount.Observer/Event/Op contract exactly.

### Scope
#### May
- Create the test files named for T1 in `tasks.md`: internal/compat/crawler/counter_test.go (+ internal/compat/crawler/zz_agentic_shim_t1.go)
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

## S2-07/T2 · attempt 1 · red-worker · 2026-09-22T03:11:07Z

### Mandate
Write every T2 test bullet in `plan.md` § T2 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. Wave 3 parallel REDs on the same base; each owns its own test file + shim zz_agentic_shim_<task>.go; declare ONLY your task's symbols. Pure unit tests (no mount, no restic run, no network) unless plan.md says otherwise. Declare the Status type here (T3 imports it later).

### Scope
#### May
- Create the test files named for T2 in `tasks.md`: internal/compat/crawler/tools_test.go (+ internal/compat/crawler/zz_agentic_shim_t2.go)
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

## S2-07/T4 · attempt 1 · red-worker · 2026-09-22T03:11:07Z

### Mandate
Write every T4 test bullet in `plan.md` § T4 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. Wave 3 parallel REDs on the same base; each owns its own test file + shim zz_agentic_shim_<task>.go; declare ONLY your task's symbols. Pure unit tests (no mount, no restic run, no network) unless plan.md says otherwise. Do not declare T1/T2/T3 symbols.

### Scope
#### May
- Create the test files named for T4 in `tasks.md`: internal/compat/crawler/seed_test.go (+ internal/compat/crawler/zz_agentic_shim_t4.go)
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

## S2-07/T3 · attempt 1 · red-worker · 2026-09-22T03:16:13Z

### Mandate
Write every T3 test bullet in `plan.md` § T3 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. Parallel with a scaffolder converting the T1/T2 shims into production files and with the other RED; declare ONLY your task's symbols in your own shim; use existing chain symbols (Meta, Result, …) as declared by the T1/T2 shims. Import T2's Status type from the chain; the VS Code row is a FixedRow (T2 decision) that T3 may convert.

### Scope
#### May
- Create the test files named for T3 in `tasks.md`: internal/compat/crawler/evidence_test.go (+ zz_agentic_shim_t3.go)
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

## S2-07/T1 · attempt 1 · green-worker · 2026-09-22T03:19:33Z

### Mandate
Implement S2-07 T1 per `tasks.md` § T1 with the least change that makes exactly the 5 T1 counter tests pass. Parallel GREENs own disjoint files; run only your tests by exact name (sibling stubs panic). Counter must be safe for concurrent Observe (sync.Mutex or atomics); ByOp returns a copy; zero value usable.

### Scope
#### May
- internal/compat/crawler/counter.go only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T1, `plan-ready.md` § T1, `validate.md` § T1. Chain base `chain/s2-07` @ 2e01ad8.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`internal/compat/crawler/counter.go`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S2-07/T2 · attempt 1 · green-worker · 2026-09-22T03:19:33Z

### Mandate
Implement S2-07 T2 per `tasks.md` § T2 with the least change that makes exactly the 8 T2 tool tests pass. Parallel GREENs own disjoint files; run only your tests by exact name (sibling stubs panic).

### Scope
#### May
- internal/compat/crawler/tools.go only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T2, `plan-ready.md` § T2, `validate.md` § T2. Chain base `chain/s2-07` @ 2e01ad8.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`internal/compat/crawler/tools.go`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S2-07/T4 · attempt 1 · green-worker · 2026-09-22T03:19:33Z

### Mandate
Implement S2-07 T4 per `tasks.md` § T4 with the least change that makes exactly the 3 T4 seed tests pass. Parallel GREENs own disjoint files; run only your tests by exact name (sibling stubs panic). Refuse roots outside os.TempDir().

### Scope
#### May
- internal/compat/crawler/seed.go only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T4, `plan-ready.md` § T4, `validate.md` § T4. Chain base `chain/s2-07` @ 2e01ad8.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`internal/compat/crawler/seed.go`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S2-07/T3 · attempt 1 · green-worker · 2026-09-22T03:23:47Z

### Mandate
Implement S2-07 T3 per `tasks.md` § T3 with the least change that makes exactly the 8 T3 evidence tests pass. Add snake_case JSON tags to Row/Report as the scaffold recipe lists (hits as *int, omitted/null when not set). Keep StatusNotTestedHere (T2's VSCodeRow uses it). Whole crawler package must pass -race; lint clean.

### Scope
#### May
- internal/compat/crawler/evidence.go only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T3, `plan-ready.md` § T3, `validate.md` § T3. Chain base `chain/s2-07` @ 43cf0aa.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`internal/compat/crawler/evidence.go`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S2-07/T5 · attempt 1 · red-worker · 2026-09-22T03:26:05Z

### Mandate
Write every T5 test bullet in `plan.md` § T5 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. Last task of S2-07. //go:build integration, gating exactly as tasks.md § T5. T1-T4 are implemented and S2-04's adapter is merged, so after writing it run it for real on this Mac: SNAPBACK_FUSE_TESTS=1 SNAPBACK_EVIDENCE_DIR=$(mktemp -d) GOTOOLCHAIN=auto go test -race -tags=integration -run <names> -v ./internal/compat/crawler/ (macFUSE installed; rg/fd/find/rsync from PATH, missing tools must produce not-tested-here rows, never skip-as-pass). If it PASSES report PASS-ON-RED (do not weaken); if it FAILS report the exact assertion. Never leave a mount behind (umount the temp mountpoint on failure). Do not commit evidence JSON.

### Scope
#### May
- Create the test files named for T5 in `tasks.md`: internal/compat/crawler/crawler_integration_test.go only
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

## S2-07/T5 · attempt 1 · green-worker · 2026-09-22T03:29:54Z

### Mandate
Implement S2-07 T5 per `tasks.md` § T5 with the least change that makes exactly TestCrawlerPositiveControl, TestCrawlerToolMatrix (integration) + full default-build matrix pass. LAST task of S2-07. RED was PASS-ON-RED (accepted): no production change expected. mkdir -p docs/reports/stage1, then SNAPBACK_FUSE_TESTS=1 SNAPBACK_EVIDENCE_DIR=$(pwd)/docs/reports/stage1 GOTOOLCHAIN=auto go test -race -count=1 -tags=integration -run 'TestCrawler' -v ./internal/compat/crawler/; confirm no leftover mount; commit ONLY the generated crawler-darwin.json. Then full standards matrix incl. coverage >=80%.

### Scope
#### May
- docs/reports/stage1/crawler-darwin.json (generated) only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T5, `plan-ready.md` § T5, `validate.md` § T5. Chain base `chain/s2-07` @ 46a91cf.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`docs/reports/stage1/crawler-darwin.json`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.
