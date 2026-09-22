---
type: plan
story: S2-06
scope: "tests only"
---

# S2-06 test plan (tests only)

Contracts under test come from `tasks.md`: `Compare`/`CompareAll`/`Precision`/`MTimeTolerance`, `ParseResticLs`, `Fixtures`/`Observe`, `WriteEvidence`/`ReadEvidence`/`MissingPrereq`, and the gated integration flow. Package: `github.com/adeelahmad/snapback/internal/compat/fidelity`. Unit tests run in the default build (`go test -race ./internal/compat/fidelity/`) and fail by assertion, never by a missing prerequisite. The integration test is `//go:build integration` and skips only with a message naming the missing prerequisite.

## T1 — comparator (unit)

- [ ] `internal/compat/fidelity/compare_test.go::TestCompareExactMatchPasses` — expected == observed (size 1024, mode 0644, mtime with ns); `Compare(e, o, MTimeTolerance)`; assert `SizeOK && ModeOK && MTimeOK` and `MTimeDelta == 0`.
- [ ] `internal/compat/fidelity/compare_test.go::TestCompareMTimeDriftOneSecondFails` — observed mtime = expected + 1s; assert `MTimeOK == false`, `MTimeDelta == time.Second`, size/mode still OK.
- [ ] `internal/compat/fidelity/compare_test.go::TestCompareSubSecondTruncationFailsAndIsMeasured` — expected `…:07.123456789Z`, observed truncated to `…:07Z`; tol 0; assert `MTimeOK == false` and `MTimeDelta == -123456789ns` (the loss is recorded, not hidden).
- [ ] `internal/compat/fidelity/compare_test.go::TestCompareModeDriftFails` — table: 0644 vs 0600, 0755 vs 0644, 0444 vs 0644, regular vs `fs.ModeSymlink|0777`; assert `ModeOK == false` for every row.
- [ ] `internal/compat/fidelity/compare_test.go::TestCompareSizeDriftFails` — 0 vs 1, 3145728 vs 3145727; assert `SizeOK == false`.
- [ ] `internal/compat/fidelity/compare_test.go::TestMTimeToleranceIsZero` — assert `MTimeTolerance == 0` (widening it must show up as a failing test and a diff, not happen silently).
- [ ] `internal/compat/fidelity/compare_test.go::TestCompareAllRejectsEmptyExpected` — `CompareAll(nil, map{…}, 0)` and `CompareAll([]Meta{}, …)`; assert non-nil error for both (M-002: no vacuous pass).
- [ ] `internal/compat/fidelity/compare_test.go::TestCompareAllMissingObservedErrors` — 3 expected, observed lacks `with space.txt`; assert error whose message contains `with space.txt`.
- [ ] `internal/compat/fidelity/compare_test.go::TestCompareAllOneResultPerExpectedInOrder` — 6 expected, observed has all 6 plus an extra; assert `len(results) == 6`, `results[i].Path == expected[i].Path`.
- [ ] `internal/compat/fidelity/compare_test.go::TestPrecision` — table: all ns-precise -> 1ns; all multiples of 1µs -> 1µs; of 1ms -> 1ms; whole seconds -> 1s; mixed 1s + one 1ms value -> 1ms.

## T2 — `restic ls --json` parser (unit)

- [ ] `internal/compat/fidelity/resticls_test.go::TestParseResticLsFilesAndRootStrip` — literal restic 0.19.0 ndjson: snapshot line, `dir` node, two `file` nodes under `/tmp/x/tree/…`, root `/tmp/x/tree`; assert 2 entries with relative paths, `size`, `Mode == 0644/0755` (from numeric `mode`), and `MTime` equal to the RFC 3339 ns value.
- [ ] `internal/compat/fidelity/resticls_test.go::TestParseResticLsSymlink` — a `symlink` node with `linktarget: "small.txt"` and mode `fs.ModeSymlink|0777`; assert `LinkTarget == "small.txt"` and `Mode&fs.ModeSymlink != 0`.
- [ ] `internal/compat/fidelity/resticls_test.go::TestParseResticLsSpacesAndUnicode` — nodes named `with space.txt` and `café-日本.txt` (NFC bytes); assert both paths come back byte-identical.
- [ ] `internal/compat/fidelity/resticls_test.go::TestParseResticLsRejectsMalformed` — table: truncated JSON on line 3, `mtime: "yesterday"`; assert error containing `line 3` for the first and non-nil for the second.

## T3 — fixtures and observation (unit)

- [ ] `internal/compat/fidelity/fixtures_test.go::TestFixturesCoverDistinctMetadata` — `Fixtures()`; assert >= 6 regular files, pairwise-distinct mtimes, one with non-zero `Nanosecond()`, one before 2000-01-01, sizes include 0, one in (0, 64KiB) and one >= 2MiB, modes include 0644, 0600, 0755, 0444.
- [ ] `internal/compat/fidelity/fixtures_test.go::TestFixturesIncludeSymlinkSpaceUnicode` — assert one entry with `ModeSymlink` and `LinkTarget == "small.txt"` whose target is itself a fixture, one path containing a space, one path with a non-ASCII rune that is valid UTF-8, and one path containing `/`.
- [ ] `internal/compat/fidelity/observe_test.go::TestObserveReportsSizeModeMTime` — write 1234 bytes to a `t.TempDir()` file, chmod 0600, `os.Chtimes` to `1998-03-14T09:26:53.5Z`; `Observe`; assert `Size == 1234`, `Mode.Perm() == 0600`, `MTime.Equal(set)` (asserts the observer itself does not lose precision on a local filesystem).
- [ ] `internal/compat/fidelity/observe_test.go::TestObserveSymlinkUsesLstat` — symlink `l` -> `target.txt` (target 5000 bytes); `Observe(l)`; assert `Mode&fs.ModeSymlink != 0`, `LinkTarget == "target.txt"`, `Size != 5000`.
- [ ] `internal/compat/fidelity/observe_test.go::TestObserveRecordsCTimeAndBirthPerPlatform` — `Observe` a new temp file; assert `!CTime.IsZero()`; on darwin assert `BirthTime != nil`, on linux assert `BirthTime == nil`.

## T4 — evidence and prerequisite probe (unit)

- [ ] `internal/compat/fidelity/evidence_test.go::TestEvidenceRoundTrip` — build an `Evidence` with 2 files (one failing mtime); `WriteEvidence(t.TempDir(), ev)` then `ReadEvidence`; assert deep-equal and `pass == false` preserved.
- [ ] `internal/compat/fidelity/evidence_test.go::TestWriteEvidenceFileNameAndMode` — assert returned path is `<dir>/fidelity-<runtime.GOOS>.json`, file mode 0644, content is valid indented JSON; a non-existent dir returns an error and creates nothing.
- [ ] `internal/compat/fidelity/evidence_test.go::TestEvidenceCarriesToleranceAndPrecision` — encode, decode into `map[string]any`; assert keys `mtime_tolerance_ns`, `mtime_precision_ns`, `files_generated`, `files_compared`, `resolved_inside_restic_mount`, `snapshot_id` exist, and per-file `mtime_delta_ns`, `expected`, `observed`, `restic_ls`.
- [ ] `internal/compat/fidelity/evidence_test.go::TestEvidenceMarksCTimeBirthNotClaimed` — encode a file with ctime set and birth time nil; assert `ctime.claimed == false`, `birth_time.claimed == false`, `birth_time.value == null`, and both notes contain `recorded, not claimed`; after asserting the files array is non-empty (M-002), assert no key named `ctime_ok` or `birth_time_ok` exists.
- [ ] `internal/compat/fidelity/prereq_test.go::TestMissingPrereqNamesEach` — table with fake `getenv`/`lookPath`/`exists`: env unset, restic missing, linux without `/dev/fuse`, darwin without `/Library/Filesystems/macfuse.fs`; assert the exact message for each row.
- [ ] `internal/compat/fidelity/prereq_test.go::TestMissingPrereqEmptyWhenAllPresent` — all fakes present, for `linux` and `darwin`; assert `""`.

## T5 — gated integration (restic + FUSE)

- [ ] `internal/compat/fidelity/fidelity_integration_test.go::TestFidelityThroughSnapshotAlias` — `//go:build integration`; skip with `MissingPrereq(...)`'s message if non-empty; run the full flow from tasks.md T5; assert `EvalSymlinks(<catalogMnt>/.snapshot/<name>)` has the restic mountpoint as its prefix; assert `files_compared == files_generated >= 6` plus the symlink; assert every file's size, mode and mtime equal the generator's values and the `restic ls --json` values with `MTimeTolerance`; write `fidelity-<goos>.json` to `$SNAPBACK_EVIDENCE_DIR` (when set) in a cleanup that runs before unmount, even on failure; cleanup leaves both mountpoints empty and the repo destroyed.

Test count: 26 (25 unit in the default build, 1 gated integration).
