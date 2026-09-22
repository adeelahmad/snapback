# Logging

Snapback writes structured logs through one shared setting: a level, a format and an optional
file. The same three settings are spelled as command-line flags and as a `logging:` section in
the config file.

## Flags

`snapback run` and `snapback web` each take all three flags:

| Flag | What it sets | Accepted values |
|---|---|---|
| `--log-level` | how much is logged | `debug`, `info`, `warn`, `error` |
| `--log-format` | how a record is encoded | `text`, `json` |
| `--log-file` | a file to write records to | an absolute path |

A flag that is given on the command line wins over the matching `logging:` key, and a setting
that neither gives falls back to the default. An unknown level or format is refused by name,
with the accepted words listed: `snapback run --log-level=verbose` exits 2 with a usage error
before the daemon takes its instance lock, so a typo never leaves a half-started daemon behind.

Every other command is configured through the `logging:` section only.

## Config keys

```yaml
logging:
  level: debug
  format: json
  file: /var/log/snapback.log
```

Every key is optional. Without them, Snapback logs at `info` in `text` format and writes to the
process's own error stream. `logging.file` must be an absolute, clean path.

## The log file

When a file is configured, Snapback appends to it and never truncates it, and it creates the
file and any missing parent directory with the `files.dir_mode` and `files.file_mode` from the
configuration, so the log inherits the same permissions as the rest of Snapback's state.

## What `debug` adds

At `debug` Snapback logs the restic command line it runs, with the password file path redacted.
Every credential environment value is redacted the same way. Debug also records:

- each restic invocation's subcommand, duration and exit code;
- catalog lookups in the history mount: lookup, readdir, readlink and read, with the inode and
  whether the entry was found;
- reader-policy decisions, with the process and the verdict;
- refresh cycles: start, per-repository snapshot counts and evictions.

`info` and above stay quiet about all of that.

## Bundles

`snapback doctor --bundle` copies the daemon's log file into the bundle, after the same
redaction it applies to the configuration and the doctor output. It reads `logging.file` when
the configuration names one, and otherwise `daemon.log` in the state directory. A configured
file that is missing becomes a short placeholder naming the path, so the bundle says why the
log is absent instead of dropping it silently. See [Privacy](privacy.md) for what a bundle
holds.
