# Architecture

This is a skeleton. [SPEC.md](SPEC.md) is the authoritative design; this file tracks what is built against it.

## Current state

Only two packages exist today:

- `cmd/snapback` — the CLI entry point.
- `internal/version` — build version information.

Everything below is planned and not yet implemented.

## Planned modules

From [SPEC.md](SPEC.md) section 4:

| Module | Responsibility |
| --- | --- |
| `config` | Parse, validate, migrate and atomically persist configuration. |
| `provider` | The `SnapshotProvider` interface. |
| `provider/restic` | Restic implementation: subprocess execution, mount supervision, `backup`, `ls`. |
| `resolver` | Root and per-directory snapshot selection, tree-path mapping, timestamp aliases. |
| `projection` | Pure catalog model independent of any FUSE library. |
| `mount` | OS FUSE adapter for the history catalog. |
| `links` | Safe creation, registry, repair and removal of managed symlinks. |
| `discovery/seed` | Targeted path seeding and a watcher for new directories. |
| `discovery/onaccess` | On-access hooks (fanotify on Linux, Endpoint Security on macOS). |
| `discovery/explicit` | Shell hook events, `link`/`open` commands, Finder Sync events. |
| `prewarm` | Metadata pre-warm of the newest selected snapshots. |
| `daemon` | Process lifecycle, instance lock, local IPC, refresh, health. |
| `web` | Embedded assets and the authenticated local HTTP API. |
| `service` | launchd, systemd and OpenRC adapters. |
| `macos/FinderCompanion` | Swift app and Finder Sync extension. |

## Provider seam

All snapshot access goes through the `SnapshotProvider` interface (SPEC.md section 5): validate a repository, list snapshots, mount read-only, resolve a snapshot root, probe a tree path, create an on-demand snapshot, pre-warm metadata, and stop the mount.

Restic is the only implementation. The seam exists so that a second implementation would be a new package with no changes to `resolver`, `projection`, `links`, `discovery`, `daemon` or `web`.
