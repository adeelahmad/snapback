---
type: init
story: S2-03
---

## S2-03/T1 · attempt 1 · red-worker · 2026-09-22T02:13:47Z

### Mandate
Write every T1 test bullet in `plan.md` § T1 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. version check and repo guard. Follow /private/tmp/claude-501/-Users-adeelahmad-work-snapback/3a548722-26e5-4d22-8251-dbadece5e939/scratchpad/red-protocol.md.

### Scope
#### May
- Create the test files named for T1 in `tasks.md`: see tasks.md § T1
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

## S2-03/T2 · attempt 1 · red-worker · 2026-09-22T02:13:48Z

### Mandate
Write every T2 test bullet in `plan.md` § T2 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. argument builders. Follow /private/tmp/claude-501/-Users-adeelahmad-work-snapback/3a548722-26e5-4d22-8251-dbadece5e939/scratchpad/red-protocol.md.

### Scope
#### May
- Create the test files named for T2 in `tasks.md`: see tasks.md § T2
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

## S2-03/T3 · attempt 1 · red-worker · 2026-09-22T02:13:48Z

### Mandate
Write every T3 test bullet in `plan.md` § T3 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. snapshot JSON parsing. Follow /private/tmp/claude-501/-Users-adeelahmad-work-snapback/3a548722-26e5-4d22-8251-dbadece5e939/scratchpad/red-protocol.md.

### Scope
#### May
- Create the test files named for T3 in `tasks.md`: see tasks.md § T3
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

## S2-03/T4 · attempt 1 · red-worker · 2026-09-22T02:13:48Z

### Mandate
Write every T4 test bullet in `plan.md` § T4 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. fixture tree and password file. Follow /private/tmp/claude-501/-Users-adeelahmad-work-snapback/3a548722-26e5-4d22-8251-dbadece5e939/scratchpad/red-protocol.md.

### Scope
#### May
- Create the test files named for T4 in `tasks.md`: see tasks.md § T4
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

## S2-03/T5 · attempt 1 · red-worker · 2026-09-22T02:16:36Z

### Mandate
Write every T5 test bullet in `plan.md` § T5 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. Runner interface + repo lifecycle with fake runner; base has T1+T2+T3 shims — reuse their symbols, do not redeclare (NewPasswordFile is already in the T2 shim). Follow /private/tmp/claude-501/-Users-adeelahmad-work-snapback/3a548722-26e5-4d22-8251-dbadece5e939/scratchpad/red-protocol.md.

### Scope
#### May
- Create the test files named for T5 in `tasks.md`: see tasks.md § T5
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

## S2-03/T6 · attempt 1 · red-worker · 2026-09-22T02:16:36Z

### Mandate
Write every T6 test bullet in `plan.md` § T6 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. evidence JSON + MissingPrerequisite; base is 2158ade (no other shims) — T6 is independent; if you need NewPasswordFile note that T2/T4 also declare it: prefer NOT to declare it; if unavoidable report it. Follow /private/tmp/claude-501/-Users-adeelahmad-work-snapback/3a548722-26e5-4d22-8251-dbadece5e939/scratchpad/red-protocol.md.

### Scope
#### May
- Create the test files named for T6 in `tasks.md`: see tasks.md § T6
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

## S2-03/scaffold · attempt 1 · scaffolder · 2026-09-22T02:20:00Z

### Mandate
S2-03 RED T1–T6 are verified (T7 is RED'd after this scaffold because it needs every symbol). The combined branch `chain2/s2-03` @ 6cf7488 does NOT compile: six marked shims `internal/compat/resticfx/zz_agentic_shim_t{1..6}.go` overlap (`NewPasswordFile` declared in both t2 and t4; possibly duplicate package doc comments). Move EVERY shim symbol into its canonical production file per `tasks.md` (version.go, guard.go, args.go, snapshots.go, tree.go, password.go, runner.go, fixture.go, evidence.go, prereq.go — use the file names tasks.md assigns), each defined EXACTLY ONCE with the signature the tests call, body `panic("SUB-AGENT-TODO: <recipe>")` (consts/vars/types keep their declared shape; exported consts that tests compare may keep a placeholder value marked `// SUB-AGENT-TODO`). `NewPasswordFile` lives only in password.go. One package doc comment (put it in doc.go or version.go). Delete all six shims. Write the symbol list to `.agentic/scaffold-symbols`.

### Scope
#### May
- Create the production files above (stubs only), delete the shims.
#### May Not
- Implement bodies; edit tests; touch other packages.

### Acceptance
`go vet ./internal/compat/resticfx/` compiles; `go test ./internal/compat/resticfx/` has no newly passing test (panics/failures expected); no `agentic:shim` left; each symbol once (`ctx-symbols`); `### Scaffold` block in output.md; selfcheck PASS.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md in); orchestrator relays output.md back.

## S2-03/T1 · attempt 1 · green-worker · 2026-09-22T02:24:35Z

### Mandate
Implement S2-03 T1 per `tasks.md` § T1 with the least change that makes exactly the T1 tests (version_test.go, guard_test.go) pass. PinnedResticVersion = "0.19.0". Other S2-03 GREEN tasks run in parallel on other files — touch only yours. GATE_RUN_MATRIX=0 (intermediate).

### Scope
#### May
- ,  only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T1, `plan-ready.md` § T1, `validate.md` § T1. Chain base `chain/s2-03` @ abf9354.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=` `; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S2-03/T2 · attempt 1 · green-worker · 2026-09-22T02:24:35Z

### Mandate
Implement S2-03 T2 per `tasks.md` § T2 with the least change that makes exactly the T2 tests (args_test.go) pass. argv arrays only; --path-template and ids/%I separate; never the password. Other S2-03 GREEN tasks run in parallel on other files — touch only yours. GATE_RUN_MATRIX=0 (intermediate).

### Scope
#### May
-  only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T2, `plan-ready.md` § T2, `validate.md` § T2. Chain base `chain/s2-03` @ abf9354.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=``; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S2-03/T3 · attempt 1 · green-worker · 2026-09-22T02:24:35Z

### Mandate
Implement S2-03 T3 per `tasks.md` § T3 with the least change that makes exactly the T3 tests (snapshots_test.go) pass. 64-hex full IDs only; never read short_id. Other S2-03 GREEN tasks run in parallel on other files — touch only yours. GATE_RUN_MATRIX=0 (intermediate).

### Scope
#### May
-  only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T3, `plan-ready.md` § T3, `validate.md` § T3. Chain base `chain/s2-03` @ abf9354.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=``; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S2-03/T4 · attempt 1 · green-worker · 2026-09-22T02:24:35Z

### Mandate
Implement S2-03 T4 per `tasks.md` § T4 with the least change that makes exactly the T4 tests (tree_test.go, password_test.go) pass. reject escaping paths; password file 0600, random hex. Other S2-03 GREEN tasks run in parallel on other files — touch only yours. GATE_RUN_MATRIX=0 (intermediate).

### Scope
#### May
- ,  only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T4, `plan-ready.md` § T4, `validate.md` § T4. Chain base `chain/s2-03` @ abf9354.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=` `; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S2-03/T6 · attempt 1 · green-worker · 2026-09-22T02:24:35Z

### Mandate
Implement S2-03 T6 per `tasks.md` § T6 with the least change that makes exactly the T6 tests (evidence_test.go, prereq_test.go) pass. no password/password-file path in evidence; skip messages name the missing prerequisite. Other S2-03 GREEN tasks run in parallel on other files — touch only yours. GATE_RUN_MATRIX=0 (intermediate).

### Scope
#### May
- ,  only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T6, `plan-ready.md` § T6, `validate.md` § T6. Chain base `chain/s2-03` @ abf9354.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=` `; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S2-03/T5 · attempt 1 · green-worker · 2026-09-22T02:30:37Z

### Mandate
Implement S2-03 T5 per `tasks.md` § T5 with the least change that makes exactly the 7 T5 tests (runner_test.go, fixture_test.go) pass. Only exec.CommandContext with argv arrays. Also check golangci-lint on the package; report any remaining SA4006 in test files (do not edit tests).

### Scope
#### May
- `internal/compat/resticfx/runner.go`, `internal/compat/resticfx/fixture.go` only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T5, `plan-ready.md` § T5, `validate.md` § T5. Chain base `chain/s2-03` @ e06cf09.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`internal/compat/resticfx/runner.go internal/compat/resticfx/fixture.go`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S2-03/T7 · attempt 1 · red-worker · 2026-09-22T02:32:48Z

### Mandate
Write every T7 test bullet in `plan.md` § T7 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. mount supervisor (no timeout, idempotent Stop registered in t.Cleanup) + gated integration test (//go:build integration, SNAPBACK_FUSE_TESTS=1, skip naming prerequisite) that proves --path-template ids/%I against real restic 0.19.0 and writes pathtemplate-<goos>.json. Base 436df17 has ALL T1-T6 GREEN — unit tests for the supervisor must fail by assertion via a zz_agentic_shim_t7.go stub. Follow /private/tmp/claude-501/-Users-adeelahmad-work-snapback/3a548722-26e5-4d22-8251-dbadece5e939/scratchpad/red-protocol.md.

### Scope
#### May
- Create the test files named for T7 in `tasks.md`: see tasks.md § T7
- Test helpers that a T7 test itself exercises (e.g. `repoRoot`, `readRepoFile`, `yamlBlock`, `runInstaller`) go in a marked shim `zz_agentic_shim_test.go` in the same test package (first line `// agentic:shim`) with deliberately WRONG bodies so those tests fail by assertion; helpers NOT under test may be real and live in the named helpers file.
#### May Not
- Create or edit any non-test artifact (workflow YAML, JSON config, install.sh, .goreleaser.yaml, SPEC.md/README.md, go.mod) — those are GREEN's job; touch other stories' files; suppress/skip tests (except the tool-absent `t.Skip` cases plan.md explicitly allows).

### Inputs
- `plan.md` § T7, `tasks.md` § T7 (contracts), `validate.md` § T7. Chain base: the story chain branch named in your prompt (stage-1 @ 2158ade = master + sprint2 plan).

### Acceptance
`go test ./test/...` for this story's package compiles; every new test FAILS by assertion; `go vet ./...` clean; diff vs BASE_REF = only this task's test files (+ shim); output.md block appended; selfcheck PASS.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md in first); orchestrator relays output.md back.
- red-worker: make shim bodies differ so comparison tests cannot pass by accident.
- all: errcheck flags unchecked writes — use explicit `_, _ =` discard, never nolint.

## S2-03/scaffold-T7 · attempt 1 · scaffolder · 2026-09-22T02:39:43Z

### Mandate
Chain `chain2/s2-03` @ 218d3c9 (T1–T6 GREEN; T7 RED verified). Move the T7 shim `internal/compat/resticfx/zz_agentic_shim_t7.go` symbols (MountStarter, Mount, StartMount, WaitReady, Stop, Exited) into `internal/compat/resticfx/mount.go` exactly once with identical signatures; bodies `panic("SUB-AGENT-TODO: …")`. Delete the shim. Omit from `.agentic/scaffold-symbols` any bare name that collides elsewhere in the repo (e.g. `Mount`, `Stop`) and say so.

### Acceptance
`go vet ./internal/compat/resticfx/` and `go vet -tags integration ./internal/compat/resticfx/` compile; T1–T6 tests still pass; T7 unit tests still fail; no `agentic:shim`; `### Scaffold` block; selfcheck PASS; commit has NO AI attribution trailers.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree.
- all: commit only as the configured git user; no Co-Authored-By / Claude-Session trailers.

## S2-03/T7 · attempt 1 · green-worker · 2026-09-22T02:44:03Z

### Mandate
Implement S2-03 T7 per `tasks.md` § T7 with the least change that makes exactly TestMountWaitReadyAndStop and the other T7 unit tests, plus the integration-tagged path-template test pass. LAST task of S2-03: full matrix + coverage >=80%. After unit tests pass, run the integration test locally (restic 0.19 + macFUSE installed): SNAPBACK_EVIDENCE_DIR=$(pwd)/docs/reports/stage1 GOTOOLCHAIN=auto go test -tags integration -run PathTemplate -v ./internal/compat/resticfx/ and commit the generated pathtemplate-darwin.json (never hand-write it). Disposable repo in t.TempDir only, generated data only, no real backends, no passwords in the repo. If the integration test SKIPs or cannot mount within the time limit, report that honestly and commit only mount.go.

### Scope
#### May
- internal/compat/resticfx/mount.go and the generated docs/reports/stage1/pathtemplate-darwin.json only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T7, `plan-ready.md` § T7, `validate.md` § T7. Chain base `chain/s2-03` @ 965cde8.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`internal/compat/resticfx/mount.go docs/reports/stage1/pathtemplate-darwin.json`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S2-03/fix-hang · attempt 1 · red-worker · 2026-09-22T02:53:21Z

### Mandate
Write every fix-hang test bullet in `plan.md` § fix-hang at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. Follow-up to T7 (see output.md T7 GREEN block): the helper-process 'hang' mode uses `select {}` with no other goroutines, so Go aborts it with 'all goroutines are asleep - deadlock!' (exit 2, ~40ms) and the Stop-on-ignored-SIGINT path is never exercised. (1) Change hang mode to `signal.Ignore(os.Interrupt); time.Sleep(time.Hour)` so it genuinely hangs and ignores interrupt. (2) Add `TestMountWaitReadyReturnsWhenProcessExits`: start the 'exit' helper (never creates ids), call WaitReady with a generous ctx (e.g. 10s), assert it returns a non-nil error that is NOT context.DeadlineExceeded within ~2s. Existing T7 tests must still pass with the hang fix; the new test must FAIL BY ASSERTION on current mount.go. No production code.

### Scope
#### May
- Create the test files named for fix-hang in `tasks.md`: internal/compat/resticfx/mount_test.go only
- Test helpers that a fix-hang test itself exercises (e.g. `repoRoot`, `readRepoFile`, `yamlBlock`, `runInstaller`) go in a marked shim `zz_agentic_shim_test.go` in the same test package (first line `// agentic:shim`) with deliberately WRONG bodies so those tests fail by assertion; helpers NOT under test may be real and live in the named helpers file.
#### May Not
- Create or edit any non-test artifact (workflow YAML, JSON config, install.sh, .goreleaser.yaml, SPEC.md/README.md, go.mod) — those are GREEN's job; touch other stories' files; suppress/skip tests (except the tool-absent `t.Skip` cases plan.md explicitly allows).

### Inputs
- `plan.md` § fix-hang, `tasks.md` § fix-hang (contracts), `validate.md` § fix-hang. Chain base: the story chain branch named in your prompt (stage-1 @ 2158ade = master + sprint2 plan).

### Acceptance
`go test ./test/...` for this story's package compiles; every new test FAILS by assertion; `go vet ./...` clean; diff vs BASE_REF = only this task's test files (+ shim); output.md block appended; selfcheck PASS.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md in first); orchestrator relays output.md back.
- red-worker: make shim bodies differ so comparison tests cannot pass by accident.
- all: errcheck flags unchecked writes — use explicit `_, _ =` discard, never nolint.

## S2-03/fix-hang · attempt 1 · green-worker · 2026-09-22T02:55:49Z

### Mandate
Implement S2-03 fix-hang per `tasks.md` § fix-hang with the least change that makes exactly TestMountWaitReadyReturnsWhenProcessExits (plus all existing resticfx tests) pass. WaitReady must also watch the mount process exiting: check <mnt>/ids first on each poll (a mount that created ids then exited still counts ready); if the process has exited before ids appears, return a wrapped process-exit error, not the context error. Full matrix (fix to a merged story).

### Scope
#### May
- internal/compat/resticfx/mount.go only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § fix-hang, `plan-ready.md` § fix-hang, `validate.md` § fix-hang. Chain base `chain/s2-03` @ 642f50d.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`internal/compat/resticfx/mount.go`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.
