# Snapback

Time Machine-style restore for Restic backups: a read-only `.snapshot` entry in every directory.

## Stack

- Go 1.27 (toolchain go1.27.1), module `github.com/adeelahmad/snapback`
- Restic repositories as the only backend
- Linux first; FUSE for the `.snapshot` view (planned)

## Commands

```
go build ./...                       # Build all packages
go test -race ./...                  # Run all tests with the race detector
go vet ./...                         # Vet
golangci-lint run                    # Lint
GOTOOLCHAIN=auto go test -race ./... # If the local Go is older than go.mod
```

## Structure

```
cmd/snapback/     # CLI entry point (main.go, run.go)
internal/version/ # Build version info
test/projectdocs/ # Tests that pin the project docs (README, CLAUDE, DEVLOG, ...)
docs/agents/      # Sprint planning and agent channel files (not committed per task)
SPEC.md           # Product spec: what Snapback is and is not
```

## Architecture

Today the binary is a skeleton: `cmd/snapback` supports only `snapback version`,
which prints build info from `internal/version`. The planned design puts a read-only FUSE view behind each
directory's `.snapshot` entry, fed by Restic snapshot metadata through a single
provider seam. See ARCHITECTURE.md for current state versus planned modules.

## Conventions

- Strict TDD: tests first (RED), minimal code to pass (GREEN), then review
- Standard library first; every new dependency needs a reason and a NOTICE entry
- Keep packages under `internal/` unless they are a deliberate public API
- Conventional commits (`feat:`, `fix:`, `docs:`, `test:`, `ci:`, `chore:`)
- Docs stay honest: no claims about features that do not exist yet
- Pin tools and CI actions to released versions, never `latest`

## Key Context

- SPEC.md: product scope, non-goals and wording rules for public docs
- ARCHITECTURE.md: current state, planned modules and the provider seam
- DEVLOG.md: session journal, decisions, mistakes and technical debt
- docs/agents/sprint1/standards.md: Go rule digest and the gate matrix
- Restic is the only supported backend; do not name or promise others
- Read-only by design: Snapback never writes to the Restic repository
