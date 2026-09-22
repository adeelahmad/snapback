---
type: tasks
story: S1-04
---
# S1-04 — Conventional-commit enforcement: tasks

Owned set (exclusive): `.github/workflows/commitlint.yml`, `.commitlintrc.json`, `test/commitlint/**`.
Depends on S1-01 (`go.mod`, module `github.com/adeelahmad/snapback`). Tasks run in order T1 -> T2 -> T3 (T2/T3 share the workflow file; T2 reuses T1's helper).

**Allowed commit types (the contract S1-05's CONTRIBUTING.md mirrors, source `~/.claude/rules/common/git-workflow.md`):**
`feat`, `fix`, `refactor`, `docs`, `test`, `chore`, `perf`, `ci` — exactly these eight, no others (config-conventional's `build`, `style`, `revert` are excluded by the override).

**Parsing decision:** S1-01 is stdlib-only and `go.mod` is not in this story's owned set, so tests must not add `gopkg.in/yaml.v3`. JSON is parsed with `encoding/json`; the workflow YAML is checked with a small stdlib indentation scanner (`blockChildKeys`) plus literal/regexp assertions. The workflow must therefore use block-style `on:` / `jobs:` / `permissions:` mappings (no inline `on: [pull_request]`).

## T1 — Commitlint config and shared test helper

Files: `.commitlintrc.json`, `test/commitlint/helpers_test.go`, `test/commitlint/config_test.go`.

- `.commitlintrc.json`: `{"extends": ["@commitlint/config-conventional"], "rules": {"type-enum": [2, "always", [<the eight types>]]}}`.
- `helpers_test.go` (package `commitlint`): `readRepoFile(t, rel string) string` (resolves repo root as `../..`, fails the test on read error) and `blockChildKeys(text, parent string) []string` (returns the direct child keys of a top-level YAML block by indentation).
- `config_test.go`: the five config tests in plan.md, including the `commitlint`-on-PATH CLI check that skips when the CLI is absent.

## T2 — Workflow trigger, job and lint range

Files: `.github/workflows/commitlint.yml`, `test/commitlint/workflow_test.go`.

- Workflow `name: commitlint`; `on:` block with only `pull_request` (types opened, synchronize, reopened, edited); job `commitlint` on `ubuntu-latest`.
- `actions/checkout` with `fetch-depth: 0`; lint step runs `npx --no-install commitlint --from "${{ github.event.pull_request.base.sha }}" --to "${{ github.event.pull_request.head.sha }}" --verbose`. The base..head range is what keeps the two legacy commits ("Initial commit", "Update README.md") out of scope.
- PR-title step passes the title through `env: PR_TITLE: ${{ github.event.pull_request.title }}` and runs `printf '%s\n' "$PR_TITLE" | npx --no-install commitlint --verbose` (never interpolates the title inside `run:`).
- `workflow_test.go`: the five workflow-behaviour tests in plan.md.

## T3 — Workflow pinning and least privilege

Files: `.github/workflows/commitlint.yml`, `test/commitlint/workflow_pins_test.go`.

- `actions/setup-node` with an exact `node-version: "X.Y.Z"`; install step `npm install --no-save @commitlint/cli@X.Y.Z @commitlint/config-conventional@X.Y.Z` (exact versions, same major for both).
- Every `uses:` pinned to a version tag or 40-char SHA (never `@main`, `@master`, `@latest`).
- Top-level `permissions:` block with `contents: read` and nothing broader.
- `workflow_pins_test.go`: the four pinning/permission tests in plan.md.
