# Configuration

Snapback reads one YAML file. The CLI and the web UI write that same file through the same
validator, so a file one accepts, the other accepts too.

## Where the file lives

Print the path Snapback uses:

```sh
snapback config path          # plain path
snapback config path --json   # {"path": "..."}
```

`config path` only prints the path. It never reads or creates the file.

The default path is the same on Linux and macOS:

| Condition | Path |
|---|---|
| `XDG_CONFIG_HOME` is set to an absolute path | `$XDG_CONFIG_HOME/snapback/config.yaml` |
| Otherwise | `$HOME/.config/snapback/config.yaml` |

To use another file, put `--config PATH` before the command, for example
`snapback --config /etc/snapback/config.yaml config validate`.

## First run

```sh
snapback config
```

With no subcommand, `snapback config` starts the local web UI and opens its **Setup** screen in
your browser. If no config file exists yet, the UI starts from the built-in defaults, so the
command works before you have written anything. With no desktop session, it prints the URL for
you to open yourself. The server runs until you stop it with Ctrl-C.

The Setup screen asks for:

- **Restic binary**: the `restic` executable.
- **Rclone binary**: the `rclone` executable, needed only for `rclone:` repositories. Choose
  "None" otherwise.
- **Repository**: the Restic repository, such as `/srv/restic` or `rclone:gdrive:Backups/restic`.
- **Password file**: the file that holds the repository password.
- **Roots**: the local directories to show history for, one per line.

The web UI saves the configuration through the full validator and a revision check. It saves
only when both pass. An invalid configuration is rejected with its field errors, and the file on
disk stays as it was. If the file has changed since the page loaded it, the save is refused, so
it never overwrites a newer version.

## Non-interactive setup

```sh
snapback config --file ./config.yaml
```

`--file` loads and validates the given file, then saves it to the config path. That is the
default path, or the `--config` override. It does not start the web UI. If validation fails, it
prints the field errors and exits non-zero, and any existing config stays untouched. This also
works when the existing config no longer parses.

The file is written atomically with mode `0600`. If its directory does not exist, Snapback
creates it with mode `0700`.

## Checking and viewing

```sh
snapback config validate          # prints "configuration valid" or the field errors
snapback config validate --json   # the same result as a JSON envelope
snapback config show              # prints the effective configuration as YAML
```

`validate` and `show` both load the file, apply defaults and run every check. `show` prints the
result with the defaults filled in. It replaces every repository `environment` value with
`***`, so the output is safe to share.

## Complete annotated example

Every key below is part of the schema. Unknown keys are rejected. If you leave a key out, it
takes the default listed in the field reference.

```yaml
version: 1                    # required; must be 1

link_name: .snapshot          # name of the history entry in each directory
timestamps: utc               # utc | local

state_dir: /home/alex/.local/state/snapback
history_mount: /home/alex/.local/state/snapback/mounts/history
backend_mount_dir: /home/alex/.local/state/snapback/mounts/repositories

web:
  enabled: true
  listen: 127.0.0.1:0         # loopback only; port 0 picks a free port
  open_browser: false
  assets_dir: ""              # load UI templates and assets from a directory instead

catalog:
  refresh_interval: 60s       # how often snapshot lists are refreshed
  prewarm_snapshots: 2
  prewarm_concurrency: 2
  probe_concurrency: 4
  presence_cache_entries: 10000
  presence_cache_ttl: 5m
  reader_policy:
    deny_processes: [rg, fd, find, rsync, mdworker, mds]
    burst_limit: 50

views:
  rsnapshot: false
  rsnapshot_keep: {hourly: 0, daily: 7, weekly: 4, monthly: 6}

discovery:
  mode: seed                  # only seed is available in v0.1
  shell: true
  finder: false               # accepted, but has no effect in v0.1
  seed:
    inode_threshold: 0.90
    max_links_per_path: 500000
  on_access:
    allow_processes: []
    handler_timeout: 20ms

repositories:
  - id: personal
    repository: rclone:gdrive:Backups/restic
    restic_binary: /usr/bin/restic
    rclone_binary: /usr/bin/rclone                     # required for rclone: repositories
    password_file: /home/alex/.config/restic/password  # mode 0600 or 0400
    cache_dir: /home/alex/.cache/restic
    no_cache: false
    lock_mode: normal                                  # normal | none
    environment:
      RCLONE_CONFIG: /home/alex/.config/rclone/rclone.conf

roots:
  - id: work
    local_path: /home/alex/work
    repository_id: personal
    prefix_map:
      - hostname: alex-laptop
        source_path: /home/alex/work
        tree_prefix: /home/alex/work
    snapshots:
      hostname: ""            # empty means all hosts
      tags_all: []
      source_paths_exact: []
    seed_paths:
      - path: project
        max_depth: 6
    exclude_relative_paths:
      - .git
      - node_modules
    snap:
      tags: [snapback:adhoc]

service:
  manager: auto               # auto | systemd
  scope: user                 # user | system
  run_as_user: ""
```

Durations use Go syntax: `20ms`, `60s`, `5m`, `1h`.

## Field reference

### Top level

| Key | Type | Default | Meaning |
|---|---|---|---|
| `version` | integer | none (required) | Schema version. Must be `1`. |
| `link_name` | string | `.snapshot` | Name of the history entry. Any file name except `.` and `..`, with no `/`. |
| `timestamps` | string | `utc` | `utc` or `local` for timestamp names. |
| `state_dir` | absolute path | `$XDG_STATE_HOME/snapback`, else `$HOME/.local/state/snapback` | Where Snapback keeps its state. |
| `history_mount` | absolute path | `<state_dir>/mounts/history` | Mount point of the history view. |
| `backend_mount_dir` | absolute path | `<state_dir>/mounts/repositories` | Directory for repository mounts. |

### `web`

| Key | Type | Default | Meaning |
|---|---|---|---|
| `web.enabled` | boolean | `true` | Whether the web UI is enabled. |
| `web.listen` | host:port | `127.0.0.1:0` | Listen address. The host must be `127.0.0.1`, `::1` or `localhost`. |
| `web.open_browser` | boolean | `false` | Whether to open a browser. |
| `web.assets_dir` | path | empty | Load UI templates and assets from this directory instead of the built-in ones. |

### `catalog`

| Key | Type | Default | Meaning |
|---|---|---|---|
| `catalog.refresh_interval` | duration | `60s` | How often snapshot lists are refreshed. Must be greater than 0. |
| `catalog.prewarm_snapshots` | integer | `2` | Number of newest snapshots to prewarm. At least 0. |
| `catalog.prewarm_concurrency` | integer | `2` | Parallel prewarm workers. At least 1. |
| `catalog.probe_concurrency` | integer | `4` | Parallel presence probes. At least 1. |
| `catalog.presence_cache_entries` | integer | `10000` | Size of the presence cache. At least 0. |
| `catalog.presence_cache_ttl` | duration | `5m` | How long presence results are cached. At least 0. |
| `catalog.reader_policy.deny_processes` | list of strings | `[rg, fd, find, rsync, mdworker, mds]` | Process names that are throttled when they enumerate snapshot directories. |
| `catalog.reader_policy.burst_limit` | integer | `50` | Number of snapshot directories a process may enumerate before it is throttled. At least 1. |

### `views`

| Key | Type | Default | Meaning |
|---|---|---|---|
| `views.rsnapshot` | boolean | `false` | Turn on the optional `daily.N`/`weekly.N` view names. |
| `views.rsnapshot_keep.hourly` | integer | `0` | Number of hourly entries to show. At least 0. |
| `views.rsnapshot_keep.daily` | integer | `7` | Number of daily entries to show. At least 0. |
| `views.rsnapshot_keep.weekly` | integer | `4` | Number of weekly entries to show. At least 0. |
| `views.rsnapshot_keep.monthly` | integer | `6` | Number of monthly entries to show. At least 0. |

### `discovery`

| Key | Type | Default | Meaning |
|---|---|---|---|
| `discovery.mode` | string | `seed` | How `.snapshot` entries appear. Only `seed` is available in v0.1, and `on-access` is rejected. |
| `discovery.shell` | boolean | `true` | Shell integration. |
| `discovery.finder` | boolean | `false` | Finder integration. Accepted, but it has no effect in v0.1: the Finder companion is planned (see ARCHITECTURE.md). |
| `discovery.seed.inode_threshold` | number | `0.90` | Inode usage above which seeding stops. Greater than 0 and at most 1. |
| `discovery.seed.max_links_per_path` | integer | `500000` | Maximum number of links seeded under one seed path. At least 1. |
| `discovery.on_access.allow_processes` | list of strings | empty | Processes allowed to trigger on-access discovery. |
| `discovery.on_access.handler_timeout` | duration | `20ms` | Time limit for one on-access handler. |

### `repositories` (at least one)

| Key | Type | Default | Meaning |
|---|---|---|---|
| `id` | identifier | none (required) | Unique ID matching `^[a-z][a-z0-9_-]{0,31}$`. |
| `repository` | string | none (required) | The Restic repository, passed to Restic as given. |
| `restic_binary` | absolute path | none (required) | The `restic` executable. |
| `rclone_binary` | absolute path | empty | The `rclone` executable. Required when `repository` starts with `rclone:`. |
| `password_file` | absolute path | none (required) | File that holds the repository password. It must exist, be a regular file and have mode `0600` or `0400`. |
| `cache_dir` | absolute path | empty | Restic cache directory. Cannot be combined with `no_cache: true`. |
| `no_cache` | boolean | `false` | Run Restic without a cache. |
| `lock_mode` | string | `normal` | `normal`, or `none` to pass `--no-lock`. `none` removes coordination with repository maintenance. |
| `environment` | map | empty | Extra environment variables for Restic. Keys must match `^[A-Z_][A-Z0-9_]*$`. `RESTIC_PASSWORD*` keys are rejected; use `password_file` instead. |

### `roots` (at least one)

| Key | Type | Default | Meaning |
|---|---|---|---|
| `id` | identifier | none (required) | Unique ID matching `^[a-z][a-z0-9_-]{0,31}$`. |
| `local_path` | absolute path | none (required) | The local directory tree. |
| `repository_id` | string | none (required) | The `id` of a repository in this file. |
| `prefix_map[].hostname` | string | empty | The host whose snapshots this mapping applies to. |
| `prefix_map[].source_path` | absolute path | none | The path that was backed up on that host. |
| `prefix_map[].tree_prefix` | absolute path | none | Where that path sits inside the snapshot tree. |
| `snapshots.hostname` | string | empty | Show only this host's snapshots. Empty means all hosts. |
| `snapshots.tags_all` | list of strings | empty | Show only snapshots that carry all of these tags. |
| `snapshots.source_paths_exact` | list of absolute paths | empty | Show only snapshots with exactly these source paths. |
| `seed_paths[].path` | relative path | none | Directory under `local_path` to seed `.snapshot` entries in. |
| `seed_paths[].max_depth` | integer | none | How many levels deep to seed. At least 1. |
| `exclude_relative_paths` | list of relative paths | empty | Directories under `local_path` to skip. |

On Linux, the daemon also watches each `seed_paths` entry with inotify and links directories
created under it afterwards, staying inside that entry's `max_depth` and skipping the root's
`exclude_relative_paths`. It watches only those entries, not the whole `local_path`, so a root
without `seed_paths` is not watched. On macOS there is no watcher: run `snapback seed` again after
creating directories.
| `snap.tags` | list of strings | empty | Tags added to snapshots taken with `snapback snap`. |

Relative paths must be clean and must not contain `..`.

### `service`

| Key | Type | Default | Meaning |
|---|---|---|---|
| `service.manager` | string | `auto` | `auto` or `systemd`. `launchd` and `openrc` are rejected in v0.1. |
| `service.scope` | string | `user` | `user` or `system`. |
| `service.run_as_user` | string | empty | The user the service runs as. |

### Layout rules

The validator also checks how the paths relate to each other:

- `history_mount` and `backend_mount_dir` must not be at or above any root's `local_path`.
- The two mounts must not overlap each other or the directory of a local repository.
- `state_dir` must not be inside either mount.

## Validation errors

Each error names its field as a YAML path, such as `repositories[0].id`. Snapback reports every
failure at once, not only the first. For example, this file has an uppercase repository ID. The
root's `repository_id` therefore matches no repository, and the password file does not exist:

```yaml
version: 1
repositories:
  - id: Personal
    repository: /srv/restic
    restic_binary: /usr/bin/restic
    password_file: /etc/restic/password
roots:
  - id: home
    local_path: /home/alex
    repository_id: personal
```

`snapback config validate` prints this and exits with status 1:

```text
snapback config: repositories[0].id: must match ^[a-z][a-z0-9_-]{0,31}$; roots[0].repository_id: names no repository; repositories[0].password_file: does not exist
fix: run 'snapback config validate' and correct the reported fields
```

With `--json`, the same result is a single object:

```json
{"ok":false,"code":"invalid_configuration","error":"repositories[0].id: must match ^[a-z][a-z0-9_-]{0,31}$; roots[0].repository_id: names no repository; repositories[0].password_file: does not exist","fix":"run 'snapback config validate' and correct the reported fields"}
```

Unknown keys are rejected when the file is read. A `password` key in a repository is rejected
with a hint to use `password_file`, because passwords never go in the config file.
