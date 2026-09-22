---
type: output
story: S2-02
---

## S2-02/T1 · attempt 1 · red-worker · 2026-09-22T02:20:00Z
status: escalate
### Summary
Wrote the single T1 bullet, `internal/projection/deps_test.go::TestDepsStdlibOnly`, exactly as plan.md specifies. No shim is needed or created (tasks.md T1 says "RED shims: none"). The test compiles, and `go vet ./...` and `golangci-lint run ./internal/projection/...` are both clean. **The test PASSES on RED. It does not fail.** The tasks.md premise is wrong for go1.27.1: with only `deps_test.go` in the directory, `go list -deps ./internal/projection` exits 0 and prints just `github.com/adeelahmad/snapback/internal/projection`. That output meets every assertion in the bullet: it is non-empty, contains the package itself, has no go-fuse or mount entries, and is stdlib-only. The test is a static negative guard, so per M-003 I kept it as written, did not weaken it, and flag it for orchestrator adjudication. Adding an assertion such as "doc.go exists" would go beyond the plan bullet, so I did not add one. Commit 5ed3852 on branch worktree-agent-a0b6257df27d6c343. The diff against 2158ade is only `internal/projection/deps_test.go`.
### Result
| Check | Status | Detail |
|---|---|---|
| `internal/projection/deps_test.go::TestDepsStdlibOnly` | PASS-ON-RED | static guard; `go list -deps` succeeds on a test-only package in go1.27.1 and lists only the projection package, so there is no RED failure mode (M-003) |
| `go vet ./...` | PASS | clean |
| `golangci-lint run ./internal/projection/...` | PASS | 0 issues |
| scope (`git diff --stat 2158ade HEAD`) | PASS | 1 file: internal/projection/deps_test.go |
### Next
Orchestrator: accept TestDepsStdlibOnly as PASS-ON-RED. Then scaffold/GREEN creates `internal/projection/doc.go` (package doc comment plus `package projection` only) and re-runs the T1 validate rows. Planner: correct the tasks.md T1 claim that `go list -deps` fails with no non-test Go file.

## S2-02/T2 · attempt 1 · red-worker · 2026-09-22T02:20:00Z
status: ok
### Summary
Wrote the single T2 test bullet, `TestValidateName`, as a table test: 9 names that must be accepted and 6 that must be rejected. Each rejected row checks `errors.Is(err, ErrInvalidName)` and that the error message contains `%q` of the name. The compile shim `zz_agentic_shim_t2.go` (`package projection`, first line `// agentic:shim`) declares `ErrInvalidName` and a `validateName` that returns nil for every name. With that shim all 6 rejection rows fail by assertion and the package compiles without T1's doc.go. The accepted rows pass under the shim, which is expected because they check that valid names return nil. The 64-hex local is named `snapshotID` so it cannot collide with T4's package-level `const fullID`. `go vet ./...` is clean, golangci-lint reports 0 issues and gofmt is clean. Commit 9feb8f7 changes only 2 files.
### Result
| Check | Status | Detail |
|---|---|---|
| `internal/projection/name_test.go::TestValidateName` | FAIL | 6 reject rows fail by assertion: `validateName("") = nil, want error wrapping ErrInvalidName` (the same for `.`, `..`, `a/b`, `/`, `a\x00b`); 9 accept subtests pass under the shim |
| `go vet ./...` | PASS | clean |
| `golangci-lint run ./internal/projection/...` | PASS | 0 issues |
### Next
Scaffold `internal/projection/name.go` with `ErrInvalidName` and `validateName`, then delete `zz_agentic_shim_t2.go`.

## S2-02/T3 · attempt 1 · red-worker · 2026-09-22T02:18:06Z
status: ok
### Summary
Wrote the three T3 tests from plan.md in `internal/projection/build_test.go`. They are pinned to the tasks.md contract: `Build(Spec) (*Generation, error)`, `fmt.Errorf("projection: %q: %w", path, err)` wrapping, a nil generation on error, and duplicates checked across the union of Dirs and Links. The compile shim `zz_agentic_shim_t3.go` declares `RootIno`, `ErrDuplicateName` (with a wrong message), `Spec`, `Dir`, `Link`, an empty `Generation`, and a `Build` that always returns `&Generation{}, nil`. It reuses the T2 shim's `ErrInvalidName` and `validateName` and does not redeclare them. Each test compiles and fails by assertion. Committed as a6e9554 on top of a6e3740.
### Result
| Check | Status | Detail |
|---|---|---|
| `internal/projection/build_test.go::TestBuildRejectsInvalidNames` | FAIL | all 20 rows (5 names x root dir, root link, nested dir, nested link) fail with `Build error = <nil>, want errors.Is ErrInvalidName` |
| `internal/projection/build_test.go::TestBuildRejectsDuplicateSiblings` | FAIL | the 4 duplicate rows fail with `Build error = <nil>, want errors.Is ErrDuplicateName`; the control row (x under two parents) passes as designed |
| `internal/projection/build_test.go::TestBuildErrorWrapsWithPath` | FAIL | `Build error = nil, want wrapped ErrInvalidName` |
| `go vet ./...` | PASS | clean |
| `golangci-lint run ./internal/projection/...` | PASS | 0 issues |
### Next
Scaffold `spec.go` (Spec/Dir/Link/RootIno/ErrDuplicateName) and `build.go` (Generation, Build), then delete `zz_agentic_shim_t3.go`. The T2 shim stays until T2 scaffold.

## S2-02/T4 · attempt 1 · red-worker · 2026-09-22T02:20:27Z
status: ok
### Summary
Wrote the five T4 tests in `internal/projection/lookup_test.go` and the shared fixture in `internal/projection/fixture_test.go`. The fixture holds `fixtureSpec()`, `fullID`, `fixturePaths()`, `buildFixture`, `lookupPath` and `mustLookupPath`, and root entries are declared out of byte order. The only new shim is `zz_agentic_shim_t4.go`: `(*Generation).Lookup(parent uint64, name string) (ino uint64, isDir bool, found bool)`, which returns `0, false, false`. The T2/T3 shims are reused and nothing is redeclared. All five tests fail by assertion under `go test -race`. `go vet ./...` and `golangci-lint run ./internal/projection/...` report 0 issues. The diff against a6e9554 touches only these 3 files, in commit a640c37.
### Result
| Check | Status | Detail |
|---|---|---|
| `internal/projection/lookup_test.go::TestLookup` | FAIL | lookupPath("docs") not found |
| `internal/projection/lookup_test.go::TestInodesStableWithinGeneration` | FAIL | lookupPath("snapshots") not found |
| `internal/projection/lookup_test.go::TestBuildIsDeterministic` | FAIL | lookupPath("snapshots") not found |
| `internal/projection/lookup_test.go::TestGenerationsAreIndependent` | FAIL | lookupPath("snapshots") not found |
| `internal/projection/lookup_test.go::TestConcurrentReadsDuringRebuild` | FAIL | g1.Lookup(RootIno, docs) = (0, false, false), want a found directory |
| `go vet ./...` / golangci-lint | PASS | 0 issues |
### Next
Scaffold: replace the shim with a `(*Generation).Lookup` stub in `internal/projection/catalog.go` and delete `zz_agentic_shim_t4.go`. GREEN: Generation needs a node table so that Lookup can resolve (parent, name) to (ino, isDir), with inodes assigned in depth-first preorder in byte order. The table has to be deep-copied at Build time, and `catalog.go` must not use `sync.`. Later tasks reuse `fixture_test.go` from T5/T6 without editing it.

## S2-02/T5 · attempt 1 · red-worker · 2026-09-22T02:22:17Z
status: ok
### Summary
Wrote the 4 T5 tests in `internal/projection/readdir_test.go`, reusing the T4 helpers (`buildFixture`, `lookupPath`) and the T3/T4 shims. Nothing is redeclared. Added `zz_agentic_shim_t5.go` with `(*Generation).ReadDir(dir uint64) (names []string, found bool)` stubbed to `nil, false`, the RED shim tasks.md specifies. All 4 tests compile and fail by assertion or t.Fatal. `go vet ./...` is clean, and golangci-lint reports 0 issues. Commit e98208a changes only the test file and the shim.
### Result
| Check | Status | Detail |
|---|---|---|
| `internal/projection/readdir_test.go::TestReadDirSortedByteOrder` | FAIL | call 0: ReadDir(RootIno) found = false, want true |
| `internal/projection/readdir_test.go::TestReadDirNotFound` | FAIL | lookupPath("rel") not found (M-002 precondition; the Lookup shim finds nothing) |
| `internal/projection/readdir_test.go::TestReadDirReturnsCopy` | FAIL | ReadDir(RootIno) = ([], false), want a found non-empty listing |
| `internal/projection/readdir_test.go::TestReadDirEmptyDirectory` | FAIL | Lookup(RootIno, "empty") not found (M-002 precondition) |
| `go vet ./...` / golangci-lint | PASS | 0 issues |
### Next
Scaffold: replace the shim with a `(*Generation).ReadDir` stub in `internal/projection/catalog.go` and delete `zz_agentic_shim_t5.go`. GREEN: store children sorted by byte order at Build time and return a fresh copy on each call. Return an empty non-nil slice for an empty dir, and `nil, false` for a symlink, 0 or an unknown inode.

## S2-02/T6 · attempt 1 · red-worker · 2026-09-22T02:24:29Z
status: ok
### Summary
Wrote the 4 T6 tests from plan.md, reusing the T4 fixture helpers (fixtureSpec, fullID, buildFixture, lookupPath) and redeclaring nothing. The shim `zz_agentic_shim_t6.go` stubs `(*Generation).Readlink(ino uint64) (target string, found bool)` with a wrong body that returns ("shim: not the stored target", true). TestGenerationSatisfiesCatalogContract keeps the method-set check and the fixture walk in one test. It defines the `catalog` interface inside the function, asserts `var _ catalog = (*Generation)(nil)`, and does not import mount. All 4 tests compile and fail by assertion. On this base the first assertion to trip is the Lookup/ReadDir shims from T4/T5; once those are green, the Readlink shim is what fails. go vet ./... is clean, and golangci-lint on ./internal/projection/... reports 0 issues. Commit e51df7e changes 3 files only.
### Result
| Check | Status | Detail |
|---|---|---|
| `internal/projection/readlink_test.go::TestReadlinkExactBytes` | FAIL | Lookup(RootIno, "l0") found=false, want a found symlink (9 subtests) |
| `internal/projection/readlink_test.go::TestReadlinkNotFound` | FAIL | lookupPath("rel") found=false, want a found symlink |
| `internal/projection/readlink_test.go::TestGenerationsIndependentTargets` | FAIL | g1: Lookup(RootIno, "rel") found=false |
| `internal/projection/contract_test.go::TestGenerationSatisfiesCatalogContract` | FAIL | ReadDir(RootIno) = ([], false), want a non-empty listing |
| `go vet ./...` / golangci-lint | PASS | 0 issues |
### Next
Scaffold: move Readlink into internal/projection/catalog.go as a SUB-AGENT-TODO stub and delete zz_agentic_shim_t6.go. GREEN: return the stored target verbatim; then run the full standards matrix plus the >=80% coverage check (validate.md Final sign-off).

## S2-02/scaffold · attempt 1 · scaffolder · 2026-09-22T02:28:14Z
status: ok
### Summary
Moved every RED shim symbol from `internal/projection/zz_agentic_shim_t{2..6}.go` into its canonical file, each defined exactly once with the binding signatures, and deleted all five shims. Function and method bodies are `panic("SUB-AGENT-TODO: ...")` recipes. Types and constants keep their shape. The sentinel errors hold placeholder messages marked SUB-AGENT-TODO, and spec.go carries a TODO for the unexported node representation. Added `doc.go` (package doc only). No production logic was implemented. `.agentic/scaffold-symbols` lists all 12 symbols. No method name collides with another definition in the repo, so none was left off.
### Scaffold
+ create package doc @ internal/projection/doc.go
+ create ErrInvalidName, validateName @ internal/projection/name.go
+ create RootIno, Spec, Dir, Link @ internal/projection/spec.go
+ create ErrDuplicateName, Generation, Build @ internal/projection/build.go
+ create (*Generation).Lookup, (*Generation).ReadDir, (*Generation).Readlink @ internal/projection/catalog.go
- delete internal/projection/zz_agentic_shim_t2.go
- delete internal/projection/zz_agentic_shim_t3.go
- delete internal/projection/zz_agentic_shim_t4.go
- delete internal/projection/zz_agentic_shim_t5.go
- delete internal/projection/zz_agentic_shim_t6.go
### Result
| Check | Status | Detail |
|---|---|---|
| stubs panic+TODO | PASS | 12 symbols; all bodies panic("SUB-AGENT-TODO: ...") |
| `GOTOOLCHAIN=auto go vet ./internal/projection/` | PASS | compiles clean |
| no newly passing tests | PASS | ran each test on its own: 17/18 FAIL; only TestDepsStdlibOnly passes (known PASS-ON-RED guard). The contract test's method-set part holds, but the test fails on the ReadDir panic |
| `grep -rn agentic:shim internal/projection` | PASS | empty |
| each symbol once | PASS | no definition outside internal/projection |
### Next
green: fill the stubs task by task (T2 validateName, T3 Build/node representation, T4 Lookup, T5 ReadDir, T6 Readlink), then run the full standards matrix and the >=80% coverage check.

## S2-02/T2 · attempt 1 · green-worker · 2026-09-22T02:40:00Z
status: ok
### Summary
Filled `validateName` in `internal/projection/name.go`: returns `fmt.Errorf("%q: %w", name, ErrInvalidName)` for "", ".", "..", or any name containing '/' or NUL (via `strings.ContainsAny(name, "/\x00")`); nil otherwise. `ErrInvalidName` message is now "invalid name". Stdlib only. Diff confined to SCOPE. Commit 333187e.
### Result
| Check | Status | Detail |
|---|---|---|
| `internal/projection/name_test.go::TestValidateName` | PASS | 9 accept + 6 reject subtests green under -race |
| `internal/projection/deps_test.go::TestDepsStdlibOnly` | PASS | still green |
| `go vet ./...` | PASS | clean |
| `golangci-lint run ./internal/projection/...` | NOTE | 1 staticcheck SA4008 at `lookup_test.go:128` — pre-existing in T4 RED test (a640c37), outside T2 scope; nothing in name.go flagged |
| diff scope | PASS | only `internal/projection/name.go` |
### Next
T3 (spec types and Build). Flag SA4008 in `lookup_test.go:128` (`for i := 0; i < loops; i++` where `loops` is constant/unchanged) to the T4 owner/orchestrator before T6's full matrix.

## S2-02/T3 · attempt 1 · green-worker · 2026-09-22T02:33:40Z
status: ok
### Summary
Filled the T3 stubs in `spec.go` and `build.go` (commit 891cc7d on top of 333187e). `spec.go` gains an unexported `node` type: `isDir`, `target`, pre-sorted `names`, and a `children` map from name to inode. `Generation` holds `nodes map[uint64]node`. `Build` walks the spec recursively. For each directory it validates every child name, rejects duplicate siblings (dir and link names share one namespace) and wraps both errors as `projection: %q: %w` with the slash-joined path. It then sorts the children by byte order and numbers them depth-first (root = `RootIno` = 1, the rest from 2 in pre-order by name). It copies the names and link targets into fresh nodes, so later changes to the spec cannot reach a built generation. It returns a nil `*Generation` on any error. `ErrDuplicateName` now reads "duplicate name". Lookup, ReadDir and Readlink are untouched (T4-T6).
### Result
| Check | Status | Detail |
|---|---|---|
| `internal/projection/build_test.go::TestBuildRejectsInvalidNames` | PASS | green (-race) |
| `internal/projection/build_test.go::TestBuildRejectsDuplicateSiblings` | PASS | green (-race) |
| `internal/projection/build_test.go::TestBuildErrorWrapsWithPath` | PASS | green (-race) |
| `internal/projection/name_test.go::TestValidateName` | PASS | still green |
| `internal/projection/deps_test.go::TestDepsStdlibOnly` | PASS | still green (only errors/fmt/sort/strings) |
| `go vet ./...` | PASS | clean |
| `golangci-lint run ./internal/projection/...` | NOTE | 3 staticcheck findings, all in T4/T5 RED test files (`lookup_test.go:90` SA4006, `lookup_test.go:128` SA4008, `readdir_test.go:29` SA4006). They follow calls to the catalog.go methods that still panic, and were already there at base; none are in spec.go or build.go |
| diff scope | PASS | only `internal/projection/spec.go`, `internal/projection/build.go` |
### Next
T4 (Lookup / inode identity) can read `g.nodes[parent].children[name]` and `isDir`. T5's ReadDir copies `names`. T6's Readlink reads `target`. Re-check the staticcheck findings once T4/T5 fill the catalog stubs.

## S2-02/T4 · attempt 1 · green-worker · 2026-09-22T02:35:22Z
status: ok
### Summary
Filled only `(*Generation).Lookup` in internal/projection/catalog.go: look up the parent in `g.nodes`; if the parent is missing or is not a directory, return not-found; otherwise resolve the name through `children` and report the child's `isDir`. ReadDir and Readlink are still SUB-AGENT-TODO stubs for T5 and T6. Commit 6a7e6f8 on worktree-agent-a57314d745f8b7d74 (base 891cc7d).
### Result
| Check | Status | Detail |
|---|---|---|
| `internal/projection/lookup_test.go::TestLookup` (10 subtests) | PASS | -race |
| `internal/projection/lookup_test.go::TestInodesStableWithinGeneration` | PASS | -race |
| `internal/projection/lookup_test.go::TestBuildIsDeterministic` | PASS | -race |
| `internal/projection/lookup_test.go::TestGenerationsAreIndependent` | PASS | -race |
| `internal/projection/lookup_test.go::TestConcurrentReadsDuringRebuild` | PASS | -race |
| Earlier S2-02 tests (TestBuildRejects*, TestBuildErrorWrapsWithPath, TestValidateName, TestDepsStdlibOnly) | PASS | -race |
| TestGenerationSatisfiesCatalogContract, TestGenerationsIndependentTargets | EXPECTED FAIL | T6 scope; they panic in the ReadDir/Readlink stubs |
| `go vet ./...` | PASS | clean |
| `golangci-lint run ./internal/projection/...` | 1 finding (test file) | readdir_test.go:29:2 SA4006: value of `want` never used (staticcheck). Outside green scope, nothing in catalog.go |
### Next
Structural review, then T5 (ReadDir). Test owner should fix the SA4006 in readdir_test.go:29.
