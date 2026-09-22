---
type: validate
story: S2-07
---

# S2-07 validate — PASS/FAIL rubric

All commands run from `/Users/adeelahmad/work/snapback` (or the task worktree, M-001). PASS requires every row's expected output. Any other output is FAIL. Package import path: `github.com/adeelahmad/snapback/internal/compat/crawler`.

## Pre-flight

| Check | Command | Expected | FAIL cites |
| --- | --- | --- | --- |
| S2-01 Observer exists | `go doc ./internal/mount Observer` | prints a type declaration, exit 0 | stories.md S2-07 Connections (depends on S2-04, which depends on S2-01 `Observer`) |
| S2-02 projection exists | `go doc ./internal/projection` | package doc, exit 0 | stories.md S2-07 Connections (S2-02) |
| S2-04 adapter merged | `test -f internal/mount/gofuse/adapter.go && go doc ./internal/mount/gofuse Adapter` | a type declaration, exit 0 | sprint plan.md wave 3 ("S2-07: S2-04 merged green") |
| Default build green before start | `go build ./... && go vet ./...` | exit 0, no output | standards.md Cross-cutting gates |
| Only owned files changed | `git diff --name-only master... \| grep -vE '^(internal/compat/crawler/\|docs/reports/stage1/crawler-(darwin\|linux)\.json$)'` | no output | stories.md S2-07 Owned files; M-012 |
| No go.mod/go.sum edit | `git diff --name-only master... -- go.mod go.sum` | no output | stories.md S2-04 Constraints (S2-01 owns go.mod) |
| Crawler tools present locally | `for b in rg fd find rsync ls; do command -v "$b" >/dev/null && echo "$b ok" \|\| echo "$b MISSING"; done` | `ok` for each; a `MISSING` tool is allowed but its row must be `not-tested-here` in the evidence | stories.md S2-07 Constraints (missing tool skips that row only) |

## T1 — Catalog hit counter

| Check | Command | Expected | FAIL cites |
| --- | --- | --- | --- |
| T1 tests | `go test -race -count=1 -run 'TestCounter' -v ./internal/compat/crawler/` | 5 `--- PASS` lines, final `ok` | stories.md S2-07 Failure scenarios (counter not reset) |

## T2 — Tool table and binary resolution

| Check | Command | Expected | FAIL cites |
| --- | --- | --- | --- |
| T2 tests | `go test -race -count=1 -run 'TestToolTable\|TestToolArgs\|TestSearchTools\|TestRsyncArgs\|TestResolve' -v ./internal/compat/crawler/` | 8 `--- PASS` lines, final `ok` | stories.md S2-07 Success scenarios (tool table, fdfind/fd resolution) |
| No shell invocation | `grep -rnE '"(sh\|bash)"\|"-c"' internal/compat/crawler/*.go \| grep -v _test.go` | no output | SPEC.md §5 (argument arrays, never a shell) |

## T3 — Result schema, acceptance check and JSON writer

| Check | Command | Expected | FAIL cites |
| --- | --- | --- | --- |
| T3 tests | `go test -race -count=1 -run 'TestEvidence\|TestReport\|TestSkipped\|TestValidate\|TestNonFollowing\|TestWriteEvidence' -v ./internal/compat/crawler/` | 8 `--- PASS` lines, final `ok` | stories.md S2-07 Failure scenarios (skipped tool folded into a zero-hit pass) |
| No reader policy implemented | `grep -rniE 'reader_policy\|throttl\|deny' internal/compat/crawler/*.go \| grep -v _test.go` | no output | SPEC.md §20 acceptance 12 second half is Stage 3; stories.md S2-07 Constraints |

## T4 — Seeded tree builder

| Check | Command | Expected | FAIL cites |
| --- | --- | --- | --- |
| T4 tests | `go test -race -count=1 -run 'TestSeed' -v ./internal/compat/crawler/` | 3 `--- PASS` lines, final `ok` | stories.md S2-07 Failure scenarios (no `.snapshot` links, vacuous zero) |

## T5 — Gated crawler integration test

| Check | Command | Expected | FAIL cites |
| --- | --- | --- | --- |
| Build tag present | `head -1 internal/compat/crawler/crawler_integration_test.go` | `//go:build integration` | standards.md "Sprint 2 / Stage 1 integration-test gating" |
| Gate skips when unset | `env -u SNAPBACK_FUSE_TESTS go test -race -count=1 -tags=integration -run TestCrawler -v ./internal/compat/crawler/` | 2 `--- SKIP` lines whose messages contain `SNAPBACK_FUSE_TESTS`; final `ok` | standards.md skip discipline |
| Integration run (macOS, local macFUSE) | `mkdir -p .tmp/ev && SNAPBACK_FUSE_TESTS=1 SNAPBACK_EVIDENCE_DIR=$PWD/.tmp/ev go test -race -count=1 -tags=integration -run TestCrawler -v ./internal/compat/crawler/` | `--- PASS: TestCrawlerPositiveControl`, `--- PASS: TestCrawlerToolMatrix`; subtests PASS for each present tool, `--- SKIP` naming the binary for each missing one; final `ok` | SPEC.md §20 acceptance 12 (first half); §21 row 1 |
| Evidence non-following zero | `jq -e '[.rows[] \| select(.follows==false and .status=="tested") \| .hits] \| length > 0 and all(. == 0)' .tmp/ev/crawler-darwin.json` | `true` | SPEC.md §20 acceptance 12 ("zero catalog reads") |
| Evidence following recorded | `jq -e '[.rows[] \| select(.follows==true and .status=="tested") \| .hits \| type] \| all(. == "number")' .tmp/ev/crawler-darwin.json` | `true` | stories.md S2-07 Success scenarios (recorded count per following tool) |
| Evidence VS Code honest | `jq -r '.rows[] \| select(.tool \| test("VS Code")) \| .status' .tmp/ev/crawler-darwin.json` | `not-tested-here` | stories.md S2-07 Failure scenarios (VS Code reported as tested) |
| Evidence positive control | `jq -e '.positive_control_hits > 0' .tmp/ev/crawler-darwin.json` | `true` | M-002; stories.md S2-07 Failure scenarios |
| No stale mount | `mount \| grep -c snapback-crawler` | `0` | stories.md S2-04 Failure scenarios (stale mount) |

## Final sign-off

The full standards matrix runs here and only here (M-005: last task of the story).

| Check | Command | Expected | FAIL cites |
| --- | --- | --- | --- |
| Unit test count | `go test -list '.*' ./internal/compat/crawler/ \| grep -c '^Test'` | `24` | this plan.md (T1–T4 checkboxes) |
| Integration test count | `go test -tags=integration -list '.*' ./internal/compat/crawler/ \| grep -c '^Test'` | `26` | this plan.md (26 checkboxes) |
| Package coverage | `go test -race -covermode=atomic -coverprofile=.tmp/crawler.out ./internal/compat/crawler/ && go tool cover -func=.tmp/crawler.out \| tail -1` | total `>= 80.0%` | common/testing.md; standards.md coverage |
| Format | `test -z "$(gofmt -l .)" && test -z "$(goimports -l .)" && echo clean` | `clean` | standards.md Cross-cutting gates |
| Build | `CGO_ENABLED=0 go build ./...` | exit 0 | standards.md Cross-cutting gates; SPEC.md §5 |
| Vet | `go vet ./... && go vet -tags=integration ./internal/compat/crawler/` | exit 0, no output | standards.md Cross-cutting gates |
| Lint | `golangci-lint run && golangci-lint run --build-tags=integration ./internal/compat/crawler/...` | exit 0 | standards.md Cross-cutting gates; M-004 |
| Test + race (whole repo) | `go test -race -count=1 ./...` | every package `ok` | standards.md Cross-cutting gates |
| Coverage (whole repo) | `go test -race -covermode=atomic -coverprofile=coverage.out ./... && go tool cover -func=coverage.out \| tail -1` | total `>= 80.0%` | standards.md "How coverage is measured" |
| Vulnerability audit | `govulncheck ./...` | `No vulnerabilities found.` | standards.md Cross-cutting gates |
| Actions lint | `actionlint` | exit 0, no output | standards.md Cross-cutting gates |
| Shell lint | `shellcheck -s sh install.sh` | exit 0 | standards.md Cross-cutting gates |
| Docs build | `mkdocs build --strict --site-dir site` | exit 0 | standards.md Cross-cutting gates |
| Release config | `goreleaser check` | exit 0 | standards.md Cross-cutting gates |
| Full integration suite | `SNAPBACK_FUSE_TESTS=1 go test -race -count=1 -tags=integration ./internal/...` | every package `ok`; skips only name missing prerequisites | sprint plan.md "macOS evidence (local)" |
| No suppressions | `grep -rnE 'nolint\|t\.Skip\(' internal/compat/crawler \| grep -vE 'SNAPBACK_FUSE_TESTS\|not on PATH\|FUSE'` | no output | standards.md (zero suppressions; skip discipline) |
| Darwin evidence committed (orchestrator, post-merge) | `SNAPBACK_FUSE_TESTS=1 SNAPBACK_EVIDENCE_DIR=docs/reports/stage1 go test -race -count=1 -tags=integration -run TestCrawler ./internal/compat/crawler/ && git ls-files docs/reports/stage1/crawler-darwin.json` | `ok`, then the path | sprint plan.md "macOS evidence (local)" |
| Linux evidence (orchestrator, after S2-05 run) | `test -s docs/reports/stage1/crawler-linux.json && jq -r .goos docs/reports/stage1/crawler-linux.json` | `linux` | stories.md S2-05 Connections; S2-07 Success scenarios |
