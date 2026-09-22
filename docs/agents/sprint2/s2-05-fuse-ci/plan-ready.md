---
type: plan-ready
story: S2-05
from_red_at: 2026-09-22T02:19:09Z
---

# S2-05 plan-ready (RED verified by orchestrator at chain @ 4c8bd48; every box currently FAILs by assertion unless noted)


# S2-05 plan — tests only

Package `test/ci` (`package ci_test`), stdlib only, text-based checks over `.github/workflows/ci.yml`. New file `test/ci/fuse_job_test.go`, 16 tests. It reuses the existing `readCI`, `jobBlock`, `stepsContaining`, `stepContaining`, `indentOf` and `pinnedRef` unchanged. It adds one local helper, `fuseJobBlock(t)`, which returns `jobBlock(readCI(t), "fuse-linux")` and calls `t.Fatalf` if that is empty. Every "must not contain" check calls `fuseJobBlock` first, so it cannot pass on an empty block (M-002). The 22 existing tests keep running unmodified, so the package total is 38. Run: `go test -race ./test/ci/...`.

## T1 — Job skeleton and pinned tool installs

- [x] `test/ci/fuse_job_test.go::TestFuseJobExistsOnUbuntuLatest` — input: ci.yml text; action: `fuseJobBlock(t)`; assertion: the block is non-empty and contains a `runs-on: ubuntu-latest` line.
- [x] `test/ci/fuse_job_test.go::TestFuseJobCheckoutAndSetupGoPinned` — input: fuse-linux block; action: substring and regex search; assertion: it contains `actions/checkout@v4` and `actions/setup-go@v5`, and the setup-go step (via `stepContaining` on the block) matches `go-version-file:\s*['"]?go\.mod`.
- [x] `test/ci/fuse_job_test.go::TestFuseJobInstallsFuse3` — input: fuse-linux block; action: `stepContaining(block, "apt-get install")`; assertion: the step is non-empty, contains `apt-get update`, and its install line has the token `fuse3`.
- [x] `test/ci/fuse_job_test.go::TestFuseJobInstallsCrawlerTools` — input: the apt install step; action: split the `apt-get install` line into whitespace tokens; assertion: the tokens include `ripgrep`, `fd-find` and `rsync`, and not a bare `fd` token (that package name does not exist on Ubuntu).
- [x] `test/ci/fuse_job_test.go::TestFuseJobDoesNotInstallDistroRestic` — input: the apt install step (asserted non-empty first); action: tokenise the install line; assertion: no token equals `restic`.
- [x] `test/ci/fuse_job_test.go::TestFuseJobDownloadsPinnedRestic` — input: fuse-linux block; action: substring search; assertion: it contains the literal `https://github.com/restic/restic/releases/download/v0.19.0/restic_0.19.0_linux_amd64.bz2` and contains neither `releases/latest` nor `${{` inside that URL's step.
- [x] `test/ci/fuse_job_test.go::TestFuseJobVerifiesResticChecksum` — input: the step containing the restic URL; action: regex `\b[0-9a-f]{64}\b` plus a substring search; assertion: the step has exactly one 64-hex literal, that literal equals the official 0.19.0 linux_amd64 SHA-256 `13176fe6d89d4357947a2cd107218ab2873a5f9d8e1ac2d4cd1c8e07e6839c21` (a test constant), and the step runs `sha256sum -c` (or `sha256sum --check`) before `bunzip2`.
- [x] `test/ci/fuse_job_test.go::TestFuseJobAssertsResticVersion` — input: fuse-linux block; action: `stepContaining(block, "restic version")`; assertion: the step is non-empty and checks the output for `0.19.0` (for example `grep -F 'restic 0.19.0'`).

## T2 — Gated integration run and evidence artifact

- [x] `test/ci/fuse_job_test.go::TestFuseJobSetsFuseTestsEnv` — input: fuse-linux block; action: regex `SNAPBACK_FUSE_TESTS:\s*["']1["']`; assertion: exactly one match (the value is the quoted string "1").
- [x] `test/ci/fuse_job_test.go::TestFuseJobSetsEvidenceDir` — input: fuse-linux block; action: regex `SNAPBACK_EVIDENCE_DIR:\s*(\S.*)`; assertion: a match exists with a non-empty value.
- [x] `test/ci/fuse_job_test.go::TestFuseJobChecksDevFuseBeforeTests` — input: fuse-linux block; action: find the line index of `/dev/fuse` (with `test -c`, `test -e` or `[ -c`/`[ -e`) and of the `go test ` line; assertion: both exist, the device check comes first, and a `mkdir -p` of `SNAPBACK_EVIDENCE_DIR` also comes before the `go test ` line.
- [x] `test/ci/fuse_job_test.go::TestFuseJobRunsIntegrationTestsWithRace` — input: fuse-linux block; action: take the first line containing `go test `; assertion: it contains `-race`, `-tags=integration` and `./...`.
- [x] `test/ci/fuse_job_test.go::TestFuseJobUploadsEvidenceAlways` — input: fuse-linux block; action: `stepContaining(block, "actions/upload-artifact")`; assertion: the step is non-empty and uses `actions/upload-artifact@v4`, contains `name: stage1-evidence-linux`, `if: always()`, a `path:` referencing `SNAPBACK_EVIDENCE_DIR`, and an `if-no-files-found:` key whose value is not `ignore`.
- [x] `test/ci/fuse_job_test.go::TestFuseJobNeverSetsRcloneRemote` — input: fuse-linux block (non-empty via helper) and the whole ci.yml; action: substring search; assertion: neither contains `SNAPBACK_RCLONE_REMOTE`.
- [x] `test/ci/fuse_job_test.go::TestFuseJobNoLatestAndActionsPinned` — input: fuse-linux block; action: collect `uses:` values and search for `@latest`; assertion: at least one `uses:` exists, each matches `pinnedRef`, and the block contains no `@latest` and no `releases/latest`.
- [x] `test/ci/fuse_job_test.go::TestFuseJobHasNoSecrets` — input: fuse-linux block (non-empty via helper); action: substring search; assertion: it contains no `secrets.` and no `continue-on-error`.

## T3 — Evidence upload fails on missing files (final wave)

- [x] `test/ci/fuse_job_test.go::TestFuseJobUploadFailsOnMissingEvidence` — input: fuse-linux block; action: `stepContaining(block, "actions/upload-artifact")`; assertion: the step is non-empty (M-002) and its `if-no-files-found:` value is exactly `error` (fails by assertion while T2's `warn` is in place).

Test count: 17 (T1 8, T2 8, T3 1).
