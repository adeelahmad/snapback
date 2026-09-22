---
type: output
story: S2-14
---

## S2-14/T1 · attempt 1 · green-worker · 2026-09-22T04:26:00Z
status: ok
### Summary
Applied the human's 8-point SPEC.md review as surgical edits in the spec's voice, with no section renumbering; the only new heading is §22.1. §1 now states the central idea and the design anchor, and it corrects the FUSE rationale. Static linkage is claimed for Linux only, with evidence required; macOS is described as a self-contained Go application executable, with linkage still to verify. The Versions panel lists snapshot occurrences, and size plus mtime is allowed only as a "likely identical" presentational grouping. `source_paths_exact` now compares canonical sets. `catalog.reader_policy` is stated not to be an access-control boundary. The new §22.1 release contract is referenced from the §2 table, and §2, §9, §17, §18, §20, §22 and §24 are reconciled with it. Every listed invariant is untouched. One test pin was updated: test/projectdocs/spec_test.go specSHA256/specSize (and its stale restore hint), per the orchestrator exception. Commit 30122aa.
### Result
| Check | Status | Detail |
|---|---|---|
| `GOTOOLCHAIN=auto go test ./test/...` | PASS | all 5 packages ok |
| `go test -race ./...` | PASS | no failures |
| `go vet ./...` | PASS | clean |
| `golangci-lint run` | PASS | 0 issues |
| `mkdocs build --strict` | PASS | built |
| `test/projectdocs/spec_test.go::TestSpecHashMatchesStage0Readme` | PASS | pin updated to d345ca55..., size 68524 |
| diff scope | PASS | SPEC.md, test/projectdocs/spec_test.go only |
### Next
Structural review / human read of §22.1 (OpenRC placed in the follow-up row because the v0.1 list names only systemd).
