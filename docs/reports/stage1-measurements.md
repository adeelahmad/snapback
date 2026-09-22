# Stage 1 measurements

This report collects the Stage 1 measurements for Snapback. Every figure is
copied from the evidence files under `docs/reports/stage1/`. The darwin files
come from a local macOS host; the linux files come from the `fuse-linux` CI
job (run 35683903691); `latency.json` comes from a run against an
rclone Google Drive remote. The report records results. The latency call is
left to a human; see Open items.

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

Each row is one `rows[]` entry of a `crawler-<goos>.json` file: the tool was run
over a seeded tree, and Hits is the `hits` value recorded for it. `—` means the tool
was not run here, so there is no count.

| Platform | Tool | Status | Hits |
| --- | --- | --- | --- |
| macOS (Apple Silicon, macFUSE, local host) | rg | tested | 0 |
| macOS (Apple Silicon, macFUSE, local host) | rg -L | tested | 761 |
| macOS (Apple Silicon, macFUSE, local host) | fd | tested | 0 |
| macOS (Apple Silicon, macFUSE, local host) | fd -L | tested | 481 |
| macOS (Apple Silicon, macFUSE, local host) | find | tested | 0 |
| macOS (Apple Silicon, macFUSE, local host) | find -L | tested | 400 |
| macOS (Apple Silicon, macFUSE, local host) | rsync -a | tested | 0 |
| macOS (Apple Silicon, macFUSE, local host) | vscode search | not-tested-here | — |
| Linux (ubuntu-latest CI, fuse3) | rg | tested | 0 |
| Linux (ubuntu-latest CI, fuse3) | rg -L | tested | 640 |
| Linux (ubuntu-latest CI, fuse3) | fd | tested | 0 |
| Linux (ubuntu-latest CI, fuse3) | fd -L | tested | 640 |
| Linux (ubuntu-latest CI, fuse3) | find | tested | 0 |
| Linux (ubuntu-latest CI, fuse3) | find -L | tested | 880 |
| Linux (ubuntu-latest CI, fuse3) | rsync -a | tested | 0 |
| Linux (ubuntu-latest CI, fuse3) | vscode search | not-tested-here | — |

Not tested here, macOS (Apple Silicon, macFUSE, local host), `vscode search`: VS Code search needs a GUI session; it follows symlinks only when search.followSymlinks is true (the default)

Not tested here, Linux (ubuntu-latest CI, fuse3), `vscode search`: VS Code search needs a GUI session; it follows symlinks only when search.followSymlinks is true (the default)

## Requirement matrix

Each row is one SPEC §22 row 1 item. The status is backed by the evidence
files above where one exists.

| Item | Status | Reason |
| --- | --- | --- |
| Pin dependencies | implemented and tested | go.mod pins go-fuse and the Go toolchain; the Pinned versions table is checked against go.mod and the evidence files |
| Disposable Restic repo | implemented and tested | `internal/compat/resticfx` creates and removes a throwaway repository; its unit tests and the path-template and fidelity runs use it; no separate evidence file records it |
| Verify --path-template ids/%I | implemented and tested | pathtemplate-darwin.json and pathtemplate-linux.json record pass |
| Tiny directory/symlink FUSE catalog on Linux | implemented and tested | catalog-linux.json records pass |
| Tiny directory/symlink FUSE catalog on macOS | implemented and tested | catalog-darwin.json records pass |
| Metadata-fidelity check | implemented and tested | fidelity-darwin.json and fidelity-linux.json record pass; ctime and birth time are recorded, not claimed |
| rclone/Google Drive latency measurements | implemented and tested | latency.json records four measurements and remote_deleted true |
| Crawler test | implemented and tested | crawler-darwin.json and crawler-linux.json record tested rows; VS Code search is not tested here on either platform (needs a GUI session) |

## Missing evidence

None. All nine expected evidence files are present under `docs/reports/stage1/`:
`pathtemplate-darwin.json`, `pathtemplate-linux.json`, `catalog-darwin.json`,
`catalog-linux.json`, `fidelity-darwin.json`, `fidelity-linux.json`,
`crawler-darwin.json`, `crawler-linux.json` and `latency.json`.

## Open items

Latency go/no-go: PENDING — human decision

- Proceed to Stage 2 with the current latency numbers.
- Add a deeper pre-warm of the snapshot catalog before Stage 2.
- Pull the §23 local hot-cache repository forward into an earlier stage.

## Addendum A §11.1: per-file content identity (restic 0.19.0)

Run on darwin/arm64 against two disposable repositories (`restic 0.19.0
compiled with go1.26.4`), deleted after the run. Tree: `big.bin` (1 500 000 B
random), `huge.bin` (9 000 000 B random), `mid.txt` (250 B), `small.txt` (12 B).

### Fields a file node carries

`restic ls --json <id>` emits per file: `name`, `type`, `path`, `uid`, `gid`,
`size`, `mode`, `permissions`, `mtime`, `atime`, `ctime`, `inode`. It carries
**no** `content` and no whole-file hash.

`restic cat tree <snapshot-id>` (and `restic cat blob <tree-id>` for subtrees)
emits per file: `name`, `type`, `mode`, `mtime`, `atime`, `ctime`, `uid`,
`gid`, `user`, `group`, `inode`, `device_id`, `size`, `links`, and `content` —
a list of data-blob IDs. Directory nodes carry `content: null` and `subtree`.
There is **no** whole-file hash field; `content` is the only content-derived
identity, and it is a list, not a scalar.

### Cold vs warm cost

With `RESTIC_CACHE_DIR` emptied, `restic cat blob <subtree-id>`:

| Call | Wall time |
| --- | --- |
| cold (cache removed) | 0.826 s |
| warm, run 1 | 1.287 s |
| warm, runs 2-4 | 0.702 s, 0.713 s, 0.719 s |

Both are dominated by process start and key derivation, not by the read. After
the cold call the cache held `snapshots/`, `index/` and `data/`, totalling
**16 K** with a single 4.0 K entry under `data/` (the tree pack), while
`repo1/data` was **1.4 M**. No data pack was fetched. Honest limits: `fs_usage`
needs root and was not run, so this is a cache-contents inference, not a
syscall trace; per-call timings include restic start-up and cannot isolate the
network or disk read.

### Re-chunking

Within one repository the `chunker_polynomial` is fixed (`restic cat config`
reported `2edee7b50f2aa7` for repo1). A second backup of the unchanged tree
produced byte-identical `content` lists for `big.bin`, `mid.txt` and
`small.txt`. Modifying `huge.bin`, reverting it and backing up again with
`--pack-size 64` produced the same 7-element `content` list: pack size does not
affect chunking.

A second repository (`restic init`, polynomial `38f8a858b3b20f`) backed up the
same tree:

| File | repo1 `content` | repo2 `content` |
| --- | --- | --- |
| `big.bin` (1 500 000 B) | 1 chunk `ebc35cc8…` | 1 chunk `ebc35cc8…` (identical) |
| `mid.txt` (250 B) | 1 chunk `f8017b1a…` | identical |
| `small.txt` (12 B) | 1 chunk `efde3611…` | identical |
| `huge.bin` (9 000 000 B) | 7 chunks `711199dc…`…`6b9635f9…` | 8 chunks `1c3c6661…`…`15ed3165…`, no ID in common |

Single-chunk files hash the whole file, so their one blob ID is polynomial-
independent and matches across repositories. Multi-chunk files do not: the cut
points differ, so every blob ID differs and there is no overlap to compare.

### Conclusion

Snapback can read a per-file content identity from tree metadata alone, with no
data-blob reads: the node's `content` list. Within one repository that list is
stable across backups and across pack-size changes, so rung 1 can compare it
directly for same-repository instances.

Across two repositories it is **not** a usable identity in general. It survives
re-chunking only where the node's `content` holds exactly one blob — here
`big.bin` at 1 500 000 B, `mid.txt` and `small.txt` did; for multi-chunk files
the lists are disjoint, and restic exposes no whole-file hash to fall back on.
For cross-repository comparison rung 1 must therefore use size plus mtime, and
may use `content` only as an opportunistic match when both nodes have exactly
one chunk.

### Refinements (human review, 2026-09-22)

**The single-chunk rule is structural, not size-based.** Chunk boundaries come
from the per-repository Rabin polynomial (`chunker_polynomial` in the repo
config), so blob IDs for multi-chunk files diverge across repositories. A blob
ID is SHA-256 of the plaintext chunk, so a node whose `content` has exactly one
entry carries an ID equal to sha256(whole file), comparable everywhere. The
rule is `len(content) == 1`, never a byte size: files between the 512 KiB
minimum and 8 MiB maximum chunk size can land on either side, depending on
where the polynomial cuts.

**Within one repository, `content` is the strongest signal.** The same
polynomial means the same boundaries: an equal blob-ID list is exact content
identity, and equal tree IDs mean identical subtrees. The primary `.snapshot`
overlay is per repository, so for per-repository rung 1 size+mtime is the fast
pre-filter and `content` the confirmation, not the other way round.

**Chunker-copy edge case.** Stated by the human; not measured in this spike.
`restic copy` moves blobs verbatim; it does not re-chunk. A hot-cache repository
created without `--copy-chunker-params` holds the primary's boundaries for
copied snapshots but its own for anything backed up into it directly, so one
file can carry two different multi-chunk ID lists in one repository. The
single-blob rule still holds there (still sha256(file)). Requirement for the
§23 hot-repository item: create the cache with `--copy-chunker-params`.
