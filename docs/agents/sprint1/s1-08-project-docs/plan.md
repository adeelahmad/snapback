---
type: plan
story: S1-08
scope: "tests only"
---
# S1-08 — Project docs: test plan (tests only)

Package `test/projectdocs` (package `projectdocs`), Go standard library only
(`crypto/sha256`, `encoding/hex`, `os`, `path/filepath`, `regexp`, `strings`,
`bufio`). Paths resolve from `repoRoot(t)`. Every test fails by assertion (not
panic) when its file is absent. Honesty/backend scans target README.md and
ARCHITECTURE.md only; SPEC.md is verbatim and never scanned (it legitimately
contains "static", "cross-platform", "Finder-integrated", ZFS, btrfs, EC2, etc.).
Word checks use `regexp.MustCompile(`(?i)\b(...)\b`)` so substrings such as
`staticcheck` or `statically` do not match.

Frozen constant: `specSHA256 = "6bc35dadf698271b23790db43d2b521859d724081fae39dd1e281726e0fafe28"`,
`specSize = 64847`.

## T1 — Move the spec to SPEC.md byte-for-byte
- [ ] `test/projectdocs/helpers_test.go::TestRepoRootHasGoMod` — input: test working dir; action: `repoRoot(t)`; assertion: returned dir has `go.mod` whose `module` line is `github.com/adeelahmad/snapback`.
- [ ] `test/projectdocs/helpers_test.go::TestSectionExtractsBody` — input: in-memory doc `"# X\n## A\na1\n## B\nb1\n"`; action: `section(doc, "## A")`; assertion: returns `"a1\n"` (trimmed compare), and `section(doc, "## Z")` returns `""`.
- [ ] `test/projectdocs/spec_test.go::TestSpecHashMatchesStage0Readme` — input: raw bytes of `SPEC.md`; action: `sha256.Sum256` then `hex.EncodeToString`; assertion: equals `6bc35dadf698271b23790db43d2b521859d724081fae39dd1e281726e0fafe28` (failure message prints actual hash and says "spec altered; restore from git show e13192d:README.md").
- [ ] `test/projectdocs/spec_test.go::TestSpecSizeAndFirstLine` — input: `SPEC.md`; action: `len(bytes)` and first line; assertion: size is `64847` and first line is `# Snapback — Implementation Specification`.
- [ ] `test/projectdocs/spec_test.go::TestReadmeIsNotTheSpec` — input: `README.md` (after T2); action: sha256 of README bytes and substring search; assertion: hash differs from `specSHA256` and README does not contain `# Snapback — Implementation Specification`.

## T2 — New README: restore-story lead, GIF TODO, pre-release status, install
- [ ] `test/projectdocs/readme_test.go::TestReadmeLeadsWithRestoreStory` — input: README text before the first line starting `## `; action: case-insensitive substring search; assertion: contains `restoring a file should be as easy as it was in 2008` and `cp .snapshot/`.
- [ ] `test/projectdocs/readme_test.go::TestReadmeTitleIsSnapback` — input: README first non-empty line; action: trim; assertion: equals `# Snapback`.
- [ ] `test/projectdocs/readme_test.go::TestReadmeHasGifTodoMarkerAtTop` — input: README text before the first `## `; action: substring search; assertion: contains `<!-- TODO: restore GIF`.
- [ ] `test/projectdocs/readme_test.go::TestReadmeStatesPreRelease` — input: `section(readme, "## Status")`; action: case-insensitive search; assertion: section non-empty and contains `pre-release`.
- [ ] `test/projectdocs/readme_test.go::TestReadmeOneLineInstallUsesPlaceholderDomain` — input: `section(readme, "## Install")`; action: regex `curl -fsSL https://[a-z0-9.-]*example\.com/install\.sh \| sh`; assertion: exactly one match, and the section contains the word `placeholder`.
- [ ] `test/projectdocs/readme_test.go::TestReadmeClaimsNoWorkingRestore` — input: whole README; action: case-insensitive regex `(?i)\b(now available|ready to use|fully working|stable release)\b`; assertion: no match (pre-release honesty; SPEC §18).

## T3 — README: httm credit and links
- [ ] `test/projectdocs/readme_links_test.go::TestReadmeCreditsHttm` — input: `section(readme, "## Prior art and credit")`; action: substring search; assertion: contains `httm`, `https://github.com/kimono-koans/httm`, `kimono-koans` and `MPL-2.0`.
- [ ] `test/projectdocs/readme_links_test.go::TestHttmCreditIsProminent` — input: README lines; action: index of `## Prior art and credit` among `## ` headings; assertion: it is present and is within the first four `## ` headings.
- [ ] `test/projectdocs/readme_links_test.go::TestReadmeLinksDocs` — input: `section(readme, "## Documentation")`; action: regex for Markdown links `\]\((SPEC\.md|ARCHITECTURE\.md|CONTRIBUTING\.md|SECURITY\.md)\)` (subtest per target); assertion: all four link targets present, and `https://adeelahmad.github.io/snapback/` present.

## T4 — README honesty and Restic-only scan
- [ ] `test/projectdocs/honesty_test.go::TestNoHonestyWordsInPublicDocs` — input: README.md and ARCHITECTURE.md (subtest per file; SPEC.md excluded by construction); action: `(?i)\b(production-ready|cross-platform|static|finder-integrated)\b`; assertion: zero matches, failure lists line number and word.
- [ ] `test/projectdocs/honesty_test.go::TestHonestyRegexIsWholeWord` — input: strings `"staticcheck"`, `"statically"`, `"a static binary"`, `"Cross-Platform"`; action: apply the honesty regex; assertion: first two do not match, last two match (guards the substring trap).
- [ ] `test/projectdocs/honesty_test.go::TestNoFirstOfKindOrMultiBackendClaims` — input: README.md and ARCHITECTURE.md; action: `(?i)(first[- ]of[- ]its[- ]kind|multi-backend)`; assertion: zero matches (SPEC §19).
- [ ] `test/projectdocs/honesty_test.go::TestReadmeMentionsOnlyResticBackend` — input: README.md; action: `(?i)\b(zfs|btrfs|nilfs2?|borg|borgbackup|kopia|duplicity|duplicati|ec2|ebs|rsnapshot|tarsnap|bup|snapper)\b`; assertion: zero matches, and README contains `Restic` (SPEC §1 internal rule, §19).
- [ ] `test/projectdocs/honesty_test.go::TestArchitectureNamesNoOtherBackend` — input: ARCHITECTURE.md; action: same backend regex as above; assertion: zero matches.
- [ ] `test/projectdocs/honesty_test.go::TestReadmeHidesProviderSeam` — input: README.md; action: case-insensitive substring `snapshotprovider`; assertion: absent (seam is internal-only per SPEC §1).
- [ ] `test/projectdocs/honesty_test.go::TestSpecIsNotScanned` — input: the file list used by the honesty scans; action: inspect the `publicDocs` slice; assertion: equals exactly `[]string{"README.md", "ARCHITECTURE.md"}` and does not contain `SPEC.md`.

## T5 — ARCHITECTURE.md skeleton
- [ ] `test/projectdocs/architecture_test.go::TestArchitectureHeadings` — input: ARCHITECTURE.md; action: collect heading lines; assertion: contains `# Architecture`, `## Current state`, `## Planned modules`, `## Provider seam`.
- [ ] `test/projectdocs/architecture_test.go::TestArchitectureStatesCurrentState` — input: `section(arch, "## Current state")`; action: substring search; assertion: contains `` `cmd/snapback` ``, `` `internal/version` `` and `exist today`.
- [ ] `test/projectdocs/architecture_test.go::TestArchitectureListsSpecModules` — input: `section(arch, "## Planned modules")`; action: backticked-name search, subtest per module (config, provider, provider/restic, resolver, projection, mount, links, discovery/seed, discovery/onaccess, discovery/explicit, prewarm, daemon, web, service, macos/FinderCompanion); assertion: each appears as `` `name` ``, and the section links `SPEC.md`.
- [ ] `test/projectdocs/architecture_test.go::TestArchitectureDescribesProviderSeam` — input: `section(arch, "## Provider seam")`; action: substring search; assertion: contains `SnapshotProvider` and `Restic` and the phrase `only implementation`.

## T6 — NOTICE
- [ ] `test/projectdocs/notice_test.go::TestNoticeListsGoStdlib` — input: NOTICE; action: substring search; assertion: non-empty, contains `Go standard library` and `BSD-3-Clause`.
- [ ] `test/projectdocs/notice_test.go::TestNoticeSaysListIsMaintained` — input: NOTICE; action: case-insensitive search; assertion: contains `updated as dependencies are added`.
- [ ] `test/projectdocs/notice_test.go::TestNoticeCoversGoModRequires` — input: `go.mod` require lines (single-line and block form) and NOTICE; action: parse module paths, search NOTICE for each; assertion: every required module path appears (vacuously passes with zero requires today; fails when a dependency is added without a NOTICE entry).

## T7 — CLAUDE.md
- [ ] `test/projectdocs/claude_md_test.go::TestClaudeMdAtMost80Lines` — input: CLAUDE.md; action: count `\n`-separated lines (trailing newline not counted as an extra line); assertion: count is between 10 and 80 inclusive.
- [ ] `test/projectdocs/claude_md_test.go::TestClaudeMdFollowsTemplate` — input: CLAUDE.md; action: first line and `## ` headings; assertion: first line is `# Snapback`, headings include `## Stack`, `## Commands`, `## Structure`, `## Architecture`, `## Conventions`, `## Key Context`.
- [ ] `test/projectdocs/claude_md_test.go::TestClaudeMdLinksContextDocs` — input: `section(claude, "## Key Context")`; action: substring search; assertion: contains `SPEC.md`, `DEVLOG.md` and `ARCHITECTURE.md` or `standards.md`.
- [ ] `test/projectdocs/claude_md_test.go::TestClaudeMdCommandsAreGo` — input: `section(claude, "## Commands")`; action: substring search; assertion: contains `go test -race ./...` and `go vet ./...` and no `npm ` token.

## T8 — DEVLOG.md
- [ ] `test/projectdocs/devlog_test.go::TestDevlogTopLevelSections` — input: DEVLOG.md; action: `## ` headings; assertion: includes `## Working State`, `## Session Archive`, `## Milestones`, `## Mistakes & Lessons`, `## Technical Debt & Future Ideas` in that order.
- [ ] `test/projectdocs/devlog_test.go::TestDevlogWorkingStateShape` — input: `section(devlog, "## Working State")`; action: regex `\*\*Session:\*\* \d+ \| \*\*Date:\*\* \d{4}-\d{2}-\d{2}` and `### ` subheadings; assertion: session line matches, and `### Active Task`, `### Key Files (current shape)`, `### Decisions (active)`, `### Next Steps`, `### Blockers`, `### Watch Out` are present.
- [ ] `test/projectdocs/devlog_test.go::TestDevlogWorkingStateAtMost80Lines` — input: Working State section; action: line count; assertion: <= 80.
- [ ] `test/projectdocs/devlog_test.go::TestDevlogKeyFilesAtMostFive` — input: the `### Key Files (current shape)` subsection; action: count lines starting `` **` ``; assertion: between 1 and 5 inclusive.
