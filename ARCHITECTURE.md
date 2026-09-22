# Architecture

[SPEC.md](SPEC.md) is the authoritative design; this file tracks what is built against it.

## Current state

These packages exist today:

- `cmd/snapback` — the CLI entry point and daemon wiring.
- `internal/version` — build version information.
- `internal/cli` — the subcommands behind a shared contract.
- `internal/config` — configuration types, validation and persistence.
- `internal/provider` — the `SnapshotProvider` interface; `internal/provider/restic` implements it.
- `internal/resolver`, `internal/aliases`, `internal/history` — per-directory snapshot selection, timestamp aliases and the history view.
- `internal/projection` — the immutable in-memory catalog.
- `internal/mount`, `internal/mount/gofuse` — the FUSE seam and its adapter; `internal/readerpolicy` decides which processes may read the mount; `internal/recovery` cleans up mounts a crashed daemon left behind.
- `internal/links` — managed `.snapshot` symlinks.
- `internal/discovery/seed` — targeted path seeding and a watcher for new directories.
- `internal/shellhook` — shell snippets that report directory changes.
- `internal/prewarm`, `internal/refresh` — metadata pre-warm and snapshot refresh.
- `internal/daemon`, `internal/ipc`, `internal/status` — process lifecycle, instance lock, the local JSON control channel and status reporting.
- `internal/web`, `internal/webui` — the authenticated local HTTP API and web UI with embedded assets.
- `internal/service` — background service install and control.
- `internal/doctor` — read-only health checks.
- `internal/errcode`, `internal/pathutil`, `internal/rawpath` — error codes and path helpers.
- `internal/compat` — compatibility, fidelity and latency test harnesses.

## Planned modules

From [SPEC.md](SPEC.md) section 4. Status reflects the packages above.

| Module | Responsibility | Status |
| --- | --- | --- |
| `config` | Parse, validate, migrate and atomically persist configuration. | Implemented |
| `provider` | The `SnapshotProvider` interface. | Implemented |
| `provider/restic` | Restic implementation: subprocess execution, mount supervision, `backup`, `ls`. | Implemented |
| `resolver` | Root and per-directory snapshot selection, tree-path mapping, timestamp aliases. | Implemented |
| `projection` | Pure catalog model independent of any FUSE library. | Implemented |
| `mount` | OS FUSE adapter for the history catalog. | Implemented |
| `links` | Safe creation, registry, repair and removal of managed symlinks. | Implemented |
| `discovery/seed` | Targeted path seeding and a watcher for new directories. | Implemented |
| `discovery/onaccess` | On-access hooks (fanotify on Linux, Endpoint Security on macOS). | Planned |
| `discovery/explicit` | Shell hook events, `link`/`open` commands, Finder Sync events. | Partial: shell hooks in `internal/shellhook` |
| `prewarm` | Metadata pre-warm of the newest selected snapshots. | Implemented |
| `daemon` | Process lifecycle, instance lock, local IPC, refresh, health. | Implemented |
| `web` | Embedded assets and the authenticated local HTTP API. | Implemented |
| `service` | launchd, systemd and OpenRC adapters. | Implemented |
| `macos/FinderCompanion` | Swift app and Finder Sync extension. | Planned |

## Provider seam

All snapshot access goes through the `SnapshotProvider` interface (SPEC.md section 5): validate a repository, list snapshots, mount read-only, resolve a snapshot root, probe a tree path, create an on-demand snapshot, pre-warm metadata, and stop the mount.

Restic is the only implementation. The seam exists so that a second implementation would be a new package with no changes to `resolver`, `projection`, `links`, `discovery`, `daemon` or `web`.
