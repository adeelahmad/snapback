---
type: validate
story: S1-04
---
# S1-04 — Validation rubric

All commands run from `/Users/adeelahmad/work/snapback` in the worktree root. Each check is PASS only when the literal expectation holds; anything else is FAIL.

## Pre-flight

- PASS if `test -f go.mod && head -1 go.mod` prints `module github.com/adeelahmad/snapback` (S1-01 merged). FAIL: stop, S1-04 is blocked on S1-01 (sprint1/plan.md wave 2 starts when S1-01 is green).
- PASS if `git diff --name-only master... | grep -vE '^(\.github/workflows/commitlint\.yml|\.commitlintrc\.json|test/commitlint/)'` prints nothing. FAIL: an edit escaped the owned set (sprint1/plan.md overlap rule; `go.mod`/`go.sum` belong to S1-01, so no YAML dependency may be added).
- PASS if `go version` succeeds. FAIL: toolchain missing (standards.md, pinned Go toolchain).

## T1 — Commitlint config and shared test helper

- PASS if `jq -r '.extends[]' .commitlintrc.json` prints exactly `@commitlint/config-conventional`.
- PASS if `jq -c '.rules["type-enum"]' .commitlintrc.json` prints `[2,"always",[...]]` whose list, sorted, is `chore ci docs feat fix perf refactor test`. FAIL: type list diverges from `~/.claude/rules/common/git-workflow.md` (story failure scenario; breaks the S1-05 CONTRIBUTING contract).
- PASS if `go test -race -count=1 -v -run 'TestConfig|TestCommitlintCLI' ./test/commitlint/...` ends with `ok` and shows `--- PASS` for the four `TestConfig*` tests (`TestCommitlintCLIVerdicts` may show `--- SKIP` when `commitlint` is absent). FAIL: any `--- FAIL` (`~/.claude/rules/golang/testing.md`, tests run with `-race`).

## T2 — Workflow trigger, job and lint range

- PASS if `grep -nE '^\s+push:|pull_request_target' .github/workflows/commitlint.yml` prints nothing. FAIL: runs on push to `master` where the range is undefined (story failure scenario).
- PASS if `grep -c 'github.event.pull_request.base.sha' .github/workflows/commitlint.yml` prints at least `1` and `grep -n 'max-parents=0' .github/workflows/commitlint.yml` prints nothing. FAIL: full history is linted and fails forever on the two legacy commits (intake assumption).
- PASS if `grep -n 'run:.*pull_request.title' .github/workflows/commitlint.yml` prints nothing. FAIL: script injection via PR title (`~/.claude/rules/common/security.md`, validate all inputs).
- PASS if `go test -race -count=1 -v -run 'TestWorkflow(Triggers|HasCommitlint|LintRange|Checkout|LintsPRTitle)' ./test/commitlint/...` shows five `--- PASS` and `ok`.

## T3 — Workflow pinning and least privilege

- PASS if `grep -nE '@(main|master|latest)\b|node-version:.*(x|lts|latest)|@commitlint/[a-z-]+@[\^~]' .github/workflows/commitlint.yml` prints nothing. FAIL: unpinned tool versions (story constraint "Node tool versions are pinned"; README.md §18 pinned toolchain).
- PASS if `grep -nE 'write' .github/workflows/commitlint.yml` prints nothing. FAIL: permissions broader than `contents: read` (least privilege, `~/.claude/rules/common/security.md`).
- PASS if `go test -race -count=1 -v -run 'TestWorkflow(PinsNode|PinsCommitlint|PinsActions|LeastPrivilege)' ./test/commitlint/...` shows four `--- PASS` and `ok`.

## Final sign-off

- PASS if `go test -race -count=1 ./test/commitlint/...` prints `ok  	github.com/adeelahmad/snapback/test/commitlint`, and `go test -count=1 -v ./test/commitlint/... | grep -c -- '--- PASS'` prints `14` (or `13` plus one `--- SKIP` for `TestCommitlintCLIVerdicts`).
- PASS if `test -z "$(gofmt -l test/commitlint)"` and `go vet ./test/commitlint/...` both exit 0 (standards.md Cross-cutting gates: gofmt, vet).
- PASS if `grep -L 'SUB-AGENT-TODO' test/commitlint/*.go` lists every file (no stubs left) and no `//nolint` or `t.Skip` exists other than the CLI-absent skip. FAIL: suppression (standards.md, zero suppressions).
- Orchestrator-only exit evidence (not a unit test): the first PR after merge shows the `commitlint` check running green (`gh pr checks <n>` lists `commitlint  pass`).
