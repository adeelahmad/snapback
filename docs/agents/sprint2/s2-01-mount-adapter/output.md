---
type: output
story: S2-01
---

## S2-01/T1 · attempt 1 · red-worker · 2026-09-22T02:20:00Z
status: ok
### Summary
Wrote the three T1 tests from plan.md at the exact path::fn, no shims (tasks.md T1: "RED shims: none"; both test files are test-only packages that compile without production code). Each fails by assertion on the chain base 2158ade: go.mod has no go-fuse require line, go.sum does not exist, and `go list -deps` of the empty gofuse package lacks the go-fuse fs/fuse packages. `go vet ./...` clean, `golangci-lint run ./internal/mount/...` 0 issues, gofmt clean. Diff vs BASE_REF is exactly the two test files. Commit 3bfcc97 on branch worktree-agent-ac0f10b9b2735d9a3.
### Result
| Check | Status | Detail |
|---|---|---|
| `internal/mount/mount_test.go::TestGoModPinsGoFuseExactTag` | FAIL | go.mod require lines for github.com/hanwen/go-fuse/v2 = 0, want exactly 1 |
| `internal/mount/mount_test.go::TestGoSumHasGoFuseHashes` | FAIL | read go.sum: open ../../go.sum: no such file or directory (t.Fatalf) |
| `internal/mount/gofuse/attr_test.go::TestGofuseDepsIncludeFsAndFuse` | FAIL | go list -deps lacks github.com/hanwen/go-fuse/v2/fs and .../fuse |
| `go vet ./...` | PASS | clean |
| `golangci-lint run ./internal/mount/...` | PASS | 0 issues |
### Next
GREEN: `go get github.com/hanwen/go-fuse/v2@v2.11.0`, create `internal/mount/gofuse/attr.go` (package doc + blank imports of go-fuse fs and fuse), `go mod tidy`, confirm `CGO_ENABLED=0 go build ./...`. No scaffold symbols needed.

## S2-01/T2 · attempt 1 · red-worker · 2026-09-22T02:18:42Z
status: ok
### Summary
Added the four T2 tests from plan.md to `internal/mount/mount_test.go`, along with the test fakes (`fakeCatalog`, `fakeAdapter`, `recordingObserver`) and compile-time `var _ Catalog/Adapter/Observer` assertions. All T2 declarations live in the marked shim `internal/mount/zz_agentic_shim_t2.go`. It uses the binding predeclared-type Catalog signatures and `RootIno uint64 = 1` exactly as tasks.md states. Kind/Op constants are deliberately 0 and `Op.String()` returns "". TestOpString and TestKindAndRootConstants fail by assertion. TestMountDepsExcludeGoFuse and TestFakesSatisfyInterfaces are PASS-ON-RED guards (M-003), as tasks.md § T2 predicts. T1's go.mod/go.sum tests still fail, which is expected (go.mod is GREEN work). Commit 3cf9adb on top of BASE 3bfcc97. Note: the `chain2/s2-01` ref points at 2158ade, not BASE, and BASE lives on `chain/s2-01`. The worktree was reset to 3bfcc97 directly.
### Result
| Check | Status | Detail |
|---|---|---|
| `internal/mount/mount_test.go::TestMountDepsExcludeGoFuse` | PASS-ON-RED | seam guard (M-003): mount has no go-fuse deps; non-empty list includes the package itself (M-002) |
| `internal/mount/mount_test.go::TestOpString` | FAIL | `Op(0).String() = "", want "lookup"`; valid ops zero and not distinct |
| `internal/mount/mount_test.go::TestKindAndRootConstants` | FAIL | `KindDir = 0, want 1`; `KindSymlink = 0, want 2` |
| `internal/mount/mount_test.go::TestFakesSatisfyInterfaces` | PASS-ON-RED | compile-time interface guard plus fake behaviour (M-003); must not be weakened |
| `go vet ./...` | PASS | clean |
| `golangci-lint run ./internal/mount/...` | PASS | 0 issues |
| diff vs BASE | PASS | only `internal/mount/mount_test.go` and `internal/mount/zz_agentic_shim_t2.go` |
### Next
SCAFFOLD/GREEN: move the declarations from `zz_agentic_shim_t2.go` into `internal/mount/mount.go` (stdlib only) and delete the shim. Set `KindDir=1`, `KindSymlink=2`, and `OpLookup/OpReadDir/OpReadlink` = 1/2/3. Implement `Op.String()` (lookup/readdir/readlink, else unknown).

## S2-01/T3 · attempt 1 · red-worker · 2026-09-22T02:35:00Z
status: escalate
### Summary
Wrote all 8 T3 bullets from plan.md § T3 in `internal/mount/gofuse/attr_test.go`, keeping T1's `TestGofuseDepsIncludeFsAndFuse`. Added one marked shim, `internal/mount/gofuse/zz_agentic_shim_t3.go`, with the planned signatures, zero-value bodies, all four constants set to 0, `ReadOnlyErrno`=0 and `DaemonOwner`=MaxUint32, so the tests cannot pass by accident. The planned API uses go-fuse types (`fuse.Owner`, `fuse.Attr`, `fuse.EntryOut`, `fuse.AttrOut`, `fs.StableAttr`), so the tests have to import go-fuse. go-fuse is not in go.mod on base 3cf9adb, because T1 GREEN has not merged. The package therefore does not compile on this base. The exact error from `GOTOOLCHAIN=auto go test ./internal/mount/gofuse/`:
`zz_agentic_shim_t3.go:12:2: no required module provides package github.com/hanwen/go-fuse/v2/fs` (the same error appears for `.../v2/fuse`), `[setup failed]`. go.mod was not edited. Dry run: a throwaway copy in the scratchpad, with `go get github.com/hanwen/go-fuse/v2@v2.11.0` and `go mod tidy`, compiles. In that copy all 8 T3 tests FAIL by assertion, `go vet` is clean and `golangci-lint run ./internal/mount/...` reports 0 issues.
### Result
| Check | Status | Detail |
|---|---|---|
| compile on base 3cf9adb | ESCALATE | go-fuse not in go.mod (T1 GREEN prerequisite): "no required module provides package github.com/hanwen/go-fuse/v2/fs" |
| `internal/mount/gofuse/attr_test.go::TestStableAttr` | FAIL | (with go-fuse) Ino = 0, want 1; Mode = 0, want 040000 |
| `internal/mount/gofuse/attr_test.go::TestAttr` | FAIL | (with go-fuse) Mode type = 0, want 040000; Owner {0 0}; Nlink = 0 |
| `internal/mount/gofuse/attr_test.go::TestAttrUnknownKindHasNoMode` | FAIL | (with go-fuse) Ino = 0, want 4 |
| `internal/mount/gofuse/attr_test.go::TestEntryOutTimeouts` | FAIL | (with go-fuse) EntryTimeout() = 0s, want > 0 |
| `internal/mount/gofuse/attr_test.go::TestAttrOutTimeout` | FAIL | (with go-fuse) Timeout() = 0s, want > 0 |
| `internal/mount/gofuse/attr_test.go::TestTimeoutsBounded` | FAIL | (with go-fuse) EntryTimeout = 0s, want in (0, 1m] |
| `internal/mount/gofuse/attr_test.go::TestReadOnlyErrnoIsEROFS` | FAIL | (with go-fuse) ReadOnlyErrno() = 0, want EROFS |
| `internal/mount/gofuse/attr_test.go::TestDaemonOwner` | FAIL | (with go-fuse) DaemonOwner() = {4294967295 4294967295}, want 501/20 |
| diff vs BASE | PASS | only `attr_test.go` and `zz_agentic_shim_t3.go` (commit 4746ea7) |
### Next
Orchestrator: rebase this commit onto a chain that has T1 GREEN (go-fuse v2.11.0 in go.mod/go.sum) and T2 GREEN (real `mount.KindDir`/`KindSymlink`), then re-verify that the 8 T3 tests FAIL by assertion. SCAFFOLD then moves the shim declarations into `internal/mount/gofuse/attr.go`, replacing T1's blank imports, and deletes `zz_agentic_shim_t3.go`.

## S2-01/scaffold · attempt 1 · scaffolder · 2026-09-22T02:24:28Z
status: ok
### Summary
Moved every T2 shim symbol from `internal/mount/zz_agentic_shim_t2.go` into `internal/mount/mount.go`, each declared exactly once, and deleted the shim. Types and interfaces keep their binding shapes (Catalog signatures, `RootIno uint64 = 1`); `KindDir`/`KindSymlink`/`OpLookup`/`OpReadDir`/`OpReadlink` stay declared as placeholder 0 marked `// SUB-AGENT-TODO`; `Op.String` body is `panic("SUB-AGENT-TODO: ...")`. go.mod/go.sum and internal/mount/gofuse untouched; no test edited. Commit d41aa8b on chain base 3cf9adb. Symbol list in `.agentic/scaffold-symbols`.
### Scaffold
+ create Kind (type) @ internal/mount/mount.go
+ create KindDir, KindSymlink (const, placeholder TODO) @ internal/mount/mount.go
+ create Entry (struct) @ internal/mount/mount.go
+ create RootIno (const = 1) @ internal/mount/mount.go
+ create Op (type) @ internal/mount/mount.go
+ create OpLookup, OpReadDir, OpReadlink (const, placeholder TODO) @ internal/mount/mount.go
+ create Op.String (panic stub) @ internal/mount/mount.go
+ create Event (struct) @ internal/mount/mount.go
+ create Catalog, Adapter, Observer (interfaces) @ internal/mount/mount.go
- delete internal/mount/zz_agentic_shim_t2.go (shim)
### Result
| Check | Status | Detail |
|---|---|---|
| go vet ./internal/mount/ | PASS | compiles |
| stubs panic+TODO | PASS | 1 func stub (Op.String), 5 constants TODO-marked, 8 types/interfaces |
| no newly passing test | PASS | FAIL: TestGoModPinsGoFuseExactTag, TestGoSumHasGoFuseHashes, TestOpString (panic), TestKindAndRootConstants; PASS-ON-RED guards TestMountDepsExcludeGoFuse, TestFakesSatisfyInterfaces still pass (as at RED) |
| agentic:shim in internal/mount/*.go | PASS | none |
| each symbol once | PASS | single declaration file; `Op.String` omitted from .agentic/scaffold-symbols because the name-based gate counts the unrelated pre-existing `internal/version.String()` (method is defined once on Op) |
| go.mod / gofuse untouched | PASS | diff is shim -> mount.go rename only |
### Next
green: T1 pins go-fuse in go.mod/go.sum; T2 fills constants and Op.String in internal/mount/mount.go. T3 RED (4746ea7) to be replayed after GREEN T1.

## S2-01/T1 · attempt 1 · green-worker · 2026-09-22T00:00:00Z
status: ok
### Summary
Ran `go get github.com/hanwen/go-fuse/v2@v2.11.0`, created `internal/mount/gofuse/attr.go` as a placeholder that blank-imports `go-fuse/v2/fs` and `go-fuse/v2/fuse` so the requirement stays direct through `go mod tidy`, then ran tidy. go.mod now has `require github.com/hanwen/go-fuse/v2 v2.11.0` (direct) and `golang.org/x/sys v0.28.0 // indirect`. mount.go was not touched. Commit 6658766 on top of d41aa8b.
### Result
| Check | Status | Detail |
|---|---|---|
| `internal/mount/mount_test.go::TestGoModPinsGoFuseExactTag` | PASS | green |
| `internal/mount/mount_test.go::TestGoSumHasGoFuseHashes` | PASS | green |
| `internal/mount/gofuse/attr_test.go::TestGofuseDepsIncludeFsAndFuse` | PASS | green |
| `internal/mount/mount_test.go::TestMountDepsExcludeGoFuse` | PASS | still green |
| `CGO_ENABLED=0 go build ./...` | PASS | ok |
| `go mod tidy -diff` | PASS | empty |
| `go vet ./...` | PASS | clean |
| `golangci-lint run ./internal/mount/...` | PASS | 0 issues |
| `govulncheck ./...` | PASS | 0 vulns reachable. go-fuse v2.11.0 has no known vulns. There is 1 module-level vuln: GO-2026-5024 in golang.org/x/sys@v0.28.0 (indirect, windows only, fixed in v0.44.0), and no code path reaches it |
### Next
T2 (mount.go) runs in parallel. Structural review after merge. Optional: bump golang.org/x/sys to >= v0.44.0 to clear the windows-only module advisory.

## S2-01/T2 · attempt 1 · green-worker · 2026-09-22T02:27:23Z
status: ok
### Summary
Filled the SUB-AGENT-TODO stubs in internal/mount/mount.go only: KindDir=1, KindSymlink=2, OpLookup=1, OpReadDir=2, OpReadlink=3, and Op.String as a switch returning lookup/readdir/readlink/unknown, per tasks.md decisions. No imports added, no test edits. Commit b4696ef on worktree-agent-a7617cd7b81d6b128 (base d41aa8b). T1's go.mod/go.sum tests still fail on this branch as expected (T1 runs in parallel and owns go.mod/go.sum).
### Result
| Check | Status | Detail |
|---|---|---|
| `internal/mount/mount_test.go::TestOpString` | PASS | green |
| `internal/mount/mount_test.go::TestKindAndRootConstants` | PASS | green |
| `internal/mount/mount_test.go::TestMountDepsExcludeGoFuse` | PASS | still passing (PASS-ON-RED guard) |
| `internal/mount/mount_test.go::TestFakesSatisfyInterfaces` | PASS | still passing (PASS-ON-RED guard) |
| `internal/mount/mount_test.go::TestGoModPinsGoFuseExactTag` | FAIL | expected: T1 scope (go.mod), not on this branch |
| `internal/mount/mount_test.go::TestGoSumHasGoFuseHashes` | FAIL | expected: T1 scope (go.sum), not on this branch |
| `go vet ./...` | PASS | clean |
| `golangci-lint run ./internal/mount/...` | PASS | 0 issues |
| gofmt / SUB-AGENT-TODO residue in mount.go | PASS | none |
### Next
Merge with T1, then T3 (gofuse attr translation, full matrix); structural review after merge.

## S2-01/scaffold-T3 · attempt 1 · scaffolder · 2026-09-22T02:32:11Z
status: ok
### Summary
Moved all 10 T3 shim symbols from internal/mount/gofuse/zz_agentic_shim_t3.go into internal/mount/gofuse/attr.go exactly once, replacing T1's blank go-fuse imports with real imports (syscall, time, fs, fuse, mount). Six functions have panic("SUB-AGENT-TODO: <recipe>") bodies; the four constants keep their name/type with placeholder 0 values marked // SUB-AGENT-TODO (target values in the comment). Shim deleted. .agentic/scaffold-symbols lists EntryTimeout, AttrTimeout, DirPerm, SymlinkPerm, DaemonOwner, ReadOnlyErrno; StableAttr, Attr, EntryOut and AttrOut are omitted because they collide by bare name with go-fuse types (fs.StableAttr, fuse.Attr, fuse.EntryOut, fuse.AttrOut). Commit 71cdb8b on base 2606f2c.
### Scaffold
+ update imports @ internal/mount/gofuse/attr.go (blank imports -> syscall, time, fs, fuse, mount)
+ create const EntryTimeout @ internal/mount/gofuse/attr.go
+ create const AttrTimeout @ internal/mount/gofuse/attr.go
+ create const DirPerm @ internal/mount/gofuse/attr.go
+ create const SymlinkPerm @ internal/mount/gofuse/attr.go
+ create func StableAttr @ internal/mount/gofuse/attr.go
+ create func Attr @ internal/mount/gofuse/attr.go
+ create func EntryOut @ internal/mount/gofuse/attr.go
+ create func AttrOut @ internal/mount/gofuse/attr.go
+ create func DaemonOwner @ internal/mount/gofuse/attr.go
+ create func ReadOnlyErrno @ internal/mount/gofuse/attr.go
- delete shim @ internal/mount/gofuse/zz_agentic_shim_t3.go
### Result
| Check | Status | Detail |
|---|---|---|
| stubs panic+TODO | PASS | 6 funcs panic; 4 consts placeholder 0 + SUB-AGENT-TODO |
| go vet ./internal/mount/... | PASS | GOTOOLCHAIN=auto |
| CGO_ENABLED=0 go build ./... | PASS | |
| go mod tidy -diff | PASS | empty |
| no newly passing T3 test | PASS | all 8 T3 tests FAIL run individually; T1 TestGofuseDepsIncludeFsAndFuse still passes |
| grep agentic:shim internal/ | PASS | empty |
| test files touched | PASS | none |
### Next
green: fill the T3 stubs in internal/mount/gofuse/attr.go per tasks.md § T3 (set consts to time.Second / 0o555, add os import for DaemonOwner), then run the full matrix (GATE_RUN_MATRIX=1).

## S2-01/T3 · attempt 1 · green-worker · 2026-09-22T02:35:19Z
status: escalate

### Summary
Filled the four constants (EntryTimeout/AttrTimeout = 1s, DirPerm/SymlinkPerm = 0o555) and six functions in internal/mount/gofuse/attr.go (StableAttr, Attr, EntryOut, AttrOut, DaemonOwner, ReadOnlyErrno=EROFS), plus one unexported typeBits helper used by StableAttr. Commit af79267 on top of 71cdb8b; the diff touches only attr.go. All 8 T3 tests and all S2-01 tests pass under -race. The full cross-cutting matrix is green EXCEPT `go test -race ./...`: test/projectdocs::TestNoticeCoversGoModRequires fails because NOTICE lacks the go.mod requires github.com/hanwen/go-fuse/v2 and golang.org/x/sys (added to go.mod by earlier S2-01 work, so already failing at 71cdb8b). NOTICE is outside SCOPE_GLOBS, so it is not fixed here; escalating for a scoped fix that adds both entries to NOTICE.

### Result
| Check | Status | Detail |
|---|---|---|
| `internal/mount/gofuse/attr_test.go::TestGofuseDepsIncludeFsAndFuse` | PASS | green |
| `internal/mount/gofuse/attr_test.go::TestStableAttr` | PASS | dir/symlink/unknown |
| `internal/mount/gofuse/attr_test.go::TestAttr` | PASS | dir/symlink |
| `internal/mount/gofuse/attr_test.go::TestAttrUnknownKindHasNoMode` | PASS | green |
| `internal/mount/gofuse/attr_test.go::TestEntryOutTimeouts` | PASS | dir/symlink |
| `internal/mount/gofuse/attr_test.go::TestAttrOutTimeout` | PASS | green |
| `internal/mount/gofuse/attr_test.go::TestTimeoutsBounded` | PASS | green |
| `internal/mount/gofuse/attr_test.go::TestReadOnlyErrnoIsEROFS` | PASS | green |
| `internal/mount/gofuse/attr_test.go::TestDaemonOwner` | PASS | green |
| `go test -race ./internal/mount/...` | PASS | mount + gofuse ok |
| gofmt / goimports | PASS | no output |
| `CGO_ENABLED=0 go build ./...` | PASS | exit 0 |
| `go vet ./...` | PASS | exit 0 |
| `golangci-lint run` | PASS | 0 issues |
| `go test -race ./...` | FAIL | test/projectdocs::TestNoticeCoversGoModRequires: NOTICE missing go-fuse/v2 and golang.org/x/sys (out of scope, pre-existing at 71cdb8b) |
| coverage gate | PASS | 96.9% (>=80%) |
| `govulncheck ./...` | PASS | exit 0; 0 called vulns; 1 module-level finding in a required module, not called (informational) |
| actionlint | PASS | exit 0 |
| `shellcheck -s sh install.sh` | PASS | exit 0 |
| `mkdocs build --strict` | PASS | exit 0 |
| `goreleaser check` | PASS | exit 0 |

### Next
Orchestrator: add a scoped fix for NOTICE (list github.com/hanwen/go-fuse/v2 and golang.org/x/sys with their licenses) so test/projectdocs passes, then structural review and merge of af79267.

## S2-01/fix-notice · attempt 1 · green-worker · 2026-09-22T02:42:07Z
status: ok
### Summary
Added `github.com/hanwen/go-fuse/v2` and `golang.org/x/sys` to NOTICE in the existing `- <name> — <licence> (<url>)` format, keeping the "updated as dependencies are added" sentence. Both licences were checked against the LICENSE files in the module cache (go-fuse v2.11.0: "New BSD License", three clauses; x/sys v0.28.0: Go Authors three-clause BSD), so both are BSD-3-Clause. NOTICE was the only file changed. Commit 968e890 on top of 36dd8a7, authored by Adeel Ahmad with no AI trailers.
### Result
| Check | Status | Detail |
|---|---|---|
| `test/projectdocs::TestNoticeListsGoStdlib` | PASS | green |
| `test/projectdocs::TestNoticeSaysListIsMaintained` | PASS | green |
| `test/projectdocs::TestNoticeCoversGoModRequires` | PASS | was red, now green |
| gofmt / goimports | PASS | no output |
| `CGO_ENABLED=0 go build ./...` | PASS | |
| `go vet ./...` | PASS | |
| `golangci-lint run` | PASS | 0 issues |
| `go test -race ./...` | PASS | all packages ok |
| coverage | PASS | 96.9% total (>=80%) |
| `govulncheck ./...` | PASS | 0 called vulnerabilities (1 in a required module, not reachable from our code) |
| actionlint / shellcheck -s sh install.sh | PASS | |
| `mkdocs build --strict` | PASS | |
| `goreleaser check` | PASS | config validated |
### Next
Structural review, then the S2-01 final gate / merge.

## S2-01/structural · attempt 1 · structural-reviewer · 2026-09-22T02:56:46Z

status: ok

### Summary
Reviewed `internal/mount/mount.go`, `internal/mount/gofuse/attr.go` and their tests plus the go.mod/go.sum/NOTICE deltas against `e83fd2f` using `git grep` and `go vet`. The mount seam and gofuse attribute translator are each defined exactly once, no shim/TODO leftovers remain in `internal/`, `go vet ./internal/mount/...` is clean, and the go-fuse/x-sys dependency additions are mirrored correctly in NOTICE. `mount.Adapter`, `mount.Observer`, `mount.Catalog` and `mount.Event`/`mount.Op` are not yet consumed by any production code outside `internal/mount` itself (only exercised via test-only fakes and a structural shape-pin in `internal/projection`), but this matches S2-01's stated scope as a seam-definition story consumed by later sprint-2 tasks (S2-02/S2-04), so it is noted as informational, not a defect.

### Result
| Check | Status | Detail |
|---|---|---|
| orphan / unused-exported symbols | PASS (informational) | `mount.Adapter`/`Observer`/`Catalog`/`Event`/`Op` have no production consumer yet outside `internal/mount`'s own tests and `internal/projection/contract_test.go`'s structural shape-pin comment; expected for a seam-only story per init.md (T2 mandate: "mount.go types + Catalog/Adapter/Observer interfaces" ahead of consumers). `mount.Entry`/`mount.Kind` are consumed by `internal/mount/gofuse/attr.go`. No dead/unreachable code found. |
| parallel implementations | PASS | No second FUSE-attribute or catalog-seam implementation found elsewhere in the repo (`internal/compat`, `internal/version`, `internal/projection` do not duplicate this abstraction; projection's catalog interface is a local structural pin, not a competing implementation). |
| duplicate helpers | PASS | `goModRequireLines` (mount_test.go), `typeBits`, `DaemonOwner`, and the EntryTimeout/AttrTimeout/DirPerm/SymlinkPerm/EROFS constants are each defined exactly once repo-wide (`git grep` confirms no second definition). |
| shim / TODO leftovers | PASS | No `zz_agentic_shim`, `agentic:shim`, or `SUB-AGENT-TODO` matches anywhere under `internal/` (file search and content grep both empty). |
| EROFS / timeout constants defined once | PASS | `ReadOnlyErrno()` returns `syscall.EROFS` (single definition, attr.go:72); `EntryTimeout`, `AttrTimeout`, `DirPerm`, `SymlinkPerm` each declared once in the attr.go const block (lines 17-20) and referenced only, never redeclared. |
| go vet | PASS | `go vet ./internal/mount/...` produced no output. |

### Findings
none

### Next
continue

## S2-01/structural · attempt 1 · structural-reviewer · 2026-09-22T02:59:47Z

status: ok

### Summary
Ran the mandatory `selfcheck` (`gate-structural-integrity`) after the manual git-grep/go-vet review above. It exited 2, but the failure is dominated by scope mismatch, not a defect introduced by S2-01: `BASE_REF=e83fd2f` predates the merge of S2-01, S2-02 and S2-03 alike (45 commits ahead), so `ctx-symbols conflicts --tree` diffs the whole merged worktree against a tree that has none of sprint 2, making every sprint-2 symbol collision (across three different stories) look "new." Manually triaging the reported HIGH lines that touch an S2-01 file: `Lookup`/`ReadDir`/`Readlink` (mount_test.go's `fakeCatalog` vs. projection's real `Generation`) and `Mount` (mount_test.go's `fakeAdapter` vs. resticfx's real adapter) are expected — a fake and a real implementation of the same interface necessarily share method names; `String` (mount.go's `Op.String` vs. internal/version's `String`) is the ordinary `fmt.Stringer` idiom; bare-identifier collisions (`_`, `out`, `unmounts`) are coincidental local-variable names in unrelated functions, not shared implementations. One finding is a genuine duplicate worth flagging: `RootIno uint64 = 1` is declared verbatim in both `internal/mount/mount.go:22` (S2-01, in scope) and `internal/projection/spec.go:12` (S2-02, out of scope) — projection re-declares the constant instead of importing `mount.RootIno`, so the two can drift silently. This is a downstream-consumer choice in S2-02, not a defect S2-01 introduced, so it does not change S2-01's own status, but it is exactly the "duplicate helper under a different name" this audit watches for and should be raised to whoever reviews S2-02/S2-04 (the gofuse-catalog wiring is the natural place to import `mount.RootIno` instead).

### Result
| Check | Status | Detail |
|---|---|---|
| selfcheck / gate-structural-integrity | FAIL (exit 2), triaged as non-blocking for S2-01 | `BASE_REF=e83fd2f` predates all of sprint 2 (S2-01+S2-02+S2-03, 45 commits), so the baseline diff cannot isolate S2-01-only introductions; every cross-story HIGH line reported is either (a) an interface/fake pair, (b) the Stringer idiom, (c) a coincidental local-variable name, or (d) the genuine `RootIno` duplicate below. |
| duplicate helper (RootIno) | FAIL (cross-story, out of S2-01's write scope) | `internal/mount/mount.go:22` and `internal/projection/spec.go:12` both declare `const RootIno uint64 = 1` independently instead of projection importing `mount.RootIno`. |

### Findings
- MEDIUM `internal/projection/spec.go:12` duplicates `internal/mount/mount.go:22`'s `RootIno uint64 = 1` instead of importing it — flag for S2-02/S2-04 follow-up, not an S2-01 defect.
- (all other findings from this review's earlier block still stand: none within S2-01's own files)

### Next
continue — S2-01 itself is clean; route the RootIno duplicate finding to the S2-02/S2-04 chain (projection/gofuse-catalog) as a follow-up so a later task imports `mount.RootIno` instead of redeclaring it. Re-run `gate-structural-integrity` with a `BASE_REF` at or after the sprint-2 merge point once available, since the current baseline cannot distinguish this story's introductions from its siblings'.
