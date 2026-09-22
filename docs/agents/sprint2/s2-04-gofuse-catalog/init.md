---
type: init
story: S2-04
---

## S2-04/T1 · attempt 1 · red-worker · 2026-09-22T02:46:45Z

### Mandate
Write every T1 test bullet in `plan.md` § T1 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. Build catalogs with the merged internal/projection (projection.Build) and the merged internal/mount contract; use the exact merged identifiers for mount.Observer/Entry and attr.go helpers (read them in internal/mount/mount.go and internal/mount/gofuse/attr.go). Attach nodes in-process with fs.NewNodeFS — no kernel mount. The shared fixture helper lives in fs_test.go.

### Scope
#### May
- Create the test files named for T1 in `tasks.md`: internal/mount/gofuse/fs_test.go (+ internal/mount/gofuse/zz_agentic_shim_t1.go)
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

## S2-04/T1 · attempt 1 · scaffolder · 2026-09-22T02:50:22Z

### Mandate
Replace `internal/mount/gofuse/zz_agentic_shim_t1.go` with production stubs in `internal/mount/gofuse/fs.go`: `dirNode`, `symlinkNode` (fields as the shim), `newRoot`, and the dirNode methods Lookup/Readdir/Getattr/Statfs and symlinkNode Readlink/Getattr with the go-fuse v2 signatures. Every body is `panic("SUB-AGENT-TODO: <recipe>")`. Delete the shim. Each symbol exactly once.

### Acceptance
`go vet ./...` compiles; existing attr tests PASS; T1 tests fail (SUB-AGENT-TODO panic acceptable); no `agentic:shim` left; gofmt clean; `.agentic/scaffold-symbols` lists distinctive symbols only; commit `chore: scaffold S2-04 read-path node stubs`, no AI trailers.

## S2-04/T1 · attempt 1 · green-worker · 2026-09-22T02:52:26Z

### Mandate
Implement S2-04 T1 per `tasks.md` § T1 with the least change that makes exactly the 9 T1 tests in fs_test.go pass. Use attr.go helpers (AttrOut, DaemonOwner, timeout constants) and the mount.Catalog/Observer contract; Observer root path is "" and child paths join with '/'. Run the attr tests by exact name if filtering (the regex 'Stable' also matches a T1 test).

### Scope
#### May
- internal/mount/gofuse/fs.go only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T1, `plan-ready.md` § T1, `validate.md` § T1. Chain base `chain/s2-04` @ e8bde6b.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`internal/mount/gofuse/fs.go`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S2-04/T2 · attempt 1 · red-worker · 2026-09-22T02:54:22Z

### Mandate
Write every T2 test bullet in `plan.md` § T2 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. Runs in parallel with T3 RED (adapter_test.go, shim_t3) — do not touch those. Add mutation methods to dirNode/symlinkNode in your shim returning a wrong errno (e.g. syscall.EPERM) so tests fail by assertion. Reuse the fixture helpers already in fs_test.go.

### Scope
#### May
- Create the test files named for T2 in `tasks.md`: internal/mount/gofuse/fs_test.go (+ internal/mount/gofuse/zz_agentic_shim_t2.go)
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

## S2-04/T3 · attempt 1 · red-worker · 2026-09-22T02:54:22Z

### Mandate
Write every T3 test bullet in `plan.md` § T3 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. Runs in parallel with T2 RED (fs_test.go, shim_t2) — do not touch those. Reuse fixtureSpec/newFixture from fs_test.go (read-only). Shim: Adapter type, its constructor, mountOptions func, Mount/Unmount matching mount.Adapter's merged signature — wrong behaviour (e.g. options without ro, Mount returns nil on bad dirs WITHOUT calling fs.Mount) so tests fail by assertion; never perform a real FUSE mount in default-build tests.

### Scope
#### May
- Create the test files named for T3 in `tasks.md`: internal/mount/gofuse/adapter_test.go (+ internal/mount/gofuse/zz_agentic_shim_t3.go)
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

## S2-04/T2 · attempt 1 · green-worker · 2026-09-22T02:59:46Z

### Mandate
Implement S2-04 T2 per `tasks.md` § T2 with the least change that makes exactly the 4 T2 tests pass. Parallel GREENs: T2 owns fs.go, T3 owns adapter.go — touch only yours. Run only your tests by exact name (the sibling's stubs panic). Every mutation returns attr.go's EROFS errno, creates no inode, fires no Observer event; Write returns count 0.

### Scope
#### May
- internal/mount/gofuse/fs.go only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T2, `plan-ready.md` § T2, `validate.md` § T2. Chain base `chain/s2-04` @ 986195a.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`internal/mount/gofuse/fs.go`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S2-04/T3 · attempt 1 · green-worker · 2026-09-22T02:59:46Z

### Mandate
Implement S2-04 T3 per `tasks.md` § T3 with the least change that makes exactly the 7 T3 tests pass. Parallel GREENs: T2 owns fs.go, T3 owns adapter.go — touch only yours. Run only your tests by exact name (the sibling's stubs panic). Preflight (exists, is dir, empty) before any FUSE call; Mount builds newRoot and calls fs.Mount with mountOptions() (ro, AllowOther false, fixed FsName/Name); server lifetime not bound to a timeout context; Unmount calls server Unmount then Wait, wrapped error when nothing mounted. Default-build tests never really mount.

### Scope
#### May
- internal/mount/gofuse/adapter.go only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T3, `plan-ready.md` § T3, `validate.md` § T3. Chain base `chain/s2-04` @ 986195a.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`internal/mount/gofuse/adapter.go`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S2-04/T4 · attempt 1 · red-worker · 2026-09-22T03:02:24Z

### Mandate
Write every T4 test bullet in `plan.md` § T4 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. File tagged //go:build integration with the exact skip gating in tasks.md § T4. T1-T3 are implemented, so after writing it run it for real on this Mac: SNAPBACK_FUSE_TESTS=1 SNAPBACK_EVIDENCE_DIR=$(mktemp -d) GOTOOLCHAIN=auto go test -race -tags=integration -run TestCatalogMountLifecycle -v ./internal/mount/gofuse/ (macFUSE installed). If it PASSES, report PASS-ON-RED (implementation already satisfies it) — do not weaken it; if it FAILS, report the exact assertion (that becomes GREEN work). Always ensure the mount is gone afterwards (umount the temp mountpoint if a failure left it). Do not commit the evidence JSON (orchestrator/GREEN does). Also confirm the default build (no tag) still passes.

### Scope
#### May
- Create the test files named for T4 in `tasks.md`: internal/mount/gofuse/mount_integration_test.go only
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

## S2-04/T4 · attempt 1 · green-worker · 2026-09-22T03:05:45Z

### Mandate
Implement S2-04 T4 per `tasks.md` § T4 with the least change that makes exactly TestCatalogMountLifecycle (integration) + full default-build matrix pass. LAST task of S2-04. RED was PASS-ON-RED (accepted by orchestrator): no production change expected. Run SNAPBACK_FUSE_TESTS=1 SNAPBACK_EVIDENCE_DIR=$(pwd)/docs/reports/stage1 GOTOOLCHAIN=auto go test -race -count=1 -tags=integration -run TestCatalogMountLifecycle -v ./internal/mount/gofuse/, confirm result pass and no leftover mount, commit ONLY the generated catalog-darwin.json (never hand-edit). Then run the full standards matrix incl. coverage >=80%.

### Scope
#### May
- docs/reports/stage1/catalog-darwin.json (generated) only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T4, `plan-ready.md` § T4, `validate.md` § T4. Chain base `chain/s2-04` @ 864025d.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`docs/reports/stage1/catalog-darwin.json`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.
- NOTE: `docs/reports/stage1/` does not exist on this chain (S2-03 evidence landed on stage-1 later) and the test does not create the evidence dir — run `mkdir -p docs/reports/stage1` first.
