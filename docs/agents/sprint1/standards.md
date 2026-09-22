---
sprint: 1
type: standards
---

# Sprint 1 standards — Snapback

The Lawkeeper's rule digest for this sprint. Every bullet below cites a real
source (spec section, a `~/.claude/rules/**` file, or `~/.claude/CLAUDE.md`).
Nothing here is invented; where the spec is silent, this file says so instead
of guessing.

## Detected stack

- Language: Go core, static target (`CGO_ENABLED=0`) — README.md §2, §5, §17.
- Native companion (out of scope for this Go digest, tracked separately): Swift, macOS Finder Sync extension — README.md §2, §9.3.
- No Rust, no other backend language — README.md §2, §21.
- FUSE (`macFUSE` on macOS, `fuse3` on Linux) is an external hard prerequisite, not part of the Go toolchain — README.md §2, §5.
- FUSE library dependency: `github.com/hanwen/go-fuse/v2`, pinned — README.md §5.
- CI: GitHub Actions; Conventional Commits + semantic-release for versioning — README.md §18.
- Toolchain: a **pinned** Go toolchain is required — README.md §18. The local environment currently has `go1.22.2` installed (see "Toolchain pinning" below for the open decision this creates).

Confirms the task brief: this is a Go project and the Go-specific rule set
(`~/.claude/rules/golang/*.md`) applies on top of the common rules
(`~/.claude/rules/common/*.md`).

## Active rules

**Go language & style**

- gofmt and goimports are mandatory, no style debates — `~/.claude/rules/golang/coding-style.md`.
- Accept interfaces, return structs; keep interfaces small (1-3 methods) — `~/.claude/rules/golang/coding-style.md`.
- Always wrap errors with context (`fmt.Errorf("...: %w", err)`) — `~/.claude/rules/golang/coding-style.md`.
- Define interfaces where they are used, not where they are implemented; prefer functional-options constructors and constructor-based dependency injection — `~/.claude/rules/golang/patterns.md`.
- Immutability: always create new objects, never mutate in place — `~/.claude/rules/common/coding-style.md`. The spec's own architecture leans on this same principle: catalog generations are immutable and published atomically — README.md §4 (`projection` module), §6.
- Many small files over few large ones (200-400 lines typical, 800 max); organize by feature/domain, high cohesion low coupling — `~/.claude/rules/common/coding-style.md`.
- Functions small (<50 lines), no deep nesting (>4 levels), no magic numbers, no hardcoded values — `~/.claude/rules/common/coding-style.md`.
- Validate only at system boundaries (user input, external APIs), fail fast with clear messages — `~/.claude/rules/common/coding-style.md`. The spec applies this directly to config: reject unknown fields and unsupported values with field-specific errors — README.md §12.

**Go-specific engineering practice**

- Use `exec.CommandContext` with argument arrays for every Restic invocation; never build a shell command from paths, repository URIs, passwords, tags or UI input — README.md §5.
- `CGO_ENABLED=0` for the Go core on every Linux target; static linkage must be verified with `ldd`/`file` before being called static, per target — README.md §5, §17.
- Always use `context.Context` for timeout control on finite operations; never apply cancellation/timeouts to the long-running mount — `~/.claude/rules/golang/security.md`; README.md §5.
- Use `gosec` for static security analysis — `~/.claude/rules/golang/security.md`. (Not installed locally — see Cross-cutting gates; not in the orchestrator's "at minimum" list, so it is documented as a recommended, non-blocking check rather than added to the enforced matrix.)

**Testing**

- Standard `go test` with table-driven tests — `~/.claude/rules/golang/testing.md`.
- Always run tests with `-race` — `~/.claude/rules/golang/testing.md`; README.md §18 ("unit and integration tests with `-race`").
- Coverage measured with `go test -cover` — `~/.claude/rules/golang/testing.md`; enforced minimum is 80% — `~/.claude/rules/common/testing.md`. See "How coverage is measured" below for the exact command.
- Mandatory TDD workflow: write the test first (RED), it must fail, write the minimal implementation (GREEN), refactor, then verify coverage — `~/.claude/rules/common/testing.md`.
- Real mount/browse/unmount tests on Linux runners and on macOS runners with macFUSE are required CI evidence, not just unit tests — README.md §18, §20.

**Security**

- No hardcoded secrets (API keys, passwords, tokens); use environment variables or a secret manager — `~/.claude/rules/common/security.md`; `~/.claude/rules/golang/security.md`.
- Credentials in config are references to protected files, never plaintext in YAML; no secrets in arguments or URLs; no secrets in status, logs or HTML — README.md §12.
- Mandatory pre-commit checklist: no hardcoded secrets, input validation, parameterized queries, sanitized HTML/XSS prevention, CSRF protection, auth verified, rate limiting, no sensitive data in error messages — `~/.claude/rules/common/security.md`. The web UI section of the spec restates several of these directly: bind to loopback, bootstrap token + authenticated session, reject unexpected Host/Origin, CSRF and session checks, no credentials in URLs, render filenames as text never HTML, validate every path against configured roots — README.md §15.

**CI / quality gate (cross-cutting, binds §"Cross-cutting gates" below)**

- CI must run: `go build`, `go vet`, `staticcheck`/`golangci-lint`, unit and integration tests with `-race`, cross-compilation of every release target, a coverage report, `govulncheck`, against a pinned Go toolchain — README.md §18.
- Conventional commits and semantic-release drive version bumps, tags and changelog — README.md §18; `~/.claude/rules/common/git-workflow.md`.
- Code review (CRITICAL/HIGH issues blocking) is required after writing or modifying code and before any commit to a shared branch — `~/.claude/rules/common/code-review.md`.
- Honesty gate: never claim "production-ready", "cross-platform", "static" or "Finder-integrated" without linked evidence; every unimplemented or untested requirement must be listed in the final report — README.md §1, §18, §22.

**Project docs (cross-cutting, not Go-specific)**

- Every project needs `CLAUDE.md` + `DEVLOG.md`; `DEVLOG.md` is a real-time session journal, updated in real time, not retroactively — `~/.claude/CLAUDE.md`; `~/.claude/rules/project-docs.md`.
- Self-anneal workflow: fix, update the script, test, update docs — `~/.claude/CLAUDE.md`.

## Cross-cutting gates

This is the matrix `gate-green-verify` and `gate-final` execute verbatim
(`run_standards_matrix` in the plugin's `bin/_gatelib.sh` greps this file for a
`## Cross-cutting gates` / `## Gate matrix` heading and runs every fenced
command line it finds, in order, failing the whole gate on the first nonzero
exit). Nothing below is decorative — every line in the fenced block actually
runs.

| Gate | Command | Threshold / notes | Local status | Source |
| --- | --- | --- | --- | --- |
| Format (gofmt) | `test -z "$(gofmt -l .)"` | must produce no output | INSTALLED (`gofmt` ships with `go`) | `~/.claude/rules/golang/coding-style.md` |
| Format (goimports) | `test -z "$(goimports -l .)"` | must produce no output | **NOT INSTALLED** — verified locally with `command -v goimports` (not in orchestrator's environment note; found independently) | `~/.claude/rules/golang/coding-style.md` |
| Build / typecheck | `CGO_ENABLED=0 go build ./...` | Go has no separate typecheck step; the compiler performs type-checking as part of `go build`, so this one command satisfies both rows | INSTALLED (`go1.22.2`) | README.md §2, §5, §17 |
| Vet | `go vet ./...` | static analysis, part of the standard toolchain | INSTALLED | README.md §18 |
| Lint | `golangci-lint run` | README.md §18 says "staticcheck/golangci-lint" (either/or); `golangci-lint` bundles the `staticcheck` linter, so this single command satisfies that requirement without a redundant second install/run | **NOT INSTALLED** (orchestrator-flagged) | README.md §18 |
| Test + race | `go test -race ./...` | | INSTALLED | README.md §18; `~/.claude/rules/golang/testing.md` |
| Coverage | `go test -race -covermode=atomic -coverprofile=coverage.out ./...` then the threshold check below | ≥80% total, see "How coverage is measured" | INSTALLED (`go test -cover` ships with `go`) | `~/.claude/rules/common/testing.md`; `~/.claude/rules/golang/testing.md` |
| Vulnerability audit | `govulncheck ./...` | | **NOT INSTALLED** (orchestrator-flagged) | README.md §18 |

**BLOCKING — install before dispatching any GREEN or FINAL gate worker:**
`goimports`, `golangci-lint`, `govulncheck` are not present in this
environment (verified with `command -v`). Their rows are kept in the
executable matrix below, not silently dropped, so the gate fails loudly with
"command not found" rather than passing a weaker matrix. The orchestrator
must install all three before any `gate-green-verify` / `gate-final` run is
expected to pass; local install commands (`go install
golang.org/x/tools/cmd/goimports@latest`, `go install
golang.org/x/vuln/cmd/govulncheck@latest`,
`brew install golangci-lint` or the pinned CI method) are an install-order
decision for the orchestrator, not something this file should prescribe.

**Also expected to fail right now for an unrelated reason:** the repository is
still greenfield (README.md, LICENSE, `.gitignore` only, per this task's own
brief). `go build`/`go vet`/`go test` will fail with "directory prefix . does
not contain main module" until stage-0 scaffolding creates `go.mod` and a
minimal `main.go` (verified locally: `gofmt -l .` exits 0 today, but `go vet
./...` and `go build ./...` already fail for lack of a module). That failure
is expected at the current tick and is not a tooling gap — README.md §22,
stage 0 exit evidence is "Green CI on an empty binary," which is the first
point this matrix is meant to pass.

**How coverage is measured.** Two-step: generate a race-enabled coverage
profile, then compute total statement coverage from `go tool cover -func` and
fail if it is under the 80% minimum from `~/.claude/rules/common/testing.md`.
Per-package coverage floors are not mandated by any source consulted, so none
is invented here; only the total is gated.

```bash
test -z "$(gofmt -l .)"
test -z "$(goimports -l .)"
CGO_ENABLED=0 go build ./...
go vet ./...
golangci-lint run
go test -race ./...
go test -race -covermode=atomic -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | awk '/^total:/{gsub("%","",$NF); if ($NF+0 < 80) { print "coverage " $NF "% is below the 80% minimum (common/testing.md)"; exit 1 } else { print "coverage " $NF "% OK (>=80%)" } }'
govulncheck ./...
```

**Not part of the routine per-task matrix above (scoped out deliberately, not
dropped silently):**

- Cross-compilation of every release target (linux/amd64, linux/arm64,
  darwin/amd64, darwin/arm64, plus unverified linux/arm, linux/mips,
  linux/mipsle) — README.md §17, §18. This is a CI/release-pipeline
  responsibility (stage 0 CI setup, stage 7 release), not a per-task local
  GREEN gate; running a 7-target cross-compile on every task would be far
  slower than the loop it is meant to protect.
- Static-linkage verification (`file`/`ldd` against a built Linux binary) —
  README.md §5, §17. This needs an actual compiled binary artifact and only
  applies to Linux targets; it belongs in the same CI/release step as
  cross-compilation, once there is a binary to inspect.
- `gosec ./...` — `~/.claude/rules/golang/security.md`. Recommended, but not
  in the orchestrator's "at minimum" list for this sprint's matrix; add it
  once `gosec` is confirmed available, without waiting on the three blocking
  installs above.

## Toolchain pinning (open item, not a rule to invent)

README.md §18 requires a "pinned Go toolchain" but does not name a version.
The local environment has `go1.22.2` installed. Go's `go.mod` supports a `go`
directive (minimum language version) and, since Go 1.21, an optional
`toolchain` directive (e.g. `toolchain go1.22.2`) that pins the exact
toolchain `go` will download and use for the module. Recording this option
here rather than picking a version: choosing the exact pinned version is a
planning decision (does the team pin to the currently installed `1.22.2`, or
a newer patch release?), not something this standards digest should decide
unilaterally.

## Sources consulted

- `/Users/adeelahmad/work/snapback/README.md` (Revision 2, all sections cited above by §).
- `/Users/adeelahmad/.claude/CLAUDE.md`.
- `/Users/adeelahmad/.claude/rules/common/coding-style.md`, `testing.md`, `security.md`, `code-review.md`, `git-workflow.md`.
- `/Users/adeelahmad/.claude/rules/golang/coding-style.md`, `testing.md`, `security.md`, `patterns.md`, `hooks.md`.
- `/Users/adeelahmad/.claude/rules/project-docs.md`.
- `/Users/adeelahmad/work/agentic-agile/plugin/agents/standards.md` (this role's own mandate/persona).
- `/Users/adeelahmad/work/agentic-agile/plugin/bin/_gatelib.sh`, `gate-green-verify`, `gate-final`, `gate-standards-cited` (gate mechanics this file's shape must satisfy).
- `/Users/adeelahmad/work/agentic-agile/plugin/schemas/planning-artifacts.kdl` (frontmatter shape for the `standards` type: `type`, `sprint`, both required, no fixed sections).
- Local tool probing via `command -v`: `gofmt`, `go` present; `goimports`, `golangci-lint`, `staticcheck`, `govulncheck`, `gosec` absent.
