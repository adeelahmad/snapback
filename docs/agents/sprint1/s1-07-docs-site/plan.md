---
type: plan
story: S1-07
scope: "tests only"
---
# S1-07 — Docs site and GitHub Pages deploy: test plan (tests only)

Package `test/docs` (package `docs`), Go standard library only; YAML is checked as
text. Paths resolve from `repoRoot(t)` (T1). Each test fails by assertion, not panic,
when its target file is absent. The live "Pages URL returns 200" check is exit
evidence verified by the orchestrator after merge, not a unit test (see validate.md).

## T1 — Shared test helpers
- [ ] `test/docs/helpers_test.go::TestRepoRootHasGoMod` — input: the test working dir; action: `repoRoot(t)`; assertion: returned dir holds `go.mod` whose `module` line is `github.com/adeelahmad/snapback`.
- [ ] `test/docs/helpers_test.go::TestYAMLScalar` — input: table of snippets (`docs_dir: docs-site`, `site_url: "https://x/"`, indented `  name: material`, missing key); action: `yamlScalar(text, key)`; assertion: returns `docs-site`/true, `https://x/`/true (quotes stripped), false for an indented (non-top-level) key, false for a missing key.

## T2 — Pinned docs requirements
- [ ] `test/docs/requirements_test.go::TestRequirementsPinned` — input: `requirements-docs.txt`; action: drop blank and `#` lines, check each; assertion: every line matches `^[A-Za-z0-9._-]+==[0-9]+(\.[0-9]+)+$` (no `>=`, `~=`, or bare names).
- [ ] `test/docs/requirements_test.go::TestRequirementsListMkdocsMaterial` — input: the same file; action: collect package names; assertion: contains exactly `mkdocs` and `mkdocs-material`.

## T3 — mkdocs.yml configuration
- [ ] `test/docs/mkdocs_config_test.go::TestDocsDirIsDocsSite` — input: `mkdocs.yml`; action: `yamlScalar(text, "docs_dir")`; assertion: value is `docs-site` (so the default `docs/`, which holds `docs/agents`, is never published).
- [ ] `test/docs/mkdocs_config_test.go::TestDocsAgentsNeverPublished` — input: `mkdocs.yml`; action: scan every line; assertion: no line contains `docs/agents`, and no top-level `docs_dir` value equals `docs` or `docs/`.
- [ ] `test/docs/mkdocs_config_test.go::TestSiteURLAndName` — input: `mkdocs.yml`; action: `yamlScalar` for `site_url`, `site_name`, `repo_url`; assertion: `https://adeelahmad.github.io/snapback/`, `Snapback`, `https://github.com/adeelahmad/snapback`.
- [ ] `test/docs/mkdocs_config_test.go::TestThemeIsMaterial` — input: `mkdocs.yml`; action: find the `theme:` block and its indented children; assertion: a child line `name: material` exists within the block.
- [ ] `test/docs/mkdocs_config_test.go::TestNavEntriesExist` — input: `mkdocs.yml` `nav:` block; action: regex `:\s*([^\s]+\.md)\s*$` over indented list lines, stat `docs-site/<file>`; assertion: at least one entry, and every referenced file exists (missing ones named per subtest).

## T4 — Landing page stub
- [ ] `test/docs/landing_test.go::TestLandingLeadsWithRestoreStory` — input: `docs-site/index.md`; action: take the first `# ` heading and the first non-empty paragraph after it; assertion: heading is `# Snapback` and that paragraph contains `restore` (case-insensitive) before any `## ` heading.
- [ ] `test/docs/landing_test.go::TestLandingHasRequiredSections` — input: `docs-site/index.md`; action: collect `## ` headings; assertion: contains `## Install`, `## Status`, `## Versions`, with `## Install` appearing before `## Status`.
- [ ] `test/docs/landing_test.go::TestLandingInstallIsComingSoon` — input: the `## Install` section; action: substring search; assertion: contains `Coming soon` (no working install command is claimed at Stage 0).
- [ ] `test/docs/landing_test.go::TestLandingStatesPreRelease` — input: the `## Status` section; action: case-insensitive search; assertion: contains `pre-release`.
- [ ] `test/docs/landing_test.go::TestLandingVersioningDeferred` — input: the `## Versions` section; action: case-insensitive search; assertion: contains `not yet` and `per release` (versioning deferred, not faked).
- [ ] `test/docs/landing_test.go::TestLandingNamesRestic` — input: `docs-site/index.md`; action: case-insensitive search; assertion: contains `restic`.

## T5 — Docs workflow
- [ ] `test/docs/workflow_test.go::TestWorkflowTriggers` — input: `.github/workflows/docs.yml`; action: text scan of the `on:` block; assertion: contains `pull_request:`, `push:`, a `branches:` entry with `master`, and a `tags:` entry with `v*`.
- [ ] `test/docs/workflow_test.go::TestWorkflowBuildsStrict` — input: same file; action: substring search; assertion: contains `pip install -r requirements-docs.txt` and `mkdocs build --strict`.
- [ ] `test/docs/workflow_test.go::TestWorkflowUploadsPagesArtifact` — input: same file; action: substring search; assertion: contains `actions/upload-pages-artifact@` and `path: site`.
- [ ] `test/docs/workflow_test.go::TestWorkflowDeployJob` — input: the `deploy:` job block (indented lines until the next job key); action: substring search within the block; assertion: contains `needs: build`, `actions/deploy-pages@`, `environment:` with `github-pages`, `pages: write`, `id-token: write`.
- [ ] `test/docs/workflow_test.go::TestWorkflowDeploySkippedOnPR` — input: the `deploy:` job block; action: substring search; assertion: contains an `if:` line with `github.event_name != 'pull_request'`.
- [ ] `test/docs/workflow_test.go::TestWorkflowTopLevelPermissionsReadOnly` — input: same file; action: find the column-0 `permissions:` block; assertion: it contains `contents: read` and does not contain `write`.

## T6 — Honesty and backend scan
- [ ] `test/docs/honesty_test.go::TestNoBannedHonestyWords` — input: every file under `docs-site/` plus `mkdocs.yml`; action: lowercase substring search for `production-ready`, `cross-platform`, `static`, `finder-integrated`; assertion: zero hits (subtest per file, hit and line number named).
- [ ] `test/docs/honesty_test.go::TestOnlyResticBackendNamed` — input: the same file set; action: case-insensitive word-boundary regex for `borg|borgbackup|kopia|duplicity|duplicati|tarsnap`; assertion: zero hits.
- [ ] `test/docs/honesty_test.go::TestDocsSiteHasNoAgentArtifacts` — input: `filepath.WalkDir(docs-site)`; action: inspect each path and each file's first 10 lines; assertion: no path segment named `agents` and no file begins with frontmatter `type: (tasks|validate|plan|stories)`.

## T7 — Optional strict build smoke
- [ ] `test/docs/build_test.go::TestMkdocsStrictBuild` — input: repo `mkdocs.yml`; action: skip if `exec.LookPath("mkdocs")` fails, else run `mkdocs build --strict --config-file <root>/mkdocs.yml --site-dir <t.TempDir()>`; assertion: exit 0 (stderr included on failure) and `<tmp>/index.html` exists.
- [ ] `test/docs/build_test.go::TestBuiltSiteExcludesAgents` — input: same build output (skip when mkdocs absent); action: walk the output dir; assertion: no path contains `agents` and no output file contains `type: tasks`.
