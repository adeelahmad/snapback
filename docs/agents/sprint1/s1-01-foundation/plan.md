---
type: plan
story: S1-01
scope: "tests only"
---

# S1-01 test plan (tests only)

Contracts under test (from `tasks.md`): `Format` renders `snapback <v> (commit <c>, target <t>)\n`; usage line is `usage: snapback version`; `run` returns 0 for `version`, 2 otherwise.

## T2 — internal/version

- [ ] `internal/version/version_test.go::TestFormat` — table: {"dev","none","linux/amd64"}, {"v1.2.3","abc1234","darwin/arm64"}, {"","",""}; call `Format(v,c,t)`; assert exact string `snapback <v> (commit <c>, target <t>)\n` per row.
- [ ] `internal/version/version_test.go::TestDefaults` — no ldflags; read package vars; assert `Version=="dev"`, `Commit=="none"`, `Target==runtime.GOOS+"/"+runtime.GOARCH`.
- [ ] `internal/version/version_test.go::TestStringUsesPackageVars` — set `Version="v9.9.9"`, `Commit="deadbee"`, `Target="linux/mipsle"` (restore via `t.Cleanup`); call `String()`; assert equals `Format("v9.9.9","deadbee","linux/mipsle")`.

## T3 — cmd/snapback run

- [ ] `cmd/snapback/run_test.go::TestRunVersion` — args `["version"]`, bytes.Buffer stdout/stderr; call `run`; assert return 0, stdout == `version.String()`, stderr empty.
- [ ] `cmd/snapback/run_test.go::TestRunUsageErrors` — table: `nil`, `[]`, `["bogus"]`, `["--version"]`, `["version","extra"]`, `["Version"]`; call `run`; assert return 2 (non-zero), stdout empty, stderr starts with `usage: snapback version`.

## T4 — cmd/snapback binary (ldflags)

- [ ] `cmd/snapback/main_test.go::TestBinaryVersionLdflags` — `go build` with `CGO_ENABLED=0` and `-ldflags "-X github.com/adeelahmad/snapback/internal/version.Version=v1.2.3 -X ...Commit=abc1234 -X ...Target=linux/arm64"` into `t.TempDir()`; exec `snapback version`; assert exit 0, stdout exactly `snapback v1.2.3 (commit abc1234, target linux/arm64)\n`, stderr empty (catches `-X` path drift and output-on-stderr).
- [ ] `cmd/snapback/main_test.go::TestBinaryDefaultsWithoutLdflags` — `go build` with no ldflags into `t.TempDir()`; exec `snapback version`; assert exit 0 and stdout contains `snapback dev (commit none, target <GOOS>/<GOARCH>)`.
- [ ] `cmd/snapback/main_test.go::TestBinaryUsageExitCode` — reuse a built binary; exec with no args and with `bogus`; assert `*exec.ExitError` with `ExitCode()==2`, stdout empty, stderr contains `usage: snapback version` (unknown subcommand must not exit 0).

Test count: 8.
