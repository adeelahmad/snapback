---
type: init
story: S2-14
---

## S2-14/T1 · attempt 1 · green-worker · 2026-09-22T04:25:00Z

### Mandate (human SPEC review, 2026-09-22, verbatim intent)
Revise SPEC.md with these refinements. Preserve the architectural direction and every invariant; this is not a redesign.

1. **FUSE wording near the top.** Replace any claim that people running Restic are already "in FUSE territory". Restic works without FUSE, and FUSE is needed only because Snapback deliberately uses `restic mount`. Use wording equivalent to: "FUSE is a hard prerequisite because Snapback exposes Restic snapshots as a browsable filesystem. `snapback doctor` detects it and explains how to install the platform prerequisite; Snapback never installs FUSE automatically."
2. **No unconditional "static" claim for every platform** (e.g. §"Shape of the software" line ~11, and the release-binaries row ~507).
   - Linux: keep the static-build requirement, claimable where `CGO_ENABLED=0` plus `file`/`ldd` evidence proves it.
   - macOS: call it a "self-contained Go application executable", with linkage and runtime requirements to be verified.
   - Do not weaken the Linux requirement.
3. **Versions-panel deduplication (~line 478).** Size plus mtime do not establish content identity, so never call such entries "distinct versions". Preferred semantics: show every snapshot occurrence unless exact content identity is available. Size plus mtime may be used only as a presentational grouping labelled "likely identical", never as proof of equality. Do not introduce content reads just for the UI.
4. **`source_paths_exact` (~line 190).** Replace "deduplicated, ordered recorded set" with a canonical comparison: preserve exact path bytes, deduplicate exact entries, sort deterministically, then compare canonical sets. The same set in a different original order must select the same backup set.
5. **`catalog.reader_policy`.** Add wording equivalent to: "`catalog.reader_policy` is a resource-protection and UX mechanism, not an access-control boundary. Processes running as the same user can change executable identity or perform equivalent filesystem operations." Keep the existing throttling and allow/deny behaviour.
6. **Explicit first public-release contract.** Elevate the §22 staged rollout into an explicit release boundary (a clear subsection, referenced from the top summary table), without deleting later features:
   - **v0.1 / first public release candidate:** Restic backend; `.snapshot` filesystem/history model; Linux first; explicit plus targeted seeded discovery; `snapback snap`; daemon; web UI; systemd; a working installer and release binary; safe link registry; refresh/prewarm; all applicable acceptance tests.
   - **Follow-up / separately proven:** macOS core plus macFUSE; launchd; Finder companion; additional package channels.
   - **Experimental until acceptance proof:** Linux fanotify on-access; macOS Endpoint Security on-access.
   Make sure nothing else in the spec implies that everything, everywhere is a prerequisite for the first release. For example, the §17 "every channel ships from the first tagged release" statement must be reconciled with this contract.
7. **Preserve, and do not simplify away:**
   - no overlay, union, passthrough or FUSE layer over live data
   - the owned `.snapshot` symlink is the only live mutation
   - `latest` = newest eligible snapshot, never "search backward until something exists"
   - a subdirectory `snap` never becomes its parent's `latest`
   - per-directory eligibility
   - per-snapshot host/source prefix mapping
   - full snapshot IDs are canonical
   - foreign `.snapshot` entries survive untouched
   - link ownership = registry evidence plus exact expected target
   - no recursive-delete repair
   - distinct failure states
   - a failed refresh keeps the last known-good immutable generation
   - cross-compilation ≠ platform support
   - Finder source ≠ Finder evidence
   - macOS support needs a real macFUSE test
   - on-access stays opt-in until measured and proven fail-open
   - no hidden fallback to live data
   - no rename inference or union filesystem
   - no invented Restic behaviour or flags
   - public claims need evidence
8. **Keep the central idea obvious.** "Every directory gets a doorway into its own backup history." The design anchor stays:
   ```
   cd ~/project
   ls .snapshot
   cp .snapshot/2026-09-20_0300Z/report.docx .
   ```
   If the spec does not already state these near the top, add them there, briefly.

### Scope
- **May:** SPEC.md. Also the tests that pin old SPEC wording, but only where a test pins wording this mandate removes; update those pins minimally, and list each in output.md (orchestrator exception, since the spec is the authoritative ground truth changed by the human).
- **May not:** anything else. No renumbering of existing section numbers (other docs reference them). No new features.

### Acceptance
- `GOTOOLCHAIN=auto go test ./test/...` all pass.
- The full standards matrix is green, including `mkdocs build --strict`.
- SPEC.md passes the repo's honesty/wording tests.
- Diff limited to scope.
- Commit `docs: refine spec per review (FUSE wording, linkage claims, versions dedup, canonical source paths, reader policy, v0.1 contract)`, with no AI trailers.
