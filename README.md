# Snapback

<!-- TODO: restore GIF (a three-second `cp .snapshot/...` restore) goes here -->

**Time Machine-style restore for your Restic backups, in every directory.**

Backing up was never the problem; restoring is. Restoring a file should be as easy as it was in 2008: the history of a directory sits right next to it, and getting yesterday's copy back is one ordinary copy command.

Snapback puts a read-only `.snapshot` entry inside your directories. Any program (a shell, a file manager, a file dialog, a script) can browse it and see the directory as it was in past Restic snapshots:

```sh
cp .snapshot/2026-09-20_0300/report.docx .
```

Your live files never move. The only change to a live directory is one managed `.snapshot` symlink; everything behind it is virtual and read-only.

## Status

Snapback is pre-release. It is under active development, there is no tagged release yet, and the restore workflow described above is the design goal, not something you can rely on today. Expect breaking changes.

## Install

Once the first release is published, the one-line installer will be:

```sh
curl -fsSL https://snapback.example.com/install.sh | sh
```

The domain above is a placeholder; the real install URL will be announced with the first release. Snapback targets Linux and macOS, requires the `restic` CLI and FUSE (macFUSE on macOS), and `snapback doctor` will explain anything missing.

The planned onboarding is three commands: `snapback config`, `snapback web` and `snapback install service`.

## Prior art and credit

Snapback stands on the shoulders of [httm](https://github.com/kimono-koans/httm) by kimono-koans, released under the MPL-2.0 license. httm showed how pleasant it is to browse and restore past versions of a file right from where it lives, and Snapback's in-directory `.snapshot` idea owes a great deal to it.

## Documentation

- [SPEC.md](SPEC.md): the product specification.
- [ARCHITECTURE.md](ARCHITECTURE.md): how the code is organised.
- [CONTRIBUTING.md](CONTRIBUTING.md): how to build, test and send changes.
- [SECURITY.md](SECURITY.md): how to report a vulnerability.
- Docs site: https://adeelahmad.github.io/snapback/
