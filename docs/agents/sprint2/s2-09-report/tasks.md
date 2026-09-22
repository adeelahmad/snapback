---
type: tasks
story: S2-09
---

# S2-09 tasks — Stage 1 measurements report and requirement matrix

Story intent and owned files: `docs/agents/sprint2/stories.md` (S2-09). Owned set: `docs/reports/stage1-measurements.md`, `test/reports/**`. Nothing else is touched. The evidence JSON under `docs/reports/stage1/` is read, never written (owners: S2-03, S2-04, S2-06, S2-07, S2-08). Tests: `plan.md`. Gates: `validate.md`.

Decisions fixed here so workers do not guess:

- Package `test/reports`, declared `package reports_test`, test files only (same shape as `test/ci`). Standard library only: no new `go.mod` requirement (go.mod is S2-01's).
- No production code exists in this story, so RED needs no shims. Every test starts with `readReport(t)`, which calls `t.Fatalf` when `docs/reports/stage1-measurements.md` is missing or empty. That is how every test fails by assertion before GREEN. No test uses `t.Skip` for a missing report or a missing evidence file.
- Anti-vacuity (M-002): every negative check (banned words, no threshold, no verdict) first asserts the report is non-empty, the named section exists, and the section has at least one table row. Every cross-check first asserts the evidence set it iterates is non-empty.
- Evidence schemas are the ones fixed in the owning stories:
  - `pathtemplate-<goos>.json` (S2-03 T6): `goos`, `goarch`, `restic_version`, `path_template`, `snapshot_id`, `observed_dir`, `ids_entries`, `result` (`pass`|`fail`), `reason`, `generated_at`.
  - `catalog-<goos>.json` (S2-04 T4): `platform` (`GOOS/GOARCH`), `go_fuse_version`, `operations`, `result`, `timestamp`.
  - `fidelity-<goos>.json` (S2-06 T4): `platform`, `restic_version`, `snapshot_id`, `snapshot_alias`, `resolved_inside_restic_mount`, `mtime_tolerance_ns`, `mtime_precision_ns`, `files_generated`, `files_compared`, `pass`, `files[]` (with `ctime`/`birth_time` objects carrying `claimed: false`).
  - `crawler-<goos>.json` (S2-07 T3): `goos`, `goarch`, `rows[]` with `tool`, `status` (`tested`|`not-tested-here`|`timed-out`), `hits` (int or null), `reason`.
  - `latency.json` (S2-08 T2): `measurements` (four names, each `samples_ms`, `median_ms`, `min_ms`, `max_ms`), `restic_version`, `rclone_version`, `os`, `arch`, `timestamp_utc`, `remote`, `data_file_count`, `data_file_bytes`, `remote_deleted`.
  Tests decode into `map[string]any` and look keys up by name, so a schema drift fails with the missing key named, not a zero value.
- Expected evidence set (9 files): `pathtemplate-{darwin,linux}.json`, `catalog-{darwin,linux}.json`, `fidelity-{darwin,linux}.json`, `crawler-{darwin,linux}.json`, `latency.json`. `latency.json` is mandatory (the go/no-go needs its numbers). Any other file may be absent only if the report's `## Missing evidence` section lists it as `` - `<file>`: implemented but not tested here — <reason> `` with a non-empty reason.
- Machine-checkable report format (the workers write exactly this; the tests parse it):
  - Section headings, in order: `## Pinned versions`, `## Path-template verification`, `## Catalog mount, browse and unmount`, `## Metadata fidelity`, `## Latency over rclone:gdrive`, `## Crawler hit counts`, `## Requirement matrix`, `## Missing evidence`, `## Open items`.
  - Tables are pipe tables. A helper `tableRows(section)` returns the data rows (header and `---` rows dropped) as trimmed cells.
  - Pinned versions rows: `| go-fuse | <go.mod version> |`, `| restic | <version> |`, `| rclone | <version> |`, `| Go toolchain | <go.mod toolchain> |`.
  - Platform labels: darwin is exactly `macOS (Apple Silicon, macFUSE, local host)`; linux is exactly `Linux (ubuntu-latest CI, fuse3)`.
  - Path-template rows: `| <file> | <label> | <result> | <snapshot_id> |`. Catalog rows: `| <file> | <label> | <result> | <go_fuse_version> |`. Fidelity rows: `| <file> | <label> | <files_compared> | <mtime_precision_ns> | <pass|fail> |`, plus the line `ctime and birth time: FUSE-approximated; recorded, not claimed.`
  - Latency rows: `| <measurement name> | <median_ms> | <min_ms> | <max_ms> | <sample count> |`, every ms value formatted as `strconv.FormatFloat(v, 'f', 1, 64)` of the JSON value, sample count = `len(samples_ms)`. Then the lines `Data: <data_file_count> files, <data_file_bytes> bytes` and `Remote deleted: <remote_deleted>`.
  - Crawler rows: `| <label> | <tool> | <status> | <hits or —> |`, one per JSON row, plus a VS Code row with status `not-tested-here` per platform (S2-07 appends it to the JSON).
  - Matrix rows: `| <item> | <status> | <reason> |`, status one of `implemented and tested`, `implemented but not tested here`, `not implemented`. The eight §22 row 1 items, exact text: `Pin dependencies`, `Disposable Restic repo`, `Verify --path-template ids/%I`, `Tiny directory/symlink FUSE catalog on Linux`, `Tiny directory/symlink FUSE catalog on macOS`, `Metadata-fidelity check`, `rclone/Google Drive latency measurements`, `Crawler test`.
  - Open items contains exactly the line `Latency go/no-go: PENDING — human decision`, followed by a bullet list of the three options: proceed to Stage 2; deeper pre-warm; pull the §23 local hot-cache repository forward. No threshold and no recommendation.
- Every number is copied from JSON. The worker generates the tables by reading the JSON files (e.g. with `jq`), never by typing figures. No helper program is committed (owned set is the report and the tests).
- All six tasks edit the same report file, so they run strictly in order. The story starts only after every evidence file is committed (sprint plan wave 4).

## T1 — Test helpers, report skeleton and evidence inventory

Helpers (`repoRoot`, `readReport`, `section`, `tableRows`, `loadEvidence`, `presentEvidence`, `missingListed`) and the inventory tests. Report: create the file with the nine headings, the intro, and `## Missing evidence` filled from what is actually absent.

Files: `test/reports/helpers_test.go`, `test/reports/inventory_test.go`, `docs/reports/stage1-measurements.md`.

## T2 — Pinned versions

Fill `## Pinned versions` from `go.mod` (go-fuse require line, toolchain directive), `pathtemplate-*.json` and `latency.json` (restic, rclone).

Files: `test/reports/versions_test.go`, `docs/reports/stage1-measurements.md`.

## T3 — Per-platform path-template, catalog and fidelity sections

Fill the three per-platform sections from their JSON, with the fixed platform labels and the ctime/birth-time line.

Files: `test/reports/platform_test.go`, `docs/reports/stage1-measurements.md`.

## T4 — Latency numbers

Fill `## Latency over rclone:gdrive` from `latency.json`: four rows, data size, `Remote deleted`.

Files: `test/reports/latency_test.go`, `docs/reports/stage1-measurements.md`.

## T5 — Crawler hit counts

Fill `## Crawler hit counts` from `crawler-*.json`, including every `not-tested-here` row and its reason, and the VS Code row.

Files: `test/reports/crawler_test.go`, `docs/reports/stage1-measurements.md`.

## T6 — Requirement matrix, open item, honesty gate and full matrix

Fill `## Requirement matrix` (eight rows, statuses consistent with the evidence) and `## Open items` (pending line and the three options). This is the story's last task: its GREEN runs the full standards matrix (M-005).

Files: `test/reports/matrix_test.go`, `docs/reports/stage1-measurements.md`.
