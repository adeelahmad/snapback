---
type: output
story: S2-05
---

## S2-05/T1 · attempt 1 · red-worker · 2026-09-22T02:15:31Z
status: ok
### Summary
Created `test/ci/fuse_job_test.go` with the 8 T1 tests from plan.md § T1, the `fuseJobBlock(t)` helper (calls `t.Fatalf` when the `fuse-linux` job block is empty, per M-002), and two small helpers: `aptInstallTokens` and `hasToken`. The file reuses `readCI`, `jobBlock`, `stepContaining` and `assertContainsAll` unchanged. No shim was needed because every symbol used already exists, and the 22 existing tests are untouched. The 8 new tests fail with `t.Fatalf("ci.yml has no \"fuse-linux\" job")`, because no such job exists yet. Package count is now 30 (22 + 8). As a sanity check, I temporarily appended the job exactly as tasks.md § T1 describes. With it, all 30 tests passed. I then reverted ci.yml, so the diff against BASE_REF is only `test/ci/fuse_job_test.go`. `go vet ./...` is clean and `golangci-lint run ./test/ci/...` reports 0 issues. Commit: c1aaa8d.
### Result
| Check | Status | Detail |
|---|---|---|
| `test/ci/fuse_job_test.go::TestFuseJobExistsOnUbuntuLatest` | FAIL | ci.yml has no "fuse-linux" job |
| `test/ci/fuse_job_test.go::TestFuseJobCheckoutAndSetupGoPinned` | FAIL | ci.yml has no "fuse-linux" job |
| `test/ci/fuse_job_test.go::TestFuseJobInstallsFuse3` | FAIL | ci.yml has no "fuse-linux" job |
| `test/ci/fuse_job_test.go::TestFuseJobInstallsCrawlerTools` | FAIL | ci.yml has no "fuse-linux" job |
| `test/ci/fuse_job_test.go::TestFuseJobDoesNotInstallDistroRestic` | FAIL | ci.yml has no "fuse-linux" job |
| `test/ci/fuse_job_test.go::TestFuseJobDownloadsPinnedRestic` | FAIL | ci.yml has no "fuse-linux" job |
| `test/ci/fuse_job_test.go::TestFuseJobVerifiesResticChecksum` | FAIL | ci.yml has no "fuse-linux" job |
| `test/ci/fuse_job_test.go::TestFuseJobAssertsResticVersion` | FAIL | ci.yml has no "fuse-linux" job |
| existing 22 test/ci tests | PASS | unchanged |
| `go vet ./...` | PASS | clean |
| `golangci-lint run ./test/ci/...` | PASS | 0 issues |
### Next
SCAFFOLD is a no-op because there are no production symbols. GREEN appends the `fuse-linux` job to `.github/workflows/ci.yml` after `cross-compile`, as tasks.md § T1 specifies. In `TestFuseJobAssertsResticVersion`, only text after `restic version` in the step counts, so the restic URL cannot satisfy the `0.19.0` check.

## S2-05/T2 · attempt 1 · red-worker · 2026-09-22T02:18:33Z
status: ok
### Summary
I appended the 8 T2 tests from plan.md § T2 to `test/ci/fuse_job_test.go`. They reuse `fuseJobBlock`, `stepContaining`, `assertContainsAll`, `readCI` and `pinnedRef`. The file also gains two small unshimmed helpers, `firstLineIndex` and `isGoTestLine`. No shim was needed because every symbol exists. There are no production edits: ci.yml and the existing tests are untouched. At base c1aaa8d, ci.yml has no `fuse-linux` job, so every T2 test fails via the `fuseJobBlock` t.Fatal. None fails at compile time. Commit 4c8bd48. The package now lists 38 tests (22 existing + 16 fuse), matching validate.md.
### Result
| Check | Status | Detail |
|---|---|---|
| `test/ci/fuse_job_test.go::TestFuseJobSetsFuseTestsEnv` | FAIL | ci.yml has no "fuse-linux" job |
| `test/ci/fuse_job_test.go::TestFuseJobSetsEvidenceDir` | FAIL | ci.yml has no "fuse-linux" job |
| `test/ci/fuse_job_test.go::TestFuseJobChecksDevFuseBeforeTests` | FAIL | ci.yml has no "fuse-linux" job |
| `test/ci/fuse_job_test.go::TestFuseJobRunsIntegrationTestsWithRace` | FAIL | ci.yml has no "fuse-linux" job |
| `test/ci/fuse_job_test.go::TestFuseJobUploadsEvidenceAlways` | FAIL | ci.yml has no "fuse-linux" job |
| `test/ci/fuse_job_test.go::TestFuseJobNeverSetsRcloneRemote` | FAIL | ci.yml has no "fuse-linux" job |
| `test/ci/fuse_job_test.go::TestFuseJobNoLatestAndActionsPinned` | FAIL | ci.yml has no "fuse-linux" job |
| `test/ci/fuse_job_test.go::TestFuseJobHasNoSecrets` | FAIL | ci.yml has no "fuse-linux" job |
| `go vet ./...` | PASS | clean |
| `golangci-lint run ./test/ci/...` | PASS | 0 issues |
| `go test -list` count | PASS | 38 |
### Next
SCAFFOLD is a no-op because there are no production symbols. GREEN, stacked on T1 GREEN, adds the job-level `env:` (`SNAPBACK_FUSE_TESTS: "1"`, `SNAPBACK_EVIDENCE_DIR`), the `/dev/fuse` check + `mkdir -p "$SNAPBACK_EVIDENCE_DIR"` step, `go test -race -tags=integration ./...`, and the `actions/upload-artifact@v4` step (`if: always()`, `if-no-files-found: warn`). The `/dev/fuse` check regex is `(test -[ce]|\[ -[ce]).*/dev/fuse` and matches a single line. The upload `path:` line must name `SNAPBACK_EVIDENCE_DIR` literally, e.g. `${{ env.SNAPBACK_EVIDENCE_DIR }}`.

## S2-05/T1 · attempt 1 · green-worker · 2026-09-22T02:21:19Z
status: ok
### Summary
I added the `fuse-linux` job at the end of `jobs:` in `.github/workflows/ci.yml`, after `cross-compile`, and changed no other file. It uses `actions/checkout@v4` and `actions/setup-go@v5` with `go-version-file: go.mod`. It installs `fuse3 ripgrep fd-find rsync` with apt (`apt-get update` first) and installs no distro restic. It downloads restic 0.19.0 from the pinned literal URL and checks it with `sha256sum -c` against `13176fe6d89d4357947a2cd107218ab2873a5f9d8e1ac2d4cd1c8e07e6839c21` before `bunzip2`. It then installs the binary to /usr/local/bin and runs `restic version | grep -F 'restic 0.19.0'`. No step name contains mount, browse or unmount. The job has no secrets, no `continue-on-error` and no `latest`. Commit 5e3fbea is on top of 4c8bd48.
### Result
| Check | Status | Detail |
|---|---|---|
| `test/ci::TestFuseJobExistsOnUbuntuLatest` | PASS | T1 |
| `test/ci::TestFuseJobCheckoutAndSetupGoPinned` | PASS | T1 |
| `test/ci::TestFuseJobInstallsFuse3` | PASS | T1 |
| `test/ci::TestFuseJobInstallsCrawlerTools` | PASS | T1 |
| `test/ci::TestFuseJobDoesNotInstallDistroRestic` | PASS | T1 |
| `test/ci::TestFuseJobDownloadsPinnedRestic` | PASS | T1 |
| `test/ci::TestFuseJobVerifiesResticChecksum` | PASS | T1 |
| `test/ci::TestFuseJobAssertsResticVersion` | PASS | T1 |
| 22 pre-existing tests (ci_test, lint_test, matrix_test) | PASS | unchanged |
| `TestFuseJobNeverSetsRcloneRemote`, `TestFuseJobNoLatestAndActionsPinned`, `TestFuseJobHasNoSecrets` | PASS | T2 guards, already satisfied |
| `TestFuseJobSetsFuseTestsEnv`, `TestFuseJobSetsEvidenceDir`, `TestFuseJobChecksDevFuseBeforeTests`, `TestFuseJobRunsIntegrationTestsWithRace`, `TestFuseJobUploadsEvidenceAlways` | FAIL | T2 scope, expected |
| `actionlint .github/workflows/ci.yml` | PASS | clean |
### Next
T2 green-worker: add the job-level env, the /dev/fuse check, the integration test run and the evidence upload. Then run the full standards matrix.

## S2-05/T2 · attempt 1 · green-worker · 2026-09-22T02:23:22Z
status: ok
### Summary
I completed the fuse-linux job in `.github/workflows/ci.yml`, the only file changed (+19 lines). The job now sets two job-level env values: `SNAPBACK_FUSE_TESTS: "1"` and `SNAPBACK_EVIDENCE_DIR: ${{ github.workspace }}/stage1-evidence`. A new step runs a one-line `test -c /dev/fuse` check. The integration step runs `mkdir -p "$SNAPBACK_EVIDENCE_DIR"` and then `go test -race -tags=integration ./...`. An always-run `actions/upload-artifact@v4` step uploads `stage1-evidence-linux` from `path: ${{ env.SNAPBACK_EVIDENCE_DIR }}` with `if-no-files-found: warn`. No step name contains mount, browse or unmount. The change adds no secrets, no continue-on-error and no test edits. Commit 1b75c3a is on top of 5e3fbea.
### Result
| Check | Status | Detail |
|---|---|---|
| `test/ci/fuse_job_test.go::TestFuseJobSetsFuseTestsEnv` | PASS | green |
| `test/ci/fuse_job_test.go::TestFuseJobSetsEvidenceDir` | PASS | green |
| `test/ci/fuse_job_test.go::TestFuseJobChecksDevFuseBeforeTests` | PASS | green |
| `test/ci/fuse_job_test.go::TestFuseJobRunsIntegrationTestsWithRace` | PASS | green |
| `test/ci/fuse_job_test.go::TestFuseJobUploadsEvidenceAlways` | PASS | green |
| `test/ci/fuse_job_test.go::TestFuseJobNeverSetsRcloneRemote / NoLatestAndActionsPinned / HasNoSecrets` | PASS | still green |
| `GOTOOLCHAIN=auto go test -race ./test/ci/ -v` | PASS | 38/38 |
| Cross-cutting gates (gofmt, goimports, build, vet, golangci-lint 0 issues, test -race, coverage 87.5%, govulncheck, actionlint, shellcheck, mkdocs --strict, goreleaser check) | PASS | full matrix |
### Next
Structural review of S2-05, then final gate.

## S2-05/T3 · attempt 1 · red-worker · 2026-09-22T03:40:00Z
status: ok
### Summary
Added `TestFuseJobUploadFailsOnMissingEvidence` to test/ci/fuse_job_test.go (commit f552e5d on base aa0974f = chain2/s2-05). It takes the fuse-linux `actions/upload-artifact` step via the existing `stepContaining`, fails with t.Fatal if the step or its `if-no-files-found:` key is missing (M-002), and checks that the value is exactly `error`. No shim was needed because every helper already exists (`fuseJobBlock`, `stepContaining`, `noFilesFoundRe`). go vet is clean and golangci-lint reports 0 issues. The other 16 tests in the package still pass.
### Result
| Check | Status | Detail |
|---|---|---|
| `test/ci/fuse_job_test.go::TestFuseJobUploadFailsOnMissingEvidence` | FAIL | fails by assertion: `upload step if-no-files-found = "warn", want "error"` |
### Next
GREEN (wave 4, after S2-03/S2-04/S2-06/S2-07 merge): change `if-no-files-found: warn` to `error` in the fuse-linux "Upload evidence" step of ci.yml.

## S2-05/T3 · attempt 1 · green-worker · 2026-09-22T03:45:00Z
status: ok
### Summary
I made one change on base f552e5d (chain2/s2-05): in the fuse-linux "Upload evidence" step of `.github/workflows/ci.yml`, `if-no-files-found: warn` is now `if-no-files-found: error`. This commit is 1e62a07. No other file changed and no test was edited. Because S2-03/04/06/07 are merged, the linux job now writes the evidence JSON. If that evidence is missing, the job now fails instead of passing with a warning.
### Result
| Check | Status | Detail |
|---|---|---|
| `test/ci/fuse_job_test.go::TestFuseJobUploadFailsOnMissingEvidence` | PASS | green |
| `test/ci/...` (all 39 tests) | PASS | go test -race |
| gofmt / goimports | PASS | no output |
| CGO_ENABLED=0 go build ./... | PASS | |
| go vet ./... | PASS | |
| golangci-lint run | PASS | 0 issues |
| go test -race ./... + coverage | PASS | total 82.1% (>=80%) |
| govulncheck ./... | PASS | no called vulnerabilities |
| actionlint | PASS | clean |
| shellcheck -s sh install.sh | PASS | |
| mkdocs build --strict | PASS | |
| goreleaser check | PASS | 1 config validated |
### Next
Structural review of S2-05, then final gate.
