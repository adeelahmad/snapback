---
type: plan
story: S1-02
scope: "tests only"
---

# S1-02 plan — tests only

Package `test/ci` (`package ci_test`), stdlib only. All tests read files relative to the repo root found by `repoRoot(t)` (walk up to `go.mod`). Run: `go test -race ./test/ci/...`.

## T1 — Core CI workflow

- [ ] `test/ci/ci_test.go::TestCIWorkflowExists` — input: repo root; action: read `.github/workflows/ci.yml`; assertion: file exists and is non-empty.
- [ ] `test/ci/ci_test.go::TestCITriggersPushAndPullRequest` — input: ci.yml text; action: locate the top-level `on:` block; assertion: it contains both `push` and `pull_request` keys.
- [ ] `test/ci/ci_test.go::TestCISetupGoUsesGoModVersionFile` — input: ci.yml text; action: find every `actions/setup-go` step; assertion: at least one exists, each is followed by `go-version-file: go.mod`, and the file has no `go-version:` key.
- [ ] `test/ci/ci_test.go::TestCIFormatCheck` — input: ci.yml text; action: substring search; assertion: contains `gofmt -l` and `goimports -l`.
- [ ] `test/ci/ci_test.go::TestCIBuildAndVet` — input: ci.yml text; action: substring search; assertion: contains `go build ./...` and `go vet ./...`.
- [ ] `test/ci/ci_test.go::TestCITestUsesRaceAndCoverage` — input: ci.yml text; action: find the `go test` run line; assertion: it contains `-race`, `-covermode=atomic` and `-coverprofile=coverage.out`.
- [ ] `test/ci/ci_test.go::TestCICoverageThresholdFailsJob` — input: ci.yml text; action: find the step containing `go tool cover -func=coverage.out`; assertion: the same step contains `80` and an `exit 1` on the below-threshold branch.
- [ ] `test/ci/ci_test.go::TestCIUploadsCoverageArtifact` — input: ci.yml text; action: find `actions/upload-artifact`; assertion: present and its `path:` references `coverage.out`.
- [ ] `test/ci/ci_test.go::TestCIRunsGovulncheck` — input: ci.yml text; action: substring search; assertion: contains `govulncheck ./...`.
- [ ] `test/ci/ci_test.go::TestCIShellcheckToleratesZeroFiles` — input: ci.yml text; action: find the shellcheck step; assertion: it uses `git ls-files '*.sh'` and guards the empty list (e.g. `xargs -r` or an explicit empty check) so zero files does not fail.
- [ ] `test/ci/ci_test.go::TestCINoContinueOnError` — input: ci.yml text; action: regex `continue-on-error:\s*true`; assertion: zero matches.
- [ ] `test/ci/ci_test.go::TestCIActionsArePinned` — input: ci.yml text; action: collect every `uses:` value; assertion: each has `@` followed by `v<digit>` or a 40-hex SHA, and none ends in `@main`, `@master` or lacks `@`.
- [ ] `test/ci/ci_test.go::TestCINoMountJobs` — input: ci.yml text; action: case-insensitive search; assertion: no job or step name contains `mount`, `browse` or `unmount`.
- [ ] `test/ci/ci_test.go::TestCINoDependencyOnOtherStoryFiles` — input: ci.yml text; action: substring search for `install.sh`, `.goreleaser`, `.releaserc`, `mkdocs`, `commitlint`; assertion: none present.

## T2 — Cross-compile matrix

- [ ] `test/ci/matrix_test.go::TestMatrixHasAllSevenTargets` — input: ci.yml text; action: extract `goos`/`goarch` pairs from the cross-compile matrix; assertion: the set equals exactly {linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, linux/arm, linux/mips, linux/mipsle}.
- [ ] `test/ci/matrix_test.go::TestMatrixUnverifiedLabelExact` — input: extracted matrix entries; action: read each entry's label/name field; assertion: exactly linux/arm, linux/mips and linux/mipsle carry `unverified`, and none of the four verified targets do.
- [ ] `test/ci/matrix_test.go::TestMatrixJobNameShowsLabel` — input: ci.yml text; action: find the cross-compile job `name:`; assertion: it interpolates the matrix label (`${{ matrix.` ... `}}`) so `unverified` appears in the rendered job name.
- [ ] `test/ci/matrix_test.go::TestMatrixLinuxCgoDisabled` — input: ci.yml text; action: search the cross-compile job; assertion: contains `CGO_ENABLED: 0` (or `CGO_ENABLED=0`).
- [ ] `test/ci/matrix_test.go::TestMatrixRunsFileOnLinuxBinaries` — input: ci.yml text; action: search the cross-compile job; assertion: a step runs `file ` on the built binary guarded by a linux condition.

## T3 — Lint config and actionlint

- [ ] `test/ci/lint_test.go::TestGolangciConfigEnablesRequiredLinters` — input: `.golangci.yml` text; action: substring search inside the `enable` list; assertion: contains `staticcheck`, `govet`, `errcheck` and `goimports`.
- [ ] `test/ci/lint_test.go::TestCIGolangciLintActionPinnedVersion` — input: ci.yml text; action: find `golangci/golangci-lint-action`; assertion: present, pinned, followed by a `version:` input whose value is not `latest`.
- [ ] `test/ci/lint_test.go::TestCIActionlint` — input: ci.yml path; action: `exec.LookPath("actionlint")`, `t.Skip("actionlint not on PATH")` if absent, otherwise run `actionlint .github/workflows/ci.yml`; assertion: exit status 0 with empty output.
