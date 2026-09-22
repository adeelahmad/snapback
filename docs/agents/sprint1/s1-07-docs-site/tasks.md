---
type: tasks
story: S1-07
---
# S1-07 — Docs site (MkDocs Material) and GitHub Pages deploy: tasks

Owned set (stories.md S1-07): `mkdocs.yml`, `docs-site/**`, `requirements-docs.txt`,
`.github/workflows/docs.yml`, `test/docs/**`. Spec: README.md §18 (docs site on GitHub
Pages, generated from repo Markdown, built and deployed by CI, versioned per release;
landing page leads with the restore story and one-line install). Stage 0 ships a
skeleton that builds and deploys, not full content. Depends on S1-01 (`go.mod`, module
`github.com/adeelahmad/snapback`). Tests are Go standard library only and check YAML
as text (no YAML parser, no new `go.mod` dependency). `docs_dir` is `docs-site` so
`docs/agents/**` is never published. Honesty gate applies to every file in
`docs-site/**`: no "production-ready", "cross-platform", "static", "Finder-integrated",
and no backend other than Restic is named. Per-release versioning is deferred and the
page says so; it is not faked.

Each task owns disjoint files, except that T1's helpers are read by T2-T7, so T1 runs
first; T2-T6 may then run in parallel; T7 runs last because it builds the whole site.

## T1 — Shared test helpers for the docs package
Files: `test/docs/helpers_test.go`.
`package docs` helpers: `repoRoot(t)` (walk up to the dir holding `go.mod`),
`readRepoFile(t, rel) string` (fails naming the path if missing or empty), and
`yamlScalar(text, key) (string, bool)` returning the value of a top-level `key: value`
line with surrounding quotes stripped. Test-only; no production code.

## T2 — Pinned docs requirements
Files: `requirements-docs.txt`, `test/docs/requirements_test.go`.
Exactly two non-comment lines, each pinned with `==`: `mkdocs==1.6.1` and
`mkdocs-material==9.6.14`. No unpinned or range (`>=`, `~=`) specifiers.

## T3 — mkdocs.yml configuration
Files: `mkdocs.yml`, `test/docs/mkdocs_config_test.go`.
Top-level keys: `site_name: Snapback`, `site_url: https://adeelahmad.github.io/snapback/`,
`repo_url: https://github.com/adeelahmad/snapback`, `docs_dir: docs-site`,
`theme:` with `name: material`, and a `nav:` list whose only entry is `- Home: index.md`.
No `exclude_docs`/`docs_dir` pointing at `docs`. No plugins beyond the material theme defaults.

## T4 — Landing page stub
Files: `docs-site/index.md`, `test/docs/landing_test.go`.
First `# ` heading is `# Snapback`; the first paragraph after it tells the restore
story (contains the word `restore`). Sections `## Install` (states the one-line install
is coming soon and marks it `Coming soon`), `## Status` (states `pre-release`) and
`## Versions` (states per-release versioned docs are not yet published). Mentions
Restic as the backend; names no other backend.

## T5 — Docs workflow: build on PR, deploy to Pages on master and tags
Files: `.github/workflows/docs.yml`, `test/docs/workflow_test.go`.
Triggers: `pull_request`, `push` with `branches: [master]` and `tags: ['v*']`.
Job `build`: checkout, `actions/setup-python` (python 3.12, pip cache on
`requirements-docs.txt`), `pip install -r requirements-docs.txt`,
`mkdocs build --strict --site-dir site`, then `actions/upload-pages-artifact` with
`path: site` (guarded `if: github.event_name != 'pull_request'`). Job `deploy`:
`needs: build`, same non-PR guard, `environment: github-pages`,
`permissions:` `pages: write` and `id-token: write`, step `actions/deploy-pages`.
Workflow-level `permissions: contents: read`; `concurrency: group: pages`.

## T6 — Honesty and backend scan of published content
Files: `test/docs/honesty_test.go`.
Walk every file under `docs-site/` plus `mkdocs.yml`; fail on any case-insensitive
banned phrase (production-ready, cross-platform, static, finder-integrated) or any
non-Restic backend name (borg, borgbackup, kopia, duplicity, duplicati, tarsnap).
Also assert nothing under `docs-site/` is a path copied from `docs/agents`.

## T7 — Optional strict build smoke
Files: `test/docs/build_test.go`.
If `mkdocs` is not on PATH, `t.Skip("mkdocs not installed")`. Otherwise run
`mkdocs build --strict --config-file <root>/mkdocs.yml --site-dir <t.TempDir()>`,
expect exit 0, `index.html` present, and no file under the output whose path contains
`agents`.
