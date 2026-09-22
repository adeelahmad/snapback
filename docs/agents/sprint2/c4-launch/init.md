---
type: init
story: S2-13
---

## S2-13/T1 · attempt 1 · red-worker · 2026-09-22T03:37:06Z

### Context (human request: GitHub repo launch kit)
The human supplied a kit at `/private/tmp/claude-501/-Users-adeelahmad-work-snapback/3a548722-26e5-4d22-8251-dbadece5e939/scratchpad/kit/github/` (README.md, REPO-SETUP.md, docs/brand/*, .github/ISSUE_TEMPLATE/{bug,feature,config}.yml, CONTRIBUTING.md, SECURITY.md). Repo settings (About, website, topics, Discussions, Pages HTTPS) are already applied by the orchestrator. The kit README over-claims (commands `snapback config|snap|web|install service` don't exist; "overlay works"; FUSE overlay into live dirs — SPEC forbids an overlay over the live tree; owner snapback-dev/branch main/URL /install are wrong). Adopt its SHAPE and brand, keep content TRUE (SPEC.md wording rules; today = Stage 0 skeleton with `snapback version` + Stage 1 compatibility evidence in docs/reports/stage1/).

### Mandate (tests only)
Update/add tests (test/projectdocs/, test/community/ or a new test/brand/ package) that FAIL BY ASSERTION today:
1. README starts with a `<picture>` whose dark source is `docs/brand/readme-banner-dark.png` and img src `docs/brand/readme-banner-light.png` with non-empty alt; both files exist and are PNGs.
2. Badge row: every badge URL references `adeelahmad/snapback` (never snapback-dev), CI badge uses `ci.yml` and `branch=master`; no badge for a target that does not exist (no Homebrew/apt/Go Report Card).
3. H1 is `# snapback` (lowercase); a bold one-line pitch follows; a `console` code block shows the canonical `cp .snapshot/latest/…` example clearly labelled as the design goal (not a present-tense claim) — honesty.
4. Section order: Why, Install, How it works, What it doesn't do, Status, Documentation, License (Quick start only if every command in it exists).
5. Honesty: README mentions no `snapback` subcommand other than `version` as working; contains no "snapback-dev", "branch=main", "snapback.run/install " (without .sh), "overlay"; Status section says early / not yet usable for restores and links the Stage 1 evidence.
6. Keep existing pins: install line `curl -fsSL https://snapback.run/install.sh | sh`, docs link `https://snapback.run/`, License links LICENSE (MIT).
7. `.github/ISSUE_TEMPLATE/bug.yml`, `feature.yml` exist and parse as YAML with name/description/body; `config.yml` has `blank_issues_enabled: false` and a contact link to `https://github.com/adeelahmad/snapback/discussions`.
8. `docs/brand/` holds mark.svg, favicon.svg, favicon-16.png, favicon-32.png, app-icon-180.png, og-card-dark.png, github-social-dark.png (1280×640 PNG, check IHDR), readme-banner-light.png, readme-banner-dark.png; no brand SVG contains `<script`.
Update any existing README test that conflicts (note each in output.md).

### Scope
May: test files only (test/**). May Not: README.md, .github/, docs/brand/, anything else.
