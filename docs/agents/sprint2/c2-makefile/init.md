---
type: init
story: C2
---

## C2/T1 · attempt 1 · red-worker · 2026-09-22T03:16:04Z

### Mandate (human request 2026-09-22: "make sure we have make file taking care of install build etc")
Write `test/makefile/makefile_test.go` (package makefile_test) pinning a repo-root `Makefile`; each test FAILS BY ASSERTION today (Makefile absent → t.Fatal naming it, not a compile error):
- Targets exist (parse `^name:` rules): help (and is the default/first target), build, install, uninstall, test, cover, lint, fmt, fmt-check, vet, vuln, docs, release-check, ci, clean; all listed in `.PHONY`.
- `make -n build` (skip if `make` not on PATH, naming it) prints a `go build` with `CGO_ENABLED=0`, output `bin/snapback`, package `./cmd/snapback`, and `-ldflags` setting `github.com/adeelahmad/snapback/internal/version.Version`, `.Commit` and `.Target`.
- `make -n install PREFIX=/tmp/x` installs `bin/snapback` to `/tmp/x/bin/snapback` with mode 0755 and honours DESTDIR; `make -n uninstall PREFIX=/tmp/x` removes it.
- `make -n ci` includes every Cross-cutting gate command from docs/agents/sprint2/standards.md (gofmt -l, goimports -l, go build, go vet, golangci-lint run, go test -race, coverage >=80 check, govulncheck, actionlint, shellcheck -s sh install.sh, mkdocs build --strict, goreleaser check).
- `GOTOOLCHAIN` defaults to `auto` (exported, overridable).
- CONTRIBUTING.md mentions `make ci` and `make install`.
Tests only; the test must not actually build or install anything (use `make -n`).

### Scope
May: test/makefile/makefile_test.go. May Not: Makefile, CONTRIBUTING.md, anything else.

## C2/T1 · attempt 1 · green-worker · 2026-09-22T03:20:10Z

### Mandate
Add a repo-root `Makefile` (GNU make) that makes all 8 tests in `test/makefile/makefile_test.go` pass, and add a short "Make targets" note to `CONTRIBUTING.md` mentioning `make ci` and `make install`. Targets: help (first/default; lists targets), build (CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X github.com/adeelahmad/snapback/internal/version.Version=$(VERSION) -X ….Commit=$(COMMIT) -X ….Target=$(GOOS)/$(GOARCH)" -o bin/snapback ./cmd/snapback; VERSION from `git describe --tags --always --dirty`, COMMIT from `git rev-parse --short HEAD`), install (install -d + install -m 0755 bin/snapback $(DESTDIR)$(PREFIX)/bin/snapback; PREFIX ?= /usr/local), uninstall, test (go test -race ./...), cover (coverage profile + fail if total <80%), lint, fmt, fmt-check, vet, vuln, docs (mkdocs build --strict), release-check (goreleaser check), ci (every Cross-cutting gate in standards.md, in order), clean (rm -rf bin/ site/ coverage.out). `export GOTOOLCHAIN ?= auto`. All targets in .PHONY. Keep it simple and readable; pin nothing to 'latest'. Match the version ldflags used by .goreleaser.yaml.

### Scope
May: Makefile, CONTRIBUTING.md, .gitignore (add `/bin/` — it is not ignored today). May Not: tests, other files.

### Acceptance
`GOTOOLCHAIN=auto go test ./test/...` all PASS (incl. test/makefile, test/community); `make -n ci` sane; full standards matrix green (run `make ci` itself as the matrix if it passes). No AI trailers.
