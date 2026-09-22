---
type: validate
story: S1-06
---

# S1-06 validate — PASS/FAIL rubric

All commands run from `/Users/adeelahmad/work/snapback`. PASS requires every row's expected output. Any other output is FAIL.

## Pre-flight

| Check | Command | Expected | FAIL cites |
| --- | --- | --- | --- |
| S1-01 landed | `test -f go.mod && head -1 go.mod` | `module github.com/adeelahmad/snapback` | stories.md S1-06 Connections (hard-depends on S1-01) |
| S1-01 version vars exist | `grep -cE '^var (Version\|Commit\|Target) ' internal/version/version.go` | `3` | s1-01-foundation/tasks.md T2 (ldflags contract) |
| Only owned files changed | `git diff --name-only master... \| grep -vE '^(\.github/workflows/release\.yml\|\.releaserc\.json\|\.goreleaser\.yaml\|CHANGELOG\.md\|test/release/)'` | no output | stories.md S1-06 Owned files |
| No go.mod/go.sum edit | `git diff --name-only master... -- go.mod go.sum` | no output | stories.md S1-01 owns go.mod |

## T1 — Test helpers and GoReleaser build section (targets, CGO, ldflags)

| Check | Command | Expected | FAIL cites |
| --- | --- | --- | --- |
| T1 tests | `go test -race -run 'TestGoreleaser(ConfigHeader\|BuildsCmdSnapback\|CGODisabled\|TargetsExactlySeven\|LdflagsTargetVersionVars\|VersionVarsCompile\|LdflagsProduceVersionedBinary)$' -v ./test/release/` | 7 `--- PASS` lines, final `ok  	github.com/adeelahmad/snapback/test/release` | README.md §17 (7 targets, CGO_ENABLED=0, version/commit/target); stories.md S1-06 Failure scenarios (ldflags drift) |
| ldflags paths | `grep -oE 'internal/version\.(Version\|Commit\|Target)=' .goreleaser.yaml \| sort` | exactly `internal/version.Commit=`, `internal/version.Target=`, `internal/version.Version=` | s1-01-foundation/tasks.md contract |
| CGO off | `grep -c 'CGO_ENABLED=0' .goreleaser.yaml` | a number `>= 1` | README.md §17 |

## T2 — Archives, checksums, cosign keyless signing, unverified labelling

| Check | Command | Expected | FAIL cites |
| --- | --- | --- | --- |
| T2 tests | `go test -race -run 'TestGoreleaser(ArchiveNameTemplate\|ArchiveNamesRenderForAllTargets\|ChecksumSha256\|SignsChecksumKeyless\|ReleaseNotesLabelUnverified\|NoPackagePublishers\|ReleaseAppendsToExisting\|HonestyWords\|Check)$' -v ./test/release/` | 8 `--- PASS` lines plus `--- PASS: TestGoreleaserCheck` if goreleaser installed, else `--- SKIP: TestGoreleaserCheck` with `goreleaser not on PATH`; final `ok` | s1-03-installer/tasks.md asset contract; README.md §17 (checksums, signatures, unverified) |
| Checksum file | `grep -E 'checksums\.txt' .goreleaser.yaml \| wc -l` | a number `>= 1` | s1-03-installer/tasks.md |
| Keyless | `grep -cE -- '--key\|COSIGN_PRIVATE_KEY\|COSIGN_PASSWORD' .goreleaser.yaml` | `0` | intake: no secrets; keyless OIDC |
| No stage-7 publishers | `grep -cE '^(brews\|homebrew_casks\|nfpms\|publishers\|snapcrafts\|dockers\|aurs\|scoops\|blobs\|uploads):' .goreleaser.yaml` | `0` | stories.md S1-06 Constraints; README.md §22 |
| goreleaser check (if installed) | `goreleaser check` | exit 0; if not installed record NOT RUN (standards.md: NOT INSTALLED) | stories.md S1-06 Success scenarios |

## T3 — semantic-release config and seeded CHANGELOG

| Check | Command | Expected | FAIL cites |
| --- | --- | --- | --- |
| T3 tests | `go test -race -run 'Test(Releaserc(BranchMaster\|PluginsInOrder\|ConventionalCommitsPreset\|ChangelogAndGitAssets\|GithubNoIssueWrites\|TolerantOfLegacyCommits)\|ChangelogSeededHeader)$' -v ./test/release/` | 7 `--- PASS` lines, final `ok  	github.com/adeelahmad/snapback/test/release` | README.md §18 (semantic release, CHANGELOG.md in repo) |
| Valid JSON | `python3 -m json.tool .releaserc.json >/dev/null && echo valid` | `valid` | stories.md S1-06 What's wanted |
| Changelog header | `head -1 CHANGELOG.md` | `# Changelog` | stories.md S1-06 (seeded header) |

## T4 — Release workflow (semantic-release then GoReleaser)

| Check | Command | Expected | FAIL cites |
| --- | --- | --- | --- |
| T4 tests | `go test -race -run 'TestWorkflow(Triggers\|DryRunInput\|ActionsPinned\|ToolVersionsPinned\|LeastPrivilege\|GoreleaserGatedOnNewRelease\|SetupGoFromGoMod\|SecretsOnlyGithubToken\|NoStage7Channels\|NoContinueOnError\|Actionlint)$' -v ./test/release/` | 10 `--- PASS` lines plus `--- PASS: TestWorkflowActionlint` if actionlint installed, else `--- SKIP: TestWorkflowActionlint` with `actionlint not on PATH`; final `ok` | stories.md S1-06 Constraints (pinned, least privilege) and Failure scenarios (runs on PRs) |
| Not on PRs | `grep -cE 'pull_request' .github/workflows/release.yml` | `0` | stories.md S1-06 Failure scenarios |
| id-token scoped | `grep -c 'id-token: write' .github/workflows/release.yml` | `1` | stories.md S1-06 Constraints |
| Only GITHUB_TOKEN | `grep -oE 'secrets\.[A-Za-z_]+' .github/workflows/release.yml \| sort -u` | `secrets.GITHUB_TOKEN` | intake: no secrets in repo |

## Final sign-off

| Check | Command | Expected | FAIL cites |
| --- | --- | --- | --- |
| Whole package | `go test -race -count=1 ./test/release/...` | `ok  	github.com/adeelahmad/snapback/test/release` | common/testing.md |
| Test count | `go test -list '.*' ./test/release/ \| grep -c '^Test'` | `34` | this plan.md (34 checkboxes) |
| Format | `test -z "$(gofmt -l test/release)" && echo clean` | `clean` | golang/coding-style.md |
| Vet | `go vet ./test/release/...` | no output, exit 0 | README.md §18 |
| No suppressions | `grep -rnE 't\.Skip\(' test/release \| grep -vE 'goreleaser not on PATH\|actionlint not on PATH'` | no output | standards.md (zero suppressions) |
| No new deps | `git diff --name-only master... -- go.mod go.sum` | no output | S1-01 owns go.mod |
| Exit evidence (orchestrator, post-merge) | `gh workflow run release.yml -f dry_run=true && gh run list --workflow release.yml --limit 1 --json conclusion -q '.[0].conclusion'` then inspect the log for `The next release version is` | `success` and a computed next version in the log | stories.md S1-06 Exit evidence |
