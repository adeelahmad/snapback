---
type: tasks
story: S1-03
---
# S1-03 — `install.sh` skeleton: tasks

Owned set: `install.sh`, `test/installer/**`. No task touches any other path.
Tests run from the repo root as `go test ./test/installer/...`; every test shells out to `sh install.sh` (never `bash`) so bashisms fail.

## Asset naming contract (cross-story, S1-06 tests the other half)

- Archive: `snapback_<os>_<arch>[_unverified].tar.gz`, no version in the name, so `.../releases/latest/download/<asset>` URLs resolve (rclone model).
  - `<os>` is `linux` or `darwin`. `<arch>` is `amd64`, `arm64`, `arm`, `mips` or `mipsle`.
  - The `_unverified` suffix is on `linux_arm`, `linux_mips` and `linux_mipsle` only.
  - The 7 assets: `snapback_linux_amd64.tar.gz`, `snapback_linux_arm64.tar.gz`, `snapback_darwin_amd64.tar.gz`, `snapback_darwin_arm64.tar.gz`, `snapback_linux_arm_unverified.tar.gz`, `snapback_linux_mips_unverified.tar.gz`, `snapback_linux_mipsle_unverified.tar.gz`.
- Checksum file: `checksums.txt`, in the same directory as the assets. Its format is GoReleaser's sha256 output, one `<64-hex>  <asset>` line per asset.
- The archive holds the binary `snapback` at its root.
- The S1-06 GoReleaser equivalents are `archives.name_template: "{{ .ProjectName }}_{{ .Os }}_{{ .Arch }}{{ if or (eq .Arch \"arm\") (eq .Arch \"mips\") (eq .Arch \"mipsle\") }}_unverified{{ end }}"` with `format: tar.gz`, and `checksum.name_template: checksums.txt` with `algorithm: sha256`. They also need `goarm`/`gomips` settings that keep `.Arch` bare (no `v6`/`_softfloat` suffix).
- Download URL: `${SNAPBACK_BASE_URL}/<asset>` and `${SNAPBACK_BASE_URL}/checksums.txt`. The default base is the placeholder `https://example.invalid/snapback/releases/latest/download`, which never resolves.

## T1 — Test harness helper and platform-to-asset mapping

- Files: `test/installer/helpers_test.go` (new; `runInstaller(t, env map[string]string) (stdout, stderr string, exitCode int)` runs `sh ../../install.sh` with a clean env plus the overrides), `test/installer/detect_test.go`, `install.sh` (detect_os, detect_arch, asset_name).
- Maps `uname -s`/`uname -m` (or `SNAPBACK_OS`/`SNAPBACK_ARCH`) to the contract names and prints the asset name in dry-run mode.
- Header: `#!/bin/sh`, `set -eu`.

## T2 — Unsupported OS/arch rejection and unverified warning

- Files: `test/installer/detect_test.go`, `install.sh`.
- Windows, FreeBSD, `riscv64`, `i686`, `s390x`, and `darwin` paired with any arch outside amd64/arm64, each exit non-zero with `unsupported` on stderr and install nothing.
- arm/mips/mipsle print a warning containing `unverified`.

## T3 — Dry-run plan and next-steps output

- Files: `test/installer/dryrun_test.go`, `install.sh`.
- `SNAPBACK_DRY_RUN=1` prints the asset URL, checksum URL and target path, then exits 0 without any network access.
- Next steps include `snapback config` and FUSE guidance (`macFUSE` on darwin, `fuse3` on linux) worded as instructions, not actions.

## T4 — Download and SHA-256 verification against httptest

- Files: `test/installer/install_test.go`, `install.sh` (download via curl or wget; verify via `sha256sum` or `shasum -a 256`; extract with `tar`).
- With a good checksum, the binary is installed to `SNAPBACK_INSTALL_DIR` and is executable.
- With a bad checksum, or the asset missing from `checksums.txt`, the script exits non-zero with `checksum` on stderr and nothing is written to the install dir.

## T5 — Install directory fallback

- Files: `test/installer/install_test.go`, `install.sh`.
- With `SNAPBACK_INSTALL_DIR` unset, the target is `$HOME/.local/bin`, created if absent.
- A warning is printed when that dir is not on `PATH`.

## T6 — Static safety checks on the script text

- Files: `test/installer/static_test.go`, `install.sh`.
- No package-manager install invocation (`brew`, `apt`, `apt-get`, `dnf`, `yum`, `opkg`, `pacman`, `apk`, `port` followed by `install`/`add` outside an `echo`/`printf` line).
- The default base URL host is `example.invalid`.
- The shebang is `#!/bin/sh`, `set -eu` is present, and there are no bashisms (`[[`, `function `, `local -`, `$'`, `<<<`).

## T7 — shellcheck when available

- Files: `test/installer/shellcheck_test.go`.
- When `shellcheck` is on PATH, run `shellcheck -s sh install.sh` and require exit 0.
- Otherwise call `t.Skip` (recorded as a skip, not a pass; CI runs shellcheck via S1-02).
