---
type: validate
story: S1-07
---

# S1-07 validate — PASS/FAIL rubric

Run every command from the repo root `/Users/adeelahmad/work/snapback`. A row PASSES only when the literal expected output matches; anything else is FAIL and cites the rule it violates (`docs/agents/sprint1/standards.md`, stories.md S1-07, README.md §18).

## Pre-flight

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| Toolchain resolvable | `go version` | `go version go1.27.1 <os>/<arch>` | standards.md "Toolchain pinning" |
| Module present (S1-01 merged) | `go list -m` | `github.com/adeelahmad/snapback` | stories.md S1-07 connections (depends on S1-01) |
| Gate tools present | `command -v goimports golangci-lint govulncheck` | three paths, exit 0 | standards.md gate-tool blocker |
| mkdocs (optional) | `command -v mkdocs \|\| echo absent` | a path, or `absent` (T7 then SKIPs; CI still runs the strict build) | stories.md S1-07 success scenario |
| Scope | `git diff --name-only master -- . ':!docs/agents'` | only `mkdocs.yml`, `requirements-docs.txt`, `.github/workflows/docs.yml`, `docs-site/*`, `test/docs/*` | stories.md S1-07 owned files |
| No go.mod change | `git diff --quiet master -- go.mod go.sum; echo "exit=$?"` | `exit=0` | task mandate (stdlib only; go.mod owned by S1-01) |

## T1 — Shared test helpers for the docs package

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| Helper tests | `go test -race -run 'TestRepoRootHasGoMod\|TestYAMLScalar' ./test/docs/` | `ok  	github.com/adeelahmad/snapback/test/docs` | standards.md Testing |
| Stdlib only | `go list -f '{{join .TestImports "\n"}}' ./test/docs/ \| grep '\.' ` | no output | task mandate (Go stdlib only) |

## T2 — Pinned docs requirements

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| Tests | `go test -race -run 'TestRequirements' ./test/docs/` | `ok  	github.com/adeelahmad/snapback/test/docs` | stories.md S1-07 intent (pins mkdocs-material) |
| Literal pins | `grep -vE '^\s*(#\|$)' requirements-docs.txt` | `mkdocs==1.6.1` and `mkdocs-material==9.6.14` | README.md §18 pinned tooling |

## T3 — mkdocs.yml configuration

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| Tests | `go test -race -run 'TestDocsDirIsDocsSite\|TestDocsAgentsNeverPublished\|TestSiteURLAndName\|TestThemeIsMaterial\|TestNavEntriesExist' ./test/docs/` | `ok  	github.com/adeelahmad/snapback/test/docs` | stories.md S1-07 failure scenario (docs_dir defaults to docs/) |
| docs_dir literal | `grep -E '^docs_dir:' mkdocs.yml` | `docs_dir: docs-site` | stories.md S1-07 intent |
| site_url literal | `grep -E '^site_url:' mkdocs.yml` | `site_url: https://adeelahmad.github.io/snapback/` | stories.md S1-07 constraints |

## T4 — Landing page stub

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| Tests | `go test -race -run 'TestLanding' ./test/docs/` | `ok  	github.com/adeelahmad/snapback/test/docs` | README.md §18 (restore story, one-line install) |
| Headings | `grep -E '^#{1,2} ' docs-site/index.md` | `# Snapback`, `## Install`, `## Status`, `## Versions` in that order | stories.md S1-07 intent |

## T5 — Docs workflow: build on PR, deploy to Pages on master and tags

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| Tests | `go test -race -run 'TestWorkflow' ./test/docs/` | `ok  	github.com/adeelahmad/snapback/test/docs` | stories.md S1-07 failure scenarios (pages permission missing) |
| Permissions | `grep -E 'pages: write\|id-token: write' .github/workflows/docs.yml` | both lines printed | stories.md S1-07 intent |
| Actions pinned to a major | `grep -E 'uses: actions/(upload-pages-artifact\|deploy-pages\|setup-python\|checkout)@' .github/workflows/docs.yml \| wc -l` | `4` or more | README.md §18 CI |

## T6 — Honesty and backend scan of published content

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| Tests | `go test -race -run 'TestNoBannedHonestyWords\|TestOnlyResticBackendNamed\|TestDocsSiteHasNoAgentArtifacts' ./test/docs/` | `ok  	github.com/adeelahmad/snapback/test/docs` | standards.md honesty gate |
| Banned words | `grep -rniE 'production-ready\|cross-platform\|static\|finder-integrated' docs-site mkdocs.yml` | no output | standards.md honesty gate |
| Other backends | `grep -rniwE 'borg\|borgbackup\|kopia\|duplicity\|duplicati\|tarsnap' docs-site mkdocs.yml` | no output | task mandate (Restic only in user-facing text) |

## T7 — Optional strict build smoke

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| Tests | `go test -race -v -run 'TestMkdocsStrictBuild\|TestBuiltSiteExcludesAgents' ./test/docs/` | both `--- PASS`, or both `--- SKIP` with `mkdocs not installed`; final `ok` | stories.md S1-07 success scenario |
| Manual strict build (if mkdocs present) | `mkdocs build --strict --site-dir /tmp/s107-site && test -f /tmp/s107-site/index.html && ! find /tmp/s107-site -path '*agents*' \| grep -q . && echo OK` | `OK` | stories.md S1-07 failure scenario (broken nav) |

## Final sign-off

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| gofmt | `test -z "$(gofmt -l .)"` | exit 0, no output | golang/coding-style.md |
| goimports | `test -z "$(goimports -l .)"` | exit 0, no output | golang/coding-style.md |
| Vet | `go vet ./...` | exit 0, no output | README.md §18 |
| Lint | `golangci-lint run` | exit 0, no output | README.md §18 |
| Race | `go test -race ./...` | every package `ok` | README.md §18 |
| Vulns | `govulncheck ./...` | `No vulnerabilities found.` | README.md §18 |
| Plan ticked | `grep -c '^- \[ \]' docs/agents/sprint1/s1-07-docs-site/plan.md` | `0` after execution | standards.md TDD workflow |

### Exit evidence (orchestrator, after merge — not a unit test)

"Pages URL returns 200" cannot be checked by a unit test; the orchestrator verifies it on `master` after merge. It is required for §22 row 0 ("docs site deploys").

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| Pages build type | `gh api repos/adeelahmad/snapback/pages --jq .build_type` | `workflow` | stories.md S1-07 failure scenario (Pages source misconfigured) |
| Deploy job green | `gh run list --workflow docs.yml --branch master --limit 1 --json conclusion --jq '.[0].conclusion'` | `success` | stories.md S1-07 exit evidence |
| Site live | `curl -fsSI https://adeelahmad.github.io/snapback/ \| head -1` | `HTTP/2 200` | README.md §18; §22 row 0 |
