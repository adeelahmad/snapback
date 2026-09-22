---
type: tasks
story: S2-05
---

# S2-05 tasks — Linux FUSE CI job

Owned set (stories.md S2-05, sprint2/plan.md row S2-05): `.github/workflows/ci.yml` (edit: append one job) and `test/ci/fuse_job_test.go` (create). Nothing else is touched. `test/ci/ci_test.go`, `matrix_test.go`, `lint_test.go` and `helpers_test.go` are read-only (sprint2/plan.md; M-012). If an existing assertion conflicts with the new job, stop and report to the orchestrator; do not edit it.

Test package: `test/ci` (`package ci_test`), standard library only, text-based checks over the raw YAML (no YAML dependency; `go.mod` is S2-01's). The new file reuses the existing helpers without redefining them: `readCI`, `jobBlock` (matrix_test.go), `stepsContaining`/`stepContaining`, `indentOf` (helpers_test.go) and `pinnedRef` (ci_test.go). It adds one local helper, `fuseJobBlock(t)`, that returns `jobBlock(readCI(t), "fuse-linux")` and calls `t.Fatalf` when it is empty, so every negative check first proves the job block is non-empty (M-002). All new test names start with `TestFuseJob` so they cannot collide with the 22 existing tests.

Tasks run in order T1 -> T2 (both edit `ci.yml` and the same test file, so they are not parallel). There are no production Go symbols, so SCAFFOLD is a no-op for both tasks.

## Constraints both tasks must respect (from the existing 22 tests)

- **Append the job at the end of `jobs:`**, after `cross-compile`. `TestCITestUsesRaceAndCoverage` takes the *first* `go test ` line in the file and `TestCIUploadsCoverageArtifact` takes the *first* `actions/upload-artifact` step. Both must stay in the `test` job.
- `TestCISetupGoUsesGoModVersionFile`: the new `actions/setup-go` step must use `go-version-file: go.mod`, and no `go-version:` key may appear.
- `TestCINoMountJobs`: no `name:` line anywhere (job or step, including the artifact `name:`) may contain `mount`, `browse` or `unmount`, case-insensitive. The job id `fuse-linux` is fine. Use step names such as "Install FUSE and crawler tools", "Install restic 0.19.0", "Check /dev/fuse", "Integration tests", "Upload evidence".
- `TestCINoDependencyOnOtherStoryFiles`: the job must not mention `install.sh`, `.goreleaser`, `.releaserc`, `mkdocs` or `commitlint`.
- `TestCINoContinueOnError` and `TestCIActionsArePinned`: no `continue-on-error: true`, and every `uses:` pinned.
- Pins match the existing file: `actions/checkout@v4`, `actions/setup-go@v5`, `actions/upload-artifact@v4`. No `@latest` and no `releases/latest` URL (M-006).
- No `secrets.` reference and no `SNAPBACK_RCLONE_REMOTE`, so latency tests skip in CI.

## T1 — `fuse-linux` job skeleton and pinned tool installs

- **Files:** `.github/workflows/ci.yml` (edit: append `fuse-linux` job), `test/ci/fuse_job_test.go` (create, with `fuseJobBlock` and the T1 tests).
- **Does:** Adds `jobs.fuse-linux` with `runs-on: ubuntu-latest`. It inherits the workflow's existing `on:` (push to `master`, pull_request) and `permissions: contents: read`, and adds no job-level `if:`. Steps:
  1. `actions/checkout@v4`, then `actions/setup-go@v5` with `go-version-file: go.mod`.
  2. `sudo apt-get update` and `sudo apt-get install -y fuse3 ripgrep fd-find rsync`. Note that `fd-find` installs the binary `fdfind`, not `fd`. S2-07 resolves that name, so this job does not symlink it. The apt line must not install `restic` (the distro package is not 0.19.0).
  3. Install restic from the official release asset, with the URL and checksum as pinned literals (no version variable):
     - URL: `https://github.com/restic/restic/releases/download/v0.19.0/restic_0.19.0_linux_amd64.bz2`
     - SHA-256: `13176fe6d89d4357947a2cd107218ab2873a5f9d8e1ac2d4cd1c8e07e6839c21`. The planner read this value on 2026-09-22 from `https://github.com/restic/restic/releases/download/v0.19.0/SHA256SUMS`, line `restic_0.19.0_linux_amd64.bz2`. Pre-flight re-reads it.
     - The step runs `curl -fsSL -o restic.bz2 <URL>`, then `echo "<sha>  restic.bz2" | sha256sum -c -` (a mismatch fails the step), then `bunzip2 restic.bz2`, then `sudo install -m 0755 restic /usr/local/bin/restic`, then `restic version | grep -F 'restic 0.19.0'`.
- **Tests:** plan.md T1 block (8 tests).

## T2 — Gated integration run and evidence artifact, then the full matrix

- **Files:** `.github/workflows/ci.yml` (edit the `fuse-linux` job only), `test/ci/fuse_job_test.go` (append the T2 tests).
- **Does:** Adds a job-level `env:` to `fuse-linux` with `SNAPBACK_FUSE_TESTS: "1"` (quoted string) and `SNAPBACK_EVIDENCE_DIR: ${{ github.workspace }}/evidence`. Then appends these steps after the installs:
  1. A "Check /dev/fuse" step that runs `test -c /dev/fuse` and also `mkdir -p "$SNAPBACK_EVIDENCE_DIR"`, so the upload path exists even if every test fails early.
  2. An "Integration tests" step that runs `go test -race -tags=integration ./...`.
  3. An "Upload evidence" step with `actions/upload-artifact@v4`, `if: always()`, `name: stage1-evidence-linux`, `path: ${{ env.SNAPBACK_EVIDENCE_DIR }}` and `if-no-files-found: warn`. The upload runs on failure too, so failure evidence is not hidden. Because the directory is created beforehand, the step never fails confusingly on a missing path. `warn` keeps the wave-1 merge green before S2-03, S2-04, S2-06 and S2-07 write any JSON. The orchestrator's exit evidence check (validate.md) is what proves the files exist. See the open question in validate.md.
  4. After the T2 tests pass, it runs `actionlint` locally (installed at `/opt/homebrew/bin/actionlint`) and then **the full standards matrix** (sprint2/standards.md "Cross-cutting gates"). This is the story's last task (M-005).
- **Tests:** plan.md T2 block (8 tests).

## T3 — Evidence upload fails on missing files (final wave)

- **Files:** `.github/workflows/ci.yml` (edit the `fuse-linux` upload step only), `test/ci/fuse_job_test.go` (append the T3 test). Same owned set as T1/T2; no new file.
- **Does:** Changes `if-no-files-found: warn` to `if-no-files-found: error` in the "Upload evidence" step. Runs only after S2-03, S2-04, S2-06 and S2-07 are merged on `master` (sprint plan wave 4), because only then do the integration tests write JSON; before that, `error` would turn every run red. Its GREEN runs the full standards matrix again (last GREEN of the story, M-005).
