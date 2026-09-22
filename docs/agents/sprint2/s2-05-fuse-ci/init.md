---
type: init
story: S2-05
---

## S2-05/T1 · attempt 1 · red-worker · 2026-09-22T02:13:48Z

### Mandate
Write every T1 test bullet in `plan.md` § T1 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. fuse-linux job skeleton; the job does not exist yet. Follow /private/tmp/claude-501/-Users-adeelahmad-work-snapback/3a548722-26e5-4d22-8251-dbadece5e939/scratchpad/red-protocol.md.

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

## S2-05/T2 · attempt 1 · red-worker · 2026-09-22T02:16:36Z

### Mandate
Write every T2 test bullet in `plan.md` § T2 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. integration run step + evidence upload; base has T1 tests (fuse_job_test.go) — append T2 tests to the same file. Follow /private/tmp/claude-501/-Users-adeelahmad-work-snapback/3a548722-26e5-4d22-8251-dbadece5e939/scratchpad/red-protocol.md.

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

## S2-05/T1 · attempt 1 · green-worker · 2026-09-22T02:19:09Z

### Mandate
Implement S2-05 T1 per `tasks.md` § T1 with the least change that makes exactly the 8 T1 TestFuseJob* tests in test/ci/fuse_job_test.go; all 22 pre-existing test/ci tests must stay passing pass. SCAFFOLD vacuous (no Go symbols, no shims). T2 tests may still fail. Avoid "mount"/"browse"/"unmount" in any step name:; go-version-file: go.mod; pin actions; actionlint clean.

### Scope
#### May
- `.github/workflows/ci.yml` (edit: append the `fuse-linux` job after `cross-compile` — skeleton + pinned tool installs + restic 0.19.0 download with pinned SHA256 + version assertion) only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T1, `plan-ready.md` § T1, `validate.md` § T1. Chain base `chain/s2-05` @ 4c8bd48.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`.github/workflows/ci.yml`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S2-05/T2 · attempt 1 · green-worker · 2026-09-22T02:21:50Z

### Mandate
Implement S2-05 T2 per `tasks.md` § T2 with the least change that makes exactly the 8 T2 tests in test/ci/fuse_job_test.go (all 38 test/ci tests must pass) pass. Full matrix runs (all repo tests pass on this chain; integration tests are build-tagged). Step names must avoid mount/browse/unmount.

### Scope
#### May
- `.github/workflows/ci.yml` (edit the fuse-linux job: SNAPBACK_FUSE_TESTS="1" + SNAPBACK_EVIDENCE_DIR job env; one-line /dev/fuse check + mkdir evidence dir before tests; `go test -race -tags=integration ./...`; upload-artifact stage1-evidence-linux with if: always(), if-no-files-found: warn, path naming SNAPBACK_EVIDENCE_DIR literally) only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T2, `plan-ready.md` § T2, `validate.md` § T2. Chain base `chain/s2-05` @ 5e3fbea.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`.github/workflows/ci.yml`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.
