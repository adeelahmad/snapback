---
type: tasks
story: S1-05
---
# S1-05 — Community and governance furniture: tasks

Owned set (stories.md S1-05): `CONTRIBUTING.md`, `SECURITY.md`, `CODE_OF_CONDUCT.md`,
`.github/ISSUE_TEMPLATE/**`, `.github/pull_request_template.md`, `test/community/**`.
Depends on S1-01 (`go.mod`, module `github.com/adeelahmad/snapback`). Tests use the
Go standard library only (no new module dependencies). Every file is Markdown and
must pass the honesty gate (standards.md: no "production-ready", "cross-platform",
"static", "Finder-integrated"). No email addresses are invented; contact paths are
GitHub mechanisms for `adeelahmad/snapback`.

## T1 — Shared test helpers for the community package
Files: `test/community/helpers_test.go`.
Add `package community` helpers: `repoRoot(t)` (walks up from the test dir to the
directory holding `go.mod`), `readOwned(t, rel) string` (fails with the relative
path if missing or empty), and `ownedFiles` (the fixed list of the six owned
furniture paths below). Test-only file; no production code.

## T2 — CONTRIBUTING.md
Files: `CONTRIBUTING.md`, `test/community/contributing_test.go`.
Headings `## Development workflow (TDD)`, `## Commit messages`, `## Running the gate matrix locally`.
The commit section lists the eight Conventional Commit types, each as a
`` `type` `` bullet: feat, fix, refactor, docs, test, chore, perf, ci (the same
set S1-04's commitlint config enforces). The gate section names `go test -race ./...`,
`go vet ./...` and `golangci-lint run`, and points at standards.md's gate matrix. Do not write the word `staticcheck` or
`SUB-AGENT-TODO` here: T7's substring scan for "static" and "TODO" would flag them.

## T3 — SECURITY.md with a disclosure path
Files: `SECURITY.md`, `test/community/security_test.go`.
Headings `## Supported versions`, `## Reporting a vulnerability`. The reporting
section directs reporters to GitHub private vulnerability reporting at
`https://github.com/adeelahmad/snapback/security/advisories/new`, states "do not open a public issue",
and gives an acknowledgement expectation. The supported-versions table states
that no release is supported yet (pre-release). No email address.

## T4 — CODE_OF_CONDUCT.md
Files: `CODE_OF_CONDUCT.md`, `test/community/conduct_test.go`.
Contributor Covenant 2.1 text with headings `## Our Pledge`, `## Enforcement`,
`## Attribution`; attribution names "version 2.1". The enforcement contact is a
GitHub mechanism (private vulnerability report link or contacting the maintainer
`@adeelahmad` via GitHub), not an invented email address.

## T5 — Issue templates (bug, feature)
Files: `.github/ISSUE_TEMPLATE/bug_report.md`, `.github/ISSUE_TEMPLATE/feature_request.md`,
`.github/ISSUE_TEMPLATE/config.yml`, `test/community/issue_templates_test.go`.
Each `.md` template opens with a `---` YAML front-matter block containing
non-empty `name:`, `about:`, `title:` (bug: `"fix: "`, feature: `"feat: "`) and
`labels:` keys, then a body with real prompts (bug: steps to reproduce, expected,
actual, `snapback version` output, OS/arch; feature: problem, proposal, alternatives).
`config.yml` sets `blank_issues_enabled: false` and a `contact_links` entry
pointing security reports at the private vulnerability reporting URL.

## T6 — Pull request template
Files: `.github/pull_request_template.md`, `test/community/pr_template_test.go`.
Contains a reminder that the PR title must be a Conventional Commit
(`type(scope): subject`) and a checklist with at least these unchecked items:
tests written first (RED before GREEN), `go test -race ./...` passes, docs updated.

## T7 — Cross-file honesty and hygiene gate
Files: `test/community/honesty_test.go`.
Iterates `ownedFiles` plus every file under `.github/ISSUE_TEMPLATE/`: asserts no
banned honesty phrase appears (case-insensitive), no email address pattern
appears, and no leftover placeholder token (`TODO`, `TBD`, `<fill`, `lorem`).
