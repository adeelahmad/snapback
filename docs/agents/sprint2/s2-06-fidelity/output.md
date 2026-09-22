---
type: output
story: S2-06
---

## S2-06/T1 · attempt 1 · red-worker · 2026-09-22T03:20:00Z
status: ok
### Summary
Wrote all 10 plan.md § T1 comparator tests in `internal/compat/fidelity/compare_test.go` (table-driven, got/want messages, no assertion libraries). T1 symbols (`Meta`, `Result`, `Compare`, `CompareAll`, `Precision`, `MTimeTolerance`) live only in the marked shim `zz_agentic_shim_t1.go` with deliberately wrong bodies: `Compare` reports every field OK with delta -1ns, `CompareAll` returns (nil, nil), `Precision` returns 0, `MTimeTolerance` is 1s. Because of those bodies, no test can pass by accident. The package compiles, every test fails by assertion, `go vet ./...` is clean, and `golangci-lint run ./internal/compat/fidelity/...` reports 0 issues. Commit 8168a5d; the diff against 366b1a1 is exactly the 2 files. The shim also carries the package doc comment, so GREEN's `compare.go` should take it over when the scaffolder deletes the shim.
### Result
| Check | Status | Detail |
|---|---|---|
| `internal/compat/fidelity/compare_test.go::TestCompareExactMatchPasses` | FAIL | MTimeDelta = -1ns, want 0 |
| `internal/compat/fidelity/compare_test.go::TestCompareMTimeDriftOneSecondFails` | FAIL | MTimeOK = true, want false; MTimeDelta = -1ns, want 1s |
| `internal/compat/fidelity/compare_test.go::TestCompareSubSecondTruncationFailsAndIsMeasured` | FAIL | MTimeOK = true; MTimeDelta = -1ns, want -123.456789ms |
| `internal/compat/fidelity/compare_test.go::TestCompareModeDriftFails` | FAIL | ModeOK = true, want false (all 4 rows) |
| `internal/compat/fidelity/compare_test.go::TestCompareSizeDriftFails` | FAIL | SizeOK = true, want false (both rows) |
| `internal/compat/fidelity/compare_test.go::TestMTimeToleranceIsZero` | FAIL | MTimeTolerance = 1s, want 0 |
| `internal/compat/fidelity/compare_test.go::TestCompareAllRejectsEmptyExpected` | FAIL | nil and empty expected both return a nil error |
| `internal/compat/fidelity/compare_test.go::TestCompareAllMissingObservedErrors` | FAIL | error = nil, want one containing "with space.txt" |
| `internal/compat/fidelity/compare_test.go::TestCompareAllOneResultPerExpectedInOrder` | FAIL | len = 0, want 6 |
| `internal/compat/fidelity/compare_test.go::TestPrecision` | FAIL | returns 0s for all 5 rows (want 1ns/1µs/1ms/1s/1ms) |
| `go vet ./...` | PASS | clean |
| `golangci-lint run ./internal/compat/fidelity/...` | PASS | 0 issues |
### Next
Scaffold `internal/compat/fidelity/compare.go` with `Meta`, `Result`, `Compare`, `CompareAll`, `Precision` and `const MTimeTolerance time.Duration = 0`, then delete `zz_agentic_shim_t1.go`. T3 and T4 depend on `Meta` and `Result`.

## S2-06/T2 · attempt 1 · red-worker · 2026-09-22T03:20:00Z
status: ok
### Summary
Wrote the four plan.md § T2 tests in `internal/compat/fidelity/resticls_test.go`. They use literal restic 0.19.0 ndjson (snapshot line, dir nodes, file and symlink nodes, numeric Go `fs.FileMode` modes, RFC 3339 ns mtimes). The shim `zz_agentic_shim_t2.go` declares only `ParseResticLs`, and its body returns `nil, nil`, so the tests fail by assertion. `Meta` is T1's and I did not declare it. That means this branch compiles only once it is merged with the T1 RED shim. To check that the tests compile and fail, I added a temporary uncommitted `Meta` that matches the tasks.md T1 contract, ran the tests, and deleted the file. `go vet ./...` was clean and golangci-lint reported 0 issues. The commit is 398449b, and the diff against 366b1a1 is 2 files only.
### Result
| Check | Status | Detail |
|---|---|---|
| `internal/compat/fidelity/resticls_test.go::TestParseResticLsFilesAndRootStrip` | FAIL | ParseResticLs() returned 0 entries ([]), want 2 |
| `internal/compat/fidelity/resticls_test.go::TestParseResticLsSymlink` | FAIL | ParseResticLs() returned 0 entries ([]), want 1 |
| `internal/compat/fidelity/resticls_test.go::TestParseResticLsSpacesAndUnicode` | FAIL | ParseResticLs() returned 0 entries ([]), want 2 |
| `internal/compat/fidelity/resticls_test.go::TestParseResticLsRejectsMalformed` | FAIL | both subtests: error = nil, want non-nil |
| go vet / golangci-lint | PASS | clean / 0 issues (verified with temporary Meta stand-in) |
### Next
Merge after T1 RED, which supplies `Meta`. Then scaffold `ParseResticLs` in `resticls.go` and delete `zz_agentic_shim_t2.go`.

## S2-06/T1+T2 · attempt 1 · scaffolder · 2026-09-22T03:18:10Z
status: ok
### Summary
Moved the T1 shim symbols into `compare.go` and the T2 shim symbol into `resticls.go`, moved the package doc comment to `doc.go`, and deleted both `zz_agentic_shim_t*.go` files. `Meta`, `Result` and `const MTimeTolerance time.Duration = 0` keep real declarations; `Compare`, `CompareAll`, `Precision` and `ParseResticLs` are `panic("SUB-AGENT-TODO: ...")` with the tasks.md recipe. Each symbol is defined once. Commit 4ec21dc.
### Scaffold
+ create package doc @ internal/compat/fidelity/doc.go
+ create const MTimeTolerance (= 0) @ internal/compat/fidelity/compare.go
+ create type Meta @ internal/compat/fidelity/compare.go
+ create type Result @ internal/compat/fidelity/compare.go
+ create func Compare (stub) @ internal/compat/fidelity/compare.go
+ create func CompareAll (stub) @ internal/compat/fidelity/compare.go
+ create func Precision (stub) @ internal/compat/fidelity/compare.go
+ create func ParseResticLs (stub) @ internal/compat/fidelity/resticls.go
- delete internal/compat/fidelity/zz_agentic_shim_t1.go
- delete internal/compat/fidelity/zz_agentic_shim_t2.go
### Result
| Check | Status | Detail |
|---|---|---|
| stubs panic+TODO | PASS | 4 funcs; 3 real type/const decls |
| single definition | PASS | each symbol once, in fidelity package only |
| MTimeTolerance exact | PASS | `const MTimeTolerance time.Duration = 0` |
| shim markers | PASS | no `agentic:shim` left |
| go vet ./... | PASS | clean |
| gofmt -l | PASS | clean |
| tests | FAIL (expected) | fidelity tests panic on SUB-AGENT-TODO |
### Next
green: fill Compare, CompareAll, Precision (T1) and ParseResticLs (T2).

## S2-06/T3 · attempt 1 · red-worker · 2026-09-22T03:18:01Z
status: ok
### Summary
I wrote all 5 T3 bullets from plan.md at the exact path::fn, in `fixtures_test.go` and `observe_test.go`. The only shim is `zz_agentic_shim_t3.go`. It declares `Observed` (embeds the T1 `Meta` and adds `CTime` and `BirthTime *time.Time`), `Fixtures()` returning one bogus entry, and `Observe()` returning a zero value with Size -1. One untagged shim is enough because it is wrong on every platform. The per-platform split (`observe_darwin.go`/`observe_linux.go`) is left for the scaffolder or GREEN. Every test fails by assertion on darwin. Commit f985130 on top of 6c43d9e contains only these 3 files and does not touch the T1/T2 shims or tests. `observe_test.go` never calls `os.Stat(`, because the validate.md T3 grep `observe*.go` also matches that test file.
### Result
| Check | Status | Detail |
|---|---|---|
| `internal/compat/fidelity/fixtures_test.go::TestFixturesCoverDistinctMetadata` | FAIL | Fixtures() has 1 regular files, want >= 6 |
| `internal/compat/fidelity/fixtures_test.go::TestFixturesIncludeSymlinkSpaceUnicode` | FAIL | no symlink->small.txt, space path, non-ASCII path, nested path |
| `internal/compat/fidelity/observe_test.go::TestObserveReportsSizeModeMTime` | FAIL | Size = -1, want 1234; Mode.Perm() = 0, want 0600; MTime zero |
| `internal/compat/fidelity/observe_test.go::TestObserveSymlinkUsesLstat` | FAIL | Mode lacks ModeSymlink; LinkTarget = "", want "target.txt" |
| `internal/compat/fidelity/observe_test.go::TestObserveRecordsCTimeAndBirthPerPlatform` | FAIL | CTime is zero; BirthTime = nil on darwin (on linux, only the CTime assertion fails) |
| go vet (darwin + GOOS=linux) / golangci-lint | PASS | clean / 0 issues |
### Next
Scaffold `Fixtures` in `fixtures.go`, `Observed`/`Observe` in `observe.go`, and the platform ctime/birth helpers in `observe_darwin.go`/`observe_linux.go`, then delete `zz_agentic_shim_t3.go`.

## S2-06/T4 · attempt 1 · red-worker · 2026-09-22T03:18:47Z
status: ok
### Summary
Wrote all 6 plan.md § T4 tests in `internal/compat/fidelity/evidence_test.go` and `prereq_test.go`, compiling against a single shim `zz_agentic_shim_t4.go` (Evidence, FileEvidence, Attrs, Unclaimed, WriteEvidence, ReadEvidence, MissingPrereq) with deliberately wrong bodies. Every test fails by assertion or t.Fatal on the shim's WriteEvidence error. Commit 30505c2 on top of 6c43d9e; diff = these 3 files only. Design choices the scaffolder and GREEN must keep: the JSON field names come from tasks.md; ctime and birth time use `Unclaimed{Value *time.Time}`, whose MarshalJSON always emits `{value, claimed:false, note}` (a nil Value encodes as null). The Go type therefore cannot set claimed=true, which is how SPEC §5 is enforced. It needs a matching UnmarshalJSON so the round trip is deep-equal. The JSON shape tests encode through WriteEvidence, not json.Marshal, so they cannot pass against a shim. The banned-key check builds `"ctime" + "_ok"` by concatenation so that validate.md's `grep ctime_ok|birth_time_ok` stays meaningful for production code. The `*Ns` field names follow the Go style guide (ns is a unit, not an initialism).
### Result
| Check | Status | Detail |
|---|---|---|
| `internal/compat/fidelity/evidence_test.go::TestEvidenceRoundTrip` | FAIL | WriteEvidence() error: shim: not implemented |
| `internal/compat/fidelity/evidence_test.go::TestWriteEvidenceFileNameAndMode` | FAIL | WriteEvidence(dir) error: shim: not implemented |
| `internal/compat/fidelity/evidence_test.go::TestEvidenceCarriesToleranceAndPrecision` | FAIL | WriteEvidence() error: shim: not implemented |
| `internal/compat/fidelity/evidence_test.go::TestEvidenceMarksCTimeBirthNotClaimed` | FAIL | WriteEvidence() error: shim: not implemented (shim Unclaimed also emits claimed:true, empty note) |
| `internal/compat/fidelity/prereq_test.go::TestMissingPrereqNamesEach` | FAIL | 4 subtests: MissingPrereq(...) = "shim", want exact skip message |
| `internal/compat/fidelity/prereq_test.go::TestMissingPrereqEmptyWhenAllPresent` | FAIL | MissingPrereq(all present, linux/darwin) = "shim", want "" |
| go vet ./... / golangci-lint ./internal/compat/fidelity/... | PASS | clean / 0 issues |
| validate.md ctime/birth grep | PASS | no output |
### Next
Scaffold `evidence.go` (Evidence, FileEvidence, Attrs, Unclaimed + MarshalJSON/UnmarshalJSON, WriteEvidence, ReadEvidence) and `prereq.go` (MissingPrereq) from `zz_agentic_shim_t4.go`, then delete the shim.

## S2-06/T1+T2 · attempt 1 · green-worker · 2026-09-22T03:22:32Z
status: ok
### Summary
Filled the T1 and T2 SUB-AGENT-TODO stubs, touching only compare.go and resticls.go. Compare masks the mode with ModeType|ModePerm, sets MTimeDelta to observed minus expected, and treats mtime as OK when |delta| <= tol. CompareAll returns an error when expected is empty (M-002) or when an expected path is missing from observed, with the path in the message; otherwise it returns results in input order. Precision returns the coarsest of 1s, 1ms and 1us that divides every Nanosecond(), falling back to 1ns. ParseResticLs reads the NDJSON line by line with bufio.Scanner. It keeps only struct_type=node entries of type file or symlink, strips the root prefix, decodes the mode as a numeric fs.FileMode and parses mtime as RFC3339Nano. Errors on bad JSON or a bad mtime name the line number. MTimeTolerance is unchanged at 0. Commit fd1bb5f on top of 1369260.
### Result
| Check | Status | Detail |
|---|---|---|
| `internal/compat/fidelity/compare_test.go::TestCompareExactMatchPasses` | PASS | green |
| `internal/compat/fidelity/compare_test.go::TestCompareMTimeDriftOneSecondFails` | PASS | green |
| `internal/compat/fidelity/compare_test.go::TestCompareSubSecondTruncationFailsAndIsMeasured` | PASS | green |
| `internal/compat/fidelity/compare_test.go::TestCompareModeDriftFails` | PASS | green |
| `internal/compat/fidelity/compare_test.go::TestCompareSizeDriftFails` | PASS | green |
| `internal/compat/fidelity/compare_test.go::TestMTimeToleranceIsZero` | PASS | green |
| `internal/compat/fidelity/compare_test.go::TestCompareAllRejectsEmptyExpected` | PASS | green |
| `internal/compat/fidelity/compare_test.go::TestCompareAllMissingObservedErrors` | PASS | green |
| `internal/compat/fidelity/compare_test.go::TestCompareAllOneResultPerExpectedInOrder` | PASS | green |
| `internal/compat/fidelity/compare_test.go::TestPrecision` | PASS | green |
| `internal/compat/fidelity/resticls_test.go::TestParseResticLsFilesAndRootStrip` | PASS | green |
| `internal/compat/fidelity/resticls_test.go::TestParseResticLsSymlink` | PASS | green |
| `internal/compat/fidelity/resticls_test.go::TestParseResticLsSpacesAndUnicode` | PASS | green |
| `internal/compat/fidelity/resticls_test.go::TestParseResticLsRejectsMalformed` | PASS | green |
| go vet ./... | PASS | clean |
| golangci-lint run ./internal/compat/fidelity/... | PASS | 0 issues |
| T3/T4 tests (fixtures, observe, evidence, prereq) | EXPECTED-RED | still on shims; later tasks |
| diff scope | PASS | only compare.go, resticls.go |
### Next
Structural review of T1+T2, then T3 GREEN (fixtures/observe) and T4 GREEN (evidence/prereq).

## S2-06/T3+T4 · attempt 1 · scaffolder · 2026-09-22T03:22:26Z
status: ok
### Summary
Moved the T3 and T4 compile shims into canonical files and deleted both shims (commit c002200 on top of 1369260). Types (Observed, Evidence, FileEvidence, Attrs, Unclaimed{Value *time.Time}) are kept exactly as declared, including JSON tags and field names such as MTimeToleranceNs. Every function and method body is now panic("SUB-AGENT-TODO: ...") carrying the tasks.md T3/T4 recipe. Per-platform ctime/birth-time extraction is one unexported helper, statTimes(fs.FileInfo), declared once per build-tagged file (darwin, linux). No os.Stat( in observe*.go; no ctime_ok/birth_time_ok in production code. compare.go and resticls.go untouched.
### Scaffold
+ create Fixtures @ internal/compat/fidelity/fixtures.go
+ create Observed, Observe @ internal/compat/fidelity/observe.go
+ create statTimes (darwin) @ internal/compat/fidelity/observe_darwin.go
+ create statTimes (linux) @ internal/compat/fidelity/observe_linux.go
+ create Evidence, FileEvidence, Attrs, Unclaimed, Unclaimed.MarshalJSON, Unclaimed.UnmarshalJSON, WriteEvidence, ReadEvidence @ internal/compat/fidelity/evidence.go
+ create MissingPrereq @ internal/compat/fidelity/prereq.go
- delete internal/compat/fidelity/zz_agentic_shim_t3.go
- delete internal/compat/fidelity/zz_agentic_shim_t4.go
### Result
| Check | Status | Detail |
|---|---|---|
| stubs panic+TODO | PASS | 9 function/method stubs (statTimes counted per platform) |
| each symbol defined once | PASS | statTimes once per GOOS via build tags |
| go vet ./... (darwin) | PASS | clean |
| GOOS=linux go vet ./internal/compat/fidelity/ | PASS | clean |
| gofmt -l internal/ | PASS | no output |
| shim markers / os.Stat( / ctime_ok / birth_time_ok in production | PASS | no matches |
| selfcheck gate-scaffold-verify | PASS | .agentic/scaffold-symbols = Fixtures, FileEvidence, Unclaimed, MissingPrereq. Left out: Observe/Observed (substring matches), statTimes (once per GOOS), WriteEvidence/ReadEvidence (resticfx has same-named funcs from S2-03; different package and record type, which tasks.md requires) |
### Next
green: fill the T3 stubs (Fixtures, Observe, statTimes darwin/linux) and T4 stubs (Unclaimed JSON, WriteEvidence, ReadEvidence, MissingPrereq); Observe must call statTimes so it is not unused.

## S2-06/T3 · attempt 1 · green-worker · 2026-09-22T03:40:00Z
status: ok
### Summary
Filled the four T3 stubs. `Fixtures()` returns 8 regular files (distinct mtimes incl. 2022-01-02T03:04:07.123456789Z and 1998-03-14T09:26:53Z; sizes 0, 1024, 3 MiB; modes 0644/0600/0755/0444; "with space.txt", NFC "café-日本.txt", "sub/nested.txt") plus the relative symlink link-to-small -> small.txt. `Observe` uses os.Lstat and os.Readlink for symlinks, then the per-platform `statTimes`: darwin sets CTime from Ctimespec and BirthTime from Birthtimespec; linux sets CTime from Ctim and leaves BirthTime nil. No os.Stat in observe*.go. Commit 1fb32d1 on top of 53c16ae, diff limited to SCOPE.
### Result
| Check | Status | Detail |
|---|---|---|
| `internal/compat/fidelity/fixtures_test.go::TestFixturesCoverDistinctMetadata` | PASS | green -race |
| `internal/compat/fidelity/fixtures_test.go::TestFixturesIncludeSymlinkSpaceUnicode` | PASS | green -race |
| `internal/compat/fidelity/observe_test.go::TestObserveReportsSizeModeMTime` | PASS | green -race |
| `internal/compat/fidelity/observe_test.go::TestObserveSymlinkUsesLstat` | PASS | green -race |
| `internal/compat/fidelity/observe_test.go::TestObserveRecordsCTimeAndBirthPerPlatform` | PASS | green -race (darwin) |
| T1/T2 compare + resticls tests | PASS | still green |
| `GOOS=linux go vet ./internal/compat/fidelity/` + `go vet ./...` | PASS | clean |
| `golangci-lint run ./internal/compat/fidelity/...` (darwin + GOOS=linux) | PASS | 0 issues |
### Next
Merge with T4 (evidence.go/prereq.go) on chain2/s2-06; structural review.

## S2-06/T4 · attempt 1 · green-worker · 2026-09-22T03:25:58Z
status: ok
### Summary
Filled the T4 stubs in evidence.go and prereq.go. Unclaimed marshals {value, claimed:false, note} via a private wire struct; value is null when nil. The note is "FUSE-approximated; recorded, not claimed" (from tasks.md §5), or "not exposed; recorded, not claimed" when nil. Conflict noted: the stub recipe said "recorded-not-asserted", but the test and tasks.md require "recorded, not claimed", so the contract wins. WriteEvidence stats dir first, so nothing is created when it is missing. It writes indented JSON to fidelity-<GOOS>.json with mode 0644. ReadEvidence wraps errors with the path. MissingPrereq checks prerequisites in the contract order. Production code has no ctime_ok or birth_time_ok. Commit 0278d76 on worktree-agent-a46ad7d5752cf1815.
### Result
| Check | Status | Detail |
|---|---|---|
| `internal/compat/fidelity/evidence_test.go::TestEvidenceRoundTrip` | PASS | green (-race) |
| `internal/compat/fidelity/evidence_test.go::TestWriteEvidenceFileNameAndMode` | PASS | green (-race) |
| `internal/compat/fidelity/evidence_test.go::TestEvidenceCarriesToleranceAndPrecision` | PASS | green (-race) |
| `internal/compat/fidelity/evidence_test.go::TestEvidenceMarksCTimeBirthNotClaimed` | PASS | green (-race) |
| `internal/compat/fidelity/prereq_test.go::TestMissingPrereqNamesEach` | PASS | 4 subtests green |
| `internal/compat/fidelity/prereq_test.go::TestMissingPrereqEmptyWhenAllPresent` | PASS | green |
| previously passing (TestCompare/TestMTimeToleranceIsZero/TestPrecision/TestParseResticLs) | PASS | ok |
| go vet ./... | PASS | clean |
| golangci-lint run ./internal/compat/fidelity/... | PASS | 0 issues |
| diff scope | PASS | only evidence.go, prereq.go |
### Next
Merge alongside T3 (fixtures/observe), then structural review.

## S2-06/T5 · attempt 1 · red-worker · 2026-09-22T03:31:00Z
status: ok

### Summary
Wrote `TestFidelityThroughSnapshotAlias` (`//go:build integration`) in `internal/compat/fidelity/fidelity_integration_test.go`; no shim needed (T1-T4, S2-03, S2-04 symbols all exist). Default build skips it; `SNAPBACK_FUSE_TESTS=` gives `--- SKIP` with `SNAPBACK_FUSE_TESTS not set: fidelity tests skipped`. Real run on darwin/arm64 (restic 0.19.0, macFUSE) FAILS by assertion, only on the symlink versus `restic ls --json`: restic 0.19 `ls --json` omits `size` and `linktarget` for symlink nodes (verified with a raw probe: the symlink line has no `size`/`linktarget` keys), so the ls view says size 0 / target "" while the mount serves size 9 / target "small.txt". All 8 regular files match both the generator values and `ls --json` in size, mode and mtime with `MTimeTolerance = 0` (delta 0ns each); the symlink's mode and mtime match. `resolved_inside_restic_mount` true, files_generated = files_compared = 9, mtime_precision_ns 1. Evidence was written before unmount (to a scratchpad dir, not committed) with `pass: false`. Both mountpoints were empty after cleanup, the repo was destroyed, and `mount | grep -c "$TMPDIR"` = 0. Implementation note: the temp root is resolved with EvalSymlinks first so restic's recorded paths and the EvalSymlinks prefix check agree on macOS (/var -> /private/var). The symlink is created with os.Symlink because resticfx.WriteTree writes regular files only.

### Result
| Check | Status | Detail |
|---|---|---|
| `internal/compat/fidelity/fidelity_integration_test.go::TestFidelityThroughSnapshotAlias` | FAIL | `symlink link-to-small target: observed "small.txt", restic ls "", want "small.txt"`; `link-to-small vs restic ls --json: size 9 want 0 (ok=false), mode Lrwxr-xr-x want Lrwxr-xr-x (ok=true), mtime ... delta 0s (ok=true)` |
| default build (`go test -race ./internal/compat/fidelity/`) | PASS | integration test excluded |
| skip path (`SNAPBACK_FUSE_TESTS=`) | SKIP | `SNAPBACK_FUSE_TESTS not set: fidelity tests skipped` |
| leaked mounts | PASS | 0 |
| `go vet ./...`, `golangci-lint run --build-tags=integration ./internal/compat/fidelity/...` | PASS | 0 issues |

### Next
Orchestrator/planner adjudication needed before GREEN: this failure cannot be fixed in production code because restic 0.19 `ls --json` does not report symlink size or link target. Options: (a) ParseResticLs/contract keeps symlinks out of the size and link-target comparison against `ls --json` (compare mode and mtime only, and check the link target against the generator), or (b) take the symlink's ls-side values from another restic source. Do not widen MTimeTolerance; the regular-file fidelity is exact.

## S2-06/T5b · attempt 1 · red-worker · 2026-09-22T03:33:13Z
status: ok

### Summary
Amended `TestFidelityThroughSnapshotAlias` in `internal/compat/fidelity/fidelity_integration_test.go` per the orchestrator's T5b contract ruling (plan.md/tasks.md carry no T5b section; the init.md ruling is the contract). Symlink entries are now compared against `restic ls --json` on mode and mtime only (tolerance stays 0, new helper `checkSymlinkVsLs`). The link target seen through the catalog alias must equal the generator's target (`small.txt`). Symlink size is not compared against ls, and the symlink is no longer fed to the ls-side `CompareAll`. Regular files are unchanged: size, mode and mtime are checked against both the generator and ls. In evidence, the symlink's `size_ok` stays false (not compared, so not claimed as a pass). Evidence has no field for this and the brief allows only this file to change, so the test post-processes the written JSON (`annotateUnreported`). It adds `"restic_ls_unreported": {"fields": ["size","linktarget"], "note": "not reported by restic ls --json 0.19"}` to each symlink entry. No shim was needed. Real run on darwin/arm64 with restic 0.19.0 and macFUSE: PASS (PASS-ON-RED, expected and waived). Cleanup left no FUSE mount; the only remaining mount under /private/var/folders is an unrelated ProxMobo nullfs. The repo was destroyed. Evidence was written to the scratchpad and not committed. Commit 5675db1.

### Result
| Check | Status | Detail |
|---|---|---|
| `internal/compat/fidelity/fidelity_integration_test.go::TestFidelityThroughSnapshotAlias` | PASS-ON-RED | contract amendment; the only T5 failure was the symlink ls size/linktarget, which the ruling removes from the comparison |
| evidence header | observed | platform darwin/arm64, restic 0.19.0, alias .snapshot/2026-09-22_0332Z, resolved_inside_restic_mount true, tolerance 0, precision 1ns, generated 9 = compared 9, pass true |
| empty.txt | observed | size 0/0/0 (obs/gen/ls), mode 0644, size/mode/mtime ok, delta 0ns |
| small.txt | observed | size 1024/1024/1024, mode 0644, ok ok ok, delta 0ns |
| large.bin | observed | size 3145728/3145728/3145728, mode 0600, ok ok ok, delta 0ns |
| run.sh | observed | size 64/64/64, mode 0755, ok ok ok, delta 0ns |
| readonly.txt | observed | size 128/128/128, mode 0444, ok ok ok, delta 0ns |
| with space.txt | observed | size 32/32/32, mode 0644, ok ok ok, delta 0ns |
| café-日本.txt | observed | size 48/48/48, mode 0644, ok ok ok, delta 0ns |
| sub/nested.txt | observed | size 16/16/16, mode 0644, ok ok ok, delta 0ns |
| link-to-small | observed | size 9/0/0 (not compared, size_ok false), mode Lrwxr-xr-x ok, mtime ok, delta 0ns, target through alias "small.txt", restic_ls_unreported [size, linktarget] |
| default build `go test -race ./internal/compat/fidelity/` | PASS | integration test excluded |
| skip path (`SNAPBACK_FUSE_TESTS=`) | SKIP | `SNAPBACK_FUSE_TESTS not set: fidelity tests skipped` |
| leaked FUSE mounts | PASS | 0 |
| `go vet ./...`, `go vet -tags=integration`, `golangci-lint run --build-tags=integration ./internal/compat/fidelity/...` | PASS | 0 issues |

### Next
Orchestrator waives PASS-ON-RED. No scaffold needed. GREEN has no production change for T5b. Optionally, a later story could add a first-class Evidence field for unreported ls attributes in place of the test-side JSON annotation.

## S2-06/T5 · attempt 1 · green-worker · 2026-09-22T03:35:49Z
status: ok
### Summary
No production change was needed; RED's T5 test plus the T5b contract ruling already pass against the existing code. I ran the real restic + macFUSE integration (`SNAPBACK_FUSE_TESTS=1 SNAPBACK_EVIDENCE_DIR=<worktree>/docs/reports/stage1 go test -race -count=1 -tags=integration -run TestFidelityThroughSnapshotAlias ./internal/compat/fidelity/`). It PASSED in 6.93s and wrote `docs/reports/stage1/fidelity-darwin.json` (308 lines: per-file expected/observed/restic_ls mode, size and mtime, with mtime_delta_ns 0; ctime and birth time are recorded as FUSE approximations with claimed=false). Afterwards `mount | grep -i fuse` was empty, so no mount was left behind. The JSON contains no local paths or secrets. Commit 229a1a6 `feat: record darwin fidelity evidence (S2-06 T5)` on top of 5675db1 contains only that file and has no AI trailers. This is the story's LAST task, so the full Cross-cutting gates matrix ran and is green.
### Result
| Check | Status | Detail |
|---|---|---|
| `internal/compat/fidelity::TestFidelityThroughSnapshotAlias` (integration, real restic + macFUSE) | PASS | 6.93s, evidence written, no leftover mount |
| gofmt -l / goimports -l | PASS | no output |
| CGO_ENABLED=0 go build ./... | PASS | |
| go vet ./... | PASS | |
| golangci-lint run | PASS | 0 issues |
| go test -race ./... (coverage run) | PASS | all 14 packages ok |
| coverage >= 80% | PASS | total 87.1% |
| govulncheck ./... | PASS | no called vulnerabilities |
| actionlint | PASS | |
| shellcheck -s sh install.sh | PASS | |
| mkdocs build --strict | PASS | site dir sent to scratchpad to keep the tree clean |
| goreleaser check | PASS | 1 config validated (gomips/goarm info notes only) |
| diff within SCOPE_GLOBS | PASS | only docs/reports/stage1/fidelity-darwin.json |
### Next
Structural review of S2-06, then final gate / merge of chain2/s2-06.

## S2-06/fix-gate · attempt 1 · red-worker · 2026-09-22T04:02:49Z
status: ok
### Summary
Added two rows to the TestMissingPrereqNamesEach table in internal/compat/fidelity/prereq_test.go: SNAPBACK_FUSE_TESTS="0" and "yes" must return the same skip message as unset. They fail by assertion against the current prereq.go, which treats any non-empty value as enabled. No shim needed. plan.md, tasks.md and validate.md have no fix-gate section, so the init.md brief served as the contract. Commit 0a4e9d5 on the worktree branch; go vet and golangci-lint are clean.
### Result
| Check | Status | Detail |
|---|---|---|
| `internal/compat/fidelity/prereq_test.go::TestMissingPrereqNamesEach/env_set_to_0` | FAIL | MissingPrereq(env set to 0, "linux") = "", want "SNAPBACK_FUSE_TESTS not set: fidelity tests skipped" |
| `internal/compat/fidelity/prereq_test.go::TestMissingPrereqNamesEach/env_set_to_yes` | FAIL | MissingPrereq(env set to yes, "linux") = "", want "SNAPBACK_FUSE_TESTS not set: fidelity tests skipped" |
| existing prereq rows + TestMissingPrereqEmptyWhenAllPresent | PASS | unchanged, still green |
| go vet ./... / golangci-lint | PASS | clean |
### Next
GREEN: change the check in prereq.go to getenv("SNAPBACK_FUSE_TESTS") != "1", matching resticfx.MissingPrerequisite. No new symbols to scaffold.

## S2-06/fix-gate · attempt 1 · green-worker · 2026-09-22T04:05:00Z
status: ok
### Summary
One-line fix in internal/compat/fidelity/prereq.go: MissingPrereq now skips unless SNAPBACK_FUSE_TESTS is exactly "1" (was: skip only when empty), keeping the same skip message. Values such as "0" or "yes" now skip, matching the new TestMissingPrereqNamesEach rows. Committed as 4a8a6be on worktree-agent-a133e9f7ff62410f1 (base 0a4e9d5). No tests edited; diff is within SCOPE_GLOBS.
### Result
| Check | Status | Detail |
|---|---|---|
| `internal/compat/fidelity::TestMissingPrereqNamesEach` | PASS | new 0/yes rows green |
| `internal/compat/fidelity::TestMissingPrereqEmptyWhenAllPresent` | PASS | green |
| `go test -race ./internal/compat/fidelity/...` | PASS | all 25 tests green |
| gofmt / goimports | PASS | no output |
| `CGO_ENABLED=0 go build ./...` | PASS | |
| `go vet ./...` | PASS | |
| `golangci-lint run` | PASS | 0 issues |
| `go test -race ./...` | PASS | all packages ok |
| coverage | PASS | total 84.0% (>=80%) |
| `govulncheck ./...` | PASS | 0 called vulnerabilities (1 in required module, not reachable) |
| actionlint | PASS | |
| `shellcheck -s sh install.sh` | PASS | |
| `mkdocs build --strict` | PASS | |
| `goreleaser check` | PASS | |
### Next
Structural review, then merge fix-gate into chain/stage-1.
