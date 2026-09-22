---
type: plan
story: S2-09
scope: "tests only"
---

# S2-09 test plan (tests only)

Contracts under test: `tasks.md` (S2-09). Package `test/reports` (`package reports_test`), standard library only. Every test first calls `readReport(t)`, which fails by assertion (`t.Fatalf`) while the report is missing, so all tests are RED today. No test skips on missing input. Negative checks first assert the positive content exists (M-002): report non-empty, section present, at least one row, evidence set non-empty.

## T1 — Test helpers, report skeleton and evidence inventory

- [ ] `test/reports/inventory_test.go::TestReportHasAllSectionsInOrder` — read the report; assert it is non-empty and the nine `## ` headings from tasks.md appear exactly once each, in order.
- [ ] `test/reports/inventory_test.go::TestEvidenceInventory` — for each of the 9 expected files: present under `docs/reports/stage1/` or listed in `## Missing evidence` as implemented but not tested here with a non-empty reason; assert present + listed == 9, present >= 1, `latency.json` present, and no present file is also listed as missing.
- [ ] `test/reports/inventory_test.go::TestEvidenceFilesParse` — assert the present set is non-empty; each present file decodes to a non-empty JSON object; a malformed file fails naming the file.

## T2 — Pinned versions

- [ ] `test/reports/versions_test.go::TestGoFuseVersionMatchesGoMod` — parse `go.mod`; assert a `github.com/hanwen/go-fuse/v2` require line exists (fail if not found); assert the `go-fuse` row's value equals that version and equals every present `catalog-*.json` `go_fuse_version`.
- [ ] `test/reports/versions_test.go::TestPinnedVersionsTable` — assert rows `restic`, `rclone`, `Go toolchain` exist; restic equals `0.19.0`, every present `pathtemplate-*.json` `restic_version` and `latency.json` `restic_version`; rclone equals `latency.json` `rclone_version`; Go toolchain equals the `toolchain` directive in `go.mod`.

## T3 — Per-platform path-template, catalog and fidelity sections

- [ ] `test/reports/platform_test.go::TestPathTemplateRowsMatchEvidence` — assert at least one present pathtemplate file; for each, a row names the file with the platform label, `result` and `snapshot_id` equal to the JSON; the snapshot_id is 64 lowercase hex.
- [ ] `test/reports/platform_test.go::TestCatalogRowsMatchEvidence` — assert at least one present catalog file; for each, a row with the label, `result` and `go_fuse_version` equal to the JSON.
- [ ] `test/reports/platform_test.go::TestFidelityRowsMatchEvidence` — assert at least one present fidelity file; for each, a row with the label, `files_compared`, `mtime_precision_ns` and pass/fail equal to the JSON.
- [ ] `test/reports/platform_test.go::TestCtimeBirthTimeNotClaimed` — assert every present fidelity file has at least one `files[]` entry and every `ctime`/`birth_time` has `claimed: false`; assert the fidelity section contains the exact ctime line and no sentence pairs `ctime` with `accurate`, `exact` or `verified`.
- [ ] `test/reports/platform_test.go::TestPlatformLabels` — assert at least one darwin row exists and uses exactly `macOS (Apple Silicon, macFUSE, local host)`, linux rows use exactly `Linux (ubuntu-latest CI, fuse3)`, and the report contains no `macOS support`.

## T4 — Latency numbers

- [ ] `test/reports/latency_test.go::TestLatencyMediansMinMaxMatchJSON` — assert `latency.json` has all four measurement names; for each, the row exists and median/min/max equal `strconv.FormatFloat(json, 'f', 1, 64)`.
- [ ] `test/reports/latency_test.go::TestLatencySampleCounts` — for each measurement, assert `samples_ms` is non-empty and the row's count equals `len(samples_ms)`; assert every measurement with count 1 is one of the two cold measurements.
- [ ] `test/reports/latency_test.go::TestLatencyDataAndRemoteDeleted` — assert `latency.json` `remote_deleted` is `true` and `remote` is `gdrive:snapback-stage1`; assert the report lines `Data: <count> files, <bytes> bytes` and `Remote deleted: true` match the JSON.

## T5 — Crawler hit counts

- [ ] `test/reports/crawler_test.go::TestCrawlerRowsMatchEvidence` — assert at least one present crawler file with at least one `tested` row; for each JSON row, a report row with the platform label, tool, status and hits (`—` when null) equal to the JSON.
- [ ] `test/reports/crawler_test.go::TestCrawlerNotTestedShownHonestly` — assert each platform has a VS Code row with status `not-tested-here`; every JSON row with status `not-tested-here` appears as such with `—` hits (never `0`), and its reason appears in the section.
- [ ] `test/reports/crawler_test.go::TestCrawlerRowCountEqualsEvidence` — assert the number of crawler report rows per platform equals that file's JSON row count (no omitted tool).

## T6 — Requirement matrix, open item, honesty gate and full matrix

- [ ] `test/reports/matrix_test.go::TestMatrixCoversAllStage1Items` — assert the matrix has exactly the eight items from tasks.md, each once, each status one of the three allowed values, and a non-empty reason on every row not `implemented and tested`.
- [ ] `test/reports/matrix_test.go::TestMatrixStatusConsistentWithEvidence` — an item is `implemented and tested` only if its evidence files are present and passing (Linux catalog needs `catalog-linux.json` with `result: pass`; macOS catalog needs `catalog-darwin.json`; latency needs `remote_deleted: true`); an item whose evidence is listed missing is not `implemented and tested`.
- [ ] `test/reports/matrix_test.go::TestPendingDecisionLine` — assert `## Open items` contains exactly `Latency go/no-go: PENDING — human decision` and three option bullets mentioning `proceed`, `pre-warm` and `§23`.
- [ ] `test/reports/matrix_test.go::TestNoThresholdOrVerdict` — after asserting the latency section has four rows, assert no line outside the options list matches `(?i)threshold|acceptable|\bno-go\b|\bGO\b` or a decision phrase (`decision: go`, `recommend`), and no `ms` value appears with `<`, `>` or `under`/`below` qualifiers.
- [ ] `test/reports/matrix_test.go::TestNoBannedHonestyWords` — after asserting the report is non-empty, assert it contains none of `production-ready`, `cross-platform`, `static`, `Finder-integrated`, `Finder integrated` (case-insensitive).

Test count: 21 (T1 3, T2 2, T3 5, T4 3, T5 3, T6 5).
