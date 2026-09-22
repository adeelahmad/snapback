---
type: plan
story: S1-03
scope: "tests only"
---
# S1-03 — test plan (tests only)

All tests live in package `installer` under `test/installer/` and run `sh install.sh` through `runInstaller` (T1 helper).
Asset names follow the tasks.md contract: `snapback_<os>_<arch>[_unverified].tar.gz` plus `checksums.txt`.

## T1 — Test harness helper and platform-to-asset mapping

- [ ] `test/installer/detect_test.go::TestAssetNameAllTargets` — table of 7 pairs (Linux/x86_64, Linux/aarch64, Darwin/x86_64, Darwin/arm64, Linux/armv7l, Linux/mips, Linux/mipsel). Action: run with `SNAPBACK_DRY_RUN=1` and the `SNAPBACK_OS`/`SNAPBACK_ARCH` overrides. Assert: exit 0 and stdout contains the exact contract asset name.
- [ ] `test/installer/detect_test.go::TestArchAliases` — input `amd64`, `arm64` and `armv6l` under Linux. Action: dry-run. Assert: they map to `linux_amd64`, `linux_arm64` and `linux_arm_unverified` respectively.

## T2 — Unsupported OS/arch rejection and unverified warning

- [ ] `test/installer/detect_test.go::TestUnsupportedPlatformFails` — table (Windows_NT/x86_64, MINGW64_NT/x86_64, FreeBSD/amd64, Linux/riscv64, Linux/i686, Linux/s390x, Darwin/mips). Action: dry-run with overrides and a temp `SNAPBACK_INSTALL_DIR`. Assert: non-zero exit, stderr contains `unsupported`, and the install dir is empty.
- [ ] `test/installer/detect_test.go::TestUnverifiedWarning` — input Linux with armv7l, mips and mipsel. Action: dry-run. Assert: combined output contains `unverified`. Verified targets (Linux/x86_64, Darwin/arm64) do not contain `unverified`.

## T3 — Dry-run plan and next-steps output

- [ ] `test/installer/dryrun_test.go::TestDryRunPrintsPlanWithoutNetwork` — input `SNAPBACK_DRY_RUN=1`, Linux/x86_64, and `SNAPBACK_BASE_URL` pointing at an httptest server that counts requests. Action: run. Assert: exit 0, stdout contains `<base>/snapback_linux_amd64.tar.gz`, `<base>/checksums.txt` and the install path, and the server saw 0 requests.
- [ ] `test/installer/dryrun_test.go::TestDryRunDefaultBaseURLIsPlaceholder` — input dry-run with no `SNAPBACK_BASE_URL`. Action: run. Assert: stdout contains `example.invalid`.
- [ ] `test/installer/dryrun_test.go::TestNextStepsDarwin` — input Darwin/arm64 dry-run. Action: run. Assert: output contains `snapback config` and `macFUSE`.
- [ ] `test/installer/dryrun_test.go::TestNextStepsLinux` — input Linux/amd64 dry-run. Action: run. Assert: output contains `snapback config` and `fuse3`.

## T4 — Download and SHA-256 verification against httptest

- [ ] `test/installer/install_test.go::TestInstallGoodChecksum` — input an httptest server serving `snapback_linux_amd64.tar.gz` (a tar.gz holding a `snapback` script that echoes `ok`) and a matching `checksums.txt`, with `SNAPBACK_OS=Linux`, `SNAPBACK_ARCH=x86_64`, `SNAPBACK_BASE_URL=<srv>` and `SNAPBACK_INSTALL_DIR=<tmp>`. Action: run. Assert: exit 0, `<tmp>/snapback` exists with the executable bit set, and running it prints `ok`.
- [ ] `test/installer/install_test.go::TestInstallBadChecksumRefuses` — input the same server, but `checksums.txt` lists a wrong 64-hex digest for the asset. Action: run. Assert: non-zero exit, stderr contains `checksum`, and `<tmp>` contains no files.
- [ ] `test/installer/install_test.go::TestInstallAssetMissingFromChecksums` — input a `checksums.txt` that lists only other assets. Action: run. Assert: non-zero exit and `<tmp>` is empty.
- [ ] `test/installer/install_test.go::TestInstallDownloadFailure` — input a server that returns 404 for the asset. Action: run. Assert: non-zero exit and `<tmp>` is empty.

## T5 — Install directory fallback

- [ ] `test/installer/install_test.go::TestInstallDirFallbackHomeLocalBin` — input the good-checksum server, `SNAPBACK_INSTALL_DIR` unset, `HOME=<tmp>`, and a `PATH` without `<tmp>/.local/bin`. Action: run. Assert: exit 0, `<tmp>/.local/bin/snapback` exists, and output warns that the dir is not on PATH.

## T6 — Static safety checks on the script text

- [ ] `test/installer/static_test.go::TestStaticNoPackageManagerInstall` — input the contents of `install.sh`. Action: regex `(brew|apt|apt-get|dnf|yum|opkg|pacman|apk|port)\s+(install|add)` over lines not starting with `echo`/`printf`. Assert: no matches.
- [ ] `test/installer/static_test.go::TestStaticPosixHeader` — input `install.sh`. Action: read it. Assert: the first line is `#!/bin/sh` and a `set -eu` line exists.
- [ ] `test/installer/static_test.go::TestStaticNoBashisms` — input `install.sh`. Action: search for `[[`, `function `, `<<<`, `$'` and `source `. Assert: none found.

## T7 — shellcheck when available

- [ ] `test/installer/shellcheck_test.go::TestShellcheck` — input `install.sh`. Action: `exec.LookPath("shellcheck")`; if it is absent, `t.Skip("shellcheck not installed; CI runs it via S1-02")`, otherwise run `shellcheck -s sh install.sh`. Assert: exit 0 with empty output.
