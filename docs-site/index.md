# Snapback

Snapback is about one thing: getting a file back. Restoring a file should be as easy as it was in 2008 — pick the moment, pick the file, restore it. Snapback is a restore tool for your backups. It supports Restic today, which stores the snapshots.

## Install

Coming soon. There are no builds to download yet.

## Status

Snapback is pre-release software under active development. Expect breaking changes.

## Roadmap

These backends are planned, not supported yet:

- Borg: planned.
- Kopia: planned.
- ZFS snapshots: planned.
- Btrfs snapshots: planned.

When the daemon runs, a snapshot taken while it runs can take up to about a minute to appear under `.snapshot`, because the view reloads Restic snapshot metadata on a refresh interval.

## Versions

Versioned documentation is not yet available. Docs will be published per release once the first release ships.
