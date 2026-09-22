---
type: tasks
story: S2-01
---

# S2-01 tasks — go-fuse pin and `mount` adapter seam

Story intent and owned files: `docs/agents/sprint2/stories.md` (S2-01). Every task touches only S2-01's owned set: `go.mod`, `go.sum`, `internal/mount/mount.go`, `internal/mount/mount_test.go`, `internal/mount/gofuse/attr.go`, `internal/mount/gofuse/attr_test.go`. Tests per task are in `plan.md`; gate commands are in `validate.md`.

Sequencing: T1 -> T2 -> T3, strictly serial. T1 and T2 share `internal/mount/mount_test.go`; T1 and T3 share `internal/mount/gofuse/attr.go` and `attr_test.go`; T3 imports T2's `mount.Entry`. No task may run in parallel with another.

Matrix rule (M-005): T1 and T2 GREEN run with `GATE_RUN_MATRIX=0` plus the package-scoped commands in `validate.md`. Only T3, the story's last task, runs the full standards matrix.

No test in this story mounts anything or needs `/dev/fuse` or macFUSE. Every test is pure Go, file reads, or `go list` subprocesses.

Decisions fixed here so workers do not guess:

- Pin: `github.com/hanwen/go-fuse/v2 v2.11.0`, a direct requirement (no `// indirect`), exact tag (newest from `go list -m -versions` on 2026-09-22). No pseudo-version, no `replace`.
- `internal/mount` imports only the standard library. Only `internal/mount/gofuse` imports go-fuse.
- The `Catalog` method signatures use only predeclared types. That is required so `internal/projection` (S2-02) can satisfy `mount.Catalog` structurally without importing `mount`: Go matches method sets by identical types, so a named type such as `mount.Entry` in a signature would force the import. The exact signatures are the cross-story contract:
  - `Lookup(parent uint64, name string) (ino uint64, isDir bool, found bool)`
  - `ReadDir(dir uint64) (names []string, found bool)` — names in ascending byte order; `found` is false for a missing inode or a non-directory.
  - `Readlink(ino uint64) (target string, found bool)` — target bytes verbatim; `found` is false for a missing inode or a non-symlink.
  - `const RootIno uint64 = 1` — the root directory's inode in every catalog.
- `Adapter`: `Mount(dir string, cat Catalog) error` and `Unmount() error`. The observer reaches S2-04's adapter through its constructor, not this interface.
- `Observer`: `Observe(ev Event)`. `Event{Op Op; Path string}`. `Op` is a `uint8` enum: `OpLookup`, `OpReadDir`, `OpReadlink` (starting at 1; 0 is invalid). `Op.String()` returns `lookup`, `readdir`, `readlink`, and `unknown` for anything else.
- `Kind` is a `uint8` enum: `KindDir = 1`, `KindSymlink = 2` (0 is invalid). `Entry{Ino uint64; Kind Kind; Name string}`.
- `attr.go` constants: `EntryTimeout = time.Second`, `AttrTimeout = time.Second`, `DirPerm = 0o555`, `SymlinkPerm = 0o555` (read-only; no write bits anywhere).

## T1 — Pin go-fuse v2.11.0 in go.mod and go.sum

Run `go get github.com/hanwen/go-fuse/v2@v2.11.0`. Create `internal/mount/gofuse/attr.go` with only `package gofuse`, a package doc comment, and blank imports of `github.com/hanwen/go-fuse/v2/fs` and `github.com/hanwen/go-fuse/v2/fuse`. That way something imports both packages and `go mod tidy` keeps the requirement and writes a complete `go.sum` (the story's first failure scenario). T3 replaces the blank imports with real uses. Then run `go mod tidy` and confirm `go build ./...` passes with `CGO_ENABLED=0`. Tests (RED): the go.mod pin test and the go.sum hash test in `mount_test.go` (package `mount`; a test-only package compiles without `mount.go`), plus the gofuse deps test in `attr_test.go`. RED shims: none. The tests fail by assertion because `go.mod` has no go-fuse line, `go.sum` does not exist, and `go list -deps` on the empty gofuse dir fails. Each test reports that failure through `t.Fatalf` with the command output.

Files: `go.mod`, `go.sum`, `internal/mount/gofuse/attr.go`, `internal/mount/mount_test.go`, `internal/mount/gofuse/attr_test.go`.

## T2 — `internal/mount` interfaces and types

Write `internal/mount/mount.go` with the `Kind`, `Entry`, `RootIno`, `Op`, `Event`, `Catalog`, `Adapter` and `Observer` declarations fixed above, plus `func (o Op) String() string`. Standard library imports only (in practice none). Each interface has 1 to 3 methods. Add to `mount_test.go`: the deps test, `Op.String` table, `Kind` constants test, and test fakes (`fakeCatalog`, `fakeAdapter`, `recordingObserver`) with compile-time assertions `var _ Catalog = fakeCatalog{}` and the equivalents. RED shims (`mount.go`): all type and interface declarations as above, but `KindDir`, `KindSymlink`, `OpLookup`, `OpReadDir` and `OpReadlink` all set to `0`, and `String()` returning `""`, so the constants and `String` tests fail by assertion. The deps test and the compile-time assertions pass on RED. They are legitimate PASS-ON-RED guards (M-003) and must not be weakened.

Files: `internal/mount/mount.go`, `internal/mount/mount_test.go`.

## T3 — `gofuse/attr.go` translation, EROFS errno and full matrix

Replace T1's blank imports in `internal/mount/gofuse/attr.go` with the pure translation:

- `StableAttr(e mount.Entry) fs.StableAttr`: `Ino` = `e.Ino`; `Mode` = `syscall.S_IFDIR` or `syscall.S_IFLNK`; 0 for an unknown kind.
- `Attr(e mount.Entry, owner fuse.Owner) fuse.Attr`: `Ino`; `Mode` = type bits OR `DirPerm`/`SymlinkPerm`; `Owner` = `owner`; `Nlink` 1. An unknown kind gives `Mode == 0` (no permission bits).
- `EntryOut(e, owner) fuse.EntryOut`: `Attr` from `Attr`, with `SetEntryTimeout(EntryTimeout)` and `SetAttrTimeout(AttrTimeout)`.
- `AttrOut(e, owner) fuse.AttrOut`: `Attr` from `Attr`, with `SetTimeout(AttrTimeout)`.
- `DaemonOwner() fuse.Owner`: the process's `os.Getuid()`/`os.Getgid()`.
- `ReadOnlyErrno() syscall.Errno`: returns `syscall.EROFS`, used by S2-04 for every mutation.

Constructors return values; nothing mutates caller state. Add the table-driven tests to `attr_test.go`. RED shims (`attr.go`): the same signatures returning zero values, `ReadOnlyErrno` returning `0`, and all four constants set to `0`, so every test fails by assertion. `go mod tidy -diff` must print nothing; if tidy wants a `go.sum` change, commit it here (the file is S2-01's). This is the story's LAST task: GREEN runs the full standards matrix (`GATE_RUN_MATRIX=1`), not a package-scoped subset.

Files: `internal/mount/gofuse/attr.go`, `internal/mount/gofuse/attr_test.go`, `go.sum` (only if `go mod tidy` changes it).
