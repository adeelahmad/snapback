# Privacy

## What Snapback reports today: nothing

Snapback ships no telemetry code. The CLI and the daemon make no network call
other than to your Restic repository and to whatever storage backend that
repository lives on. There is no usage reporting, no crash reporting, no
phone-home check and no identifier on disk.

The rest of this page describes the contract that any future telemetry must
meet. It is a promise about what the code will be allowed to do, not a
description of something you can turn on now.

## The opt-in contract for telemetry

Telemetry is off by default in every build and on every platform. The
installer, package upgrades, `snapback setup`, the daemon's first run and the
web UI will never turn it on, never pre-tick a consent box and never ask a
question whose default answer is "on". An upgrade will keep whatever you chose
and will not ask again. While telemetry is off, snapback will make no telemetry
network call and will keep no telemetry queue on disk.

### What telemetry may report

This list is exhaustive. A field that is not named here will never leave your
machine, and adding one needs a revision of the published addendum.

1. Snapback version, build commit and release channel.
2. Operating system name, OS major version and CPU architecture.
3. Install method (`install.sh`, package manager name, container image, built
   from source).
4. Coarse event counters, each an integer count per reporting window:
    - setup completed;
    - first `.snapshot` entry listed;
    - restores performed;
    - `snap` runs;
    - daemon starts.
5. Error codes from `internal/errcode` — the stable numeric or symbolic code
   only, never the message and never any value interpolated into it.
6. Timing buckets: durations reported as coarse buckets (for example `<100ms`,
   `<1s`, `<10s`, `>=10s`), never exact timings that could fingerprint a
   repository.
7. An installation identifier that is a locally generated random value, not
   derived from any hostname, MAC address, username or repository, and deleted
   when you turn telemetry off.

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

### Where it would go

Telemetry will use OpenTelemetry over OTLP/HTTP to a collector **you**
configure with `telemetry.endpoint`. There is no built-in endpoint and the
project runs no default collector, so with telemetry on but no endpoint set,
nothing will leave the machine. The endpoint must be HTTPS unless it resolves
to a loopback address, and delivery will be asynchronous and bounded so it can
never block or slow a command.

### Crash reports are a separate choice

Crash reports are their own opt-in (`telemetry.crash_reports`). Turning
telemetry on will not turn crash reports on. They will use the Sentry protocol
so you can point them at a self-hosted GlitchTip or compatible endpoint; there
is no built-in DSN. Crash reports obey the prohibition list in full: frames
carry function and package names from snapback's own modules only, and paths,
arguments, local variables and environment are stripped before a report is
queued, not at the receiver.

### Seeing and deleting what was collected

Two commands are planned; neither is available today.

- `snapback telemetry status` will print exactly what is on, the configured
  endpoint, the installation identifier and the full last payload as it was or
  would be delivered, in the machine-readable JSON form used elsewhere in
  snapback.
- `snapback telemetry off` will stop all reporting immediately, delete the
  local queue and the installation identifier, and keep the setting off across
  upgrades.

Nothing will ever be uploaded retroactively: whatever was not reported while
telemetry was on is gone.

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
