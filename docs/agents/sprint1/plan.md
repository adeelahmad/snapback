---
type: sprint-plan
sprint: 1
stage: "2"
---

# Sprint 1 plan — Stage 0: Scaffolding (Stage-2 contract)

Stories and intents: `docs/agents/sprint1/stories.md`. Gate rules: `docs/agents/sprint1/standards.md`.

## Sprint goal

Stand up Snapback Stage 0 with no feature code. The sprint delivers a `snapback version` binary with ldflags-injected version, commit and target; green GitHub Actions CI (build, vet, lint, race, coverage, govulncheck, 7-target cross-compile with the unverified targets labelled); conventional-commit enforcement; a semantic-release, GoReleaser and cosign pipeline; an MkDocs Material site on GitHub Pages; repository furniture; and an `install.sh` skeleton. Exit evidence (§22 row 0): green CI on the binary, and the docs site deploys.

## Totals

8 stories, 43 tasks, 177 tests (unique `Test*` functions across the story `plan.md` files).

## Waves

Story-level waves are unchanged from Stage 1: no story appears in two waves. Inside each wave, tasks are dispatched in ticks. A tick lists the tasks that may run concurrently. Their file sets are pairwise disjoint, and every task in a tick depends only on tasks merged green in earlier ticks.

| Wave | Stories | Starts when |
| --- | --- | --- |
| 1 | S1-01 | immediately |
| 2 | S1-02, S1-03, S1-04, S1-05, S1-06, S1-07, S1-08 | S1-01 T4 merged green |

### Wave 1: S1-01 (serial, 1 worker per tick)

| Tick | Task | Files | Tests |
| --- | --- | --- | --- |
| 1.1 | S1-01 T1 | `go.mod`, `.gitignore` | 0 (gate-only: build/vet) |
| 1.2 | S1-01 T2 | `internal/version/version.go`, `internal/version/version_test.go` | 3 |
| 1.3 | S1-01 T3 | `cmd/snapback/run.go`, `cmd/snapback/run_test.go` | 2 |
| 1.4 | S1-01 T4 | `cmd/snapback/main.go`, `cmd/snapback/main_test.go` | 3 |

The tasks are serial because each one needs its predecessor: T2 needs `go.mod`, T3 imports `internal/version`, and T4 shares package `main` with T3.

### Wave 2, tick 2.1: 7 concurrent workers (story foundations)

| Task | Files (exclusive this tick) |
| --- | --- |
| S1-02 T1 | `.github/workflows/ci.yml`, `test/ci/ci_test.go`, `test/ci/helpers_test.go` |
| S1-03 T1 | `install.sh`, `test/installer/helpers_test.go`, `test/installer/detect_test.go` |
| S1-04 T1 | `.commitlintrc.json`, `test/commitlint/helpers_test.go`, `test/commitlint/config_test.go` |
| S1-05 T1 | `test/community/helpers_test.go` |
| S1-06 T1 | `.goreleaser.yaml`, `test/release/helpers_test.go`, `test/release/goreleaser_build_test.go` |
| S1-07 T1 | `test/docs/helpers_test.go` |
| S1-08 T1 | `README.md` -> `SPEC.md` (`git mv`, byte-for-byte), `test/projectdocs/helpers_test.go`, `test/projectdocs/spec_test.go` |

### Wave 2, tick 2.2: 21 concurrent workers (dispatch in batches of at least 5; any subset is safe)

| Task | Files (exclusive this tick) |
| --- | --- |
| S1-02 T2 | `.github/workflows/ci.yml` (edit), `test/ci/matrix_test.go` |
| S1-03 T2 | `install.sh`, `test/installer/detect_test.go` |
| S1-04 T2 | `.github/workflows/commitlint.yml`, `test/commitlint/workflow_test.go` |
| S1-05 T2 | `CONTRIBUTING.md`, `test/community/contributing_test.go` |
| S1-05 T3 | `SECURITY.md`, `test/community/security_test.go` |
| S1-05 T4 | `CODE_OF_CONDUCT.md`, `test/community/conduct_test.go` |
| S1-05 T5 | `.github/ISSUE_TEMPLATE/bug_report.md`, `.github/ISSUE_TEMPLATE/feature_request.md`, `.github/ISSUE_TEMPLATE/config.yml`, `test/community/issue_templates_test.go` |
| S1-05 T6 | `.github/pull_request_template.md`, `test/community/pr_template_test.go` |
| S1-06 T2 | `.goreleaser.yaml` (edit), `test/release/goreleaser_artifacts_test.go` |
| S1-06 T3 | `.releaserc.json`, `CHANGELOG.md`, `test/release/releaserc_test.go` |
| S1-06 T4 | `.github/workflows/release.yml`, `test/release/workflow_test.go` |
| S1-07 T2 | `requirements-docs.txt`, `test/docs/requirements_test.go` |
| S1-07 T3 | `mkdocs.yml`, `test/docs/mkdocs_config_test.go` |
| S1-07 T4 | `docs-site/index.md`, `test/docs/landing_test.go` |
| S1-07 T5 | `.github/workflows/docs.yml`, `test/docs/workflow_test.go` |
| S1-07 T6 | `test/docs/honesty_test.go` |
| S1-08 T2 | `README.md` (new), `test/projectdocs/readme_test.go` |
| S1-08 T5 | `ARCHITECTURE.md`, `test/projectdocs/architecture_test.go` |
| S1-08 T6 | `NOTICE`, `test/projectdocs/notice_test.go` |
| S1-08 T7 | `CLAUDE.md`, `test/projectdocs/claude_md_test.go` |
| S1-08 T8 | `DEVLOG.md`, `test/projectdocs/devlog_test.go` |

S1-07 T6 scans `docs-site/**` and `mkdocs.yml` while T3 and T4 write them. Its worktree reads them only after they merge, so a T6 GREEN that fails on merged content goes back to the owning task (T3 or T4). T6 never edits those files.

### Wave 2, tick 2.3: 6 concurrent workers

| Task | Files (exclusive this tick) |
| --- | --- |
| S1-02 T3 | `.golangci.yml`, `.github/workflows/ci.yml` (lint `version:` pin only), `test/ci/lint_test.go` |
| S1-03 T3 | `install.sh`, `test/installer/dryrun_test.go` |
| S1-04 T3 | `.github/workflows/commitlint.yml`, `test/commitlint/workflow_pins_test.go` |
| S1-05 T7 | `test/community/honesty_test.go` (reads all S1-05 files, so it runs after T2 through T6) |
| S1-07 T7 | `test/docs/build_test.go` (builds the whole site, so it runs after T2 through T6) |
| S1-08 T3 | `README.md`, `test/projectdocs/readme_links_test.go` |

### Wave 2, tick 2.4: 2 concurrent workers

| Task | Files |
| --- | --- |
| S1-03 T4 | `install.sh`, `test/installer/install_test.go` |
| S1-08 T4 | `test/projectdocs/honesty_test.go`, `README.md` (only if the scan fails) |

### Wave 2, ticks 2.5 to 2.7: S1-03 tail (1 worker each)

| Tick | Task | Files |
| --- | --- | --- |
| 2.5 | S1-03 T5 | `install.sh`, `test/installer/install_test.go` |
| 2.6 | S1-03 T6 | `install.sh`, `test/installer/static_test.go` |
| 2.7 | S1-03 T7 | `test/installer/shellcheck_test.go` (last, so any shellcheck fix to `install.sh` never collides with T2 through T6) |

Idle capacity in ticks 2.4 to 2.7 is expected. S1-03 is the serial bottleneck because all six of its implementation tasks edit `install.sh`.

## Dependency graph (task level)

```
S1-01: T1 -> T2 -> T3 -> T4 ──> (all of wave 2)
S1-02: T1 -> T2 -> T3                    (all edit ci.yml)
S1-03: T1 -> T2 -> T3 -> T4 -> T5 -> T6 -> T7  (T1-T6 edit install.sh; T7 last)
S1-04: T1 -> T2 -> T3                    (T2/T3 edit commitlint.yml; T2 uses T1 helper)
S1-05: T1 -> {T2, T3, T4, T5, T6} -> T7
S1-06: T1 -> {T2, T3, T4}                (T1 -> T2 share .goreleaser.yaml)
S1-07: T1 -> {T2, T3, T4, T5, T6} -> T7
S1-08: T1 (git mv README.md SPEC.md) -> {T2 -> T3 -> T4, T5, T6, T7, T8}
```

Soft cross-story contracts are textual only. They share no files and impose no ordering, and each side tests its own half:
- S1-03 and S1-06 share the asset naming contract: `snapback_<os>_<arch>[_unverified].tar.gz`, no version in the name, and `checksums.txt`.
- S1-06 depends on S1-01 through the ldflags `-X` paths `github.com/adeelahmad/snapback/internal/version.{Version,Commit,Target}`, which are fixed in S1-01 tasks.md.
- S1-04 feeds S1-06 softly: conventional commits drive semantic-release.

## Parallelism and cross-story file ownership

| Story | Wave | Owned files (exclusive) | Tasks | Tests |
| --- | --- | --- | --- | --- |
| S1-01 | 1 | `go.mod`, `cmd/snapback/**`, `internal/version/**`, `.gitignore` (no `go.sum`; stdlib only) | 4 | 8 |
| S1-02 | 2 | `.github/workflows/ci.yml`, `.golangci.yml`, `test/ci/**` | 3 | 22 |
| S1-03 | 2 | `install.sh`, `test/installer/**` | 7 | 17 |
| S1-04 | 2 | `.github/workflows/commitlint.yml`, `.commitlintrc.json`, `test/commitlint/**` | 3 | 14 |
| S1-05 | 2 | `CONTRIBUTING.md`, `SECURITY.md`, `CODE_OF_CONDUCT.md`, `.github/ISSUE_TEMPLATE/**`, `.github/pull_request_template.md`, `test/community/**` | 7 | 20 |
| S1-06 | 2 | `.github/workflows/release.yml`, `.releaserc.json`, `.goreleaser.yaml`, `CHANGELOG.md`, `test/release/**` | 4 | 34 |
| S1-07 | 2 | `mkdocs.yml`, `docs-site/**`, `requirements-docs.txt`, `.github/workflows/docs.yml`, `test/docs/**` | 7 | 26 |
| S1-08 | 2 | `README.md`, `SPEC.md`, `ARCHITECTURE.md`, `NOTICE`, `CLAUDE.md`, `DEVLOG.md`, `test/projectdocs/**` | 8 | 36 |
| **Total** | | | **43** | **177** |

Ownership rules the orchestrator enforces at dispatch:
- `README.md` belongs to S1-08 only. S1-08 T1 moves it to `SPEC.md`, and T2 through T4 author the new one. No other story, and no orchestrator bookkeeping, may touch `README.md` or `SPEC.md`. Other stories link to the README; they never edit it.
- `DEVLOG.md` and `CLAUDE.md` at the repo root belong to S1-08 (T8 and T7). The orchestrator must not write them while those tasks are in flight.
- `.github/` is split by exact file or subdirectory, and `test/` is split by subdirectory. No path appears in two story rows. `LICENSE` (MIT) already exists and no task touches it.
- Within a story, the same file never appears in two tasks of the same tick (see the tick tables).

Maximum fan-out is 21 concurrent workers (tick 2.2). Ticks 2.1, 2.2 and 2.3 each meet the orchestrator's minimum of 5 workers per tick.

## Critical path

S1-01 T1 -> T2 -> T3 -> T4 -> S1-03 T1 -> T2 -> T3 -> T4 -> T5 -> T6 -> T7 -> orchestrator exit-evidence verification. That is 11 serial task steps (ticks 1.1 to 2.7).

- S1-01 is the only wave-1 story, and its four tasks are strictly serial.
- S1-03 is the longest wave-2 chain: six serial `install.sh` edits, then the shellcheck tail. The next longest chains are S1-08 T1 -> T2 -> T3 -> T4 (4 steps) and S1-02/S1-04/S1-05/S1-07 (3 steps each).
- S1-06 carries the most risk: it has the most tests (34), the ldflags coupling to S1-01, and a semantic-release dry run that can only be verified on GitHub. It is still off the critical path at 2 steps.
- After wave 2 merges, the orchestrator verifies what no unit test can, as post-merge exit evidence: (1) the CI run on `master` is green (S1-02); (2) the Pages deploy succeeds and `https://adeelahmad.github.io/snapback/` returns 200 (S1-07); (3) the release `workflow_dispatch` dry run computes a version (S1-06); (4) the commitlint check runs on the first PR (S1-04).

Blocking prerequisite (from standards.md): the orchestrator installs `goimports`, `golangci-lint` and `govulncheck` locally before any RED, GREEN or FINAL gate. S1-01 pins `go 1.27` / `toolchain go1.27.1`. The local Go is 1.22.2, so gates need `GOTOOLCHAIN=auto` and network access to download it.

## Per-story plan pointers

| Story | Stage-2 dir | Tasks | Tests |
| --- | --- | --- | --- |
| S1-01 | `docs/agents/sprint1/s1-01-foundation/` (tasks.md, validate.md, plan.md) | 4 | 8 |
| S1-02 | `docs/agents/sprint1/s1-02-ci/` (tasks.md, validate.md, plan.md) | 3 | 22 |
| S1-03 | `docs/agents/sprint1/s1-03-installer/` (tasks.md, validate.md, plan.md) | 7 | 17 |
| S1-04 | `docs/agents/sprint1/s1-04-commitlint/` (tasks.md, validate.md, plan.md) | 3 | 14 |
| S1-05 | `docs/agents/sprint1/s1-05-community/` (tasks.md, validate.md, plan.md) | 7 | 20 |
| S1-06 | `docs/agents/sprint1/s1-06-release/` (tasks.md, validate.md, plan.md) | 4 | 34 |
| S1-07 | `docs/agents/sprint1/s1-07-docs-site/` (tasks.md, validate.md, plan.md) | 7 | 26 |
| S1-08 | `docs/agents/sprint1/s1-08-project-docs/` (tasks.md, validate.md, plan.md) | 8 | 36 |

Per-task test counts, as listed in each story's plan.md: S1-01 T2=3, T3=2, T4=3. S1-02 T1=14, T2=5, T3=3. S1-03 T1=2, T2=2, T3=4, T4=4, T5=1, T6=3, T7=1. S1-04 T1=5, T2=5, T3=4. S1-05 T1=2, T2=3, T3=4, T4=2, T5=4, T6=2, T7=3. S1-06 T1=7, T2=8, T3=7, T4=10, plus 2 listed ahead of the T1 heading. S1-07 T1=2, T2=2, T3=5, T4=6, T5=6, T6=3, T7=2. S1-08 T1=5, T2=6, T3=3, T4=7, T5=4, T6=3, T7=4, T8=4.

## Decisions requiring human approval

Execution may not start until a human has approved each item.

1. **README.md -> SPEC.md move** (S1-08 T1). The Rev 2 spec moves byte-for-byte with `git mv`, frozen at SHA256 `6bc35dadf698271b23790db43d2b521859d724081fae39dd1e281726e0fafe28` (64847 bytes). A new user-facing README.md replaces it. After the move, spec citations "README.md §N" read as "SPEC.md §N".
2. **Go pin `go1.27.1`** (S1-01 T1). The local toolchain is go1.22.2, so every local gate needs `GOTOOLCHAIN=auto` and a network download of the toolchain.
3. **Stdlib-only, text-based YAML checks.** The test packages `test/ci`, `test/commitlint`, `test/release`, `test/docs` and `test/community` use line and indentation matching, not a YAML library, so `go.mod` gains no dependency. The cost: these checks are less strict than a real parser.
4. **Asset naming.** Assets use the `_unverified` suffix on linux/arm, linux/mips and linux/mipsle, and names carry no version, so `releases/latest/download/<asset>` URLs resolve (S1-03 and S1-06 contract).
5. **Placeholder installer domain.** The install one-liner in README and install.sh points at a placeholder host until the first release.
6. **`rsnapshot` in the README backend ban list** (S1-08 T4 `TestReadmeMentionsOnlyResticBackend`, validate.md). SPEC §1 (thesis) and §3 (views) refer to rsnapshot-style snapshot naming. A human must decide whether a README reference to that naming style counts as naming a backend (keep the ban) or not (drop `rsnapshot` from the list).
7. **Docs versions chosen but not yet verified.** `mkdocs==1.6.1` and `mkdocs-material==9.6.14` (S1-07 T2) must be confirmed as current and compatible before GREEN.
8. **Missing local tools.** `goimports`, `golangci-lint` and `govulncheck` are not installed and must be installed before any RED or GREEN gate runs.
9. **Post-merge exit evidence is verified by the orchestrator, not by tests:** CI green on `master`, the Pages URL returns 200, the semantic-release `workflow_dispatch` dry run computes a version, and commitlint runs on the first PR.

## Cross-cutting gates

Copied verbatim from `docs/agents/sprint1/standards.md` § Cross-cutting gates. That file is the source of truth, and this copy must not diverge.

| Gate | Command | Threshold / notes | Local status | Source |
| --- | --- | --- | --- | --- |
| Format (gofmt) | `test -z "$(gofmt -l .)"` | must produce no output | INSTALLED | `~/.claude/rules/golang/coding-style.md` |
| Format (goimports) | `test -z "$(goimports -l .)"` | must produce no output | **NOT INSTALLED** | `~/.claude/rules/golang/coding-style.md` |
| Build / typecheck | `CGO_ENABLED=0 go build ./...` | compiler type-checks | INSTALLED (`go1.22.2`) | README.md §2, §5, §17 |
| Vet | `go vet ./...` | | INSTALLED | README.md §18 |
| Lint | `golangci-lint run` | bundles staticcheck | **NOT INSTALLED** | README.md §18 |
| Test + race | `go test -race ./...` | | INSTALLED | README.md §18; `~/.claude/rules/golang/testing.md` |
| Coverage | `go test -race -covermode=atomic -coverprofile=coverage.out ./...` + threshold | >= 80% total | INSTALLED | `~/.claude/rules/common/testing.md` |
| Vulnerability audit | `govulncheck ./...` | | **NOT INSTALLED** | README.md §18 |

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

Standards.md deliberately keeps these out of the per-task matrix. They run in CI (S1-02) or the release pipeline (S1-06): the 7-target cross-compile, static-linkage `file`/`ldd` checks, and `gosec` (recommended, non-blocking).
