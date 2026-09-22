---
type: validate
story: S1-02
---

# S1-02 validate — PASS/FAIL rubric

All commands run from `/Users/adeelahmad/work/snapback`. PASS requires every row's expected output. Any other output is FAIL.

## Pre-flight

| Check | Command | Expected | FAIL cites |
| --- | --- | --- | --- |
| S1-01 landed | `test -f go.mod && head -1 go.mod` | `module github.com/adeelahmad/snapback` | stories.md S1-02 Connections (depends on S1-01) |
| Only owned files changed | `git diff --name-only master... \| grep -vE '^(\.github/workflows/ci\.yml\|\.golangci\.yml\|test/ci/)'` | no output | stories.md S1-02 Owned files |
| No go.mod/go.sum edit | `git diff --name-only master... -- go.mod go.sum` | no output | stories.md S1-01 owns go.mod |

## T1 — Core CI workflow (triggers, toolchain, gates)

| Check | Command | Expected | FAIL cites |
| --- | --- | --- | --- |
| T1 tests | `go test -race -run 'TestCI(WorkflowExists\|Triggers\|SetupGo\|FormatCheck\|BuildAndVet\|TestUses\|CoverageThreshold\|UploadsCoverage\|RunsGovulncheck\|Shellcheck\|NoContinue\|ActionsArePinned\|NoMount\|NoDependency)' -v ./test/ci/` | 14 `--- PASS` lines, final `ok  	github.com/adeelahmad/snapback/test/ci` | README.md §18 CI requirements; standards.md gate matrix |
| No continue-on-error | `grep -c 'continue-on-error: *true' .github/workflows/ci.yml` | `0` | stories.md S1-02 Failure scenarios (lint with continue-on-error) |
| No hard-coded Go version | `grep -c 'go-version:' .github/workflows/ci.yml` | `0` | stories.md S1-02 Constraints (toolchain from go.mod) |

## T2 — Cross-compile matrix with unverified labelling

| Check | Command | Expected | FAIL cites |
| --- | --- | --- | --- |
| T2 tests | `go test -race -run 'TestMatrix' -v ./test/ci/` | 5 `--- PASS` lines, final `ok  	github.com/adeelahmad/snapback/test/ci` | stories.md S1-02 Failure scenarios (dropped target, unlabelled unverified) |
| Unverified count | `grep -c 'unverified' .github/workflows/ci.yml` | a number `>= 3` | stories.md S1-02 What's wanted |

## T3 — golangci-lint config, lint job pin, and actionlint check

| Check | Command | Expected | FAIL cites |
| --- | --- | --- | --- |
| T3 tests | `go test -race -run 'TestGolangci\|TestCIGolangci\|TestCIActionlint' -v ./test/ci/` | 2 `--- PASS` lines plus `--- PASS: TestCIActionlint` if actionlint installed, else `--- SKIP: TestCIActionlint` with `actionlint not on PATH`; final `ok` | stories.md S1-02 Success scenarios (skip, not pass, when absent) |
| Lint locally (if installed) | `golangci-lint run ./...` | exit 0, no output; if not installed record NOT RUN (standards.md: NOT INSTALLED) | README.md §18 staticcheck/golangci-lint |

## Final sign-off

| Check | Command | Expected | FAIL cites |
| --- | --- | --- | --- |
| Whole package | `go test -race -count=1 ./test/ci/...` | `ok  	github.com/adeelahmad/snapback/test/ci` | common/testing.md |
| Test count | `go test -list '.*' ./test/ci/ \| grep -c '^Test'` | `22` | this plan.md (22 checkboxes) |
| Format | `test -z "$(gofmt -l test/ci)" && echo clean` | `clean` | golang/coding-style.md |
| Vet | `go vet ./test/ci/...` | no output, exit 0 | README.md §18 |
| No suppressions | `grep -rnE 't\.Skip\(' test/ci \| grep -v actionlint` | no output | standards.md (zero suppressions) |
| Exit evidence (orchestrator, post-merge) | `gh run list --workflow ci.yml --branch master --limit 1 --json conclusion -q '.[0].conclusion'` | `success` | stories.md S1-02 Exit evidence; §22 row 0 |
