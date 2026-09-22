---
type: plan
story: S2-01
scope: "tests only"
---

# S2-01 test plan (tests only)

The contracts under test are fixed in `tasks.md`. Paths are relative to the repo root. Tests read files through `../../` (mount) or `../../../` (gofuse) relative paths and run `go list` as a subprocess. None mounts anything or needs FUSE, macFUSE or `/dev/fuse`.

## T1 — go-fuse pin

- [ ] `internal/mount/mount_test.go::TestGoModPinsGoFuseExactTag` — input `../../go.mod`; action: find the require line for `github.com/hanwen/go-fuse/v2`; assertion: exactly one such line, version matches `^v\d+\.\d+\.\d+$` and equals `v2.11.0`, no `// indirect`, no pseudo-version (`-0.`), no `replace` directive naming go-fuse.
- [ ] `internal/mount/mount_test.go::TestGoSumHasGoFuseHashes` — input `../../go.sum`; action: read the file (fatal with the error if it is missing); assertion: the file is non-empty (M-002) and contains both `github.com/hanwen/go-fuse/v2 v2.11.0 h1:` and `github.com/hanwen/go-fuse/v2 v2.11.0/go.mod h1:`.
- [ ] `internal/mount/gofuse/attr_test.go::TestGofuseDepsIncludeFsAndFuse` — input package `github.com/adeelahmad/snapback/internal/mount/gofuse`; action: run `go list -deps` and split the output into lines; assertion: the command succeeds, the list is non-empty (M-002), and it contains both `github.com/hanwen/go-fuse/v2/fs` and `github.com/hanwen/go-fuse/v2/fuse`.

## T2 — mount interfaces and types

- [ ] `internal/mount/mount_test.go::TestMountDepsExcludeGoFuse` — input package `github.com/adeelahmad/snapback/internal/mount`; action: run `go list -deps`; assertion: the command succeeds, the list is non-empty and contains `github.com/adeelahmad/snapback/internal/mount` itself (M-002 guard), and no line contains `hanwen/go-fuse`.
- [ ] `internal/mount/mount_test.go::TestOpString` — table: `OpLookup`, `OpReadDir`, `OpReadlink`, `Op(0)`, `Op(255)`; action: call `String()`; assertion: `lookup`, `readdir`, `readlink`, `unknown`, `unknown`, and the three valid ops are distinct and non-zero.
- [ ] `internal/mount/mount_test.go::TestKindAndRootConstants` — input: the exported constants; action: compare them; assertion: `KindDir == 1`, `KindSymlink == 2` (distinct and non-zero, so the zero `Entry` is invalid), and `RootIno == 1`.
- [ ] `internal/mount/mount_test.go::TestFakesSatisfyInterfaces` — input: `fakeCatalog` (root 1 has a `docs` directory at ino 2 and a `link` symlink at ino 3 targeting `../x`), `fakeAdapter`, and `recordingObserver`, with compile-time `var _ Catalog/Adapter/Observer = ...` assertions; action: call `Lookup(RootIno,"docs")`, `ReadDir(RootIno)`, `Readlink(3)`, `Readlink(2)` and `Lookup(RootIno,"nope")` through a `Catalog` variable, then `Mount(t.TempDir(), cat)` and `Unmount()` through an `Adapter` variable, then `Observe(Event{OpLookup,"/docs"})` through an `Observer` variable; assertion: the lookups return `(2,true,true)` and `(0,false,false)`, `ReadDir` returns `["docs","link"]` and true, `Readlink(3)` returns `("../x",true)`, `Readlink(2)` returns `found=false`, the adapter recorded the mount dir and one unmount, and the observer recorded exactly one event equal to the one sent (PASS-ON-RED guard, M-003).

## T3 — gofuse attribute translation

- [ ] `internal/mount/gofuse/attr_test.go::TestStableAttr` — table: `Entry{Ino:1,Kind:KindDir}`, `Entry{Ino:7,Kind:KindSymlink}`, `Entry{Ino:9,Kind:0}`; action: call `StableAttr`; assertion: `Ino` is echoed, `Mode` is `syscall.S_IFDIR`, `syscall.S_IFLNK` and `0` respectively.
- [ ] `internal/mount/gofuse/attr_test.go::TestAttr` — table: a directory and a symlink entry, with `owner := fuse.Owner{Uid:501,Gid:20}`; action: call `Attr(e, owner)`; assertion: `Mode&0o222 == 0` (no write bits), `Mode&syscall.S_IFMT` is the matching type bit, `Mode&0o777` equals `DirPerm`/`SymlinkPerm`, `Owner == owner`, `Ino == e.Ino`, `Nlink == 1`.
- [ ] `internal/mount/gofuse/attr_test.go::TestAttrUnknownKindHasNoMode` — input `Entry{Ino:4,Kind:0}`; action: call `Attr`; assertion: `Mode == 0`, so an invalid entry never gets permission bits.
- [ ] `internal/mount/gofuse/attr_test.go::TestEntryOutTimeouts` — table: directory and symlink entries; action: call `EntryOut(e, owner)`; assertion: `EntryTimeout() == EntryTimeout`, `AttrTimeout() == AttrTimeout`, and `out.Attr` equals `Attr(e, owner)`.
- [ ] `internal/mount/gofuse/attr_test.go::TestAttrOutTimeout` — input: a directory entry; action: call `AttrOut(e, owner)`; assertion: `Timeout() == AttrTimeout` and `out.Attr` equals `Attr(e, owner)`.
- [ ] `internal/mount/gofuse/attr_test.go::TestTimeoutsBounded` — input: the constants `EntryTimeout` and `AttrTimeout`; action: compare them with the bounds; assertion: each is `> 0` and `<= time.Minute` (bounded lifetimes, SPEC §5), and `DirPerm&0o222 == 0` and `SymlinkPerm&0o222 == 0`.
- [ ] `internal/mount/gofuse/attr_test.go::TestReadOnlyErrnoIsEROFS` — input: none; action: call `ReadOnlyErrno()`; assertion: it equals `syscall.EROFS`, and neither `syscall.EPERM` nor `syscall.ENOSYS`.
- [ ] `internal/mount/gofuse/attr_test.go::TestDaemonOwner` — input: the process identity; action: call `DaemonOwner()`; assertion: `Uid == uint32(os.Getuid())` and `Gid == uint32(os.Getgid())`.

Test count: 15.
