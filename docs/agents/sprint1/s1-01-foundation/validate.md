---
type: validate
story: S1-01
---

# S1-01 validate — PASS/FAIL rubric

Run every command from the repo root `/Users/adeelahmad/work/snapback`. A row PASSES only when the literal expected output matches; anything else is FAIL and cites the rule it violates (`docs/agents/sprint1/standards.md`).

## Pre-flight

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| Toolchain resolvable | `go version` | `go version go1.27.1 <os>/<arch>` (via `GOTOOLCHAIN=auto` from `go.mod`) | standards.md "Toolchain pinning"; README.md §18 pinned toolchain |
| Gate tools present | `command -v goimports golangci-lint govulncheck` | three paths printed, exit 0 | standards.md "BLOCKING — install before dispatching any GREEN or FINAL gate worker" |
| Scope | `git diff --name-only master -- . ':!docs'` | only `go.mod`, `.gitignore`, `cmd/snapback/*`, `internal/version/*` | stories.md S1-01 owned files; standards.md honesty gate (no feature code) |

## T1 — Module and ignore file

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| Module path | `go list -m` | `github.com/adeelahmad/snapback` | README.md §2 (named `snapback`) |
| Toolchain pin | `grep -E '^(go\|toolchain) ' go.mod` | `go 1.27` and `toolchain go1.27.1` | standards.md "Toolchain pinning"; README.md §18 |
| Stdlib only | `grep -c '^require' go.mod; test ! -e go.sum && echo nosum` | `0` then `nosum` | stories.md S1-01 constraint "Standard library only" |
| Ignore entries | `grep -xE 'coverage.out\|dist/' .gitignore` | both lines printed | stories.md S1-01 intent |

## T2 — `internal/version` vars and pure formatter

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| Unit tests | `go test -race -run 'TestFormat\|TestDefaults\|TestStringUsesPackageVars' ./internal/version/` | `ok  	github.com/adeelahmad/snapback/internal/version` | standards.md Testing (table-driven, `-race`) |
| Vars, not consts | `grep -E '^var (Version\|Commit\|Target)' internal/version/version.go \| wc -l` | `3` (or one `var (...)` block with all three) | stories.md failure scenario: `-X` silently ignored |

## T3 — Command dispatch (`run`) with injectable streams

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| Dispatch tests | `go test -race -run 'TestRunVersion\|TestRunUsageErrors' ./cmd/snapback/` | `ok  	github.com/adeelahmad/snapback/cmd/snapback` | stories.md failure scenarios (unknown exits 0; output on stderr) |
| Function size | `golangci-lint run ./cmd/... ./internal/...` | no output, exit 0 | standards.md "Functions small (<50 lines)"; Lint gate |

## T4 — Thin `main` and ldflags binary test

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| Binary tests | `go test -race -run 'TestBinary' ./cmd/snapback/` | `ok  	github.com/adeelahmad/snapback/cmd/snapback` | stories.md success scenario (ldflags into `t.TempDir()`) |
| Manual ldflags smoke | `CGO_ENABLED=0 go build -o /tmp/sb -ldflags "-X github.com/adeelahmad/snapback/internal/version.Version=v0.0.1 -X github.com/adeelahmad/snapback/internal/version.Commit=c0ffee -X github.com/adeelahmad/snapback/internal/version.Target=linux/amd64" ./cmd/snapback && /tmp/sb version; echo "exit=$?"` | `snapback v0.0.1 (commit c0ffee, target linux/amd64)` then `exit=0` | stories.md failure scenario (ldflags path mismatch); S1-06 contract |
| Usage exit | `/tmp/sb bogus; echo "exit=$?"` | stderr `usage: snapback version`, then `exit=2` | stories.md failure scenario (unknown subcommand exits 0) |
| Thin main | `grep -c . cmd/snapback/main.go` | <= 10 non-empty lines; body is only `os.Exit(run(...))` | standards.md coverage >= 80% (logic in testable `run`) |

## Final sign-off

Run the standards matrix verbatim; every line must exit 0.

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| gofmt | `test -z "$(gofmt -l .)"` | exit 0, no output | `~/.claude/rules/golang/coding-style.md` |
| goimports | `test -z "$(goimports -l .)"` | exit 0, no output | `~/.claude/rules/golang/coding-style.md` |
| Build | `CGO_ENABLED=0 go build ./...` | exit 0, no output | README.md §2, §5, §17 |
| Vet | `go vet ./...` | exit 0, no output | README.md §18 |
| Lint | `golangci-lint run` | exit 0, no output | README.md §18 |
| Race | `go test -race ./...` | every package `ok` | README.md §18; golang/testing.md |
| Coverage | `go test -race -covermode=atomic -coverprofile=coverage.out ./... && go tool cover -func=coverage.out \| awk '/^total:/{gsub("%","",$NF); if ($NF+0 < 80) exit 1; print "coverage " $NF "% OK (>=80%)"}'` | `coverage NN.N% OK (>=80%)` | `~/.claude/rules/common/testing.md` (80% minimum) |
| Vulns | `govulncheck ./...` | `No vulnerabilities found.` | README.md §18; standards.md toolchain note |
| Honesty | `grep -rniE 'production-ready\|cross-platform\|static\|finder-integrated' cmd internal` | no output | standards.md honesty gate |
| Plan ticked | `grep -c '^- \[ \]' docs/agents/sprint1/s1-01-foundation/plan.md` | `0` after execution | standards.md TDD workflow |
