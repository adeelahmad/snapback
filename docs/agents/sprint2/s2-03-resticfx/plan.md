---
type: plan
story: S2-03
scope: "tests only"
---

# S2-03 test plan (tests only)

Contracts under test are fixed in `tasks.md`. Package `github.com/adeelahmad/snapback/internal/compat/resticfx`. All default-build tests run without restic, FUSE or network and fail by assertion against scaffolded stubs. Every filesystem fixture is under `t.TempDir()`; no test reads or writes the real home directory. Negative checks first assert the positive set is non-empty (M-002).

## T1 — version parsing and repository guard

- [ ] `internal/compat/resticfx/version_test.go::TestParseResticVersion` — table: `restic 0.19.0 compiled with go1.26.4 on darwin/arm64` -> `0.19.0`; `restic 0.18.1 compiled with go1.24.0 on linux/amd64` -> `0.18.1`; `""`, `garbage`, `restic`, `restic x.y.z` -> error; assert exact version string or non-nil error per row.
- [ ] `internal/compat/resticfx/version_test.go::TestCheckPinnedVersion` — `0.19.0` output -> nil; `0.18.1` output -> error whose message contains both `0.18.1` and `0.19.0`; garbage -> error.
- [ ] `internal/compat/resticfx/guard_test.go::TestGuardRepo` — `Guard{TempRoot: t.TempDir(), Home: <second t.TempDir() acting as home>}`; allowed: `<TempRoot>/repo`, `<TempRoot>/a/b`, `rclone:gdrive:snapback-stage1`; refused: `<Home>/x`, `rclone:gdrive:`, `rclone:gdrive:other`, `rclone:other:snapback-stage1`, `/`, `<TempRoot>` itself, `relative/repo`, `<TempRoot>/../escape`, `sftp:host:/r`; assert nil vs non-nil per row and that the allowed set is non-empty.

## T2 — argument builders and unmount command

- [ ] `internal/compat/resticfx/args_test.go::TestInitArgs` — `InitArgs("/tmp/x/repo","/tmp/x/pw")`; assert exact slice `[-r /tmp/x/repo --password-file /tmp/x/pw init]`.
- [ ] `internal/compat/resticfx/args_test.go::TestBackupArgs` — assert exact slice `[-r R --password-file P backup --quiet D]`, including a `D` containing a space and `;` kept as one element (shell metacharacters are data).
- [ ] `internal/compat/resticfx/args_test.go::TestSnapshotsAndLsArgs` — `SnapshotsArgs` ends `snapshots --json`; `LsArgs(R,P,<64-hex>)` ends `ls --json <64-hex>`; assert exact slices.
- [ ] `internal/compat/resticfx/args_test.go::TestMountArgs` — assert exact slice `[-r R --password-file P mount --path-template ids/%I M]`; assert `--path-template` and `ids/%I` are adjacent separate elements and `M` is last.
- [ ] `internal/compat/resticfx/args_test.go::TestArgsNeverContainPassword` — write a real password file with `NewPasswordFile(t.TempDir())`, read its value; for every builder assert the password-file path IS present and the password value is in no element; also assert two calls return slices with distinct backing arrays (mutating one does not change the other).
- [ ] `internal/compat/resticfx/args_test.go::TestUnmountCommand` — darwin -> `umount [M]`; linux -> `fusermount3 [-u M]`; windows -> error.

## T3 — snapshot JSON and ids/ observation

- [ ] `internal/compat/resticfx/snapshots_test.go::TestParseSnapshotsFullIDs` — two-entry JSON fixture shaped like `restic snapshots --json` (fields `id`, `short_id`, `time`, `hostname`, `paths`); assert 2 snapshots, each `ID` is the 64-hex `id` (not `short_id`), time and paths decoded.
- [ ] `internal/compat/resticfx/snapshots_test.go::TestParseSnapshotsRejectsBadIDs` — table: `id` 8-hex, `id` 63-hex, `id` uppercase hex, `id` missing (only `short_id`), `id` with a non-hex char; assert error per row.
- [ ] `internal/compat/resticfx/snapshots_test.go::TestParseSnapshotsEmptyAndGarbage` — `[]` -> empty slice, nil error; `null` -> empty, nil; `not json` -> error.
- [ ] `internal/compat/resticfx/snapshots_test.go::TestCheckObservedIDs` — table with `full` = 64-hex: `[full]` pass; `[full, other64]` pass; `[full[:8]]` fail; `[full, full[:8]]` fail (prefix entry present); `["somedir"]` fail; `[]` fail; invalid `fullID` fail.

## T4 — generated tree and password file

- [ ] `internal/compat/resticfx/tree_test.go::TestWriteTreeMetadata` — specs: `a.txt` 0 bytes 0644, `dir with space/b.bin` 4096 bytes 0600, `nested/deep/c.txt` 17 bytes 0755, each with a distinct fixed `ModTime` (second precision, UTC); assert returned metadata and a fresh `os.Lstat` match size, `Mode().Perm()` and `ModTime().Equal` for every spec, and len == 3.
- [ ] `internal/compat/resticfx/tree_test.go::TestWriteTreeRejectsEscapingPaths` — specs `/etc/x`, `../x`, `a/../../x`; assert error and that nothing was created outside `root` (parent dir listing unchanged).
- [ ] `internal/compat/resticfx/password_test.go::TestNewPasswordFile` — assert path is inside the given dir, mode is exactly `0600`, content is 64 hex chars, and two calls produce different contents.

## T5 — runner and fixture lifecycle

- [ ] `internal/compat/resticfx/runner_test.go::TestExecRunnerPassesArgvVerbatim` — run the test binary itself as a helper (`os.Args[0]`, `-test.run=TestHelperProcess`, `GO_WANT_HELPER_PROCESS=1`) with args including `a b`, `$(echo pwned)`, `;ls`; assert stdout echoes each arg as one line exactly (no shell expansion).
- [ ] `internal/compat/resticfx/runner_test.go::TestExecRunnerErrorIncludesStderr` — helper exits 3 writing `boom` to stderr; assert non-nil error containing `boom` and exit status.
- [ ] `internal/compat/resticfx/fixture_test.go::TestNewFixtureRefusesUnguardedRepo` — `NewFixture` with repo `<fake home>/repo` and `rclone:gdrive:other`; assert error and that the fake runner recorded zero calls.
- [ ] `internal/compat/resticfx/fixture_test.go::TestFixtureInitBackupSnapshots` — fake runner scripted to return a snapshots JSON with one 64-hex ID; call Init, Backup, Snapshots; assert recorded calls are `restic` with exactly `InitArgs`, `BackupArgs`, `SnapshotsArgs`, and the returned ID is the full 64-hex value.
- [ ] `internal/compat/resticfx/fixture_test.go::TestFixtureInitRefusesExistingRepo` — pre-create a non-empty dir under `t.TempDir()`; assert Init errors and the runner recorded no `init` call.
- [ ] `internal/compat/resticfx/fixture_test.go::TestFixtureDestroyOnlyOwnRepo` — a fixture that did not run Init: Destroy leaves the dir intact; after a successful (faked) Init that created the dir: Destroy removes it; assert both outcomes and that a sibling file in `TempRoot` survives.
- [ ] `internal/compat/resticfx/fixture_test.go::TestToolVersions` — fake runner returns restic and rclone version text; assert parsed `0.19.0` and rclone version string; with rclone reported not found, assert rclone `""` and nil error.

## T6 — evidence JSON and prerequisite probe

- [ ] `internal/compat/resticfx/evidence_test.go::TestEvidenceRoundTrip` — populated `Evidence` -> `WriteEvidence(t.TempDir(), ev)` -> `ReadEvidence`; assert deep-equal, file ends with `\n`, and JSON keys include `restic_version`, `snapshot_id`, `observed_dir`, `path_template`, `result`.
- [ ] `internal/compat/resticfx/evidence_test.go::TestWriteEvidenceFileName` — `GOOS` darwin -> `pathtemplate-darwin.json`, linux -> `pathtemplate-linux.json`; assert returned path and file existence; empty dir argument -> error.
- [ ] `internal/compat/resticfx/evidence_test.go::TestEvidenceOmitsPassword` — generate a password file, build evidence from a fixture run with the fake runner; assert the written file is non-empty and contains neither the password value nor the password-file path.
- [ ] `internal/compat/resticfx/evidence_test.go::TestEvidenceDirFromEnv` — fake getenv returning a temp dir -> that dir; returning "" -> "".
- [ ] `internal/compat/resticfx/prereq_test.go::TestMissingPrerequisite` — table over fake `Probe`: all present (linux and darwin) -> ""; `SNAPBACK_FUSE_TESTS` unset -> message containing `SNAPBACK_FUSE_TESTS`; restic not on PATH -> contains `restic`; linux without `/dev/fuse` -> contains `/dev/fuse`; linux without `fusermount3` -> contains `fusermount3`; darwin without macFUSE -> contains `macfuse`; version `0.18.1` -> contains `0.18.1` and `0.19.0`; assert every non-empty message is non-generic (names the item).

## T7 — mount supervisor and path-template integration

- [ ] `internal/compat/resticfx/mount_test.go::TestMountWaitReadyAndStop` — start the helper process (long-running, creates `<mnt>/ids` then blocks until interrupted) via `StartMount`; assert `WaitReady` returns nil, `Stop` returns nil, the process has exited, and the injected unmount func was invoked at most once.
- [ ] `internal/compat/resticfx/mount_test.go::TestMountWaitReadyTimeout` — helper that never creates `ids`; `WaitReady` with a 200 ms ctx returns an error wrapping `context.DeadlineExceeded`; `Stop` still reaps the process.
- [ ] `internal/compat/resticfx/mount_test.go::TestMountStopIdempotent` — helper that exits immediately; call `Stop` twice; assert both return nil (or the same recorded exit error) and no panic.
- [ ] `internal/compat/resticfx/pathtemplate_integration_test.go::TestPathTemplateIntegration` — `//go:build integration`; skip with `MissingPrerequisite` message if any prerequisite is absent; otherwise run the full flow from `tasks.md` T7 against real restic 0.19.0; assert `<mnt>/ids/<64-hex fullID>` exists, `CheckObservedIDs` passes, the listing under the snapshot root contains the three generated names, the mount is gone after cleanup, and when `SNAPBACK_EVIDENCE_DIR` is set the evidence file exists with `result=pass`, `restic_version=0.19.0`, and `snapshot_id` equal to `observed_dir`.

Test count: 32 (31 default-build, 1 integration-tagged). `TestHelperProcess` in `runner_test.go` is a helper entry point, not a counted test.
