# Using Snapback

Snapback v0.1 is early, pre-release software. This page describes only the commands that exist in
the current build. Restic is the only backend.

## Getting started

1. **Build.** Build from source with `go build ./cmd/snapback`. The install script fetches the
   latest tagged release, which predates v0.1.
2. **Configure.** Write `~/.config/snapback/config.yaml` with your Restic repository, then run
   `snapback config`. It starts the local web UI and opens its setup page. `snapback web` serves
   the same UI without jumping to setup. See [Configuration](configuration.md) for every key,
   `snapback config --file` and `snapback config validate`.
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

## Machine-readable status: `status --json`

`snapback status --json` prints one JSON object on stdout. It is an envelope: `ok` reports whether
the daemon answered, and `data` carries the status snapshot. Every key is snake_case.

```json
{
  "ok": true,
  "data": {
    "state": "degraded",
    "repos": [
      { "id": "personal", "state": "ready", "code": "" },
      { "id": "archive", "state": "failed", "code": "repository_unavailable" }
    ],
    "last_refresh": "2026-09-22T06:00:00Z",
    "generation": 7,
    "eligible_count": { "personal": 3, "archive": 0 },
    "links": 4,
    "warm": {
      "5f2d9c1b8a7e4630df51cc2a9b04e7f318d6a5be2c7f091d43ab68e5c70f29d1": true
    },
    "prewarm": {
      "warm": 1,
      "cold": 2,
      "pending": 0,
      "last_prewarm": "2026-09-22T06:00:00Z"
    },
    "pending": [],
    "discovery": "running",
    "throttle": [
      { "pid": 4821, "process": "mds", "rule": "deny", "at": "2026-09-22T05:58:12Z" }
    ],
    "web_url": "http://127.0.0.1:8080/",
    "recovery": { "unmounted": [], "foreign": [] }
  }
}
```

The output carries no repository passwords or environment values.

## Daemon logs

The daemon writes its log to stderr. It logs one `Info` line whenever the refresh outcome changes
— on the first refresh, and after that only when the generation, a repository's state, the
eligible count or the failed repositories change, so an idle daemon stays quiet:

```
refresh generation=7 personal=ready archive=failed eligible=3/4 failed=archive
```

`eligible` is the number of linked directories that have at least one snapshot, over the number of
linked directories; `failed` appears only when a repository failed. When a repository's mount
fails, the daemon logs one `Error` line carrying `repos=<sorted ids>` and `err=<error>`. Under the
systemd user service, read both with:

```
journalctl --user -u snapback.service
```

## Web UI

`snapback web` refuses any listen address that is not loopback; the default is `127.0.0.1` on a
random port. Each run creates a one-time token and prints a URL of the form
`http://127.0.0.1:PORT/auth?token=...`. Opening that URL exchanges the token for a session.
`--open` also opens the URL in a browser when a desktop session is present. The
[web UI guide](web-ui.md) walks through each page with screenshots.

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

- Linux is the v0.1 target. Linux acceptance evidence is pending.
- The v0.1 acceptance run so far is on macOS with macFUSE; every applicable item passed there,
  and Acc 2, 12 and 17 need Linux. See
  `docs/reports/v0.1-acceptance.md` in the repository.
- The service installer is tested with systemd in user scope only.
