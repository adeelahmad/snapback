---
type: validate
story: S2-06
---

# S2-06 validate — PASS/FAIL rubric

Run every command from the repo root `/Users/adeelahmad/work/snapback`. A row PASSES only when the literal expected output matches. Anything else is FAIL and cites the rule it violates (`docs/agents/sprint2/standards.md`, `docs/agents/sprint2/stories.md` S2-06, SPEC.md §5/§7/§20).

## Pre-flight

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| Upstream merged | `go list ./internal/compat/resticfx ./internal/mount/gofuse ./internal/projection` | three import paths, exit 0 | sprint2/plan.md Waves (wave 3 needs S2-03 and S2-04 merged green) |
| resticfx contract present | `go doc ./internal/compat/resticfx` | lists the version check, tree generator, init, backup, snapshots, mount/unmount, ls args and destroy | stories.md S2-03 "What's wanted"; open question 1 if names differ |
| Gate tools present | `command -v goimports golangci-lint govulncheck actionlint` | four paths, exit 0 | standards.md Cross-cutting gates |
| Integration prerequisites (macOS evidence run only) | `restic version; test -d /Library/Filesystems/macfuse.fs && echo macfuse` | `restic 0.19.0 …` then `macfuse` | standards.md Stage 1 integration-test gating; stories.md S2-03 (pinned 0.19.0) |
| Scope | `git diff --name-only master -- . ':!docs/agents'` | only paths under `internal/compat/fidelity/` (plus `docs/reports/stage1/fidelity-*.json` on the orchestrator evidence commit) | sprint2/plan.md ownership table; M-012 |
| No go.mod edit | `git diff --name-only master -- go.mod go.sum` | no output | sprint2/plan.md "go.mod/go.sum belong to S2-01 only" |

## T1 — Pure comparator and mtime-precision probe

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| Unit tests | `go test -race -run 'TestCompare\|TestMTimeToleranceIsZero\|TestPrecision' ./internal/compat/fidelity/` | `ok  	github.com/adeelahmad/snapback/internal/compat/fidelity` | stories.md S2-06 success scenarios (exact passes, 1 s drift fails, mode drift fails, precision recorded) |
| Tolerance is a named zero constant | `grep -nE 'MTimeTolerance +(time\.Duration +)?= +0' internal/compat/fidelity/compare.go` | one line | stories.md S2-06 constraint (named tolerance, never silently widened) |
| Package-scoped matrix | `GATE_RUN_MATRIX=0` plus `go vet ./internal/compat/fidelity/ && golangci-lint run ./internal/compat/fidelity/...` | no output, exit 0 | M-005; standards.md Vet, Lint |

## T2 — `restic ls --json` parser

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| Unit tests | `go test -race -run 'TestParseResticLs' ./internal/compat/fidelity/` | `ok  	github.com/adeelahmad/snapback/internal/compat/fidelity` | stories.md S2-06 (compare against `restic ls --json`) |
| No restic internals | `grep -rn 'restic/internal' internal/compat/fidelity/` | no output | SPEC.md §5 (CLI only); stories.md S2-03 constraint |

## T3 — Fixture spec and platform observation

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| Unit tests | `go test -race -run 'TestFixtures\|TestObserve' ./internal/compat/fidelity/` | `ok  	github.com/adeelahmad/snapback/internal/compat/fidelity` | stories.md S2-06 (distinct mtimes incl. sub-second and pre-2000, sizes 0/small/multi-MiB, modes 0644/0600/0755/read-only) |
| Cross-compile | `for t in linux/amd64 linux/arm64 linux/arm linux/mips linux/mipsle darwin/amd64 darwin/arm64; do GOOS=${t%/*} GOARCH=${t#*/} CGO_ENABLED=0 go build ./internal/compat/fidelity/ \|\| echo FAIL $t; done` | no output | `.github/workflows/ci.yml` cross-compile matrix; SPEC.md §17 |
| Lstat, never Stat, for observation | `grep -n 'os\.Stat(' internal/compat/fidelity/observe*.go` | no output | tasks.md T3 (symlink observed as itself) |

## T4 — Evidence JSON and prerequisite probe

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| Unit tests | `go test -race -run 'TestEvidence\|TestWriteEvidence\|TestMissingPrereq' ./internal/compat/fidelity/` | `ok  	github.com/adeelahmad/snapback/internal/compat/fidelity` | stories.md S2-06 (evidence round-trip); standards.md skip discipline |
| ctime/birth never asserted | `grep -rnE 'ctime_ok\|birth_time_ok\|CTimeOK\|BirthOK' internal/compat/fidelity/*.go` | no output | SPEC.md §5 ("documented, not claimed"); stories.md S2-06 failure scenario |

## T5 — Gated integration test and full matrix

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| Build tag | `head -1 internal/compat/fidelity/fidelity_integration_test.go` | `//go:build integration` | standards.md Stage 1 integration-test gating (build tag) |
| Default build skips cleanly | `go test -race ./internal/compat/fidelity/` | `ok …` and no integration test listed under `-v` | standards.md (integration excluded from default matrix) |
| Skip names prerequisite | `SNAPBACK_FUSE_TESTS= go test -tags=integration -run TestFidelityThroughSnapshotAlias -v ./internal/compat/fidelity/` | `--- SKIP` with `SNAPBACK_FUSE_TESTS not set: fidelity tests skipped` | standards.md skip discipline |
| Real run (macOS, local) | `SNAPBACK_FUSE_TESTS=1 SNAPBACK_EVIDENCE_DIR=docs/reports/stage1 go test -race -tags=integration -run TestFidelityThroughSnapshotAlias -v ./internal/compat/fidelity/` | `--- PASS`, and `docs/reports/stage1/fidelity-darwin.json` exists with `"pass": true`, `files_compared == files_generated >= 6`, `"resolved_inside_restic_mount": true` | SPEC.md §5 metadata fidelity; §7; §20; stories.md S2-06 success |
| No leaked mounts | `mount \| grep -c "$TMPDIR"` after the run | `0` | stories.md S2-03 failure scenario (leaked mount) |
| Full gate matrix (last GREEN) | every row of standards.md § Cross-cutting gates, including the coverage threshold `>= 80%` total | all PASS, no suppressions | M-005; standards.md Cross-cutting gates |

## Final sign-off

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| All tests present | `grep -c '^- \[ \]' docs/agents/sprint2/s2-06-fidelity/plan.md` | `26` at plan time, all ticked in `plan-ready.md` after RED | plan.md test count |
| Full matrix green | standards.md § Cross-cutting gates run on the story's final GREEN | every gate PASS; coverage >= 80% | standards.md; sprint2/stories.md Definition of Done |
| darwin evidence | `jq '.pass, .files_compared, .files_generated, .mtime_tolerance_ns, .mtime_precision_ns' docs/reports/stage1/fidelity-darwin.json` | `true`, N, N (N >= 6), `0`, a recorded value | SPEC.md §5; stories.md S2-06 |
| linux evidence | same `jq` on `docs/reports/stage1/fidelity-linux.json`, copied from the `stage1-evidence-linux` artifact of a green `fuse-linux` run on `master` | same shape; `pass` true, or false carried into S2-09 as "measured and failing" (never skipped) | sprint2/plan.md Linux evidence step; stories.md S2-06 |
| ctime/birth recorded, not claimed | `jq '[.files[] \| .ctime.claimed, .birth_time.claimed] \| unique' docs/reports/stage1/fidelity-*.json` | `[false]` for each file | SPEC.md §5 |
| Tolerance unchanged | `jq .mtime_tolerance_ns docs/reports/stage1/fidelity-*.json` | `0`, unless a human approved a non-zero named tolerance, recorded in DEVLOG.md | stories.md S2-06 constraint (no silent widening) |
| Structural review | `gate-structural-integrity` on `internal/compat/fidelity/` | no HIGH findings (duplicate small test helpers are expected per M-010) | M-009, M-010 |
