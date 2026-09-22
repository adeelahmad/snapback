---
type: validate
story: S2-04
---

# S2-04 validate — PASS/FAIL rubric

All commands run from `/Users/adeelahmad/work/snapback` (or the task worktree). PASS requires every row's expected output. Any other output is FAIL. Per M-011, a `files: []` line from md-db is not a failure by itself; judge by exit code.

**Where each platform's evidence comes from.** The macOS real-mount run is **local** on this Apple Silicon host with macFUSE installed (`/Library/Filesystems/macfuse.fs`); there is no macOS CI runner this sprint (stories.md Out of scope). The Linux real-mount run happens **in CI**, in the `fuse-linux` job that S2-05 adds; `catalog-linux.json` is copied by the orchestrator from that job's `stage1-evidence-linux` artifact on a green `master` run. It is not produced by a worker and must not be hand-written. Until both files exist, the platform without evidence is reported as **implemented but not tested here** (standards.md integration-test gating).

## Pre-flight

| Check | Command | Expected | FAIL cites |
| --- | --- | --- | --- |
| S2-01 pin landed | `grep -E '^\s*github.com/hanwen/go-fuse/v2 v[0-9]+\.[0-9]+\.[0-9]+$' go.mod` | one line, exact `vX.Y.Z` tag (v2.11.0 per stories.md) | stories.md S2-04 Connections (depends on S2-01) |
| S2-01 seam and attr.go landed | `test -f internal/mount/mount.go && test -f internal/mount/gofuse/attr.go && echo ok` | `ok` | stories.md S2-01 Owned files |
| S2-02 projection landed | `go list ./internal/projection` | `github.com/adeelahmad/snapback/internal/projection` | stories.md S2-04 Connections (depends on S2-02) |
| S2-01/S2-02 green | `go test -race -count=1 ./internal/mount/ ./internal/projection/...` | every line starts `ok` | sprint plan wave 2 start condition |
| Only owned files changed | `git diff --name-only master... \| grep -vE '^(internal/mount/gofuse/(fs\|fs_test\|adapter\|adapter_test\|mount_integration_test)\.go\|docs/reports/stage1/catalog-(darwin\|linux)\.json\|docs/agents/sprint2/s2-04-gofuse-catalog/)'` | no output | stories.md S2-04 Owned files; M-012 |
| No go.mod/go.sum/attr.go edit | `git diff --name-only master... -- go.mod go.sum internal/mount/mount.go internal/mount/gofuse/attr.go` | no output | stories.md S2-04 Constraints (S2-01 owns these) |
| macFUSE present for the local run (T4 only) | `test -d /Library/Filesystems/macfuse.fs && echo present` | `present` | SPEC §5 (real macOS compatibility test) |

## T1 — Read-path nodes

| Check | Command | Expected | FAIL cites |
| --- | --- | --- | --- |
| T1 tests | `go test -race -count=1 -run 'TestLookupReturnsCatalogChild\|TestLookupMissingReturnsENOENT\|TestLookupInodeStableAcrossCalls\|TestReaddirListsCatalogEntriesInOrder\|TestReadlinkReturnsExactTargetBytes\|TestGetattrMatchesAttrTranslation\|TestLookupSetsEntryAndAttrTimeouts\|TestStatfsReportsNoFreeSpace\|TestObserverReceivesOneEventPerOp' -v ./internal/mount/gofuse/` | 9 top-level `--- PASS` lines, final `ok` | stories.md S2-04 Success scenarios; SPEC §7 |
| No file-content read path | `grep -nE 'fs\.Node(Opener\|Reader\|Opendirer)' internal/mount/gofuse/fs.go` | no output | stories.md S2-04 Constraints |
| Package gate (M-005) | `GATE_RUN_MATRIX=0` plus `go vet ./internal/mount/gofuse/ && golangci-lint run ./internal/mount/gofuse/...` | exit 0, no findings | standards.md gate matrix |

## T2 — Mutations return EROFS

| Check | Command | Expected | FAIL cites |
| --- | --- | --- | --- |
| T2 tests | `go test -race -count=1 -run 'TestNodesImplementMutationInterfaces\|TestDirMutationsReturnEROFS\|TestSymlinkSetattrReturnsEROFS\|TestWriteReturnsEROFS' -v ./internal/mount/gofuse/` | 4 top-level `--- PASS` lines, final `ok` | SPEC §7 ("Return EROFS for mutations"); stories.md S2-04 Failure scenarios |
| No wrong errno | `grep -nE 'syscall\.(EPERM\|ENOSYS\|ENOTSUP\|EACCES)' internal/mount/gofuse/fs.go` | no output | stories.md S2-04 Failure scenarios (EPERM/ENOSYS instead of EROFS) |

## T3 — Adapter

| Check | Command | Expected | FAIL cites |
| --- | --- | --- | --- |
| T3 tests | `go test -race -count=1 -run 'TestAdapterSatisfiesMountAdapter\|TestProjectionGenerationSatisfiesMountCatalog\|TestMountOptionsReadOnlyAndPrivate\|TestMountRejectsNonEmptyDir\|TestMountRejectsMissingDir\|TestMountRejectsNonDirectory\|TestUnmountBeforeMountReturnsError' -v ./internal/mount/gofuse/` | 7 top-level `--- PASS` lines, final `ok` | stories.md S2-04 Success scenarios; SPEC §1 (never mount over a live directory) |
| No timeout on the mount lifetime | `grep -nE 'context\.With(Timeout\|Deadline)' internal/mount/gofuse/adapter.go` | no output | SPEC §5; standards.md (no timeouts on the long-running mount) |
| allow_other never set | `grep -nE 'AllowOther:\s*true\|allow_other' internal/mount/gofuse/adapter.go` | no output | SPEC §7; stories.md S2-04 Constraints |

## T4 — Real-mount integration test

| Check | Command | Expected | FAIL cites |
| --- | --- | --- | --- |
| Build tag present | `head -1 internal/mount/gofuse/mount_integration_test.go` | `//go:build integration` | standards.md integration-test gating |
| Skip names the prerequisite | `go test -race -count=1 -tags=integration -run TestCatalogMountLifecycle -v ./internal/mount/gofuse/` with `SNAPBACK_FUSE_TESTS` unset | `--- SKIP: TestCatalogMountLifecycle` with `SNAPBACK_FUSE_TESTS not set: FUSE catalog tests skipped` | standards.md skip discipline |
| macOS real mount (local, macFUSE) | `mkdir -p .tmp/ev && SNAPBACK_FUSE_TESTS=1 SNAPBACK_EVIDENCE_DIR=$PWD/.tmp/ev go test -race -count=1 -tags=integration -run TestCatalogMountLifecycle -v ./internal/mount/gofuse/` | `--- PASS: TestCatalogMountLifecycle` (not SKIP), final `ok` | SPEC §5 (real macOS test on Apple Silicon); stories.md S2-04 Success scenarios |
| macOS evidence content | `jq -e '.platform=="darwin/arm64" and .result=="pass" and (.go_fuse_version\|test("^v[0-9]+\\.[0-9]+\\.[0-9]+$")) and (.operations\|length>0)' .tmp/ev/catalog-darwin.json` | `true` | stories.md S2-04 What's wanted (evidence fields) |
| No stale mount after the run | `mount \| grep -c TestCatalogMountLifecycle` | `0` | stories.md S2-04 Failure scenarios (stale mount) |
| go-fuse imported only here | `grep -rlE '"github.com/hanwen/go-fuse' --include='*.go' . \| grep -v '^./internal/mount/gofuse/'` | no output | stories.md S2-04 Constraints; SPEC §5 |

## Final sign-off

Run by T4 (the last task) against the merged story. This is the full standards matrix, not the package-scoped one (M-005).

| Check | Command | Expected | FAIL cites |
| --- | --- | --- | --- |
| Format (gofmt) | `test -z "$(gofmt -l .)" && echo clean` | `clean` | standards.md gate matrix |
| Format (goimports) | `test -z "$(goimports -l .)" && echo clean` | `clean` | standards.md gate matrix |
| Build (static) | `CGO_ENABLED=0 go build ./...` | exit 0, no output | SPEC §5, §17 |
| Build with integration tag | `CGO_ENABLED=0 go vet -tags=integration ./...` | exit 0, no output | standards.md integration-test gating |
| Vet | `go vet ./...` | exit 0, no output | SPEC §18 |
| Lint | `golangci-lint run` | exit 0, no findings (test files included, M-004) | standards.md gate matrix |
| Test + race (whole repo) | `go test -race -count=1 ./...` | every package `ok` or `no test files` | SPEC §18, §20 |
| Coverage (whole repo and package) | `go test -race -covermode=atomic -coverprofile=coverage.out ./... && go tool cover -func=coverage.out \| tail -1` | total >= 80.0%; `go test -cover ./internal/mount/gofuse/` also >= 80.0% | standards.md "How coverage is measured" |
| Vulnerability audit | `govulncheck ./...` | `No vulnerabilities found.` | standards.md gate matrix |
| actionlint | `actionlint` | exit 0, no output | standards.md gate matrix |
| shellcheck | `shellcheck -s sh install.sh` | exit 0, no output | standards.md gate matrix |
| Default-build test count | `go test -list '.*' ./internal/mount/gofuse/ \| grep -c '^Test'` | `20` plus the count of S2-01's `attr_test.go` tests in the same package (T4 adds none to the default build) | this plan.md (20 default-build checkboxes) |
| Integration test count | `go test -tags=integration -list '.*' ./internal/mount/gofuse/ \| grep -c '^TestCatalogMountLifecycle$'` | `1` | this plan.md T4 |
| Integration suite on macOS (local) | `SNAPBACK_FUSE_TESTS=1 SNAPBACK_EVIDENCE_DIR=docs/reports/stage1 go test -race -count=1 -tags=integration ./internal/mount/...` | `--- PASS: TestCatalogMountLifecycle`, no SKIP for it, final `ok`; `docs/reports/stage1/catalog-darwin.json` has `result` `pass` | sprint plan "macOS evidence (local)"; SPEC §5 |
| Linux evidence (orchestrator, post-merge, via S2-05 CI) | `gh run list --workflow ci.yml --branch master --limit 1 --json conclusion -q '.[0].conclusion'`, then download `stage1-evidence-linux` and `jq -e '.platform\|startswith("linux/") and .result=="pass"' catalog-linux.json` | `success`, then `true`; file committed at `docs/reports/stage1/catalog-linux.json` | sprint plan (Linux evidence from the `fuse-linux` artifact); stories.md S2-05 |
| No suppressions | `grep -nE 't\.Skip\(\|//nolint' internal/mount/gofuse/*.go \| grep -vE 'SNAPBACK_FUSE_TESTS not set\|/dev/fuse not present\|macFUSE not installed'` | no output | standards.md (zero suppressions; skips must name the prerequisite) |
| Honest claims | a `SKIP` of `TestCatalogMountLifecycle` on any platform | recorded as **implemented but not tested here** for that platform, never as PASS | standards.md gate-final skip counting; SPEC §22 |
