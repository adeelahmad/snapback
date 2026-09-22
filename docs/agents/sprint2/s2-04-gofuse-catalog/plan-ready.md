---
type: plan-ready
story: S2-04
from_red_at: 2026-09-22T02:50:00Z
---

# S2-04 plan-ready (RED verified by orchestrator at chain @ 3a027b8; every box currently FAILs by assertion unless noted)


# S2-04 plan — tests only

Package `internal/mount/gofuse` (`package gofuse`). Default-build tests (T1-T3) build their catalogs with `internal/projection` and attach the node tree in-process with `fs.NewNodeFS`; they never mount. Run: `go test -race ./internal/mount/gofuse/...`. The T4 test is tagged `integration`. Run it with `SNAPBACK_FUSE_TESTS=1 go test -race -tags=integration -run TestCatalogMountLifecycle -v ./internal/mount/gofuse/`.

Shared fixture (test helper in `fs_test.go`): a projection with root entries `docs/` (directory containing `readme-link` -> `../target`), `rel` -> `a/b`, `up` -> `../outside` and `abs` -> `/var/tmp/x`.

Total: 21 tests (20 default build, 1 integration).

## T1 — Read-path nodes

- [x] `internal/mount/gofuse/fs_test.go::TestLookupReturnsCatalogChild` — input: fixture root; action: `Lookup(ctx, "docs", &out)` and `Lookup(ctx, "rel", &out)`; assertion: errno 0, non-nil inodes, `docs` has `S_IFDIR` and `rel` has `S_IFLNK` in `StableAttr().Mode`.
- [x] `internal/mount/gofuse/fs_test.go::TestLookupMissingReturnsENOENT` — input: fixture root; action: `Lookup(ctx, "nope", &out)`; assertion: errno is `syscall.ENOENT` and the returned inode is nil.
- [x] `internal/mount/gofuse/fs_test.go::TestLookupInodeStableAcrossCalls` — input: fixture root; action: look up `docs` twice; assertion: both `StableAttr().Ino` values are non-zero, equal to each other and equal to the catalog's inode for `docs`.
- [x] `internal/mount/gofuse/fs_test.go::TestReaddirListsCatalogEntriesInOrder` — input: fixture root; action: `Readdir(ctx)` and drain the stream; assertion: the list is non-empty, names equal the catalog's ordered listing exactly (`abs`, `docs`, `rel`, `up`), and each entry's mode type bits match its kind.
- [x] `internal/mount/gofuse/fs_test.go::TestReadlinkReturnsExactTargetBytes` — input: table of `rel` -> `a/b`, `up` -> `../outside`, `abs` -> `/var/tmp/x`, `docs/readme-link` -> `../target`; action: look up each symlink and call `Readlink(ctx)`; assertion: errno 0 and bytes equal the table target byte-for-byte (no cleaning or resolution).
- [x] `internal/mount/gofuse/fs_test.go::TestGetattrMatchesAttrTranslation` — input: table of the `docs` directory and the `rel` symlink; action: `Getattr(ctx, nil, &out)` on each node; assertion: `out.Attr` mode, uid, gid and `out` attr timeout equal what `attr.go`'s translation returns for the same `mount.Entry`, and the mode has no write bits.
- [x] `internal/mount/gofuse/fs_test.go::TestLookupSetsEntryAndAttrTimeouts` — input: fixture root; action: `Lookup(ctx, "docs", &out)`; assertion: `out.EntryTimeout()` and `out.AttrTimeout()` equal `attr.go`'s named timeout constants and are non-zero.
- [x] `internal/mount/gofuse/fs_test.go::TestStatfsReportsNoFreeSpace` — input: fixture root; action: `Statfs(ctx, &out)`; assertion: errno 0 and `Bfree`, `Bavail` and `Ffree` are 0.
- [x] `internal/mount/gofuse/fs_test.go::TestObserverReceivesOneEventPerOp` — input: table of ops lookup `docs`, readdir root, readlink `rel`, each with a fresh recording Observer; action: perform the op once; assertion: exactly one event recorded, its op kind matches the op and its path matches the node path (`docs`, root, `rel`).

## T2 — Mutations return EROFS

- [x] `internal/mount/gofuse/fs_test.go::TestNodesImplementMutationInterfaces` — input: a directory node and a symlink node from the fixture; action: runtime type assertions; assertion: the directory node satisfies `fs.NodeMkdirer`, `fs.NodeCreater`, `fs.NodeUnlinker`, `fs.NodeRmdirer`, `fs.NodeRenamer`, `fs.NodeSymlinker`, `fs.NodeLinker`, `fs.NodeSetattrer` and `fs.NodeWriter`, and the symlink node satisfies `fs.NodeSetattrer` (a missing one would make go-fuse answer ENOSYS/ENOTSUP).
- [x] `internal/mount/gofuse/fs_test.go::TestDirMutationsReturnEROFS` — input: table of mkdir, create, unlink, rmdir, rename, symlink, link and setattr invoked on the fixture root; action: call each method; assertion: errno equals `syscall.EROFS` (not EPERM, ENOSYS or ENOTSUP), no inode is returned, and a following `Readdir` equals the original listing.
- [x] `internal/mount/gofuse/fs_test.go::TestSymlinkSetattrReturnsEROFS` — input: the `rel` symlink node; action: `Setattr(ctx, nil, &in, &out)` with a mode change; assertion: errno equals `syscall.EROFS` and a following `Readlink` still returns `a/b`.
- [x] `internal/mount/gofuse/fs_test.go::TestWriteReturnsEROFS` — input: the fixture root node; action: `Write(ctx, nil, []byte("x"), 0)`; assertion: errno equals `syscall.EROFS` and the written count is 0.

## T3 — Adapter

- [x] `internal/mount/gofuse/adapter_test.go::TestAdapterSatisfiesMountAdapter` — input: `*Adapter` from its constructor; action: assign to a `mount.Adapter` variable and type-assert at runtime; assertion: the assertion succeeds and the value is non-nil.
- [x] `internal/mount/gofuse/adapter_test.go::TestProjectionGenerationSatisfiesMountCatalog` — input: a generation built by `projection` from the fixture spec; action: pass it to a function taking `mount.Catalog` and call the Catalog lookup for `docs` under the root; assertion: the call compiles and returns the `docs` directory entry.
- [x] `internal/mount/gofuse/adapter_test.go::TestMountOptionsReadOnlyAndPrivate` — input: none; action: call the mount-options function; assertion: `Options` contains `ro`, `AllowOther` is false, and `FsName` is non-empty.
- [x] `internal/mount/gofuse/adapter_test.go::TestMountRejectsNonEmptyDir` — input: a `t.TempDir()` containing one file; action: `Mount` the fixture catalog there; assertion: a non-nil error mentioning the path, the file is still present and unchanged, and no mount was attempted (Unmount afterwards reports nothing mounted).
- [x] `internal/mount/gofuse/adapter_test.go::TestMountRejectsMissingDir` — input: a path under `t.TempDir()` that does not exist; action: `Mount`; assertion: a non-nil error that wraps `fs.ErrNotExist` (`errors.Is`).
- [x] `internal/mount/gofuse/adapter_test.go::TestMountRejectsNonDirectory` — input: a regular file under `t.TempDir()`; action: `Mount` with that file as the mountpoint; assertion: a non-nil error mentioning the path, and the file is unchanged.
- [x] `internal/mount/gofuse/adapter_test.go::TestUnmountBeforeMountReturnsError` — input: a fresh `*Adapter`; action: `Unmount`; assertion: a non-nil error, no panic.

## T4 — Real-mount integration test

- [x] `internal/mount/gofuse/mount_integration_test.go::TestCatalogMountLifecycle` — input: `//go:build integration`, `SNAPBACK_FUSE_TESTS=1`, FUSE device probe (else `t.Skip` naming `SNAPBACK_FUSE_TESTS`, `/dev/fuse` or macFUSE), the fixture projection, mountpoint `filepath.Join(t.TempDir(), "mnt")`; action: Mount via `*Adapter` with a `t.Cleanup` unmount, `os.ReadDir` root and `docs`, `os.Readlink` every symlink, `os.Mkdir` and `os.WriteFile` inside the mount, `os.Lstat` `docs` twice, Unmount; assertion: expected names present first (M-002), exact readlink targets, both mutations fail with `errors.Is(err, syscall.EROFS)`, equal inodes on both stats, Unmount returns nil, the mountpoint is an empty directory on its parent's device, and when `SNAPBACK_EVIDENCE_DIR` is set `catalog-<GOOS>.json` exists there and decodes with `platform`, `go_fuse_version`, `operations` (non-empty) and `result` equal to `pass`.
