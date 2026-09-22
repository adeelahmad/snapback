---
type: plan
story: S1-05
scope: "tests only"
---
# S1-05 — Community and governance furniture: test plan (tests only)

Package `test/community` (package `community`), Go standard library only. All paths
are resolved from the repo root found by `repoRoot(t)` (T1). Each test fails by
assertion, not by panic, when its target file is absent.

## T1 — Shared test helpers
- [ ] `test/community/helpers_test.go::TestRepoRootHasGoMod` — input: the test's working dir; action: `repoRoot(t)`; assertion: returned dir contains a `go.mod` whose `module` line is `github.com/adeelahmad/snapback`.
- [ ] `test/community/helpers_test.go::TestOwnedFilesExistAndNonTrivial` — input: `ownedFiles` (CONTRIBUTING.md, SECURITY.md, CODE_OF_CONDUCT.md, .github/ISSUE_TEMPLATE/bug_report.md, .github/ISSUE_TEMPLATE/feature_request.md, .github/pull_request_template.md); action: stat + read each; assertion: each exists, is a regular file, and has at least 200 bytes of non-whitespace-trimmed content (subtest per path).

## T2 — CONTRIBUTING.md
- [ ] `test/community/contributing_test.go::TestContributingHasRequiredHeadings` — input: CONTRIBUTING.md; action: collect lines starting with `## `; assertion: contains `## Development workflow (TDD)`, `## Commit messages`, `## Running the gate matrix locally`.
- [ ] `test/community/contributing_test.go::TestContributingListsEightCommitTypes` — input: the `## Commit messages` section body; action: regex-match `` `(feat|fix|refactor|docs|test|chore|perf|ci)` `` per type; assertion: all eight of feat, fix, refactor, docs, test, chore, perf, ci are present (missing types named in the failure).
- [ ] `test/community/contributing_test.go::TestContributingNamesLocalGateCommands` — input: the `## Running the gate matrix locally` section body; action: substring search; assertion: contains `go test -race ./...`, `go vet ./...` and `golangci-lint run`.

## T3 — SECURITY.md
- [ ] `test/community/security_test.go::TestSecurityHasRequiredHeadings` — input: SECURITY.md; action: heading scan; assertion: contains `## Supported versions` and `## Reporting a vulnerability`.
- [ ] `test/community/security_test.go::TestSecurityHasPrivateReportingChannel` — input: the `## Reporting a vulnerability` section; action: substring search; assertion: contains `https://github.com/adeelahmad/snapback/security/advisories/new` and the phrase `private vulnerability reporting` (case-insensitive).
- [ ] `test/community/security_test.go::TestSecurityWarnsAgainstPublicIssues` — input: the reporting section; action: case-insensitive search; assertion: contains `do not open a public issue`.
- [ ] `test/community/security_test.go::TestSecuritySupportedVersionsTable` — input: the `## Supported versions` section; action: find Markdown table rows (lines starting `|`); assertion: at least a header row, a separator row and one data row exist, and the section mentions `pre-release` (case-insensitive).

## T4 — CODE_OF_CONDUCT.md
- [ ] `test/community/conduct_test.go::TestConductIsContributorCovenant21` — input: CODE_OF_CONDUCT.md; action: heading + substring scan; assertion: headings `## Our Pledge`, `## Enforcement`, `## Attribution` present and the text contains `Contributor Covenant` and `version 2.1`.
- [ ] `test/community/conduct_test.go::TestConductHasGitHubContactMethod` — input: the `## Enforcement` section; action: substring search; assertion: contains `github.com/adeelahmad/snapback` or `@adeelahmad`, and the `[INSERT CONTACT METHOD]` placeholder from the upstream template is absent.

## T5 — Issue templates
- [ ] `test/community/issue_templates_test.go::TestIssueTemplatesHaveValidFrontMatter` — input: bug_report.md and feature_request.md; action: require line 1 == `---`, find the closing `---`, parse each line between as `key: value` (stdlib split on first `:`, key matches `^[a-z_]+$`); assertion: every line parses, and keys `name`, `about`, `title`, `labels` are present with non-empty values (subtest per file).
- [ ] `test/community/issue_templates_test.go::TestIssueTemplateTitlesAreConventional` — input: parsed front matter; action: read `title` (quotes stripped); assertion: bug_report title is `fix: ` and feature_request title is `feat: `.
- [ ] `test/community/issue_templates_test.go::TestIssueTemplateBodiesAreNotPlaceholder` — input: body after front matter; action: heading/substring scan; assertion: bug body mentions `Steps to reproduce`, `Expected`, `Actual`, `snapback version`; feature body mentions `Problem`, `Proposal`, `Alternatives`.
- [ ] `test/community/issue_templates_test.go::TestIssueConfigRoutesSecurityPrivately` — input: `.github/ISSUE_TEMPLATE/config.yml`; action: line scan; assertion: contains `blank_issues_enabled: false` and a `url:` line equal to `https://github.com/adeelahmad/snapback/security/advisories/new`.

## T6 — Pull request template
- [ ] `test/community/pr_template_test.go::TestPRTemplateHasConventionalTitleReminder` — input: `.github/pull_request_template.md`; action: case-insensitive search; assertion: contains `conventional commit` and `type(scope): subject`.
- [ ] `test/community/pr_template_test.go::TestPRTemplateHasTestChecklist` — input: the PR template; action: collect lines matching `^- \[ \] `; assertion: at least three items, and items mention `RED`, `go test -race ./...` and `docs`.

## T7 — Cross-file honesty and hygiene gate
- [ ] `test/community/honesty_test.go::TestNoBannedHonestyClaims` — input: `ownedFiles` plus every file in `.github/ISSUE_TEMPLATE/`; action: lower-case each and search; assertion: none contains `production-ready`, `cross-platform`, `static`, or `finder-integrated` (failure names file + phrase).
- [ ] `test/community/honesty_test.go::TestNoEmailAddresses` — input: same file set; action: regex `[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}`; assertion: zero matches (GitHub handles like `@adeelahmad` do not match).
- [ ] `test/community/honesty_test.go::TestNoPlaceholderTokens` — input: same file set; action: case-insensitive search; assertion: none contains `TODO`, `TBD`, `<fill`, or `lorem`.
