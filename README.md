<picture>
  <source media="(prefers-color-scheme: dark)" srcset="docs/brand/readme-banner-dark.png">
  <img alt="snapback: Time Machine-style restore for your Restic backups, in every directory" src="docs/brand/readme-banner-light.png" width="1200">
</picture>

<p align="center">
  <a href="https://github.com/adeelahmad/snapback/releases"><img alt="release" src="https://img.shields.io/github/v/release/adeelahmad/snapback?color=1e7e43&label=release"></a>
  <a href="https://github.com/adeelahmad/snapback/actions/workflows/ci.yml"><img alt="ci" src="https://img.shields.io/github/actions/workflow/status/adeelahmad/snapback/ci.yml?branch=master&label=ci"></a>
  <a href="LICENSE"><img alt="license" src="https://img.shields.io/github/license/adeelahmad/snapback?color=5f6368"></a>
  <a href="https://snapback.run/"><img alt="docs" src="https://img.shields.io/badge/docs-snapback.run-1a73e8"></a>
</p>

# snapback

**Restoring a file should be as easy as it was in 2008.**

snapback is a restore tool for your backups. It supports Restic today; other backends are on the [roadmap](#roadmap).

snapback puts a read-only `.snapshot` entry inside your directories, backed by the Restic snapshots you already have. Here is what you get in a directory once `snapback setup` has run:

```console
$ ls .snapshot/
2026-09-21_0300Z/  2026-09-22_0300Z/  latest/

$ cp .snapshot/latest/report.docx .
```

That is the whole restore: no app to open, no browser tab, no restore wizard. Your backups, shown as ordinary read-only directories, right where the files live. The v0.1 acceptance run on Linux (CI, fuse3) exercised this restore, with macOS (macFUSE) as supplementary evidence; the results are in [docs/reports/v0.1-acceptance.md](docs/reports/v0.1-acceptance.md).

## Why

Backing up was never the problem; restoring is. Most tools make you leave the directory you are working in, open something else, find the file again and download it somewhere. Older setups solved this with a `.snapshot` directory next to the data. snapback brings that back for the Restic backups you already have.

- **Next to the data.** A `.snapshot` entry sits inside the directories your backups cover, not in a separate mount you have to go looking for.
- **Any program.** `cp`, `diff`, an editor, a file manager, a file dialog: if it can read a directory, it can restore.
- **Your existing backups.** snapback reads [Restic](https://restic.net) repositories. It does not replace your backup tool; it makes restoring from it trivial.

## Install

```sh
curl -fsSL https://snapback.run/install.sh | sh
```

The installer downloads the latest release (v1.4.0) for your operating system and architecture, verifies it against the signed `checksums.txt`, and installs the binary to `/usr/local/bin`, or to `~/.local/bin` when that directory is not writable or not on your `PATH`.

You need FUSE (`fuse3` on Linux, [macFUSE](https://macfuse.github.io) on macOS), the `restic` CLI on your `PATH`, and an existing Restic repository holding at least one snapshot of the directory you want to browse.

Or build from source with Go:

```sh
go build ./cmd/snapback
```

Every binary is built with `CGO_ENABLED=0`, so the Linux builds are statically linked. The `linux/arm`, `linux/mips` and `linux/mipsle` archives carry an `_unverified` suffix: they are cross-compiled but never run on that hardware, and the macOS binaries' linkage and runtime are unverified too. All assets are on the [releases page](https://github.com/adeelahmad/snapback/releases).

### Install from GitHub Packages

```sh
docker pull ghcr.io/adeelahmad/snapback:latest
```

The image is published by the release workflow and first appears with the next tagged release. See the [Docker page](docs-site/docker.md) for what the container can and cannot do.

## Quick start

```sh
curl -fsSL https://snapback.run/install.sh | sh
cd ~/project && snapback setup
ls .snapshot
```

`snapback setup` reads the machine instead of asking you to describe it: it takes the repository from `RESTIC_REPOSITORY` and the password file from `RESTIC_PASSWORD_FILE`, takes the directory you ran it in as the backup root, asks the repository for the snapshot host and paths, prints each fact it found and writes the configuration for you. Restoring is then an ordinary copy:

```sh
cp .snapshot/latest/report.docx .
```

Three commands on Linux, four on macOS — measured, see [docs/reports/sprint5/actions.md](docs/reports/sprint5/actions.md).

On macOS, setup writes the configuration only: start the daemon yourself with `snapback run`, which stays in the foreground, because the login service is Linux-only for now. Every flag and command is in the [usage guide](docs-site/usage.md).

## How it works

The full design is in [SPEC.md](SPEC.md).

```
~/Documents/
├── report.docx                 ← today's file, never moved
└── .snapshot                   ← one managed symlink, the only change to the directory
    ├── 2026-09-21_0300Z/report.docx
    ├── 2026-09-22_0300Z/report.docx
    └── latest/                 ← the newest snapshot for this directory
```

Your live files stay exactly where they are, on their own filesystem. The only change snapback makes to a live directory is one managed `.snapshot` symlink. It points into a small read-only FUSE catalog that lists the Restic snapshots containing that directory, newest first, plus a `latest` alias, and resolves them through a private, read-only `restic mount`. Files are read from the repository only when you open them.

While the daemon runs, a new snapshot can take up to about a minute to appear under `.snapshot`, because the view reloads Restic snapshot metadata on a refresh interval. `snapback refresh` asks the daemon to reload sooner.

## What it doesn't do

- It will not schedule backups, prune or forget snapshots. Keep using Restic for that.
- It will not write to your Restic repository or replace your backup tool.
- It will not run on Windows. FUSE is a hard requirement.
- It only reads Restic repositories.

## Status

snapback v0.1 is early, pre-release software. Expect breaking changes. Linux is the v0.1 target.

What exists and passed the v0.1 acceptance run on Linux (CI, fuse3), with macOS (macFUSE) as supplementary evidence: the `.snapshot` view with its `latest` alias, `snapback snap` for ad-hoc snapshots, `snapback seed` and `snapback link` to create `.snapshot` entries, the daemon (`snapback run`), the local web UI (`snapback web`) and `snapback doctor`. The systemd user service (`snapback install service`) passed on Linux too; Acc 2, 12 and 17 are Linux-only and were skipped in the macOS runs. Item-by-item results are in [docs/reports/v0.1-acceptance.md](docs/reports/v0.1-acceptance.md).

Stage 1 compatibility evidence (FUSE catalog, `restic mount` path templates, metadata fidelity, crawler safety and latency measurements) is in [docs/reports/stage1/](docs/reports/stage1/). Issues and pull requests are welcome; see [CONTRIBUTING.md](CONTRIBUTING.md).

## Roadmap

snapback supports Restic today. These backends are planned, not supported yet:

- Borg: planned.
- Kopia: planned.
- ZFS snapshots: planned.
- Btrfs snapshots: planned.

## Prior art and credit

snapback stands on the shoulders of [httm](https://github.com/kimono-koans/httm) by kimono-koans, released under the MPL-2.0 license. httm showed how pleasant it is to browse and restore past versions of a file right from where it lives, and snapback's in-directory `.snapshot` idea owes a great deal to it.

## Documentation

- Docs site: https://snapback.run/docs/
- [SPEC.md](SPEC.md): the product specification.
- [ARCHITECTURE.md](ARCHITECTURE.md): how the code is organised.
- [CONTRIBUTING.md](CONTRIBUTING.md): how to build, test and send changes.
- [SECURITY.md](SECURITY.md): how to report a vulnerability.

## License

[MIT](LICENSE)

<p align="center"><img src="docs/brand/mark.svg" width="24" alt=""></p>
