---
type: sprint-plan
sprint: 2
stage: "2"
---

# Sprint 2 plan — Stage 2: Compatibility milestone (task-level plan)

Stories and intents: `docs/agents/sprint2/stories.md`. Gate rules: `docs/agents/sprint2/standards.md`. Per-story tasks, tests and rubrics live in the story dirs listed under "Per-story plan pointers". Totals: 45 tasks, 210 tests.

## Sprint goal

Pin go-fuse behind a `mount` adapter seam, with a pure `projection` model. Verify `restic mount --path-template ids/%I` against restic 0.19.0 on a disposable repo. Prove a tiny directory/symlink catalog mounts, browses and unmounts on macOS (local macFUSE) and Linux (new `fuse-linux` CI job). Check mtime, size and mode fidelity through `.snapshot/<timestamp>/`. Measure the four §11 latencies over `gdrive:snapback-stage1`, then delete that repo. Count crawler catalog hits. The sprint ends with every number in `docs/reports/stage1-measurements.md` (backed by JSON in `docs/reports/stage1/`) and a latency go/no-go item open for the human. No threshold is set by the plan.

## Waves

Waves are story-level; ticks are task-level. A task starts only when the tasks and stories it depends on are merged green. Tasks in the same tick own disjoint files (task file lists are in each story's `tasks.md`). Within a story, tasks that share a file run in separate ticks, in the story's task order.

| Wave | Stories | Starts when |
| --- | --- | --- |
| 1 | S2-01, S2-02, S2-03, S2-05 (T1, T2) | immediately |
| 2 | S2-04, S2-08 | S2-04: S2-01 T3 and S2-02 T6 merged green. S2-08: S2-03 T7 merged green |
| 3 | S2-06, S2-07 | S2-06: S2-03 and S2-04 merged green. S2-07: S2-04 merged green |
| 4 | S2-05 T3, then S2-09 | S2-05 T3: S2-03, S2-04, S2-06, S2-07 on `master`. S2-09: all evidence JSON committed (darwin files run locally; linux files from a green `fuse-linux` artifact on `master` after S2-05 T3; `latency.json` from the human-confirmed gdrive run) |
| — | Human go/no-go on latency | S2-09 merged |

### Tick table (concurrent tasks per tick)

| Tick | Wave | Concurrent tasks | Files touched (disjoint per tick) |
| --- | --- | --- | --- |
| 1 | 1 | S2-01 T1, S2-02 T1, S2-02 T2, S2-03 T1, S2-03 T2, S2-03 T3, S2-03 T4, S2-05 T1 | `go.mod`, `go.sum`, `internal/mount/mount_test.go`, `internal/mount/gofuse/attr{,_test}.go`; `internal/projection/{doc.go,deps_test.go}`; `internal/projection/name{,_test}.go`; `resticfx/{version,guard}{,_test}.go`; `resticfx/args{,_test}.go`; `resticfx/snapshots{,_test}.go`; `resticfx/{tree,password}{,_test}.go`; `.github/workflows/ci.yml`, `test/ci/fuse_job_test.go` |
| 2 | 1 | S2-01 T2, S2-02 T3, S2-03 T5, S2-03 T6, S2-05 T2 | `internal/mount/mount{,_test}.go`; `internal/projection/{spec,build,build_test}.go`; `resticfx/{runner,fixture}{,_test}.go`; `resticfx/{evidence,prereq}{,_test}.go`; `ci.yml`, `fuse_job_test.go` |
| 3 | 1 | S2-01 T3, S2-02 T4, S2-03 T7 | `internal/mount/gofuse/attr{,_test}.go`, `go.sum`; `internal/projection/{catalog.go,fixture_test.go,lookup_test.go}`; `resticfx/mount{,_test}.go`, `resticfx/pathtemplate_integration_test.go` |
| 4 | 1-2 | S2-02 T5, S2-08 T1, S2-08 T2, S2-08 T3 | `internal/projection/{catalog.go,readdir_test.go}`; `latency/guard{,_test}.go`; `latency/result{,_test}.go`; `latency/scratch{,_test}.go` |
| 5 | 1-2 | S2-02 T6, S2-08 T4 | `internal/projection/{catalog.go,readlink_test.go,contract_test.go}`; `latency/cleanup{,_test}.go` |
| 6 | 2 | S2-04 T1, S2-08 T5 | `internal/mount/gofuse/fs{,_test}.go`; `latency/run{,_test}.go` |
| 7 | 2 | S2-04 T2, S2-04 T3, S2-08 T6 | `gofuse/fs{,_test}.go`; `gofuse/adapter{,_test}.go`; `tools/stage1latency/{main,deps,run,run_test}.go` |
| 8 | 2 | S2-04 T4 | `gofuse/mount_integration_test.go` |
| 9 | 3 | S2-06 T1, S2-06 T2, S2-06 T3, S2-06 T4, S2-07 T1, S2-07 T2, S2-07 T3, S2-07 T4 | `fidelity/compare*`; `fidelity/resticls*`; `fidelity/{fixtures,observe}*`; `fidelity/{evidence,prereq}*`; `crawler/counter*`; `crawler/tools*`; `crawler/evidence*`; `crawler/seed*` |
| 10 | 3 | S2-06 T5, S2-07 T5 | `fidelity/fidelity_integration_test.go`; `crawler/crawler_integration_test.go` |
| 11 | 4 | S2-05 T3 | `.github/workflows/ci.yml`, `test/ci/fuse_job_test.go` |
| 12 | 4 | S2-09 T1 | `test/reports/{helpers,inventory}_test.go`, `docs/reports/stage1-measurements.md` |
| 13 | 4 | S2-09 T2 | `test/reports/versions_test.go`, report |
| 14 | 4 | S2-09 T3 | `test/reports/platform_test.go`, report |
| 15 | 4 | S2-09 T4 | `test/reports/latency_test.go`, report |
| 16 | 4 | S2-09 T5 | `test/reports/crawler_test.go`, report |
| 17 | 4 | S2-09 T6 | `test/reports/matrix_test.go`, report |

Paths abbreviated `resticfx/`, `latency/`, `fidelity/`, `crawler/` are under `internal/compat/`; `gofuse/` is `internal/mount/gofuse/`. Ticks 1, 2 and 9 meet the 5-worker minimum; the others are narrow because of hard dependencies or a shared file (S2-01 `mount_test.go`/`attr.go`, S2-02 `catalog.go`, S2-04 `fs.go`, S2-09's single report file).

Evidence runs are orchestrator steps, not worker tasks, because they need the real host and network:
- **macOS evidence (local):** after each of S2-03 T7 (tick 3), S2-04 T4 (tick 8), S2-06 T5 and S2-07 T5 (tick 10) merges, run `SNAPBACK_FUSE_TESTS=1 SNAPBACK_EVIDENCE_DIR=docs/reports/stage1 go test -race -tags=integration ./internal/...` on this Mac. Commit the `*-darwin.json` file to its owning story's path.
- **Linux evidence (CI):** after S2-05 T3 (tick 11) is on `master`, wait for the green `fuse-linux` run. Download `stage1-evidence-linux` and commit the `*-linux.json` files to their owning paths. This gates tick 12.
- **Latency (network, human-confirmed remote):** after S2-08 T6 merges (tick 7) and only after the human explicitly confirms the run, run `SNAPBACK_RCLONE_REMOTE=gdrive:snapback-stage1 go run ./tools/stage1latency -out docs/reports/stage1/latency.json` once. Confirm `remote_deleted: true`, and that `rclone lsf gdrive:snapback-stage1` reports the path absent. S2-08 has no gated network integration test.

## Parallelism and cross-story file ownership

| Story | Wave | Owned files (exclusive) | Hard deps |
| --- | --- | --- | --- |
| S2-01 | 1 | `go.mod`, `go.sum`, `internal/mount/mount.go`, `internal/mount/mount_test.go`, `internal/mount/gofuse/attr.go`, `internal/mount/gofuse/attr_test.go` | none |
| S2-02 | 1 | `internal/projection/**` | none |
| S2-03 | 1 | `internal/compat/resticfx/**`, `docs/reports/stage1/pathtemplate-darwin.json`, `docs/reports/stage1/pathtemplate-linux.json` | none |
| S2-05 | 1 (T1, T2), 4 (T3) | `.github/workflows/ci.yml`, `test/ci/fuse_job_test.go` | none for T1/T2; T3 needs S2-03, S2-04, S2-06, S2-07 merged (soft: runs the others' integration tests) |
| S2-04 | 2 | `internal/mount/gofuse/fs.go`, `internal/mount/gofuse/fs_test.go`, `internal/mount/gofuse/adapter.go`, `internal/mount/gofuse/adapter_test.go`, `internal/mount/gofuse/mount_integration_test.go`, `docs/reports/stage1/catalog-darwin.json`, `docs/reports/stage1/catalog-linux.json` | S2-01, S2-02 |
| S2-08 | 2 | `internal/compat/latency/**`, `tools/stage1latency/**`, `docs/reports/stage1/latency.json` | S2-03 |
| S2-06 | 3 | `internal/compat/fidelity/**`, `docs/reports/stage1/fidelity-darwin.json`, `docs/reports/stage1/fidelity-linux.json` | S2-03, S2-04 |
| S2-07 | 3 | `internal/compat/crawler/**`, `docs/reports/stage1/crawler-darwin.json`, `docs/reports/stage1/crawler-linux.json` | S2-04 |
| S2-09 | 4 | `docs/reports/stage1-measurements.md`, `test/reports/**` | S2-03, S2-04, S2-06, S2-07, S2-08 |

Ownership rules the orchestrator enforces at dispatch:
- `go.mod`/`go.sum` belong to S2-01 only. If a later story needs a new requirement, that is an S2-01 fix task.
- `internal/mount/gofuse/` and `docs/reports/stage1/` are split by exact file. No glob in one row covers a file in another.
- `test/ci/ci_test.go`, `matrix_test.go`, `lint_test.go` and `helpers_test.go` (from S1-02) are read-only this sprint. S2-05 adds `fuse_job_test.go` only. If an existing assertion conflicts with the new job, that goes to the orchestrator, not to an edit (M-012).
- `cmd/snapback/**`, `internal/version/**`, `README.md`, `SPEC.md`, `CLAUDE.md`, `DEVLOG.md`, `docs-site/**` and `mkdocs.yml` are untouched by every story.

Maximum story fan-out: 4 (wave 1). Maximum task fan-out: 8 (ticks 1 and 9).

## Critical path

S2-02 T1..T6 (ticks 1-5) -> S2-04 T1, T3, T4 (ticks 6-8) -> S2-06 T1..T5 (ticks 9-10) -> S2-05 T3 (tick 11) -> Linux evidence (green `fuse-linux` artifact on `master`) -> S2-09 T1..T6 (ticks 12-17) -> human go/no-go. 17 ticks plus two orchestrator gates.

- S2-04 carries the most risk: the first real go-fuse mount on macFUSE (Apple Silicon). If go-fuse fails there, §5 requires a documented stop and a human decision on a substitute adapter (see decisions below). That would stall S2-06 and S2-07.
- S2-06 is the longest chain because it combines a restic mount, a catalog mount and per-file comparison. S2-07 runs in parallel and is shorter (no restic).
- S2-08 is off the critical path but gated by the network and a real credential. It can run as soon as S2-03 lands. The human-approved remote, and deleting it afterwards, are mandatory.
- The Linux evidence depends on the S2-05 job being merged and green on `master` after S2-03, S2-04, S2-06 and S2-07. That is an orchestrator wait, not a task.

## Per-story plan pointers

| Story | Stage-2 dir | Tasks | Tests | Files |
| --- | --- | --- | --- | --- |
| S2-01 | `docs/agents/sprint2/s2-01-mount-adapter/` | 3 | 15 | `tasks.md`, `validate.md`, `plan.md` |
| S2-02 | `docs/agents/sprint2/s2-02-projection/` | 6 | 18 | `tasks.md`, `validate.md`, `plan.md` |
| S2-03 | `docs/agents/sprint2/s2-03-resticfx/` | 7 | 32 | `tasks.md`, `validate.md`, `plan.md` |
| S2-04 | `docs/agents/sprint2/s2-04-gofuse-catalog/` | 4 | 21 | `tasks.md`, `validate.md`, `plan.md` |
| S2-05 | `docs/agents/sprint2/s2-05-fuse-ci/` | 3 | 17 | `tasks.md`, `validate.md`, `plan.md` |
| S2-06 | `docs/agents/sprint2/s2-06-fidelity/` | 5 | 26 | `tasks.md`, `validate.md`, `plan.md` |
| S2-07 | `docs/agents/sprint2/s2-07-crawler/` | 5 | 26 | `tasks.md`, `validate.md`, `plan.md` |
| S2-08 | `docs/agents/sprint2/s2-08-latency/` | 6 | 34 | `tasks.md`, `validate.md`, `plan.md` |
| S2-09 | `docs/agents/sprint2/s2-09-report/` | 6 | 21 | `tasks.md`, `validate.md`, `plan.md` |
| **Total** | | **45** | **210** | |

Stage-2 sequencing notes carried from memory:
- M-005: intermediate GREENs run with `GATE_RUN_MATRIX=0` plus a package-scoped matrix. Only each story's last GREEN runs the full matrix, including the 80% total coverage. Integration-only code lives in `_test.go` files with the `integration` build tag, so default-build coverage is not diluted.
- M-007: S2-05's workflow is validated with `actionlint` locally and then a real CI run. Restic's `--path-template` was checked against the installed binary during planning (`restic mount --help` on 0.19.0 lists `--path-template`). S2-03 still proves it at runtime.
- M-010: each new test package will duplicate small helpers (`repoRoot`, evidence readers). That is expected and is not a structural-review failure.

## Decisions requiring human action

These need the human. No agent makes them, and nothing proceeds past them silently.

1. **go-fuse -> cgofuse fallback (only if macFUSE fails).** If S2-04 T4 cannot pass on macFUSE, the worker stops and reports. §5 allows a documented cgofuse substitute behind the same `mount.Adapter` interface, at the cost of cgo on macOS. The human decides before any substitution.
2. **Latency go/no-go (after the numbers).** After S2-09 merges, the human reads the report and chooses: proceed to Stage 2, deeper pre-warm, or pull the §23 local hot-cache repository forward. The plan sets no threshold, and the report records only `Latency go/no-go: PENDING — human decision`.
3. **Confirm each real Google Drive run.** Every run of `go run ./tools/stage1latency` against `gdrive:snapback-stage1` needs explicit human confirmation beforehand, including any re-run. There is no automated or CI path to it.
4. **Push to GitHub.** Pushing to `master` (which triggers the `fuse-linux` run that produces Linux evidence) needs human approval.

Settled at planning (recorded, no further action unless the human objects):
- go-fuse pin `github.com/hanwen/go-fuse/v2 v2.11.0`.
- `SNAPBACK_EVIDENCE_DIR` controls output only and gates no test.
- `SNAPBACK_RCLONE_REMOTE` must equal `gdrive:snapback-stage1`; any other value is refused.
- `fd` is not installed on this Mac, so the darwin `fd`/`fd -L` rows will be `not-tested-here` unless the human installs it before the S2-07 evidence run. Linux CI installs `fd-find`.
- Restic on Linux CI is the pinned 0.19.0 release asset with a SHA-256 check (literal in S2-05 validate.md).
- S2-05 uploads evidence with `if-no-files-found: warn` in T2; S2-05 T3 (tick 11) flips it to `error` once S2-03/04/06/07 are merged.
- Post-merge exit evidence the orchestrator verifies (not unit tests): a green `fuse-linux` run on `master` with its artifact, the local macOS integration run, the confirmed gdrive latency run with the remote deleted, and every evidence JSON committed.

## Cross-cutting gates

Copied verbatim from `docs/agents/sprint2/standards.md` § Cross-cutting gates. That file is the source of truth, and this copy must not diverge. The integration suite (`go test -race -tags=integration ./...` with the env gating in standards.md § "Sprint 2 / Stage 1 integration-test gating") is an additional evidence step, not part of this per-task matrix.

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
