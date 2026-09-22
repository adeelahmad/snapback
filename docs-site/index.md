# Snapback

Snapback is about one thing: getting a file back. Restoring a file should be as easy as it was in 2008 — pick the moment, pick the file, restore it. Snapback is a restore tool for your backups. It supports Restic today, which stores the snapshots.

## Install

```sh
curl -fsSL https://snapback.run/install.sh | sh
```

The installer downloads the latest release (v1.4.0) for your operating system and architecture, verifies it against the signed `checksums.txt`, and installs the binary to `/usr/local/bin`, or to `~/.local/bin` when that directory is not writable or not on your `PATH`.

Before running it you need FUSE (`fuse3` on Linux, [macFUSE](https://macfuse.github.io) on macOS), the `restic` CLI and an existing Restic repository. You can also build from source with `go build ./cmd/snapback`. The [usage guide](usage.md) covers setup, the service and every command.

### Install from GitHub Packages

```sh
docker pull ghcr.io/adeelahmad/snapback:latest
```

The image is published by the release workflow and first appears with the next tagged release. See the [Docker page](docker.md) for what the container can and cannot do.

## How it works

Each directory your backups cover gets a read-only `.snapshot` entry. It lists the Restic snapshots that contain the directory, newest first, plus a `latest` alias. Restore with any program that reads files, for example `cp .snapshot/latest/report.docx .`. Browsing and restoring never write to the Restic repository, and every read runs `restic … --no-lock`; `snapback snap` is the one command that adds a snapshot, and only when you run it. Nothing in Snapback deletes, prunes or rewrites repository data.

While the daemon runs, a new snapshot can take up to about a minute to appear under `.snapshot`, because the view reloads Restic snapshot metadata on a refresh interval.

## Status

Snapback v0.1 is early, pre-release software. Expect breaking changes. Linux is the v0.1 target.

The `.snapshot` view, `latest`, `snap`, `seed`, `link`, the daemon, the web UI, `doctor` and the systemd user service passed the v0.1 acceptance run on Linux (CI, fuse3) — Acc 1 to 17 — with macOS (macFUSE) as supplementary evidence. Acc 2, 12 and 17 are Linux-only and were skipped in the macOS runs. Item-by-item results are in the [v0.1 acceptance report](https://github.com/adeelahmad/snapback/blob/master/docs/reports/v0.1-acceptance.md).

## Roadmap

These backends are planned, not supported yet:

- Borg: planned.
- Kopia: planned.
- ZFS snapshots: planned.
- Btrfs snapshots: planned.

## Versions

This site is built from the `master` branch by the docs workflow (`.github/workflows/docs.yml`) and describes the current release, v1.4.0; per-release documentation is not published. Released versions are listed on the [GitHub releases page](https://github.com/adeelahmad/snapback/releases).

## Privacy

This site uses Google Analytics to count visits.
