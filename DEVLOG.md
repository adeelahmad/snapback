# Snapback Dev Log

## Working State
**Session:** 1 | **Date:** 2026-09-22

### Active Task
Sprint 1 / Stage 0: scaffolding (repository furniture, CI, lint, race tests,
semantic release, docs site pipeline, installer skeleton). Exit evidence is green
CI on an empty binary and a docs site that deploys. No product feature exists yet.
- [x] Planning: intake, standards, 8 stories (S1-01..S1-08), human approval
- [x] S1-01 foundation: go.mod, `internal/version`, `cmd/snapback` stub (merged)
- [x] S1-02 CI workflow, S1-04 commitlint, S1-06 release config (merged)
- [x] S1-03 installer skeleton (merged)
- [ ] S1-05 community files, S1-07 docs site, S1-08 project docs <-- CURRENT
- [ ] Structural reviews, final gate, retrospective

### Key Files (current shape)
**`SPEC.md`** (MOVED from README.md, Revision 2)
The full product spec, byte-identical to the old README (sha256 verified). Every
design decision and the stage table (§22) live here.

**`README.md`** (NEW)
User-facing entry point. States plainly that Snapback is pre-release, has no
tagged release, and that the restore workflow is a design goal; links SPEC.md.

**`.github/workflows/ci.yml`** (NEW)
Go race tests, vet, gofmt/goimports, golangci-lint, govulncheck, coverage.
Tool installs are pinned (goimports v0.50.0, govulncheck v1.8.0), never `@latest`.

**`install.sh`** (NEW)
POSIX installer skeleton: OS/arch detection, download plus sha256 verification,
`~/.local/bin` fallback. FUSE hints are printed as instructions, never executed.

**`test/projectdocs/`** (NEW)
Go tests that pin the shape and honesty of the project docs (README, SPEC,
ARCHITECTURE, CLAUDE.md, DEVLOG.md, NOTICE).

### Decisions (active)
- Spec moved from README.md to SPEC.md: the README must speak to users about what
  works today; the spec describes a future product and would overclaim if it
  stayed as the landing page.
- Strict TDD per task (RED, SCAFFOLD, GREEN) with one worktree per worker; the
  orchestrator re-runs `go test` itself because plugin gates skip Go.
- Intermediate GREENs run with GATE_RUN_MATRIX=0; the full matrix is enforced at
  each story's merge into `stage-0`.
- Human enabled auto-approve after reviewing the plan (including the rsnapshot ban).

### Next Steps
1. Finish S1-08 (T3-T8), S1-05 (T5-T6), S1-07 (T5) and merge into `stage-0`.
2. Run structural review per story, then the sprint final gate.
3. Run the retrospective; report plugin gate issues upstream.
4. Stage 1: compatibility milestone (pin deps, Restic path-template, FUSE catalog).

### Blockers
- None hard. Disk hit 100% at tick 10 (Go caches 3.1 GB); finished worktrees
  were removed. The human should free disk before Stage 1.

### Watch Out
- Keep the shell at the repo root: the transcripts hook writes relative to cwd.
- `STANDARDS_FILE` must be absolute in task.env, else gates fall back to cargo.
- gate-structural-integrity flags Go `_test.go` local duplicates as HIGH (false positive).

---
---

## Session Archive

### Session 1 -- 2026-09-22: Sprint 1 planning and Stage 0 scaffolding
**What we did:** Planned Sprint 1 (8 stories) and ran two-phase TDD across parallel
worktree workers; S1-01/02/03/04/06 merged into `stage-0` with the full gate green
(coverage 87.5%, govulncheck clean, actionlint clean).
**Files:** go.mod, cmd/snapback, internal/version, .github/workflows/*, install.sh,
.goreleaser.yaml, .releaserc.json, .commitlintrc.json, SPEC.md, README.md.
**Decisions:** Spec moved to SPEC.md; orchestrator compensates for Rust-only gates.

---

## Milestones
- [ ] Stage 0 - Scaffolding: green CI on an empty binary; docs site deploys
- [ ] Stage 1 - Compatibility milestone
- [ ] Stage 2 - Core vertical slice
- [ ] Stage 3 - Reliable background operation
- [ ] Stage 4 - Web UI and services
- [ ] Stage 5 - macOS proof
- [ ] Stage 6 - On-access mode
- [ ] Stage 7 - Release

## Mistakes & Lessons
### 2026-09-22 - Unpinned tools in CI
**What happened:** ci.yml installed goimports and govulncheck with `@latest`.
**Root cause:** Green-worker took the shortest install line; no test pinned versions.
**How we fixed it:** Structural review flagged it; fix task S1-02/fix1 pinned
goimports v0.50.0 and govulncheck v1.8.0.
**Lesson:** Pin every action and tool to a released version that really exists.

### 2026-09-22 - Plugin gates are Rust-only
**What happened:** gate-red-verify and gate-green-verify WARN and skip for Go;
gate-plan-shape exits 0 when given a directory.
**Root cause:** Dynamic checks hardcode `cargo test`; directory input is unguarded.
**How we fixed it:** Orchestrator re-runs `go test` per worktree and always passes
the plan.md file path. Not fixed in the plugin.
**Lesson:** A green gate is only evidence when it actually ran the tests.

### 2026-09-22 - Stray transcript broke a planning gate
**What happened:** gate-stage2-complete failed on a transcript containing 'TBW'.
**Root cause:** The transcripts hook wrote `.agentic/transcripts` under a sprint dir
because the shell cwd was there.
**How we fixed it:** Moved the file aside and re-ran the gate; shell stays at root.
**Lesson:** Hooks that write relative paths depend on cwd.

### 2026-09-22 - T1 of S1-01 could not pass the matrix alone
**What happened:** `go vet ./...` fails on a module with no packages.
**Root cause:** Plan split go.mod into a task with no code.
**How we fixed it:** Kept T1 as the chain base and merged it with T2.
**Lesson:** Every task must leave the repo buildable on its own.

## Technical Debt & Future Ideas
- Test-helper consolidation: `repoRoot`, `section`, `indentOf` and similar helpers
  are duplicated across Go test packages (test/projectdocs, S1-05, S1-07). Go test
  packages cannot share unexported code; a small shared internal test package
  would remove the copies.
- S1-05 `securitySection` is a near-duplicate of `contributingSection`.
- Report upstream to agentic-agile: Rust-only dynamic gates, gate-plan-shape
  directory false pass, cwd-relative transcripts hook, `md-db validate` printing
  `files: []`, missing `_test.go` carve-out in gate-structural-integrity, and
  whole-repo matrix per task conflicting with RED-first stories.
- Worktree agents cannot write the main-tree STORY_DIR; output.md is relayed by
  the orchestrator.
- goreleaser and commitlint are not installed locally, so their configs are not
  tool-validated; CI must validate them.
- Installer fallback to `~/.local/bin` triggers when `/usr/local/bin` is unwritable
  or not on PATH; confirm with the human.
