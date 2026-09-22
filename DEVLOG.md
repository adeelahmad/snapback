# Snapback Dev Log

## Working State
**Session:** 2 | **Date:** 2026-09-23

### Active Task
Sprint 5 (adoption): make the first install frictionless — a measured action
count, real packaging, debug logging, explicit file modes, a restore mount point
and a configurable web UI. Chain 5 waves w1-w3 are merged on `chain5/w3`; the
structural review of that chain is being worked off before the final gate.
- [x] Rulings 1-7 on the intake and the web UI plan (docs/agents/sprint5-adoption/rulings.md)
- [x] S5-09 action counts measured by acceptance tests: 3 on Linux, 4 on macOS
- [x] S5-36 logging, S5-37 file modes, S5-38 mount point, S5-30..S5-34 web UI
- [x] S5-35 pinned Dockerfile and ghcr.io images; S5-27 QEMU smoke for arm/mips
- [ ] Structural review fixes: S5-36/T16 wiring, CLEAN-1, CLEAN-2, CLEAN-3 <-- CURRENT
- [ ] Final gate, retrospective, then the Telemetry sprint

### Key Files (current shape)
**`cmd/snapback/daemondeps.go`** (MODIFIED)
Builds every daemon dependency: providers, reader policy, catalog, refresher,
mount linker. `daemonBuilder` still takes no `*slog.Logger`, which is why the
four S5-36 debug decorators are unwired; S5-36/T16 threads the logger in.

**`internal/logging/options.go`** (NEW, S5-36)
Parses level and format, builds the slog logger and the appending file sink that
`snapback run`, `snapback web` and the doctor bundle all share.

**`internal/daemon/mountpoint.go`** (NEW, S5-38)
Ensures the managed `.snapshot` link at a repository mount point once per
generation and removes it on teardown, keeping the directory (ruling 7).

**`internal/web/bindpolicy.go`** (NEW, S5-34)
Decides loopback versus remote binds and refuses a remote bind without
`--allow-remote`. It disagrees with `server.go`'s inline check on `"localhost"`
(CLEAN-1).

**`test/acceptance/actioncount_linux_test.go`** (NEW, S5-09)
Runs the counted zero-to-`ls .snapshot` script and compares the `# action` lines
with the count the script prints, so the README number cannot drift.

### Decisions (active)
- Bar for adoption: 3 actions on Linux, 4 on macOS, and a test must measure it
  before any doc may claim the number (ruling: intake, S5-09).
- `setup` never writes to the repository; on an empty repository it detects
  "no snapshots" and prints the next command.
- The release image bundles restic, pinned by version and SHA256, on a
  digest-pinned alpine base, published to `ghcr.io/adeelahmad/snapback`.
- `snapback web --with-daemon` is opt-in; `snapback web` alone starts no daemon.
- The mount point is one prompt or flag with one default (`/mnt/<instance>` on
  Linux), no picker and no repository browser.
- The backend view is `ids/` only: the mount point promises
  `<mount_point>/.snapshot/ids/<snapshot id>` and nothing else (ruling 6).

### Next Steps
1. S5-36/T16: thread the daemon logger into `daemonBuilder` and wrap the restic
   runner, catalog, reader policy and refresher with their `Log*` decorators.
2. CLEAN-1 loopback disagreement, CLEAN-2 dead `configControls`, CLEAN-3 setup
   template partials.
3. Re-run the structural review, then the sprint final gate and the retrospective.
4. Record and commit `docs-site/img/demo.gif`, then start the Telemetry sprint.

### Blockers
- None hard. The launch checklist's outward-facing steps (social preview,
  Discussions, first posts) need the human.

### Watch Out
- commitlint rejects `build:` subjects and checks the PR title as well.
- `snapback setup` refuses a project under `/tmp`, so acceptance projects must
  live elsewhere.
- Workers have a 5-minute cap; a task that touches several templates has to be
  split before it is dispatched.

---
---

## Session Archive

### Session 2 -- 2026-09-23: Sprint 5 (adoption)
**What we did:** Planned Sprint 5 and ran chain 5 (waves w1-w3, ~91 commits on top of
`stage-5`): measured the action count (3 Linux, 4 macOS), debug logging end to end
(S5-36), explicit file and directory modes (S5-37), a repository mount point for
restore onto another machine (S5-38), the sectioned web UI with setup, instances,
daemon control and bind policy (S5-30..S5-34), a pinned restic-bundling image on
ghcr.io (S5-35), and QEMU smoke tests that retired the "unverified" arm/mips label
(S5-27). Sprints 2-4 were not journaled here; this entry starts again at Sprint 5.
**Files:** internal/logging, internal/fsmode wiring in internal/daemon, internal/web,
internal/webui/templates, cmd/snapback/daemondeps.go, Dockerfile, .goreleaser.yaml,
test/acceptance, docs-site, docs/reports/sprint5/actions.md.
**Decisions:** rulings 1-7 (docs/agents/sprint5-adoption/rulings.md) — pinned image with
bundled restic, opt-in `--with-daemon`, refuse remote binds, one mount point question,
`ids/`-only backend view, teardown keeps the directory.

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

### 2026-09-23 - commitlint rejected `build:` subjects and the PR title
**What happened:** Commits and a pull request using a `build:` type failed the
commit lint job twice.
**Root cause:** `.commitlintrc.json` allows only feat, fix, docs, test, ci, chore,
refactor and perf, and the workflow lints the PR title as well as the commits.
**How we fixed it:** Reworded the subjects to `chore:` and renamed the PR title.
**Lesson:** The allowed type list is the whole list; it binds PR titles too.

### 2026-09-23 - The Linux action-count test put its project under /tmp
**What happened:** `TestActionCountLinux` failed because `snapback setup` refused
the directory it was pointed at.
**Root cause:** The script created the counted project inside the test's temp dir,
and setup declines to manage paths under `/tmp`.
**How we fixed it:** ACC-LINUX-1 moved the counted project outside the temp dir.
**Lesson:** Acceptance harnesses must respect the product's own refusals rather
than work around them.

### 2026-09-23 - S5-30/T7 exceeded the 5-minute worker cap
**What happened:** The GREEN worker for the every-section config save ran past the
5-minute limit and was killed mid-task.
**Root cause:** One task covered the form model, the control partials and the save
path across several templates — far more than a single worker can finish.
**How we fixed it:** A planner split it into T7a (retire the flat form tests and fix
the fixture), T7b (load the partials, section the config view) and T7c (save every
section with field errors); all three landed inside the cap.
**Lesson:** Size a task by what fits in the cap, and split before dispatching, not
after a kill.

### 2026-09-23 - The setup form had no typed-password control
**What happened:** The web setup form could not accept a repository password typed
by the user; only a pre-existing password file worked.
**Root cause:** The form model mirrored config keys only, and no test asked for the
end-to-end path from a typed password to a started daemon.
**How we fixed it:** The S5-30/T10 acceptance test pinned a typed-password setup
starting the daemon, and the form now writes a 0600 credential file.
**Lesson:** A pinned end-to-end path finds the missing control that a field-by-field
test never asks for.

### 2026-09-23 - The S5-38 design promised mount point views that do not exist
**What happened:** The mount point design text said a user would find
`ids/`, `hosts/`, `snapshots/` and `tags/` under the mount point.
**Root cause:** It described Restic's mount layout, but our backend is mounted with
`--path-template ids/%I`, pinned in `internal/provider/restic/args_test.go`, so only
`ids/` is ever published.
**How we fixed it:** Ruling 6 fixed the promise at `ids/` only, and the T9 acceptance
test and the docs were corrected to match.
**Lesson:** A design may only promise what the pinned arguments actually produce.

## Technical Debt & Future Ideas
- `demo.gif` is not recorded or committed yet: the tape lives at the quick-start
  path and the docs workflow's demo job uploads an artifact, but
  `docs-site/img/demo.gif` is still missing from the tree.
- The web UI screenshots in the docs are stale after the sectioned form landed
  (S5-30/T9b).
- `USAGE-DASH`: single-dash help (`-help`) renders differently from `--help`.
- `NEXT-1`: the next-step line after `setup` is chosen per platform and service and
  needs one more pass for the no-service path.
- Log-record test helpers (`decodeRecords` and the JSON logger builder) are
  reimplemented in three test packages; one shared internal test package would do.
- SPEC wording: "unverified" for arm and mips was replaced with QEMU smoke coverage
  once S5-27/T3 landed; keep the wording in step with what CI really runs.
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
