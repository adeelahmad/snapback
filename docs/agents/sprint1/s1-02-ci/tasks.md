---
type: tasks
story: S1-02
---

# S1-02 tasks — CI workflow and lint configuration

Owned set (stories.md S1-02): `.github/workflows/ci.yml`, `.golangci.yml`, `test/ci/**`. No task touches anything else. In particular `go.mod`/`go.sum` belong to S1-01, so `test/ci` uses the standard library only (no YAML dependency): assertions are line/substring based over the raw file text. Test package: `test/ci`, test-only (`package ci_test`), locating the repo root by walking up from the test's working directory to the directory holding `go.mod`.

Tasks run in order T1 -> T2 -> T3 (T1 and T2 both edit `ci.yml`, so they are not parallel).

## T1 — Core CI workflow (triggers, toolchain, gates)

- **Files:** `.github/workflows/ci.yml` (create), `test/ci/ci_test.go` (create), `test/ci/helpers_test.go` (create: `repoRoot`, `readRepoFile`).
- **Does:** `ci.yml` triggers on `push` and `pull_request`; every job uses `actions/setup-go` pinned to a version tag or SHA with `go-version-file: go.mod` and no `go-version:` key. Jobs/steps: `gofmt -l` + `goimports -l` check; `go build ./...`; `go vet ./...`; `go test -race -covermode=atomic -coverprofile=coverage.out ./...`; the 80% threshold check (standards.md "How coverage is measured" awk, which `exit 1`s below 80); `actions/upload-artifact` of `coverage.out`; `govulncheck ./...`; `shellcheck` over `git ls-files '*.sh'` that succeeds on zero files. No `continue-on-error: true` anywhere. Every `uses:` is pinned (`@vN...` or 40-hex SHA, never `@main`/`@master`/unpinned). No mount/browse/unmount job. No reference to S1-03..S1-08 files.
- **Tests:** plan.md T1 block.

## T2 — Cross-compile matrix with unverified labelling

- **Files:** `.github/workflows/ci.yml` (edit: add `cross-compile` job), `test/ci/matrix_test.go` (create).
- **Does:** A matrix job covering exactly seven `goos/goarch` targets: `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64` (verified) and `linux/arm`, `linux/mips`, `linux/mipsle` (job name contains `unverified`). The job `name:` expression renders the `unverified` label for exactly those three (for example a matrix `label` field used in `name:`). Linux builds set `CGO_ENABLED: 0`; a step runs `file` on Linux binaries as static-linkage evidence.
- **Tests:** plan.md T2 block.

## T3 — golangci-lint config, lint job pin, and actionlint check

- **Files:** `.golangci.yml` (create), `.github/workflows/ci.yml` (edit only if the lint step is missing its pinned `version:`), `test/ci/lint_test.go` (create).
- **Does:** `.golangci.yml` enables at least `staticcheck`, `govet`, `errcheck`, `goimports`. The `ci.yml` lint step uses `golangci/golangci-lint-action` pinned, with an explicit `version:` input (not `latest`). A test runs `actionlint` on `ci.yml` when it is on PATH and calls `t.Skip` (recorded skip, not pass) when absent.
- **Tests:** plan.md T3 block.
