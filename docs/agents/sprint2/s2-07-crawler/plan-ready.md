---
type: plan-ready
story: S2-07
from_red_at: 2026-09-22T03:15:15Z
---

# S2-07 plan-ready (RED verified by orchestrator at chain @ c326407; every box currently FAILs by assertion unless noted)


# S2-07 plan — tests only

Package `internal/compat/crawler` (`package crawler`). Unit tests (T1–T4) run in the default build: `go test -race ./internal/compat/crawler/...`. Each fails by assertion against the SCAFFOLD stubs, never by compile error or panic alone. Integration tests (T5) carry `//go:build integration`: `SNAPBACK_FUSE_TESTS=1 go test -race -tags=integration -run TestCrawler -v ./internal/compat/crawler/`.

## T1 — Catalog hit counter

- [x] `internal/compat/crawler/counter_test.go::TestCounterSatisfiesObserver` — input: `var _ mount.Observer = (*Counter)(nil)` plus a new `Counter`; action: deliver one lookup event through the `mount.Observer` interface value; assertion: `Total()` is 1 (compile-time and runtime conformance).
- [x] `internal/compat/crawler/counter_test.go::TestCounterCountsEachEvent` — input: new `Counter`; action: deliver 3 lookups, 2 readdirs, 1 readlink; assertion: `Total()` == 6.
- [x] `internal/compat/crawler/counter_test.go::TestCounterByOpKind` — input: the same 6 events; action: call `ByOp()`, then mutate the returned map; assertion: counts are lookup 3, readdir 2, readlink 1, and a second `ByOp()` is unchanged by the mutation (returns a copy).
- [x] `internal/compat/crawler/counter_test.go::TestCounterResetZeroes` — input: `Counter` with 5 events; action: `Reset()`, then deliver 1 readdir; assertion: `Total()` == 1 and `ByOp()` has only readdir 1 (reset between tools cannot inflate later rows).
- [x] `internal/compat/crawler/counter_test.go::TestCounterConcurrentSafe` — input: new `Counter`; action: 50 goroutines each deliver 100 events, joined with a WaitGroup, under `-race`; assertion: `Total()` == 5000 and the race detector reports nothing.

## T2 — Tool table and binary resolution

- [x] `internal/compat/crawler/tools_test.go::TestToolTableExactRows` — input: `Tools()`; action: collect display names in order; assertion: exactly `rg`, `rg -L`, `fd`, `fd -L`, `find`, `find -L`, `rsync -a` (7 rows, no extras).
- [x] `internal/compat/crawler/tools_test.go::TestToolTableFollowFlags` — input: `Tools()`; action: read `Follows` per row; assertion: true for exactly `rg -L`, `fd -L`, `find -L`; false for the other four; each following row's argv contains `-L` and each non-following row's argv does not.
- [x] `internal/compat/crawler/tools_test.go::TestToolArgsAreArraysWithRoot` — input: each row's argv for root `/tmp/x` and dest `/tmp/y`; action: inspect elements; assertion: argv is non-empty, contains the root as its own element, no element is `sh`/`bash`/`-c`, and no element contains `;`, `|`, `&&`, `$(` or a backtick.
- [x] `internal/compat/crawler/tools_test.go::TestSearchToolsVisitHiddenEntries` — input: argv for the four `rg`/`fd` rows; action: inspect flags; assertion: `rg` rows contain `--hidden` and `--no-ignore`, `fd` rows contain `-H` and `-I` (otherwise `.snapshot` is never visited and zero hits is vacuous).
- [x] `internal/compat/crawler/tools_test.go::TestRsyncArgsUseDestination` — input: `rsync -a` row, root `/tmp/x`, dest `/tmp/y`; action: build argv; assertion: argv contains `-a`, does not contain `-L`/`--copy-links`, and its last element is the destination `/tmp/y/`.
- [x] `internal/compat/crawler/tools_test.go::TestResolvePrefersFd` — input: `fd` row, fake `lookPath` that finds both `fd` and `fdfind`; action: `Resolve`; assertion: returns the `fd` path.
- [x] `internal/compat/crawler/tools_test.go::TestResolveFallsBackToFdfind` — input: `fd -L` row, fake `lookPath` that finds only `fdfind`; action: `Resolve`; assertion: returns the `fdfind` path with nil error.
- [x] `internal/compat/crawler/tools_test.go::TestResolveMissingNamesBinaries` — input: `fd` row and `rsync -a` row, fake `lookPath` that finds nothing; action: `Resolve`; assertion: `errors.Is(err, ErrToolMissing)`, the `fd` error message contains both `fd` and `fdfind`, the rsync message contains `rsync`; and `VSCodeRow()` has status `not-tested-here`, a non-empty reason mentioning `search.followSymlinks`, and no argv.

## T3 — Result schema, acceptance check and JSON writer

- [x] `internal/compat/crawler/evidence_test.go::TestEvidenceFileName` — input: goos values `darwin`, `linux`, empty; action: `EvidenceFileName`; assertion: `crawler-darwin.json`, `crawler-linux.json`, and an error for empty.
- [x] `internal/compat/crawler/evidence_test.go::TestReportJSONRoundTrip` — input: a `Report` with one tested row (hits 0), one tested following row (hits 42, per-op counts), one `not-tested-here` row, the VS Code row; action: marshal then unmarshal; assertion: the decoded value deep-equals the original and the JSON contains keys `goos`, `rows`, `hits`, `status`, `version`, `argv`.
- [x] `internal/compat/crawler/evidence_test.go::TestSkippedRowEncodesNoHitCount` — input: a `not-tested-here` row with reason `fd not on PATH`; action: marshal; assertion: the row's `hits` is `null` or absent (never `0`) and `reason` is present.
- [x] `internal/compat/crawler/evidence_test.go::TestValidateRejectsSkipFoldedIntoPass` — input: table of bad rows: `not-tested-here` with hits 0; `not-tested-here` with empty reason; `tested` with nil hits; unknown status; action: `Report.Validate()`; assertion: each returns a non-nil error, and a report of only valid rows returns nil.
- [x] `internal/compat/crawler/evidence_test.go::TestNonFollowingViolationsFlagsHits` — input: report where tested `rg` has hits 3 and tested `find` is `timed-out`; action: `NonFollowingViolations`; assertion: returns exactly those two rows by name.
- [x] `internal/compat/crawler/evidence_test.go::TestNonFollowingViolationsIgnoresFollowingAndSkipped` — input: report where `rg -L` has hits 900, `find -L` is `timed-out`, `fd` is `not-tested-here`, `rsync -a` has hits 0; action: `NonFollowingViolations`; assertion: first checks the report has 4 rows (M-002), then asserts the result is empty.
- [x] `internal/compat/crawler/evidence_test.go::TestWriteEvidenceCreatesFile` — input: `t.TempDir()`, a valid darwin report; action: `WriteEvidence`; assertion: `crawler-darwin.json` exists, parses as JSON into an equal `Report`, and ends with a newline.
- [x] `internal/compat/crawler/evidence_test.go::TestWriteEvidenceRejectsBadInput` — input: empty dir with a valid report, then `t.TempDir()` with an invalid report; action: `WriteEvidence`; assertion: both return errors and no `crawler-*.json` file exists in the temp dir afterwards.

## T4 — Seeded tree builder

- [x] `internal/compat/crawler/seed_test.go::TestSeedTreeShape` — input: `t.TempDir()` root, a temp dir as link target, default shape; action: `Seed`; assertion: at least 20 directories returned, the maximum depth below root is 3, and every directory contains at least one regular file.
- [x] `internal/compat/crawler/seed_test.go::TestSeedTreeLinkInEveryDir` — input: the same seeded tree; action: walk with `filepath.WalkDir` (no following) and `os.Readlink` each `.snapshot`; assertion: link count equals directory count (non-zero), each entry is a symlink (`Lstat` mode), and each raw target equals `linkTarget` byte-for-byte.
- [x] `internal/compat/crawler/seed_test.go::TestSeedRejectsNonTempRoot` — input: roots `$HOME/snapback-crawler-seed` and `/`; action: `Seed`; assertion: each returns an error and nothing is created at the `$HOME` path (`os.Lstat` reports not-exist).

## T5 — Gated crawler integration test

- [x] `internal/compat/crawler/crawler_integration_test.go::TestCrawlerPositiveControl` — input: gate (`SNAPBACK_FUSE_TESTS=1` + FUSE probe, else `t.Skip` naming the missing prerequisite), mounted catalog with `Counter`, seeded tree linked into the mount; action: count links, `Reset()`, run `ls -L <dir>/.snapshot/` via `exec.CommandContext`; assertion: link count equals directory count (>= 20) and `Total()` > 0 afterwards; the mountpoint is unmounted and empty after cleanup.
- [x] `internal/compat/crawler/crawler_integration_test.go::TestCrawlerToolMatrix` — input: same gated setup; action: per T2 row as a subtest, `Reset()`, `Resolve` (missing: `t.Skipf("<binary> not on PATH")` and row `not-tested-here`), read version, run with the per-tool timeout and `cmd.Dir` in the temp tree, record hits; append `VSCodeRow()`; if `SNAPBACK_EVIDENCE_DIR` is set, `WriteEvidence`; assertion: every present non-following row has hits == 0 and status `tested` (`NonFollowingViolations` empty), every present following row has non-nil hits, every missing tool is `not-tested-here` (never `tested` with 0), the VS Code row is `not-tested-here`, and when the env var is set `crawler-<goos>.json` exists and passes `Validate()`.
