---
type: plan-ready
story: S2-08
from_red_at: 2026-09-22T02:52:59Z
---

# S2-08 plan-ready (RED verified by orchestrator at chain @ e85e18b; every box currently FAILs by assertion unless noted)


# S2-08 test plan (tests only)

Contracts under test: `tasks.md` (S2-08). Every test is pure: no network, no rclone, no restic, no FUSE. Orchestration uses a fake `Runner` that records `(name, args)` and returns scripted output/errors, and a fake `Clock` that advances a fixed step per `Now()`. Negative checks assert the expected non-empty content first (M-002). The real Google Drive run is not a test; it is exit evidence in `validate.md`.

## T1 — Remote guard

- [x] `internal/compat/latency/guard_test.go::TestCheckRemoteAllowsOnlyExactRemote` — table: `gdrive:snapback-stage1` -> nil; `""`, `gdrive:`, `gdrive:snapback-stage1/x`, `gdrive:snapback-stage1 `, `GDRIVE:snapback-stage1`, `gdrive:other`, `s3:x`, `rclone:gdrive:snapback-stage1`, `/tmp/repo` -> error whose text contains `gdrive:snapback-stage1`.
- [x] `internal/compat/latency/guard_test.go::TestRemoteFromEnv` — fake getenv: unset -> error naming `SNAPBACK_RCLONE_REMOTE` and `gdrive:snapback-stage1`; `gdrive:other` -> refused; allowed value -> returned unchanged, nil error.
- [x] `internal/compat/latency/guard_test.go::TestRepoSpecIsConstant` — assert `RepoSpec() == "rclone:gdrive:snapback-stage1"` and `AllowedRemote == "gdrive:snapback-stage1"`.

## T2 — Timing stats and result schema

- [x] `internal/compat/latency/result_test.go::TestSummarize` — table: `[5ms]` -> 5/5/5; `[30ms,10ms,20ms]` -> median 20, min 10, max 30; `[4ms,1ms,3ms,2ms]` -> median 2.5; `[1500us]` -> 1.5 (sub-ms precision kept); assert exact float values.
- [x] `internal/compat/latency/result_test.go::TestSummarizeEmptyErrors` — `Summarize(nil)` and `Summarize([]time.Duration{})` return a non-nil error (no median claimed from zero samples).
- [x] `internal/compat/latency/result_test.go::TestEncodeSchemaKeys` — build a full Result; `Encode`; unmarshal to `map[string]any`; assert every required key is present (`measurements`, `restic_version`, `rclone_version`, `os`, `arch`, `timestamp_utc`, `remote`, `data_file_count`, `data_file_bytes`, `remote_deleted`) and the four measurement names each carry `samples_ms`, `median_ms`, `min_ms`, `max_ms`.
- [x] `internal/compat/latency/result_test.go::TestEncodeRemoteDeletedFalseIsPresent` — Result with `RemoteDeleted=false`; assert decoded map has key `remote_deleted` with value `false` (no `omitempty` drop).
- [x] `internal/compat/latency/result_test.go::TestEncodeHasNoThresholdOrVerdict` — after asserting the encoded JSON has the four measurements (M-002), walk every key recursively and assert none matches `(?i)threshold|verdict|pass|fail|^go$|nogo|no_go|^ok$|acceptable|limit|target`.
- [x] `internal/compat/latency/result_test.go::TestEncodeRejectsMissingMeasurement` — Result with only three measurements; assert `Encode` returns an error naming the missing one.
- [x] `internal/compat/latency/result_test.go::TestTimestampIsUTCRFC3339` — Result built from a non-UTC fake time; assert `timestamp_utc` parses with `time.RFC3339` and ends in `Z`.
- [x] `internal/compat/latency/result_test.go::TestNoThresholdSymbolsInPackage` — parse every non-test `.go` file in the package with `go/parser` (assert >= 1 file found, M-002); assert no identifier or string literal matches `(?i)threshold|verdict|nogo|no_go|acceptable`.

## T3 — Disposable secrets and scratch dirs

- [x] `internal/compat/latency/scratch_test.go::TestNewScratchLayout` — `newScratch`; assert root is under `os.TempDir()`, password file mode is `0600`, content is >= 64 hex chars, cache/data/mount dirs exist and are empty.
- [x] `internal/compat/latency/scratch_test.go::TestNewScratchPasswordsDiffer` — two scratches; assert different password contents.
- [x] `internal/compat/latency/scratch_test.go::TestScratchCloseRemovesEverything` — `Close`; assert root no longer exists (`os.IsNotExist`).
- [x] `internal/compat/latency/scratch_test.go::TestScratchOutsideRepoTree` — assert root does not have the repo working directory (`os.Getwd` at test start) as a prefix, so the password is never written where it could be committed.

## T4 — Cleanup and deletion verification

- [x] `internal/compat/latency/cleanup_test.go::TestVerifyDeleted` — table: (empty, nil) -> true; (nil, error "directory not found") -> true; (`"data/\n"`, nil) -> false; (nil, error "couldn't connect") -> false; (`"keys/\n"`, error "directory not found") -> false.
- [x] `internal/compat/latency/cleanup_test.go::TestPurgeAndVerifyCommands` — fake runner; `purgeAndVerify`; assert exactly two calls in order: `rclone purge gdrive:snapback-stage1`, `rclone lsf gdrive:snapback-stage1` (args arrays, no shell) and result true when lsf reports not found.
- [x] `internal/compat/latency/cleanup_test.go::TestPurgeErrorStillVerifies` — purge returns an error and lsf returns `data/`; assert lsf still called and result false.
- [x] `internal/compat/latency/cleanup_test.go::TestCheckEmptyRefusesNonEmptyRemote` — table: lsf "directory not found" -> nil; lsf empty -> nil; lsf `config\n` -> error mentioning `non-empty`; lsf other error -> error.

## T5 — Orchestrated run

- [x] `internal/compat/latency/run_test.go::TestRunCommandSequence` — scripted fake runner; `Run`; assert the recorded program/subcommand order is exactly: lsf (pre-check), restic version, rclone version, init, backup, snapshots, mount, list, ls, list x`Samples`, read, unmount, mount, list x`Samples`, unmount, purge, lsf.
- [x] `internal/compat/latency/run_test.go::TestRunUsesFreshTempCacheDir` — assert every restic call carries `--cache-dir <d>` with the same `d` under `os.TempDir()`, and `d` is not `os.UserCacheDir()` or below it; both mounts share `d`.
- [x] `internal/compat/latency/run_test.go::TestRunPrewarmUsesFullSnapshotID` — snapshots returns a 64-hex ID; assert the prewarm call is `ls --json <that full ID>` and mount uses `--path-template ids/%I`.
- [x] `internal/compat/latency/run_test.go::TestRunPasswordNeverInArgsOrResult` — read the password file contents via a runner hook; assert no recorded arg and no byte of `Encode(result)` contains it, and `--password-file` is present on restic calls.
- [x] `internal/compat/latency/run_test.go::TestRunRecordsSamplesFromClock` — fake clock step 7ms; assert each measurement's samples equal 7ms, cold ones have exactly 1 sample, warm ones have `Samples` samples, and medians match `Summarize`.
- [x] `internal/compat/latency/run_test.go::TestRunCleanupAfterMidRunFailure` — runner fails the prewarm `ls`; assert `Run` returns an error, unmount, purge and lsf were still called after the failure, and returned Result has `RemoteDeleted==true` when lsf reports not found.
- [x] `internal/compat/latency/run_test.go::TestRunReportsFailedDelete` — purge succeeds but lsf lists `data/`; assert `Result.RemoteDeleted==false` and the returned error mentions `gdrive:snapback-stage1` and `not deleted`.
- [x] `internal/compat/latency/run_test.go::TestRunAbortsOnNonEmptyRemoteBeforeWrite` — pre-check lsf lists `config`; assert error, and no init, backup, mount or purge call was recorded (it never deletes what it did not create).
- [x] `internal/compat/latency/run_test.go::TestRunRefusesOtherRemoteBeforeAnyCommand` — `Config.Remote="gdrive:other"`; assert error naming `gdrive:snapback-stage1` and zero runner calls.
- [x] `internal/compat/latency/run_test.go::TestRunRemovesScratch` — capture the scratch root via the password-file arg; after `Run` (success and failure cases) assert the root no longer exists.

## T6 — Thin CLI and full matrix

- [x] `tools/stage1latency/run_test.go::TestRunMissingOutFlagExits2` — args `[]`; assert return 2, stderr contains `-out`, no runner call.
- [x] `tools/stage1latency/run_test.go::TestRunRefusedRemoteExits2` — getenv returns `gdrive:other` then unset; assert return 2, stderr contains `gdrive:snapback-stage1`, no file at `-out`.
- [x] `tools/stage1latency/run_test.go::TestRunWritesJSONAndExits0` — injected fake latency deps; `-out <t.TempDir()>/latency.json`; assert return 0, file parses, has four measurements and `remote_deleted: true`.
- [x] `tools/stage1latency/run_test.go::TestRunRemoteNotDeletedExits1Loudly` — fake deps with lsf listing `data/`; assert return 1, JSON still written with `remote_deleted: false`, stderr contains `NOT DELETED` and `gdrive:snapback-stage1`.
- [x] `tools/stage1latency/run_test.go::TestMainIsThin` — read `main.go`; assert it contains `os.Exit(run(` and has <= 12 non-empty lines.

Test count: 34 (T1 3, T2 8, T3 4, T4 4, T5 10, T6 5).
