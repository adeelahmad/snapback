---
type: init
story: S2-02
---

## S2-02/T1 · attempt 1 · red-worker · 2026-09-22T02:13:47Z

### Mandate
Write every T1 test bullet in `plan.md` § T1 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. package skeleton and dependency guard. Follow /private/tmp/claude-501/-Users-adeelahmad-work-snapback/3a548722-26e5-4d22-8251-dbadece5e939/scratchpad/red-protocol.md.

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

## S2-02/T2 · attempt 1 · red-worker · 2026-09-22T02:13:47Z

### Mandate
Write every T2 test bullet in `plan.md` § T2 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. name validation. Follow /private/tmp/claude-501/-Users-adeelahmad-work-snapback/3a548722-26e5-4d22-8251-dbadece5e939/scratchpad/red-protocol.md.

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

## S2-02/T3 · attempt 1 · red-worker · 2026-09-22T02:16:36Z

### Mandate
Write every T3 test bullet in `plan.md` § T3 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. spec types, Build, ErrDuplicateName, deep copy, deterministic inodes; base has T1+T2 (T2 shim declares ErrInvalidName/validateName — reuse, do not redeclare). Follow /private/tmp/claude-501/-Users-adeelahmad-work-snapback/3a548722-26e5-4d22-8251-dbadece5e939/scratchpad/red-protocol.md.

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

## S2-02/T4 · attempt 1 · red-worker · 2026-09-22T02:18:37Z

### Mandate
Write every T4 test bullet in `plan.md` § T4 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. Lookup + fixture + generation isolation + concurrent reads; base has T1-T3 shims — reuse, do not redeclare; new stubs in zz_agentic_shim_t4.go. Follow /private/tmp/claude-501/-Users-adeelahmad-work-snapback/3a548722-26e5-4d22-8251-dbadece5e939/scratchpad/red-protocol.md.

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

## S2-02/T5 · attempt 1 · red-worker · 2026-09-22T02:20:57Z

### Mandate
Write every T5 test bullet in `plan.md` § T5 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. ReadDir byte-order + fresh copy + empty non-nil; base has T1-T4 incl. fixture_test.go helpers — reuse them; stub (*Generation).ReadDir(dir uint64) (names []string, found bool) in zz_agentic_shim_t5.go. Follow /private/tmp/claude-501/-Users-adeelahmad-work-snapback/3a548722-26e5-4d22-8251-dbadece5e939/scratchpad/red-protocol.md.

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

## S2-02/T6 · attempt 1 · red-worker · 2026-09-22T02:22:45Z

### Mandate
Write every T6 test bullet in `plan.md` § T6 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. Readlink byte-for-byte + catalog contract test (method-set check + fixture walk in ONE test, interface defined locally, NOT importing mount); stub (*Generation).Readlink(ino uint64) (target string, found bool) in zz_agentic_shim_t6.go; base has T1-T5. Follow /private/tmp/claude-501/-Users-adeelahmad-work-snapback/3a548722-26e5-4d22-8251-dbadece5e939/scratchpad/red-protocol.md.

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

## S2-02/scaffold · attempt 1 · scaffolder · 2026-09-22T02:25:26Z

### Mandate
S2-02 RED T1–T6 verified (T1 = PASS-ON-RED static guard). Chain `chain2/s2-02` @ e51df7e holds shims `internal/projection/zz_agentic_shim_t{2..6}.go`. Move every shim symbol into its canonical file per tasks.md (`doc.go` package doc, `name.go` ErrInvalidName/validateName, `spec.go` Spec/Dir/Link/RootIno, `build.go` Build/ErrDuplicateName, `catalog.go` Generation + Lookup/ReadDir/Readlink) EXACTLY once with the exact signatures (binding: Lookup(parent uint64, name string) (ino uint64, isDir bool, found bool); ReadDir(dir uint64) (names []string, found bool); Readlink(ino uint64) (target string, found bool); RootIno = 1; Build(Spec) (*Generation, error)). Bodies = `panic("SUB-AGENT-TODO: …")`; types/consts/vars keep their shape (sentinel errors keep a placeholder message marked TODO). Delete all shims; write `.agentic/scaffold-symbols`.

### Acceptance
`go vet ./internal/projection/` compiles; no newly passing test besides the known PASS-ON-RED guards (TestDepsStdlibOnly, method-set part of the contract test); no `agentic:shim`; each symbol once; `### Scaffold` block; selfcheck PASS.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree.

## S2-02/T2 · attempt 1 · green-worker · 2026-09-22T02:30:37Z

### Mandate
Implement S2-02 T2 per `tasks.md` § T2 with the least change that makes exactly TestValidateName pass. Reject empty, ".", "..", names with "/" or NUL; error wraps ErrInvalidName and quotes the name.

### Scope
#### May
- `internal/projection/name.go` only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T2, `plan-ready.md` § T2, `validate.md` § T2. Chain base `chain/s2-02` @ b50e403.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`internal/projection/name.go`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S2-02/T3 · attempt 1 · green-worker · 2026-09-22T02:32:05Z

### Mandate
Implement S2-02 T3 per `tasks.md` § T3 with the least change that makes exactly the 3 T3 tests (build_test.go) pass. Build validates every name (wrapping errors with the offending quoted path), rejects duplicate siblings (ErrDuplicateName), deep-copies the spec, numbers inodes deterministically (root=1, others from 2 depth-first by name). Lookup/ReadDir/Readlink stay stubs (T4-T6).

### Scope
#### May
- `internal/projection/spec.go`, `internal/projection/build.go` only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T3, `plan-ready.md` § T3, `validate.md` § T3. Chain base `chain/s2-02` @ 333187e.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`internal/projection/spec.go internal/projection/build.go`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S2-02/T4 · attempt 1 · green-worker · 2026-09-22T02:34:12Z

### Mandate
Implement S2-02 T4 per `tasks.md` § T4 with the least change that makes exactly the 5 T4 tests in lookup_test.go (incl. concurrent reads under -race) pass. Lookup(parent, name) via Generation.nodes; ReadDir/Readlink stay stubs (T5/T6). Anchor -run regexes with ^...$.

### Scope
#### May
- `internal/projection/catalog.go` (Lookup only) only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T4, `plan-ready.md` § T4, `validate.md` § T4. Chain base `chain/s2-02` @ 891cc7d.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`internal/projection/catalog.go`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S2-02/T5 · attempt 1 · green-worker · 2026-09-22T02:39:29Z

### Mandate
Implement S2-02 T5 per `tasks.md` § T5 with the least change that makes exactly the 4 T5 tests in readdir_test.go pass. Names in ascending byte order, fresh copy each call, empty dir → empty non-nil. Readlink stays a stub (T6). Anchor -run regexes. Report any remaining lint findings in test files.

### Scope
#### May
- `internal/projection/catalog.go` (ReadDir only) only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T5, `plan-ready.md` § T5, `validate.md` § T5. Chain base `chain/s2-02` @ b98f8d1.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`internal/projection/catalog.go`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S2-02/T6 · attempt 1 · green-worker · 2026-09-22T02:41:30Z

### Mandate
Implement S2-02 T6 per `tasks.md` § T6 with the least change that makes exactly the T6 tests (readlink_test.go, contract_test.go); ALL S2-02 tests must pass pass. This is S2-02's LAST task: FULL standards matrix runs, plus package coverage >= 80% (go test -cover ./internal/projection/). Readlink returns the stored target byte-for-byte; not found for dirs/unknown inodes.

### Scope
#### May
- `internal/projection/catalog.go` (Readlink only) only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T6, `plan-ready.md` § T6, `validate.md` § T6. Chain base `chain/s2-02` @ ce4c091.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`internal/projection/catalog.go`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.
