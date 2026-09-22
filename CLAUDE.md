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
cmd/snapback/     # CLI entry point and daemon wiring
internal/cli/     # Subcommands behind a shared contract
internal/daemon/  # Instance lock, lifecycle; internal/ipc is its local JSON channel
internal/provider/ # SnapshotProvider seam; provider/restic is the only backend
internal/resolver/ # Live dir -> Restic snapshots; aliases, projection, history
internal/mount/   # FUSE seam; mount/gofuse adapts it; readerpolicy gates readers
internal/links/   # Managed .snapshot symlinks; discovery/seed plans and watches
internal/web/     # Local web UI and authenticated API; assets in internal/webui
internal/service/ # Background service install and control
internal/doctor/  # Read-only health checks
test/projectdocs/ # Tests that pin the project docs
SPEC.md           # Product spec: what Snapback is and is not
```

## Architecture

`cmd/snapback` wires the CLI (`internal/cli`) and a daemon (`internal/daemon`) that the
CLI reaches over `internal/ipc`. The daemon reads Restic snapshot metadata through the
`provider` seam, resolves each directory's snapshots (`resolver`, `aliases`, `history`)
and serves a read-only history mount (`mount`, `mount/gofuse`), linked into directories
as `.snapshot` by `links` and `discovery/seed`. `web`, `service` and `doctor` add the
local UI, background service control and health checks. See ARCHITECTURE.md.

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
- Restic is the only supported backend; name others only in a Roadmap section, marked
  "planned, not supported", and never say "multi-backend" until a second provider ships
- Read-only by design: Snapback never writes to the Restic repository
