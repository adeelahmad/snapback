# Stage 1 measurements

This report collects the Stage 1 measurements for Snapback. Every figure is
copied from the evidence files under `docs/reports/stage1/`. The darwin files
come from a local macOS host; the linux files come from the `fuse-linux` CI
job (run 35683903691); `latency.json` comes from a run against an
rclone Google Drive remote. The report records results. It does not make the
latency go/no-go decision; that decision is left to a human.

## Pinned versions

| Component | Version |
| --- | --- |
| go-fuse | v2.11.0 |
| restic | 0.19.0 |
| rclone | v1.75.0 |
| Go toolchain | go1.27.1 |

## Path-template verification

Not filled yet.

## Catalog mount, browse and unmount

Not filled yet.

## Metadata fidelity

Not filled yet.

## Latency over rclone:gdrive

Not filled yet.

## Crawler hit counts

Not filled yet.

## Requirement matrix

Not filled yet.

## Missing evidence

None. All nine expected evidence files are present under `docs/reports/stage1/`:
`pathtemplate-darwin.json`, `pathtemplate-linux.json`, `catalog-darwin.json`,
`catalog-linux.json`, `fidelity-darwin.json`, `fidelity-linux.json`,
`crawler-darwin.json`, `crawler-linux.json` and `latency.json`.

## Open items

Not filled yet.
