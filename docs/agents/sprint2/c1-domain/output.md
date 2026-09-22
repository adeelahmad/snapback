---
type: output
story: C1
---

## C1/T1 · attempt 1 · red-worker · 2026-09-22T03:18:41Z
status: ok
### Summary
Repointed the placeholder-domain tests at the real endpoints. The README install line must be exactly `curl -fsSL https://snapback.run/install.sh | sh`, and the README must contain no example.com or example.invalid anywhere. The install.sh default base URL must be https://github.com/adeelahmad/snapback/releases/latest/download. The mkdocs site_url must be https://snapback.run/. The docs.yml build job must copy install.sh into site/ between `mkdocs build` and upload-pages-artifact. The old README assertion that the domain is called a "placeholder" was removed because it contradicts the mandate. Commit 48d9718 on branch worktree-agent-a52468206417e969d. No shims were needed; go vet and golangci-lint are clean. Out-of-scope conflict: test/projectdocs/readme_links_test.go still pins docsSiteURL = https://adeelahmad.github.io/snapback/ (README line 43). That file is not in the May-scope, so I left it untouched. The planner should decide whether to extend scope.
### Result
| Check | Status | Detail |
|---|---|---|
| `test/projectdocs/readme_test.go::TestReadmeOneLineInstallUsesSnapbackRun` | FAIL | 0 lines "curl -fsSL https://snapback.run/install.sh \| sh", want exactly 1 |
| `test/projectdocs/readme_test.go::TestReadmeHasNoPlaceholderDomains` | FAIL | README contains placeholder domain "example.com" |
| `test/installer/dryrun_test.go::TestDryRunDefaultBaseURLIsGitHubReleases` | FAIL | stdout lacks the GitHub releases URL and contains example.invalid |
| `test/docs/mkdocs_config_test.go::TestSiteURLAndName/site_url` | FAIL | site_url = adeelahmad.github.io/snapback/, want https://snapback.run/ |
| `test/docs/workflow_test.go::TestWorkflowPublishesInstallScript` | FAIL | build job: no step copies install.sh into site/ |
### Next
GREEN: edit the README.md Install section, the install.sh default SNAPBACK_BASE_URL and the mkdocs.yml site_url. Add a `cp install.sh site/` step to .github/workflows/docs.yml before upload-pages-artifact. No symbols to scaffold.

## C1/T1b · attempt 1 · red-worker · 2026-09-22T03:21:37Z
status: ok
### Summary
Repinned the README docs-site URL in test/projectdocs/readme_links_test.go from https://adeelahmad.github.io/snapback/ to https://snapback.run/, and added a whole-README check forbidding the old github.io docs URL. Both checks fail by assertion on the current README (line 43 still lists the github.io site). No shim needed; only that file changed. Commit 4c4d3f6 on branch worktree-agent-a86d546f68153c8c3.
### Result
| Check | Status | Detail |
|---|---|---|
| `test/projectdocs/readme_links_test.go::TestReadmeLinksDocs` | FAIL | README "## Documentation" section missing docs site "https://snapback.run/"; README still links the old docs site "adeelahmad.github.io/snapback" |
| go vet / golangci-lint ./test/projectdocs/... | PASS | 0 issues |
### Next
GREEN: change the README.md Documentation section to list https://snapback.run/ and drop the adeelahmad.github.io/snapback link. No symbols to scaffold.

## C1/T1 · attempt 1 · green-worker · 2026-09-22T03:24:15Z
status: ok
### Summary
Pointed install and docs links at snapback.run with the smallest possible edits. README Install now shows exactly `curl -fsSL https://snapback.run/install.sh | sh`, and the placeholder sentence is gone. The docs-site link is https://snapback.run/. install.sh defaults SNAPBACK_BASE_URL to https://github.com/adeelahmad/snapback/releases/latest/download and still accepts an env override. mkdocs.yml site_url is https://snapback.run/. The docs.yml build job gains `cp install.sh site/install.sh` between the mkdocs build and upload-pages-artifact, and the pinned actions are unchanged. No tests were touched. Commit a996332 is authored by Adeel Ahmad only, with no AI trailers. ORCHESTRATOR.md still mentions github.io and example.invalid, but it is out of scope and was left alone. Nothing was pushed or deployed.
### Result
| Check | Status | Detail |
|---|---|---|
| `test/projectdocs::TestReadmeOneLineInstallUsesSnapbackRun` | PASS | green |
| `test/projectdocs::TestReadmeHasNoPlaceholderDomains` | PASS | green |
| `test/projectdocs::TestReadmeLinksDocs` | PASS | green |
| `test/docs::TestSiteURLAndName` | PASS | green |
| `test/docs::TestWorkflowPublishesInstallScript` | PASS | green |
| `test/installer::TestDryRunDefaultBaseURLIsGitHubReleases` | PASS | green |
| `go test -race ./...` | PASS | all packages ok |
| `go vet ./...` | PASS | |
| `golangci-lint run ./...` | PASS | 0 issues |
| `gofmt -l .` | PASS | no output |
| `shellcheck -s sh install.sh` | PASS | |
| `actionlint` | PASS | |
| `mkdocs build --strict` | PASS | |
### Next
Structural review, then merge into chain2/c1-domain. A human still has to configure the Pages custom domain (CNAME) for snapback.run.
