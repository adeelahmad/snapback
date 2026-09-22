---
type: validate
story: S1-05
---
# S1-05 — Community and governance furniture: validation rubric

Each section is PASS only if every command exits 0 with the expected output.
Run from the repo root `/Users/adeelahmad/work/snapback`.

## Pre-flight
- `test -f go.mod && grep -q '^module github.com/adeelahmad/snapback$' go.mod` — exits 0. FAIL: S1-01 not merged; S1-05 is wave 2 (sprint1/plan.md, "S1-01 merged green").
- `git diff --name-only master... | grep -vE '^(CONTRIBUTING.md|SECURITY.md|CODE_OF_CONDUCT.md|\.github/ISSUE_TEMPLATE/|\.github/pull_request_template.md|test/community/)'` — prints nothing. FAIL: edited outside the owned set (stories.md S1-05 "Owned files").
- `go list -m all | wc -l` is unchanged from before the story. FAIL: a module dependency was added; tests must use the standard library only (tasks.md header).

## T1 — Shared test helpers for the community package
- `go test ./test/community/ -run 'TestRepoRootHasGoMod|TestOwnedFilesExistAndNonTrivial' -v` — output contains `--- PASS: TestRepoRootHasGoMod` and `--- PASS: TestOwnedFilesExistAndNonTrivial`, ends `ok`. FAIL: an owned file missing or stub-sized (stories.md S1-05 failure scenario "placeholder-only").

## T2 — CONTRIBUTING.md
- `go test ./test/community/ -run 'TestContributing' -v` — three `--- PASS: TestContributing...` lines, `ok`.
- `for t in feat fix refactor docs test chore perf ci; do grep -q "\`$t\`" CONTRIBUTING.md || echo MISSING $t; done` — prints nothing. FAIL: commit types contradict S1-04's config (stories.md S1-05 failure scenario; `~/.claude/rules/common/git-workflow.md`).

## T3 — SECURITY.md with a disclosure path
- `go test ./test/community/ -run 'TestSecurity' -v` — four `--- PASS: TestSecurity...` lines, `ok`.
- `grep -c 'https://github.com/adeelahmad/snapback/security/advisories/new' SECURITY.md` — prints `1` or more. FAIL: no actual reporting channel (stories.md S1-05 failure scenario; `~/.claude/rules/common/security.md`).

## T4 — CODE_OF_CONDUCT.md
- `go test ./test/community/ -run 'TestConduct' -v` — two `--- PASS: TestConduct...` lines, `ok`.
- `grep -c 'INSERT CONTACT METHOD' CODE_OF_CONDUCT.md` — prints `0`. FAIL: upstream placeholder left in (stories.md S1-05 "contact method uses GitHub mechanisms").

## T5 — Issue templates (bug, feature)
- `go test ./test/community/ -run 'TestIssue' -v` — four `--- PASS: TestIssue...` lines, `ok`.
- `head -1 .github/ISSUE_TEMPLATE/bug_report.md .github/ISSUE_TEMPLATE/feature_request.md` — each shows `---`. FAIL: no YAML front matter (stories.md S1-05 success scenario "valid YAML front matter").

## T6 — Pull request template
- `go test ./test/community/ -run 'TestPRTemplate' -v` — two `--- PASS: TestPRTemplate...` lines, `ok`.

## T7 — Cross-file honesty and hygiene gate
- `go test ./test/community/ -run 'TestNo' -v` — three `--- PASS: TestNo...` lines, `ok`.
- `grep -riE 'production-ready|cross-platform|static|finder-integrated' CONTRIBUTING.md SECURITY.md CODE_OF_CONDUCT.md .github/ISSUE_TEMPLATE .github/pull_request_template.md` — prints nothing, exits 1. FAIL: honesty gate (standards.md "Honesty gate"; README.md §1, §18, §22).
- `grep -rEo '[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}' CONTRIBUTING.md SECURITY.md CODE_OF_CONDUCT.md .github/` — prints nothing. FAIL: personal email or secret committed (`~/.claude/rules/common/security.md` "NEVER hardcode secrets"; stories.md S1-05 constraints).

## Final sign-off
- `go test -race -count=1 ./test/community/...` — ends `ok  github.com/adeelahmad/snapback/test/community`; 20 tests pass (plan.md checkbox count).
- `test -z "$(gofmt -l test/community)"` and `go vet ./test/community/...` — both exit 0 (standards.md gate matrix).
- Every checkbox in `docs/agents/sprint1/s1-05-community/plan.md` is ticked in the story's plan-ready.md. FAIL: untested requirement unreported (standards.md honesty gate).
