# Privacy

## What Snapback ships and reports today

- Snapback ships telemetry code: five events, a hand-rolled OTLP/HTTP JSON exporter, and crash reporting.
- All of it is off by default in every build and on every platform; nothing is collected, queued or sent unless you turn it on.
- `telemetry.enabled` (events) and `telemetry.crash_reports` (crash reports) are two separate, independent switches — turning one on never turns the other on.
- Events are buffered in memory only; there is no on-disk queue, ever. If the collector cannot be reached, or is too slow, the event is dropped and counted, never retried and never written to disk.
- A 128-bit `crypto/rand` install identifier is generated once telemetry is actually used — the first event delivered, not at startup — stored at `<state_dir>/telemetry/install_id` mode `0600`, and deleted by `snapback telemetry disable` (`ForgetInstallID`) so a future `enable` starts with a fresh identity.
- Retention is whatever you configure on your own collector: Snapback and its maintainers never receive or retain anything, because the project runs no maintainer-run collector or endpoint.

The installer, package upgrades, `snapback setup`, the daemon's first run and
the web UI will never turn telemetry on, never pre-tick a consent box and
never ask a question whose default answer is "on". An upgrade keeps whatever
you chose and does not ask again.

### The five events

- `setup.completed`
- `daemon.started`
- `mount.ready`
- `doctor.failed`
- `error`

This is the closed, exhaustive set. No other event name exists.

### Event attributes

Every event carries `version`, `os` and `arch`; some carry more.

| Event | Attributes |
|---|---|
| `setup.completed` | `version`, `os`, `arch`, `outcome`, `duration` |
| `daemon.started` | `version`, `os`, `arch` |
| `mount.ready` | `version`, `os`, `arch`, `duration` |
| `doctor.failed` | `version`, `os`, `arch`, `check` |
| `error` | `version`, `os`, `arch`, `code` |

No other attribute exists on any of them; adding one means changing the
schema in `internal/telemetry` and this page together. See
[Telemetry](telemetry.md) for the configuration keys and the endpoint rules.

### Duration buckets

Every duration an event reports is one of five coarse labels, never an exact
number, so a timing cannot fingerprint a repository:

- `<100ms`
- `<1s`
- `<10s`
- `<60s`
- `>=60s`

### What will never be reported

None of the following may leave your machine, in any field, in any encoding,
hashed or in the clear:

1. File names, file paths, directory names or path fragments.
2. Repository URIs, bucket names, host names or IP addresses of any backend.
3. Your hostname, domain, usernames or home directory name.
4. Snapshot IDs, tree IDs, blob IDs or backup tags.
5. Config file contents, including instance names and environment variables.
6. Free-text error strings, stack traces, command lines, arguments or log lines.
7. Any file content, size distribution or directory listing.

A field that cannot be produced without one of the above will be dropped, not
redacted.

### Where it goes

Telemetry is a hand-rolled OTLP/HTTP JSON exporter — not the OpenTelemetry
SDK — POSTed to a collector **you** configure with `telemetry.endpoint`.
There is no built-in endpoint and the project runs no default collector, so
with telemetry on but no endpoint set, nothing leaves the machine, and
`enable` refuses to turn telemetry on until `telemetry.endpoint` is set. The
endpoint must be HTTPS unless it resolves to a loopback address (`127.0.0.1`,
`::1` or `localhost`), and delivery is asynchronous and bounded so it can
never block or slow a command.

### Crash reports are a separate choice

Crash reports are their own opt-in (`telemetry.crash_reports`). Turning
telemetry on does not turn crash reports on. They use the Sentry envelope
protocol so you can point them at a self-hosted GlitchTip or compatible
endpoint; there is no built-in DSN. Crash reports obey the prohibition list
in full: frames carry function and package names from snapback's own modules
only, and paths, arguments, local variables and environment are stripped
before a report is sent, not at the receiver.

### Seeing and changing what is sent

`snapback telemetry` has four verbs:

- `status` prints whether telemetry and crash reports are on, the configured endpoints and the install identifier.
- `show` prints the exact bytes that would be POSTed for a sample of all five events — whether telemetry is on or off — and sends nothing.
- `enable` turns telemetry on, refusing when no `telemetry.endpoint` is configured.
- `disable` turns telemetry off and deletes the install identifier.

Nothing is ever uploaded retroactively: whatever was not sent while telemetry
was on is gone. See [Telemetry](telemetry.md) and [the CLI
reference](cli.md) for the full detail on each verb.

## Sharing diagnostics without telemetry

`snapback doctor --bundle` is available today and is the supported way to get
diagnostics to the project. It works with telemetry off, makes no network call,
and writes one local archive at a path you choose, redacted by the prohibition
list above. It prints the path and a summary of what the archive contains, and
you decide whether to attach it to an issue.

## This documentation site uses Google Analytics

The site you are reading is served with Google Analytics (measurement ID
`G-BFWW49ZP0E`). It records page views and the usual browser-side details that
come with them — page URL, referrer, approximate location derived from IP
address, browser and device type — and Google sets cookies to count returning
visitors. Blocking the script or using a content blocker does not affect the
docs.

Site analytics are entirely separate from snapback itself. The CLI and the
daemon never talk to Google Analytics, and nothing you do on this site is
linked to anything on your machine.
