---
type: tasks
story: S1-06
---

# S1-06 tasks — Semantic-release and release-artifact pipeline

Owned set (stories.md S1-06): `.github/workflows/release.yml`, `.releaserc.json`, `.goreleaser.yaml`, `CHANGELOG.md`, `test/release/**`. No task touches anything else. `go.mod`/`go.sum` belong to S1-01, so `test/release` is stdlib only: YAML is checked as raw text (line/substring/regex), `.releaserc.json` is decoded with `encoding/json`, and the archive `name_template` is rendered with `text/template`. Test package: `test/release`, test-only (`package release_test`), locating the repo root by walking up to the directory that holds `go.mod`.

Out of scope (README §22, stage 7): Homebrew tap, apt repo, OpenWrt feed, docs deploy (S1-07). No secrets in the repo; signing is cosign keyless via GitHub OIDC.

## Cross-story contracts (stated, each side tests its own half)

- **S1-01 ldflags (read from `s1-01-foundation/tasks.md`, not assumed):** `-X github.com/adeelahmad/snapback/internal/version.Version={{ .Version }}`, `-X github.com/adeelahmad/snapback/internal/version.Commit={{ .ShortCommit }}`, `-X github.com/adeelahmad/snapback/internal/version.Target={{ .Os }}/{{ .Arch }}`. Target renders as `GOOS/GOARCH`, the same shape as S1-01's default `runtime.GOOS + "/" + runtime.GOARCH`. S1-01 output line: `snapback <Version> (commit <Commit>, target <Target>)`.
- **S1-03 asset naming (read from `s1-03-installer/tasks.md`):** `archives.name_template: "{{ .ProjectName }}_{{ .Os }}_{{ .Arch }}{{ if or (eq .Arch \"arm\") (eq .Arch \"mips\") (eq .Arch \"mipsle\") }}_unverified{{ end }}"`, `formats: [tar.gz]` (GoReleaser v2; `format: tar.gz` is also accepted by the test), `checksum.name_template: checksums.txt`, `algorithm: sha256`. The 7 assets are `snapback_linux_amd64.tar.gz`, `snapback_linux_arm64.tar.gz`, `snapback_darwin_amd64.tar.gz`, `snapback_darwin_arm64.tar.gz`, `snapback_linux_arm_unverified.tar.gz`, `snapback_linux_mips_unverified.tar.gz`, `snapback_linux_mipsle_unverified.tar.gz`. `goarm: 7` and `gomips: softfloat` are carried in `.Arm`/`.Mips`, so `.Arch` stays bare (no `v7`/`_softfloat` in asset names). Binary at the archive root; no version in the asset name.

## Sequencing

T1 first (it creates `helpers_test.go` and `.goreleaser.yaml`). Then T2, T3, T4 may run in parallel: T2 is the only other task that edits `.goreleaser.yaml` (sequenced after T1), T3 owns `.releaserc.json` + `CHANGELOG.md`, T4 owns `release.yml`. No two parallel tasks edit the same file.

## T1 — Test helpers and GoReleaser build section (targets, CGO, ldflags)

- **Files:** `test/release/helpers_test.go` (create: `repoRoot(t)`, `readRepoFile(t, rel)`, `yamlBlock(text, key)` returning the indented block under a key), `test/release/goreleaser_build_test.go` (create), `.goreleaser.yaml` (create: `version: 2`, `project_name: snapback`, `builds:` only).
- **Does:** One build `id: snapback`, `main: ./cmd/snapback`, `binary: snapback`, `env: [CGO_ENABLED=0]`, `flags: [-trimpath]`, and an explicit `targets:` list of exactly `linux_amd64_v1`, `linux_arm64_v8.0`, `darwin_amd64_v1`, `darwin_arm64_v8.0`, `linux_arm_7`, `linux_mips_softfloat`, `linux_mipsle_softfloat` (an explicit list, not goos x goarch with ignores, so the set is text-checkable). `ldflags` are `-s -w` plus exactly the three S1-01 `-X` flags above. The test imports `github.com/adeelahmad/snapback/internal/version` and takes the address of `Version`, `Commit`, `Target`, so a rename breaks compilation. A build-and-exec test builds `./cmd/snapback` with the ldflags taken from `.goreleaser.yaml` (templates substituted) and runs `snapback version`.
- **Tests:** plan.md T1 block (7 tests).

## T2 — Archives, checksums, cosign keyless signing, unverified labelling

- **Files:** `.goreleaser.yaml` (edit, after T1: add `archives:`, `checksum:`, `signs:`, `release:`, `changelog:`), `test/release/goreleaser_artifacts_test.go` (create).
- **Does:** Archive `name_template` and format exactly as the S1-03 contract, with no `wrap_in_directory: true`, so the `snapback` binary sits at the archive root (installer contract). `checksum: {name_template: checksums.txt, algorithm: sha256}`. `signs:` with `cmd: cosign`, `artifacts: checksum`, args `sign-blob --output-signature=${signature} --output-certificate=${certificate} ${artifact} --yes`, and no `--key` / `COSIGN_PRIVATE_KEY` / `COSIGN_PASSWORD` (keyless via OIDC). `release: {mode: append}` so GoReleaser attaches assets to the GitHub release semantic-release already created, and `changelog: {disable: true}` so notes are not duplicated. `release.footer` states that linux/arm, linux/mips and linux/mipsle are **unverified** until a real mount test exists. No `brews`, `homebrew_casks`, `nfpms`, `publishers`, `snapcrafts`, `dockers`, `aurs`, `scoops`, `blobs`, `uploads`. No honesty-gate word (`production-ready`, `cross-platform`, `static`, `Finder-integrated`) anywhere in the file. `goreleaser check` runs when installed, and is a recorded skip otherwise.
- **Tests:** plan.md T2 block (9 tests).

## T3 — semantic-release config and seeded CHANGELOG

- **Files:** `.releaserc.json` (create), `CHANGELOG.md` (create), `test/release/releaserc_test.go` (create).
- **Does:** `branches: ["master"]`. `plugins` in this order: `@semantic-release/commit-analyzer` and `@semantic-release/release-notes-generator` (both `preset: "conventionalcommits"`), `@semantic-release/changelog` (`changelogFile: "CHANGELOG.md"`), `@semantic-release/git` (`assets: ["CHANGELOG.md"]`, message containing `[skip ci]`), `@semantic-release/github` (`successComment: false`, `failComment: false`, `releasedLabels: false`, so the job needs no issues/PR write). No `releaseRules` or `parserOpts` override, so the legacy non-conventional commits ("Initial commit", "Update README.md") parse as no-release instead of failing. `CHANGELOG.md` is only a `# Changelog` header plus one line saying entries are generated by semantic-release; it carries no honesty-gate word.
- **Tests:** plan.md T3 block (7 tests).

## T4 — Release workflow (semantic-release then GoReleaser)

- **Files:** `.github/workflows/release.yml` (create), `test/release/workflow_test.go` (create).
- **Does:** `on: push: branches: [master]` and `workflow_dispatch` with a boolean `dry_run` input; never `pull_request`/`pull_request_target`. Top-level `permissions: contents: read`. Job `semantic-release`: `permissions: contents: write`, `actions/checkout` (pinned, `fetch-depth: 0`), `cycjimmy/semantic-release-action` (pinned) with exact `semantic_version`, `extra_plugins` each pinned to an exact semver (`@semantic-release/changelog`, `@semantic-release/git`, `conventional-changelog-conventionalcommits`), `dry_run: ${{ inputs.dry_run }}`, `GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}`; it exposes `new_release_published` and `new_release_git_tag` as job outputs. Job `goreleaser`: `needs: semantic-release`, `if: needs.semantic-release.outputs.new_release_published == 'true'` (a tag pushed with `GITHUB_TOKEN` does not trigger a second workflow, so the tag path runs in-workflow), `permissions: contents: write, id-token: write` (the only place `id-token: write` appears), checkout at the new tag with `fetch-depth: 0`, `actions/setup-go` with `go-version-file: go.mod`, `sigstore/cosign-installer` pinned with exact `cosign-release`, `goreleaser/goreleaser-action` pinned with exact `version: v2.x.y` and `args: release --clean`. Every `uses:` is pinned (`@vN` or a 40-hex SHA). No `continue-on-error: true`. The only secret is `secrets.GITHUB_TOKEN`. No Homebrew/apt/OpenWrt/Pages steps. `actionlint` runs when installed, and is a recorded skip otherwise.
- **Tests:** plan.md T4 block (11 tests).
