---
type: tasks
story: S2-08
---

# S2-08 tasks — rclone/Google Drive latency harness and measurements

Story intent and owned files: `docs/agents/sprint2/stories.md` (S2-08). Owned set: `internal/compat/latency/**`, `tools/stage1latency/**`, `docs/reports/stage1/latency.json`. Depends on S2-03 (`internal/compat/resticfx`: runner, guard, repo init/backup/mount/unmount, versions) merged green. Tests: `plan.md`. Gates and exit evidence: `validate.md`.

Decisions fixed here so workers do not guess:

- Allowed remote is the constant `AllowedRemote = "gdrive:snapback-stage1"`; the restic repo spec is `"rclone:" + AllowedRemote`. Any other value (including `gdrive:`, `gdrive:snapback-stage1/x`, `gdrive:other`, `s3:x`, empty) is refused with an error that names `gdrive:snapback-stage1`. The `rclone purge` and `rclone lsf` targets are built only from `AllowedRemote`, never from input.
- The password is random, written to a 0600 file inside a fresh `os.MkdirTemp` dir that the run removes; it is passed only via `--password-file` (resticfx). It never appears in args, logs or JSON, and no file under the repo tree is ever written with it.
- Test data is generated only (resticfx generated-tree writer). `DataFileCount` and `DataFileBytes` are named constants recorded in the JSON. `Samples` (named constant, >= 3) is the repeat count where a measurement allows repetition; cold measurements are 1 sample each because a second read is no longer cold, and the JSON records the actual count per measurement.
- The cache dir is a fresh empty `os.MkdirTemp` dir passed as `--cache-dir`; the same dir is reused for the remount ("after restart"). The default user cache is never used.
- No threshold, verdict, pass/fail or go/no-go is computed or emitted anywhere (human decision). The JSON has no key matching `threshold|verdict|pass|fail|go|ok|acceptable`.
- Deletion is `rclone purge gdrive:snapback-stage1`, verified by `rclone lsf gdrive:snapback-stage1`: deleted means lsf returns a "directory not found" error or empty output. It runs in a deferred cleanup even after a mid-run failure; a failed or unverified delete sets `remote_deleted: false` and makes the CLI exit non-zero with a loud stderr line.
- Collaborators are injected: `Runner` (finite commands and the long-running mount, satisfied by resticfx's runner) and `Clock` (`Now() time.Time`), so all orchestration is unit-tested with fakes and no network.
- The real run is NOT a test. It is the orchestrator command `SNAPBACK_RCLONE_REMOTE=gdrive:snapback-stage1 go run ./tools/stage1latency -out docs/reports/stage1/latency.json`, recorded in `validate.md` as exit evidence. No `//go:build integration` test is added for S2-08 (resolved in stories.md S2-08: the run happens only after explicit human confirmation).

## T1 — Remote guard

`AllowedRemote` constant, `CheckRemote(remote string) error` (exact match only), and `RemoteFromEnv(getenv func(string) string) (string, error)` reading `SNAPBACK_RCLONE_REMOTE` (unset is an error naming the variable and the allowed value). `RepoSpec() string` returns `"rclone:gdrive:snapback-stage1"`.

Files: `internal/compat/latency/guard.go`, `internal/compat/latency/guard_test.go`.

## T2 — Timing stats and result schema

`Summarize(samples []time.Duration) (Stats, error)` returns median/min/max in milliseconds (float64; even count median is the mean of the middle two); empty input is an error. `Result` struct with JSON tags: `measurements` (map of the four names `cold_listing`, `warm_prewarmed_listing`, `cold_first_file_read`, `warm_listing_after_restart` to `{samples_ms, median_ms, min_ms, max_ms}`), `restic_version`, `rclone_version`, `os`, `arch`, `timestamp_utc` (RFC 3339, UTC), `remote`, `data_file_count`, `data_file_bytes`, `remote_deleted` (bool, always present, no `omitempty`). `Encode(Result) ([]byte, error)` produces indented JSON and refuses a Result missing any of the four measurements.

Files: `internal/compat/latency/result.go`, `internal/compat/latency/result_test.go`.

## T3 — Disposable secrets and scratch dirs

`newScratch() (scratch, error)` creates one `os.MkdirTemp` root holding a random password file (mode 0600, >= 32 random bytes hex-encoded), an empty cache dir, a data dir and a mount dir; `scratch.Close() error` removes the root. Refuses to create anything outside `os.TempDir()`.

Files: `internal/compat/latency/scratch.go`, `internal/compat/latency/scratch_test.go`.

## T4 — Cleanup and deletion verification

`VerifyDeleted(lsfOut []byte, lsfErr error) bool` (true only for empty output with nil error, or an error whose text contains `directory not found`; any other error or any listed entry is false). `purgeAndVerify(ctx, r Runner) bool` runs `rclone purge gdrive:snapback-stage1` then `rclone lsf gdrive:snapback-stage1` and returns the verified result; a purge error still runs lsf. `checkEmpty(ctx, r Runner) error` runs the same lsf before any write and errors if the path exists and is non-empty.

Files: `internal/compat/latency/cleanup.go`, `internal/compat/latency/cleanup_test.go`.

## T5 — Orchestrated run

`Run(ctx context.Context, cfg Config) (Result, error)` where `Config{Remote string; Runner Runner; Clock Clock}`. Order: guard -> `checkEmpty` -> scratch -> versions -> init -> backup -> snapshots (full ID) -> mount -> time cold listing -> (mount stays up) -> `restic ls --json <fullID>` prewarm -> time warm listing x`Samples` -> time cold first-file read -> unmount -> remount same cache dir -> time warm-after-restart listing x`Samples` -> unmount -> deferred `purgeAndVerify` -> scratch close. Every timing is `Clock.Now()` before and after a single operation. A mid-run failure still unmounts, purges and verifies, and returns the error together with the partial Result carrying `remote_deleted`.

Files: `internal/compat/latency/run.go`, `internal/compat/latency/run_test.go`.

## T6 — Thin CLI and full matrix

`tools/stage1latency`: `run(args []string, getenv func(string) string, stdout, stderr io.Writer) int` parses `-out <path>` (required), reads the remote via `RemoteFromEnv`, calls `latency.Run`, writes `Encode` output to `-out`, prints the path. Exit 0 only when Run succeeded and `remote_deleted` is true; refused remote or missing flag exits 2; any other failure exits 1 and still writes the JSON if a Result exists. `main` is only `os.Exit(run(os.Args[1:], os.Getenv, os.Stdout, os.Stderr))`, with the real resticfx runner and `time.Now` clock wired in a small `deps.go`. This is the story's last task: its GREEN runs the full standards matrix (M-005).

Files: `tools/stage1latency/main.go`, `tools/stage1latency/deps.go`, `tools/stage1latency/run.go`, `tools/stage1latency/run_test.go`.
