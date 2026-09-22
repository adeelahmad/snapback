# Using Snapback

Snapback v0.1 is early, pre-release software. This page describes only the commands that exist in
the current build. Restic is the only backend.

## Getting started

1. **Install.** Run the install script from the README, or build from source with
   `go build ./cmd/snapback`.
2. **Configure.** Write `~/.config/snapback/config.yaml` with your Restic repository, then run
   `snapback config`. It starts the local web UI and opens its setup page. `snapback web` serves
   the same UI without jumping to setup.
3. **Install the service.** Run `snapback install service`. The default scope is the current user.
4. **Check.** Run `snapback doctor` to check prerequisites and repository health.

## Command reference

These are the commands `snapback --help` lists, with their own summaries.

| Command | Summary |
|---|---|
| `snapback config` | validate or show the configuration |
| `snapback doctor` | check prerequisites and repository health |
| `snapback install` | install the Snapback background service |
| `snapback link` | create the .snapshot link in a directory |
| `snapback links` | list, repair or remove managed .snapshot links |
| `snapback notify` | (no summary in help) |
| `snapback open` | open a directory's .snapshot history in the file manager |
| `snapback refresh` | ask the running daemon to refresh its snapshot view |
| `snapback run` | run the daemon in the foreground |
| `snapback seed` | pre-create .snapshot links under a directory or the configured roots |
| `snapback service` | start, stop, restart, inspect or uninstall the service |
| `snapback shell-hook` | (no summary in help) |
| `snapback snap` | take an ad-hoc snapshot of a directory now |
| `snapback status` | print the running daemon's status |
| `snapback version` | print build information |
| `snapback web` | serve the local web UI |

## Web UI

`snapback web` refuses any listen address that is not loopback; the default is `127.0.0.1` on a
random port. Each run creates a one-time token and prints a URL of the form
`http://127.0.0.1:PORT/auth?token=...`. Opening that URL exchanges the token for a session.
`--open` also opens the URL in a browser when a desktop session is present.

## How `.snapshot` looks

Each registered directory gets a read-only `.snapshot` entry containing:

- `latest`: a link to the newest snapshot of that directory.
- Timestamp-named links, one per snapshot, for browsing by time.
- `by-date/`: one folder per day, holding that day's timestamp links.
- `snapshots/`: one entry per Restic snapshot, named by its full snapshot ID.
- `info.json`: metadata about the directory and its snapshots.

Snapback never writes to the Restic repository.

## Latency

After a backup finishes, a new snapshot can take up to about a minute to appear under `.snapshot`.
`snapback refresh` asks the running daemon to refresh sooner.

## Limits

- Linux first. Release testing covers Linux.
- macOS: the release is a self-contained executable, but its runtime is not verified.
- The service installer is tested with systemd in user scope only.
