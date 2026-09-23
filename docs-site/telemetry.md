# Telemetry

What Snapback can report, exactly, and the two switches that control it. For
what to run on your own infrastructure to receive it, see
[Self-hosted telemetry collector](telemetry-collector.md); for the full
prohibition list and every attribute an event may carry, see
[Privacy](privacy.md).

## Off until you turn it on

Telemetry is disabled in every build, on every platform, by default. Nothing
in setup, an install or an upgrade turns it on for you, and turning it on is a
separate, explicit step from turning crash reports on.

## Configuration keys

The `telemetry:` section of your config has four keys:

```yaml
telemetry:
  enabled: false
  endpoint: ""
  crash_reports: false
  crash_endpoint: ""
```

- `telemetry.enabled` — turns event reporting on or off. Off by default.
- `telemetry.endpoint` — the OTLP/HTTP collector base URL events are sent to.
- `telemetry.crash_reports` — a separate opt-in for crash reports.
- `telemetry.crash_endpoint` — the Sentry-protocol endpoint crash reports are
  POSTed to.

## `enable` refuses without an endpoint

`telemetry.endpoint` must be set before enabling telemetry: `snapback telemetry enable` refuses to turn telemetry on otherwise.

There is nothing built in for events to reach, so set the endpoint first,
then run `enable`.

## The five events

The closed set of events Snapback can ever emit:

- `setup.completed`
- `daemon.started`
- `mount.ready`
- `doctor.failed`
- `error`

No other event name exists. Adding one means changing the schema in
`internal/telemetry` and this page together.

## Duration buckets

Every duration an event reports is one of five coarse labels, never an exact
number, so a timing cannot fingerprint a repository:

- `<100ms`
- `<1s`
- `<10s`
- `<60s`
- `>=60s`

## The loopback/HTTPS rule

An endpoint using the `http` scheme is only accepted when its host is a
loopback literal: `127.0.0.1`, `::1` or `localhost`. Any other host must use
`https`; an `http` endpoint to a non-loopback host is rejected outright.

## No collector ships with Snapback

Snapback ships no collector and no default endpoint. With telemetry on but no
endpoint configured, nothing leaves the machine — and, as above, `enable`
will not even let you get into that state. To receive events you stand up
your own collector; bring the two example services up with:

```sh
docker compose up -d
```

against the `otel-collector` and `glitchtip` services described in
[Self-hosted telemetry collector](telemetry-collector.md).
