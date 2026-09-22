---
type: tasks
story: S2-04
---

# S2-04 tasks — Tiny go-fuse history catalog behind the adapter

Owned set (stories.md S2-04): `internal/mount/gofuse/fs.go`, `internal/mount/gofuse/fs_test.go`, `internal/mount/gofuse/adapter.go`, `internal/mount/gofuse/adapter_test.go`, `internal/mount/gofuse/mount_integration_test.go`, `docs/reports/stage1/catalog-darwin.json`, `docs/reports/stage1/catalog-linux.json`. No task touches anything else. `go.mod`/`go.sum`, `internal/mount/mount.go` and `internal/mount/gofuse/attr.go` belong to S2-01; `internal/projection/**` belongs to S2-02. If a new module requirement or an `attr.go` change is needed, stop and report it as an S2-01 fix task (stories.md S2-04 Constraints).

Package: `github.com/adeelahmad/snapback/internal/mount/gofuse`, tests in-package (`package gofuse`) so node types can be exercised directly. Test catalogs are built with `internal/projection` (S2-02), never a hand-rolled fake catalog, so the cross-story `mount.Catalog` contract is exercised for real. Where this file names an S2-01 or S2-02 symbol (`mount.Catalog`, `mount.Adapter`, `mount.Observer`, `mount.Entry`, the `attr.go` translation function and EROFS errno, the projection build function and generation type), the implementing workers use the exact identifiers those stories merged; the names here describe the contract, not new symbols.

Split of evidence (user brief):

- **Pure unit tests (default build, T1-T3):** node/attr mapping, Observer events, readlink targets, EROFS on every mutation, mount options and mountpoint preflight. They use go-fuse's in-process API (`fs.NewNodeFS` on the root node to attach the inode tree, then calling node methods directly). No kernel mount, no `/dev/fuse`, no macFUSE. They carry the package's default-build coverage (>= 80%).
- **Real-mount integration test (T4):** `//go:build integration`, gated on `SNAPBACK_FUSE_TESTS=1` plus a device probe, with a `t.Skip` message naming the missing prerequisite.

Order: T1 -> (T2 || T3) -> T4. T2 and T3 own disjoint files (`fs.go`/`fs_test.go` vs `adapter.go`/`adapter_test.go`) and both depend only on T1's root-node constructor, so they may run in parallel. T4 needs both. Per M-005, T1-T3 GREENs run with `GATE_RUN_MATRIX=0` plus the package-scoped matrix; T4 is the last task and runs the full matrix.

Cross-story names (binding): `mount.Catalog` (S2-01) uses builtin-type signatures only — `Lookup(parent uint64, name string) (ino uint64, isDir bool, found bool)`, `ReadDir(dir uint64) (names []string, found bool)`, `Readlink(ino uint64) (target string, found bool)` — with `mount.RootIno = 1` as the root inode. Catalogs are built with `projection.Build(Spec) (*Generation, error)` (S2-02); `*projection.Generation` satisfies `mount.Catalog` structurally. Timestamp aliases use the format `2006-01-02_1504Z` (SPEC §3).

## T1 — Read-path nodes: lookup, getattr, readdir, readlink, statfs, Observer

- **Files:** `internal/mount/gofuse/fs.go` (create), `internal/mount/gofuse/fs_test.go` (create).
- **Does:** Defines a directory node and a symlink node (each embedding `fs.Inode`) over a `mount.Catalog`, plus an unexported root constructor taking the catalog and a `mount.Observer`. Directory nodes implement `NodeLookuper`, `NodeReaddirer`, `NodeGetattrer` and `NodeStatfser`; symlink nodes implement `NodeReadlinker` and `NodeGetattrer`. Lookup creates child inodes with `StableAttr.Ino` equal to the catalog's stable inode and the correct type bits, fills `fuse.EntryOut` through `attr.go` (mode, uid/gid, entry and attr timeouts as `attr.go`'s named constants) and returns `ENOENT` on a miss. Readdir returns the catalog's ordered listing. Readlink returns the catalog's target bytes unchanged (never cleaned or resolved). Getattr fills `fuse.AttrOut` from `attr.go`. Statfs returns success with zero free and available blocks and inodes (read-only catalog). Lookup, readdir and readlink each fire exactly one Observer event carrying the op kind and the node path. There is no regular-file node and no read path for file content.
- **Tests:** plan.md T1 block (9 tests).

## T2 — Mutations return EROFS

- **Files:** `internal/mount/gofuse/fs.go` (edit), `internal/mount/gofuse/fs_test.go` (edit).
- **Does:** The directory node implements `NodeMkdirer`, `NodeCreater`, `NodeUnlinker`, `NodeRmdirer`, `NodeRenamer`, `NodeSymlinker`, `NodeLinker`, `NodeSetattrer` and `NodeWriter`; the symlink node implements `NodeSetattrer`. Every one returns `attr.go`'s EROFS errno and creates no inode. Implementing the interfaces explicitly matters: a node that lacks one makes go-fuse answer `ENOTSUP`/`ENOSYS`, which is the stories.md failure scenario. Mutations fire no Observer event (the Observer counts catalog reads only).
- **Tests:** plan.md T2 block (4 tests).

## T3 — Adapter: mount options, mountpoint preflight, contract checks

- **Files:** `internal/mount/gofuse/adapter.go` (create), `internal/mount/gofuse/adapter_test.go` (create).
- **Does:** `*Adapter` implements `mount.Adapter`. A pure, unexported mount-options function returns `fuse.MountOptions` with the read-only option set, `AllowOther` false and a fixed `FsName`/`Name`. Mount preflights the target: it must exist, be a directory and be empty, else it returns a wrapped error before any FUSE call (never mount over a live directory, SPEC §1). Mount builds the T1 root and calls `fs.Mount`; the returned server's lifetime is not bound to a timeout context (SPEC §5). Unmount calls the server's `Unmount`, then `Wait`, and returns a wrapped error when nothing is mounted. Compile-time assertions: `var _ mount.Adapter = (*Adapter)(nil)` and the projection generation type assigned to `mount.Catalog`.
- **Tests:** plan.md T3 block (7 tests).

## T4 — Real-mount integration test, evidence file, full matrix

- **Files:** `internal/mount/gofuse/mount_integration_test.go` (create), `docs/reports/stage1/catalog-darwin.json` (produced by the local macOS run; committed by the orchestrator), `docs/reports/stage1/catalog-linux.json` (copied by the orchestrator from the S2-05 `stage1-evidence-linux` artifact).
- **Does:** File is tagged `//go:build integration`. Gating, in order: `SNAPBACK_FUSE_TESTS` != `1` -> `t.Skip("SNAPBACK_FUSE_TESTS not set: FUSE catalog tests skipped")`; on linux, `/dev/fuse` missing -> `t.Skip("fuse3 device /dev/fuse not present")`; on darwin, `/Library/Filesystems/macfuse.fs` missing -> `t.Skip("macFUSE not installed: /Library/Filesystems/macfuse.fs missing")`. The test builds a small projection (at least one subdirectory, one relative symlink, one `..` symlink, one absolute symlink), mounts it through `*Adapter` at a fresh `t.TempDir()` subdirectory, registers a `t.Cleanup` that unmounts even on failure, then: `os.ReadDir` of the root and subdirectory (asserting the expected names exist before anything else, M-002); `os.Readlink` of each symlink (exact targets); `os.Mkdir` and `os.WriteFile` inside the mount fail with `syscall.EROFS`; two `os.Lstat` calls on the same entry report the same inode; explicit Unmount returns nil; the mountpoint is afterwards an empty directory that is no longer a mount (its device ID equals its parent's). When `SNAPBACK_EVIDENCE_DIR` is set, the test writes `catalog-<runtime.GOOS>.json` after the last assertion and reads it back, with `platform` (`GOOS/GOARCH`), `go_fuse_version` (from `debug.ReadBuildInfo`), `operations` (the ordered list exercised), `result` and `timestamp`; a `t.Cleanup` rewrites it with `result: "fail"` and the failed operation when `t.Failed()`, so failure evidence is never lost; the evidence helper lives only in this `_test.go` file. The task finishes by running the full standards matrix (default build) and the integration suite on this Mac, per validate.md Final sign-off. If go-fuse cannot pass on macFUSE, the worker stops and reports; it does not swap in cgofuse (SPEC §5, sprint plan decision 3).
- **Tests:** plan.md T4 block (1 test).
