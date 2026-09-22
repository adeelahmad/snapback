---
type: output
story: S2-07
---

## S2-07/T1 · attempt 1 · red-worker · 2026-09-22T03:13:45Z
status: ok
### Summary
Wrote the 5 plan.md T1 tests in internal/compat/crawler/counter_test.go against the merged mount.Observer/Event/Op contract (Observe(ev mount.Event); OpLookup/OpReadDir/OpReadlink). Compile shim zz_agentic_shim_t1.go (first line `// agentic:shim`) declares only T1 symbols: type Counter struct{} with Observe (no-op), Total() int (returns -1), ByOp() map[mount.Op]int (returns {lookup:-1}), Reset() (no-op). The zero-value Counter is used (no constructor). All 5 tests compile and fail by assertion under -race; go vet ./... clean; golangci-lint 0 issues. Commit 67d997a on the worktree branch; diff vs 366b1a1 = the two T1 files only.
### Result
| Check | Status | Detail |
|---|---|---|
| `internal/compat/crawler/counter_test.go::TestCounterSatisfiesObserver` | FAIL | Total() after one lookup = -1, want 1 |
| `internal/compat/crawler/counter_test.go::TestCounterCountsEachEvent` | FAIL | Total() after 6 events = -1, want 6 |
| `internal/compat/crawler/counter_test.go::TestCounterByOpKind` | FAIL | ByOp() = map[lookup:-1], want map[lookup:3 readdir:2 readlink:1] |
| `internal/compat/crawler/counter_test.go::TestCounterResetZeroes` | FAIL | Total() after Reset and 1 readdir = -1, want 1 |
| `internal/compat/crawler/counter_test.go::TestCounterConcurrentSafe` | FAIL | Total() after 50 goroutines x 100 events = -1, want 5000 |
| go vet ./... | PASS | clean |
| golangci-lint run ./internal/compat/crawler/... | PASS | 0 issues |
### Next
Scaffold Counter in internal/compat/crawler/counter.go (Observe, Total() int, ByOp() map[mount.Op]int returning a copy, Reset; zero value usable, concurrency-safe) and delete zz_agentic_shim_t1.go.

## S2-07/T2 · attempt 1 · red-worker · 2026-09-22T03:14:11Z
status: ok
### Summary
Wrote all 8 plan.md § T2 tests in `internal/compat/crawler/tools_test.go` with a single shim `zz_agentic_shim_t2.go` (first line `// agentic:shim`) declaring T2 symbols only: `Status` (type only, no constants; T3 owns the values), `ErrToolMissing`, `Tool{Name, Candidates, Follows, VersionArgs}` + method `Tool.Argv(root, dest string) []string`, `FixedRow{Name, Status, Reason, Argv}`, `Tools() []Tool`, `Resolve(tool Tool, lookPath func(string) (string, error)) (string, error)`, `VSCodeRow() FixedRow`. Shim bodies are deliberately wrong (empty table, `Resolve` returns "", nil, `VSCodeRow` returns status `tested` with argv). Per M-002, per-row checks first require all 7 rows (`allTools` t.Fatal), so they cannot pass on an empty table. The name lookups also t.Fatal when a row is missing. Design choice to note: tasks.md gives no type for the VS Code row, and T3's `Row` is not available to T2, so T2 adds its own `FixedRow`, and T3 can convert it into a `Row`. Commit e503ace; the diff against 366b1a1 touches only these 2 files. `go vet ./...` is clean, `golangci-lint run ./internal/compat/crawler/...` reports 0 issues, and gofmt is clean.
### Result
| Check | Status | Detail |
|---|---|---|
| `internal/compat/crawler/tools_test.go::TestToolTableExactRows` | FAIL | Tools() names = [], want 7 rows |
| `internal/compat/crawler/tools_test.go::TestToolTableFollowFlags` | FAIL | len(Tools()) = 0, want 7 |
| `internal/compat/crawler/tools_test.go::TestToolArgsAreArraysWithRoot` | FAIL | len(Tools()) = 0, want 7 |
| `internal/compat/crawler/tools_test.go::TestSearchToolsVisitHiddenEntries` | FAIL | Tools() has no row "rg" |
| `internal/compat/crawler/tools_test.go::TestRsyncArgsUseDestination` | FAIL | Tools() has no row "rsync -a" |
| `internal/compat/crawler/tools_test.go::TestResolvePrefersFd` | FAIL | Tools() has no row "fd" |
| `internal/compat/crawler/tools_test.go::TestResolveFallsBackToFdfind` | FAIL | Tools() has no row "fd -L" |
| `internal/compat/crawler/tools_test.go::TestResolveMissingNamesBinaries` | FAIL | Tools() has no row "fd" |
### Next
Scaffold `internal/compat/crawler/tools.go` with the named symbols (`Status`, `ErrToolMissing`, `Tool`, `Tool.Argv`, `FixedRow`, `Tools`, `Resolve`, `VSCodeRow`) and delete `zz_agentic_shim_t2.go`. T3 imports `Status` and adds its constants.

## S2-07/T4 · attempt 1 · red-worker · 2026-09-22T03:30:00Z
status: ok

### Summary
Wrote the three plan.md § T4 tests in `internal/compat/crawler/seed_test.go` against the T4-only shim `zz_agentic_shim_t4.go` (`Shape{Depth, Fanout}`, `DefaultShape() Shape`, `Seed(root, linkTarget string, shape Shape) (dirs, links []string, err error)`). The shim bodies are deliberately wrong: `Seed` creates nothing, returns nothing and never errors, and `DefaultShape` returns a zero Shape. All three tests compile and fail by assertion. go vet is clean and golangci-lint reports 0 issues. Commit c2ba9ae; the diff against 366b1a1 contains only the two files. Signature choice: the contract says "returns the list of directories and links", so `Seed` returns two slices and an error. DefaultShape is a func rather than a var so SCAFFOLD can stub it. No conflicts with the style guide.

### Result
| Check | Status | Detail |
|---|---|---|
| `internal/compat/crawler/seed_test.go::TestSeedTreeShape` | FAIL | Seed(...) returned 0 directories, want >= 20 |
| `internal/compat/crawler/seed_test.go::TestSeedTreeLinkInEveryDir` | FAIL | found 0 .snapshot links, want one per directory (M-002 non-empty guard) |
| `internal/compat/crawler/seed_test.go::TestSeedRejectsNonTempRoot` | FAIL | Seed("$HOME/snapback-crawler-seed"/"/") error = nil, want an error |

### Next
Scaffold `Shape`, `DefaultShape`, `Seed` in `internal/compat/crawler/seed.go` and delete `zz_agentic_shim_t4.go`. GREEN: the default shape must give >= 20 dirs with max depth exactly 3 below root (root counts as depth 0; e.g. Depth 3, Fanout 3 = 40 dirs), one regular file and one `.snapshot` symlink (raw target == linkTarget) per dir including root, and refuse a root outside `os.TempDir()` before creating anything.

## S2-07/T1+T2+T4 · attempt 1 · scaffolder · 2026-09-22T03:17:59Z
status: ok
### Summary
Moved the T1, T2 and T4 shim symbols into their canonical production files with the RED-chosen signatures unchanged, added the package doc comment in doc.go, and deleted the three shims. Types, Status, ErrToolMissing and struct fields keep real declarations; every function and method body is panic("SUB-AGENT-TODO: ...") carrying its tasks.md recipe. T3 files (evidence_test.go, zz_agentic_shim_t3.go) were not touched. Commit 2e01ad8 on top of c326407.
### Scaffold
+ create package doc @ internal/compat/crawler/doc.go
+ create Counter, Counter.Observe, Counter.Total, Counter.ByOp, Counter.Reset @ internal/compat/crawler/counter.go
+ create Status, ErrToolMissing, Tool, Tool.Argv, FixedRow, Tools, Resolve, VSCodeRow @ internal/compat/crawler/tools.go
+ create Shape, DefaultShape, Seed @ internal/compat/crawler/seed.go
- delete internal/compat/crawler/zz_agentic_shim_t1.go
- delete internal/compat/crawler/zz_agentic_shim_t2.go
- delete internal/compat/crawler/zz_agentic_shim_t4.go
### Result
| Check | Status | Detail |
|---|---|---|
| stubs panic+TODO | PASS | 10 function/method bodies |
| single definition | PASS | each symbol defined once in package crawler |
| go vet ./... | PASS | GOTOOLCHAIN=auto |
| gofmt -l | PASS | clean |
| shim markers T1/T2/T4 | PASS | none left |
| tests | EXPECTED FAIL | stubs panic until GREEN |
### Next
green: fill the T1 (counter.go), T2 (tools.go) and T4 (seed.go) stubs

## S2-07/T3 · attempt 1 · red-worker · 2026-09-22T03:40:00Z
status: ok

### Summary
Wrote all 8 plan.md § T3 tests in `internal/compat/crawler/evidence_test.go` with one shim `zz_agentic_shim_t3.go` (first line `// agentic:shim`) that declares only T3 symbols. These are the `Status` constants `StatusTested`/`StatusNotTestedHere`/`StatusTimedOut` (T2 owns the type and T3 owns the values), `SchemaVersion`, `Row{Tool, Argv, Follows, Status, Hits *int, Ops map[string]int, Version, ExitCode, Duration, Reason}`, `Report{SchemaVersion, GOOS, GOARCH, Seed Shape, PositiveControlHits, Rows}`, `Report.Validate() error`, `NonFollowingViolations(Report) []Row`, `EvidenceFileName(string) (string, error)` and `WriteEvidence(dir string, r Report) error`. Shim bodies are deliberately wrong: Validate and WriteEvidence return nil without writing, NonFollowingViolations returns every row, and EvidenceFileName returns "evidence.json" with no error. The struct types deliberately have no JSON tags, so the key and `reason` checks fail. The contract is snake_case tags (goos, rows, hits, status, version, argv, reason, ...), with `hits` null or omitted when nil. Per M-002, the ignore-test first asserts that the report has 4 rows. The VS Code row is converted from T2's `VSCodeRow()` FixedRow into a `Row` inside the test, so no production converter is needed. `TestWriteEvidenceRejectsBadInput` uses `t.Chdir(tempdir)` so that a wrongly accepted empty dir would leave a file the glob can see. Commit 60de721. The diff against c326407 touches only these 2 files. `go vet ./...` is clean, golangci-lint reports 0 issues, and gofmt is clean. No conflicts with the style guide.

### Result
| Check | Status | Detail |
|---|---|---|
| `internal/compat/crawler/evidence_test.go::TestEvidenceFileName` | FAIL | EvidenceFileName("darwin") = "evidence.json", <nil>, want "crawler-darwin.json", nil; empty goos gives no error |
| `internal/compat/crawler/evidence_test.go::TestReportJSONRoundTrip` | FAIL | json.Marshal(report) = ..., want key "goos" (also rows, hits, status, version, argv) |
| `internal/compat/crawler/evidence_test.go::TestSkippedRowEncodesNoHitCount` | FAIL | json.Marshal(skipped row) = ..., want reason "fd not on PATH" |
| `internal/compat/crawler/evidence_test.go::TestValidateRejectsSkipFoldedIntoPass` | FAIL | Report{...}.Validate() = nil, want error (all 4 bad cases) |
| `internal/compat/crawler/evidence_test.go::TestNonFollowingViolationsFlagsHits` | FAIL | NonFollowingViolations(report) = [rg rg -L fd find rsync -a], want [rg find] |
| `internal/compat/crawler/evidence_test.go::TestNonFollowingViolationsIgnoresFollowingAndSkipped` | FAIL | NonFollowingViolations(report) = [rg -L find -L fd rsync -a], want none |
| `internal/compat/crawler/evidence_test.go::TestWriteEvidenceCreatesFile` | FAIL | os.ReadFile(.../crawler-darwin.json): no such file or directory |
| `internal/compat/crawler/evidence_test.go::TestWriteEvidenceRejectsBadInput` | FAIL | WriteEvidence("", valid report) = nil, want error; invalid report = nil, want error |

### Next
Scaffold `internal/compat/crawler/evidence.go` with the named symbols (Status constants, `SchemaVersion`, `Row`, `Report`, `Report.Validate`, `NonFollowingViolations`, `EvidenceFileName`, `WriteEvidence`) and delete `zz_agentic_shim_t3.go`. `Report.Seed` uses T4's `Shape`. GREEN adds snake_case JSON tags (`hits` as `*int`, null or omitempty).

## S2-07/T1 · attempt 1 · green-worker · 2026-09-22T03:21:37Z
status: ok
### Summary
Filled the four SUB-AGENT-TODO stubs in `internal/compat/crawler/counter.go`. `Counter` holds a `sync.Mutex`, a total and a lazily allocated `map[mount.Op]int`, so the zero value is usable. `Observe` bumps both under the lock, `ByOp` returns `maps.Clone` of the map (never the internal one), `Reset` zeroes the total and drops the map. Commit e1a75c2 on base 7bde0d0 (chain2/s2-07, which carries the T3 RED); the diff touches only counter.go. No style or contract conflicts.
### Result
| Check | Status | Detail |
|---|---|---|
| `internal/compat/crawler/counter_test.go::TestCounterSatisfiesObserver` | PASS | -race |
| `internal/compat/crawler/counter_test.go::TestCounterCountsEachEvent` | PASS | -race |
| `internal/compat/crawler/counter_test.go::TestCounterByOpKind` | PASS | -race; copy semantics verified |
| `internal/compat/crawler/counter_test.go::TestCounterResetZeroes` | PASS | -race |
| `internal/compat/crawler/counter_test.go::TestCounterConcurrentSafe` | PASS | -race, 50x100 events |
| `go vet ./...` | PASS | clean |
| `golangci-lint run ./internal/compat/crawler/...` | PASS (counter.go) | 0 findings in counter.go. 6 staticcheck SA4006 findings exist in seed_test.go and tools_test.go: T2/T4 stubs still panic, so later assignments look unused. They are outside this scope and should clear once T2/T4 GREEN land |
| Scope | PASS | diff = internal/compat/crawler/counter.go only |
### Next
Structural review, then merge alongside the T2/T4 GREENs. Re-run package lint after all three merge to confirm the SA4006 findings clear.

## S2-07/T2 · attempt 1 · green-worker · 2026-09-22T04:10:00Z
status: ok
### Summary
Filled the four T2 stubs in `internal/compat/crawler/tools.go` (commit 1b28f2e on base 7bde0d0). `Tools()` returns the seven ordered rows. Both fd rows list the candidates `fd` then `fdfind`, and every row gets `--version` as VersionArgs. `Argv` builds plain argument arrays: `rg --files --hidden --no-ignore [-L] root`, `fd -H -I [-L] . root`, `find [-L] root` and `rsync -a root dest/`. `Resolve` returns the first candidate that lookPath finds. Otherwise it returns an error that wraps `ErrToolMissing` and names every candidate it tried. `VSCodeRow` returns status `not-tested-here` using the T3 shim constant `StatusNotTestedHere`, a reason that mentions `search.followSymlinks`, and no argv. No tests were edited and no files outside SCOPE were touched.
### Result
| Check | Status | Detail |
|---|---|---|
| `internal/compat/crawler/tools_test.go::TestToolTableExactRows` | PASS | green |
| `internal/compat/crawler/tools_test.go::TestToolTableFollowFlags` | PASS | green |
| `internal/compat/crawler/tools_test.go::TestToolArgsAreArraysWithRoot` | PASS | green |
| `internal/compat/crawler/tools_test.go::TestSearchToolsVisitHiddenEntries` | PASS | green |
| `internal/compat/crawler/tools_test.go::TestRsyncArgsUseDestination` | PASS | green |
| `internal/compat/crawler/tools_test.go::TestResolvePrefersFd` | PASS | green |
| `internal/compat/crawler/tools_test.go::TestResolveFallsBackToFdfind` | PASS | green |
| `internal/compat/crawler/tools_test.go::TestResolveMissingNamesBinaries` | PASS | green |
| `go vet ./...` | PASS | clean |
| `golangci-lint run ./internal/compat/crawler/...` | NOTE | no findings in tools.go; 5 pre-existing SA4006 (staticcheck) in RED file `seed_test.go` lines 30,31,71,72,128 (unused `root`/`linkTarget` before reassignment), outside T2 scope, needs a RED/T4 fix |
| sibling tests (T1 counter, T4 seed) | EXPECTED FAIL | sibling SUB-AGENT-TODO stubs panic; not T2 |
### Next
Merge into chain2/s2-07. Before the story's final gate, the orchestrator should route the `seed_test.go` SA4006 lint findings to the T4 RED owner. Then structural review.

## S2-07/T4 · attempt 1 · green-worker · 2026-09-22T03:25:00Z
status: ok
### Summary
Filled the two SUB-AGENT-TODO stubs in internal/compat/crawler/seed.go. DefaultShape returns Depth 3, Fanout 3 (40 dirs, max depth 3). Seed refuses any root whose absolute path is not under filepath.Abs(os.TempDir()) before touching the filesystem, then recursively creates each directory with one regular file (file.txt) and a .snapshot symlink whose raw target is linkTarget, returning the dirs and links. Commit b754df4 on base 7bde0d0; only seed.go changed.
### Result
| Check | Status | Detail |
|---|---|---|
| `internal/compat/crawler/seed_test.go::TestSeedTreeShape` | PASS | green under -race |
| `internal/compat/crawler/seed_test.go::TestSeedTreeLinkInEveryDir` | PASS | green under -race |
| `internal/compat/crawler/seed_test.go::TestSeedRejectsNonTempRoot` | PASS | green under -race |
| `go vet ./...` | PASS | clean |
| `golangci-lint run ./internal/compat/crawler/...` | PASS (seed.go) | 0 findings in seed.go; 1 pre-existing SA4006 in tools_test.go:76 (T2 RED file, out of scope) |
| Sibling tests (T1/T2/T3) | N/A | still panic on their own SUB-AGENT-TODO stubs, as expected on this chain |
### Next
Orchestrator: flag SA4006 (`banned` unused) in tools_test.go:76 to the T2 owner. Then structural review / T5.

## S2-07/T3 · attempt 1 · scaffolder · 2026-09-22T03:22:06Z
status: ok
### Summary
Moved every symbol from zz_agentic_shim_t3.go into the canonical internal/compat/crawler/evidence.go and deleted the shim. Status constants, SchemaVersion, Row and Report stay real (no JSON tags; GREEN adds them). Validate, NonFollowingViolations, EvidenceFileName and WriteEvidence panic with SUB-AGENT-TODO recipes from tasks.md T3. counter.go, tools.go and seed.go untouched. Commit 076473c on base 7bde0d0.
### Scaffold
+ create StatusTested/StatusNotTestedHere/StatusTimedOut, SchemaVersion @ internal/compat/crawler/evidence.go
+ create Row, Report (no JSON tags) @ internal/compat/crawler/evidence.go
+ create (Report).Validate stub @ internal/compat/crawler/evidence.go
+ create NonFollowingViolations stub @ internal/compat/crawler/evidence.go
+ create EvidenceFileName stub @ internal/compat/crawler/evidence.go
+ create WriteEvidence stub @ internal/compat/crawler/evidence.go
- delete internal/compat/crawler/zz_agentic_shim_t3.go
### Result
| Check | Status | Detail |
|---|---|---|
| stubs panic+TODO | PASS | 4 function stubs |
| go vet ./... | PASS | GOTOOLCHAIN=auto |
| gofmt -l | PASS | clean |
| shim marker | PASS | no agentic:shim left |
| selfcheck uniqueness | PASS | the gate matches by name only, so WriteEvidence/Validate were left out of .agentic/scaffold-symbols; resticfx.WriteEvidence (S2-03) is a separate package, not a duplicate |
### Next
green: fill the evidence.go stubs and add the snake_case JSON tags to Row and Report

## S2-07/T3 · attempt 1 · green-worker · 2026-09-22T03:35:00Z
status: ok
### Summary
Filled the four SUB-AGENT-TODO stubs in `internal/compat/crawler/evidence.go` and added snake_case JSON tags to `Row` and `Report` (`hits` is `*int` with `omitempty`, so a skipped row never encodes `0`). `Validate` rejects a not-tested-here row with hits or without a reason, a tested row with nil hits, and any unknown status; timed-out rows pass. `NonFollowingViolations` returns non-following rows that are tested with hits > 0 or timed out. `EvidenceFileName` returns `crawler-<goos>.json` and errors on empty goos. `WriteEvidence` checks dir, validates, then writes indented JSON plus a trailing newline, so rejected input creates no file. Commit 9bf61b7 on top of 43cf0aa. No test edits and no files outside SCOPE. No conflicts with the style guide.
### Result
| Check | Status | Detail |
|---|---|---|
| `internal/compat/crawler/evidence_test.go::TestEvidenceFileName` | PASS | green |
| `internal/compat/crawler/evidence_test.go::TestReportJSONRoundTrip` | PASS | green |
| `internal/compat/crawler/evidence_test.go::TestSkippedRowEncodesNoHitCount` | PASS | green |
| `internal/compat/crawler/evidence_test.go::TestValidateRejectsSkipFoldedIntoPass` | PASS | green |
| `internal/compat/crawler/evidence_test.go::TestNonFollowingViolationsFlagsHits` | PASS | green |
| `internal/compat/crawler/evidence_test.go::TestNonFollowingViolationsIgnoresFollowingAndSkipped` | PASS | green |
| `internal/compat/crawler/evidence_test.go::TestWriteEvidenceCreatesFile` | PASS | green |
| `internal/compat/crawler/evidence_test.go::TestWriteEvidenceRejectsBadInput` | PASS | green |
| `go test -race ./internal/compat/crawler/` | PASS | all 24 tests (T1, T2, T4 included) |
| `go vet ./...` | PASS | clean |
| `golangci-lint run ./internal/compat/crawler/...` | PASS | 0 issues |
### Next
Structural review, then T5 (gated crawler integration test).

## S2-07/T5 · attempt 1 · red-worker · 2026-09-22T03:28:56Z
status: escalate
### Summary
I wrote both plan.md § T5 tests in `internal/compat/crawler/crawler_integration_test.go` (`//go:build integration`). No shim was needed because every symbol already exists: T1-T4 are implemented and the S2-04 `gofuse` adapter is merged. Setup builds 3 temp history dirs, a `projection` catalog that links to them, and a mount at `<tmp>/snapback-crawler-mnt` with `Counter` as the Observer. It seeds the live tree (DefaultShape, 40 dirs) with every `.snapshot` pointing into the mount. Unmount is registered in `t.Cleanup`, so it runs even when a test fails. The gate skips with a message naming the missing prerequisite (`SNAPBACK_FUSE_TESTS`, `/dev/fuse`, or the macFUSE bundle). I ran it for real on this Mac with macFUSE. Both tests PASS on first run, because the feature under test was already built by T1-T4 and S2-04. This is PASS-ON-RED per M-003, and the tests are not weakened. The positive control proves the counter sees reads (see below), so the zero-hit rows mean something (M-002). Measured hits: rg/fd/find/rsync -a = 0; rg -L = 761; fd -L = 482; find -L = 400. Positive control = 4. The VS Code row is not-tested-here. No mount was left behind (`mount | grep -c snapback-crawler` = 0). The evidence JSON went to the scratchpad only and was not committed.
### Result
| Check | Status | Detail |
|---|---|---|
| `internal/compat/crawler/crawler_integration_test.go::TestCrawlerPositiveControl` | PASS-ON-RED | T1/T4 + S2-04 adapter already implemented; links == dirs (40 >= 20), `ls -L` Total() = 4 > 0, mountpoint empty and on the parent device after Unmount |
| `internal/compat/crawler/crawler_integration_test.go::TestCrawlerToolMatrix` | PASS-ON-RED | T2/T3 already implemented; all 7 tools present (subtests PASS), non-following hits 0 / status tested, following hits recorded, VS Code not-tested-here, crawler-darwin.json written and Validate() nil |
| gate unset (`SNAPBACK_FUSE_TESTS=`) | PASS | 2 SKIP lines naming SNAPBACK_FUSE_TESTS, final ok |
| validate.md jq checks (non-following zero, following numeric) | PASS | both `true` on scratchpad evidence |
| `go vet ./...` + `go vet -tags=integration` | PASS | clean |
| `golangci-lint run --build-tags=integration ./internal/compat/crawler/...` | PASS | 0 issues |
| stale mount | PASS | 0 |
### Next
selfcheck gate-red-verify BLOCKS by design (no FAIL row). I did not fabricate a FAIL, because both tests genuinely pass (M-003). The orchestrator must adjudicate. Orchestrator decides what to do with PASS-ON-RED (M-003). There are no stubs to scaffold and no GREEN bodies left. Next steps: the final sign-off matrix (validate.md § Final sign-off), then committing `docs/reports/stage1/crawler-darwin.json` from a post-merge run.

## S2-07/T5 · attempt 1 · green-worker · 2026-09-22T03:33:07Z
status: ok
### Summary
RED was PASS-ON-RED, so no production change was needed. Ran the real macFUSE crawler integration (`SNAPBACK_FUSE_TESTS=1 SNAPBACK_EVIDENCE_DIR=<worktree>/docs/reports/stage1 go test -race -count=1 -tags=integration -run TestCrawler ./internal/compat/crawler/`) on darwin/arm64; both tests passed and all 7 tool rows (rg, rg -L, fd, fd -L, find, find -L, rsync -a) were `tested`. Non-following tools recorded 0 catalog hits; following variants (-L) traverse the catalog (e.g. rg -L 761 hits), which is the recorded evidence. No mount left behind (`mount | grep fuse` empty after run). Committed only `docs/reports/stage1/crawler-darwin.json` as 1d9d114. The full Cross-cutting gates matrix is green; the generated `site/` and `coverage.out` were removed afterwards.
### Result
| Check | Status | Detail |
|---|---|---|
| `internal/compat/crawler/crawler_integration_test.go::TestCrawlerPositiveControl` | PASS | macFUSE, positive_control_hits=4 |
| `internal/compat/crawler/crawler_integration_test.go::TestCrawlerToolMatrix` | PASS | 7/7 subtests, all rows status=tested |
| evidence | PASS | docs/reports/stage1/crawler-darwin.json (155 lines), only file in commit |
| leftover mount | PASS | none |
| gofmt / goimports | PASS | no output |
| CGO_ENABLED=0 go build ./... | PASS | |
| go vet ./... | PASS | |
| golangci-lint run | PASS | 0 issues |
| go test -race ./... | PASS | 14 packages ok |
| coverage | PASS | 87.6% total (>=80%) |
| govulncheck ./... | PASS | 0 affecting vulnerabilities |
| actionlint | PASS | |
| shellcheck -s sh install.sh | PASS | |
| mkdocs build --strict | PASS | |
| goreleaser check | PASS | |
### Next
S2-07 complete; structural review of chain/s2-07, then merge to stage-1.
