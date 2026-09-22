---
type: validate
story: S2-02
---

# S2-02 validate — PASS/FAIL rubric

Run every command from the repo root `/Users/adeelahmad/work/snapback`. A row PASSES only when the literal expected output matches. Anything else FAILS and cites the rule it violates (`docs/agents/sprint2/standards.md`, `docs/agents/sprint2/stories.md` S2-02, SPEC.md). Intermediate GREENs (T1 to T5) run with `GATE_RUN_MATRIX=0` plus the package-scoped rows below. T6 runs Final sign-off in full (M-005).

## Pre-flight

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| Toolchain | `go version` | `go version go1.27.1 <os>/<arch>` | standards.md "Detected stack" (toolchain pin) |
| Gate tools present | `command -v goimports golangci-lint govulncheck actionlint shellcheck mkdocs goreleaser` | seven paths printed, exit 0 | standards.md "Cross-cutting gates" |
| Catalog contract source | `grep -cE 'Lookup\(parent uint64, name string\) \(ino uint64, isDir bool, found bool\)\|ReadDir\(dir uint64\) \(names \[\]string, found bool\)\|Readlink\(ino uint64\) \(target string, found bool\)' docs/agents/sprint2/s2-01-mount-adapter/tasks.md` | `3` (if `internal/mount/mount.go` has merged, the same three signatures appear there too) | stories.md S2-01/S2-02 Connections (structural contract) |
| Scope | `git diff --name-only master -- . ':!docs'` | only paths under `internal/projection/` | stories.md S2-02 owned files; M-012 |

## T1 — Package skeleton and dependency guard

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| Deps test | `go test -race -run 'TestDepsStdlibOnly' ./internal/projection/` | `ok  	github.com/adeelahmad/snapback/internal/projection` | stories.md S2-02 success scenario (deps test, M-002) |
| No FUSE, no mount | `go list -deps ./internal/projection \| grep -cE 'hanwen/go-fuse\|internal/mount'` | `0` | SPEC.md §4 module table ("independent of any FUSE library") |
| Stdlib only | `go list -f '{{join .Imports "\n"}}' ./internal/projection \| grep -c '\.'` | `0` | stories.md S2-02 constraint "Standard library only" |

## T2 — Name validation

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| Name tests | `go test -race -run 'TestValidateName' ./internal/projection/` | `ok  	github.com/adeelahmad/snapback/internal/projection` | stories.md S2-02 constraint (rejected name classes) |
| Package vet/lint | `go vet ./internal/projection/ && golangci-lint run ./internal/projection/...` | exit 0, no output | standards.md Lint; M-004 (lint covers test files) |

## T3 — Spec types and Build validation

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| Build tests | `go test -race -run 'TestBuild' ./internal/projection/` | `ok  	github.com/adeelahmad/snapback/internal/projection` | stories.md S2-02 constraints (wrapped errors, duplicate siblings) |
| Wrapped sentinels | `grep -nE 'ErrInvalidName\|ErrDuplicateName' internal/projection/*.go \| grep -c 'errors.New'` | `2` | tasks.md decisions (exported sentinels, `%w` wrapping) |
| Package vet/lint | `go vet ./internal/projection/ && golangci-lint run ./internal/projection/...` | exit 0, no output | standards.md Lint |

## T4 — Lookup, inode identity and generation independence

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| Lookup tests | `go test -race -count=3 -run 'TestLookup\|TestInodesStable\|TestBuildIsDeterministic\|TestGenerationsAreIndependent\|TestConcurrentReadsDuringRebuild' ./internal/projection/` | `ok  	github.com/adeelahmad/snapback/internal/projection` | SPEC.md §6 (immutable generations), §7 (stable inode identities); stories.md S2-02 failure scenarios |
| No locks needed | `grep -c 'sync\.' internal/projection/catalog.go` | `0` (a generation is read-only after Build) | tasks.md decisions (immutability) |
| Package vet/lint | `go vet ./internal/projection/ && golangci-lint run ./internal/projection/...` | exit 0, no output | standards.md Lint |

## T5 — Ordered directory listing

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| ReadDir tests | `go test -race -count=3 -run 'TestReadDir' ./internal/projection/` | `ok  	github.com/adeelahmad/snapback/internal/projection` | stories.md S2-02 failure scenario (nondeterministic listing) |
| Package vet/lint | `go vet ./internal/projection/ && golangci-lint run ./internal/projection/...` | exit 0, no output | standards.md Lint |

## T6 — Readlink, the catalog contract, and the full matrix

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| Readlink and contract tests | `go test -race -run 'TestReadlink\|TestGenerationsIndependentTargets\|TestGenerationSatisfiesCatalogContract' ./internal/projection/` | `ok  	github.com/adeelahmad/snapback/internal/projection` | stories.md S2-02 success scenarios; S2-01 tasks.md contract |
| No path cleaning | `grep -cE 'filepath\.(Clean\|EvalSymlinks\|Abs)\|path\.Clean' internal/projection/*.go` | `0` for every file | stories.md S2-02 constraint (targets stored byte-for-byte) |
| No mount import | `grep -rc 'internal/mount' internal/projection/ \| grep -v ':0$'` | no output | stories.md S2-02 ("satisfy structurally WITHOUT importing mount") |

## Final sign-off

Run the standards matrix verbatim (`docs/agents/sprint2/standards.md` § Cross-cutting gates). Every line must exit 0, and the package-specific rows must pass as well.

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| gofmt | `test -z "$(gofmt -l .)"` | exit 0, no output | `~/.claude/rules/golang/coding-style.md` |
| goimports | `test -z "$(goimports -l .)"` | exit 0, no output | `~/.claude/rules/golang/coding-style.md` |
| Build | `CGO_ENABLED=0 go build ./...` | exit 0, no output | SPEC.md §5 (static builds) |
| Vet | `go vet ./...` | exit 0, no output | SPEC.md §18 |
| Lint | `golangci-lint run` | exit 0, no output | SPEC.md §18; M-004 |
| Race | `go test -race ./...` | every package `ok` | SPEC.md §18, §20 |
| Coverage (total) | `go test -race -covermode=atomic -coverprofile=coverage.out ./... && go tool cover -func=coverage.out \| awk '/^total:/{gsub("%","",$NF); if ($NF+0 < 80) exit 1; print "coverage " $NF "% OK (>=80%)"}'` | `coverage NN.N% OK (>=80%)` | `~/.claude/rules/common/testing.md` |
| Coverage (package) | `go test -race -covermode=atomic -coverprofile=/tmp/proj.out ./internal/projection/ && go tool cover -func=/tmp/proj.out \| awk '/^total:/{gsub("%","",$NF); if ($NF+0 < 80) exit 1; print "projection " $NF "% OK"}'` | `projection NN.N% OK` | stories.md S2-02 success scenario ("Package coverage is at least 80%") |
| Vulns | `govulncheck ./...` | `No vulnerabilities found.` | SPEC.md §18 |
| actionlint | `actionlint` | exit 0, no output | standards.md Cross-cutting gates |
| shellcheck | `shellcheck -s sh install.sh` | exit 0, no output | standards.md Cross-cutting gates |
| Docs build | `mkdocs build --strict --site-dir site` | exit 0 | standards.md Cross-cutting gates |
| Release config | `goreleaser check` | exit 0 | standards.md Cross-cutting gates |
| File/function size | `wc -l internal/projection/*.go \| awk '$2!="total" && $1>400{print; bad=1} END{exit bad}'` | exit 0, no output | stories.md S2-02 constraint (files under 400 lines; functions under 50 lines are checked by review) |
| Honesty | `grep -rniE 'production-ready\|cross-platform\|static\|finder-integrated' internal/projection` | no output | standards.md honesty gate; SPEC.md §1 |
| Plan ticked | `grep -c '^- \[ \]' docs/agents/sprint2/s2-02-projection/plan.md` | `0` after execution | standards.md TDD workflow |
