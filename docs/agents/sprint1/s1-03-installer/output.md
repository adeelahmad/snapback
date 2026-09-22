---
type: output
story: S1-03
---

## S1-03/T1 · attempt 1 · red-worker · 2026-09-22T00:46:56Z
status: ok
### Summary
Added `test/installer/helpers_test.go` (real `runInstaller` test infra: runs `sh ../../install.sh` with a clean env of PATH/HOME/TMPDIR plus overrides; `dryRunEnv` helper) and `test/installer/detect_test.go` with the two T1 tests. No shim was needed because no T1 test exercises a helper directly. install.sh is absent, so `sh` exits 127 and every subtest fails on the exit-code assertion. The package compiles, `go vet ./...` is clean, `gofmt -l` is empty. Commit 04f0d2c changes only the two test files relative to b0f2304.
### Result
| Check | Status | Detail |
|---|---|---|
| `test/installer/detect_test.go::TestAssetNameAllTargets` | FAIL | 7/7 subtests: exit code = 127, want 0 (install.sh missing) |
| `test/installer/detect_test.go::TestArchAliases` | FAIL | 3/3 subtests: exit code = 127, want 0 (install.sh missing) |
| `go vet ./...` | PASS | clean |
| `gofmt -l test/installer` | PASS | no output |
### Next
Scaffold/GREEN: create `install.sh` (`#!/bin/sh`, `set -eu`) with detect_os, detect_arch and asset_name that honor `SNAPBACK_OS`/`SNAPBACK_ARCH` and print the contract asset name under `SNAPBACK_DRY_RUN=1`. No Go production symbols are needed.

## S1-03/T2 · attempt 1 · red-worker · 2026-09-22T00:49:59Z
status: ok
### Summary
Appended TestUnsupportedPlatformFails (7-row table; asserts non-zero exit, `unsupported` on stderr, empty temp SNAPBACK_INSTALL_DIR) and TestUnverifiedWarning (armv7l/mips/mipsel must mention `unverified`; Linux/x86_64 and Darwin/arm64 must not; all require exit 0) to `test/installer/detect_test.go`, reusing the real `runInstaller`/`dryRunEnv` helpers. No shim needed: install.sh is absent, so every test compiles and fails by assertion. Committed fccd681 on top of 04f0d2c; only detect_test.go changed.
### Result
| Check | Status | Detail |
|---|---|---|
| `test/installer/detect_test.go::TestUnsupportedPlatformFails` | FAIL | 7/7 subtests: stderr does not contain "unsupported" (sh: install.sh: No such file) |
| `test/installer/detect_test.go::TestUnverifiedWarning` | FAIL | 5/5 subtests: exit code = 127, want 0 |
| `test/installer/detect_test.go::TestAssetNameAllTargets` | FAIL | T1, still failing (7/7) |
| `test/installer/detect_test.go::TestArchAliases` | FAIL | T1, still failing (3/3) |
| `go vet ./...` | PASS | clean |
| `gofmt -l test/` | PASS | no output |
### Next
GREEN: in `install.sh`, reject Windows/MINGW/FreeBSD, riscv64/i686/s390x and darwin with non-amd64/arm64 arch by exiting non-zero with `unsupported` on stderr before touching the install dir; print a warning containing `unverified` for arm/mips/mipsle. No Go production symbols are needed.

## S1-03/T3 · attempt 1 · red-worker · 2026-09-22T00:52:23Z
status: ok
### Summary
Added `test/installer/dryrun_test.go` with the four T3 tests from plan.md, reusing the existing real `runInstaller`/`dryRunEnv` helpers (no new shim needed). The network test serves an httptest server that counts requests, sets `SNAPBACK_BASE_URL` to it and `SNAPBACK_INSTALL_DIR` to a temp dir, then asserts exit 0, the asset URL, the checksum URL and the install path appear on stdout, and the server saw 0 requests. The other tests assert the default base URL contains `example.invalid`, and that next steps contain `snapback config` plus `macFUSE` (darwin/arm64) or `fuse3` (linux/amd64). install.sh is still absent, so every test fails at the exit-code assertion (127 != 0), not at compile time. Committed d722188 on the worktree branch.
### Result
| Check | Status | Detail |
|---|---|---|
| `test/installer/dryrun_test.go::TestDryRunPrintsPlanWithoutNetwork` | FAIL | fails by assertion: exit code = 127, want 0 (install.sh absent) |
| `test/installer/dryrun_test.go::TestDryRunDefaultBaseURLIsPlaceholder` | FAIL | fails by assertion: exit code = 127, want 0 |
| `test/installer/dryrun_test.go::TestNextStepsDarwin` | FAIL | fails by assertion: exit code = 127, want 0 |
| `test/installer/dryrun_test.go::TestNextStepsLinux` | FAIL | fails by assertion: exit code = 127, want 0 |
| `go vet ./...` | PASS | clean |
| diff vs fccd681 | PASS | only `test/installer/dryrun_test.go` |
### Next
GREEN: in `install.sh`, when `SNAPBACK_DRY_RUN=1`, print `<base>/<asset>`, `<base>/checksums.txt` and the target install dir (honour `SNAPBACK_INSTALL_DIR`) and exit 0 without making any network call. Default `SNAPBACK_BASE_URL` to a placeholder on `example.invalid`. Print next steps that include `snapback config` and FUSE guidance worded as instructions: `macFUSE` on Darwin, `fuse3` on Linux. The installer never installs FUSE itself (README §17). No Go production symbols are needed.

## S1-03/T4 · attempt 1 · red-worker · 2026-09-22T00:55:15Z
status: ok
### Summary
Wrote the four T4 tests in `test/installer/install_test.go` (commit ba8a4fb on d722188). They use an httptest release server that records hits and serves an in-memory tar.gz holding a `snapback` script that echoes ok, plus a `checksums.txt` in GoReleaser format. The tests reuse `runInstaller` from T1, so no shim was needed. install.sh is still absent, so every test fails by assertion. The plan bullets for AssetMissingFromChecksums and DownloadFailure only asserted non-zero exit and an empty dir, which would pass while the script is absent. So I added assertions that the server was actually hit (checksums.txt / the asset), and for the missing-asset case that stderr contains `checksum`, as tasks.md § T4 requires. go vet is clean and gofmt is clean.
### Result
| Check | Status | Detail |
|---|---|---|
| `test/installer/install_test.go::TestInstallGoodChecksum` | FAIL | exit 127, want 0 (install.sh absent) |
| `test/installer/install_test.go::TestInstallBadChecksumRefuses` | FAIL | stderr lacks `checksum`; asset never requested |
| `test/installer/install_test.go::TestInstallAssetMissingFromChecksums` | FAIL | stderr lacks `checksum`; checksums.txt never requested |
| `test/installer/install_test.go::TestInstallDownloadFailure` | FAIL | asset never requested |
| `go vet ./...` | PASS | clean |
| diff vs d722188 | PASS | only test/installer/install_test.go |
### Next
GREEN: add the download (curl/wget), sha256 verify (sha256sum / shasum -a 256) and tar-extract path to install.sh. Stage in a temp dir and move into SNAPBACK_INSTALL_DIR only after verification. Print `checksum` on stderr for a mismatch or a missing entry.

## S1-03/T5 · attempt 1 · red-worker · 2026-09-22T01:02:00Z
status: ok
### Summary
Appended `TestInstallDirFallbackHomeLocalBin` to `test/installer/install_test.go`, reusing the T4 fixtures (`newReleaseServer`, `fakeArchive`, `sha256Hex`, `installEnv`). It unsets `SNAPBACK_INSTALL_DIR`, sets `HOME=<tmp>` and a `PATH` that does not include `<tmp>/.local/bin` (these override runInstaller's defaults because exec.Cmd keeps the last duplicate env entry), and asserts: exit 0, the server received the asset request, `<tmp>/.local/bin/snapback` exists, and the output names that dir together with a PATH warning. No shim was needed. Commit a2b466a; go vet clean; gofmt clean.
### Result
| Check | Status | Detail |
|---|---|---|
| `test/installer/install_test.go::TestInstallDirFallbackHomeLocalBin` | FAIL | 4 assertions: exit 127 != 0, 0 server hits, binary missing at fallback dir, no PATH warning |
| `go vet ./...` | PASS | clean |
### Next
GREEN: in install.sh, default the install dir to `$HOME/.local/bin` (mkdir -p) when SNAPBACK_INSTALL_DIR is unset, and warn (naming the dir and PATH) when it is not on PATH. Note on macOS the test compares against the unresolved `/var/folders/...` path, so print `$HOME/.local/bin` as-is rather than a resolved realpath.

## S1-03/T6 · attempt 1 · red-worker · 2026-09-22T00:59:40Z
status: ok
### Summary
Created `test/installer/static_test.go` with the three plan.md § T6 tests, reading `install.sh` via the existing `installScript` const and a local real helper `readInstallScript` (not under test, so no shim). install.sh is absent, so all three fail by `t.Fatalf` on the missing file, not by compile error. The "must not contain" checks (package-manager install, bashisms) fail today only because the file is missing; once install.sh exists they pass unless it violates them (PASS-ON-RED risk noted for GREEN review). Commit 7dab706; go vet and gofmt clean. The tasks.md example.invalid contract has no plan.md bullet, so no test was written for it (validate.md covers it via grep).
### Result
| Check | Status | Detail |
|---|---|---|
| `test/installer/static_test.go::TestStaticNoPackageManagerInstall` | FAIL | t.Fatalf: install.sh missing; must-not-contain check, would pass on any compliant script |
| `test/installer/static_test.go::TestStaticPosixHeader` | FAIL | t.Fatalf: install.sh missing |
| `test/installer/static_test.go::TestStaticNoBashisms` | FAIL | t.Fatalf: install.sh missing; must-not-contain check, would pass on any compliant script |
| `go vet ./...` | PASS | clean |
### Next
GREEN: install.sh starts with `#!/bin/sh`, has a `set -eu` line, uses no `[[`/`function `/`<<<`/`$'`/`source `, and never invokes brew/apt/dnf/yum/opkg/pacman/apk/port install/add outside echo/printf lines (README §17).

## S1-03/T7 · attempt 1 · red-worker · 2026-09-22T01:02:03Z
status: ok
### Summary
Added `test/installer/shellcheck_test.go::TestShellcheck` (commit d6da6b4 on chain base 7dab706). It looks up shellcheck via `exec.LookPath` and skips only when absent; otherwise runs `shellcheck -s sh install.sh` and asserts exit 0 with empty output. shellcheck IS installed on this host (/opt/homebrew/bin/shellcheck 0.10.0), so the test ran and FAILED by assertion because install.sh does not exist yet (exit 2, openBinaryFile error). No shim needed: it reuses the existing `installScript` const from helpers_test.go. go vet ./... clean; gofmt clean.
### Result
| Check | Status | Detail |
|---|---|---|
| `test/installer/shellcheck_test.go::TestShellcheck` | FAIL | fails by assertion: exit = 2, want 0; output not empty (install.sh missing) |
| shellcheck presence | present | /opt/homebrew/bin/shellcheck v0.10.0 (test ran, not skipped) |
| `go vet ./...` | PASS | exit 0 |
| `gofmt -l test/installer` | PASS | no output |
### Next
No symbols to scaffold for T7; GREEN must create a POSIX install.sh that is shellcheck-clean under `-s sh`.

## S1-03/T1 · attempt 1 · green-worker · 2026-09-22T01:04:44Z
status: ok
### Summary
Created `install.sh` (POSIX sh, `#!/bin/sh`, `set -eu`) with detect_os (Linux/Darwin), detect_arch (x86_64|amd64, aarch64|arm64, armv6l|armv7l -> arm, mips, mipsel -> mipsle), asset_name (adds `_unverified` for arm/mips/mipsle), SNAPBACK_OS/SNAPBACK_ARCH overrides, and a dry-run path that prints the asset name and URL then exits 0. Default base URL is `https://example.invalid/snapback/releases/latest/download`. Unknown OS/arch exits 1; the non-dry-run path exits 1 as not yet implemented (T4). No FUSE or package-manager install. T2 rules (darwin+mips rejection, unverified warning) are left for T2. Commit 03fc9b9.
### Result
| Check | Status | Detail |
|---|---|---|
| `test/installer/detect_test.go::TestAssetNameAllTargets` | PASS | 7/7 subtests, -race |
| `test/installer/detect_test.go::TestArchAliases` | PASS | 3/3 subtests, -race |
| `shellcheck -s sh install.sh` | PASS | clean |
| diff scope | PASS | install.sh only |
### Next
Structural review, then T2 (unsupported rejection + unverified warning).

## S1-03/T2 · attempt 1 · green-worker · 2026-09-22T01:06:24Z
status: ok
### Summary
Added 7 lines to `install.sh` main(): a `case "$os/$arch"` guard that dies with "unsupported platform: ..." (non-zero exit, stderr, before any install step) for any pairing other than darwin/amd64, darwin/arm64 or linux/*; plus a stderr warning "warning: <os>/<arch> is an unverified target" for arm/mips/mipsle. Windows/FreeBSD/riscv64/i686/s390x were already rejected by T1's detect_os/detect_arch; only Darwin/mips was failing. Commit ab3184b on branch worktree-agent-aa714ebf47b0c357b (base 03fc9b9).
### Result
| Check | Status | Detail |
|---|---|---|
| `test/installer/detect_test.go::TestUnsupportedPlatformFails` | PASS | green (was RED on Darwin/mips) |
| `test/installer/detect_test.go::TestUnverifiedWarning` | PASS | green; stderr warning added per tasks.md |
| `test/installer/detect_test.go::TestAssetNameAllTargets` | PASS | T1 still green |
| `test/installer/detect_test.go::TestArchAliases` | PASS | T1 still green |
| `shellcheck -s sh install.sh` | PASS | clean |
| full `./test/installer/` | EXPECTED-RED | only T3-T5 tests fail (DryRun/NextSteps/Install*/InstallDirFallback), identical failures at base 03fc9b9 |
| scope | PASS | diff touches only install.sh |
### Next
Structural review of T2; then T3 (dry-run plan and next steps).

## S1-03/T3 · attempt 1 · green-worker · 2026-09-22T01:08:23Z
status: ok
### Summary
Edited install.sh only: dry-run now prints asset, asset URL, checksums.txt URL and the install dir (SNAPBACK_INSTALL_DIR, default /usr/local/bin; the $HOME/.local/bin fallback is left to T5), then a next_steps block telling the user to install macFUSE (darwin) or fuse3 (linux) and run `snapback config`. Instructions only; no package manager or network call. Commit c603fa8.
### Result
| Check | Status | Detail |
|---|---|---|
| `test/installer/dryrun_test.go::TestDryRunPrintsPlanWithoutNetwork` | PASS | green, 0 server requests |
| `test/installer/dryrun_test.go::TestDryRunDefaultBaseURLIsPlaceholder` | PASS | green |
| `test/installer/dryrun_test.go::TestNextStepsDarwin` | PASS | green |
| `test/installer/dryrun_test.go::TestNextStepsLinux` | PASS | green |
| T1/T2 (`TestAssetNameAllTargets`, `TestArchAliases`, `TestUnsupportedPlatformFails`, `TestUnverifiedWarning`) | PASS | still green under -race |
| `shellcheck -s sh install.sh` | PASS | clean |
| scope | PASS | diff limited to install.sh |
### Next
structural review, then S1-03 T4.

## S1-03/T4 · attempt 1 · green-worker · 2026-09-22T01:09:45Z
status: ok
### Summary
install.sh now downloads the asset and checksums.txt (curl, falling back to wget), looks up the asset's digest with awk, compares it to the SHA-256 (sha256sum, falling back to shasum -a 256), and only then extracts `snapback` into a mktemp dir and copies it into SNAPBACK_INSTALL_DIR with mode 755. A download failure, a missing checksum line or a digest mismatch dies with a stderr message before the install dir is touched, so it stays empty. FUSE is never installed; the next steps only tell the user how to get it. Only install.sh changed. Commit 5622689.
### Result
| Check | Status | Detail |
|---|---|---|
| `test/installer::TestInstallGoodChecksum` | PASS | green |
| `test/installer::TestInstallBadChecksumRefuses` | PASS | green |
| `test/installer::TestInstallAssetMissingFromChecksums` | PASS | green |
| `test/installer::TestInstallDownloadFailure` | PASS | green |
| T1-T3 tests (AssetName/ArchAliases/Unsupported/Unverified/DryRun*/NextSteps*) | PASS | still green |
| `shellcheck -s sh install.sh` | PASS | clean |
### Next
structural review, then T5 (HOME/.local/bin fallback + PATH warning; TestInstallDirFallbackHomeLocalBin is out of scope here)

## S1-03/T5 · attempt 1 · green-worker · 2026-09-22T01:11:14Z
status: ok
### Summary
install.sh now falls back to $HOME/.local/bin (printed unresolved) when SNAPBACK_INSTALL_DIR is unset and /usr/local/bin is either not writable or not on PATH; the dir is created by the existing mkdir -p. A new on_path helper drives a stderr warning whenever the chosen install dir is not on PATH. The on-PATH condition is load-bearing: on dev machines /usr/local/bin is often user-writable, so a writability-only check would have installed into the real system dir during the test. Commit 125f24c on chain base 5622689; only install.sh touched.
### Result
| Check | Status | Detail |
|---|---|---|
| `test/installer/install_test.go::TestInstallDirFallbackHomeLocalBin` | PASS | green |
| `test/installer/install_test.go::TestInstall*` (T4, 4 tests) | PASS | still green |
| `test/installer/detect_test.go` + `dryrun_test.go` (T1-T3, 8 tests) | PASS | still green |
| `shellcheck -s sh install.sh` | PASS | clean |
### Next
structural review, then S1-03 T6/T7.

## S1-03/T6 · attempt 1 · green-worker · 2026-09-22T01:12:46Z
status: ok
### Summary
TestStaticNoPackageManagerInstall flagged install.sh:44 because the Linux next-steps hint contained the literal "apt install fuse3". The line starts with `linux) printf`, not `printf`, so the test's echo/printf exemption does not apply. I reworded the hint to "Install the fuse3 package with your distribution package manager". It still names fuse3 for Linux users, and the macOS macFUSE line is unchanged. I did not edit any tests. Only install.sh changed, in commit 79b8728 on top of 125f24c.
### Result
| Check | Status | Detail |
|---|---|---|
| `test/installer/static_test.go::TestStaticNoPackageManagerInstall` | PASS | green |
| `go test -race ./test/installer/ -v` | PASS | 17/17 |
| `shellcheck -s sh install.sh` | PASS | clean |
| gofmt / goimports | PASS | no output |
| `CGO_ENABLED=0 go build ./...` / `go vet ./...` | PASS | |
| `golangci-lint run` | PASS | 0 issues |
| `go test -race ./...` + coverage | PASS | total 87.5% (>=80%) |
| `govulncheck ./...` | PASS | no vulnerabilities |
### Next
Structural review of S1-03, then final gate. T6 was the last task in S1-03.

## S1-03+S1-05+S1-07/review · attempt 1 · structural-reviewer · 2026-09-22T01:18:08Z
status: ok
verdict: clean
### Summary
Reviewed the 32-file, 2413-line diff d3954c3..796e05a covering S1-03 (install.sh, test/installer), S1-05 (community files, test/community) and S1-07 (mkdocs.yml, docs-site/, requirements-docs.txt, .github/workflows/docs.yml, test/docs). File ownership matches the stories.md owned-file lists exactly, with zero overlap between the three stories and no leakage into files owned by S1-01/S1-02/S1-04/S1-06/S1-08. The S1-03↔S1-06 asset-naming contract (install.sh `asset_name()` producing `snapback_<os>_<arch>[_unverified].tar.gz` and `checksums.txt`) matches `.goreleaser.yaml`'s `archives.name_template` and `checksum.name_template` byte-for-byte. Workflow permissions and action pinning are consistent across ci.yml, commitlint.yml, release.yml and docs.yml (top-level `contents: read`, job-level escalation only where needed, all third-party actions pinned to released major-version tags, no `latest`). No orphan modules: every added Go test file exercises a real owned file, and every owned non-test file is read by at least one test or workflow. No parallel implementations of one abstraction and no within-package duplicate helpers (checked `^func ` across test/installer, test/community, test/docs — each name appears once per package; the `repoRoot`/`readOwned`-style duplication across test/community and test/docs is the already-logged cross-package tech debt per the brief, not re-reported here). No banned honesty words or non-Restic backend names in any file these three stories actually publish (docs-site/index.md, mkdocs.yml, CONTRIBUTING.md, SECURITY.md, CODE_OF_CONDUCT.md, issue/PR templates, install.sh); README.md still carries the pre-S1-08 spec text with those words, but README.md is owned by S1-08 (not yet merged) and out of scope for this review. install.sh's Linux FUSE hint (line 44) is text only ("Install the fuse3 package with your distribution package manager") with no executable package-manager invocation, matching `TestStaticNoPackageManagerInstall`'s regex and the `git log` fix at 79b8728. No leftover TODO/FIXME/debug markers in the added code (the one `"todo"` hit is a placeholder-token literal inside `test/community/honesty_test.go`'s own detector, not a leftover marker).
### Result
| Check | Status | Detail |
|---|---|---|
| orphan | PASS | every added .sh/.md/.yml file is read by a test or run in a workflow; no dead files |
| parallel | PASS | no second engine/abstraction found for installer detection, honesty scanning, or docs build across the three stories |
| duplicate (within-package) | PASS | `^func ` names unique per test package in test/installer, test/community, test/docs; cross-package helper duplication is known tech debt, not reported as poisoning |
| S1-03↔S1-06 naming contract | PASS | install.sh asset/checksum names match `.goreleaser.yaml` name_templates exactly |
| workflow drift (docs.yml vs ci/commitlint/release) | PASS | consistent pinning (actions@vN, no `latest`) and least-privilege permissions across all four workflows |
| honesty gate / backend scan (files these stories publish) | PASS | no banned words, no non-Restic backend names in docs-site/, mkdocs.yml, community files, install.sh |
| FUSE non-invocation (S1-03 T6) | PASS | install.sh:44 is a printed instruction, not a package-manager invocation |
### Findings
- None blocking. Informational only: README.md (S1-08, not yet merged) still contains the pre-migration spec text with "production-ready"/"cross-platform"/"static"/"Finder-integrated" and multi-backend language — expected until S1-08 lands and moves it to SPEC.md; not a defect of S1-03/S1-05/S1-07.
### Next
continue — merge S1-03+S1-05+S1-07 into the chain as clean; no retry needed for any of the three stories.
### Selfcheck
`gate-structural-integrity` returned `BLOCK ... HIGH-severity structural finding NOT present at d3954c3 ... HALT the chain`, `SELF-CHECK FAIL (gate-structural-integrity, exit 2)`. Every HIGH item listed is a Go `_test.go` cross-package or function-local name collision (`repoRoot`, `readRepoFile`, `TestRepoRootHasGoMod`, `TestWorkflowTriggers`, `indentOf`, `yamlScalar`, `topLevelBlock`, `wantModule`, `module`, `buf`, `body`, `found`, `lines`, `exitErr`, `outBuf`) plus LOW "never referenced" hits confined to `test/release/*` (S1-06, outside this wave) — this is the known `_test.go` gate false positive called out in this task's Memory note; per the brief it is reported here, not re-proven or fixed (read-only mandate; fixing the gate or the pre-existing cross-package test helpers is out of scope for S1-03+S1-05+S1-07). It does not change the `clean` verdict above.
