---
type: plan-ready
story: S2-02
from_red_at: 2026-09-22T02:25:26Z
---

# S2-02 plan-ready (RED verified by orchestrator at chain @ e51df7e; every box currently FAILs by assertion unless noted)


# S2-02 test plan (tests only)

The contracts under test are fixed in `tasks.md`. All tests are in-package (`package projection`) unit tests. Run them with `go test -race ./internal/projection/`. Every test must fail by assertion on RED, with the shims listed per task in `tasks.md`. `fixtureSpec()` and `fullID` live in `internal/projection/fixture_test.go` (created in T4).

## T1 — Package skeleton and dependency guard

- [ ] `internal/projection/deps_test.go::TestDepsStdlibOnly` — input package `github.com/adeelahmad/snapback/internal/projection`; action: run `go list -deps` (fatal with the output if it fails) and split the output into lines; assertion: the list is non-empty and contains the projection package itself (M-002 guard); no line contains `hanwen/go-fuse` or `internal/mount`; every line other than the projection package has no `.` in its first path element (standard library only).

## T2 — Name validation

- [x] `internal/projection/name_test.go::TestValidateName` — table: accepted `a`, `latest`, `.hidden`, `a b`, `...`, `UPPER`, `Ünïcode`, `2026-09-22T10-00-00Z`, a 64-hex ID; rejected `""`, `.`, `..`, `a/b`, `/`, `a\x00b`; action: call `validateName`; assertion: accepted rows return nil; rejected rows return an error with `errors.Is(err, ErrInvalidName)` whose message contains `%q` of the name.

## T3 — Spec types and Build validation

- [x] `internal/projection/build_test.go::TestBuildRejectsInvalidNames` — table: an invalid name (`""`, `.`, `..`, `a/b`, `a\x00b`) placed as a root dir, a root link, a nested dir under `docs`, and a nested link under `docs`; action: `Build(spec)`; assertion: `errors.Is(err, ErrInvalidName)`, the returned generation is nil, and nested rows' error text contains the parent path `docs`.
- [x] `internal/projection/build_test.go::TestBuildRejectsDuplicateSiblings` — table: two root dirs named `x`; two root links named `x`; root dir `x` plus root link `x`; a duplicate inside `docs`; plus a control row with `x` under two different parents; action: `Build(spec)`; assertion: duplicate rows give `errors.Is(err, ErrDuplicateName)`, a nil generation and error text containing `"x"`; the control row gives a nil error and a non-nil generation.
- [x] `internal/projection/build_test.go::TestBuildErrorWrapsWithPath` — input: a spec with link `a\x00b` inside dir `docs`; action: `Build(spec)`; assertion: the error is non-nil, `errors.Is(err, ErrInvalidName)`, `errors.Unwrap` chain reaches the sentinel, and `err.Error()` contains `projection:` and the quoted path `"docs/a\x00b"`.

## T4 — Lookup, inode identity and generation independence

- [x] `internal/projection/lookup_test.go::TestLookup` — input: `Build(fixtureSpec())`; table: `(RootIno,"docs")` dir hit; `(RootIno,"rel")` symlink hit with `isDir=false`; `(RootIno,"missing")` miss; `(docsIno,"readme-link")` nested hit; `(relIno,"x")` lookup under a symlink parent -> miss; `(9999,"docs")` unknown parent -> miss; `(RootIno,"..")` and `(RootIno,".")` -> miss; `(snapshotsIno, fullID)` hit; `(snapshotsIno, fullID[:8])` -> miss (no short-prefix match); action: `Lookup`; assertion: `found` and `isDir` match per row, hits have `ino != 0` and `ino != RootIno`, misses return `ino == 0`.
- [x] `internal/projection/lookup_test.go::TestInodesStableWithinGeneration` — input: `Build(fixtureSpec())`; action: look up every fixture path twice (walking from `RootIno`); assertion: every lookup is found (M-002), both calls return the same inode, and all inodes are distinct and greater than `RootIno`.
- [x] `internal/projection/lookup_test.go::TestBuildIsDeterministic` — input: two independent `Build(fixtureSpec())` calls; action: look up every fixture path in both; assertion: each path is found in both (M-002) with the same inode in both generations.
- [x] `internal/projection/lookup_test.go::TestGenerationsAreIndependent` — input: `spec := fixtureSpec()`, `g1 := Build(spec)`, then append dir `new` to `spec.Dirs`, rename `spec.Dirs[0]` in place, and build `g2 := Build(spec)`; action: look up `new` and the original name in both; assertion: `g2` finds `new`; `g1` does not find `new` and still finds the original name, with the same inode as before the spec was mutated (Build deep-copies; generation N never sees N+1).
- [x] `internal/projection/lookup_test.go::TestConcurrentReadsDuringRebuild` — input: `g1 := Build(fixtureSpec())`; action: 8 goroutines each loop 100 times over `g1.Lookup(RootIno,"docs")` while the main goroutine builds 20 new generations from modified copies of the spec; `sync.WaitGroup`, run under `-race`; assertion: every `g1` lookup returned the same `(ino,true,true)` as the first, which is found (M-002), and the race detector reports nothing.

## T5 — Ordered directory listing

- [ ] `internal/projection/readdir_test.go::TestReadDirSortedByteOrder` — input: a spec whose root holds `b`, `B`, `a`, `Z`, `_`, `é`, `.dot` (mixed dirs and links, declared out of order); action: `ReadDir(RootIno)` 5 times; assertion: `found` is true, the result equals `[".dot","B","Z","_","a","b","é"]` (ascending byte order), and all 5 calls return identical slices.
- [ ] `internal/projection/readdir_test.go::TestReadDirNotFound` — input: `Build(fixtureSpec())`; table: control `RootIno` -> found with non-empty names (M-002); symlink `rel` ino, `0`, and `9999` -> not found; action: `ReadDir`; assertion: `found` matches per row, and not-found rows return `names == nil`.
- [ ] `internal/projection/readdir_test.go::TestReadDirReturnsCopy` — input: `Build(fixtureSpec())`; action: `names, _ := ReadDir(RootIno)`, overwrite `names[0] = "mutated"`, call `ReadDir(RootIno)` again; assertion: the first result was non-empty (M-002), and the second result does not contain `mutated` and equals the original sorted list.
- [ ] `internal/projection/readdir_test.go::TestReadDirEmptyDirectory` — input: `Build(fixtureSpec())`, `emptyIno` from `Lookup(RootIno,"empty")`; action: `ReadDir(emptyIno)`; assertion: the lookup found a directory (M-002), `found` is true, and names is non-nil with length 0.

## T6 — Readlink, the catalog contract, and the full matrix

- [ ] `internal/projection/readlink_test.go::TestReadlinkExactBytes` — table of link targets: `../target`, `/var/tmp/x`, `..`, `a//b/./c/`, `x/../y`, `trailing/`, ` spaced `, `Ünïcode/ß`, `../ids/<fullID>`; action: build a root holding one link per target, look each up, then call `Readlink`; assertion: each lookup is found with `isDir=false` (M-002), and `Readlink` returns `found=true` with target byte-equal to the input (no clean, resolve or trim; the 64-hex ID is not truncated).
- [ ] `internal/projection/readlink_test.go::TestReadlinkNotFound` — input: `Build(fixtureSpec())`; table: control `rel` ino -> `("a/b", true)` (M-002); `docs` dir ino, `RootIno`, `0`, `9999` -> not found; action: `Readlink`; assertion: results match per row, and not-found rows return `target == ""`.
- [ ] `internal/projection/readlink_test.go::TestGenerationsIndependentTargets` — input: `spec := fixtureSpec()`, `g1 := Build(spec)`, then change link `rel`'s target in place to `changed/target`, `g2 := Build(spec)`; action: `Readlink` on `rel` in each generation; assertion: `g1` returns `a/b` and `g2` returns `changed/target`, both found.
- [ ] `internal/projection/contract_test.go::TestGenerationSatisfiesCatalogContract` — input: a test-local `type catalog interface { Lookup(uint64, string) (uint64, bool, bool); ReadDir(uint64) ([]string, bool); Readlink(uint64) (string, bool) }` (defined here, not imported from `mount`), `var _ catalog = (*Generation)(nil)`, and `var c catalog = Build(fixtureSpec())`; action: through `c`, `ReadDir(RootIno)`, `Lookup(RootIno,"snapshots")`, `ReadDir(snapshotsIno)`, `Lookup(snapshotsIno, fullID)`, `Readlink(linkIno)`; then use `reflect.TypeOf((*Generation)(nil))` to list its exported methods; assertion: root listing is non-empty and contains `snapshots`; the snapshots listing is exactly `[fullID]` (64 characters); readlink returns `../ids/<fullID>`; the exported method set is exactly `{Lookup, ReadDir, Readlink}` (read-only surface: no mutator exists, so mutations are the adapter's EROFS, §7).

Test count: 18.
