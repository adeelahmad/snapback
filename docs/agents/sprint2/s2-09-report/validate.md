---
type: validate
story: S2-09
---

# S2-09 validate — PASS/FAIL rubric

All commands run from `/Users/adeelahmad/work/snapback` (or the task worktree). PASS requires every row's expected output. Anything else is FAIL.

## Pre-flight

| Check | Command | Expected | FAIL cites |
| --- | --- | --- | --- |
| Upstream stories merged | `git log --oneline master \| grep -cE 'S2-0(3\|4\|6\|7\|8)'` | at least 5 (each dependency merged green) | sprint2/plan.md wave 4 |
| Evidence committed | `ls docs/reports/stage1/*.json \| wc -l` | 9, or fewer only with each absent file named by the orchestrator as not-tested-here | stories.md S2-09 Success scenarios |
| latency.json present, remote deleted | `jq -e '.remote_deleted == true and .remote == "gdrive:snapback-stage1"' docs/reports/stage1/latency.json` | `true` | stories.md S2-08, S2-09 |
| Evidence parses | `for f in docs/reports/stage1/*.json; do jq -e type "$f" >/dev/null \|\| echo BAD $f; done` | no output | stories.md S2-09 |
| RED is by assertion | `go test -count=1 ./test/reports/ 2>&1 \| grep -cE 'panic:\|build failed'` | `0` on RED; failures cite the missing report | memory.md M-002, M-003 |
| Only owned files change | `git diff --name-only master... \| grep -vE '^(docs/reports/stage1-measurements\.md\|test/reports/.*)$'` | no output | stories.md S2-09 Owned files |

## T1 — Helpers, skeleton, inventory

| Check | Command | Expected | FAIL cites |
| --- | --- | --- | --- |
| T1 tests | `go test -race -count=1 -run 'TestReportHasAllSectionsInOrder\|TestEvidenceInventory\|TestEvidenceFilesParse' -v ./test/reports/` | 3 `--- PASS`, final `ok` | plan.md T1 |

## T2 — Pinned versions

| Check | Command | Expected | FAIL cites |
| --- | --- | --- | --- |
| T2 tests | `go test -race -count=1 -run 'TestGoFuseVersionMatchesGoMod\|TestPinnedVersionsTable' -v ./test/reports/` | 2 `--- PASS`, final `ok` | plan.md T2 |

## T3 — Per-platform sections

| Check | Command | Expected | FAIL cites |
| --- | --- | --- | --- |
| T3 tests | `go test -race -count=1 -run 'TestPathTemplateRowsMatchEvidence\|TestCatalogRowsMatchEvidence\|TestFidelityRowsMatchEvidence\|TestCtimeBirthTimeNotClaimed\|TestPlatformLabels' -v ./test/reports/` | 5 `--- PASS`, final `ok` | plan.md T3 |

## T4 — Latency

| Check | Command | Expected | FAIL cites |
| --- | --- | --- | --- |
| T4 tests | `go test -race -count=1 -run 'TestLatency' -v ./test/reports/` | 3 `--- PASS`, final `ok` | plan.md T4 |

## T5 — Crawler

| Check | Command | Expected | FAIL cites |
| --- | --- | --- | --- |
| T5 tests | `go test -race -count=1 -run 'TestCrawler' -v ./test/reports/` | 3 `--- PASS`, final `ok` | plan.md T5 |

## T6 — Matrix, open item, honesty

| Check | Command | Expected | FAIL cites |
| --- | --- | --- | --- |
| T6 tests | `go test -race -count=1 -run 'TestMatrix\|TestPendingDecisionLine\|TestNoThresholdOrVerdict\|TestNoBannedHonestyWords' -v ./test/reports/` | 5 `--- PASS`, final `ok` | plan.md T6 |

## Final sign-off

| Check | Command | Expected | FAIL cites |
| --- | --- | --- | --- |
| Whole package | `go test -race -count=1 ./test/reports/...` | `ok` | common/testing.md |
| Test count | `go test -list '.*' ./test/reports/ \| grep -c '^Test'` | `21` | this plan.md |
| No skips | `go test -race -count=1 -v ./test/reports/ \| grep -c -- '--- SKIP'` | `0` | sprint2/standards.md (zero suppressions); memory.md M-002 |
| Stdlib only | `go list -f '{{join .XTestImports "\n"}}' ./test/reports/ \| grep -c '\.'` | `0` (no import path contains a dot, so no third-party module) | task brief (stdlib only) |
| Evidence untouched | `git diff --name-only master... -- docs/reports/stage1/` | no output | stories.md S2-09 Constraints |
| Pending decision present, no verdict | `grep -c 'Latency go/no-go: PENDING — human decision' docs/reports/stage1-measurements.md` | `1` | stories.md S2-09; sprint2/plan.md Decisions |
| Outside docs-site | `grep -c 'stage1-measurements' mkdocs.yml` | `0` | stories.md S2-09 Constraints |
| Full standards matrix (T6 only, last task) | every command in the fenced block of sprint2/standards.md "Cross-cutting gates", in order | all exit 0, coverage at least 80% | memory.md M-005 |
| Owned files only | `git diff --stat master...` | only `docs/reports/stage1-measurements.md` and `test/reports/*_test.go` | stories.md S2-09 Owned files |
| **Human go/no-go (not an agent step)** | the human reads the report and records the decision | the agent never writes a verdict | sprint2/plan.md Decisions requiring human action |
