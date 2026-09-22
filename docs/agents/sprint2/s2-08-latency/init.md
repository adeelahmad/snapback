---
type: init
story: S2-08
---

## S2-08/T1 guard · attempt 1 · red-worker · 2026-09-22T02:50:22Z

### Mandate
Write every T1 guard test bullet in `plan.md` § T1 guard at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. Package internal/compat/latency is new. Pure unit tests only — no network, no rclone/restic execution, no real remote. Other S2-08 REDs run in parallel with their own shims (zz_agentic_shim_t1/t2/t3.go); declare in your shim ONLY the symbols your task's tests reference and that tasks.md assigns to your task.

### Scope
#### May
- Create the test files named for T1 guard in `tasks.md`: internal/compat/latency/_test.go (+ internal/compat/latency/zz_agentic_shim_t1 guard.go)
- Test helpers that a T1 guard test itself exercises (e.g. `repoRoot`, `readRepoFile`, `yamlBlock`, `runInstaller`) go in a marked shim `zz_agentic_shim_test.go` in the same test package (first line `// agentic:shim`) with deliberately WRONG bodies so those tests fail by assertion; helpers NOT under test may be real and live in the named helpers file.
#### May Not
- Create or edit any non-test artifact (workflow YAML, JSON config, install.sh, .goreleaser.yaml, SPEC.md/README.md, go.mod) — those are GREEN's job; touch other stories' files; suppress/skip tests (except the tool-absent `t.Skip` cases plan.md explicitly allows).

### Inputs
- `plan.md` § T1 guard, `tasks.md` § T1 guard (contracts), `validate.md` § T1 guard. Chain base: the story chain branch named in your prompt (stage-1 @ 2158ade = master + sprint2 plan).

### Acceptance
`go test ./test/...` for this story's package compiles; every new test FAILS by assertion; `go vet ./...` clean; diff vs BASE_REF = only this task's test files (+ shim); output.md block appended; selfcheck PASS.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md in first); orchestrator relays output.md back.
- red-worker: make shim bodies differ so comparison tests cannot pass by accident.
- all: errcheck flags unchecked writes — use explicit `_, _ =` discard, never nolint.

## S2-08/T2 result · attempt 1 · red-worker · 2026-09-22T02:50:22Z

### Mandate
Write every T2 result test bullet in `plan.md` § T2 result at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. Package internal/compat/latency is new. Pure unit tests only — no network, no rclone/restic execution, no real remote. Other S2-08 REDs run in parallel with their own shims (zz_agentic_shim_t1/t2/t3.go); declare in your shim ONLY the symbols your task's tests reference and that tasks.md assigns to your task.

### Scope
#### May
- Create the test files named for T2 result in `tasks.md`: internal/compat/latency/_test.go (+ internal/compat/latency/zz_agentic_shim_t2 result.go)
- Test helpers that a T2 result test itself exercises (e.g. `repoRoot`, `readRepoFile`, `yamlBlock`, `runInstaller`) go in a marked shim `zz_agentic_shim_test.go` in the same test package (first line `// agentic:shim`) with deliberately WRONG bodies so those tests fail by assertion; helpers NOT under test may be real and live in the named helpers file.
#### May Not
- Create or edit any non-test artifact (workflow YAML, JSON config, install.sh, .goreleaser.yaml, SPEC.md/README.md, go.mod) — those are GREEN's job; touch other stories' files; suppress/skip tests (except the tool-absent `t.Skip` cases plan.md explicitly allows).

### Inputs
- `plan.md` § T2 result, `tasks.md` § T2 result (contracts), `validate.md` § T2 result. Chain base: the story chain branch named in your prompt (stage-1 @ 2158ade = master + sprint2 plan).

### Acceptance
`go test ./test/...` for this story's package compiles; every new test FAILS by assertion; `go vet ./...` clean; diff vs BASE_REF = only this task's test files (+ shim); output.md block appended; selfcheck PASS.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md in first); orchestrator relays output.md back.
- red-worker: make shim bodies differ so comparison tests cannot pass by accident.
- all: errcheck flags unchecked writes — use explicit `_, _ =` discard, never nolint.

## S2-08/T3 scratch · attempt 1 · red-worker · 2026-09-22T02:50:22Z

### Mandate
Write every T3 scratch test bullet in `plan.md` § T3 scratch at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. Package internal/compat/latency is new. Pure unit tests only — no network, no rclone/restic execution, no real remote. Other S2-08 REDs run in parallel with their own shims (zz_agentic_shim_t1/t2/t3.go); declare in your shim ONLY the symbols your task's tests reference and that tasks.md assigns to your task.

### Scope
#### May
- Create the test files named for T3 scratch in `tasks.md`: internal/compat/latency/_test.go (+ internal/compat/latency/zz_agentic_shim_t3 scratch.go)
- Test helpers that a T3 scratch test itself exercises (e.g. `repoRoot`, `readRepoFile`, `yamlBlock`, `runInstaller`) go in a marked shim `zz_agentic_shim_test.go` in the same test package (first line `// agentic:shim`) with deliberately WRONG bodies so those tests fail by assertion; helpers NOT under test may be real and live in the named helpers file.
#### May Not
- Create or edit any non-test artifact (workflow YAML, JSON config, install.sh, .goreleaser.yaml, SPEC.md/README.md, go.mod) — those are GREEN's job; touch other stories' files; suppress/skip tests (except the tool-absent `t.Skip` cases plan.md explicitly allows).

### Inputs
- `plan.md` § T3 scratch, `tasks.md` § T3 scratch (contracts), `validate.md` § T3 scratch. Chain base: the story chain branch named in your prompt (stage-1 @ 2158ade = master + sprint2 plan).

### Acceptance
`go test ./test/...` for this story's package compiles; every new test FAILS by assertion; `go vet ./...` clean; diff vs BASE_REF = only this task's test files (+ shim); output.md block appended; selfcheck PASS.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md in first); orchestrator relays output.md back.
- red-worker: make shim bodies differ so comparison tests cannot pass by accident.
- all: errcheck flags unchecked writes — use explicit `_, _ =` discard, never nolint.

## S2-08/T4 · attempt 1 · red-worker · 2026-09-22T02:53:21Z

### Mandate
Write every T4 test bullet in `plan.md` § T4 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. Declare the package's Runner interface in your shim with a method set that the merged resticfx runner (internal/compat/resticfx/runner.go) satisfies — read it and mirror its signatures; tests use a fake Runner. Do not redeclare T1-T3 symbols (AllowedRemote, CheckRemote, Summarize, Result, newScratch …) — they exist on the chain. No network, no rclone execution.

### Scope
#### May
- Create the test files named for T4 in `tasks.md`: internal/compat/latency/cleanup_test.go (+ internal/compat/latency/zz_agentic_shim_t4.go)
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

## S2-08/T1 · attempt 1 · green-worker · 2026-09-22T02:56:26Z

### Mandate
Implement S2-08 T1 per `tasks.md` § T1 with the least change that makes exactly TestCheckRemoteAllowsOnlyExactRemote, TestRemoteFromEnv, TestRepoSpecIsConstant pass. Parallel GREENs T1/T2/T3 each own one file (guard.go / result.go / scratch.go) — touch only yours. Run only your task's tests by exact name (panics in sibling stubs abort whole-package runs), then golangci-lint on the package. TestNoThresholdSymbolsInPackage must stay passing (no threshold/verdict/pass/fail/go/ok/acceptable words in identifiers or strings).

### Scope
#### May
- internal/compat/latency/guard.go only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T1, `plan-ready.md` § T1, `validate.md` § T1. Chain base `chain/s2-08` @ 891548e.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`internal/compat/latency/guard.go`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S2-08/T2 · attempt 1 · green-worker · 2026-09-22T02:56:26Z

### Mandate
Implement S2-08 T2 per `tasks.md` § T2 with the least change that makes exactly the 8 T2 tests in result_test.go pass. Parallel GREENs T1/T2/T3 each own one file (guard.go / result.go / scratch.go) — touch only yours. Run only your task's tests by exact name (panics in sibling stubs abort whole-package runs), then golangci-lint on the package. TestNoThresholdSymbolsInPackage must stay passing (no threshold/verdict/pass/fail/go/ok/acceptable words in identifiers or strings). Encode must convert the timestamp to UTC.

### Scope
#### May
- internal/compat/latency/result.go only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T2, `plan-ready.md` § T2, `validate.md` § T2. Chain base `chain/s2-08` @ 891548e.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`internal/compat/latency/result.go`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S2-08/T3 · attempt 1 · green-worker · 2026-09-22T02:56:26Z

### Mandate
Implement S2-08 T3 per `tasks.md` § T3 with the least change that makes exactly the 4 T3 tests in scratch_test.go pass. Parallel GREENs T1/T2/T3 each own one file (guard.go / result.go / scratch.go) — touch only yours. Run only your task's tests by exact name (panics in sibling stubs abort whole-package runs), then golangci-lint on the package. TestNoThresholdSymbolsInPackage must stay passing (no threshold/verdict/pass/fail/go/ok/acceptable words in identifiers or strings). Password: >=32 random bytes (crypto/rand) hex-encoded, 0600, inside os.MkdirTemp under os.TempDir(); never logged.

### Scope
#### May
- internal/compat/latency/scratch.go only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T3, `plan-ready.md` § T3, `validate.md` § T3. Chain base `chain/s2-08` @ 891548e.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`internal/compat/latency/scratch.go`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S2-08/T5 · attempt 1 · red-worker · 2026-09-22T02:58:16Z

### Mandate
Write every T5 test bullet in `plan.md` § T5 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. Use the existing chain symbols (Runner from the T4 shim/cleanup.go, Result/Stats/Summarize/Encode, CheckRemote, newScratch). Shim only T5 symbols (Run, Config, Clock, Samples/DataFileCount/DataFileBytes constants and any mount-supervisor method the tests need). If Runner needs a new method for the long-running mount, do NOT edit the T4 shim — report it as a Runner-shape change for the orchestrator. Scripted fake Runner + fake Clock; no network, no real rclone/restic. The shim's Run must return a wrong Result WITHOUT calling the runner so tests fail by assertion.

### Scope
#### May
- Create the test files named for T5 in `tasks.md`: internal/compat/latency/run_test.go (+ internal/compat/latency/zz_agentic_shim_t5.go)
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

## S2-08/T4 · attempt 1 · green-worker · 2026-09-22T02:58:38Z

### Mandate
Implement S2-08 T4 per `tasks.md` § T4 with the least change that makes exactly TestVerifyDeleted, TestPurgeAndVerifyCommands, TestPurgeErrorStillVerifies, TestCheckEmptyRefusesNonEmptyRemote pass. Targets built only from AllowedRemote, never from input; args arrays, no shell. A T5 RED runs in parallel (run_test.go) — do not touch it.

### Scope
#### May
- internal/compat/latency/cleanup.go only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T4, `plan-ready.md` § T4, `validate.md` § T4. Chain base `chain/s2-08` @ 6c19b46.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`internal/compat/latency/cleanup.go`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S2-08/T5 · attempt 1 · scaffolder · 2026-09-22T03:03:03Z

### Mandate
Move `internal/compat/latency/zz_agentic_shim_t5.go` symbols into `internal/compat/latency/run.go`: Run, Config (Remote, Runner, Mounter, Clock — orchestrator accepted the Mounter/Mounted interfaces the RED introduced, since a blocking Runner.Run cannot hold the long-running mount), Clock, Mounter, Mounted, Samples/DataFileCount/DataFileBytes constants. Function bodies `panic("SUB-AGENT-TODO: <recipe from tasks.md § T5>")`. Delete the shim; each symbol once.

### Acceptance
`go vet ./...` compiles; T1-T4 tests PASS; TestNoThresholdSymbolsInPackage PASS; no shim marker; gofmt clean; commit `chore: scaffold S2-08 run stubs`, no AI trailers.

## S2-08/T6 · attempt 1 · red-worker · 2026-09-22T03:03:03Z

### Mandate
Write every T6 test bullet in `plan.md` § T6 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. New package main under tools/stage1latency. Tests inject fake latency deps (no real rclone/restic, no network, never the real gdrive remote). Shim run() with wrong behaviour (e.g. always returns 3, writes nothing). TestMainIsThin reads main.go — if main.go does not exist yet, that test must fail by assertion (t.Fatal on missing file), not compile error; do NOT create main.go/deps.go/run.go. Use the existing latency package API on the chain (Run, Config{Remote,Runner,Mounter,Clock}, Result, Encode, RemoteFromEnv). A parallel scaffolder converts the latency T5 shim into run.go — do not touch internal/compat/latency.

### Scope
#### May
- Create the test files named for T6 in `tasks.md`: tools/stage1latency/run_test.go (+ tools/stage1latency/zz_agentic_shim_t6.go)
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

## S2-08/T5 · attempt 1 · green-worker · 2026-09-22T03:04:59Z

### Mandate
Implement S2-08 T5 per `tasks.md` § T5 with the least change that makes exactly the 10 TestRun* tests in run_test.go pass. Order per tasks.md § T5 (guard -> checkEmpty -> scratch -> versions -> init -> backup -> snapshots -> mount -> cold list -> prewarm restic ls --json <fullID> -> warm list xSamples -> cold first-file read -> unmount -> remount same cache -> warm list after restart -> unmount -> purgeAndVerify in a deferred cleanup that runs even after mid-run failure; scratch removed always). Password only via --password-file; never in args, logs or Result. No threshold/verdict fields. Uses the fake Runner/Mounter/Clock in tests; never real network. A T6 RED runs in parallel under tools/ — do not touch it.

### Scope
#### May
- internal/compat/latency/run.go only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T5, `plan-ready.md` § T5, `validate.md` § T5. Chain base `chain/s2-08` @ 72df2f8.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`internal/compat/latency/run.go`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.


## S2-08/T6 · attempt 1 · scaffolder · 2026-09-22T03:06:13Z

### Mandate
Replace `tools/stage1latency/zz_agentic_shim_t6.go` with `tools/stage1latency/main.go` (exactly `func main() { os.Exit(run(os.Args[1:], os.Getenv, os.Stdout, os.Stderr)) }` plus package/imports — real, not a stub, <=12 non-empty lines), `tools/stage1latency/deps.go` (`var newConfig func(remote string) latency.Config` — keep the variable; its default value is a func literal whose body is `panic("SUB-AGENT-TODO: wire resticfx runner, mounter and time.Now clock")`), `tools/stage1latency/run.go` (`run(args []string, getenv func(string) string, stdout, stderr io.Writer) int` with body `panic("SUB-AGENT-TODO: <recipe from tasks.md § T6>")`). Delete the shim.

### Acceptance
`go vet ./...` compiles; TestMainIsThin PASSES (main is real); the other T6 tests fail (panic acceptable); gofmt clean; no shim marker; commit `chore: scaffold S2-08 CLI stubs`, no AI trailers.

## S2-08/fix-id · attempt 1 · red-worker · 2026-09-22T03:07:25Z

### Mandate
Write every fix-id test bullet in `plan.md` § fix-id at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. Follow-up to T5 GREEN: testSnapshotID (run_test.go:17) has 63 hex chars, not 64, which forced GREEN to accept any non-empty id. (1) Correct testSnapshotID to exactly 64 lowercase hex chars (all existing T5 tests must still PASS). (2) Add TestRunRejectsShortSnapshotID: the scripted snapshots reply returns an 8-char id (like a short_id); assert Run returns a non-nil error mentioning the snapshot id AND cleanup still ran (purge + lsf recorded, scratch removed). It must FAIL BY ASSERTION on current run.go. No production code. A T6 scaffolder works in tools/ in parallel.

### Scope
#### May
- Create the test files named for fix-id in `tasks.md`: internal/compat/latency/run_test.go only
- Test helpers that a fix-id test itself exercises (e.g. `repoRoot`, `readRepoFile`, `yamlBlock`, `runInstaller`) go in a marked shim `zz_agentic_shim_test.go` in the same test package (first line `// agentic:shim`) with deliberately WRONG bodies so those tests fail by assertion; helpers NOT under test may be real and live in the named helpers file.
#### May Not
- Create or edit any non-test artifact (workflow YAML, JSON config, install.sh, .goreleaser.yaml, SPEC.md/README.md, go.mod) — those are GREEN's job; touch other stories' files; suppress/skip tests (except the tool-absent `t.Skip` cases plan.md explicitly allows).

### Inputs
- `plan.md` § fix-id, `tasks.md` § fix-id (contracts), `validate.md` § fix-id. Chain base: the story chain branch named in your prompt (stage-1 @ 2158ade = master + sprint2 plan).

### Acceptance
`go test ./test/...` for this story's package compiles; every new test FAILS by assertion; `go vet ./...` clean; diff vs BASE_REF = only this task's test files (+ shim); output.md block appended; selfcheck PASS.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md in first); orchestrator relays output.md back.
- red-worker: make shim bodies differ so comparison tests cannot pass by accident.
- all: errcheck flags unchecked writes — use explicit `_, _ =` discard, never nolint.

## S2-08/T6 · attempt 1 · green-worker · 2026-09-22T03:08:28Z

### Mandate
Implement S2-08 T6 per `tasks.md` § T6 with the least change that makes exactly the 5 T6 tests in tools/stage1latency/run_test.go pass. Last planned task of S2-08: full matrix + coverage >=80%. run.go per tasks.md § T6 (exit 2 on missing -out / refused remote with stderr naming the flag / gdrive:snapback-stage1; exit 1 + loud 'NOT DELETED' stderr + JSON still written when remote_deleted false; exit 0 only on success with remote_deleted true). deps.go: default newConfig wires the real resticfx runner (Runner), a Mounter adapter over resticfx StartMount/WaitReady/Stop with os.ReadDir/os.ReadFile for List/Read, and a time.Now Clock. Do NOT run the CLI against the real remote — tests only use fakes. A fix-id RED runs in parallel in internal/compat/latency/run_test.go — do not touch internal/.

### Scope
#### May
- tools/stage1latency/run.go and tools/stage1latency/deps.go only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T6, `plan-ready.md` § T6, `validate.md` § T6. Chain base `chain/s2-08` @ 664e9b7.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`tools/stage1latency/run.go tools/stage1latency/deps.go`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S2-08/fix-id · attempt 1 · green-worker · 2026-09-22T03:08:51Z

### Mandate
Implement S2-08 fix-id per `tasks.md` § fix-id with the least change that makes exactly TestRunRejectsShortSnapshotID (+ all latency tests) pass. Reject any snapshot id that is not exactly 64 lowercase hex chars with an error mentioning 'snapshot id'; existing deferred cleanup must still run. A T6 GREEN runs in parallel in tools/ — do not touch it.

### Scope
#### May
- internal/compat/latency/run.go only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § fix-id, `plan-ready.md` § fix-id, `validate.md` § fix-id. Chain base `chain/s2-08` @ 08c43c3.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`internal/compat/latency/run.go`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S2-08/fix-cli-id · attempt 1 · red-worker · 2026-09-22T03:16:04Z

### Mandate
Write every fix-cli-id test bullet in `plan.md` § fix-cli-id at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. Integration gap found at S2-08 merge gate: the CLI tests' fake 'restic snapshots' reply uses a 63-hex-char id, which fix-id now correctly rejects, so TestRunWritesJSONAndExits0 and TestRunRemoteNotDeletedExits1Loudly fail. Change the fake id to exactly 64 lowercase hex chars. This is a test-data correction: after it, ALL tools/stage1latency tests must PASS (report PASS-ON-RED; no production change needed). Run go test -race ./... to confirm the whole repo passes.

### Scope
#### May
- Create the test files named for fix-cli-id in `tasks.md`: tools/stage1latency/run_test.go only
- Test helpers that a fix-cli-id test itself exercises (e.g. `repoRoot`, `readRepoFile`, `yamlBlock`, `runInstaller`) go in a marked shim `zz_agentic_shim_test.go` in the same test package (first line `// agentic:shim`) with deliberately WRONG bodies so those tests fail by assertion; helpers NOT under test may be real and live in the named helpers file.
#### May Not
- Create or edit any non-test artifact (workflow YAML, JSON config, install.sh, .goreleaser.yaml, SPEC.md/README.md, go.mod) — those are GREEN's job; touch other stories' files; suppress/skip tests (except the tool-absent `t.Skip` cases plan.md explicitly allows).

### Inputs
- `plan.md` § fix-cli-id, `tasks.md` § fix-cli-id (contracts), `validate.md` § fix-cli-id. Chain base: the story chain branch named in your prompt (stage-1 @ 2158ade = master + sprint2 plan).

### Acceptance
`go test ./test/...` for this story's package compiles; every new test FAILS by assertion; `go vet ./...` clean; diff vs BASE_REF = only this task's test files (+ shim); output.md block appended; selfcheck PASS.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md in first); orchestrator relays output.md back.
- red-worker: make shim bodies differ so comparison tests cannot pass by accident.
- all: errcheck flags unchecked writes — use explicit `_, _ =` discard, never nolint.
