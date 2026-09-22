---
type: stories
sprint: 1
---

# Sprint 1 stories — Snapback Stage 0: Scaffolding

Source of intent: `docs/agents/sprint1/intake.md`. Rules: `docs/agents/sprint1/standards.md`.
Spec anchors: README.md §2, §17, §18, §22 row 0, §24.

## Sprint goal

Stand up Stage 0 of the Snapback spec with no feature code: a Go module whose only
binary is `snapback version` (version/commit/target injected via `-ldflags`, §17),
GitHub Actions CI (build, vet, lint, `-race` tests, coverage, govulncheck, 7-target
cross-compile with the unverified targets labelled), conventional-commit enforcement,
a semantic-release + release-artifact pipeline, an MkDocs Material docs site deployed
to GitHub Pages, repository furniture, and an `install.sh` skeleton.

## Sprint demo

1. `go build -ldflags "-X …version.Version=v0.0.0-demo -X …Commit=abc123 -X …Target=linux/amd64" ./cmd/snapback && ./snapback version` prints the three injected values.
2. The CI workflow run on `master` is green, and its cross-compile job lists 4 verified and 3 **unverified** targets.
3. The release workflow's semantic-release dry run computes a next version from a conventional commit.
4. The GitHub Pages URL for `adeelahmad/snapback` serves the stub landing page.
5. `SNAPBACK_DRY_RUN=1 sh install.sh` on macOS/Linux prints the detected OS/arch, the asset and checksum it would fetch, and "next: `snapback config`". It never installs FUSE; it only prints how.

## Definition of Done

- Every story below is merged, and its tests pass under the standards gate matrix (`docs/agents/sprint1/standards.md` § Cross-cutting gates). That includes gofmt, goimports, `CGO_ENABLED=0 go build ./...`, `go vet ./...`, `golangci-lint run`, `go test -race ./...`, total coverage >= 80%, and `govulncheck ./...`.
- The blocking tool installs noted in standards.md (`goimports`, `golangci-lint`, `govulncheck`) are done by the orchestrator before any GREEN/FINAL gate.
- The orchestrator verifies exit evidence after merge, not a unit test: (a) the GitHub Actions CI run on `master` is green; (b) the Pages deploy succeeds and the URL returns HTTP 200; (c) the release workflow's semantic-release dry run succeeds.
- No docs text uses "production-ready", "cross-platform", "static" or "Finder-integrated" without linked evidence (honesty gate, §18). Each docs-bearing story asserts this in a test.
- No two stories modify the same file (see the owned-file lists).
- No FUSE, Restic, config, daemon or web-UI code exists.

## Out of scope

- Any feature code from stages 1 to 7: FUSE catalog, Restic provider, resolver, links registry, discovery, daemon, web UI, service installers (§22).
- Real mount/browse/unmount CI jobs (§18). These wait until mount code exists in stage 1.
- Homebrew tap, apt repo, OpenWrt feed, RPM/AUR/QNAP/Synology (§17, stage 7). The release pipeline builds and checksums artifacts but publishes to no package channel.
- Actually cutting a tagged release. A dry run is the Stage 0 evidence.
- Claiming a production domain for `install.sh`. A placeholder or the Pages URL is used.
- The README restore GIF. There is no restore flow yet; a placeholder is used.
- Docs-site content beyond a stub landing page.

## User stories

### S1-01 — Go module and `snapback version` binary (foundation)

As a maintainer, I want a buildable Go module whose only command is `snapback version` so that every later story (CI, release, installer) has a real binary to build, test and cross-compile.

**Owned files:** `go.mod`, `go.sum` (only if a dependency exists; a stdlib-only module has none, and that is acceptable), `cmd/snapback/**`, `internal/version/**`, `.gitignore`.

**Intent**

- **What's wanted:** Module path `github.com/adeelahmad/snapback` with `go` and `toolchain` directives. Package `internal/version` exposes `Version`, `Commit` and `Target` string vars, settable via `-ldflags -X`, with defaults `dev`, `none` and `runtime.GOOS/GOARCH`. It also exposes a pure formatter that renders them. `cmd/snapback` dispatches `version` to the formatter on stdout and exits 0. Any other or missing argument prints usage to stderr and exits non-zero. `.gitignore` adds `coverage.out` and `dist/`.
- **Constraints:** Standard library only. `CGO_ENABLED=0` builds (§17). Package and command are named `snapback` (§2). Pin the toolchain to the current stable Go patch release, not 1.22.x: govulncheck reports stdlib advisories against old toolchains and would turn CI red. Functions stay under 50 lines. `main` stays thin so the logic is testable and coverage stays >= 80%.
- **Failure scenarios:** The ldflags `-X` path does not match the var's import path, so injected values are silently ignored and the binary reports `dev`. Unknown subcommands exit 0. The version output goes to stderr. The toolchain pin is left at 1.22.2 and govulncheck fails. Any feature code (config, restic, fuse) creeps in.
- **Success scenarios:** A table-driven unit test covers the formatter with default and injected values. A test builds the binary with `-ldflags -X` into `t.TempDir()`, runs `snapback version` and asserts all three values on stdout with exit 0. A test checks that an unknown or missing subcommand exits non-zero with usage on stderr. `CGO_ENABLED=0 go build ./...` and `go vet ./...` pass, and coverage >= 80%.
- **Connections:** Every other story depends on it (they all need `go.mod` for their Go test packages). S1-06 (release) hard-codes the exact ldflags `-X` paths defined here. S1-02 (CI) cross-compiles this binary. S1-03 (installer) names the binary `snapback`.

### S1-02 — CI workflow and lint configuration

As a maintainer, I want GitHub Actions CI on every push and PR so that "green CI on an empty binary" (§22 row 0) is real and enforced.

**Owned files:** `.github/workflows/ci.yml`, `.golangci.yml`, `test/ci/**`.

**Intent**

- **What's wanted:** `ci.yml` triggers on `push` and `pull_request` and uses `actions/setup-go` with `go-version-file: go.mod` (the pinned toolchain). Its jobs are: gofmt/goimports check; `go build`; `go vet`; `golangci-lint` via its official action with a pinned version; `go test -race -covermode=atomic -coverprofile=coverage.out ./...` with the 80% threshold check and the coverage profile uploaded as an artifact; `govulncheck`; `shellcheck` over `git ls-files '*.sh'`, which tolerates zero files; and a cross-compile matrix of `linux/amd64`, `linux/arm64`, `darwin/amd64` and `darwin/arm64` (verified) plus `linux/arm`, `linux/mips` and `linux/mipsle`, whose job names carry an explicit `unverified` label. Linux builds use `CGO_ENABLED=0`, and `file` is run on the Linux binaries as static-linkage evidence. `.golangci.yml` enables at least `staticcheck`, `govet`, `errcheck` and `goimports`.
- **Constraints:** No mount/browse/unmount jobs (there is no mount code yet, §18). Third-party actions are pinned to a version or SHA. The Go toolchain is not hard-coded separately from `go.mod`. The workflow must not depend on files owned by S1-03..S1-08 existing.
- **Failure scenarios:** A target is silently dropped from the matrix. The unverified targets are not labelled. `-race` is missing. The coverage threshold is computed but never fails the job. A lint step runs with `continue-on-error`. govulncheck is skipped. The workflow YAML is invalid, so GitHub never runs it.
- **Success scenarios:** A Go test-only package `test/ci` reads `.github/workflows/ci.yml` and `.golangci.yml` and asserts all of the following: triggers include push and pull_request; all 7 targets are present and exactly `linux/arm`, `linux/mips` and `linux/mipsle` are marked unverified; `-race`, the coverage threshold, `govulncheck`, the golangci-lint step and `go-version-file: go.mod` are present; and there is no `continue-on-error: true`. If `actionlint` is on PATH the test also runs it, and it records a skip (not a pass) otherwise. **Exit evidence the orchestrator verifies after merge:** the CI run on `master` is green.
- **Connections:** Depends on S1-01. Its lint config governs every later Go story in later sprints. It is the CI that S1-03's shellcheck and every story's Go tests run under.

### S1-03 — `install.sh` skeleton

As a new user, I want a one-line installer skeleton matching §17 so that the install contract exists before real releases do.

**Owned files:** `install.sh`, `test/installer/**`.

**Intent**

- **What's wanted:** A POSIX `sh` script. It detects OS (`Linux`/`Darwin` mapped to `linux`/`darwin`) and arch (`x86_64`/`amd64` to `amd64`, `aarch64`/`arm64` to `arm64`, `armv7l`/`armv6l` to `arm`, `mips`, `mipsel` to `mipsle`), with test overrides via `SNAPBACK_OS`/`SNAPBACK_ARCH`. It builds the asset name and the checksum-file name. It verifies SHA-256 with `sha256sum` or `shasum -a 256`. It installs to `SNAPBACK_INSTALL_DIR`, falling back to `$HOME/.local/bin`. It prints next steps, including `snapback config`, and prints how to install FUSE (macFUSE, fuse3) without doing it. `SNAPBACK_DRY_RUN=1` prints the plan without downloading.
- **Constraints:** `set -eu`, shellcheck-clean, and POSIX sh (no bashisms). It must never invoke `brew`, `apt`, `apt-get`, `dnf`, `opkg` or similar to install FUSE. The base URL is an obvious placeholder or override (`SNAPBACK_BASE_URL`), not a live URL that claims a release exists. For unverified arch targets it prints a warning that says "unverified".
- **Failure scenarios:** An unsupported OS or arch (for example Windows or `riscv64`) exits 0 or installs garbage; it must exit non-zero with a clear message. A checksum mismatch still installs. The script calls a package manager to install FUSE. It hard-codes a real domain as if releases exist.
- **Success scenarios:** A Go test package `test/installer` runs `sh install.sh` with `SNAPBACK_DRY_RUN=1` and table-driven `SNAPBACK_OS`/`SNAPBACK_ARCH` pairs. It asserts the asset name for all 7 targets, the unverified warning for arm/mips/mipsle, a non-zero exit for unsupported pairs, and `snapback config` in the output. It serves a fake asset plus checksum from `httptest` via `SNAPBACK_BASE_URL` and checks that a good checksum installs into a temp dir and a bad one fails without installing. It greps that the script contains no package-manager install invocation. If `shellcheck` is available it is run, and the test skips otherwise (CI runs shellcheck via S1-02).
- **Connections:** Depends on S1-01 (`go.mod` for the test package; binary name). Its asset/checksum naming must match S1-06's goreleaser `name_template`/`checksum` settings; this is recorded as a cross-story contract, and each side tests its own half. It foreshadows stage 7 distribution channels.

### S1-04 — Conventional-commit enforcement

As a maintainer, I want PR commits and titles linted against Conventional Commits so that semantic-release (S1-06) can compute versions.

**Owned files:** `.github/workflows/commitlint.yml`, `.commitlintrc.json`, `test/commitlint/**`.

**Intent**

- **What's wanted:** A workflow on `pull_request` that lints the PR's commit range (base..head) with commitlint (`@commitlint/config-conventional`, pinned) and a config file that extends it with types matching `~/.claude/rules/common/git-workflow.md` (feat, fix, refactor, docs, test, chore, perf, ci).
- **Constraints:** Only commits from now on are evaluated. The two historical non-conventional commits ("Initial commit", "Update README.md") must not be linted (intake assumption). The Node tool versions are pinned.
- **Failure scenarios:** The workflow lints full history and fails forever on the two legacy commits. The type list diverges from the git-workflow rule. The workflow runs on push to `master`, where the range is undefined.
- **Success scenarios:** A Go test package `test/commitlint` parses `.commitlintrc.json` (encoding/json) and asserts that it extends config-conventional and allows exactly the eight types. It reads the workflow and asserts the `pull_request` trigger, a lint range bounded by base/head SHAs (not `--from` the root), and pinned versions. **Exit evidence:** the first PR shows the commitlint check running green, verified by the orchestrator.
- **Connections:** Depends on S1-01. It feeds S1-06 (semantic-release relies on conventional history) and pairs with S1-05's CONTRIBUTING.md, which documents the convention (text only; no shared file).

### S1-05 — Community and governance furniture

As a contributor, I want CONTRIBUTING, SECURITY, CODE_OF_CONDUCT and issue/PR templates so that the repo looks trustworthy and has a disclosure path (§18).

**Owned files:** `CONTRIBUTING.md`, `SECURITY.md`, `CODE_OF_CONDUCT.md`, `.github/ISSUE_TEMPLATE/**`, `.github/pull_request_template.md`, `test/community/**`.

**Intent**

- **What's wanted:** CONTRIBUTING covers the TDD workflow, the conventional-commit types, and how to run the gate matrix locally. SECURITY gives a concrete disclosure path: GitHub private vulnerability reporting for `adeelahmad/snapback`, plus the supported-versions table (pre-release: none). CODE_OF_CONDUCT is Contributor Covenant 2.1 with a contact method. The issue templates (bug, feature) and a PR template include a conventional-title reminder and a test checklist.
- **Constraints:** Markdown only. The honesty gate applies (no banned claims without evidence). No hardcoded personal secrets. The contact method uses GitHub mechanisms rather than inventing an email address.
- **Failure scenarios:** SECURITY.md has no actual reporting channel. A template is placeholder-only. CONTRIBUTING's commit types contradict S1-04's config. A banned honesty word appears.
- **Success scenarios:** A Go test package `test/community` asserts that each owned file exists and is non-trivial (it contains its required headings), that SECURITY.md contains a reporting-channel line, that CONTRIBUTING lists the eight commit types, that the issue templates have valid YAML front matter or form keys, and that no owned file contains "production-ready", "cross-platform", "static" or "Finder-integrated".
- **Connections:** Depends on S1-01 (`go.mod` for the test package). It textually references S1-04's commit types and the standards gate matrix. It is independent of all other stories' files.

### S1-06 — Semantic-release and release-artifact pipeline

As a maintainer, I want tags and changelog computed from conventional commits, and release artifacts built, checksummed and signed, so that the stage 7 pipeline exists from day one (§17, §18).

**Owned files:** `.github/workflows/release.yml`, `.releaserc.json`, `.goreleaser.yaml`, `CHANGELOG.md`, `test/release/**`.

**Intent**

- **What's wanted:** `.releaserc.json` configures semantic-release on branch `master` with commit-analyzer (conventionalcommits), release-notes-generator, changelog (`CHANGELOG.md`), git, and github plugins. `release.yml` runs semantic-release on push to `master` (with `workflow_dispatch` and a dry-run input), and on a new tag runs GoReleaser. `.goreleaser.yaml` builds `./cmd/snapback` for the 7 targets with `CGO_ENABLED=0`. Its ldflags `-X` paths match S1-01's `internal/version` vars exactly (Version/Commit/Target). It produces a SHA-256 `checksums.txt`, uses cosign keyless signing (`signs:` on the checksum), and marks the arm/mips/mipsle archives as unverified in the release notes and name. `CHANGELOG.md` is seeded with a header.
- **Constraints:** Node and GoReleaser versions are pinned. Workflow permissions are least-privilege (`contents: write` and `id-token: write` for cosign only where needed). No Homebrew/apt/OpenWrt publishers (stage 7). Commits that predate enforcement do not break version computation.
- **Failure scenarios:** The ldflags path drifts from `internal/version`, so released binaries report `dev`. A target is dropped. Checksums or signatures are missing. The unverified targets are presented as verified. semantic-release runs on PRs. It corrupts or overwrites `CHANGELOG.md`. It fails to compute a version because of the legacy commits.
- **Success scenarios:** A Go test package `test/release` reads `.goreleaser.yaml` and asserts that the `-X` paths equal `github.com/adeelahmad/snapback/internal/version.{Version,Commit,Target}`. The test also references those three vars directly, so a rename breaks compilation instead of silently drifting. It further asserts that all 7 goos/goarch pairs are present, that `CGO_ENABLED=0`, a sha256 checksum, a `signs` block, and "unverified" labelling are present, and that there are no brew/nfpm/publisher blocks. It parses `.releaserc.json` and asserts the branch, plugins and changelog file. It asserts that `release.yml` does not trigger on `pull_request`. `goreleaser check` runs if installed and skips otherwise. **Exit evidence:** the orchestrator verifies that a `workflow_dispatch` semantic-release dry run on GitHub computes a next version.
- **Connections:** Hard-depends on S1-01 (ldflags var paths, binary path) and softly on S1-04 (conventional history). Its asset naming is a contract with S1-03's installer. The docs deploy on tag stays in S1-07's workflow (no shared file). It is the pipeline stage 7 extends.

### S1-07 — Docs site (MkDocs Material) and GitHub Pages deploy

As a prospective user, I want a docs site deployed by CI to GitHub Pages so that "docs site deploys" (§22 row 0) is met.

**Owned files:** `mkdocs.yml`, `docs-site/**`, `requirements-docs.txt`, `.github/workflows/docs.yml`, `test/docs/**`.

**Intent**

- **What's wanted:** MkDocs Material (the routine choice allowed by §24) with `docs_dir: docs-site`, because `docs/` already holds agent planning artifacts. The stub `docs-site/index.md` leads with the restore-story framing, marks install instructions as coming soon, and states plainly that the project is pre-release. `requirements-docs.txt` pins `mkdocs-material`. `docs.yml` runs `mkdocs build --strict` on PRs, and on push to `master` (and tags) deploys with `actions/upload-pages-artifact` plus `actions/deploy-pages` using `pages: write` and `id-token: write`.
- **Constraints:** `site_url` points at `https://adeelahmad.github.io/snapback/`. The honesty gate applies. Versioning per release (§18) is deferred, noted in the page, and is not faked. It does not read `docs/agents/**`.
- **Failure scenarios:** `docs_dir` defaults to `docs/`, so agent artifacts are published. The Pages source is misconfigured, leaving the build green but nothing deployed. The pages permission is missing. A banned honesty word appears. The strict build fails on a broken nav.
- **Success scenarios:** A Go test package `test/docs` reads `mkdocs.yml` and asserts `docs_dir: docs-site`, the material theme, `site_url`, and that every `nav` entry exists under `docs-site/`. It asserts that `docs.yml` has the build-on-PR, deploy-on-master, deploy-pages and permission settings. It asserts that no file in `docs-site/**` contains a banned word. `mkdocs build --strict` runs into a temp dir if `mkdocs` is on PATH and skips otherwise. **Exit evidence:** the orchestrator verifies after merge that the Pages deploy job is green, `curl -fsSI https://adeelahmad.github.io/snapback/` returns 200, and Pages is enabled with build type `workflow` via `gh api`.
- **Connections:** Depends on S1-01 (`go.mod` for the test package). It will host `install.sh` as the interim installer URL in later stages (S1-03). Its docs content grows with stages 1 to 7.

### S1-08 — Project docs: README, ARCHITECTURE, NOTICE, CLAUDE.md, DEVLOG.md

As a visitor, I want a README with restore-story framing, an architecture overview, and dependency licence notices so that the repo meets §18 furniture and the user's project-docs standard.

**Owned files:** `README.md`, `SPEC.md` (new; the current README content moved verbatim), `ARCHITECTURE.md`, `NOTICE`, `CLAUDE.md`, `DEVLOG.md`, `test/projectdocs/**`.

**Intent**

- **What's wanted:** The current `README.md` is the full Snapback spec (Revision 2). It moves byte-for-byte to `SPEC.md`, so the spec is preserved and remains citable. The new `README.md` leads with the restore story, has a GIF placeholder marked TODO (no restore exists yet), states pre-release status, and links to SPEC.md, the docs site, CONTRIBUTING and SECURITY. `ARCHITECTURE.md` is a stub that summarises spec §4 modules and says plainly that only `cmd/snapback` and `internal/version` exist today. `NOTICE` lists dependency licences (at Stage 0: the Go standard library under BSD-3-Clause, with a note that the list updates as dependencies are added). `CLAUDE.md` (under 80 lines) and `DEVLOG.md` follow `~/.claude/rules/project-docs.md`.
- **Constraints:** The spec text must survive unchanged; `SPEC.md` equals the old README byte-for-byte. The honesty gate applies to README and ARCHITECTURE. No other story's files are edited, and references to them are links only.
- **Failure scenarios:** The spec is overwritten or truncated and lost. The README claims a working restore or uses a banned word. NOTICE is empty. CLAUDE.md exceeds 80 lines.
- **Success scenarios:** A Go test package `test/projectdocs` asserts that the SHA-256 of `SPEC.md` equals the hash of the Stage 0 README (`git show HEAD:README.md` at planning time, which the test hard-codes). It asserts that README contains the restore-story lead, the GIF TODO marker and a link to SPEC.md, and no banned words. It asserts that ARCHITECTURE.md, NOTICE and DEVLOG.md exist and are non-empty, and that CLAUDE.md is at most 80 lines.
- **Connections:** Depends on S1-01 (`go.mod` for the test package). Later planning inputs cite the spec as "README.md §N". After this story those citations resolve to `SPEC.md §N`, which the orchestrator must note for later sprints. It links to S1-05's and S1-07's outputs.

## Story dependency graph

```
S1-01 (go.mod + snapback version)  -- foundation, wave 1
  |-- S1-02 CI + lint             (wave 2)
  |-- S1-03 install.sh            (wave 2)  ~ naming contract with S1-06
  |-- S1-04 commitlint            (wave 2)  ~ soft feed into S1-06
  |-- S1-05 community furniture   (wave 2)
  |-- S1-06 release pipeline      (wave 2)  hard-depends on S1-01 ldflags paths
  |-- S1-07 docs site + Pages     (wave 2)
  `-- S1-08 project docs          (wave 2)
```

Hard edges: S1-01 to each of S1-02..S1-08. Soft (textual contract, no shared file, no ordering required): S1-03 and S1-06 (asset/checksum naming), S1-04 to S1-06 (conventional history), S1-05 to S1-04 (commit type list).
