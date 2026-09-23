# Self-hosted telemetry collector

Snapback ships no collector and no default endpoint. Telemetry is off until you
turn it on, and turning it on with `snapback telemetry enable` requires you to
supply an `endpoint` — there is nothing built in for it to talk to. If you want
to receive the events described in [Privacy](privacy.md), you run your own
OpenTelemetry collector and your own GlitchTip instance (or another
Sentry-protocol-compatible receiver in front of one); nothing in this page
runs on Snapback's infrastructure.

This page walks through one way to stand that up with Docker Compose.

## The example stack

Two files, both under `docs/examples/telemetry/` in the Snapback repository:

- `docker-compose.yml` defines two services: `otel-collector`, an
  [OpenTelemetry Collector Contrib](https://github.com/open-telemetry/opentelemetry-collector-contrib)
  instance that receives Snapback's OTLP metrics, and `glitchtip`, a
  self-hosted [GlitchTip](https://glitchtip.com/) instance that stores and
  displays them.
- `otel-collector.yaml` is the collector's own config: an `otlp` receiver with
  the `http` protocol only (no `grpc`), forwarding to GlitchTip's ingest.

Copy both files into a directory you control, then:

```sh
cp docs/examples/telemetry/docker-compose.yml .
cp docs/examples/telemetry/otel-collector.yaml .
```

### Ports stay on loopback

Every port either service publishes is bound to `127.0.0.1` only
(`127.0.0.1:4318:4318` for the collector's OTLP/HTTP receiver,
`127.0.0.1:8000:8000` for GlitchTip's web UI), never `0.0.0.0`. If you need
Snapback or a browser on another machine to reach either service, put a
reverse proxy with TLS in front of it yourself — the compose file will not
expose either port to the network on its own.

### Secrets live in your own `.env`, not in the compose file

The compose file sets no default credentials in plain text. `glitchtip`'s
`DATABASE_URL`, `REDIS_URL`, `SECRET_KEY` and `GLITCHTIP_DOMAIN` are all
`${VAR}` references resolved from an `.env` file you write next to the
compose file and never commit:

```sh
cat > .env <<'EOF'
DATABASE_URL=postgres://glitchtip:CHANGE-ME@your-postgres-host:5432/glitchtip
REDIS_URL=redis://your-redis-host:6379/0
GLITCHTIP_SECRET_KEY=CHANGE-ME
GLITCHTIP_DOMAIN=https://your-glitchtip-host:8000
GLITCHTIP_INGEST_TOKEN=CHANGE-ME
EOF
```

GlitchTip needs a Postgres database and a Redis instance of its own; this
example does not bundle either, so point `DATABASE_URL` and `REDIS_URL` at
instances you already run, or add them to your own compose override.

### Bring it up

```sh
docker compose up -d
```

## Pointing Snapback at it

Once the collector is reachable, configure Snapback with the endpoint and
turn telemetry on:

```yaml
telemetry:
  enabled: true
  endpoint: "https://your-collector-host:4318"
```

```sh
snapback telemetry enable
```

`snapback telemetry show` always prints the exact payload Snapback would send,
whether or not telemetry is on, so you can see what your collector will
receive before you turn anything on. See [Privacy](privacy.md) for the closed
list of events and attributes, and [CLI reference](cli.md) for the full
`telemetry` command surface.
