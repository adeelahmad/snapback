---
type: validate
story: S1-08
---

# S1-08 validate — PASS/FAIL rubric

Run every command from the repo root `/Users/adeelahmad/work/snapback`. A row PASSES only when the literal expected output matches; anything else is FAIL and cites the rule it violates. Frozen spec hash: `6bc35dadf698271b23790db43d2b521859d724081fae39dd1e281726e0fafe28` (64847 bytes; `git show e13192d:README.md`).

## Pre-flight

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| S1-01 merged | `go list -m` | `github.com/adeelahmad/snapback` | stories.md S1-08 "Depends on S1-01" |
| Stage-0 spec intact before T1 | `git show HEAD:README.md \| shasum -a 256` | `6bc35dadf698271b23790db43d2b521859d724081fae39dd1e281726e0fafe28  -` | stories.md S1-08 constraint "SPEC.md equals the old README byte-for-byte" |
| Scope | `git diff --name-only master -- . ':!docs'` | only `README.md`, `SPEC.md`, `ARCHITECTURE.md`, `NOTICE`, `CLAUDE.md`, `DEVLOG.md`, `test/projectdocs/*` | stories.md S1-08 owned files |
| Stdlib only | `grep -c '^require' go.mod` | `0` | standards.md; stories.md S1-01 "Standard library only" |

## T1 — Move the spec to SPEC.md byte-for-byte

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| Hash | `shasum -a 256 SPEC.md` | `6bc35dadf698271b23790db43d2b521859d724081fae39dd1e281726e0fafe28  SPEC.md` | stories.md S1-08 failure "spec overwritten or truncated" |
| Byte diff vs Stage 0 | `git show e13192d:README.md \| cmp - SPEC.md && echo same` | `same` | stories.md S1-08 constraint |
| Rename tracked | `git diff --cached -M --name-status master -- SPEC.md README.md \| head -1` (or on the merged commit `git show -M --name-status HEAD`) | `R100	README.md	SPEC.md` or `A SPEC.md` + `M README.md` with hash row PASS | tasks.md T1 (`git mv` first) |
| Tests | `go test -race -run 'TestRepoRootHasGoMod\|TestSectionExtractsBody\|TestSpec\|TestReadmeIsNotTheSpec' ./test/projectdocs/` | `ok  	github.com/adeelahmad/snapback/test/projectdocs` | plan.md T1 |

## T2 — New README: restore-story lead, GIF TODO, pre-release status, install

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| Lead | `awk '/^## /{exit} {print}' README.md \| grep -ci 'restoring a file should be as easy as it was in 2008'` | `1` | SPEC.md §19 public framing; §18 "lead with the restore story" |
| GIF TODO | `awk '/^## /{exit} {print}' README.md \| grep -c '<!-- TODO: restore GIF'` | `1` | stories.md S1-08 intent; SPEC.md §18 furniture |
| Install line | `grep -cE 'curl -fsSL https://[a-z0-9.-]*example\.com/install\.sh \| sh' README.md` | `1` | tasks.md T2; S1-03 placeholder-domain constraint |
| Tests | `go test -race -run 'TestReadmeLeads\|TestReadmeTitle\|TestReadmeHasGif\|TestReadmeStates\|TestReadmeOneLine\|TestReadmeClaimsNo' ./test/projectdocs/` | `ok  	github.com/adeelahmad/snapback/test/projectdocs` | plan.md T2 |

## T3 — README: httm credit and links

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| httm link | `grep -c 'https://github.com/kimono-koans/httm' README.md` | `>= 1` | SPEC.md §19 "Credit it prominently in the README" |
| Doc links | `grep -oE '\]\((SPEC\|ARCHITECTURE\|CONTRIBUTING\|SECURITY)\.md\)' README.md \| sort -u \| wc -l` | `4` | stories.md S1-08 intent (links to SPEC, CONTRIBUTING, SECURITY) |
| Tests | `go test -race -run 'TestReadmeCreditsHttm\|TestHttmCreditIsProminent\|TestReadmeLinksDocs' ./test/projectdocs/` | `ok  	github.com/adeelahmad/snapback/test/projectdocs` | plan.md T3 |

## T4 — README honesty and Restic-only scan

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| Honesty words (public docs only) | `grep -niwE 'production-ready\|cross-platform\|static\|finder-integrated' README.md ARCHITECTURE.md` | no output | standards.md honesty gate; SPEC.md §1, §18 |
| Other backends in README | `grep -niwE 'zfs\|btrfs\|nilfs2?\|borg\|borgbackup\|kopia\|duplicity\|duplicati\|ec2\|ebs\|rsnapshot\|tarsnap\|bup\|snapper' README.md ARCHITECTURE.md` | no output | SPEC.md §1 internal rule; §19 "do not mention any backend other than Restic" |
| Seam hidden | `grep -ci snapshotprovider README.md` | `0` | SPEC.md §1 "not for public docs" |
| SPEC not scanned | (no command) SPEC.md is deliberately excluded; it contains these words verbatim | n/a | tasks.md rules |
| Tests | `go test -race -run 'TestNoHonesty\|TestHonestyRegex\|TestNoFirstOfKind\|TestReadmeMentionsOnly\|TestArchitectureNamesNo\|TestReadmeHides\|TestSpecIsNotScanned' ./test/projectdocs/` | `ok  	github.com/adeelahmad/snapback/test/projectdocs` | plan.md T4 |

## T5 — ARCHITECTURE.md skeleton

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| Current state | `grep -c 'exist today' ARCHITECTURE.md` | `>= 1` | stories.md S1-08 "says plainly that only cmd/snapback and internal/version exist today" |
| Tests | `go test -race -run 'TestArchitecture' ./test/projectdocs/` | `ok  	github.com/adeelahmad/snapback/test/projectdocs` | plan.md T5; SPEC.md §4 |

## T6 — NOTICE

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| Content | `grep -c 'BSD-3-Clause' NOTICE` | `>= 1` | SPEC.md §18 "NOTICE with dependency licences"; stories.md failure "NOTICE is empty" |
| Tests | `go test -race -run 'TestNotice' ./test/projectdocs/` | `ok  	github.com/adeelahmad/snapback/test/projectdocs` | plan.md T6 |

## T7 — CLAUDE.md

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| Line budget | `wc -l < CLAUDE.md` | integer `<= 80` | `~/.claude/rules/project-docs.md` "under 80 lines"; stories.md failure |
| Tests | `go test -race -run 'TestClaudeMd' ./test/projectdocs/` | `ok  	github.com/adeelahmad/snapback/test/projectdocs` | plan.md T7 |

## T8 — DEVLOG.md

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| Working State | `grep -cE '^## (Working State\|Session Archive\|Milestones)$' DEVLOG.md` | `3` | `~/.claude/rules/workflow.md` DEVLOG template |
| Tests | `go test -race -run 'TestDevlog' ./test/projectdocs/` | `ok  	github.com/adeelahmad/snapback/test/projectdocs` | plan.md T8 |

## Final sign-off

Run the standards matrix verbatim; every line must exit 0.

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| gofmt | `test -z "$(gofmt -l .)"` | exit 0, no output | `~/.claude/rules/golang/coding-style.md` |
| goimports | `test -z "$(goimports -l .)"` | exit 0, no output | `~/.claude/rules/golang/coding-style.md` |
| Build | `CGO_ENABLED=0 go build ./...` | exit 0, no output | SPEC.md §2, §5, §17 |
| Vet | `go vet ./...` | exit 0, no output | SPEC.md §18 |
| Lint | `golangci-lint run` | exit 0, no output | SPEC.md §18 |
| Race | `go test -race ./...` | every package `ok` (incl. `test/projectdocs`) | SPEC.md §18; golang/testing.md |
| Coverage | `go test -race -covermode=atomic -coverprofile=coverage.out ./... && go tool cover -func=coverage.out \| awk '/^total:/{gsub("%","",$NF); if ($NF+0 < 80) exit 1; print "coverage " $NF "% OK (>=80%)"}'` | `coverage NN.N% OK (>=80%)` (test-only package adds no production statements) | `~/.claude/rules/common/testing.md` |
| Vulns | `govulncheck ./...` | `No vulnerabilities found.` | SPEC.md §18 |
| Spec hash (final) | `shasum -a 256 SPEC.md \| cut -d' ' -f1` | `6bc35dadf698271b23790db43d2b521859d724081fae39dd1e281726e0fafe28` | stories.md S1-08 constraint |
| Honesty | `grep -rniwE 'production-ready\|cross-platform\|static\|finder-integrated' README.md ARCHITECTURE.md` | no output | standards.md honesty gate |
| Plan ticked | `grep -c '^- \[ \]' docs/agents/sprint1/s1-08-project-docs/plan.md` | `0` after execution | standards.md TDD workflow |
