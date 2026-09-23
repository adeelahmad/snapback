# Snapback — Specification Addendum B: opt-in telemetry and diagnostics

**Status:** approved design, not yet planned into sprints. Section 2's field list is
superseded — see the note at the top of that section.
**Applies to:** `SPEC.md` Revision 2 and `SPEC-ADDENDUM-A.md` (Rev 2.1). This addendum adds sections; it rewrites none.
**Date:** 23 September 2026.

Nothing here weakens the spine in Addendum A §1. Telemetry is read-side only: it never
writes to a backup repository, never touches the user's real files, and never changes what
snapback reports. A telemetry failure MUST never fail a snapback command.

## 1. Opt-in, off by default

1. Telemetry MUST be off by default in every build and on every platform.
2. The installer, package upgrades, `snapback setup`, the daemon's first run and the web UI
   MUST NOT enable telemetry, MUST NOT pre-tick a consent box, and MUST NOT ask in a way
   whose default answer is "on".
3. Telemetry is enabled only by an explicit `telemetry: on` in `config.yaml` or by the user
   running `snapback telemetry on`. No other path may set it.
4. An upgrade MUST preserve the existing setting and MUST NOT re-prompt.
5. With telemetry off, snapback MUST make no network call for telemetry and MUST keep no
   telemetry queue on disk.

## 2. Exhaustive field list (what MAY be sent)

**Superseded (Ruling S6-R3, 2026-09-23):** sprint 6 planning refined this section's
speculative field list into a closed set of five named events with fixed attributes —
`setup.completed`, `daemon.started`, `mount.ready`, `doctor.failed`, `error` — and
dropped the open-ended counters below for restores, `snap` runs and the first
`.snapshot` entry listed, none of which were implemented. `internal/telemetry/event.go`
and `docs/agents/sprint6-telemetry/plan.md` are the authoritative source for what
actually ships; the list below is kept for history only.

When telemetry is on, only the following MAY leave the machine. This list is exhaustive: a
field not named here MUST NOT be sent, and adding one requires a revision of this addendum.

1. Snapback version, build commit and release channel.
2. Operating system name, OS major version and CPU architecture.
3. Install method (`install.sh`, package manager name, container image, built from source).
4. Coarse event counters, each an integer count per reporting window:
   - setup completed;
   - first `.snapshot` entry listed;
   - restores performed;
   - `snap` runs;
   - daemon starts.
5. Error codes from `internal/errcode` — the stable numeric or symbolic code only, never the
   message and never any value interpolated into it.
6. Timing buckets: durations reported as coarse buckets (for example `<100ms`, `<1s`, `<10s`,
   `>=10s`), never exact timings that could fingerprint a repository.
7. An installation identifier that is a locally generated random value, not derived from any
   hostname, MAC address, username or repository, and deleted by `snapback telemetry off`.

## 3. Prohibited content (what MUST NEVER be sent)

The following MUST NEVER be sent, in any field, in any encoding, hashed or in the clear:

1. File names, file paths, directory names or path fragments.
2. Repository URIs, bucket names, host names or IP addresses of any backend.
3. The machine's hostname, domain, usernames or home directory name.
4. Snapshot IDs, tree IDs, blob IDs or backup tags.
5. Config file contents, including instance names and environment variables.
6. Free-text error strings, stack traces, command lines, arguments or log lines.
7. Any file content, size distribution or directory listing.

A field that cannot be produced without one of the above MUST be dropped, not redacted.

## 4. Transport

1. Telemetry is emitted as OpenTelemetry over OTLP/HTTP.
2. The collector endpoint MUST be configured by the user (`telemetry.endpoint`). There is no
   built-in endpoint, and the project operates no default collector.
3. With `telemetry: on` but no endpoint set, snapback MUST send nothing and MUST say so in
   `snapback telemetry status`.
4. The endpoint MUST be HTTPS unless it resolves to a loopback address.
5. Sending MUST be asynchronous and bounded: a fixed-size local queue, a send timeout, and
   silent drop when the queue is full. Telemetry MUST NOT block or delay any command.

## 5. Crash reports

1. Crash reports are a separate opt-in (`telemetry.crash_reports: on`); turning telemetry on
   MUST NOT turn crash reports on.
2. Crash reports use the Sentry protocol so a self-hosted GlitchTip (or compatible) endpoint
   can receive them. The endpoint is user-configured; there is no built-in DSN.
3. Crash reports obey §3 in full: frames carry function and package names from snapback's own
   modules only; paths, arguments, local variables and environment MUST be stripped before
   the report is queued, not at the receiver.

## 6. See and delete

1. `snapback telemetry status` MUST print exactly what is enabled (telemetry, crash reports),
   the configured endpoint, the installation identifier, and the full last payload as it was
   or would be sent, in the machine-readable JSON form of `SPEC.md` §6.
2. `snapback telemetry off` MUST stop all sending immediately, MUST delete the local queue and
   the installation identifier, and MUST leave the setting off across upgrades.
3. There MUST be no command that uploads history retroactively; anything not sent while
   telemetry was on is gone.

## 7. `snapback doctor --bundle`

1. `snapback doctor --bundle` MUST work with telemetry off and MUST send nothing anywhere.
2. It produces one local archive at a path the user chooses, redacted by the §3 rules, and
   prints the path and a summary of what the archive contains.
3. The user attaches the archive to an issue manually. This is the supported way to get
   diagnostics to the project, and the docs MUST present it before telemetry.

## 8. Privacy page on the docs site

1. The docs site MUST carry a privacy page that lists the §2 fields verbatim, states the §3
   prohibitions, and explains how to turn telemetry on and off.
2. The same page MUST disclose that the docs site itself is served with Google Analytics, what
   that collects, and that the site's analytics are independent of snapback's telemetry.
3. The privacy page MUST be linked from the docs site navigation and from
   `snapback telemetry status`.

## 9. Acceptance

1. A default install, an upgrade, `setup` and the web UI each make zero telemetry network
   calls (observed with a loopback collector that records every request).
2. With telemetry on and a loopback collector, every field received is in the §2 list; a test
   fixture containing paths, a repository URI and a snapshot ID yields none of them.
3. `snapback telemetry off` removes the queue and identifier from disk.
4. `snapback doctor --bundle` with telemetry on makes no network call, and the archive
   contains no §3 item (scanned by the test).
