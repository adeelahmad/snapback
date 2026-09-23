# CLI reference

Every `snapback` subcommand, its synopsis and its flags. The page is written
from the binary's own help output, and `test/site/cli_flags_test.go` builds the
binary and fails if a flag here is missing or invented, so the reference cannot
drift from the code.

Run `snapback` with no arguments to list the commands, or `snapback <command> -h`
for the same synopsis you see below. Flags accept one or two dashes: `-json` and
`--json` are the same flag.

## snapback setup

Detect this machine and write a working configuration.

Setup also picks a mount point for the repository: a directory whose `.snapshot`
shows the whole repository history — every snapshot, every host — so you can
restore onto a machine that was not the one backed up. The default is
`/mnt/<repository id>` (on macOS,
`~/Library/Application Support/snapback/mounts/<repository id>`); interactive
setup asks for it, `--no-prompt` keeps the default, and `--mount ""` leaves the
repository unmounted. The daemon links it once the repository is ready.

```
snapback setup [flags] [ROOT...]
```

Args:

- `ROOT` — a directory to back up; defaults to the working directory.

Example:

```
snapback setup --repo /srv/restic ~/work
```

| Flag | Description |
|---|---|
| `--dry-run` | report the configuration without writing it |
| `--force` | overwrite an existing configuration |
| `--mount` | mount point for this repository's restores (default `/mnt/<repository id>`); empty disables it |
| `--no-prompt` | do not ask any question, keep every default |
| `--no-service` | do not install a background service |
| `--password-file` | repository password file to use instead of the detected one |
| `--repo` | Restic repository to use instead of the detected one |

## snapback config

Validate or show the configuration.

```
snapback config [flags] validate|show|path
```

Args:

- `validate` — report whether the configuration is valid.
- `show` — print the configuration with secrets redacted.
- `path` — print the resolved configuration path.

Example:

```
snapback config show --json
```

| Flag | Description |
|---|---|
| `--json` | write a JSON envelope |

## snapback link

Create the `.snapshot` link in a directory.

```
snapback link [flags] DIR
```

Args:

- `DIR` — the directory to give a `.snapshot` entry.

Example:

```
snapback link ~/work
```

| Flag | Description |
|---|---|
| `--json` | write a JSON envelope |

## snapback links

List, repair or remove managed `.snapshot` links.

```
snapback links <list|repair|remove> [flags]
```

Args:

- `list` — print the managed `.snapshot` links.
- `repair` — restore the managed links that are missing or stale.
- `remove` — remove the managed links; requires `--managed`.

Example:

```
snapback links remove --managed --json
```

| Flag | Description |
|---|---|
| `--json` | write a JSON envelope |
| `--managed` | with remove, remove every registry-owned link |

## snapback seed

Pre-create `.snapshot` links under a directory or the configured roots.

```
snapback seed [flags] DIR
```

Args:

- `DIR` — the directory to seed; defaults to the configured roots.

Example:

```
snapback seed --max-depth 2 ~/work
```

| Flag | Description |
|---|---|
| `--dry-run` | report the plan without creating links |
| `--force` | seed even when the inode budget is exceeded |
| `--json` | write a JSON envelope |
| `--max-depth` | how deep below DIR to seed (default 3) |

## snapback open

Open a directory's `.snapshot` history in the file manager.

```
snapback open [flags] DIR
```

Args:

- `DIR` — the directory whose `.snapshot` history to open.

Example:

```
snapback open --json ~/work
```

| Flag | Description |
|---|---|
| `--json` | write a JSON envelope instead of a message |

## snapback snap

Take an ad-hoc snapshot of a directory now.

```
snapback snap [flags] [DIR]
```

Args:

- `DIR` — the directory to snapshot; defaults to the working directory.

Example:

```
snapback snap --tag release --wait --timeout 5m ~/work/proj
```

| Flag | Description |
|---|---|
| `--json` | write a JSON envelope |
| `--prefix` | tree prefix for a path outside the roots |
| `--repository` | repository ID for a path outside the roots |
| `--tag` | add a tag |
| `--timeout` | how long `--wait` waits (default 2m0s) |
| `--wait` | wait until the snapshot is browsable |

## snapback run

Run the daemon in the foreground.

```
snapback run [flags]
```

Example:

```
snapback run --config /etc/snapback/config.yaml
```

| Flag | Description |
|---|---|
| `--config` | configuration file (defaults to the resolved user config path, `~/.config/snapback/config.yaml`) |
| `--log-file` | write logs to this file; overrides `logging.file` |
| `--log-format` | log format, one of text, json; overrides `logging.format` (default text) |
| `--log-level` | log level, one of debug, info, warn, error; overrides `logging.level` (default info) |

## snapback status

Print the running daemon's status.

```
snapback status [flags]
```

Example:

```
snapback status --json
```

| Flag | Description |
|---|---|
| `--json` | write a JSON envelope |

## snapback refresh

Ask the running daemon to refresh its snapshot view.

```
snapback refresh [flags]
```

Example:

```
snapback refresh --json
```

| Flag | Description |
|---|---|
| `--json` | write a JSON envelope |

## snapback doctor

Check prerequisites and repository health.

```
snapback doctor [flags] [DIR]
```

Args:

- `DIR` is where `--bundle` writes the archive; it defaults to the current directory.

`doctor --bundle` is unaffected by telemetry settings: it always writes the bundle,
whether telemetry is enabled or disabled.

Example:

```
snapback doctor --json --strict
```

| Flag | Description |
|---|---|
| `--bundle` | write a local diagnostic bundle into DIR and print its path |
| `--json` | print checks as a JSON array |
| `--mount-test` | add a mount_test check that performs a real mount |
| `--strict` | keep platform-inapplicable checks as failures |
| `--verbose` | print the probe and the raw observation under every check |

## snapback web

Serve the local web UI.

```
snapback web [flags]
```

Args:

- `web` takes no positional arguments.

Example:

```
snapback web
```

| Flag | Description |
|---|---|
| `--allow-origin` | also accept browser requests from this origin (repeatable) |
| `--allow-remote` | allow a bind address that is not loopback; prints a warning |
| `--assets` | load templates and assets from DIR |
| `--bind` | listen on ADDRESS instead of the configured address; a non-loopback address is refused without `--allow-remote` |
| `--log-file` | write logs to this file; overrides `logging.file` |
| `--log-format` | log format, one of text, json; overrides `logging.format` (default text) |
| `--log-level` | log level, one of debug, info, warn, error; overrides `logging.level` (default info) |
| `--open` | open the web UI in a browser |
| `--with-daemon` | run a snapback daemon for the lifetime of this command |

## snapback install service

Install the Snapback background service. `snapback install` accepts only this
form.

```
snapback install service [flags]
```

Args:

- `install service` takes no positional arguments beyond the word "service".

Example:

```
snapback install service --scope user --manager systemd --user alice
```

| Flag | Description |
|---|---|
| `--json` | write a JSON envelope |
| `--manager` | service manager: auto, systemd, launchd or openrc (default "auto") |
| `--scope` | service scope: user or system (default "user") |
| `--user` | run a system-scope service as this user |

## snapback service

Start, stop, restart, status or uninstall the service.

```
snapback service <start|stop|restart|status|uninstall> [flags]
```

Args:

- `start|stop|restart|status|uninstall` — the lifecycle action to take.

Example:

```
snapback service status --scope user
```

| Flag | Description |
|---|---|
| `--scope` | service scope: user or system (default "user") |

## snapback shell-hook

Print the shell integration script for a shell. It takes no flags.

```
snapback shell-hook bash|zsh|fish
```

Args:

- `bash|zsh|fish` — the shell to emit the hook script for.

Example:

```
eval "$(snapback shell-hook zsh)"
```

## snapback telemetry

Report on, enable or disable telemetry.

```
snapback telemetry <status|show|enable|disable> [flags]
```

Args:

- `status` — report whether telemetry is enabled
- `show` — print the telemetry settings
- `enable` — turn telemetry on
- `disable` — turn telemetry off

Example:

```
snapback telemetry status
```

See https://snapback.run/privacy for what telemetry collects and why.

| Flag | Description |
|---|---|
| `--json` | write a JSON envelope |

## snapback notify

Internal command invoked by the shell hook to tell the running daemon which
directory you just entered. It prints nothing, never fails a shell prompt and
is not meant to be run by hand.

```
snapback notify [options] -- DIR
```

Args:

- `DIR` — the directory the shell moved into.

## snapback version

Print build information. It takes no flags.

```
snapback version
```

Example:

```
snapback version
```
