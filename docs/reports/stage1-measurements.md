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

Each row is one `pathtemplate-<goos>.json` file: restic `mount --path-template ids/%I`
was checked to expose the snapshot under its full ID.

| Evidence | Platform | Result | Snapshot ID |
| --- | --- | --- | --- |
| pathtemplate-darwin.json | macOS (Apple Silicon, macFUSE, local host) | pass | dbd065721f05c11d20f80e67270fe6f4507b20a958cc07279fe862614aa88ede |
| pathtemplate-linux.json | Linux (ubuntu-latest CI, fuse3) | pass | cc439810023786886736bf838bb44e4a6f5fd2d1050687ca2f049abdcd3d7dea |

## Catalog mount, browse and unmount

Each row is one `catalog-<goos>.json` file: the tiny directory/symlink catalog was
mounted, browsed and unmounted.

| Evidence | Platform | Result | go-fuse |
| --- | --- | --- | --- |
| catalog-darwin.json | macOS (Apple Silicon, macFUSE, local host) | pass | v2.11.0 |
| catalog-linux.json | Linux (ubuntu-latest CI, fuse3) | pass | v2.11.0 |

## Metadata fidelity

Each row is one `fidelity-<goos>.json` file: files read through the `.snapshot`
alias were compared with `restic ls --json` on mode, size and mtime.

| Evidence | Platform | Files compared | mtime precision (ns) | Result |
| --- | --- | --- | --- | --- |
| fidelity-darwin.json | macOS (Apple Silicon, macFUSE, local host) | 9 | 1 | pass |
| fidelity-linux.json | Linux (ubuntu-latest CI, fuse3) | 9 | 1 | pass |

restic ls --json 0.19 does not report symlink size or link target. The symlink
was compared on mode and mtime, and its target was checked through the catalog
alias.

ctime and birth time: FUSE-approximated; recorded, not claimed.

## Latency over rclone:gdrive

Each row is one measurement from `latency.json`, in milliseconds.

| Measurement | Median (ms) | Min (ms) | Max (ms) | Samples |
| --- | --- | --- | --- | --- |
| cold_listing | 32.7 | 32.7 | 32.7 | 1 |
| warm_prewarmed_listing | 1.4 | 0.7 | 2.2 | 3 |
| cold_first_file_read | 3012.4 | 3012.4 | 3012.4 | 1 |
| warm_listing_after_restart | 6.4 | 4.4 | 21.1 | 3 |

Data: 100 files, 4096 bytes

Remote deleted: true

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
