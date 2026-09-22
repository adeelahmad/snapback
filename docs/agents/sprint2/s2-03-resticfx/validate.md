---
type: validate
story: S2-03
---

# S2-03 validate — PASS/FAIL rubric

Run every command from the repo root `/Users/adeelahmad/work/snapback`. A row PASSES only when the literal expected output matches; anything else is FAIL and cites the rule it violates. `PKG=./internal/compat/resticfx/`. Intermediate GREENs (T1-T6) run only their rows plus the package-scoped matrix with `GATE_RUN_MATRIX=0` (M-005); T7 runs Final sign-off.

## Pre-flight

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| Toolchain | `go version` | `go version go1.27.1 <os>/<arch>` | standards.md "Detected stack"; `go.mod` toolchain pin |
| Restic pinned | `restic version` | first line starts `restic 0.19.0 ` | SPEC.md §5 (verify path template against pinned version); stories.md S2-03 |
| FUSE present (macOS) | `test -d /Library/Filesystems/macfuse.fs && echo ok` | `ok` (else integration run is "implemented but not tested here") | standards.md "Sprint 2 / Stage 1 integration-test gating" |
| Gate tools | `command -v goimports golangci-lint govulncheck actionlint shellcheck mkdocs goreleaser` | seven paths, exit 0 | standards.md "Cross-cutting gates" |
| Scope | `git diff --name-only master -- . ':!docs/agents'` | only `internal/compat/resticfx/*` and `docs/reports/stage1/pathtemplate-*.json` | stories.md S2-03 owned files; M-012 |
| No secrets committed | `git diff master -- . ':!docs/agents' \| grep -iE 'password[^_-]*[:=][[:space:]]*"[^"]+"\|RESTIC_PASSWORD='` | no output | SPEC.md §12; stories.md S2-03 constraints |

## T1 — version parsing and repository guard

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| Unit tests | `go test -race -run 'TestParseResticVersion\|TestCheckPinnedVersion\|TestGuardRepo' ./internal/compat/resticfx/` | `ok  	github.com/adeelahmad/snapback/internal/compat/resticfx` | stories.md S2-03 success scenarios (version, guard table) |
| Pinned constant | `grep -n 'PinnedResticVersion = "0.19.0"' internal/compat/resticfx/version.go` | one line | tasks.md decisions |

## T2 — argument builders and unmount command

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| Unit tests | `go test -race -run 'TestInitArgs\|TestBackupArgs\|TestSnapshotsAndLsArgs\|TestMountArgs\|TestArgsNeverContainPassword\|TestUnmountCommand' ./internal/compat/resticfx/` | `ok  	github.com/adeelahmad/snapback/internal/compat/resticfx` | SPEC.md §5 (argv arrays, path template); §12 (no secrets in args) |
| Path template literal | `grep -c '"ids/%I"' internal/compat/resticfx/args.go` | `1` or more | SPEC.md §5 |

## T3 — snapshot JSON and ids/ observation

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| Unit tests | `go test -race -run 'TestParseSnapshots\|TestCheckObservedIDs' ./internal/compat/resticfx/` | `ok  	github.com/adeelahmad/snapback/internal/compat/resticfx` | SPEC.md §5, §6 (full IDs only); stories.md S2-03 failure scenario (short ID passes) |
| No short_id use | `grep -n 'ShortID\|short_id' internal/compat/resticfx/snapshots.go` | no output (field ignored, not decoded) | SPEC.md §5 "Never identify a snapshot by a short prefix alone" |

## T4 — generated tree and password file

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| Unit tests | `go test -race -run 'TestWriteTree\|TestNewPasswordFile' ./internal/compat/resticfx/` | `ok  	github.com/adeelahmad/snapback/internal/compat/resticfx` | SPEC.md §20 (generated disposable fixtures); §12 |
| crypto/rand | `grep -n '"crypto/rand"' internal/compat/resticfx/password.go` | one line | SPEC.md §12 |

## T5 — runner and fixture lifecycle

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| Unit tests | `go test -race -run 'TestExecRunner\|TestNewFixture\|TestFixture\|TestToolVersions' ./internal/compat/resticfx/` | `ok  	github.com/adeelahmad/snapback/internal/compat/resticfx` | stories.md S2-03 (fake runner; guard; destroy only own repo) |
| Only CommandContext | `grep -rnE 'exec\.Command\(\|"sh"\|"/bin/sh"\|"bash"\|-c"' internal/compat/resticfx --include='*.go'` | no output | SPEC.md §5; standards.md "Active rules" |
| No restic internals | `go list -deps ./internal/compat/resticfx \| grep -E 'restic/restic\|go-fuse'` | no output | SPEC.md §5; stories.md S2-03 (stdlib only) |

## T6 — evidence JSON and prerequisite probe

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| Unit tests | `go test -race -run 'TestEvidence\|TestWriteEvidenceFileName\|TestMissingPrerequisite' ./internal/compat/resticfx/` | `ok  	github.com/adeelahmad/snapback/internal/compat/resticfx` | stories.md S2-03 (`SNAPBACK_EVIDENCE_DIR` contract; skip names prerequisite) |
| Env var name | `grep -n '"SNAPBACK_EVIDENCE_DIR"' internal/compat/resticfx/evidence.go` | one line | stories.md S2-03 connections (contract used by S2-04, S2-05, S2-06, S2-07) |

## T7 — mount supervisor and path-template integration

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| Mount unit tests | `go test -race -run 'TestMount' ./internal/compat/resticfx/` | `ok  	github.com/adeelahmad/snapback/internal/compat/resticfx` | SPEC.md §5 (no timeout on mount; graceful stop) |
| Build tag | `head -1 internal/compat/resticfx/pathtemplate_integration_test.go` | `//go:build integration` | standards.md integration-test gating |
| Skip names prerequisite | `env -u SNAPBACK_FUSE_TESTS go test -race -count=1 -tags=integration -run TestPathTemplateIntegration -v ./internal/compat/resticfx/ \| grep -A2 -- '--- SKIP' \| grep -c SNAPBACK_FUSE_TESTS` | `1` or more (skip message names the missing prerequisite) | standards.md "Skip discipline" |
| Integration run (macOS) | `SNAPBACK_FUSE_TESTS=1 SNAPBACK_EVIDENCE_DIR=docs/reports/stage1 go test -race -tags=integration -run TestPathTemplateIntegration -v ./internal/compat/resticfx/` | `--- PASS: TestPathTemplateIntegration` and `ok` | SPEC.md §5, §20; stories.md sprint demo 4 |
| Evidence content | `jq -r '[.restic_version,.result,(.snapshot_id\|test("^[0-9a-f]{64}$")),(.snapshot_id==.observed_dir),.path_template]\|@tsv' docs/reports/stage1/pathtemplate-darwin.json` | `0.19.0	pass	true	true	ids/%I` | stories.md S2-03 success scenarios; SPEC.md §5 |
| No leaked mount | `mount \| grep -F "${TMPDIR%/}" \| grep -c 'TestPathTemplateIntegration'` (the test's mountpoint is a `t.TempDir()` subdirectory, whose path contains the test name under `$TMPDIR`; other macFUSE mounts on the host are ignored) | `0` after the run (cleanup unmounted even on failure) | stories.md S2-03 failure scenario (leaked mount) |
| Linux evidence | `test -s docs/reports/stage1/pathtemplate-linux.json && jq -r .result docs/reports/stage1/pathtemplate-linux.json` | `pass` (orchestrator copies it from the green S2-05 `fuse-linux` artifact; until then this row is PENDING, not FAIL) | stories.md S2-03 success scenarios; sprint2 plan.md "Linux evidence" |

## Final sign-off

Run on T7's GREEN only. The standards matrix, verbatim; every line must exit 0.

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| gofmt | `test -z "$(gofmt -l .)"` | exit 0, no output | standards.md "Cross-cutting gates" |
| goimports | `test -z "$(goimports -l .)"` | exit 0, no output | standards.md "Cross-cutting gates" |
| Build | `CGO_ENABLED=0 go build ./...` | exit 0 | SPEC.md §5, §17 |
| Vet | `go vet ./... && go vet -tags=integration ./internal/compat/resticfx/` | exit 0 | SPEC.md §18 |
| Lint | `golangci-lint run && golangci-lint run --build-tags=integration ./internal/compat/resticfx/...` | exit 0, no issues (test files included, M-004) | SPEC.md §18 |
| Race | `go test -race ./...` | every package `ok` | SPEC.md §18, §20 |
| Coverage | `go test -race -covermode=atomic -coverprofile=coverage.out ./... && go tool cover -func=coverage.out \| awk '/^total:/{gsub("%","",$NF); if ($NF+0 < 80) exit 1; print "coverage " $NF "% OK (>=80%)"}'` | `coverage NN.N% OK (>=80%)` | `~/.claude/rules/common/testing.md` |
| Vulns | `govulncheck ./...` | `No vulnerabilities found.` | SPEC.md §18 |
| actionlint | `actionlint` | exit 0 | standards.md "Cross-cutting gates" |
| shellcheck | `shellcheck -s sh install.sh` | exit 0 | standards.md "Cross-cutting gates" |
| Docs | `mkdocs build --strict --site-dir site` | exit 0 | SPEC.md §18 |
| Release config | `goreleaser check` | exit 0 | SPEC.md §17, §18 |
| Integration suite | `SNAPBACK_FUSE_TESTS=1 SNAPBACK_EVIDENCE_DIR=docs/reports/stage1 go test -race -tags=integration ./internal/compat/resticfx/` | `ok`; any SKIP is reported as "implemented but not tested here" | standards.md integration-test gating |
| Honesty | `grep -rniE 'production-ready\|cross-platform\|finder-integrated' internal/compat/resticfx` | no output | SPEC.md §1, §18 |
| Plan ticked | `grep -c '^- \[ \]' docs/agents/sprint2/s2-03-resticfx/plan.md` | `0` after execution | standards.md TDD workflow |
