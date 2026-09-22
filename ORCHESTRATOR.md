# Snapback — Orchestrator Ledger

Spec: `README.md` (Revision 2). Workflow: `agentic-agile` plugin. Agent artifacts: `docs/agents/`.

## Current state

- **Tick:** 3
- **Stage:** 0 — Scaffolding (Sprint 1)
- **Phase:** PLANNING — intake + standards passed; Stage-1 passed (8 stories, 2 waves); PLANNING COMPLETE — gate-stage2-complete exit 0, gate-tooling exit 0. 8 stories / 43 tasks / 177 tests. LOOP STOPPED (cron 72a30c79 deleted) at HUMAN APPROVAL GATE
- **Last gate:** n/a (no code yet)
- **Human gate pending:** YES — approve docs/agents/sprint1/plan.md + 9 decisions (section 'Decisions requiring human approval') before any RED dispatch

## Stage table (README §22)

| Stage | Scope | Exit evidence | Status | Evidence recorded |
| --- | --- | --- | --- | --- |
| 0. Scaffolding | Repository furniture, CI, lint, race tests, semantic release, docs site pipeline, installer script skeleton | Green CI on an empty binary; docs site deploys | planning | — |
| 1. Compatibility milestone | Pin deps; disposable Restic repo; verify `--path-template ids/%I`; tiny FUSE catalog Linux+macOS; metadata fidelity; rclone/GDrive latency; crawler test | Numbers recorded in the report; go/no-go on latency | not started | — |
| 2. Core vertical slice | Config, SnapshotProvider + Restic, resolver, private Restic mount, virtual catalog, `link`/`open`/`snap`, ownership registry | Acceptance 3–8, 10 on Linux | not started | — |
| 3. Reliable background operation | Daemon, refresh, pre-warm, IPC, shell hooks, seeding + watcher + inode budget, reader policy, crash recovery, shutdown | Acceptance 1, 2, 9, 11–15 | not started | — |
| 4. Web UI and services | Setup, Configuration, History, Status, Integrations; launchd/systemd/OpenRC; `install service`; packages + installer | Acceptance 16, 17; all channels publish from a tag | not started | — |
| 5. macOS proof | go-fuse on macFUSE/Apple Silicon, LaunchAgent, Finder companion | Acceptance 17 on macOS, 19 | not started | — |
| 6. On-access mode | fanotify, Endpoint Security, process policy, fail-open | Acceptance 18; promotion decision | not started | — |
| 7. Release | Binaries, companion pkg, checksums, signatures, notices, docs, launch assets | Requirement matrix complete | not started | — |

## Environment notes (tick 0)

- Tools present: go 1.22.2 (old), restic, rclone, macFUSE, md-db, ctx-symbols.
- Tools missing: golangci-lint, staticcheck, govulncheck — gate step `golangci-lint run` cannot run locally until installed.
- Tick timer: session cron job `72a30c79` (`*/5 * * * *`) fires each tick; foreground `sleep` is blocked in this harness. Workers are background agents; kills use TaskStop.
- Concurrency deviation: planning has only 2 independent tasks (intake, standards); planner depends on both. The 5-worker minimum cannot be met in planning without inventing roles, which the ground-truth rules forbid.

## Task log

| Tick | Task | Role | Outcome | Split applied |
| --- | --- | --- | --- | --- |
| 0 | sprint1/intake | intake | spawned 00:02Z | — |
| 0 | sprint1/standards | standards | spawned 00:02Z | — |
| 0 | sprint1/intake | intake | passed (gate-intake re-run by orchestrator: PASS; 0 blocking questions; origin=adeelahmad/snapback exists) | — |
| 1 | sprint1/standards | standards | killed at 5m30s (5-min rule); artifact standards.md already written and gate-standards-cited re-run by orchestrator: PASS, md-db ok → accepted as passed | none (output complete) |
| 1 | sprint1/planner-stage1 | planner | spawned 00:08Z (scope narrowed to stories.md + sprint plan.md to fit 5-min budget) | pre-split: Stage-2 per story to separate planner workers |
| 1 | sprint1/planner-stage1 | planner | passed (md-db 0 errors re-run by orchestrator; 8 stories, non-overlapping file sets) | — |
| 1 | sprint1/stage2/S1-01 | planner | spawned 00:11Z | — |
| 1 | sprint1/stage2/S1-02 | planner | spawned 00:11Z | — |
| 1 | sprint1/stage2/S1-03 | planner | spawned 00:11Z | — |
| 1 | sprint1/stage2/S1-04 | planner | spawned 00:11Z | — |
| 1 | sprint1/stage2/S1-05 | planner | spawned 00:11Z | — |
| 2 | sprint1/stage2/S1-01 | planner | passed (gate-plan-shape on plan.md re-run by orchestrator: exit 0; 8 tests, 4 tasks; pins go1.27.1) | — |
| 2 | sprint1/stage2/S1-02 | planner | passed (gate-plan-shape on plan.md re-run by orchestrator: exit 0; 22 tests, 3 tasks) | — |
| 2 | sprint1/stage2/S1-03 | planner | passed (gate-plan-shape on plan.md re-run by orchestrator: exit 0; 17 tests, 7 tasks; asset naming contract relayed to S1-06) | — |
| 2 | sprint1/stage2/S1-04 | planner | passed (gate-plan-shape on plan.md re-run by orchestrator: exit 0; 14 tests, 3 tasks) | — |
| 2 | sprint1/stage2/S1-05 | planner | passed (gate-plan-shape on plan.md re-run by orchestrator: exit 0; 20 tests, 7 tasks) | — |
| 2 | sprint1/stage2/S1-06 | planner | spawned 00:13Z | — |
| 2 | sprint1/stage2/S1-07 | planner | spawned 00:13Z | — |
| 2 | sprint1/stage2/S1-08 | planner | spawned 00:13Z | — |
| 2 | sprint1/stage2/S1-07 | planner | passed gate (26 tests, 7 tasks) but plan defect found by orchestrator: bans 'rclone' as a non-Restic backend, contradicting README §11/§12/§17 → sent back for fix (attempt 2) | none — targeted fix |
| 2 | sprint1/stage2/S1-07 | planner | passed attempt 2 (rclone removed from ban list; gate-plan-shape exit 0, re-verified by orchestrator) | — |
| 2 | sprint1/stage2/S1-06 | planner | passed (gate-plan-shape exit 0 re-verified; 34 tests, 4 tasks; S1-01 ldflags + S1-03 naming contracts asserted by tests) | — |
| 2 | sprint1/stage2/S1-08 | planner | passed (gate-plan-shape exit 0 re-verified; 36 tests, 8 tasks; README sha256 6bc35dad… verified) | — |
| 2 | sprint1/stage2/finalize | planner | spawned 00:16Z | — |
| 3 | sprint1/stage2/finalize | planner | passed after orchestrator fix: gate-stage2-complete first blocked on stray transcript copy in docs/agents/sprint1/.agentic (hook wrote relative to shell cwd); moved to .agentic/transcripts-misplaced-sprint1-cwd; re-run exit 0 | — |

## Plugin issues found

- `gate-plan-shape` given a directory exits 0 (grep 'Is a directory') — false pass. Orchestrator always passes the plan.md file path. Report upstream to agentic-agile.
- `md-db validate` prints `files: []` even when files are valid; S1-05 planner confirmed via a negative test that it does read them.
- `transcripts` hook writes `.agentic/transcripts` relative to the shell cwd; a cwd inside a sprint dir planted a transcript containing 'TBW' that made gate-stage2-complete fail. Keep the shell at repo root.
- `md-db validate` returns `files: []` even for a single file path; S1-05 planner's negative test showed it does detect errors, but its file selection should be checked upstream.
