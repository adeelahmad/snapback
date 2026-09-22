---
type: tasks
story: S2-03
---

# S2-03 tasks — Disposable Restic fixture and `--path-template ids/%I` verification

Story intent and owned files: `docs/agents/sprint2/stories.md` (S2-03). Every task touches only S2-03's owned set: `internal/compat/resticfx/**`, `docs/reports/stage1/pathtemplate-darwin.json`, `docs/reports/stage1/pathtemplate-linux.json`. Tests per task are in `plan.md`; gate commands in `validate.md`. Rules: `docs/agents/sprint2/standards.md` (gate matrix and "Sprint 2 / Stage 1 integration-test gating"). Spec: SPEC.md §5 (exec boundary, full IDs, path template, no timeout on mount), §6 (full snapshot identity), §12 (no secrets in args/logs), §20 (disposable fixtures, never a real home dir).

Decisions fixed here so workers do not guess:

- Package `resticfx` at `github.com/adeelahmad/snapback/internal/compat/resticfx`. Standard library only (no go-fuse, no `restic/internal`).
- `const PinnedResticVersion = "0.19.0"`. `restic version` output shape (verified locally during planning): `restic 0.19.0 compiled with go1.26.4 on darwin/arm64`.
- Every subprocess goes through one `Runner` interface: `Run(ctx context.Context, name string, args []string) (stdout []byte, err error)`. The only production implementation, `ExecRunner`, uses `exec.CommandContext(ctx, name, args...)`. No `sh -c`, no `exec.Command` without context, no string-joined commands. Unit tests use a recording fake runner.
- Passwords: `NewPasswordFile(dir)` writes 32 random bytes (hex, `crypto/rand`) to a `0600` file under a `t.TempDir()`-style dir and returns only the path. Every restic argv carries `--password-file <path>`; no argv, env var, log line or evidence field ever carries the password value. No password or key is committed.
- Guard: `GuardRepo(repo string, g Guard) error` with `Guard{TempRoot, Home string}` injected (production default: `os.TempDir()`, `os.UserHomeDir()`). Allowed: (a) an absolute path strictly under `TempRoot` (compared after `filepath.Clean` and, where the path exists, `filepath.EvalSymlinks` on both sides — macOS `/var` -> `/private/var`); (b) exactly `rclone:gdrive:snapback-stage1`. Refused with a descriptive error: relative paths, `/`, `TempRoot` itself, anything under `Home` outside `TempRoot`, any other `rclone:` spec (including `rclone:gdrive:` and `rclone:gdrive:other`), any other backend prefix.
- Full IDs: a snapshot ID is valid only if it matches `^[0-9a-f]{64}$`. `ParseSnapshots` rejects any entry whose `id` is missing, short or non-hex; `short_id` is ignored.
- Path template literal: `--path-template` followed by `ids/%I` as two separate argv elements.
- Evidence contract (`SNAPBACK_EVIDENCE_DIR`, also used by S2-04, S2-06, S2-07 and set by the S2-05 CI job): when the env var is non-empty, write `<dir>/pathtemplate-<runtime.GOOS>.json` (0644, pretty-printed, trailing newline). Fields: `goos`, `goarch`, `restic_version`, `rclone_version` (omitempty), `path_template` (`ids/%I`), `snapshot_id` (64-hex), `observed_dir` (name seen under `ids/`), `ids_entries` (all names seen), `result` (`pass` or `fail`), `reason` (omitempty), `generated_at` (RFC 3339 UTC). The integration test writes the file whether it passes or fails, then fails the test on `fail`.
- Mount supervision: the mount process gets a cancellable context with no deadline (SPEC §5). Readiness means `<mnt>/ids` exists and is a directory, polled until a caller-provided finite wait. Stop sends `os.Interrupt`, waits a bounded grace period, then runs the platform unmount command and kills the process. Stop is idempotent and is registered with `t.Cleanup` before readiness is awaited, so a failed test never leaks a mount.
- Unmount command: darwin `umount <mnt>`; linux `fusermount3 -u <mnt>`.
- Prerequisites: `MissingPrerequisite(p Probe) string` returns "" when all present, else a message naming exactly one missing item, checked in this order: `SNAPBACK_FUSE_TESTS=1`, `restic` on PATH, FUSE (linux: `/dev/fuse` exists and `fusermount3` on PATH; darwin: `/Library/Filesystems/macfuse.fs` exists), restic version equals `0.19.0`. The integration test calls `t.Skip(msg)` with that message.

## T1 — Pinned version parsing and repository guard

`ParseResticVersion(out string) (string, error)` extracts `X.Y.Z` from the first line of `restic version` output and errors on anything else. `CheckPinnedVersion(out string) error` errors (naming found and wanted versions) unless it is `0.19.0`. `Guard` type and `GuardRepo` as specified above. Pure functions; no subprocess.

Files: `internal/compat/resticfx/version.go`, `internal/compat/resticfx/guard.go`, `internal/compat/resticfx/version_test.go`, `internal/compat/resticfx/guard_test.go`.

## T2 — Argument builders and unmount command

Pure builders returning fresh `[]string` (no shared backing arrays): `VersionArgs()`, `InitArgs(repo, pwFile)`, `BackupArgs(repo, pwFile, dir)` (includes `--quiet`), `SnapshotsArgs(repo, pwFile)` (includes `--json`), `LsArgs(repo, pwFile, snapshotID)` (includes `--json`), `MountArgs(repo, pwFile, mnt)` (includes `--path-template` then `ids/%I`, then `mnt` last; no other mount flags). Each begins with `-r <repo> --password-file <pwFile>` then the subcommand. `UnmountCommand(goos, mnt string) (name string, args []string, err error)` for darwin/linux, error for any other GOOS.

Files: `internal/compat/resticfx/args.go`, `internal/compat/resticfx/args_test.go`.

## T3 — Snapshot JSON parsing and `ids/` observation check

`type Snapshot struct { ID string; Time time.Time; Hostname string; Paths []string }`. `ParseSnapshots(data []byte) ([]Snapshot, error)` decodes `restic snapshots --json` output and enforces the 64-hex rule on every entry. `CheckObservedIDs(entries []string, fullID string) error` passes only when `entries` contains `fullID` exactly and no entry is a strict prefix of any full ID (a short ID or truncated name is a failure, not a pass); errors on empty `entries` and on an invalid `fullID`.

Files: `internal/compat/resticfx/snapshots.go`, `internal/compat/resticfx/snapshots_test.go`.

## T4 — Generated tree writer and password file

`type FileSpec struct { Path string; Size int; Mode os.FileMode; ModTime time.Time }` and `type FileMeta` (same fields, as written). `WriteTree(root string, specs []FileSpec) ([]FileMeta, error)` creates parent dirs, writes deterministic content of `Size` bytes, applies `Mode` (chmod after write, not umask-dependent) and `ModTime` (`os.Chtimes`), and returns the metadata read back via `os.Lstat`. Rejects absolute paths and any `..` escape of `root`. `NewPasswordFile(dir string) (string, error)` as specified above.

Files: `internal/compat/resticfx/tree.go`, `internal/compat/resticfx/password.go`, `internal/compat/resticfx/tree_test.go`, `internal/compat/resticfx/password_test.go`.

## T5 — Runner and fixture lifecycle

`type Runner interface` and `ExecRunner` (as specified; stderr captured into the returned error, truncated to 4 KiB). `type Fixture struct` created by `NewFixture(r Runner, repo, pwFile string, g Guard) (*Fixture, error)` which calls `GuardRepo` first. Methods: `Init(ctx)` (refuses a local repo path that already exists and is non-empty; records that this fixture created it), `Backup(ctx, dir)`, `Snapshots(ctx) ([]Snapshot, error)` (via `ParseSnapshots`), `Destroy()` (removes a local repo only if this fixture created it and it still passes the guard; never touches rclone remotes, which S2-08 deletes under human control), and `ToolVersions(ctx) (restic, rclone string, err error)` (rclone is "" when `rclone` is not on PATH, never an error). Finite calls take the caller's ctx; the helper never adds a timeout to a mount. The `TestHelperProcess` fake executable (the standard `os.Args[0] -test.run=TestHelperProcess` pattern, gated on `GO_WANT_HELPER_PROCESS=1`) lives in `runner_test.go` and is reused by T7's `mount_test.go`.

Files: `internal/compat/resticfx/runner.go`, `internal/compat/resticfx/fixture.go`, `internal/compat/resticfx/runner_test.go`, `internal/compat/resticfx/fixture_test.go`.

## T6 — Evidence JSON and prerequisite probe

`type Evidence struct` with the JSON fields above. `EvidenceDir(getenv func(string) string) string` reads `SNAPBACK_EVIDENCE_DIR`. `WriteEvidence(dir string, ev Evidence) (path string, err error)` writes `pathtemplate-<ev.GOOS>.json`. `ReadEvidence(path string) (Evidence, error)`. `type Probe struct { Getenv func(string) string; LookPath func(string) (string, error); Stat func(string) (os.FileInfo, error); GOOS string; ResticVersionOut string }` and `MissingPrerequisite(p Probe) string` as specified.

Files: `internal/compat/resticfx/evidence.go`, `internal/compat/resticfx/prereq.go`, `internal/compat/resticfx/evidence_test.go`, `internal/compat/resticfx/prereq_test.go`.

## T7 — Mount supervisor, path-template integration test, full matrix (last task)

`type Mount struct` returned by `StartMount(r MountStarter, name string, args []string, mnt string) (*Mount, error)` — starts via `exec.CommandContext` with a cancel-only context. `(*Mount).WaitReady(ctx) error` polls for `<mnt>/ids`; `(*Mount).Stop() error` as specified, idempotent. Integration test file tagged `//go:build integration` runs the flow: probe prerequisites (skip naming the missing one), `t.TempDir()` repo/source/mount dirs, password file, `WriteTree` (3 files incl. a space and a nested path), init, backup, snapshots, start `restic mount --path-template ids/%I`, register `Stop` in `t.Cleanup`, wait ready, read `<mnt>/ids`, `CheckObservedIDs`, list `<mnt>/ids/<fullID>/<source path>` and compare names to the generated tree, write evidence when `SNAPBACK_EVIDENCE_DIR` is set, destroy the repo. Then run the full standards matrix (M-005: only this last GREEN runs it) and produce `docs/reports/stage1/pathtemplate-darwin.json` locally; the Linux file is copied by the orchestrator from the S2-05 CI artifact.

Files: `internal/compat/resticfx/mount.go`, `internal/compat/resticfx/mount_test.go`, `internal/compat/resticfx/pathtemplate_integration_test.go`, `docs/reports/stage1/pathtemplate-darwin.json` (generated by the test run, not hand-written).

## Ordering

T1, T2, T3, T4, T6 are independent (separate files, pure logic). T5 depends on T1, T2, T3. T7 depends on all of T1-T6. Intermediate GREENs run the package-scoped matrix with `GATE_RUN_MATRIX=0` (M-005); T7's GREEN runs the full matrix.
