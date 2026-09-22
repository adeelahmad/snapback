---
type: validate
story: S2-05
---

# S2-05 validate — PASS/FAIL rubric

All commands run from `/Users/adeelahmad/work/snapback` (or the task worktree). PASS requires every row's expected output. Any other output is FAIL.

## Pre-flight

| Check | Command | Expected | FAIL cites |
| --- | --- | --- | --- |
| No story dependency | none (S2-05 is wave 1 with no hard prerequisite) | n/a | sprint2/plan.md row S2-05 |
| Existing tests green on base | `go test -count=1 ./test/ci/` | `ok  	github.com/adeelahmad/snapback/test/ci` | stories.md S2-05 Constraints (22 S1-02 tests keep passing) |
| Existing test count | `go test -list '.*' ./test/ci/ \| grep -c '^Test'` | `22` | stories.md S2-05 Constraints |
| actionlint installed | `command -v actionlint` | a path (currently `/opt/homebrew/bin/actionlint`), so `TestCIActionlint` runs rather than skips | memory.md M-007; sprint2/standards.md gate matrix (`actionlint` INSTALLED) |
| Restic checksum is the official one | `curl -fsSL https://github.com/restic/restic/releases/download/v0.19.0/SHA256SUMS \| grep ' restic_0.19.0_linux_amd64.bz2$'` | `13176fe6d89d4357947a2cd107218ab2873a5f9d8e1ac2d4cd1c8e07e6839c21  restic_0.19.0_linux_amd64.bz2` (if it differs, stop and report; do not edit the constant silently) | stories.md S2-05 Constraints (pinned URL and checksum literals) |
| Only owned files change | `git diff --name-only master... \| grep -vE '^(\.github/workflows/ci\.yml\|test/ci/fuse_job_test\.go)$'` | no output | stories.md S2-05 Owned files; memory.md M-012 |

## T1 — Job skeleton and pinned tool installs

| Check | Command | Expected | FAIL cites |
| --- | --- | --- | --- |
| T1 tests | `go test -race -count=1 -run 'TestFuseJob(ExistsOnUbuntuLatest\|CheckoutAndSetupGoPinned\|InstallsFuse3\|InstallsCrawlerTools\|DoesNotInstallDistroRestic\|DownloadsPinnedRestic\|VerifiesResticChecksum\|AssertsResticVersion)$' -v ./test/ci/` | 8 `--- PASS` lines, final `ok` | stories.md S2-05 Success scenarios; Failure scenarios (distro restic, `fd` package name) |
| Existing 22 still pass | `go test -race -count=1 -run 'TestCI\|TestMatrix\|TestGolangci' ./test/ci/` | `ok` (`TestCIActionlint` runs, not skipped) | stories.md S2-05 Constraints |
| Package-scoped only | run with `GATE_RUN_MATRIX=0`; the full matrix is T2's job | n/a | memory.md M-005 |

## T2 — Gated integration run and evidence artifact

| Check | Command | Expected | FAIL cites |
| --- | --- | --- | --- |
| T2 tests | `go test -race -count=1 -run 'TestFuseJob(SetsFuseTestsEnv\|SetsEvidenceDir\|ChecksDevFuseBeforeTests\|RunsIntegrationTestsWithRace\|UploadsEvidenceAlways\|NeverSetsRcloneRemote\|NoLatestAndActionsPinned\|HasNoSecrets)$' -v ./test/ci/` | 8 `--- PASS` lines, final `ok` | stories.md S2-05 Success scenarios; Failure scenarios (env never set, upload only on success) |
| actionlint | `actionlint .github/workflows/ci.yml` | exit 0, no output | memory.md M-007; stories.md S2-05 Constraints |
| No @latest | `grep -c '@latest' .github/workflows/ci.yml` | `0` | memory.md M-006 |
| Step names avoid mount words | `go test -count=1 -run TestCINoMountJobs -v ./test/ci/` | `--- PASS: TestCINoMountJobs` | S1-02 test (read-only this sprint) |

## Final sign-off

| Check | Command | Expected | FAIL cites |
| --- | --- | --- | --- |
| Whole package | `go test -race -count=1 ./test/ci/...` | `ok  	github.com/adeelahmad/snapback/test/ci` | common/testing.md |
| Test count | `go test -list '.*' ./test/ci/ \| grep -c '^Test'` | `38` after T2 (22 existing plus 16 new); `39` after T3 | this plan.md |
| No skips in test/ci | `go test -race -count=1 -v ./test/ci/ \| grep -c -- '--- SKIP'` | `0` (actionlint is installed, so `TestCIActionlint` must run) | sprint2/standards.md (zero suppressions) |
| Existing test files untouched | `git diff --name-only master... -- test/ci/ci_test.go test/ci/matrix_test.go test/ci/lint_test.go test/ci/helpers_test.go` | no output | sprint2/plan.md (S1-02 tests read-only) |
| actionlint | `actionlint` | exit 0, no output | sprint2/standards.md gate matrix |
| Full standards matrix (T2 only, last task) | every command in the fenced block of sprint2/standards.md "Cross-cutting gates", in order | every command exits 0 and coverage is at least 80% | memory.md M-005; sprint2/standards.md |
| Owned files only | `git diff --stat master...` | exactly `.github/workflows/ci.yml` and `test/ci/fuse_job_test.go` | memory.md M-012 |
| **Exit evidence (orchestrator, post-merge; not a unit test)** | `gh run list --workflow ci.yml --branch master --limit 1 --json conclusion,databaseId`, then `gh run download <id> -n stage1-evidence-linux` | `fuse-linux` concluded `success`, and the artifact holds the Linux JSON files (`pathtemplate-linux.json`, `catalog-linux.json`, `fidelity-linux.json`, `crawler-linux.json` as those stories land) | stories.md S2-05 Exit evidence; memory.md M-007 |

The last row is verified by the orchestrator after merge. It cannot be a unit test: a green run on GitHub needs a real runner, a real `/dev/fuse` and the merged commit. Local tests and actionlint only prove the YAML's shape. They do not prove that apt package names, the restic download or the kernel mount work on `ubuntu-latest` (M-007). S2-05's tasks are complete once the rows above that row pass. The story is complete only when the orchestrator records this row, which can happen after S2-03, S2-04, S2-06 and S2-07 merge.

**Resolved:** T2 ships `if-no-files-found: warn`, so the wave-1 merge (before any evidence writer exists) stays green. T3 (final wave, after S2-03, S2-04, S2-06 and S2-07 are on `master`) flips it to `error`, so a run where every integration test silently skips goes red in CI.

## T3 — Evidence upload becomes strict (final wave)

| Check | Command | Expected | FAIL cites |
| --- | --- | --- | --- |
| Dependencies merged | `ls internal/compat/resticfx internal/mount/gofuse internal/compat/fidelity internal/compat/crawler` | all four exist on `master` | sprint2/plan.md wave 4 |
| T3 test | `go test -race -count=1 -run 'TestFuseJobUploadFailsOnMissingEvidence' -v ./test/ci/` | 1 `--- PASS`, final `ok` | plan.md T3 |
| Whole package after T3 | `go test -race -count=1 ./test/ci/...` and `go test -list '.*' ./test/ci/ \| grep -c '^Test'` | `ok`; `39` | this plan.md |
| actionlint | `actionlint .github/workflows/ci.yml` | exit 0, no output | memory.md M-007 |
| Real CI | next `fuse-linux` run on `master` | `success` with all four `*-linux.json` in the artifact | stories.md S2-05 Exit evidence |

