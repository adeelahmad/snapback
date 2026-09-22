---
type: validate
story: S1-03
---
# S1-03 — validation rubric

## Pre-flight

- PASS if `test -f go.mod && go version` exits 0 (S1-01 merged). FAIL: this is a hard dependency (sprint1/plan.md wave 2 gate).
- PASS if `git diff --name-only master... | grep -vE '^(install\.sh|test/installer/)'` prints nothing. FAIL: the change touches files outside the owned set (stories.md S1-03 "Owned files").
- PASS if `sh -n install.sh` exits 0.

## T1 — Test harness helper and platform-to-asset mapping

- Run `go test ./test/installer/ -run 'TestAssetName' -v`. Expected: `--- PASS` for all 7 subtests and `ok`. FAIL if any asset string differs from the tasks.md contract, which breaks the S1-06 cross-story contract.
- Run `SNAPBACK_DRY_RUN=1 SNAPBACK_OS=Linux SNAPBACK_ARCH=x86_64 sh install.sh`. Expected: stdout contains `snapback_linux_amd64.tar.gz`.

## T2 — Unsupported OS/arch rejection and unverified warning

- Run `go test ./test/installer/ -run 'TestUnsupported|TestUnverified' -v`. Expected: `ok`.
- Run `SNAPBACK_DRY_RUN=1 SNAPBACK_OS=Windows_NT SNAPBACK_ARCH=x86_64 sh install.sh; echo "exit=$?"`. Expected: stderr contains `unsupported` and `exit=` is not 0. FAIL if it exits 0 (story failure scenario 1).

## T3 — Dry-run plan and next-steps output

- Run `go test ./test/installer/ -run 'TestDryRun|TestNextSteps' -v`. Expected: `ok`.
- Run `SNAPBACK_DRY_RUN=1 SNAPBACK_OS=Darwin SNAPBACK_ARCH=arm64 sh install.sh`. Expected: stdout contains `snapback config`, `macFUSE` and `checksums.txt`. FAIL if any network request is attempted (standards.md honesty gate: dry-run must not claim or fetch a release).

## T4 — Download and SHA-256 verification against httptest

- Run `go test ./test/installer/ -run 'TestInstall' -v`. Expected: `ok`, with `TestInstallGoodChecksum` and `TestInstallBadChecksumRefuses` both `--- PASS`.
- FAIL if a bad checksum leaves any file in the install dir (story failure scenario 2).

## T5 — Install directory fallback

- Run `go test ./test/installer/ -run 'TestInstallDir' -v`. Expected: `ok`.

## T6 — Static safety checks on the script text

- Run `go test ./test/installer/ -run 'TestStatic' -v`. Expected: `ok`.
- Run `grep -nE '(brew|apt(-get)?|dnf|yum|opkg|pacman|apk)[[:space:]]+(install|add)' install.sh | grep -vE 'echo|printf'`. Expected: no output. FAIL: the script installs FUSE (README.md §17 "Never installs FUSE").
- Run `grep -c 'example.invalid' install.sh`. Expected: at least 1. FAIL: a live-looking URL (standards.md honesty gate).

## T7 — shellcheck when available

- Run `go test ./test/installer/ -run 'TestShellcheck' -v`. Expected: `--- PASS` if shellcheck is installed, otherwise `--- SKIP` naming shellcheck. A pass without shellcheck present is a FAIL.
- If installed, run `shellcheck -s sh install.sh`. Expected: exit 0, no output.

## Final sign-off

- Run `gofmt -l test/installer`. Expected: no output.
- Run `go vet ./test/installer/...`. Expected: exit 0.
- Run `go test -race ./test/installer/...`. Expected: `ok`.
- Every checkbox in `plan.md` maps to a test that exists and passes, or skips only for shellcheck.
