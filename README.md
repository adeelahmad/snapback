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

<!-- TODO: restore GIF (a three-second `cp .snapshot/...` restore) goes here -->

snapback is being built to put a read-only `.snapshot` folder inside your directories, backed by the Restic snapshots you already have. This is the design goal, not something that works today (see [Status](#status)):

```console
$ ls .snapshot/
2026-09-21_0300Z/  2026-09-22_0300Z/  latest/

$ cp .snapshot/latest/report.docx .
```

That is the whole restore: no app to open, no browser tab, no restore wizard. Your backups, shown as ordinary read-only directories, right where the files live.

## Why

Backing up was never the problem; restoring is. Most tools make you leave the directory you are working in, open something else, find the file again and download it somewhere. Older setups solved this with a `.snapshot` directory next to the data. snapback aims to bring that back for the Restic backups you already have.

- **Next to the data.** The plan is a `.snapshot` entry inside the directories your backups cover, not a separate mount you have to go looking for.
- **Any program.** `cp`, `diff`, an editor, a file manager, a file dialog: if it can read a directory, it will be able to restore.
- **Your existing backups.** snapback reads [Restic](https://restic.net) repositories. It does not replace your backup tool; it is meant to make restoring from it trivial.

## Install

```sh
curl -fsSL https://snapback.run/install.sh | sh
```

The installer downloads the latest tagged build for Linux or macOS from the [releases page](https://github.com/adeelahmad/snapback/releases). Today that build does one thing: `snapback version` prints its build information.

The planned `.snapshot` view will need the `restic` CLI and FUSE (`fuse3` on Linux, [macFUSE](https://macfuse.github.io) on macOS).

## How it works

This section describes the design in [SPEC.md](SPEC.md); none of it is built yet.

```
~/Documents/
├── report.docx                 ← today's file, never moved
└── .snapshot                   ← one managed symlink, the only change to the directory
    ├── 2026-09-21_0300Z/report.docx
    ├── 2026-09-22_0300Z/report.docx
    └── latest/                 ← the newest snapshot for this directory
```

Your live files will stay exactly where they are, on their own filesystem. The only change snapback will make to a live directory is one managed `.snapshot` symlink. It will point into a small read-only FUSE catalog that lists the Restic snapshots containing that directory, newest first, plus a `latest` alias, and resolves them through a private, read-only `restic mount`. Files will be read from the repository only when you open them.

## What it doesn't do

- It will not schedule backups, prune or forget snapshots. Keep using Restic for that.
- It will not write to your Restic repository or replace your backup tool.
- It will not run on Windows. FUSE is a hard requirement.
- It only reads Restic repositories.

## Status

snapback is early, pre-release software and not yet usable for restores. The only command that exists is `snapback version`; the `.snapshot` view described above is the design goal. Expect breaking changes.

Stage 1 compatibility evidence (FUSE catalog, `restic mount` path templates, metadata fidelity, crawler safety and latency measurements) is in [docs/reports/stage1/](docs/reports/stage1/). Issues and pull requests are welcome; see [CONTRIBUTING.md](CONTRIBUTING.md).

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
