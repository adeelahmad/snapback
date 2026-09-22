---
type: init
story: S2-01
---

## S2-01/T1 · attempt 1 · red-worker · 2026-09-22T02:13:47Z

### Mandate
Write every T1 test bullet in `plan.md` § T1 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. go-fuse absent from go.mod; tests fail on the missing requirement/package. Follow /private/tmp/claude-501/-Users-adeelahmad-work-snapback/3a548722-26e5-4d22-8251-dbadece5e939/scratchpad/red-protocol.md.

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

## S2-01/T2 · attempt 1 · red-worker · 2026-09-22T02:15:30Z

### Mandate
Write every T2 test bullet in `plan.md` § T2 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. mount.go types + Catalog/Adapter/Observer interfaces with the binding builtin-type signatures; stubs in zz_agentic_shim_t2.go. Follow /private/tmp/claude-501/-Users-adeelahmad-work-snapback/3a548722-26e5-4d22-8251-dbadece5e939/scratchpad/red-protocol.md.

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

## S2-01/T3 · attempt 1 · red-worker · 2026-09-22T02:20:00Z

### Mandate
Write every T3 test bullet in `plan.md` § T3 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. gofuse attr translation (StableAttr/Attr/EntryOut/AttrOut/DaemonOwner/ReadOnlyErrno=EROFS, named timeout/mode constants); base has T1+T2. go-fuse is NOT in go.mod yet (GREEN T1) — if the T3 tests must reference go-fuse types they will not compile: in that case write them against the planned API and report it; do not add go.mod. Follow /private/tmp/claude-501/-Users-adeelahmad-work-snapback/3a548722-26e5-4d22-8251-dbadece5e939/scratchpad/red-protocol.md.

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

## S2-01/scaffold · attempt 1 · scaffolder · 2026-09-22T02:23:16Z

### Mandate
Scaffold S2-01 for T1+T2 only (T3's RED 4746ea7 is held until go-fuse is in go.mod after GREEN T1; it needs real go-fuse types). Chain `chain2/s2-01` @ 3cf9adb has T1 tests (no shim) and T2 tests + `internal/mount/zz_agentic_shim_t2.go`. Move every T2 shim symbol into `internal/mount/mount.go` exactly once: types/interfaces keep their exact declared shape (the binding Catalog signatures, `RootIno uint64 = 1`); `Kind`/`Op` constants stay declared (values may be placeholders marked `// SUB-AGENT-TODO`); func/method bodies (e.g. `Op.String`) are `panic("SUB-AGENT-TODO: …")`. Delete the shim. Do NOT create gofuse/attr.go or touch go.mod (GREEN T1). Write names to `.agentic/scaffold-symbols`.

### Scope
#### May
- Create `internal/mount/mount.go` (stubs), delete `internal/mount/zz_agentic_shim_t2.go`.
#### May Not
- Implement bodies; edit tests; touch go.mod/go.sum or internal/mount/gofuse.

### Acceptance
`go vet ./internal/mount/` compiles; no newly passing test; no `agentic:shim` in internal/mount; each symbol once; `### Scaffold` block; selfcheck PASS.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree.

## S2-01/T1 · attempt 1 · green-worker · 2026-09-22T02:26:01Z

### Mandate
Implement S2-01 T1 per `tasks.md` § T1 with the least change that makes exactly the 3 T1 tests pass. go get github.com/hanwen/go-fuse/v2@v2.11.0; go mod tidy; CGO_ENABLED=0 go build ./... must pass. T2 runs in parallel on mount.go — do not touch it.

### Scope
#### May
- `go.mod`, `go.sum`, `internal/mount/gofuse/attr.go` (placeholder with blank imports of go-fuse fs+fuse) only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T1, `plan-ready.md` § T1, `validate.md` § T1. Chain base `chain/s2-01` @ d41aa8b.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`go.mod go.sum internal/mount/gofuse/attr.go`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S2-01/T2 · attempt 1 · green-worker · 2026-09-22T02:26:01Z

### Mandate
Implement S2-01 T2 per `tasks.md` § T2 with the least change that makes exactly TestOpString and TestKindAndRootConstants (guards must stay passing) pass. Kind/Op constants and Op.String per tasks.md. T1 runs in parallel on go.mod/attr.go — do not touch those.

### Scope
#### May
- `internal/mount/mount.go` only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T2, `plan-ready.md` § T2, `validate.md` § T2. Chain base `chain/s2-01` @ d41aa8b.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`internal/mount/mount.go`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S2-01/scaffold-T3 · attempt 1 · scaffolder · 2026-09-22T02:30:37Z

### Mandate
Chain `chain2/s2-01` @ 2606f2c: T1+T2 GREEN merged; T3 RED (8 tests, verified FAIL by assertion) with shim `internal/mount/gofuse/zz_agentic_shim_t3.go`. Move every T3 shim symbol into `internal/mount/gofuse/attr.go` exactly once (replacing T1's placeholder blank imports with real go-fuse imports as the stubs need), bodies `panic("SUB-AGENT-TODO: …")`, named consts keep shape (placeholder values marked `// SUB-AGENT-TODO`). Delete the shim; write `.agentic/scaffold-symbols` (omit names that collide by bare name elsewhere, e.g. `Attr`, and say so).

### Acceptance
`go vet ./internal/mount/...` + `CGO_ENABLED=0 go build ./...` compile; `go mod tidy -diff` empty; no newly passing T3 test; no `agentic:shim`; `### Scaffold` block; selfcheck PASS.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree.

## S2-01/T3 · attempt 1 · green-worker · 2026-09-22T02:32:32Z

### Mandate
Implement S2-01 T3 per `tasks.md` § T3 with the least change that makes exactly the 8 T3 tests in internal/mount/gofuse/attr_test.go pass. This is S2-01's LAST task: FULL standards matrix runs; all S2-01 tests must pass. Timeouts/perm modes as named constants per tasks.md; ReadOnlyErrno = syscall.EROFS; DaemonOwner from os.Getuid/Getgid.

### Scope
#### May
- `internal/mount/gofuse/attr.go` only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T3, `plan-ready.md` § T3, `validate.md` § T3. Chain base `chain/s2-01` @ 71cdb8b.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`internal/mount/gofuse/attr.go`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S2-01/fix-notice · attempt 1 · green-worker · 2026-09-22T02:39:29Z

### Mandate
Implement S2-01 fix-notice per `tasks.md` § fix-notice with the least change that makes exactly test/projectdocs TestNoticeCoversGoModRequires plus all S2-01 tests pass. S2-01's final gate is red only because NOTICE (S1-08 guard) lacks the new go.mod modules github.com/hanwen/go-fuse/v2 (BSD-3-Clause) and golang.org/x/sys (BSD-3-Clause) — verify licences from the module cache LICENSE files. Full matrix runs.

### Scope
#### May
- `NOTICE` (add go-fuse v2 and golang.org/x/sys with their licences) only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § fix-notice, `plan-ready.md` § fix-notice, `validate.md` § fix-notice. Chain base `chain/s2-01` @ 36dd8a7.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`NOTICE`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.
