---
type: tasks
story: S1-08
---
# S1-08 — Project docs: tasks

Owned set (stories.md S1-08): `README.md`, `SPEC.md`, `ARCHITECTURE.md`, `NOTICE`,
`CLAUDE.md`, `DEVLOG.md`, `test/projectdocs/**`. Depends on S1-01 (`go.mod`, module
`github.com/adeelahmad/snapback`). Tests are Go standard library only (package
`projectdocs`); no new module dependencies.

Frozen spec hash (computed 2026-09-22 from both the working-tree `README.md` and
`git show HEAD:README.md`, identical):
`SHA256(README.md) = 6bc35dadf698271b23790db43d2b521859d724081fae39dd1e281726e0fafe28`,
size 64847 bytes, first line `# Snapback — Implementation Specification`.

Rules that bind every task:
- SPEC.md is verbatim and is NEVER edited, reformatted or honesty-scanned. Honesty
  and backend-name checks apply to `README.md` and `ARCHITECTURE.md` only.
- Honesty words (standards.md honesty gate) are matched as whole words,
  case-insensitive: `production-ready`, `cross-platform`, `static`,
  `Finder-integrated`. Whole-word matching means `staticcheck` is not a hit, but
  prefer not to write it in README/ARCHITECTURE at all.
- Public docs (README) present Snapback as a Restic tool only (SPEC §1 internal
  rule, §19): no other backend or snapshot technology is named, and the
  `SnapshotProvider` seam is not mentioned. ARCHITECTURE.md (internal) may name the
  seam but no other backend.
- References to other stories' files (CONTRIBUTING.md, SECURITY.md, docs site,
  install.sh) are links only; this story creates none of them.
- After merge, spec citations "README.md §N" in later planning resolve to
  "SPEC.md §N" (orchestrator note).

## T1 — Move the spec to SPEC.md byte-for-byte
Files: `SPEC.md` (via `git mv README.md SPEC.md`), `test/projectdocs/helpers_test.go`,
`test/projectdocs/spec_test.go`.
Run `git mv README.md SPEC.md` first, before any other edit, so the Rev 2 spec keeps
its git history and no byte changes. Do not open SPEC.md in an editor that rewrites
line endings or trailing newlines. Add `package projectdocs` helpers: `repoRoot(t)`
(walks up to the dir holding `go.mod`), `readDoc(t, rel) string` (fails naming the
path when missing or empty), and `section(doc, heading) string` (body between a
`## ` heading and the next `## `). Add the hash test hard-coding the literal
constant above. README.md is absent between T1 and T2; that is expected.

## T2 — New README: restore-story lead, GIF TODO, pre-release status, install
Files: `README.md`, `test/projectdocs/readme_test.go`.
Create a new `README.md` starting `# Snapback`. Before the first `## ` heading:
the line "Restoring a file should be as easy as it was in 2008." (SPEC §19 framing),
the tagline "Time Machine-style restore for your Restic backups, in every
directory.", a `cp .snapshot/<timestamp>/report.docx .` example, and the literal
marker `<!-- TODO: restore GIF (no restore exists yet) -->`. Add `## Status` saying
Snapback is pre-release and nothing restores yet (only `snapback version` exists).
Add `## Install` with the one-line placeholder
`curl -fsSL https://snapback.example.com/install.sh | sh` and a sentence saying the
domain is a placeholder until the first release; note Restic and FUSE are
prerequisites. No claims of a working restore, no honesty words.

## T3 — README: httm credit and links
Files: `README.md`, `test/projectdocs/readme_links_test.go`.
Add `## Prior art and credit` crediting httm prominently (SPEC §19): name `httm`,
link `https://github.com/kimono-koans/httm`, author kimono-koans, licence MPL-2.0,
and say it is the direct inspiration. Describe httm as a file-level, Time
Machine-like CLI/TUI for browsing and restoring past file versions WITHOUT listing
the filesystems or tools it supports beyond Restic. Add `## Documentation` linking
`SPEC.md`, `ARCHITECTURE.md`, `CONTRIBUTING.md`, `SECURITY.md` and the docs site
`https://adeelahmad.github.io/snapback/` (links only).

## T4 — README honesty and Restic-only scan
Files: `test/projectdocs/honesty_test.go` (and `README.md` only if the scan fails).
Add whole-word, case-insensitive scans of README.md and ARCHITECTURE.md for the four
honesty words and for the phrases `first-of-its-kind`/`first of its kind` and
`multi-backend`. Add a README-only scan for other backend names (see plan.md T4 list)
and for `SnapshotProvider`. SPEC.md is explicitly excluded from both scans.

## T5 — ARCHITECTURE.md skeleton
Files: `ARCHITECTURE.md`, `test/projectdocs/architecture_test.go`.
Headings `# Architecture`, `## Current state`, `## Planned modules`,
`## Provider seam`. `## Current state` says plainly that only `cmd/snapback` and
`internal/version` exist today. `## Planned modules` is a table of the SPEC §4
modules, each in backticks: `config`, `provider`, `provider/restic`, `resolver`,
`projection`, `mount`, `links`, `discovery/seed`, `discovery/onaccess`,
`discovery/explicit`, `prewarm`, `daemon`, `web`, `service`,
`macos/FinderCompanion`, with a one-line responsibility and a link to `SPEC.md`.
`## Provider seam` describes the `SnapshotProvider` interface and states Restic is
the only implementation; it names no other backend. No honesty words (describe the
build as `CGO_ENABLED=0`, never with the banned adjective).

## T6 — NOTICE
Files: `NOTICE`, `test/projectdocs/notice_test.go`.
Plain text: project name and copyright line, then a dependency licence list. At
Stage 0 the only entry is "Go standard library — BSD-3-Clause
(https://go.dev/LICENSE)", followed by the sentence "This list is updated as
dependencies are added." The test also cross-checks every `require` module in
`go.mod` appears in NOTICE (zero today), so later dependency additions fail until
NOTICE is updated.

## T7 — CLAUDE.md
Files: `CLAUDE.md`, `test/projectdocs/claude_md_test.go`.
Follow `~/.claude/rules/project-docs.md` template, at most 80 lines: `# Snapback`
plus one-line description, then `## Stack` (Go 1.27, Restic CLI, FUSE),
`## Commands` (`go build ./cmd/snapback`, `go test -race ./...`, `go vet ./...`,
`golangci-lint run`), `## Structure` (`cmd/snapback`, `internal/version`, `test/`,
`docs/`), `## Architecture` (2-4 sentences, link ARCHITECTURE.md), `## Conventions`
(TDD, conventional commits, stdlib-first), `## Key Context` (links SPEC.md,
DEVLOG.md, `docs/agents/sprint1/standards.md`; note the honesty gate by reference).
Omit `## Environment` (no env vars yet), as the template allows.

## T8 — DEVLOG.md
Files: `DEVLOG.md`, `test/projectdocs/devlog_test.go`.
Follow `~/.claude/rules/workflow.md`: `# Snapback Dev Log`, `## Working State` with
`**Session:** 1 | **Date:** 2026-09-22`, `### Active Task` (Sprint 1 scaffolding),
`### Key Files (current shape)` (max 5 entries, each `**\`path\`**` plus inline
summary), `### Decisions (active)` (spec moved to SPEC.md and why),
`### Next Steps`, `### Blockers`, `### Watch Out`; then `## Session Archive`,
`## Milestones`, `## Mistakes & Lessons`, `## Technical Debt & Future Ideas`.
Working State at most 80 lines.
