---
type: output
story: S2-03
---

## S2-03/T1 · attempt 1 · red-worker · 2026-09-22T02:15:09Z
status: ok
### Summary
Wrote the three T1 tests from plan.md in new package `internal/compat/resticfx` (internal test package `resticfx`). Symbols `PinnedResticVersion`, `Guard{TempRoot,Home}`, `ParseResticVersion`, `CheckPinnedVersion`, `GuardRepo` live in the marked shim `zz_agentic_shim_t1.go` (first line `// agentic:shim`) with deliberately wrong bodies (constant `0.0.0-shim`, parser returns `"shim", nil`, checks return nil). Guard test asserts the allowed set is non-empty before the refused checks (M-002); `<TempRoot>/../escape` is built by string concatenation so it is not pre-cleaned. go vet clean; golangci-lint on the package: 0 issues. Commit 8422b92 on branch worktree-agent-a421ac2531fa94394; diff vs 2158ade = 3 files (2 tests + shim).
### Result
| Check | Status | Detail |
|---|---|---|
| `internal/compat/resticfx/version_test.go::TestParseResticVersion` | FAIL | all 6 rows fail by assertion (got "shim"/nil error) |
| `internal/compat/resticfx/version_test.go::TestCheckPinnedVersion` | FAIL | `CheckPinnedVersion(0.18.1) = nil; want error` |
| `internal/compat/resticfx/guard_test.go::TestGuardRepo` | FAIL | 9 refused rows return nil; want error |
| `go vet ./...` | PASS | clean |
| `golangci-lint run ./internal/compat/resticfx/...` | PASS | 0 issues |
### Next
Scaffold `version.go` (`PinnedResticVersion = "0.19.0"`, `ParseResticVersion`, `CheckPinnedVersion`) and `guard.go` (`Guard`, `GuardRepo`), then delete `zz_agentic_shim_t1.go`. Note: the shim carries the package doc comment; the scaffolder should move it to a production file.

## S2-03/T2 · attempt 1 · red-worker · 2026-09-22T02:15:21Z
status: ok
### Summary
I wrote all 6 T2 test bullets from plan.md in `internal/compat/resticfx/args_test.go`. The symbols they need come from one marked shim, `zz_agentic_shim_t2.go` (first line `// agentic:shim`). Its bodies are deliberately wrong: every builder returns the same shared `["shim"]` slice, so the backing-array check fails too, and `UnmountCommand` always errors. Every test compiles and fails by assertion. `go vet ./...` is clean and `golangci-lint run ./internal/compat/resticfx/...` reports 0 issues. Commit a79cb73 changes 2 files only.

One overlap to watch: `TestArgsNeverContainPassword` calls `NewPasswordFile`, which belongs to T4. My shim therefore also defines `NewPasswordFile`. If T4's shim defines it too, merging the two will cause a duplicate-symbol compile error. The scaffolder must keep exactly one `NewPasswordFile` stub when it deletes the shims.

`VersionArgs` has no test bullet and is not shimmed. The password check covers only the builders that take `pwFile`.
### Result
| Check | Status | Detail |
|---|---|---|
| `internal/compat/resticfx/args_test.go::TestInitArgs` | FAIL | InitArgs = ["shim"], want [-r R --password-file P init] |
| `internal/compat/resticfx/args_test.go::TestBackupArgs` | FAIL | exact slice mismatch (incl. dir with space and `;`) |
| `internal/compat/resticfx/args_test.go::TestSnapshotsAndLsArgs` | FAIL | SnapshotsArgs/LsArgs exact slice mismatch |
| `internal/compat/resticfx/args_test.go::TestMountArgs` | FAIL | exact slice mismatch; path-template adjacency and mnt-last asserted |
| `internal/compat/resticfx/args_test.go::TestArgsNeverContainPassword` | FAIL | pw-file path absent; shared backing array detected |
| `internal/compat/resticfx/args_test.go::TestUnmountCommand` | FAIL | darwin/linux return unexpected error |
| `go vet ./...` | PASS | clean |
| `golangci-lint run ./internal/compat/resticfx/...` | PASS | 0 issues |
### Next
Scaffold the named symbols: `InitArgs`, `BackupArgs`, `SnapshotsArgs`, `LsArgs`, `MountArgs`, `VersionArgs` and `UnmountCommand` in `args.go`, then delete `zz_agentic_shim_t2.go`. Keep a single `NewPasswordFile` stub, which T4 owns in `password.go`.

## S2-03/T3 · attempt 1 · red-worker · 2026-09-22T02:15:29Z
status: ok

### Summary
Wrote the four T3 tests from plan.md § T3 in `internal/compat/resticfx/snapshots_test.go` (package `resticfx`), plus the compile shim `zz_agentic_shim_t3.go` (`// agentic:shim`) defining `Snapshot`, `ParseSnapshots` (returns nil, nil) and `CheckObservedIDs` (returns nil). The shim bodies are wrong on purpose so that every test fails by assertion. `go vet ./...` is clean and `golangci-lint run ./internal/compat/resticfx/...` reports 0 issues. The diff against 2158ade is these 2 files only. Commit 158b5b8 is on branch worktree-agent-a707891bb4c6ddc3f.

### Result
| Check | Status | Detail |
|---|---|---|
| `internal/compat/resticfx/snapshots_test.go::TestParseSnapshotsFullIDs` | FAIL | got 0 snapshots, want 2 |
| `internal/compat/resticfx/snapshots_test.go::TestParseSnapshotsRejectsBadIDs` | FAIL | 5/5 rows: want error, got nil |
| `internal/compat/resticfx/snapshots_test.go::TestParseSnapshotsEmptyAndGarbage` | FAIL | ParseSnapshots("not json"): want error, got nil |
| `internal/compat/resticfx/snapshots_test.go::TestCheckObservedIDs` | FAIL | 5 negative rows want error, got nil (the 2 positive rows pass, asserted non-empty per M-002) |
| `go vet ./...` | PASS | clean |
| `golangci-lint run ./internal/compat/resticfx/...` | PASS | 0 issues |

### Next
The scaffolder should stub `Snapshot`, `ParseSnapshots` and `CheckObservedIDs` in `internal/compat/resticfx/snapshots.go` and delete `zz_agentic_shim_t3.go`. The type must not decode `short_id` (validate.md T3 grep row).

## S2-03/T4 · attempt 1 · red-worker · 2026-09-22T02:15:41Z
status: ok

### Summary
Wrote the 3 T4 tests from plan.md at the exact path::fn in package `internal/compat/resticfx`, plus the compile shim `zz_agentic_shim_t4.go` (`// agentic:shim`). The shim declares `FileSpec`, `FileMeta`, `WriteTree` (returns nil, nil) and `NewPasswordFile` (returns "", nil). Those wrong bodies make every test fail by assertion rather than by compile error or panic. The escape test checks that the parent listing is non-empty before asserting that it did not change (M-002). Package-level test helpers use T4-specific names (`treeDirNames`, `newCheckedPasswordFile`, `passwordHexRe`) so they do not collide with the parallel T1-T6 tests. `go vet ./...` is clean and `golangci-lint run ./internal/compat/resticfx/...` reports 0 issues. Commit 0f755b9 changes only these 3 files.

### Result
| Check | Status | Detail |
|---|---|---|
| `internal/compat/resticfx/tree_test.go::TestWriteTreeMetadata` | FAIL | fails by assertion: WriteTree returned 0 metas, want 3 |
| `internal/compat/resticfx/tree_test.go::TestWriteTreeRejectsEscapingPaths` | FAIL | fails by assertion: WriteTree("/etc/x"/"../x"/"a/../../x") returned nil error, want rejection |
| `internal/compat/resticfx/password_test.go::TestNewPasswordFile` | FAIL | fails by assertion: NewPasswordFile returned an empty path |
| `go vet ./...` | PASS | clean |
| `golangci-lint run ./internal/compat/resticfx/...` | PASS | 0 issues |

### Next
Scaffold these symbols in `tree.go` / `password.go` and delete `zz_agentic_shim_t4.go`: `FileSpec`, `FileMeta`, `WriteTree` and `NewPasswordFile`. Merge note: T2 (TestArgsNeverContainPassword) and T6 (TestEvidenceOmitsPassword) also call `NewPasswordFile`. If their shims also declare it, the declarations will clash when the RED chains merge, so keep exactly one declaration.

## S2-03/T5 · attempt 1 · red-worker · 2026-09-22T02:19:06Z
status: ok
### Summary
Wrote all 7 T5 plan bullets in internal/compat/resticfx/runner_test.go and fixture_test.go, plus the reusable TestHelperProcess fake executable (modes: echo, fail) for T7. New compile stubs live only in zz_agentic_shim_t5.go (Runner, ExecRunner with value receiver, Fixture, NewFixture, Init, Backup, Snapshots, Destroy() error, ToolVersions). Wrong bodies return nil/"shim". Reused T1/T2/T3 symbols (Guard, GuardRepo, InitArgs, BackupArgs, SnapshotsArgs, NewPasswordFile, Snapshot, fullIDA) and redeclared nothing. T5 needs nothing from T4 (WriteTree, FileSpec, FileMeta). The recording fakeRunner is a real test helper in fixture_test.go. Contract choices GREEN must meet: the rclone version is the trimmed first line of output ("rclone v1.68.2"); "rclone not on PATH" is faked as &exec.Error{Err: exec.ErrNotFound}; the fake simulates restic init by creating the repo dir. The helper process was checked by hand and echoes argv verbatim and exits 3 with "boom". go vet is clean. golangci-lint reports 0 issues. The diff against 93e2ce1 is those 3 files. Commit 6619940.
### Result
| Check | Status | Detail |
|---|---|---|
| `internal/compat/resticfx/runner_test.go::TestExecRunnerPassesArgvVerbatim` | FAIL | stdout = "", want "a b\n$(echo pwned)\n;ls\n" |
| `internal/compat/resticfx/runner_test.go::TestExecRunnerErrorIncludesStderr` | FAIL | expected error from helper exiting 3, got nil |
| `internal/compat/resticfx/fixture_test.go::TestNewFixtureRefusesUnguardedRepo` | FAIL | expected guard error, got nil (home, other-remote) |
| `internal/compat/resticfx/fixture_test.go::TestFixtureInitBackupSnapshots` | FAIL | recorded 0 calls, want 3 |
| `internal/compat/resticfx/fixture_test.go::TestFixtureInitRefusesExistingRepo` | FAIL | Init on existing non-empty repo dir: expected error, got nil |
| `internal/compat/resticfx/fixture_test.go::TestFixtureDestroyOnlyOwnRepo` | FAIL | after Init, repo dir missing (no init call made) |
| `internal/compat/resticfx/fixture_test.go::TestToolVersions` | FAIL | restic version = "shim", want "0.19.0"; rclone = "shim" |
| `go vet ./...` | PASS | clean |
| `golangci-lint run ./internal/compat/resticfx/...` | PASS | 0 issues |
### Next
The scaffolder replaces zz_agentic_shim_t5.go with runner.go and fixture.go stubs: Runner, ExecRunner (value receiver Run), Fixture, NewFixture, Init, Backup, Snapshots, Destroy() error, ToolVersions. TestHelperProcess stays in runner_test.go for T7.

## S2-03/T6 · attempt 1 · red-worker · 2026-09-22T02:18:20Z
status: ok
### Summary
Wrote all 5 T6 tests from plan.md at the exact path::fn in package `internal/compat/resticfx` (new package; base 2158ade). Compile shim `zz_agentic_shim_t6.go` declares `Evidence` (JSON tags per tasks.md contract, `GeneratedAt time.Time`), `EvidenceDir`, `WriteEvidence`, `ReadEvidence`, `Probe`, `MissingPrerequisite` with deliberately wrong bodies (wrong dir string, errors, generic message). `NewPasswordFile` is NOT declared: `TestEvidenceOmitsPassword` uses a test-local helper `t6PasswordFile` (crypto/rand 32 bytes hex, 0600 file). Deviation: T5's Fixture/fake runner does not exist on this independent branch, so that test builds a populated Evidence (as a fixture run would produce) directly instead of driving a Fixture; it still asserts non-empty file and absence of the password value and password-file path. `PinnedResticVersion` (T1) is not referenced; tests use the literal `0.19.0`. Commit 76c10a3 on branch worktree-agent-a2a887a24411981ba; diff vs 2158ade = 3 files. go vet clean; golangci-lint 0 issues.
### Result
| Check | Status | Detail |
|---|---|---|
| `internal/compat/resticfx/evidence_test.go::TestEvidenceRoundTrip` | FAIL | t.Fatalf WriteEvidence: agentic shim error |
| `internal/compat/resticfx/evidence_test.go::TestWriteEvidenceFileName` | FAIL | darwin/linux subtests fatal on WriteEvidence error (empty-dir subtest passes on shim, expected) |
| `internal/compat/resticfx/evidence_test.go::TestEvidenceOmitsPassword` | FAIL | t.Fatalf WriteEvidence: agentic shim error |
| `internal/compat/resticfx/evidence_test.go::TestEvidenceDirFromEnv` | FAIL | EvidenceDir = "shim-wrong-evidence-dir", want temp dir / "" |
| `internal/compat/resticfx/prereq_test.go::TestMissingPrerequisite` | FAIL | all 8 rows fail: generic shim message, want "" or message naming the item |
| `go vet ./...` | PASS | clean |
| `golangci-lint run ./internal/compat/resticfx/...` | PASS | 0 issues |
### Next
Scaffold `evidence.go` (Evidence, EvidenceDir, WriteEvidence, ReadEvidence; literal "SNAPBACK_EVIDENCE_DIR") and `prereq.go` (Probe, MissingPrerequisite) as SUB-AGENT-TODO stubs, delete `zz_agentic_shim_t6.go`. Merge note: T4 owns `NewPasswordFile`; T6 tests do not reference it.

## S2-03/scaffold · attempt 1 · scaffolder · 2026-09-22T02:23:30Z
status: ok
### Summary
Consolidated the six overlapping RED shims (zz_agentic_shim_t1..t6.go; NewPasswordFile was declared in both t2 and t4) into ten canonical production files named per tasks.md. Every function/method body is panic("SUB-AGENT-TODO: <recipe>"); types keep the shim/spec shape; PinnedResticVersion holds an empty placeholder marked SUB-AGENT-TODO. One package doc comment (version.go). 29 exported symbols, each defined once; symbol list at .agentic/scaffold-symbols. VersionArgs (tasks.md T2) is not referenced by any test, so it was not stubbed. Commit abf9354 on top of 6cf7488.
### Scaffold
+ create PinnedResticVersion, ParseResticVersion, CheckPinnedVersion @ internal/compat/resticfx/version.go (package doc)
+ create Guard, GuardRepo @ internal/compat/resticfx/guard.go
+ create InitArgs, BackupArgs, SnapshotsArgs, LsArgs, MountArgs, UnmountCommand @ internal/compat/resticfx/args.go
+ create Snapshot, ParseSnapshots, CheckObservedIDs @ internal/compat/resticfx/snapshots.go
+ create FileSpec, FileMeta, WriteTree @ internal/compat/resticfx/tree.go
+ create NewPasswordFile @ internal/compat/resticfx/password.go
+ create Runner, ExecRunner, ExecRunner.Run @ internal/compat/resticfx/runner.go
+ create Fixture, NewFixture, Fixture.Init/Backup/Snapshots/Destroy/ToolVersions @ internal/compat/resticfx/fixture.go
+ create Evidence, EvidenceDir, WriteEvidence, ReadEvidence @ internal/compat/resticfx/evidence.go
+ create Probe, MissingPrerequisite @ internal/compat/resticfx/prereq.go
- delete internal/compat/resticfx/zz_agentic_shim_t1.go .. zz_agentic_shim_t6.go (6 files)
### Result
| Check | Status | Detail |
|---|---|---|
| go vet ./internal/compat/resticfx/ | PASS | compiles (GOTOOLCHAIN=auto) |
| gofmt -l | PASS | clean |
| stubs panic+TODO | PASS | 29 symbols; every func body panics SUB-AGENT-TODO |
| no newly passing test | PASS | per-test isolated run: 28 FAIL/panic; only TestHelperProcess passes (env-gated fake exec helper, no-op without GO_WANT_HELPER_PROCESS) |
| grep agentic:shim internal/ | PASS | empty |
| tests untouched | PASS | only non-test files changed |
### Next
green: fill the stubs per tasks.md T1-T6 (set PinnedResticVersion = "0.19.0"); then RED T7 (mount.go) against these symbols.

## S2-03/T1 · attempt 1 · green-worker · 2026-09-22T02:40:00Z
status: ok
### Summary
Filled the T1 stubs in version.go and guard.go. PinnedResticVersion is now "0.19.0". ParseResticVersion matches `^restic X.Y.Z` on the first line and returns an error for any other shape. CheckPinnedVersion's error names both the found and the wanted version. GuardRepo allows exactly `rclone:gdrive:snapback-stage1`, or an absolute path strictly under TempRoot. Before comparing, it resolves symlinks on the longest existing ancestor of both paths, which handles macOS /var -> /private/var. It refuses relative paths, `/`, TempRoot itself, `..` escapes, Home, and any other backend spec. Commit abc894c. The diff covers the 2 SCOPE files only.
### Result
| Check | Status | Detail |
|---|---|---|
| `internal/compat/resticfx/version_test.go::TestParseResticVersion` | PASS | 6/6 subtests, -race |
| `internal/compat/resticfx/version_test.go::TestCheckPinnedVersion` | PASS | -race |
| `internal/compat/resticfx/guard_test.go::TestGuardRepo` | PASS | -race |
| `go vet ./...` | PASS | clean |
| `golangci-lint run ./internal/compat/resticfx/...` | PASS (own files) | 0 issues in version.go/guard.go. 8 SA4006 issues come from sibling T5/T6 test files (evidence_test.go, fixture_test.go): unreachable code after their stubs panic. Those files are untouched and should clear once their GREENs land |
| diff scope | PASS | only version.go, guard.go |
### Next
Continue with the parallel S2-03 GREENs (T2+), then the structural review. The last S2-03 GREEN runs the full matrix, including golangci-lint on the package.

## S2-03/T2 · attempt 1 · green-worker · 2026-09-22T03:10:00Z
status: ok

### Summary
Filled the six SUB-AGENT-TODO stubs in `internal/compat/resticfx/args.go`. Each builder returns a fresh slice literal (so no two calls share a backing array) that starts with `-r repo --password-file pwFile`, then the subcommand. `MountArgs` passes `--path-template` and `ids/%I` as separate elements, with `mnt` last. `UnmountCommand` returns `umount mnt` on darwin, `fusermount3 -u mnt` on linux, and an error for any other GOOS. Five of the six T2 tests pass in the worktree. `TestArgsNeverContainPassword` fails there only because it calls `NewPasswordFile`, which is still T4's panicking stub in password.go. I left password.go alone. In an isolated scratch copy with a temporary crypto/rand `NewPasswordFile`, that test PASSES against this args.go. So it will pass once T4 lands, with no args.go change needed. Commit 7583f84.

### Result
| Check | Status | Detail |
|---|---|---|
| `internal/compat/resticfx/args_test.go::TestInitArgs` | PASS | -race |
| `internal/compat/resticfx/args_test.go::TestBackupArgs` | PASS | -race, includes dir with space/semicolon |
| `internal/compat/resticfx/args_test.go::TestSnapshotsAndLsArgs` | PASS | -race |
| `internal/compat/resticfx/args_test.go::TestMountArgs` | PASS | -race |
| `internal/compat/resticfx/args_test.go::TestUnmountCommand` | PASS | -race |
| `internal/compat/resticfx/args_test.go::TestArgsNeverContainPassword` | BLOCKED-ON-T4 | panics in password.go:5 NewPasswordFile stub; PASS in scratch copy with temporary NewPasswordFile |
| `go vet ./...` | PASS | clean |
| `golangci-lint run ./internal/compat/resticfx/...` | PASS for args.go | 8 staticcheck SA4006 findings, all in evidence_test.go and fixture_test.go (other tasks' RED tests, dead code after panicking stubs); none in args.go |
| diff scope | PASS | only internal/compat/resticfx/args.go |

### Next
Once T4 (password.go) merges, re-run `TestArgsNeverContainPassword` on the chain. Then structural review. T5 can consume these builders.

## S2-03/T3 · attempt 1 · green-worker · 2026-09-22T02:40:00Z
status: ok
### Summary
Filled the two SUB-AGENT-TODO stubs in internal/compat/resticfx/snapshots.go. ParseSnapshots json-decodes into Snapshot (json tags id/time/hostname/paths only; short_id is never decoded) and rejects any entry whose id fails ^[0-9a-f]{64}$ (covers missing, short, uppercase, non-hex). CheckObservedIDs errors on empty entries or an invalid fullID, requires an exact fullID entry, and fails if any entry is a strict prefix of fullID. Commit edf4bcf; diff limited to SCOPE.
### Result
| Check | Status | Detail |
|---|---|---|
| `internal/compat/resticfx/snapshots_test.go::TestParseSnapshotsFullIDs` | PASS | green (-race) |
| `internal/compat/resticfx/snapshots_test.go::TestParseSnapshotsRejectsBadIDs` | PASS | all 5 subtests |
| `internal/compat/resticfx/snapshots_test.go::TestParseSnapshotsEmptyAndGarbage` | PASS | green |
| `internal/compat/resticfx/snapshots_test.go::TestCheckObservedIDs` | PASS | all 7 rows |
| `go vet ./...` | PASS | clean |
| `golangci-lint run ./internal/compat/resticfx/...` (snapshots.go) | PASS | no findings in scope |
| validate.md T3 `grep -n 'ShortID\|short_id' snapshots.go` | PASS | no output |
### Next
Structural review; merge edf4bcf into chain2/s2-03 alongside the parallel T-tasks.

## S2-03/T4 · attempt 1 · green-worker · 2026-09-22T02:27:47Z
status: ok
### Summary
Filled the two T4 stubs. WriteTree rejects any path that is not filepath.IsLocal (absolute, "..", "a/../../x") before touching disk, creates parent dirs, writes Size deterministic bytes, chmods to Mode after the write, sets ModTime via os.Chtimes, and returns Lstat metadata. NewPasswordFile hex-encodes 32 crypto/rand bytes into an os.CreateTemp file under dir, chmods it to 0600, and returns only the path. Commit a786f20 on chain base abf9354; the diff touches only tree.go and password.go.
### Result
| Check | Status | Detail |
|---|---|---|
| `internal/compat/resticfx/tree_test.go::TestWriteTreeMetadata` | PASS | green (-race) |
| `internal/compat/resticfx/tree_test.go::TestWriteTreeRejectsEscapingPaths` | PASS | all 3 subtests green |
| `internal/compat/resticfx/password_test.go::TestNewPasswordFile` | PASS | 0600, 64 lowercase hex chars, unique |
| crypto/rand grep | PASS | one import line in password.go |
| go vet ./... | PASS | clean |
| golangci-lint ./internal/compat/resticfx/... | PASS (own files) | 0 issues in tree.go/password.go; 3 SA4006 findings already present in sibling RED test files evidence_test.go:84,105 and fixture_test.go:84 (not in T4 scope) |
| package-wide tests | EXPECTED FAIL | sibling stubs (e.g. InitArgs) still panic SUB-AGENT-TODO; GATE_RUN_MATRIX=0 |
### Next
Structural review. Orchestrator: the SA4006 lint findings in evidence_test.go and fixture_test.go (red-worker files) will block the story's final full-matrix gate (M-004); route them to a lint-only RED retry.

## S2-03/T6 · attempt 1 · green-worker · 2026-09-22T02:40:00Z
status: ok

### Summary
I filled the four stubs in evidence.go and prereq.go and committed them as aa64f19. `EvidenceDir` reads `SNAPBACK_EVIDENCE_DIR`. `WriteEvidence` rejects an empty dir. Otherwise it writes pretty-printed JSON with a trailing newline, mode 0644, to `dir/pathtemplate-<GOOS>.json`. `ReadEvidence` decodes the file and wraps any error. The `Evidence` struct has no field for the password or its file, so neither can reach the file. `MissingPrerequisite` checks in order: `SNAPBACK_FUSE_TESTS=1`, restic on PATH, FUSE (linux: /dev/fuse then fusermount3; darwin: /Library/Filesystems/macfuse.fs), then `CheckPinnedVersion`. Each message names the missing item. The version check reuses T1's `CheckPinnedVersion` instead of a second parser. That stub still panics on this chain, so two `TestMissingPrerequisite` subtests (`all_present_*`, `wrong_restic_version`) pass only once T1's GREEN merges. I confirmed all 8 subtests PASS with a temporary T1 body. That body was reverted and not committed.

### Result
| Check | Status | Detail |
|---|---|---|
| `internal/compat/resticfx/evidence_test.go::TestEvidenceRoundTrip` | PASS | green |
| `internal/compat/resticfx/evidence_test.go::TestWriteEvidenceFileName` | PASS | darwin, linux, empty dir |
| `internal/compat/resticfx/evidence_test.go::TestEvidenceOmitsPassword` | PASS | no password value or path in the file |
| `internal/compat/resticfx/evidence_test.go::TestEvidenceDirFromEnv` | PASS | green |
| `internal/compat/resticfx/prereq_test.go::TestMissingPrerequisite` | PASS* | 8/8 with T1's CheckPinnedVersion in place; alone on abf9354 it panics in T1's stub |
| `go vet ./...` | PASS | clean |
| `golangci-lint run ./internal/compat/resticfx/...` | PASS (own files) | 6 SA4006 in fixture_test.go (T5 RED file, already at base, outside scope) |
| diff scope | PASS | evidence.go and prereq.go only |

### Next
Merge after T1 GREEN, then re-run `TestMissingPrerequisite`. Hand the fixture_test.go SA4006 lint issues to T5 (or a RED lint retry) before T7's full matrix (M-004). Then structural review.

## S2-03/T5 · attempt 1 · green-worker · 2026-09-22T02:32:12Z
status: ok

### Summary
Filled the SUB-AGENT-TODO stubs in runner.go and fixture.go. `ExecRunner.Run` uses `exec.CommandContext(ctx, name, args...)` (argv array, no shell), returns stdout and wraps the exit error with stderr truncated to 4 KiB. `Fixture` holds runner/repo/pwFile/guard/created. `NewFixture` calls `GuardRepo` before storing anything, so a refused repo never reaches the runner. `Init` refuses an existing non-empty local repo before running `restic InitArgs`, and records `created` only for local repos. `Backup` and `Snapshots` run the T1 argv builders with the caller's ctx (`Snapshots` parses via `ParseSnapshots`). `Destroy` removes the repo only when this fixture created it and it still passes the guard, so rclone remotes are never touched. `ToolVersions` parses restic via `ParseResticVersion`, returns the first line of `rclone version`, and returns "" with no error when the runner error wraps `exec.ErrNotFound`. Commit 436df17 on e06cf09.

### Result
| Check | Status | Detail |
|---|---|---|
| `internal/compat/resticfx/runner_test.go::TestExecRunnerPassesArgvVerbatim` | PASS | green |
| `internal/compat/resticfx/runner_test.go::TestExecRunnerErrorIncludesStderr` | PASS | green |
| `internal/compat/resticfx/fixture_test.go::TestNewFixtureRefusesUnguardedRepo` | PASS | green |
| `internal/compat/resticfx/fixture_test.go::TestFixtureInitBackupSnapshots` | PASS | green |
| `internal/compat/resticfx/fixture_test.go::TestFixtureInitRefusesExistingRepo` | PASS | green |
| `internal/compat/resticfx/fixture_test.go::TestFixtureDestroyOnlyOwnRepo` | PASS | green |
| `internal/compat/resticfx/fixture_test.go::TestToolVersions` | PASS | green (both subtests) |
| T1-T4, T6 tests (incl. TestArgsNeverContainPassword, TestMissingPrerequisite) | PASS | whole package green under `go test -race` |
| `go vet ./...` | PASS | clean |
| `golangci-lint run ./internal/compat/resticfx/...` | PASS | 0 issues; the earlier SA4006 in fixture_test.go are gone now that the stubs no longer panic |
| diff scope | PASS | runner.go and fixture.go only |

### Next
Structural review, then T7 (mount supervisor, integration test, full matrix).

## S2-03/T7 · attempt 1 · red-worker · 2026-09-22T03:10:00Z
status: ok
### Summary
Wrote the 4 T7 plan bullets. The mount supervisor unit tests (`mount_test.go`) run a dedicated fake executable, `TestMountHelperProcess`, with modes mount/hang/exit and gated on `GO_WANT_HELPER_PROCESS=1`. It lives in mount_test.go so T5's runner_test.go is left untouched. The tests call `zz_agentic_shim_t7.go`, which declares `MountStarter{Unmount func(string) error; Grace time.Duration}`, `Mount`, `StartMount`, `(*Mount).WaitReady`, `Stop` and `Exited` with deliberately wrong bodies, so all 3 tests fail by assertion. `pathtemplate_integration_test.go` is tagged `//go:build integration` and skips with `MissingPrerequisite`'s message. Otherwise it runs init -> backup -> snapshots -> `restic mount --path-template ids/%I`, registers Stop in t.Cleanup before WaitReady, runs CheckObservedIDs, lists `ids/<fullID>/<src>` and asserts the mount is gone after cleanup. It writes evidence whether it passes or fails and asserts result=pass, restic 0.19.0 and snapshot_id==observed_dir. A relative `SNAPBACK_EVIDENCE_DIR` is resolved against the module root, because go test runs in the package dir and validate.md passes `docs/reports/stage1`. go vet (default and `-tags integration`) and golangci-lint (default and `--build-tags integration`) are clean. Commit 961ee77 touches 3 files only.
### Result
| Check | Status | Detail |
|---|---|---|
| `internal/compat/resticfx/mount_test.go::TestMountWaitReadyAndStop` | FAIL | WaitReady: shim: WaitReady not implemented, want nil once <mnt>/ids exists |
| `internal/compat/resticfx/mount_test.go::TestMountWaitReadyTimeout` | FAIL | WaitReady = shim error, want an error wrapping context.DeadlineExceeded |
| `internal/compat/resticfx/mount_test.go::TestMountStopIdempotent` | FAIL | Stop twice = (shim err, shim err), want (nil, nil) |
| `internal/compat/resticfx/pathtemplate_integration_test.go::TestPathTemplateIntegration` | GATED | integration tag only; with SNAPBACK_FUSE_TESTS unset it SKIPs with "missing prerequisite: SNAPBACK_FUSE_TESTS=1 is not set"; not run in the default build |
| T1-T6 tests | PASS | unchanged |
| go vet ./... and -tags integration | PASS | clean |
| golangci-lint ./internal/compat/resticfx/... | PASS | 0 issues (default and integration tags) |
### Next
Scaffold mount.go with MountStarter, Mount, StartMount, WaitReady, Stop and Exited (signatures as in the shim), and delete zz_agentic_shim_t7.go. GREEN then implements a cancel-only exec.CommandContext, polling for <mnt>/ids, and a Stop that sends Interrupt, waits Grace, unmounts (at most once) and kills. Stop must be idempotent via sync.Once and return nil for exit 0.
