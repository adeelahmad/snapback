---
type: plan-ready
story: S1-04
from_red_at: 2026-09-22T00:53:23Z
---

# S1-04 plan-ready (RED verified by orchestrator at chain @ 4fb073b; every box currently FAILs by assertion unless noted)

# S1-04 — Test plan (tests only)

Package `test/commitlint` (Go, stdlib only). Run: `go test -race ./test/commitlint/...`.
Allowed commit types asserted below (and mirrored by S1-05 CONTRIBUTING.md): `feat`, `fix`, `refactor`, `docs`, `test`, `chore`, `perf`, `ci`.

## T1 — Commitlint config and shared test helper

- [x] `test/commitlint/config_test.go::TestConfigIsValidJSON` — input: `.commitlintrc.json` read via `readRepoFile`; action: `json.Unmarshal` into `map[string]any`; assertion: no error and the file is non-empty.
- [x] `test/commitlint/config_test.go::TestConfigExtendsConventional` — input: parsed config; action: read `extends` as `[]string`; assertion: it contains exactly `"@commitlint/config-conventional"`.
- [x] `test/commitlint/config_test.go::TestConfigTypeEnumExactlyEightTypes` — input: parsed `rules["type-enum"]`; action: take element [2] as `[]string`, sort, compare to sorted `{feat, fix, refactor, docs, test, chore, perf, ci}`; assertion: equal (no missing, no extra such as `build`, `style`, `revert`), no duplicates.
- [x] `test/commitlint/config_test.go::TestConfigTypeEnumIsErrorAlways` — input: parsed `rules["type-enum"]`; action: read elements [0] and [1]; assertion: level is `2` (error, not warning) and applicability is `"always"`.
- [x] `test/commitlint/config_test.go::TestCommitlintCLIVerdicts` — input: table of messages `feat: add x` (pass), `ci: pin node` (pass), `Initial commit` (fail), `Update README.md` (fail), `build: bump` (fail, type excluded); action: if `commitlint` is not on PATH `t.Skip`, else pipe each message to `commitlint --config ../../.commitlintrc.json`; assertion: exit status 0 for pass rows and non-zero for fail rows.

## T2 — Workflow trigger, job and lint range

- [x] `test/commitlint/workflow_test.go::TestWorkflowTriggersOnPullRequestOnly` — input: `.github/workflows/commitlint.yml`; action: `blockChildKeys(text, "on")`; assertion: equals `["pull_request"]` (no `push`, no `pull_request_target`, no `schedule`).
- [x] `test/commitlint/workflow_test.go::TestWorkflowHasCommitlintJob` — input: workflow text; action: `blockChildKeys(text, "jobs")` and a regexp for `runs-on:` under it; assertion: jobs contain `commitlint` and it declares `runs-on: ubuntu-latest`.
- [x] `test/commitlint/workflow_test.go::TestWorkflowLintRangeBoundedByPRShas` — input: workflow text; action: substring/regexp search; assertion: contains `--from "${{ github.event.pull_request.base.sha }}"` and `--to "${{ github.event.pull_request.head.sha }}"`, and does NOT contain `rev-list --max-parents=0`, `--from HEAD~`, or a `--from` without the base SHA (legacy commits never linted).
- [x] `test/commitlint/workflow_test.go::TestWorkflowCheckoutFetchesFullHistory` — input: workflow text; action: locate the `actions/checkout` step; assertion: its `with:` sets `fetch-depth: 0` (the base..head range needs both SHAs locally).
- [x] `test/commitlint/workflow_test.go::TestWorkflowLintsPRTitleViaEnv` — input: workflow text; action: search for the title step; assertion: `PR_TITLE: ${{ github.event.pull_request.title }}` appears under an `env:` key and no `run:` line contains `${{ github.event.pull_request.title }}` (script-injection guard).

## T3 — Workflow pinning and least privilege

- [x] `test/commitlint/workflow_pins_test.go::TestWorkflowPinsNodeVersion` — input: workflow text; action: regexp `node-version:\s*"?(\S+?)"?$`; assertion: exactly one match and it is an exact `\d+\.\d+\.\d+` (no `lts/*`, `20.x`, `latest`).
- [x] `test/commitlint/workflow_pins_test.go::TestWorkflowPinsCommitlintPackages` — input: workflow text; action: regexp for `@commitlint/cli@` and `@commitlint/config-conventional@` versions; assertion: both present, each an exact `\d+\.\d+\.\d+`, same major version, no `^`, `~`, `latest`.
- [x] `test/commitlint/workflow_pins_test.go::TestWorkflowPinsActions` — input: workflow text; action: collect every `uses:` value; assertion: at least one found and each has `@` followed by `v\d+(\.\d+)*` or a 40-hex SHA, never `@main`, `@master`, `@latest`.
- [x] `test/commitlint/workflow_pins_test.go::TestWorkflowLeastPrivilegePermissions` — input: workflow text; action: `blockChildKeys(text, "permissions")` and read the `contents:` value; assertion: keys equal `["contents"]` and value is `read` (no `write-all`, no write scopes).

Total: 14 tests.
