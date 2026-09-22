---
type: init
story: C1
---

## C1/T1 · attempt 1 · red-worker · 2026-09-22T03:16:04Z

### Mandate (human request 2026-09-22: "update readme from example to exact domain we have got snapback.run")
Update the tests that pin placeholder domains so they require the real ones, each FAILING BY ASSERTION on current files:
1. `test/projectdocs/readme_test.go` installLineRe: require exactly `curl -fsSL https://snapback.run/install.sh | sh`; add a check that README contains no `example.com` / `example.invalid`.
2. `test/installer/dryrun_test.go`: the default base URL shown by the dry run must be `https://github.com/adeelahmad/snapback/releases/latest/download` (real release assets exist there); assert no `example.invalid`.
3. `test/docs/mkdocs_config_test.go`: site_url must be `https://snapback.run/`.
4. New test in `test/ci/` (or the existing docs-workflow test file if one exists): `.github/workflows/docs.yml` must copy `install.sh` into the built `site/` before upload-pages-artifact, so https://snapback.run/install.sh is served (today it 404s).
Tests only.

### Scope
May: test/projectdocs/readme_test.go, test/installer/dryrun_test.go, test/docs/mkdocs_config_test.go, one test file under test/ci/ or test/docs/ for the docs workflow. May Not: README.md, install.sh, mkdocs.yml, workflows, production code.

## C1/T1b · attempt 1 · red-worker · 2026-09-22T03:19:33Z

### Mandate
Scope extension approved by orchestrator: `test/projectdocs/readme_links_test.go` still pins the docs-site URL `https://adeelahmad.github.io/snapback/`. Change it to require `https://snapback.run/` (and forbid the github.io docs URL in README). Must FAIL BY ASSERTION on current README. Tests only; edit only that file.

## C1/T1 · attempt 1 · green-worker · 2026-09-22T03:22:10Z

### Mandate
Make every C1 test pass with the least change:
- README.md: Install line exactly `curl -fsSL https://snapback.run/install.sh | sh` (drop any "placeholder" wording in that section); Documentation docs-site link `https://snapback.run/`; no example.com / example.invalid / adeelahmad.github.io/snapback left.
- install.sh: default `SNAPBACK_BASE_URL` = `https://github.com/adeelahmad/snapback/releases/latest/download` (still overridable by env); keep POSIX sh, shellcheck clean.
- mkdocs.yml: `site_url: https://snapback.run/`.
- .github/workflows/docs.yml: after `mkdocs build --strict --site-dir site` and before upload-pages-artifact, a step `cp install.sh site/install.sh` (pinned actions unchanged); actionlint clean.
Honesty: docs must not claim more than exists. Do not push or deploy anything.

### Scope
May: README.md, install.sh, mkdocs.yml, .github/workflows/docs.yml. May Not: tests, anything else.

### Acceptance
`GOTOOLCHAIN=auto go test ./test/...` all PASS; full standards matrix green (actionlint, shellcheck -s sh install.sh, mkdocs build --strict); no AI trailers.
