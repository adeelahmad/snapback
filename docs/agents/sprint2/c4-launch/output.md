---
type: output
story: S2-13
---

## S2-13/T1 · attempt 1 · red-worker · 2026-09-22T03:50:00Z
status: ok

### Summary
Pinned the launch README shape (picture banner, badge row, lowercase H1 + bold pitch, console example labelled as design goal, section order, only `snapback version` mentioned, Status says early / not yet usable for restores with Stage 1 link, License MIT), the brand assets under docs/brand/ (new package test/brand: existence, PNG signature, 1280x640 social card via IHDR, no `<script` in SVGs), and the YAML issue forms (bug.yml/feature.yml structure, discussions contact link, retired .md templates gone). Stdlib only, offline, no shim needed. Commit 9748856 on chain2/c4-launch base 16a1c87.

Conflicting existing tests updated:
- `test/projectdocs/readme_test.go::TestReadmeTitleIsSnapback` removed: it pinned first line `# Snapback`; superseded by `TestReadmeStartsWithPictureBanner` (first line `<picture>`) and `TestReadmeH1AndBoldPitch` (H1 `# snapback`).
- `test/projectdocs/readme_links_test.go::TestHttmCreditIsProminent` relaxed from "within the first four ## headings" to "before ## Documentation", since the new required order Why/Install/How it works/What it doesn't do/... would otherwise force the credit into slot 4. The httm credit section itself is still required.
- `test/community/issue_templates_test.go`: the front-matter tests for bug_report.md/feature_request.md (`TestIssueTemplatesHaveValidFrontMatter`, `TestIssueTemplateTitlesAreConventional`, `TestIssueTemplateBodiesAreNotPlaceholder`) replaced by YAML-form tests; the "snapback version" prompt carries over as `TestBugFormAsksForSnapbackVersion` (the kit's `snapback --version` is rejected, since that flag does not exist). `TestIssueConfigRoutesSecurityPrivately` kept, so config.yml must keep the security advisory link alongside the discussions link.
- `test/community/helpers_test.go` ownedFiles/wantOwnedFiles now list bug.yml/feature.yml.
- Kept unchanged, which the GREEN README must still satisfy: GIF TODO marker in the lead, the "restoring a file should be as easy as it was in 2008" phrase and `cp .snapshot/` in the lead, "pre-release" in Status, the httm credit section, the Documentation links incl. `https://snapback.run/` (with slash), exactly one install line, and the honesty scans (for example "static" is banned, so the kit's "Static binaries" wording cannot be used).

### Result
| Check | Status | Detail |
|---|---|---|
| `test/projectdocs::TestReadmeStartsWithPictureBanner` | FAIL | first non-empty line = "# Snapback", want "<picture>" |
| `test/projectdocs::TestReadmeBadgesPointAtRepo` | FAIL | no badge row between </picture> and the H1 |
| `test/projectdocs::TestReadmeH1AndBoldPitch` | FAIL | H1 = "# Snapback", want "# snapback"; no bold pitch |
| `test/projectdocs::TestReadmeConsoleExampleIsDesignGoal` | FAIL | lead has no ```console block |
| `test/projectdocs::TestReadmeSectionOrder` | FAIL | missing/out of order from "## Why" |
| `test/projectdocs::TestReadmeMentionsOnlyVersionSubcommand` | FAIL | mentions snapback doctor/config/web/install |
| `test/projectdocs::TestReadmeHasNoLaunchKitOverclaims` | PASS-ON-RED | negative guard: current README has no snapback-dev/branch=main/overlay/install-without-.sh; it pins the kit's over-claims out |
| `test/projectdocs::TestReadmeStatusSaysEarlyAndLinksStage1` | FAIL | Status missing "early", "not yet usable for restores", "docs/reports/stage1/" |
| `test/projectdocs::TestReadmeLicenseLinksMIT` | FAIL | "## License" section missing |
| `test/projectdocs::TestHttmCreditIsProminent` | PASS-ON-RED | relaxed existing test (see Summary) |
| `test/brand::TestBrandAssetsExist` | FAIL | all 9 docs/brand files missing |
| `test/brand::TestGithubSocialCardIs1280x640` | FAIL | github-social-dark.png missing |
| `test/brand::TestBrandSVGsHaveNoScript` | FAIL | docs/brand has no SVG files (non-empty precondition, M-002) |
| `test/brand::TestPNGSize` | PASS-ON-RED | helper self-test |
| `test/community::TestIssueFormsParseAsYAML` | FAIL | bug.yml / feature.yml missing |
| `test/community::TestBugFormAsksForSnapbackVersion` | FAIL | bug.yml missing |
| `test/community::TestRetiredMarkdownTemplatesAreGone` | FAIL | bug_report.md, feature_request.md still exist |
| `test/community::TestIssueConfigLinksDiscussions` | FAIL | contact urls lack the discussions URL |
| `test/community::TestIssueConfigRoutesSecurityPrivately` | PASS-ON-RED | existing invariant kept (rewritten on shared helpers) |
| `test/community::TestParseIssueForm` | PASS-ON-RED | helper self-test |
| `test/community::TestOwnedFilesExistAndNonTrivial` + honesty scans | FAIL | now read bug.yml/feature.yml, which are missing |
| `go vet ./...`, `golangci-lint` on the 3 packages | PASS | 0 issues |

### Next
GREEN (docs-only, no Go symbols to scaffold): add docs/brand/ assets from the kit (mark, favicon svg/png, app-icon-180, og-card-dark, github-social-dark 1280x640, readme-banner light/dark); add .github/ISSUE_TEMPLATE/bug.yml and feature.yml (ask for `snapback version`, not `--version`) and delete bug_report.md/feature_request.md; add the discussions contact link to config.yml and keep the security advisory link; rewrite README.md in the kit's shape with true content (Stage 0 skeleton, Stage 1 evidence link, no overlay, no non-existent subcommands, badges on adeelahmad/snapback with ci.yml branch=master).

## S2-13/T1 · attempt 1 · green-worker · 2026-09-22T03:46:17Z
status: ok

### Summary
Docs-only GREEN, commit 8f03a1a on chain2/c4-launch (base 0ca2ed0), authored by the repo git user with no AI trailers. README.md rewritten in the kit's shape (picture banner, badge row, `# snapback`, bold pitch, GIF TODO marker, console example labelled as the design goal, then Why / Install / How it works / What it doesn't do / Status / Prior art and credit / Documentation / License), keeping every claim true: only `snapback version` exists, How it works uses future tense and follows SPEC.md (one managed `.snapshot` symlink into a read-only FUSE catalog plus a private `restic mount`, with no overlay over the live tree), and Status says early, pre-release, not yet usable for restores, with a link to docs/reports/stage1/. Badges: release, ci (ci.yml, branch=master), license and docs, all on adeelahmad/snapback. No Go Report Card, Homebrew or apt badges. Dropped the kit's claims that do not hold: the `config/snap/web/install service` subcommands, "overlay", "Static binaries", Homebrew/apt/ipk, snapback-dev, `/install` without .sh, and the rsnapshot mention (the backend regex bans it). Deviation from the old README: it said "there is no tagged release yet", but tags v1.0.0 and v1.0.1 exist, so Install now says the installer fetches the latest tagged build, which today only runs `snapback version`. docs/brand/: all 12 kit files copied (the 9 the tests need, plus the lockup and inverse SVGs). None contains `<script`. .github/ISSUE_TEMPLATE/: bug.yml asks for `snapback version` output, not `--version`; feature.yml comes from the kit; config.yml gains the adeelahmad/snapback discussions link and keeps the security advisory link. bug_report.md and feature_request.md removed with `git rm`. No tests were edited.

### Result
| Check | Status | Detail |
|---|---|---|
| `test/projectdocs::TestReadmeStartsWithPictureBanner` | PASS | green |
| `test/projectdocs::TestReadmeBadgesPointAtRepo` | PASS | green |
| `test/projectdocs::TestReadmeH1AndBoldPitch` | PASS | green |
| `test/projectdocs::TestReadmeConsoleExampleIsDesignGoal` | PASS | green |
| `test/projectdocs::TestReadmeSectionOrder` | PASS | green |
| `test/projectdocs::TestReadmeMentionsOnlyVersionSubcommand` | PASS | green |
| `test/projectdocs::TestReadmeHasNoLaunchKitOverclaims` | PASS | green |
| `test/projectdocs::TestReadmeStatusSaysEarlyAndLinksStage1` | PASS | green |
| `test/projectdocs::TestReadmeLicenseLinksMIT` | PASS | green |
| `test/projectdocs` (credit, lead, pre-release, install line, docs links, honesty scans) | PASS | green |
| `test/brand::TestBrandAssetsExist` / `TestGithubSocialCardIs1280x640` / `TestBrandSVGsHaveNoScript` | PASS | green |
| `test/community::TestIssueFormsParseAsYAML` / `TestBugFormAsksForSnapbackVersion` / `TestRetiredMarkdownTemplatesAreGone` / `TestIssueConfigLinksDiscussions` / `TestIssueConfigRoutesSecurityPrivately` | PASS | green |
| `test/community` owned files + honesty/email/placeholder scans | PASS | green |
| gofmt / goimports | PASS | no output |
| `CGO_ENABLED=0 go build ./...` / `go vet ./...` | PASS | |
| `golangci-lint run` | PASS | 0 issues |
| `go test -race ./...` | PASS | all packages ok |
| coverage | PASS | 84.0% total (>=80%) |
| `govulncheck ./...` | PASS | 0 called vulnerabilities (1 in a required module, not reachable) |
| `actionlint` | PASS | |
| `shellcheck -s sh install.sh` | PASS | |
| `mkdocs build --strict` | PASS | built clean |
| `goreleaser check` | PASS | 1 config validated |

### Next
Structural review of 8f03a1a (docs-only; no Go symbols), then merge chain2/c4-launch. Once merged, a human should look at the rendered README on GitHub to check the banners in light and dark mode.
