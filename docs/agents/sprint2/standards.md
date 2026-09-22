---
sprint: 2
type: standards
---

# Sprint 2 standards — Snapback

The Lawkeeper's rule digest for Sprint 2, Stage 1 (the compatibility milestone:
go-fuse FUSE catalog, restic CLI via `exec`, rclone latency measurements,
crawler test — SPEC.md §22, stage 1 row). This file starts from
`docs/agents/sprint1/standards.md` and keeps every cited rule and gate
verbatim except where the task brief asked for an update; it adds only rules
that are sourced from SPEC.md, `~/.claude/rules/golang/*` or
`~/.claude/rules/common/*`. Nothing here is invented; where a source is
silent, this file says so instead of guessing.

## Detected stack (unchanged from Sprint 1, reconfirmed)

- Language: Go core, static target (`CGO_ENABLED=0`) — SPEC.md §5, §17; `CLAUDE.md` ("Stack").
- Go toolchain: `go.mod` pins `toolchain go1.27.1` for module `github.com/adeelahmad/snapback` — `/Users/adeelahmad/work/snapback/go.mod`; `CLAUDE.md` ("Stack") records this as `go 1.27` (toolchain `go1.27.1`), resolving Sprint 1's open "toolchain pinning" item.
- FUSE (`macFUSE` on macOS, `fuse3` on Linux) is an external hard prerequisite, not part of the Go toolchain — SPEC.md §1, §5.
- FUSE library dependency for the history catalog: `github.com/hanwen/go-fuse/v2`, pinned, behind a `mount` adapter interface — SPEC.md §5.
- Restic is driven only through the installed `restic` CLI via `exec`, never a library import — SPEC.md §5.
- rclone is a required, externally-configured executable for `rclone:REMOTE:PATH` repositories; Restic invokes it via `rclone serve restic` — SPEC.md §2, §12.
- CI: GitHub Actions; Conventional Commits + semantic-release for versioning — SPEC.md §18; `.github/workflows/ci.yml`, `release.yml`, `commitlint.yml`, `docs.yml`.
- No Rust, no other backend language; Restic is the only `SnapshotProvider` implementation in this release — SPEC.md §1, §21.

Confirms the task brief: this is a Go project and the Go-specific rule set
(`~/.claude/rules/golang/*.md`) applies on top of the common rules
(`~/.claude/rules/common/*.md`).

## Active rules

**Carried over from Sprint 1 (unchanged)**

- gofmt and goimports are mandatory, no style debates — `~/.claude/rules/golang/coding-style.md`.
- Accept interfaces, return structs; keep interfaces small (1-3 methods) — `~/.claude/rules/golang/coding-style.md`.
- Always wrap errors with context (`fmt.Errorf("...: %w", err)`) — `~/.claude/rules/golang/coding-style.md`.
- Define interfaces where they are used, not where they are implemented; prefer functional-options constructors and constructor-based dependency injection — `~/.claude/rules/golang/patterns.md`.
- Immutability: always create new objects, never mutate in place — `~/.claude/rules/common/coding-style.md`. The spec's own architecture leans on this same principle: each refresh publishes a new immutable catalog generation atomically — SPEC.md §6.
- Many small files over few large ones (200-400 lines typical, 800 max); organize by feature/domain, high cohesion low coupling — `~/.claude/rules/common/coding-style.md`.
- Functions small (<50 lines), no deep nesting (>4 levels), no magic numbers, no hardcoded values — `~/.claude/rules/common/coding-style.md`.
- Validate only at system boundaries (user input, external APIs), fail fast with clear messages — `~/.claude/rules/common/coding-style.md`.
- Use `exec.CommandContext` with argument arrays for every Restic invocation; never build a shell command from paths, repository URIs, passwords, tags or UI input — SPEC.md §5.
- `CGO_ENABLED=0` for the Go core on every Linux target; static linkage must be verified with `ldd`/`file` before being called static, per target — SPEC.md §5, §17.
- Always use `context.Context` for timeout control on finite operations; never apply cancellation/timeouts to the long-running mount — `~/.claude/rules/golang/security.md`; SPEC.md §5.
- Use `gosec` for static security analysis — `~/.claude/rules/golang/security.md`. (Recommended, non-blocking; see Cross-cutting gates — not in the orchestrator's literal command list for this sprint's matrix either.)
- Standard `go test` with table-driven tests — `~/.claude/rules/golang/testing.md`.
- Always run tests with `-race` — `~/.claude/rules/golang/testing.md`; SPEC.md §18, §20 ("unit and integration tests with `-race`"; "Unit, integration and race tests on the core").
- Coverage measured with `go test -cover`, enforced minimum 80% — `~/.claude/rules/golang/testing.md`; `~/.claude/rules/common/testing.md`.
- Mandatory TDD workflow: write the test first (RED), it must fail, write the minimal implementation (GREEN), refactor, then verify coverage — `~/.claude/rules/common/testing.md`.
- No hardcoded secrets (API keys, passwords, tokens); use environment variables or a secret manager — `~/.claude/rules/common/security.md`; `~/.claude/rules/golang/security.md`.
- Credentials in config are references to protected files, never plaintext in YAML; no secrets in arguments or URLs; no secrets in status, logs or HTML — SPEC.md §12.
- Code review (CRITICAL/HIGH issues blocking) is required after writing or modifying code and before any commit to a shared branch — `~/.claude/rules/common/code-review.md`.
- Conventional commits and semantic-release drive version bumps, tags and changelog — SPEC.md §18; `~/.claude/rules/common/git-workflow.md`.
- Every project needs `CLAUDE.md` + `DEVLOG.md`; `DEVLOG.md` is a real-time session journal, updated in real time, not retroactively — `~/.claude/CLAUDE.md`; `~/.claude/rules/project-docs.md`.
- Self-anneal workflow: fix, update the script, test, update docs — `~/.claude/CLAUDE.md`.

**New for Sprint 2 / Stage 1 — sourced from SPEC.md**

- Every Restic subprocess invocation uses `exec.CommandContext` with argument arrays and never a shell; never build a command from paths, repository URIs, passwords, tags or UI input — SPEC.md §5. (Restated above with the Sprint 1 citation; repeated here per the task brief's explicit list because it is the load-bearing rule for the restic-CLI-via-exec work this sprint.)
- Pin `github.com/hanwen/go-fuse/v2` for the history catalog behind a `mount` adapter interface, so a tested substitute (documented, e.g. a cgofuse fallback) could be swapped in without touching `resolver`, `projection`, `links`, `discovery`, `daemon` or `web` — SPEC.md §5.
- Do not import `restic/internal/...` or fork Restic to reach an internal API; the CLI boundary is deliberate even though both projects are Go — SPEC.md §5.
- Never identify a snapshot by a short prefix alone; store and use full repository identity and full snapshot identity — SPEC.md §5.
- The virtual catalog implements only lookup, attributes, directory enumeration, readlink, read-only generated-file reads, statfs and lifecycle; it returns `EROFS` for any mutation attempt — SPEC.md §7.
- Tests use generated, disposable temporary fixtures and a small real (but non-production) Restic repository; never test mutation behaviour against a real home directory or a production repository — SPEC.md §20.
- Honest reporting: nothing is called "production-ready", "cross-platform", "static" or "Finder-integrated" without linked evidence; the implementation report ends with a requirement matrix — **implemented and tested**, **implemented but not tested here**, or **not implemented** — with the exact reason for any unmet item, including every untested OS/architecture/service/FUSE combination — SPEC.md §1, §20, §22.

**Testing — Stage 1 specifics**

- Real mount/browse/unmount tests on Linux runners and on macOS runners with macFUSE are required CI evidence, not just unit tests — SPEC.md §18, §20.
- Verify `restic mount --path-template ids/%I` against the pinned Restic version before relying on it — SPEC.md §5.
- Verify per platform that displayed mtimes inside `.snapshot/<timestamp>/` match the backed-up values; document, do not claim, ctime/birth-time approximation by FUSE — SPEC.md §5.
- Real rclone/Google Drive cold-listing, warm-listing and cold-read latency numbers are reported from Stage 1; these decide go/no-go on the local hot-cache roadmap item — SPEC.md §20, §21, §22.
- A crawler-safety test runs `rg`, `rg -L`, `fd -L` and a VS Code search over a seeded tree and measures catalog hits — SPEC.md §7 (Stage 1 is named explicitly in this same paragraph).

## Cross-cutting gates

This is the matrix `gate-green-verify` and `gate-final` execute verbatim
(`run_standards_matrix` in the plugin's `bin/_gatelib.sh` greps this file for a
`## Cross-cutting gates` / `## Gate matrix` heading and runs every fenced
command line it finds, in order, failing the whole gate on the first nonzero
exit). Nothing below is decorative — every line in the fenced block actually
runs. This updates Sprint 1's matrix to the tool set the task brief specified
and that `.github/workflows/ci.yml` / `docs.yml` / `release.yml` now run; all
listed tools were verified present locally with `command -v` at the time this
file was written.

| Gate | Command | Threshold / notes | Local status | Source |
| --- | --- | --- | --- | --- |
| Format (gofmt) | `test -z "$(gofmt -l .)"` | must produce no output | INSTALLED | `~/.claude/rules/golang/coding-style.md`; `.github/workflows/ci.yml` |
| Format (goimports) | `test -z "$(goimports -l .)"` | must produce no output | INSTALLED | `~/.claude/rules/golang/coding-style.md`; `.github/workflows/ci.yml` |
| Build | `CGO_ENABLED=0 go build ./...` | static Linux target; Go's compiler performs type-checking as part of build, so no separate typecheck step exists | INSTALLED (`go` toolchain per `go.mod`) | SPEC.md §5, §17; `.github/workflows/ci.yml` |
| Vet | `go vet ./...` | | INSTALLED | SPEC.md §18; `.github/workflows/ci.yml` |
| Lint | `golangci-lint run` | config `.golangci.yml` (v2 schema): `errcheck`, `govet`, `staticcheck` linters, `goimports` formatter | INSTALLED (`golangci-lint`, config present at repo root) | SPEC.md §18 ("staticcheck/golangci-lint"); `.golangci.yml`; `.github/workflows/ci.yml` (pinned `v2.13.2` action) |
| Test + race | `go test -race ./...` | | INSTALLED | SPEC.md §18, §20; `~/.claude/rules/golang/testing.md` |
| Coverage | `go test -race -covermode=atomic -coverprofile=coverage.out ./...` then the threshold check below | ≥80% total, see "How coverage is measured" | INSTALLED | `~/.claude/rules/common/testing.md`; `~/.claude/rules/golang/testing.md`; `.github/workflows/ci.yml` |
| Vulnerability audit | `govulncheck ./...` | | INSTALLED | SPEC.md §18; `.github/workflows/ci.yml` (pinned `v1.8.0`) |
| GitHub Actions lint | `actionlint` | lints every workflow under `.github/workflows/` | INSTALLED | task brief (orchestrator-issued gate list, `issued_by: "@orchestrator"` on this file's own init); no repo workflow runs it yet, so it is not double-sourced from `ci.yml` — it is added here because the orchestrator specified it as a gate this sprint runs |
| Shell lint | `shellcheck -s sh install.sh` | `install.sh` starts `#!/bin/sh` (POSIX), so `-s sh` matches its actual shebang; `.github/workflows/ci.yml` currently runs a broader `shellcheck` over every tracked `*.sh` with no `-s` flag — this row narrows to what the task brief specified | INSTALLED | task brief; `install.sh` (`#!/bin/sh` shebang); `.github/workflows/ci.yml` (existing broader shellcheck step) |
| Docs build | `mkdocs build --strict --site-dir site` | fails on any warning (broken links, nav errors); matches the existing CI step verbatim | INSTALLED | SPEC.md §18 ("Docs site on GitHub Pages... MkDocs Material"); `.github/workflows/docs.yml`; `mkdocs.yml`, `requirements-docs.txt` |
| Release config check | `goreleaser check` | validates `.goreleaser.yaml` without building/publishing | INSTALLED (`goreleaser`, confirmed by the task brief as "installed locally now") | SPEC.md §17, §18 (release pipeline); `.goreleaser.yaml`; task brief |

**How coverage is measured.** Two-step: generate a race-enabled coverage
profile, then compute total statement coverage from `go tool cover -func` and
fail if it is under the 80% minimum from `~/.claude/rules/common/testing.md`.
Per-package coverage floors are not mandated by any source consulted, so none
is invented here; only the total is gated. This matches the coverage step
already implemented in `.github/workflows/ci.yml`.

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
actionlint
shellcheck -s sh install.sh
mkdocs build --strict --site-dir site
goreleaser check
```

**Not part of the routine per-task matrix above (scoped out deliberately, not
dropped silently, unchanged from Sprint 1's reasoning):**

- Cross-compilation of every release target (linux/amd64, linux/arm64,
  darwin/amd64, darwin/arm64, plus unverified linux/arm, linux/mips,
  linux/mipsle) — SPEC.md §17, §18; `.github/workflows/ci.yml` (`cross-compile`
  job). CI/release-pipeline responsibility, not a per-task local GREEN gate.
- Static-linkage verification (`file`/`ldd` against a built Linux binary) —
  SPEC.md §5, §17; `.github/workflows/ci.yml` ("Static linkage evidence" step
  in the `cross-compile` job). Needs an actual compiled binary artifact and
  only applies to Linux targets.
- `gosec ./...` — `~/.claude/rules/golang/security.md`. Recommended, but still
  not in the orchestrator's literal command list for this sprint's matrix.
- Commit-message linting (`commitlint`) — `.github/workflows/commitlint.yml`;
  `~/.claude/rules/common/git-workflow.md`. This runs on PR events against
  commit ranges, not as a per-task local gate command.

## Sprint 2 / Stage 1 integration-test gating (FUSE, restic, network)

Stage 1 work (go-fuse catalog, restic CLI mount/browse/unmount, rclone
latency) needs real prerequisites — a FUSE implementation, an installed
`restic` binary, and in one case network access to a real rclone remote —
that are not present in every environment this matrix runs in. SPEC.md §20
requires "Real mount/browse/unmount tests" and "Real rclone/Google Drive...
numbers... from stage 1"; it does not specify a gating mechanism, so the
mechanism itself (build tags, env vars) is this file's own reasonable
translation of that requirement into an executable gate, not a fabricated
SPEC.md citation — the task brief also directs it explicitly. Nothing here
was invented from README/SPEC text.

- **Build tag.** Integration tests that require FUSE or a real `restic`
  binary live in files tagged `//go:build integration` (excluded from the
  default `go test -race ./...` run in the matrix above) and are run
  separately with `go test -race -tags=integration ./...` once the
  prerequisite env vars below are confirmed set.
- **Env vars, checked at test setup, one per prerequisite:**
  - `SNAPBACK_FUSE_TESTS=1` — gates any test that mounts the go-fuse catalog
    or a real `restic mount`. Requires `fuse3` (Linux) or `macFUSE` (macOS)
    to actually be installed; the test must probe for the device/binary
    itself, not just trust the env var.
  - `SNAPBACK_RCLONE_REMOTE` — set to a real `rclone` remote spec (e.g.
    `gdrive:Backups/restic-test`) to gate the SPEC.md §20/§21 cold/warm
    latency measurement tests. Absent, those tests skip.
- **Skip discipline, not a suppression loophole.** A test gated on a missing
  prerequisite must call `t.Skip(...)` (or the harness equivalent) with a
  message naming the exact missing prerequisite (e.g. `"SNAPBACK_FUSE_TESTS
  not set: FUSE catalog tests skipped"`, `"fuse3 device /dev/fuse not
  present"`, `"SNAPBACK_RCLONE_REMOTE not set: rclone latency tests
  skipped"`) — never a bare, unexplained skip. This is the task brief's
  explicit instruction; SPEC.md §1 and §22 already require that "every
  unimplemented or untested requirement is listed" and that the final report
  states the exact reason for any unmet item, which a silent skip would
  violate.
  - `gate-final` counts skips: a run of the integration suite where
    prerequisite-gated tests skipped must be reported as **implemented but
    not tested here** in the Stage 1 exit report (SPEC.md §22's three-way
    matrix), never folded into "implemented and tested." This is how the
    honest-reporting rule above (SPEC.md §1, §20, §22) applies concretely to
    a skipped integration test.
- **CI target.** Linux CI runners for this stage must have `fuse3` installed
  so `SNAPBACK_FUSE_TESTS=1` integration tests actually run in CI rather than
  perpetually skipping — SPEC.md §18 ("Real mount/browse/unmount tests on
  Linux runners and on macOS runners with macFUSE"); SPEC.md §20 (same
  requirement, acceptance and platform-proof sections). Whether that runner
  provisioning step is added to `.github/workflows/ci.yml` in this sprint is
  a planning/implementation decision, not something this standards file
  performs; the requirement that it exist before the Stage 1 exit is claimed
  complete is sourced from SPEC.md §18/§20 above.
- `SNAPBACK_RCLONE_REMOTE`-gated tests reach a real external network service;
  they must never run as part of the default local/CI matrix above without
  the env var explicitly set, consistent with SPEC.md §20's fixture rule
  (disposable, never a production repository) extended to network remotes
  by the same reasoning the task brief applies.

## Sources consulted

- `/Users/adeelahmad/work/snapback/SPEC.md` (Revision 2, all sections cited above by §; supersedes the README.md citations used in Sprint 1's file, which SPEC.md itself supersedes per its title line).
- `/Users/adeelahmad/work/snapback/docs/agents/sprint1/standards.md` (Sprint 1 digest this file starts from).
- `/Users/adeelahmad/work/snapback/CLAUDE.md`, `ORCHESTRATOR.md` (stage/sprint mapping, current toolchain).
- `/Users/adeelahmad/work/snapback/go.mod`, `.golangci.yml`, `install.sh`, `.goreleaser.yaml`, `mkdocs.yml`.
- `/Users/adeelahmad/work/snapback/.github/workflows/ci.yml`, `docs.yml`, `release.yml`, `commitlint.yml`.
- `/Users/adeelahmad/.claude/CLAUDE.md`.
- `/Users/adeelahmad/.claude/rules/common/coding-style.md`, `testing.md`, `security.md`, `code-review.md`, `git-workflow.md`.
- `/Users/adeelahmad/.claude/rules/golang/coding-style.md`, `testing.md`, `security.md`, `patterns.md`, `hooks.md`.
- `/Users/adeelahmad/.claude/rules/project-docs.md`.
- `/Users/adeelahmad/work/agentic-agile/plugin/bin/_gatelib.sh`, `gate-standards-cited`, `selfcheck` (gate mechanics this file's shape must satisfy).
- This task's own init frontmatter (`issued_by: "@orchestrator"`), for the gate-matrix tool list and the FUSE/restic/network gating requirement, which are orchestrator directives rather than SPEC.md text — cited as such above wherever no SPEC.md § covers the specific tool choice.
- Local tool probing via `command -v`: `gofmt`, `goimports`, `go`, `golangci-lint`, `govulncheck`, `actionlint`, `shellcheck`, `mkdocs`, `goreleaser` all present.
