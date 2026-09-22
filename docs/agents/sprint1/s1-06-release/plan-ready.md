---
type: plan-ready
story: S1-06
from_red_at: 2026-09-22T00:55:03Z
---

# S1-06 plan-ready (RED verified by orchestrator at chain @ 45866ed; every box currently FAILs by assertion unless noted)


# S1-06 plan — tests only

Package `test/release` (`package release_test`), Go stdlib only (`os`, `os/exec`, `path/filepath`, `regexp`, `strings`, `encoding/json`, `text/template`, `runtime`, `testing`), plus the in-module import `github.com/adeelahmad/snapback/internal/version` (S1-01). No new go.mod deps. Files are read relative to `repoRoot(t)` (walk up to `go.mod`). YAML is text-checked; `.releaserc.json` is decoded with `encoding/json`. Run: `go test -race ./test/release/...`.

Naming convention: `Test<Area><Behavior>`, where Area is `Goreleaser` (`.goreleaser.yaml`, T1+T2), `Releaserc`/`Changelog` (T3) or `Workflow` (`release.yml`, T4). Only `TestGoreleaserCheck` and `TestWorkflowActionlint` may `t.Skip`, and only when their tool is not on PATH.

## T1 — Test helpers and GoReleaser build section

- [x] `test/release/goreleaser_build_test.go::TestGoreleaserConfigHeader` — input: repo root; action: read `.goreleaser.yaml`; assertion: file exists, contains a line `version: 2` and a line `project_name: snapback`.
- [x] `test/release/goreleaser_build_test.go::TestGoreleaserBuildsCmdSnapback` — input: `builds:` block text; action: substring search; assertion: contains `main: ./cmd/snapback` and `binary: snapback`.
- [x] `test/release/goreleaser_build_test.go::TestGoreleaserCGODisabled` — input: `builds:` block text; action: regex search; assertion: `CGO_ENABLED=0` present and `CGO_ENABLED=1` absent from the whole file.
- [x] `test/release/goreleaser_build_test.go::TestGoreleaserTargetsExactlySeven` — input: the `targets:` list items; action: map each `os_arch_variant` to `os/arch`; assertion: the set equals exactly {linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, linux/arm, linux/mips, linux/mipsle} with no duplicates, and `linux_arm_7`, `linux_mips_softfloat`, `linux_mipsle_softfloat` appear verbatim.
- [x] `test/release/goreleaser_build_test.go::TestGoreleaserLdflagsTargetVersionVars` — input: `.goreleaser.yaml` text; action: collect every `-X <path>=<value>` flag; assertion: exactly three, equal to `github.com/adeelahmad/snapback/internal/version.Version={{ .Version }}`, `...internal/version.Commit={{ .ShortCommit }}` and `...internal/version.Target={{ .Os }}/{{ .Arch }}`.
- [x] `test/release/goreleaser_build_test.go::TestGoreleaserVersionVarsCompile` — input: `version.Version`, `version.Commit`, `version.Target` imported from `internal/version`; action: take the address of each as `*string`; assertion: all three pointers are non-nil and `version.Version == "dev"` in an un-ldflagged test build (a rename breaks compilation).
- [x] `test/release/goreleaser_build_test.go::TestGoreleaserLdflagsProduceVersionedBinary` — input: the three `-X` flags from `.goreleaser.yaml` with `{{ .Version }}`->`9.9.9`, `{{ .ShortCommit }}`->`abc1234`, `{{ .Os }}/{{ .Arch }}`->`runtime.GOOS/runtime.GOARCH`; action: `go build -ldflags` of `./cmd/snapback` with `CGO_ENABLED=0` into `t.TempDir()`, then run `snapback version`; assertion: exit 0 and stdout equals `snapback 9.9.9 (commit abc1234, target <GOOS>/<GOARCH>)\n`.

## T2 — Archives, checksums, cosign keyless signing, unverified labelling

- [x] `test/release/goreleaser_artifacts_test.go::TestGoreleaserArchiveNameTemplate` — input: `archives:` block text; action: substring search; assertion: contains the exact S1-03 contract `name_template` and `tar.gz` as the archive format, and has no `wrap_in_directory: true` (binary sits at the archive root, as the installer extracts it).
- [x] `test/release/goreleaser_artifacts_test.go::TestGoreleaserArchiveNamesRenderForAllTargets` — input: the extracted archive `name_template`; action: parse with `text/template` and execute for ProjectName `snapback` and each of the 7 Os/Arch pairs, appending `.tar.gz`; assertion: the results equal exactly the 7 contract assets (`snapback_linux_amd64.tar.gz` ... `snapback_linux_mipsle_unverified.tar.gz`), `_unverified` only on arm/mips/mipsle.
- [x] `test/release/goreleaser_artifacts_test.go::TestGoreleaserChecksumSha256` — input: `checksum:` block text; action: substring search; assertion: `name_template` is `checksums.txt` (quoted or bare) and `algorithm: sha256`.
- [x] `test/release/goreleaser_artifacts_test.go::TestGoreleaserSignsChecksumKeyless` — input: `signs:` block text and whole file; action: substring search; assertion: `cmd: cosign`, `artifacts: checksum`, `sign-blob` and `--output-signature` present; `--key`, `COSIGN_PRIVATE_KEY` and `COSIGN_PASSWORD` absent from the file.
- [x] `test/release/goreleaser_artifacts_test.go::TestGoreleaserReleaseNotesLabelUnverified` — input: `release:` block text; action: case-insensitive search of its `footer`/`header`; assertion: contains `unverified` and each of `linux/arm`, `linux/mips`, `linux/mipsle`.
- [x] `test/release/goreleaser_artifacts_test.go::TestGoreleaserNoPackagePublishers` — input: `.goreleaser.yaml` text; action: regex `^(brews|homebrew_casks|nfpms|publishers|snapcrafts|dockers|aurs|scoops|blobs|uploads):` per line; assertion: zero matches (stage 7 channels are out of scope).
- [x] `test/release/goreleaser_artifacts_test.go::TestGoreleaserReleaseAppendsToExisting` — input: `release:` and `changelog:` block text; action: substring search; assertion: `mode: append` under `release:` and `disable: true` under `changelog:` (no duplicate release vs semantic-release).
- [x] `test/release/goreleaser_artifacts_test.go::TestGoreleaserHonestyWords` — input: `.goreleaser.yaml` text; action: case-insensitive search; assertion: none of `production-ready`, `cross-platform`, `static`, `finder-integrated` occurs.
- [x] `test/release/goreleaser_artifacts_test.go::TestGoreleaserCheck` — input: `exec.LookPath("goreleaser")`; action: if absent `t.Skip("goreleaser not on PATH")`, else run `goreleaser check` in the repo root; assertion: exit 0.

## T3 — semantic-release config and seeded CHANGELOG

- [x] `test/release/releaserc_test.go::TestReleasercBranchMaster` — input: `.releaserc.json` decoded with `encoding/json`; action: read `branches`; assertion: equals exactly `["master"]`.
- [x] `test/release/releaserc_test.go::TestReleasercPluginsInOrder` — input: decoded `plugins` (each a string or `[name, options]`); action: extract names in order; assertion: equals `@semantic-release/commit-analyzer`, `@semantic-release/release-notes-generator`, `@semantic-release/changelog`, `@semantic-release/git`, `@semantic-release/github`.
- [x] `test/release/releaserc_test.go::TestReleasercConventionalCommitsPreset` — input: options of commit-analyzer and release-notes-generator; action: read `preset`; assertion: both equal `conventionalcommits`.
- [x] `test/release/releaserc_test.go::TestReleasercChangelogAndGitAssets` — input: options of changelog and git plugins; action: read fields; assertion: `changelogFile == "CHANGELOG.md"`, git `assets` contains `CHANGELOG.md`, git `message` contains `[skip ci]`.
- [x] `test/release/releaserc_test.go::TestReleasercGithubNoIssueWrites` — input: options of the github plugin; action: read fields; assertion: `successComment`, `failComment` and `releasedLabels` are each JSON `false`.
- [x] `test/release/releaserc_test.go::TestReleasercTolerantOfLegacyCommits` — input: decoded config; action: search every plugin's options; assertion: no `releaseRules` and no `parserOpts` keys anywhere, so the two legacy non-conventional commits parse as no-release under the default preset.
- [x] `test/release/releaserc_test.go::TestChangelogSeededHeader` — input: `CHANGELOG.md`; action: read first non-empty line and full text; assertion: first line is `# Changelog`, and none of `production-ready`, `cross-platform`, `static`, `finder-integrated` occurs (case-insensitive).

## T4 — Release workflow

- [x] `test/release/workflow_test.go::TestWorkflowTriggers` — input: `.github/workflows/release.yml` text; action: inspect the top-level `on:` block; assertion: has `push` with `branches:` containing `master`, has `workflow_dispatch`, and the file has no `pull_request` or `pull_request_target`.
- [x] `test/release/workflow_test.go::TestWorkflowDryRunInput` — input: release.yml text; action: inspect `workflow_dispatch.inputs`; assertion: a `dry_run` input with `type: boolean` exists and `inputs.dry_run` is referenced by the semantic-release step.
- [x] `test/release/workflow_test.go::TestWorkflowActionsPinned` — input: every `uses:` value; action: regex; assertion: each has `@` followed by `v<digit>` or a 40-hex SHA; none is `@main`, `@master` or unpinned.
- [x] `test/release/workflow_test.go::TestWorkflowToolVersionsPinned` — input: release.yml text; action: regex; assertion: `semantic_version:` is an exact `\d+\.\d+\.\d+`, every `extra_plugins` entry ends `@\d+\.\d+\.\d+`, the goreleaser-action `version:` is `v\d+\.\d+\.\d+` (not `latest`/`~>`), and `cosign-release:` is `v\d+\.\d+\.\d+`.
- [x] `test/release/workflow_test.go::TestWorkflowLeastPrivilege` — input: release.yml text; action: locate `permissions:` blocks per scope; assertion: top-level grants only `contents: read`; `id-token: write` occurs exactly once and inside the `goreleaser` job; no `write-all`; the `semantic-release` job grants `contents: write` and no `id-token`.
- [x] `test/release/workflow_test.go::TestWorkflowGoreleaserGatedOnNewRelease` — input: the `goreleaser` job block; action: substring search; assertion: `needs: semantic-release`, an `if:` containing `new_release_published == 'true'`, checkout with `fetch-depth: 0` and a `ref:` using `new_release_git_tag`, and `args: release --clean`.
- [x] `test/release/workflow_test.go::TestWorkflowSetupGoFromGoMod` — input: release.yml text; action: find `actions/setup-go`; assertion: present with `go-version-file: go.mod` and no `go-version:` key in the file.
- [x] `test/release/workflow_test.go::TestWorkflowSecretsOnlyGithubToken` — input: release.yml text; action: regex `secrets\.([A-Za-z_]+)`; assertion: every match is `GITHUB_TOKEN`, and `COSIGN_` does not occur.
- [x] `test/release/workflow_test.go::TestWorkflowNoStage7Channels` — input: release.yml text; action: case-insensitive search; assertion: none of `homebrew`, `brew `, `tap`, `apt-get`, `aptly`, `opkg`, `openwrt`, `gh-pages`, `mkdocs` occurs.
- [x] `test/release/workflow_test.go::TestWorkflowNoContinueOnError` — input: release.yml text; action: regex `continue-on-error:\s*true`; assertion: zero matches.
- [x] `test/release/workflow_test.go::TestWorkflowActionlint` — input: `exec.LookPath("actionlint")`; action: if absent `t.Skip("actionlint not on PATH")`, else run `actionlint .github/workflows/release.yml`; assertion: exit 0 with no output.
