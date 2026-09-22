# Web UI

Snapback includes a small local web UI for status, history, configuration and setup. It
serves only the pages described here.

## Opening it

- `snapback web` starts the UI. Add `--open` to also open it in a browser when a desktop
  session is present.
- `snapback config` starts the same UI and goes to the Setup page.

The server listens on loopback only and refuses any other listen address. The default is
`127.0.0.1` on a random port. Each run creates a one-time token and prints a URL of the form
`http://127.0.0.1:PORT/auth?token=...`. Opening it exchanges the token for a session cookie
and redirects to `/`, so the token leaves the address bar. The token works once. The same
URL is written to `web.url` in the state directory (mode 0600) and removed when the server
stops.

Every page has the same navigation bar: History, Status, Config, Integrations and Setup.

## Status

![Status page listing the mount "demo: mounted" and metrics for last refresh, eligible snapshots, managed links, throttle events, prewarm and discovery mode](img/webui-status.png)

The Status page shows the daemon's mounts with their state, then these metrics. A metric
the daemon does not report shows "not reported".

| Field | Meaning |
|---|---|
| Last refresh | when the daemon last refreshed its snapshot view |
| Eligible snapshots | how many Restic snapshots match the configured roots and filters |
| Managed links | how many `.snapshot` links Snapback manages |
| Throttle events | how many reader throttle events the daemon has recorded |
| Prewarm | snapshots counted as warm, cold and pending |
| Discovery mode | how `.snapshot` links are created, for example `seed` |

## History

![History page with the root "home", three snapshots listed newest first and the files of the selected snapshot](img/webui-history.png)

1. Pick a root in the **Roots** panel. Each root shows its path and the repository and
   mount state.
2. Pick a folder. The folders under the root that have a `.snapshot` link are listed.
3. The **Snapshots** timeline lists the snapshots of that folder newest first, with the
   snapshot name, host and time. Each is marked warm or cold.
4. Pick a snapshot to see its files: name, size, modified time and state. A file that is
   missing in that snapshot shows "absent", and one that could not be read shows
   "read failed".

When a file is selected, a **Versions** panel lists its versions. Versions with the same
size and modified time are marked "likely identical". Each version has two actions.

- **Download** streams that version of the file to your browser.
- **Restore copy next to original** copies that version into the live folder beside the
  original. The copy is named with the snapshot's date, for example `plan (2026-09-19).txt`.
  If that name is taken, a number is added, as in `plan (2026-09-19) 2.txt`. An existing
  file is never overwritten, and the copy is created with mode 0600.

**Open in file manager** opens the folder in the host file manager.

## Configuration

![Configuration page with fields for roots, filters, exclusions, seed paths, discovery mode, cache directory and refresh interval](img/webui-config.png)

The Configuration page edits the roots, filters, exclusions and seed paths (one per line),
the discovery mode, the cache directory and the refresh interval, filled from the current
configuration. In this build the **Save configuration** button posts to `/config`, which has
no save handler, so it does not write anything. Edit the configuration file instead; see
[Configuration](configuration.md) for every key.

## Setup

![Setup page with fields for the Restic binary, the Rclone binary, the repository, the password file and the roots](img/webui-setup.png)

The Setup page picks the Restic binary and, if one is used, the Rclone binary, then takes the
repository, the password file and the roots. In this build the **Save setup** button posts
to `/setup`, which has no save handler, so it does not write anything. Write the
configuration file as described in [Configuration](configuration.md).

## Security

- **Loopback only.** The server binds only to a loopback address.
- **Host check.** A request whose `Host` is not the bound loopback address and port is
  refused with 421.
- **Origin check.** A write (any method other than GET, HEAD or OPTIONS) with an `Origin`
  other than its own host, or a `Sec-Fetch-Site` other than `same-origin` or `none`, is
  refused with 403.
- **Session.** Every page and API route needs the session cookie from the one-time token.
  The cookie is `HttpOnly` and `SameSite=Strict`.
- **CSRF.** Every write must carry the session's CSRF token, or it is refused with 403.
- **Paths.** The root must be one of the configured roots, and a path must be local to it.
  Absolute paths, `..` and NUL bytes are refused.
- **No arbitrary file read.** Download and restore open files through the snapshot
  directory with `os.OpenRoot`, so `..` and symlinks cannot leave the snapshot, and only
  regular files are served. Restore writes through the live root the same way.
- **Headers.** Responses set `X-Frame-Options: DENY`, `X-Content-Type-Options: nosniff`
  and `Referrer-Policy: no-referrer`.
