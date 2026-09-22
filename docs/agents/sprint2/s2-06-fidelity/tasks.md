---
type: tasks
story: S2-06
---

# S2-06 tasks — Metadata-fidelity check through `.snapshot/<timestamp>/`

Story intent and owned files: `docs/agents/sprint2/stories.md` (S2-06). Every task touches only `internal/compat/fidelity/**`. The evidence files `docs/reports/stage1/fidelity-darwin.json` and `fidelity-linux.json` are written by the orchestrator's evidence runs (see `docs/agents/sprint2/plan.md` § Waves), not by a worker. Tests are listed in `plan.md`; gate commands in `validate.md`. Spec anchors: SPEC.md §5 (metadata fidelity), §7 (historical metadata is truthful; the catalog never proxies file content), §20 (disposable fixtures, real mount tests).

Upstream contracts consumed (from `stories.md`, not re-implemented here):

- **S2-03 `internal/compat/resticfx`** (production code, importable): pinned-version check (`0.19.0`), generated-tree writer that sets chosen sizes, modes and mtimes and returns the expected metadata, repo init with a 0600 `--password-file`, backup, `snapshots --json` returning the full 64-hex ID, `restic mount --path-template ids/%I` as a supervised process with readiness wait and unmount, `ls --json` argument building, repo destroy, and the temp-dir repository guard.
- **S2-04 `internal/mount/gofuse`**: `*Adapter` implementing `mount.Adapter` (mount a `mount.Catalog` read-only at a caller-owned empty directory; unmount).
- **S2-02 `internal/projection`**: builds an immutable generation from a spec of named directories and symlinks with raw target strings.

Decisions fixed here so workers do not guess:

- Asserted attributes are exactly **mtime, size and mode** (`fs.ModeType|fs.ModePerm` bits). uid, gid, ctime and birth time are never asserted. ctime and birth time are recorded in the evidence with `"claimed": false` and the note `FUSE-approximated; recorded, not claimed` (§5).
- `const MTimeTolerance time.Duration = 0` (exact match). It is written into the evidence JSON as `mtime_tolerance_ns`. Any precision loss is recorded as `mtime_precision_ns` and as a per-file `mtime_delta_ns`. The tolerance is never widened inside a task. Changing it is a human decision (see validate.md Final sign-off).
- The `.snapshot` alias name follows SPEC.md §3 "Timestamp names": UTC `YYYY-MM-DD_HHMMZ` (e.g. `2026-09-21_0300Z`), i.e. `const SnapshotDirLayout = "2006-01-02_1504Z"` applied to the restic snapshot time in UTC. The collision suffixes (seconds, then a short ID) never arise with the single snapshot used here, and `timestamps: local` is out of scope for Stage 1. The catalog's projection is `/.snapshot/<name>` -> symlink with absolute target `<resticMnt>/ids/<fullID>`.
- Expected values come only from (a) the resticfx generator's returned metadata and (b) the parsed `restic ls --json` output. Nothing is ever read back from either mount to build an expectation, because that would be circular.
- Symlinks: the fixture symlink's own mtime cannot be set portably with the standard library (`os.Chtimes` follows links; `go.mod` is S2-01's). For the symlink, the check asserts the `ModeSymlink` type bit via `Lstat`, the `Readlink` target, and the mtime and size against `restic ls --json` only (not against generation time).
- Evidence file name: `fidelity-<runtime.GOOS>.json` under `$SNAPBACK_EVIDENCE_DIR`. The file is written **before** the pass/fail assertions (in a `t.Cleanup` that runs even on failure), so a failing platform is recorded as "measured and failing", never skipped.

Cross-story names (binding): `mount.Catalog` (S2-01) uses builtin-type signatures only — `Lookup(parent uint64, name string) (ino uint64, isDir bool, found bool)`, `ReadDir(dir uint64) (names []string, found bool)`, `Readlink(ino uint64) (target string, found bool)` — with `mount.RootIno = 1` as the root inode. Catalogs are built with `projection.Build(Spec) (*Generation, error)` (S2-02); `*projection.Generation` satisfies `mount.Catalog` structurally. Timestamp aliases use the format `2006-01-02_1504Z` (SPEC §3).

## T1 — Pure comparator and mtime-precision probe

`compare.go` declares `type Meta struct { Path string; Size int64; Mode fs.FileMode; MTime time.Time; LinkTarget string }` and `type Result struct { Path string; SizeOK, ModeOK, MTimeOK bool; MTimeDelta time.Duration }`. It exposes `func Compare(expected, observed Meta, tol time.Duration) Result` (pure; mode compared on `fs.ModeType|fs.ModePerm`; `|delta| <= tol`), `func CompareAll(expected []Meta, observed map[string]Meta, tol time.Duration) ([]Result, error)` (returns an error for an empty `expected` (M-002) and for any expected path missing from `observed`; one result per expected file, in input order), and `func Precision(ts []time.Time) time.Duration` (the coarsest of 1ns, 1µs, 1ms, 1s that divides every `t.Nanosecond()`). `const MTimeTolerance time.Duration = 0`.

Files: `internal/compat/fidelity/compare.go`, `internal/compat/fidelity/compare_test.go`.

## T2 — `restic ls --json` parser

`resticls.go` exposes `func ParseResticLs(r io.Reader, root string) ([]Meta, error)`. It reads restic 0.19.0's newline-delimited JSON, skips the leading `"struct_type":"snapshot"` line and `dir` nodes, keeps `file` and `symlink` nodes, maps `path` (with the `root` prefix stripped, so paths are relative with `/` separators), `size`, `mode` (numeric Go `fs.FileMode`), `mtime` (RFC 3339 with nanoseconds) and `linktarget`. Any malformed line returns an error naming the line number. It does not run restic; the integration test feeds it the stdout of the resticfx `ls --json` command.

Files: `internal/compat/fidelity/resticls.go`, `internal/compat/fidelity/resticls_test.go`.

## T3 — Fixture spec and platform observation

`fixtures.go` exposes `func Fixtures() []Meta`, the fixed fixture set handed to the resticfx generator: at least 6 regular files with pairwise-distinct mtimes (including one sub-second value like `…:07.123456789Z` and one pre-2000 value, `1998-03-14T09:26:53Z`), sizes 0, 1 KiB-ish and 3 MiB, modes 0644, 0600, 0755 and 0444, one name containing a space (`with space.txt`), one Unicode name in NFC (`café-日本.txt`), one file in a subdirectory, plus one relative symlink (`link-to-small` -> `small.txt`). `observe.go` exposes `type Observed struct { Meta; CTime time.Time; BirthTime *time.Time }` and `func Observe(path string) (Observed, error)`, which uses `os.Lstat` (never following the final component) and `os.Readlink` for symlinks. Platform files fill `CTime`/`BirthTime` from `syscall.Stat_t`: darwin sets both (`Ctimespec`, `Birthtimespec`); linux sets `CTime` (`Ctim`) and leaves `BirthTime` nil because birth time needs `statx`, which is outside the stdlib, and that is recorded as `not exposed`.

Files: `internal/compat/fidelity/fixtures.go`, `internal/compat/fidelity/fixtures_test.go`, `internal/compat/fidelity/observe.go`, `internal/compat/fidelity/observe_darwin.go`, `internal/compat/fidelity/observe_linux.go`, `internal/compat/fidelity/observe_test.go`.

## T4 — Evidence JSON and prerequisite probe

`evidence.go` declares `type Evidence struct` with `platform` (`GOOS/GOARCH`), `restic_version`, `snapshot_id` (64-hex), `snapshot_alias` (`.snapshot/<name>`), `resolved_inside_restic_mount` (bool), `mtime_tolerance_ns`, `mtime_precision_ns`, `files_generated`, `files_compared`, `pass`, and `files[]` with per-file `path`, `expected` / `observed` `{size, mode, mtime}`, `restic_ls` `{size, mode, mtime}`, `size_ok` / `mode_ok` / `mtime_ok`, `mtime_delta_ns`, plus `ctime` / `birth_time` objects `{value, claimed:false, note}` (`birth_time.value` null where not exposed). It exposes `func WriteEvidence(dir string, ev Evidence) (string, error)` (writes `fidelity-<runtime.GOOS>.json`, indented, mode 0644, errors if `dir` does not exist; returns the path) and `func ReadEvidence(path string) (Evidence, error)`. `prereq.go` exposes `func MissingPrereq(getenv func(string) string, lookPath func(string) (string, error), exists func(string) bool, goos string) string`, which returns "" when everything is present, otherwise the exact skip message for the first missing item, in this order: `SNAPBACK_FUSE_TESTS not set: fidelity tests skipped`, `restic not on PATH: fidelity tests skipped`, `fuse3 device /dev/fuse not present` (linux), `macFUSE not installed: /Library/Filesystems/macfuse.fs missing` (darwin). The restic version check stays in resticfx (S2-03); this task does not duplicate it.

Files: `internal/compat/fidelity/evidence.go`, `internal/compat/fidelity/evidence_test.go`, `internal/compat/fidelity/prereq.go`, `internal/compat/fidelity/prereq_test.go`.

## T5 — Gated integration test through the catalog alias, and full matrix

`fidelity_integration_test.go` (`//go:build integration`) holds one test. It skips with `MissingPrereq(os.Getenv, exec.LookPath, exists, runtime.GOOS)` when that returns non-empty. Otherwise, all under `t.TempDir()`, it: runs the resticfx version check; generates `Fixtures()` with the resticfx writer; inits a disposable repo and backs it up; reads the full snapshot ID and time from `snapshots --json`; parses `restic ls --json <id>` with `ParseResticLs`; starts `restic mount --path-template ids/%I` and waits for readiness; builds a projection with `/.snapshot/<name>` -> `<resticMnt>/ids/<fullID>` and mounts it with `gofuse.Adapter`; asserts that `filepath.EvalSymlinks(<catalogMnt>/.snapshot/<name>)` has the restic mountpoint as its prefix (§7: served by restic, not proxied); `Observe`s every fixture path through `<catalogMnt>/.snapshot/<name>/…`; runs `CompareAll` against the generator's values (regular files) and against the `ls --json` values (all entries, including the symlink); records ctime/birth time; writes the evidence when `SNAPBACK_EVIDENCE_DIR` is set; and asserts `files_compared == files_generated >= 6` and every file's size, mode and mtime pass. Cleanup unmounts the catalog, then restic, then destroys the repo, even on failure. This is the story's last GREEN, so it runs the **full** gate matrix (with 80% total coverage) and then the local integration run (M-005).

Files: `internal/compat/fidelity/fidelity_integration_test.go`.

## Dependencies and parallelism

| Task | Depends on | Parallel with |
| --- | --- | --- |
| T1 | S2-03 and S2-04 merged green (wave 3 entry) | T2 |
| T2 | wave 3 entry | T1, T3, T4 |
| T3 | T1 (uses `Meta`) | T2, T4 |
| T4 | T1 (uses `Meta`, `Result`) | T2, T3 |
| T5 | T1, T2, T3, T4 | none (last task; full matrix) |

T1 through T4 own disjoint files. T3 and T4 compile against T1's `Meta`/`Result` types. The RED worker for T3 or T4 may add a mod-common shim for them, which the scaffolder removes.
