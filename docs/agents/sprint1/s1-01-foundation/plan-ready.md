---
type: plan-ready
story: S1-01
from_red_at: 2026-09-22T00:32:32Z
---

# S1-01 plan-ready (RED verified by orchestrator at chain/s1-01 @ 718df3b)

## T1 — Module and ignore file
- [x] (no test bullets) go.mod + .gitignore at 08d0c5b; ticks when the story's GREEN merges with a green matrix

## T2 — internal/version
- [x] `internal/version/version_test.go::TestFormat` — FAIL: Format(...) = "", want "snapback dev (commit none, target linux/amd64)\n"
- [x] `internal/version/version_test.go::TestDefaults` — FAIL: Version = "", want "dev"
- [x] `internal/version/version_test.go::TestStringUsesPackageVars` — FAIL: String() = "shim", want Format(...)

## T3 — cmd/snapback run
- [x] `cmd/snapback/run_test.go::TestRunVersion` — FAIL: run([version]) = 1, want 0
- [x] `cmd/snapback/run_test.go::TestRunUsageErrors` — FAIL: run([...]) = 1, want 2

## T4 — cmd/snapback binary (ldflags)
- [x] `cmd/snapback/main_test.go::TestBinaryVersionLdflags` — FAIL: stdout = "", want "snapback v1.2.3 (commit abc1234, target linux/arm64)\n"
- [x] `cmd/snapback/main_test.go::TestBinaryDefaultsWithoutLdflags` — FAIL: stdout = ""
- [x] `cmd/snapback/main_test.go::TestBinaryUsageExitCode` — FAIL: err = <nil>, want *exec.ExitError code 2
