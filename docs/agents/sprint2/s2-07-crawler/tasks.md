---
type: tasks
story: S2-07
---

# S2-07 tasks — Crawler-safety hit-count test

Owned set (stories.md S2-07): `internal/compat/crawler/**`, `docs/reports/stage1/crawler-darwin.json`, `docs/reports/stage1/crawler-linux.json`. No task touches anything else. `go.mod`/`go.sum` belong to S2-01 and are not edited (standard library plus the already-pinned S2-01/S2-02/S2-04 packages only).

Scope: MEASUREMENT only (SPEC.md §7 crawler safety; §20 acceptance 12, Stage 1 portion; §21 row 1 "crawler test"). `catalog.reader_policy`, throttling and denial are Stage 3 and are not implemented, stubbed or asserted here. The only assertion on counts is zero catalog hits for the non-following tools (`rg`, `fd`, `find`, `rsync -a`). Following tools (`rg -L`, `fd -L`, `find -L`) have their counts recorded, never judged. VS Code search with `search.followSymlinks` is a fixed `not-tested-here` row.

Package: `internal/compat/crawler` (`package crawler`, tests `package crawler`). Upstream contract, taken from stories.md because S2-01/S2-04 may not have merged when this plan was written: `mount.Observer` receives one event per catalog operation (op kind + node path); `gofuse` fires it for every lookup, readdir and readlink; `*gofuse.Adapter` mounts a `mount.Catalog` at a caller-owned empty directory and unmounts; `projection` builds the catalog. RED binds to whatever exact identifiers S2-01/S2-04 merged (validate.md Pre-flight checks they exist).

Order: T1, T2, T3, T4 are independent pure units (different files, no shared symbols except T3 importing T2's `Status` type — so T2 before T3). T5 depends on T1–T4 and is last: it runs the full standards matrix plus the gated integration suite.

Cross-story names (binding): `mount.Catalog` (S2-01) uses builtin-type signatures only — `Lookup(parent uint64, name string) (ino uint64, isDir bool, found bool)`, `ReadDir(dir uint64) (names []string, found bool)`, `Readlink(ino uint64) (target string, found bool)` — with `mount.RootIno = 1` as the root inode. Catalogs are built with `projection.Build(Spec) (*Generation, error)` (S2-02); `*projection.Generation` satisfies `mount.Catalog` structurally. Timestamp aliases use the format `2006-01-02_1504Z` (SPEC §3).

## T1 — Catalog hit counter (Observer)

- **Files:** `internal/compat/crawler/counter.go` (create), `internal/compat/crawler/counter_test.go` (create).
- **Does:** `Counter` implements `mount.Observer`. It counts every event, keeps a per-op breakdown (lookup, readdir, readlink), and exposes `Total()`, `ByOp()` (returns a copy, never the internal map) and `Reset()`. It is safe for concurrent use because FUSE callbacks arrive on multiple goroutines (sync/atomic or a mutex). No policy, no throttling.
- **Tests:** plan.md T1 block (5 tests).

## T2 — Tool table and binary resolution

- **Files:** `internal/compat/crawler/tools.go` (create), `internal/compat/crawler/tools_test.go` (create).
- **Does:** A fixed, ordered table of seven `Tool` rows: `rg`, `rg -L`, `fd`, `fd -L`, `find`, `find -L`, `rsync -a`. Each row has a display name, candidate binaries (`fd` rows: `fd` then `fdfind`, for Debian/Ubuntu), a `Follows` flag, a function that builds the argv for a given root (and, for rsync, a destination), and the argv used to read the version. `rg` and `fd` rows pass `--hidden --no-ignore` (`fd`: `-H -I`) so the hidden `.snapshot` link is actually visited; without that, zero hits would be vacuous. `Resolve(tool, lookPath)` takes an injected `lookPath` func (default `exec.LookPath`) and returns the first candidate found or an error wrapping `ErrToolMissing` that names every candidate tried. A separate fixed `VSCodeRow` carries status `not-tested-here` and a reason; it has no argv and cannot be run. Argument arrays only: no `sh -c`, no shell metacharacters.
- **Tests:** plan.md T2 block (8 tests).

## T3 — Result schema, acceptance check and JSON writer

- **Files:** `internal/compat/crawler/evidence.go` (create), `internal/compat/crawler/evidence_test.go` (create).
- **Does:** `Report` (schema version, goos, goarch, seed shape, positive-control hits, rows) and `Row` (tool, argv, follows, status, hits as `*int`, per-op counts, version, exit code, duration, reason). Status values: `tested`, `not-tested-here`, `timed-out`. `Validate()` rejects a `not-tested-here` row that carries a hit count or lacks a reason, and a `tested` row with no hit count (a skipped tool must never be folded into a zero-hit pass). `NonFollowingViolations(report)` returns the tested non-following rows with hits > 0 or status `timed-out`; following and skipped rows are ignored. `EvidenceFileName(goos)` returns `crawler-<goos>.json` and rejects an empty goos. `WriteEvidence(dir, report)` validates, then writes indented JSON with a trailing newline to `dir/crawler-<goos>.json`; an empty dir is an error (the caller decides whether `SNAPBACK_EVIDENCE_DIR` is set).
- **Tests:** plan.md T3 block (8 tests).

## T4 — Seeded tree builder

- **Files:** `internal/compat/crawler/seed.go` (create), `internal/compat/crawler/seed_test.go` (create).
- **Does:** `Seed(root, linkTarget, shape)` creates a tree of at least 20 directories, 3 levels deep, one regular file per directory, and a `.snapshot` symlink in every directory whose raw target is `linkTarget`. It returns the list of directories and links created. It refuses a root that is not under `os.TempDir()` (a `t.TempDir()` passes; anything under `$HOME` outside temp is refused, §20). Pure filesystem code, so unit tests run in the default build with a plain temp directory as the link target.
- **Tests:** plan.md T4 block (3 tests).

## T5 — Gated crawler integration test and full matrix (last task)

- **Files:** `internal/compat/crawler/crawler_integration_test.go` (create, `//go:build integration`), `docs/reports/stage1/crawler-darwin.json` (produced by the local run, committed by the orchestrator per sprint plan.md "macOS evidence").
- **Does:** Gate: skip naming the prerequisite unless `SNAPBACK_FUSE_TESTS=1` and a FUSE probe passes (`/dev/fuse` on Linux; the macFUSE bundle on darwin). Setup: temp history dirs with files; a `projection` catalog whose snapshot entries are symlinks to those dirs; mount it with the S2-04 adapter at a temp mountpoint with `Counter` as the Observer; unmount in `t.Cleanup` even on failure. Seed the live tree (T4) with every `.snapshot` pointing into the mount. M-002 positive control first: assert the link count equals the directory count and that `ls -L <dir>/.snapshot/` (exec argv) yields hits > 0. Then for each T2 row, as a subtest: `Counter.Reset()`, resolve the binary (missing binary: `t.Skipf` naming it, row recorded `not-tested-here`), read the version, run via `exec.CommandContext` with a per-tool timeout constant and `cmd.Dir` set to the temp tree, record hits/ops/exit code/duration. Non-zero tool exits are recorded, not failed (rg exits 1 on no match; `find -L` may report cycles). A deadline hit records `timed-out`. Assert zero hits for present non-following tools; for following tools only assert that a count was recorded. Append the VS Code row. When `SNAPBACK_EVIDENCE_DIR` is set, `WriteEvidence` produces `crawler-<goos>.json`. Finally run the whole standards matrix (M-005: last task of the story).
- **Tests:** plan.md T5 block (2 tests).
