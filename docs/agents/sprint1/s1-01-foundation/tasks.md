---
type: tasks
story: S1-01
---

# S1-01 tasks — Go module and `snapback version` binary

Story intent and owned files: `docs/agents/sprint1/stories.md` (S1-01). Every task below touches only S1-01's owned set: `go.mod`, `go.sum` (not created; stdlib only), `cmd/snapback/**`, `internal/version/**`, `.gitignore`. Tests for every task are listed in `plan.md`; gate commands in `validate.md`.

Decisions fixed here so workers do not guess:

- Module path `github.com/adeelahmad/snapback`; `go 1.27` and `toolchain go1.27.1` (current stable patch per `https://go.dev/VERSION?m=text` on 2026-09-22; not 1.22.x, so govulncheck does not flag stdlib advisories).
- Output contract of the formatter (single line, trailing newline): `snapback <Version> (commit <Commit>, target <Target>)\n`.
- Usage contract on stderr: first line `usage: snapback version`.
- ldflags `-X` paths (consumed verbatim by S1-06): `github.com/adeelahmad/snapback/internal/version.Version`, `.Commit`, `.Target`.

## T1 — Module and ignore file

Create `go.mod` with module path `github.com/adeelahmad/snapback`, `go 1.27`, `toolchain go1.27.1`, no `require` block. Append `coverage.out` and `dist/` to `.gitignore` (keep existing lines). No `go.sum` (stdlib only).

Files: `go.mod`, `.gitignore`.

## T2 — `internal/version` vars and pure formatter

Package `version` declares `var Version = "dev"`, `var Commit = "none"`, `var Target = runtime.GOOS + "/" + runtime.GOARCH` (package-level `var`, not `const`, so `-ldflags -X` can set them). Exposes `func Format(version, commit, target string) string` (pure, no globals) and `func String() string` returning `Format(Version, Commit, Target)`.

Files: `internal/version/version.go`, `internal/version/version_test.go`.

## T3 — Command dispatch (`run`) with injectable streams

`func run(args []string, stdout, stderr io.Writer) int` in package `main`: `args == ["version"]` writes `version.String()` to stdout and returns 0; missing, unknown, or extra arguments write the usage line to stderr and return 2. No other subcommands, flags, config, restic or FUSE code.

Files: `cmd/snapback/run.go`, `cmd/snapback/run_test.go`.

## T4 — Thin `main` and ldflags binary test

`func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }` — nothing else. Add the build-and-exec test that compiles `./cmd/snapback` with `CGO_ENABLED=0` and `-ldflags -X` for all three vars into `t.TempDir()` and runs the binary.

Files: `cmd/snapback/main.go`, `cmd/snapback/main_test.go`.
