---
type: intake
sprint: 1
---

# Intake — Stage 0: Scaffolding

## Restated request

Deliver Stage 0 ("Scaffolding") of the Snapback spec (`README.md`, Revision 2, §22 row 0): stand up the repository furniture, CI, lint, race tests, semantic release, a docs-site deploy pipeline, and an installer-script skeleton — with **no feature code** (no FUSE, no Restic, no config, no daemon, no web UI) — such that CI is green on a minimal ("empty") binary and the docs site deploys to a reachable URL. This is the exit evidence stated in §22 row 0: *"Green CI on an empty binary; docs site deploys."*

## 1. What's wanted (intent)

- A buildable Go module (`go.mod`/`go.sum`) named `snapback` containing a minimal `cmd/snapback` binary — see the "empty binary" assumption below for its exact contents.
- GitHub Actions CI that runs on every push/PR: `go build`, `go vet`, `staticcheck`/`golangci-lint`, unit tests with `-race`, cross-compilation of every stage-0-relevant release target, a coverage report, `govulncheck`, and a pinned Go toolchain (§18).
- Conventional-commit tooling and a semantic-release pipeline that computes versions, writes `CHANGELOG.md`, and can cut a GitHub release from tags (§18, §2 "Quality" row).
- A docs-site pipeline (MkDocs Material or Hugo, §18) built and deployed by CI to GitHub Pages, producing a reachable URL even if content is currently a stub landing page.
- Repository furniture: README (restore-story framing, GIF placeholder — see Assumptions), `ARCHITECTURE.md`, `CONTRIBUTING.md`, `SECURITY.md` with a disclosure path, `CODE_OF_CONDUCT`, issue/PR templates, `LICENSE` (permissive OSI — already present, see Constraints), `NOTICE` for dependency licences (§18).
- An installer-script **skeleton** (`install.sh`) reflecting the one-line-installer contract of §17 (OS/arch detection, checksum verification, install to a bin dir, print next steps, never auto-install FUSE) — skeleton only; it need not point at a live domain or download real release artifacts yet.

## 2. Constraints

Fixed by the spec (must follow as written, cited):

- **Language/build:** Go core, `CGO_ENABLED=0` for Linux targets; no Rust (§2 "Language" row; §17; §5 "Static builds"). Swift is out of scope for stage 0 (only used for the macOS Finder companion, stage 5).
- **Naming:** command, module path and package name are `snapback` (§2 "Name" row).
- **Release target matrix** for cross-compilation (§17): `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64` are the verified set; `linux/arm`, `linux/mips`, `linux/mipsle` must be published/labelled **unverified** until a real mount test exists (that mount test itself is out of scope for stage 0 — see Connections).
- **CI contents** (§18): GitHub Actions; `go build`, `go vet`, `staticcheck`/`golangci-lint`, unit+integration tests with `-race`, cross-compilation of every release target, coverage report, `govulncheck`, pinned Go toolchain. The spec's "real mount/browse/unmount tests on Linux runners and macOS runners with macFUSE" also lives under §18 but has no code to exercise until stage 1+ (see Connections/Out of scope).
- **Conventional commits + semantic release** (§18, §2 "Quality" row): automatic version bumps, tags, changelog, GitHub release notes from commit history; `CHANGELOG.md` kept in the repo.
- **Docs site** (§18, §2 "Quality" row): GitHub Pages, generated from repo Markdown, built/deployed by CI, versioned per release.
- **Repository furniture list** is fixed (§18): README+GIF, ARCHITECTURE.md, CONTRIBUTING.md, SECURITY.md w/ disclosure path, CODE_OF_CONDUCT, issue/PR templates, LICENSE (permissive OSI), NOTICE, go.mod/go.sum.
  - **LICENSE check:** `LICENSE` already exists in the repo and is **MIT** — this satisfies the "choose a permissive OSI licence" requirement; no new decision needed.
- **Installer contract** (§17): `curl -fsSL https://snapback.<domain>/install.sh | sh` — detects OS/arch, downloads the matching release, verifies the checksum, installs to a user or system bin dir, prints next steps (`snapback config`); must **never** install FUSE automatically, only print how.
- **Release-artifact requirements that the pipeline must be built to support** (§17), even though no real release ships in stage 0: SHA-256 checksums and Sigstore/cosign or GPG signatures on every artifact; every artifact reports version/commit/build target via `snapback version`.
- **Honesty gate** (§18, §1): never use "production-ready", "cross-platform", "static" or "Finder-integrated" in README/release notes/docs without linked evidence. Applies to all stage-0 docs text.
- **GitHub org/repo is already fixed and reachable:** `git remote -v` shows `origin` = `git@github.com:adeelahmad/snapback.git`; `gh repo view` confirms the repo exists, is public, default branch `master`. No naming/claiming decision remains open here.

## 3. Failure scenarios (what must not happen)

- Any CI step (build, vet, lint, race-enabled test, or any target's cross-compile) is red, flaky, or silently skipped — the stage-0 exit evidence is explicitly "green CI."
- The cross-compile job silently drops a required target, or fails to label the unverified Linux targets (`arm`/`mips`/`mipsle`) as unverified, misrepresenting §17's verified/unverified distinction.
- Semantic-release is wired against non-conventional commit history (the two existing commits, "Initial commit" and "Update README.md", do not follow conventional-commit format) and either fails to compute a version or corrupts `CHANGELOG.md`.
- The docs-site pipeline builds locally but fails to deploy (wrong GitHub Pages source configuration, permissions, or build output path) — "docs site deploys" evidence is unmet even if CI is green.
- The installer skeleton implies a real release exists (e.g., hardcodes a live download URL that 404s) or attempts to auto-install FUSE, violating §17's "never installs FUSE; prints how."
- `NOTICE` or dependency-licence tracking is skipped or wrong, letting an incompatible dependency licence slip in unnoticed.
- Stage-0 scope creep: any FUSE/Restic/config/daemon/web-UI code, or any of the full distribution channels (Homebrew tap, apt repo, OpenWrt feed) gets built now instead of the stage-0 skeleton/pipeline only — this pulls in stage 1–7 work prematurely and contradicts the mandate's explicit out-of-scope list.
- README/docs text asserts "production-ready", "cross-platform", "static," or "Finder-integrated" without linked evidence, breaching the honesty gate before any feature exists to back the claim.

## 4. Success scenarios (acceptance / done)

- `go build ./...`, `go vet ./...`, and the chosen linter (`staticcheck`/`golangci-lint`) run clean in GitHub Actions on push and PR.
- Unit tests run with `-race` and pass (even if trivial at this stage); a coverage report is produced; `govulncheck` runs and reports no unaddressed advisories; the Go toolchain version is pinned in `go.mod`/CI.
- The empty binary cross-compiles for all four verified targets (`linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`) and the three unverified Linux targets, with the verified/unverified distinction visible in CI output or release notes.
- A conventional-commit push (or a tag) drives semantic-release to compute the next version and update `CHANGELOG.md`, demonstrated at least once (dry-run acceptable if no tag is cut yet).
- The docs-site pipeline builds (MkDocs Material or Hugo) and deploys via CI to GitHub Pages, producing a reachable URL.
- All required repository furniture files exist and are non-empty/non-placeholder-only where content is knowable today (LICENSE, CONTRIBUTING.md, SECURITY.md, CODE_OF_CONDUCT, issue/PR templates, ARCHITECTURE.md stub, NOTICE, go.mod/go.sum).
- `install.sh` exists, is shell-lint-clean, implements OS/arch detection and checksum-verification structure, prints `snapback config` as next steps, and does not attempt to install FUSE.

## 5. Connections (dependencies / relations to other work)

- Depends on nothing upstream — stage 0 is the first stage (§22); it is a hard prerequisite for every later stage's CI, release, and docs-site infrastructure.
- The "real mount/browse/unmount tests on Linux and macOS runners with macFUSE" line in §18 has no code to exercise until the FUSE catalog exists (stage 1) and the macOS proof (stage 5); stage 0's CI should not attempt to fabricate these jobs against nonexistent mount code.
- The installer skeleton foreshadows the full distribution channel set in §17 (Homebrew tap, apt repo, OpenWrt `.ipk`, RPM/AUR/QNAP/Synology later) — those channels are out of scope now, but the skeleton's structure should not preclude adding them in stage 7.
- The semantic-release/CI pipeline built here is the same pipeline stage 7 ("Release") will rely on for full artifact publishing (binaries, checksums, signatures, Homebrew/apt/OpenWrt updates) — stage 0 builds the pipeline and its skeleton, not the final artifact set.
- The docs-site content roadmap (§18: landing page with restore-story framing and a demo GIF, install/quickstart, the three onboarding commands, configuration reference, discovery modes and costs, troubleshooting, backup/gitignore exclusions, managed-link cleanup, operations guide) mostly cannot be written yet because the features it documents don't exist; stage 0 delivers the pipeline plus a stub landing page, with real content arriving alongside the features in later stages.
- Local environment gap recorded in `ORCHESTRATOR.md`: `golangci-lint`, `staticcheck`, and `govulncheck` are **not installed locally** (only `go 1.22.2`, `restic`, `rclone`, `macFUSE`, `md-db` are present). CI can and should run these tools regardless; local pre-commit verification of the full suite is not currently possible until these tools are installed, which is a build-phase (not intake-phase) action item.
- `gh auth status` confirms the `adeelahmad` account token carries broad admin scopes (including `repo`, `admin:repo_hook`), and `gh repo view` confirms `adeelahmad/snapback` already exists as a public repo with commits pushed matching local `HEAD` — so "green CI" and "docs site deploys" can be verified for real (actual GitHub Actions runs, actual Pages deployment), not just simulated locally.

## In scope

- `go.mod`/`go.sum`, a minimal `cmd/snapback` binary (see Assumptions for its exact contents).
- GitHub Actions workflows: build/vet/lint/race-test/cross-compile/coverage/govulncheck, pinned Go toolchain.
- Conventional-commit enforcement + semantic-release wiring; `CHANGELOG.md`.
- Docs-site generator choice, build, and GitHub Pages deploy workflow, with a stub landing page.
- Repository furniture: README (restore-story framing; GIF itself deferred, see Assumptions), `ARCHITECTURE.md`, `CONTRIBUTING.md`, `SECURITY.md`, `CODE_OF_CONDUCT`, issue/PR templates, `NOTICE`. (`LICENSE` already exists, no action needed.)
- `install.sh` skeleton matching §17's one-liner contract, without requiring a live domain or a real published release.

## Out of scope (stages 1+, per mandate)

- FUSE catalog/mount adapter, Restic `SnapshotProvider` implementation, resolver, links registry, discovery (seed/on-access/explicit), daemon, web UI, native service installers (stage 1–6 per §22).
- Real distribution channels beyond the skeleton: Homebrew tap, apt repo + signing key, OpenWrt feed, RPM/AUR/QNAP/Synology (§17; ships from stage 7's tagged release).
- macOS Finder companion, on-access kernel hooks, and their acceptance tests (§20 items 18–19).
- Real "mount/browse/unmount" CI jobs (§18) — deferred until FUSE/mount code exists.
- Any §20 acceptance criteria beyond what's incidentally exercised by trivial build/vet/lint/race-test scaffolding.
- Purchasing/configuring a production domain for `install.sh` hosting (see Assumptions).

## Open questions (blocking)

None identified. Every candidate the mandate flagged was resolvable from the spec plus direct repo/tooling inspection (see Assumptions below for how each was resolved). If the orchestrator or human disagrees with any resolution recorded there — especially the "empty binary" contents or the installer domain placeholder — flag it before planning proceeds; otherwise these stand as the working scope.

## Assumptions

- **"Empty binary" = a `cmd/snapback` binary implementing only `snapback version`,** printing version/commit/build-target injected via `-ldflags` (per §17: "every artifact carries version, commit and build target (`snapback version`)"). This is the smallest binary that already satisfies a fixed spec requirement and gives CI something real to build/vet/lint/race-test/cross-compile, without pulling in feature code (no Restic/FUSE/config detection — that's `doctor`'s job, stage 3+). Non-blocking because the spec's own §17 language gives a defensible default; flagged here so the orchestrator can override if a literal no-op `main(){}` was intended instead.
- **Installer domain (`https://snapback.<domain>/install.sh`) is not yet claimed.** For the stage-0 skeleton, `install.sh` will use an obvious placeholder (e.g. a `TODO`-commented domain, or point at the repo's own GitHub Pages URL as an interim host) rather than block on domain acquisition. Claiming a real short domain (§2 "Name" row) is a follow-up action item, not a stage-0 blocker.
- **Docs-site generator (MkDocs Material vs Hugo)** is left as a routine implementation choice for the planner/build step, per §24 ("make routine implementation choices autonomously"); the spec offers both as valid options and does not prefer one.
- **Go toolchain version to pin** is a routine choice (local `go1.22.2` is available but flagged "old" in `ORCHESTRATOR.md`); pin to a current stable release in `go.mod`'s `toolchain` directive during the build phase, not fixed by the spec.
- **Semantic-release tooling** (JS `semantic-release` via a Node step, or a Go-native alternative such as `svu`/`git-cliff`) is not fixed by the spec; left as a build-phase choice. Because the two existing commits ("Initial commit", "Update README.md") predate conventional-commit enforcement, the assumption is that commit-linting and semantic-release only evaluate commits going forward, not full history.
- **README restore-demo GIF** (§18 repository-furniture bullet) cannot be produced yet — there is no working restore flow until stage 2+. Assumption: the README ships with a placeholder/TODO for the GIF at stage 0 and the honesty gate is respected (no claim of a working restore before evidence exists); the real GIF lands once the vertical slice (stage 2) exists.
- **GitHub Pages deployment mechanism** (Actions-based `actions/deploy-pages` vs a `gh-pages` branch via `peaceiris/actions-gh-pages`) is a routine build choice; either can be enabled via the `gh` CLI's admin-scoped token already available in this environment (confirmed via `gh auth status` / `gh api repos/.../pages`), so no additional human action is assumed necessary to make "docs site deploys" verifiable end-to-end.
