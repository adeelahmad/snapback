# Snapback

Snapback is about one thing: getting a file back. Restoring a file should be as easy as it was in 2008 — pick the moment, pick the file, restore it. Snapback is a restore tool for your backups. It supports Restic today, which stores the snapshots.

## Install

The latest tagged release predates v0.1. Until v0.1 is tagged, build from source:

```sh
go build ./cmd/snapback
```

Snapback needs the `restic` CLI and FUSE (`fuse3` on Linux, macFUSE on macOS). The [usage guide](usage.md) covers setup, the service and every command.

## How it works

Each directory your backups cover gets a read-only `.snapshot` entry. It lists the Restic snapshots that contain the directory, newest first, plus a `latest` alias. Restore with any program that reads files, for example `cp .snapshot/latest/report.docx .`. Snapback never writes to the Restic repository.

While the daemon runs, a new snapshot can take up to about a minute to appear under `.snapshot`, because the view reloads Restic snapshot metadata on a refresh interval.

## Status

Snapback v0.1 is early, pre-release software. Expect breaking changes. Linux is the v0.1 target.

The `.snapshot` view, `latest`, `snap`, `seed`, `link`, the daemon, the web UI and `doctor` passed the v0.1 acceptance run on macOS with macFUSE; every applicable item passed there, and Acc 2, 12 and 17 need Linux. The systemd user service is built, but its acceptance check needs Linux. Linux acceptance evidence is pending.

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
