---
type: output
story: C2
---

## C2/T1 · attempt 1 · red-worker · 2026-09-22T03:19:22Z
status: ok
### Summary
Wrote test/makefile/makefile_test.go (package makefile_test, commit 918f32f, 1 file) pinning a repo-root Makefile: all 15 targets defined and in .PHONY, help as the default target, `make -n build` (CGO_ENABLED=0, -o bin/snapback, ./cmd/snapback, -ldflags -X for internal/version Version/Commit/Target), `make -n install PREFIX=/tmp/x` with and without DESTDIR (install -m 0755), `make -n uninstall` (rm), `make -n ci` containing every standards.md cross-cutting gate plus the 80 coverage threshold, an exported overridable `GOTOOLCHAIN ?= auto`, and CONTRIBUTING.md naming `make ci` and `make install`. Every Makefile test t.Fatal's "repo-root Makefile missing" before any make call; the make-on-PATH skip comes after that check and names the prerequisite. Only `make -n` is used. No shim needed. Sanity check: a throwaway Makefile (never committed, deleted) made all seven Makefile tests pass, so the assertions can be satisfied. Rule parsing accepts multi-target lines (`vet lint:`), which is a little looser than a strict `^name:` per line.
### Result
| Check | Status | Detail |
|---|---|---|
| `test/makefile/makefile_test.go::TestMakefileTargets` | FAIL | repo-root Makefile missing |
| `test/makefile/makefile_test.go::TestMakefileHelpIsDefaultTarget` | FAIL | repo-root Makefile missing |
| `test/makefile/makefile_test.go::TestMakeBuildDryRun` | FAIL | repo-root Makefile missing |
| `test/makefile/makefile_test.go::TestMakeInstallDryRun` (prefix, destdir) | FAIL | repo-root Makefile missing |
| `test/makefile/makefile_test.go::TestMakeUninstallDryRun` | FAIL | repo-root Makefile missing |
| `test/makefile/makefile_test.go::TestMakeCIRunsEveryGate` | FAIL | repo-root Makefile missing |
| `test/makefile/makefile_test.go::TestMakefileGoToolchainDefault` | FAIL | repo-root Makefile missing |
| `test/makefile/makefile_test.go::TestContributingMentionsMake` | FAIL | CONTRIBUTING.md lacks "make ci" / "make install" |
| golangci-lint ./test/makefile/... ; go vet ./... | PASS | 0 issues |
### Next
GREEN: add the repo-root Makefile (targets above, exported GOTOOLCHAIN ?= auto, PREFIX/DESTDIR install) and mention `make ci` / `make install` in CONTRIBUTING.md. No production symbols to scaffold.

## C2/T1 · attempt 1 · green-worker · 2026-09-22T03:25:00Z
status: ok
### Summary
Added a repo-root GNU `Makefile` (commit 0c2505c, works on make 3.81): `help` is the first/default target and lists targets from `##` comments; `build` runs `CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X …version.Version/Commit/Target"` into `bin/snapback` (same -X set as .goreleaser.yaml; VERSION from `git describe`, COMMIT from `git rev-parse --short`); `install` (depends on build) runs `install -d` + `install -m 0755` into `$(DESTDIR)$(PREFIX)/bin`, PREFIX ?= /usr/local; `uninstall`, `test`, `cover` (80% total check), `lint`, `fmt`, `fmt-check`, `vet`, `vuln`, `docs`, `release-check`, `clean`. `ci` runs every standards.md cross-cutting gate in order as prerequisites; three small extra phony targets (`compile`, `lint-actions`, `lint-shell`) carry the gates that have no human-requested target, so `ci` stays a plain prerequisite list. `export GOTOOLCHAIN ?= auto`. CONTRIBUTING.md gains a "Make targets" section naming `make build`, `make install` and `make ci`; `.gitignore` gains `/bin/`. Note: the first CONTRIBUTING draft said "static", which test/community's honesty test bans; I reworded it and the Makefile help text to "CGO off".
### Result
| Check | Status | Detail |
|---|---|---|
| `test/makefile/makefile_test.go::TestMakefileTargets` | PASS | green |
| `test/makefile/makefile_test.go::TestMakefileHelpIsDefaultTarget` | PASS | green |
| `test/makefile/makefile_test.go::TestMakeBuildDryRun` | PASS | green |
| `test/makefile/makefile_test.go::TestMakeInstallDryRun` (prefix, destdir) | PASS | green |
| `test/makefile/makefile_test.go::TestMakeUninstallDryRun` | PASS | green |
| `test/makefile/makefile_test.go::TestMakeCIRunsEveryGate` | PASS | green |
| `test/makefile/makefile_test.go::TestMakefileGoToolchainDefault` | PASS | green |
| `test/makefile/makefile_test.go::TestContributingMentionsMake` | PASS | green |
| `go test ./test/...` | PASS | all packages incl. test/community |
| `make build` (real) + `make install/uninstall DESTDIR=<scratch>` | PASS | binary 0755, `snapback version` prints ldflags values; bin/ deleted before commit |
| `make ci` (full standards matrix) | PASS | exit 0; coverage 86.6% total; goreleaser check validated |
### Next
Structural review, then merge chain2/c2-make into stage-1.
