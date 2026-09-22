---
type: validate
story: S2-08
---

# S2-08 validate — PASS/FAIL rubric

Run every command from the repo root `/Users/adeelahmad/work/snapback`. A row PASSES only when the literal expected output matches; anything else is FAIL and cites the rule it violates. Intermediate GREENs (T1-T5) run package-scoped with `GATE_RUN_MATRIX=0`; only T6's GREEN runs the full matrix (M-005).

## Pre-flight

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| Dependency merged | `go list ./internal/compat/resticfx/` | `github.com/adeelahmad/snapback/internal/compat/resticfx` | sprint2/plan.md wave 2 gate (S2-03 merged green) |
| Toolchain | `go version` | `go1.27.1` in the output | standards.md toolchain pinning |
| Gate tools | `command -v goimports golangci-lint govulncheck` | three paths, exit 0 | standards.md cross-cutting gates |
| Scope | `git diff --name-only master -- . ':!docs/agents'` | only paths under `internal/compat/latency/`, `tools/stage1latency/`, and (after exit evidence) `docs/reports/stage1/latency.json` | stories.md S2-08 owned files; M-012 |
| No secret on disk in repo | `git ls-files -o --exclude-standard \| grep -iE 'pass(word)?' ` | no output | stories.md S2-08 constraint (password only in 0600 temp file); SPEC §12 |

## T1 — Remote guard

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| Guard tests | `go test -race -run 'TestCheckRemote\|TestRemoteFromEnv\|TestRepoSpec' ./internal/compat/latency/` | `ok  	github.com/adeelahmad/snapback/internal/compat/latency` | stories.md S2-08 constraint (exact remote, human decision) |

## T2 — Timing stats and result schema

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| Schema tests | `go test -race -run 'TestSummarize\|TestEncode\|TestTimestamp\|TestNoThreshold' ./internal/compat/latency/` | `ok  	...internal/compat/latency` | stories.md S2-08 (no threshold or verdict; median needs samples) |
| No threshold in code | `grep -rniE 'threshold\|verdict\|acceptable\|no.?go' internal/compat/latency tools/stage1latency --include='*.go' \| grep -v _test.go` | no output | human decision: go/no-go is the human's |

## T3 — Disposable secrets and scratch dirs

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| Scratch tests | `go test -race -run 'TestNewScratch\|TestScratch' ./internal/compat/latency/` | `ok  	...internal/compat/latency` | SPEC §12 (password handling); stories.md S2-08 |

## T4 — Cleanup and deletion verification

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| Cleanup tests | `go test -race -run 'TestVerifyDeleted\|TestPurge\|TestCheckEmpty' ./internal/compat/latency/` | `ok  	...internal/compat/latency` | stories.md S2-08 (delete afterwards; refuse non-empty remote) |
| Purge target constant | `grep -rn '"purge"' internal/compat/latency --include='*.go' \| grep -v _test.go` | every hit builds its target from `AllowedRemote` | stories.md S2-08 failure scenario (purge from user input) |

## T5 — Orchestrated run

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| Run tests | `go test -race -run 'TestRun' ./internal/compat/latency/` | `ok  	...internal/compat/latency` | stories.md S2-08 success scenarios; SPEC §11 four measurements |
| Args arrays only | `grep -rnE 'sh"\|"-c"\|bash' internal/compat/latency tools/stage1latency --include='*.go' \| grep -v _test.go` | no output | SPEC §5 (`exec.CommandContext` argument arrays) |
| Package coverage | `go test -race -cover ./internal/compat/latency/` | `coverage: NN.N%` with NN >= 80 | common/testing.md 80% |

## T6 — Thin CLI and full matrix

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| CLI tests | `go test -race ./tools/stage1latency/` | `ok  	github.com/adeelahmad/snapback/tools/stage1latency` | stories.md S2-08 (`main` thin) |
| Refusal smoke (no network) | `SNAPBACK_RCLONE_REMOTE=gdrive:other go run ./tools/stage1latency -out /tmp/x.json; echo "exit=$?"` | stderr names `gdrive:snapback-stage1`, then `exit=2`, and `/tmp/x.json` not created | human constraint: that exact remote only |

## Exit evidence (orchestrator-run, network, not CI)

Run once on this Mac after T6 merges green, with macFUSE, `restic` 0.19.0 and `rclone` configured for `gdrive:`. Never run by a sub-agent or CI.

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| Remote empty first | `rclone lsf gdrive:snapback-stage1; echo "exit=$?"` | `directory not found` error (exit 3) or no output | stories.md S2-08 (only its own repo) |
| Real run | `SNAPBACK_RCLONE_REMOTE=gdrive:snapback-stage1 go run ./tools/stage1latency -out docs/reports/stage1/latency.json; echo "exit=$?"` | `exit=0` | SPEC §11 Stage 1 measurements; §21 risk 1 |
| Four numbers | `jq -r '.measurements \| keys[]' docs/reports/stage1/latency.json` | `cold_first_file_read`, `cold_listing`, `warm_listing_after_restart`, `warm_prewarmed_listing` | SPEC §11 |
| Samples real | `jq '[.measurements[].samples_ms \| length] \| min' docs/reports/stage1/latency.json` | `1` or more; warm entries have `Samples` entries | stories.md failure scenario (median from one sample) |
| Deleted | `jq .remote_deleted docs/reports/stage1/latency.json` | `true` | human constraint: delete afterwards |
| Deletion re-verified | `rclone lsf gdrive:snapback-stage1; echo "exit=$?"` | `directory not found` error (exit 3) | human constraint: verify deletion |
| No threshold | `jq -r '[paths \| map(tostring) \| join(".")] \| .[]' docs/reports/stage1/latency.json \| grep -iE 'threshold\|verdict\|pass\|fail\|acceptable'` | no output | human decision (go/no-go) |
| No password | `grep -ciE 'password' docs/reports/stage1/latency.json` | `0` | SPEC §12 |

## Final sign-off

Run the standards matrix verbatim on T6's GREEN; every line must exit 0.

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| gofmt | `test -z "$(gofmt -l .)"` | exit 0, no output | golang/coding-style.md |
| goimports | `test -z "$(goimports -l .)"` | exit 0, no output | golang/coding-style.md |
| Build | `CGO_ENABLED=0 go build ./...` | exit 0, no output | SPEC §5 |
| Vet | `go vet ./...` | exit 0, no output | SPEC §18 |
| Lint | `golangci-lint run` | exit 0, no output | SPEC §18; M-004 |
| Race | `go test -race ./...` | every package `ok` | SPEC §18, §20 |
| Coverage | `go test -race -covermode=atomic -coverprofile=coverage.out ./... && go tool cover -func=coverage.out \| awk '/^total:/{gsub("%","",$NF); if ($NF+0 < 80) exit 1; print "coverage " $NF "% OK (>=80%)"}'` | `coverage NN.N% OK (>=80%)` | common/testing.md |
| Vulns | `govulncheck ./...` | `No vulnerabilities found.` | SPEC §18 |
| Honesty | `grep -rniE 'production-ready\|cross-platform\|static\|finder-integrated' cmd internal tools` | no output | standards.md honesty gate |
| Default CI never touches network | `go test ./... 2>&1 \| grep -ci rclone` | `0` | standards.md Stage 1 gating (network tests never in default matrix) |
| Plan ticked | `grep -c '^- \[ \]' docs/agents/sprint2/s2-08-latency/plan.md` | `0` after execution | standards.md TDD workflow |
| Exit evidence | the Exit evidence table above | every row PASS, `latency.json` committed | sprint2/plan.md latency step |
