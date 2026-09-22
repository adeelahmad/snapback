# Snapback — Specification Addendum A (Revision 2.1)

**Status:** approved design, not yet planned into sprints.
**Applies to:** `SPEC.md` Revision 2. This addendum adds sections; it does not rewrite existing ones except where a rule below says "supersedes".
**Date:** 22 September 2026.

## 0. Instructions to the orchestrator

1. Commit this file to the repository as `SPEC-ADDENDUM-A.md` next to `SPEC.md`, and add a one-line pointer at the top of `SPEC.md`: "See SPEC-ADDENDUM-A.md (Rev 2.1) for instances, consistency, ignore files and cache configuration."
2. Do **not** reopen Stage 1. The only Stage 1 change is the extra measurement in §11.1, which is recorded in the existing compatibility report, not a new story.
3. Stage 2 (core vertical slice) stays single-instance in behaviour but must be **built on the config schema in §2**, so that later stages add instances without a migration. Treat §2 as a Stage 2 change to the config and resolver stories; everything else in this document lands in Stage 3 or later per the placement table in §11.
4. Every feature below obeys the four rules in §1. A story that needs to break one of them is a planning defect; send it back, do not implement it.
5. Where this document says "verify in the spike", the orchestrator must not let a worker assert backend internals from memory. The worker runs the real tool against a disposable repository and records what it observed in `docs/reports/stage1/`.
6. Update `ORCHESTRATOR.md`'s stage table with the acceptance numbers in §11.3 and add a `## Addendum A` row group to the task log when planning begins.
7. `ARCHITECTURE.md`, `README.md` and the docs site must follow the positioning rules in §13 whenever they are touched.

## 1. Spine (non-negotiable rules)

These four rules already hold in Rev 2; this addendum makes them explicit because every new feature leans on them.

1. **Read-side only.** Snapback never creates, schedules, prunes, repairs or writes to any backup repository or filesystem snapshot. It is not a backup runner.
2. **Never write to live data.** Snapback writes only application-owned artefacts: the `.snapshot` link (existing rule), its own state directory, its own mounts and caches. It never sets xattrs, tags, attributes or marker files on the user's real files or directories, and never creates or deletes ignore files (§7).
3. **Report, don't fix.** When snapback finds a problem it reports it and, at most, offers a normal restore-copy to a location the user chooses. It never resolves disagreements automatically, never applies majority voting, never "heals" a backend.
4. **Machine-composable.** Every report and every failure has a stable exit code and a versioned JSON form (§6). Humans get the pretty form; scripts get the truth.

Wording rule for docs and UI: matching copies **agree**; they are not "verified" or "proven good". The word **independent** is reserved for §4 rung 2 only.

## 2. Instances (supersedes `repositories:` in SPEC.md §config)

### 2.1 Model

A **backend** is a kind of provider (`restic`, later `kopia`, `borg`, `zfs`, `btrfs`). An **instance** is one configured target of a backend with a user-chosen name: `restic-nas`, `restic-usb`, `borg-hetzner`. The same backend may appear any number of times. Users think in instances ("Borg to the Hetzner box"), so every folder, report and command names the instance, never the bare backend.

The `SnapshotProvider` seam is unchanged: one `SnapshotProvider` implementation per backend type, one live provider *object* per instance.

### 2.2 Configuration model — rclone as the reference

Configuration follows rclone's remote model exactly, in both the CLI and the web UI:

- `snapback config` (existing frictionless command) gains an interactive **instance** flow: `n) New instance` → choose backend type → answer that backend's prompts → instance is validated and written → it now exists by name. `e) Edit`, `d) Delete`, `r) Rename`, `t) Test` as in rclone. Non-interactive form: `snapback instance add NAME --type restic --repository ... --password-file ...`.
- `snapback instance list|show NAME|test NAME|remove NAME|rename OLD NEW`.
- The web UI **Configuration → Instances** page is the same state machine rendered as forms: Add instance → pick type → type-specific fields → Test → Save. CLI and UI write the same `config.yaml` through the same validator (existing rule). Anything added in one appears in the other.
- Each backend type declares its prompt schema once (Go struct with tags) and both front doors render from it. No hand-written duplicate forms.
- `test` performs the backend's cheapest read-only liveness check (for Restic: `restic cat config`), never a write.

### 2.3 Schema (config version 2)

```yaml
version: 2

instances:                      # replaces `repositories:`; same rules, plus `type`
  - id: restic-nas              # restricted identifier grammar, unique, used as a folder name
    type: restic
    repository: rest:https://nas.local:8000/personal
    restic_binary: /opt/homebrew/bin/restic
    rclone_binary: /opt/homebrew/bin/rclone
    password_file: /Users/alex/.config/restic/password
    cache_dir: /Users/alex/Library/Caches/restic
    lock_mode: normal
    environment: {}
    enabled: true
    consistency: true           # participates in §4 rung 1 (default true when the root enables it)
  - id: restic-gdrive
    type: restic
    repository: rclone:gdrive:Backups/restic
    password_file: /Users/alex/.config/restic/password
    environment:
      RCLONE_CONFIG: /Users/alex/.config/rclone/rclone.conf

roots:
  - id: work
    local_path: /Users/alex/work
    instances: [restic-nas, restic-gdrive]   # replaces `repository_id`; ordered; first is the default
    layout: auto                # auto | flat | by-instance   (§3.3)
    prefix_map:                 # unchanged; now may name an instance
      - instance: restic-gdrive
        hostname: workstation
        source_path: /home/alex/work
        tree_prefix: /home/alex/work
    # snapshots:, seed_paths:, exclude_relative_paths:, snap: unchanged
```

Rules:

- Migration: a `version: 1` file with `repositories:`/`repository_id` loads with a deprecation warning and is rewritten to version 2 on the next `snapback config` save. `type: restic` is inferred. Never silently rewrite on daemon start.
- `roots[].instances` is an ordered list; the first entry is the **default instance** used by `latest`, `by-date` and `snap` when the user does not name one.
- An instance can be referenced by many roots. Deleting an instance that a root references is refused unless `--force`, which removes it from the roots too.
- Instance IDs are folder names inside the mount, so the identifier grammar excludes anything a filesystem or shell treats specially.
- Rev 2 rule preserved: credentials are file references, never plaintext.

### 2.4 Provider roster and order

| Order | Backend | Kind | Notes |
|---|---|---|---|
| 1 | Restic | repository | Only backend in Stages 1–4. Must be proven (latency gate) before any other lands. |
| 2 | Kopia | repository | Stage 6+. Read via the `kopia` CLI, JSON output, same CLI boundary rule as Restic. |
| 3 | Borg | repository | Stage 6+. Read via `borg` CLI with JSON. Note Borg has no FUSE-free random-access read of a file without extracting; the provider must express that as a cost class (§4.4). |
| 4 | ZFS | local snapshot | Stage 6+. Lists `.zfs/snapshot/<name>` on the dataset that contains the root. No remote, no egress. |
| 5 | btrfs | local snapshot | Stage 6+. Read-only subvolume snapshots. |

ZFS/btrfs rule: a dataset may carry snapshots created by znapzend, Sanoid, PVE, zrepl or by hand, all at once. Snapback presents them grouped and labelled by their naming convention/origin where detectable and **never assumes they are its own or prunes them**. Native ZFS scrub/checksum status may be surfaced as that instance's self-reported health, clearly labelled as self-reported (§4.1).

## 3. Instance ↔ root linking by discovery

### 3.1 Three separate things

- **Instance config** answers *how do I reach a repository* (like an rclone remote). It knows nothing about directories.
- **Root + `seed_paths`** answers *which local directories get a `.snapshot` entry*. It knows nothing about repositories.
- **The link between them is discovered, not configured.** Each snapshot inside a repository records its own metadata (for Restic: hostname, paths, tags, time). Snapback reads that listing per instance (Restic: `restic snapshots --json`, already used by the catalog) and matches it to the root.

### 3.2 Matching algorithm (per root, per instance)

1. Candidate snapshots = snapshots in the instance filtered by the root's existing `snapshots:` selectors (hostname, tags, exact source paths).
2. A candidate **covers** the root when one of its recorded paths is an ancestor-or-equal of the root's `local_path` on the same hostname, or a `prefix_map` entry for that instance maps the root to a path inside the snapshot tree.
3. The tree prefix for a covering snapshot is the mapped path (from `prefix_map`) or the recorded path itself.
4. Snapshots that cover the root are what the resolver exposes; snapshots that do not are invisible for this root.
5. An instance with **zero** covering snapshots shows **no folder** in that root's `.snapshot` and is reported as `no-coverage` in `snapback status`, not as an error.

`prefix_map` remains the only manual override, for the cases Rev 2 already lists: different hostname, moved directory, browsing another machine's backups.

### 3.3 Namespace layout

`layout: auto` (default) resolves to `flat` when the root has exactly one enabled instance and `by-instance` otherwise. Users may force either.

```
flat (unchanged from Rev 2):
  .snapshot/2026-09-21_0300Z/…
  .snapshot/latest -> …
  .snapshot/by-date/…

by-instance:
  .snapshot/restic-nas/2026-09-21_0300Z/…
  .snapshot/restic-nas/latest
  .snapshot/restic-gdrive/2026-09-21_0300Z/…
  .snapshot/latest        -> latest of the DEFAULT instance (first in roots[].instances)
  .snapshot/by-date/      -> default instance only
```

Rules: no merged "all instances" timeline in this revision (it hides which copy you are reading, which defeats §4). `daily.N`/`weekly.N` views, when enabled, are per-instance in `by-instance` layout. Switching layout never touches live data; it only changes what the FUSE mount renders.

### 3.4 Cost rule

Multi-instance browsing is opt-in per root and off by default in the setup wizard: N instances means up to N remote round-trips on a cold listing. Each instance is independently toggleable (`enabled`), and a disabled or unreachable instance renders as absent with a status entry, never as a hang: per-instance listing calls run concurrently with the existing `probe_concurrency` and a per-instance timeout (§8.3), and one slow instance must not block the others.

## 4. Consistency ladder

Two rungs, named precisely, never blurred. Both are read-only. Neither runs on the browse path unless §5 says so.

### 4.1 Rung 1 — cross-instance **consistency** (cheap, not independent)

Compares **what each instance reports** about the same logical file version: at minimum size and mtime; where the backend exposes a stable per-file content identity in metadata (verify in the spike, §11.1), that identity too. No file data is transferred.

- Command: `snapback consistency [ROOT|PATH] [--instances a,b] [--since DURATION] [--json]`.
- Unit of comparison: (root-relative path, snapshot time bucket). Snapshots from different instances are paired when their timestamps fall within `consistency.pair_window` (default 6h) and both cover the path; unpaired snapshots are reported as `unpaired`, not compared.
- Per-file result: `agree | disagree | missing-in <instance> | unknown` (unknown = backend gave no comparable identity).
- Must be labelled, in UI and JSON, **"reported by the backends; not an independent check"**. It catches index corruption, a botched sync, a backend that lost track of a file. It cannot catch all backends being wrong together.
- ZFS/btrfs instances contribute their native checksum/scrub status as `self_reported_health`, labelled as such.

### 4.2 Rung 2 — **independent integrity** (expensive, opt-in)

Snapback reads the bytes of a file version from each instance through the provider and computes the hash **itself** (SHA-256, one common layer), trusting no backend's stored claim. Only this rung may use the word "independent".

- Command: `snapback integrity [ROOT|PATH] --instances a,b [--sample N|--all] [--budget-bytes B] [--rate-limit MB/s] [--json]`.
- Never automatic, never on the browse path. If scheduled, the schedule is the user's (`cron`, `launchd`, the daemon's `integrity.schedule` if the user sets it), off by default, with a hard byte budget and rate limit, and it pauses when the daemon detects interactive browsing on the same instance.
- Cost is shown before running: estimated bytes per instance and, for remote instances, a reminder that egress may cost money.
- Result per file version: hashes per instance, `agree | disagree | unreadable-in <instance>`.

### 4.3 Explicitly out of scope

- Auto-heal, majority vote, "replace back in time" — out. The `disagree` report includes the restore command the user *could* run for each copy; running it is theirs.
- Invoking and re-surfacing each backend's own `check --read-data` / `kopia snapshot verify` / `borg check` as if it were snapback's verification — out. The backend marking its own homework may not sit under either rung's banner. (A plain "last self-check reported by the tool" field in status is allowed, labelled self-reported.)
- Any claim that matching hashes prove the original was ever correct. The documented claim is: "independent confirmation that your backups faithfully captured what was there."

### 4.4 Provider cost classes

Each provider declares, per operation, a cost class the ladder uses to plan and to warn: `local-metadata`, `remote-metadata`, `remote-data`, `extract-required` (Borg-style: the file cannot be read without materialising the archive item). Rung 1 refuses anything above `remote-metadata`; rung 2 accepts all and reports the class in its cost estimate.

### 4.5 Open design question (decide before Stage 6 planning)

Compare instances only against each other (pairwise, no state) or additionally against a recorded **baseline** in the state directory, Tripwire-style, which would catch "they all changed together"? Default for this revision: pairwise only; a baseline is a later, separately gated feature. Record the decision in `ORCHESTRATOR.md` when it is made.

## 5. Virtual render-time consistency labels

When rung 1 finds a disagreement for a file, the FUSE mount can **show it in the directory listing itself**, so the inconsistency is visible where the user is standing rather than buried in a report.

Rules:

- Labels are **virtual**: they exist only inside the snapback FUSE mount and are computed when a directory is rendered. Nothing is written to live data, to any repository, or to any persistent file the user owns. Because they are ephemeral, they can never be stale or leak.
- Input is the rung 1 comparison for the directory being rendered, from the metadata cache (§8.4); rendering never triggers rung 2 and never fetches file data. If the comparison is not cached, the directory renders unlabelled and the comparison is queued (bounded by `consistency.render_budget_entries`, default 500 per listing).
- Presentation is configurable; the default is the least surprising to `cp` and shell completion:

```yaml
consistency:
  labels: marker          # off | marker | suffix | xattr | all
  marker_name: _inconsistent   # rendered as .snapshot/<instance>/<ts>/_inconsistent/ listing the disagreeing files as symlinks to their copies
  suffix: " (disagrees)"  # appended to the rendered name in suffix mode; original name still resolves
  xattr: user.snapback.consistency   # in-mount only (Linux user.* / macOS com.apple.metadata:_kMDItemUserTags equivalent is NOT used)
```

- `marker` mode adds one synthetic directory entry per snapshot directory and never renames real entries; `suffix` mode renames only inside the mount and keeps the original name resolvable; `xattr` mode sets the attribute on the *virtual* file inside the mount. `all` enables the three together. Finder tags and xattrs on real files are forbidden by §1 rule 2.
- Web UI Versions panel shows the same state with the same wording as §4.1.

## 6. Machine-composable output

- Every command supports `--json`; the daemon's HTTP API returns the same documents. Schemas are versioned (`"schema": "snapback.consistency/1"`) and additive-only within a major.
- Exit codes are stable and documented in `docs/exit-codes.md`:
  `0` success/all agree · `1` usage error · `2` config invalid · `3` instance unreachable · `10` disagreements found · `11` missing-in-instance found · `12` unknown/uncomparable only · `20` budget exhausted before completion · `30` FUSE prerequisite missing (existing `doctor` case).
- Minimum fields for a consistency/integrity report: `schema`, `generated_at`, `root`, `instances[]` (id, type, cost_class, status), `pairs[]`, `files[]` (path, snapshot_time, state, per-instance {size, mtime, identity | hash}, suggested_restore[] as plain commands), `summary` counts, `truncated` flag.
- Output must be line-stable for `grep`/`jq`; `--json-lines` emits one file record per line for large roots.
- These documents are the contract on which users build their own alerting, ticketing or healing. Snapback supplies the truth; the user owns the action.

## 7. Ignore files

Two layers, both about the **live tree only**; they never hide anything that is inside a backup.

### 7.1 Global ignore (config)

```yaml
ignore:
  file_name: .snapignore        # configurable, like link_name; alternatives: .snapback-ignore
  patterns:                     # gitignore syntax, evaluated relative to each root
    - .git/
    - node_modules/
    - target/
    - build/
    - dist/
    - "**/.cache/"
    - "**/*.tmp"
```

`ignore.patterns` **supersedes** the Rev 2 `exclude_relative_paths` defaults (`.git`, `node_modules`, `target`, `build`, `dist`, `_*`): the old key still loads and is merged into `patterns` with a deprecation warning. Roots may add `ignore_patterns:` of their own, appended after the global list.

### 7.2 Per-directory marker (`.snapignore`)

- **Presence** of a file named `ignore.file_name` in a directory excludes that directory **and everything below it**. The file may be empty.
- **Contents**, if any, are gitignore-style patterns relative to that directory, applied in addition to the presence rule's subtree exclusion for *sibling* matching from the parent. (Practically: an empty `.snapignore` in `big-data/` ignores `big-data/**`; a `.snapignore` in `projects/` containing `*/node_modules/` and `scratch/` ignores those under `projects/` while `projects/` itself stays served.)
- The marker is discovered during seeding, by the watcher (FAN_CREATE/FSEvents on the marker name), and by the on-access hook before any handler work. A newly created marker causes the existing `.snapshot` links beneath it to be **unregistered and removed** by the ownership registry (they are application-owned, so removing them is allowed); the directory tree itself is never touched.

### 7.3 What "ignored" means, in every subsystem

An ignored directory is never: seeded with a link · watched · served by the on-access hook (fail-open, return immediately) · offered a `.snapshot` entry in the FUSE mount · included by `snapback snap` (refused with exit 1 and a message naming the marker) · included in rung 1/2 comparisons · pre-warmed · counted against the inode budget.

Rules:

- Snapback **never creates, edits or deletes** an ignore file (§1 rule 2). `snapback ignore add PATH` prints the exact file it *would* create and asks the user to create it, or offers to do so only with an explicit `--write` flag which is documented as the one command that writes a user-visible file.
- Ignoring applies to the live tree. Browsing `.snapshot/…/big-data/` from a parent that is served still works if the backup contains it: ignore controls where snapback *acts*, not what history exists.
- `snapback status` lists ignored directories with the reason (`global-pattern: node_modules/` or `marker: /Users/alex/work/big-data/.snapignore`), and `snapback doctor` warns when a seed path is entirely ignored.
- If the ignore file name collides with `link_name`, config validation fails.

## 8. FUSE, cache and metadata-cache configuration

All keys have OS-appropriate defaults; the setup wizard never asks for them. They exist so that the latency gate findings and power users can tune without a rebuild. Every key is validated; unknown keys are an error, not ignored.

### 8.1 FUSE

```yaml
fuse:
  implementation: auto          # auto | go-fuse | (future) cgofuse
  allow_other: false            # Linux: needs user_allow_other in /etc/fuse.conf; doctor checks
  default_permissions: true
  attr_timeout: 30s             # kernel attribute cache; snapshots are immutable so long is safe
  entry_timeout: 30s
  negative_timeout: 5s          # short: a missing snapshot may appear on next refresh
  max_read: 1MiB
  max_background: 12
  direct_io: false
  kernel_cache: true            # file contents inside a snapshot never change
  daemon_timeout: 600s          # macOS macFUSE: avoid the kernel giving up during a cold Drive read
  volume_name: Snapback         # macOS Finder sidebar name
  noappledouble: true           # macOS: suppress ._* lookups
  mount_options: []             # raw passthrough, validated against a known list, logged at mount
```

Rules: `attr_timeout`/`entry_timeout` apply to snapshot contents (immutable); the catalog level (`.snapshot/` itself, `latest`, instance folders) uses `catalog.refresh_interval` semantics and is invalidated on refresh, not by these timeouts. `doctor` reports the effective options and flags combinations known to hurt (e.g. `direct_io: true` with `kernel_cache: true`).

### 8.2 Data cache

```yaml
cache:
  dir: /Users/alex/Library/Caches/snapback        # snapback-owned; separate from each backend's own cache
  max_size: 10GiB
  max_age: 30d
  eviction: lru                 # lru | lfu
  readahead: 4MiB               # sequential reads inside a snapshot file
  min_free_disk: 5%             # stop caching, keep serving, when the volume drops below this
  per_instance:                 # optional overrides
    restic-gdrive: {max_size: 20GiB}
```

The Restic instance's own `cache_dir`/`no_cache` keys stay per instance (that is Restic's metadata/pack cache). `cache:` is snapback's file-content cache in front of any provider, so a second `cp` of the same version is local. The hot-repository (`restic copy`) idea remains on the roadmap and is not this key.

### 8.3 Provider I/O limits

```yaml
providers:
  defaults:
    list_timeout: 20s           # per instance; on expiry the instance renders absent with a status entry
    open_timeout: 60s
    max_concurrent_reads: 4
    retry: {attempts: 3, backoff: 500ms}
  per_instance:
    restic-gdrive: {list_timeout: 45s, max_concurrent_reads: 2}
```

### 8.4 Metadata cache

Extends the Rev 2 `catalog:` block. The metadata cache is persistent on disk and survives daemon restarts; it holds snapshot lists, tree listings, per-file attributes and rung 1 comparison results.

```yaml
catalog:
  refresh_interval: 60s
  prewarm_snapshots: 2          # existing
  prewarm_concurrency: 2        # existing
  probe_concurrency: 4          # existing
  presence_cache_entries: 10000 # existing
  presence_cache_ttl: 5m        # existing
  metadata:
    dir: <state_dir>/metadata
    engine: sqlite              # single-file, WAL mode
    max_size: 2GiB
    tree_ttl: 0                 # 0 = immutable; a snapshot's tree never changes, evict by size only
    snapshot_list_ttl: 60s      # aligned with refresh_interval
    consistency_ttl: 24h        # rung 1 results
    prewarm_depth: 3            # directory levels to pre-list on the newest N snapshots
    per_instance:
      restic-gdrive: {prewarm_snapshots: 1}
  reader_policy: {…}            # unchanged
```

Rules: a cache miss must never block a listing beyond `providers.list_timeout`; `snapback status` reports warm/cold per instance and cache hit rates; `snapback cache clear [--metadata|--data] [--instance ID]` exists and is the only way caches are dropped.

## 9. Command summary (new or changed)

| Command | Purpose |
|---|---|
| `snapback config` | rclone-style interactive menu; adds instance flow |
| `snapback instance add\|list\|show\|test\|remove\|rename` | non-interactive instance management |
| `snapback consistency [PATH] [--json]` | rung 1 report |
| `snapback integrity [PATH] --instances … [--sample\|--all] [--budget-bytes] [--json]` | rung 2 report |
| `snapback ignore list\|check PATH\|add PATH [--write]` | inspect and explain ignore decisions |
| `snapback cache stats\|clear` | cache management |
| `snapback status --json` | now includes per-instance coverage, reachability, warm/cold, ignored dirs |
| `snapback snap [PATH] [--instance ID]` | unchanged; defaults to the root's first instance |

## 10. Web UI additions

- **Configuration → Instances**: the rclone-style add/edit/test flow (§2.2).
- **Configuration → Roots**: instance checklist per root, layout selector, ignore patterns, read-only view of discovered coverage per instance ("restic-nas: 41 snapshots cover this root; restic-gdrive: no coverage — check prefix_map").
- **History / Versions panel**: instance column; consistency state per version with the §4.1 wording; "Restore copy next to original" per instance (never overwrites, existing rule).
- **Status**: per-instance reachability, warm/cold, cache stats, ignored directories with reasons, last rung 1/2 run and summary counts.
- **Advanced**: FUSE/cache/metadata keys from §8, with the same validator as the CLI.

## 11. Stage placement, spike addition, acceptance criteria

### 11.1 Addition to Stage 1 compatibility report (no new story)

Record, for the pinned Restic version, against the disposable repository: **can snapback obtain a stable per-file content identity from snapshot metadata alone (tree/node data, no data blob reads), and does that identity survive re-chunking (e.g. after a backup with different `--pack-size`/chunker params)?** Note the exact fields observed and the cold vs warm cost of obtaining them. This decides whether rung 1 can compare content identity for Restic or only size/mtime. Do the same for each later backend in its own compatibility milestone before its provider is planned.

### 11.2 Placement

| Feature | Stage |
|---|---|
| §2 config schema v2, `instances:`, `roots[].instances`, migration, `snapback instance *` (Restic only) | **2** (config + resolver stories) |
| §3 discovery matching, `layout: auto/flat/by-instance`, per-instance folders | 2 for single instance (flat); by-instance rendering in **3** |
| §7 ignore files (global + marker) | **3** (seeding/watcher/on-access/registry stories) |
| §8 FUSE/cache/metadata configuration | 2 for FUSE keys and metadata cache; data cache and provider limits in **3** |
| §2.2 rclone-style `config` menu and §10 UI pages | **4** |
| §4.1 rung 1 consistency + §6 JSON/exit codes | **new Stage 4b** "Consistency" after Web UI, before macOS proof; gated on §11.1 findings |
| §5 virtual labels | Stage 4b |
| §4.2 rung 2 integrity | Stage 4b, opt-in only |
| §2.4 Kopia, Borg, ZFS, btrfs providers | Stage **6+** each behind its own compatibility milestone; renumber Stage 6/7 accordingly |

### 11.3 Acceptance criteria (extend the Rev 2 list; numbers continue from 19)

20. A `version: 1` config loads, warns, and is rewritten to version 2 only on explicit save; the daemon never rewrites it.
21. `snapback instance add` then `snapback instance test` succeed against a disposable Restic repository with no write to the repository (verified by comparing repository file listing before and after).
22. With one instance, the mount layout is byte-identical to Rev 2 (`flat`); with two instances, `by-instance` folders appear and `.snapshot/latest` resolves to the first instance.
23. An instance with no covering snapshot shows no folder and `status --json` reports `no-coverage` for it.
24. A `prefix_map` entry naming an instance changes only that instance's resolution.
25. An unreachable instance times out at `providers.defaults.list_timeout`, renders absent, and the other instances list within their normal time (measured).
26. An empty `.snapignore` in a seeded subtree removes existing links beneath it within one watcher cycle and prevents new ones; `snap` on that path exits 1; the directory's own contents are untouched (hash of the tree before/after).
27. Global ignore patterns and per-directory patterns are reported by `snapback ignore check PATH` with the matching rule.
28. FUSE, cache and metadata keys are validated; an unknown key fails `doctor` and `config` with the key path in the message.
29. Data cache: second read of the same file version is served without a provider read (provider call counter).
30. Metadata cache survives daemon restart: first listing after restart of a pre-warmed snapshot makes zero provider list calls.
31. `snapback consistency --json` on two instances that hold the same backup returns all `agree`; after deleting one file from one repository's snapshot (done by the test with the backend's own tool, outside snapback) it returns `missing-in <instance>` and exit code 11; the output validates against the published schema.
32. `snapback integrity` refuses to run without `--sample` or `--all`, prints a byte estimate first, and stops at `--budget-bytes` with exit 20.
33. With `consistency.labels: marker`, a disagreeing file appears under `_inconsistent/` inside the mount and the live tree and both repositories are byte-identical before and after (hashes compared).
34. Documentation and UI never use "verified", "proven" or "independent" for rung 1 output (doc test, like the existing README honesty tests).

## 12. Non-goals (this revision)

- Running, scheduling, pruning, repairing or copying backups.
- A merged cross-instance timeline.
- Auto-heal or any automatic write to a repository.
- Surfacing a backend's own `check` as snapback's verification.
- Finder tags or xattrs on real files.
- Writing ignore files without an explicit `--write`.
- Hot local repository via `restic copy` (still roadmap, separate decision).

## 13. Positioning rules for README, docs site and ARCHITECTURE.md

- Lead with the in-directory `.snapshot` idea and the multi-instance view ("your USB, your NAS and your cloud copy, side by side, as plain folders").
- Never describe snapback as a "Restic web UI". Name Backrest and Zerobyte explicitly as write-path tools that snapback complements, and Proxmox Backup Server as the single-vendor comparison.
- Use "agree"/"consistency" for rung 1 and "independent integrity" only for rung 2. "Report, don't fix" appears in the README.
- Provider roadmap may now be announced (this supersedes the earlier decision to keep other backends unannounced), phrased as "planned", with Restic as the only supported backend until its stage evidence says otherwise.

## 14. Decisions still owed by the human (unchanged, collected here)

1. macOS service default: user LaunchAgent vs system LaunchDaemon.
2. Installer fallback to `~/.local/bin` when `/usr/local/bin` is unwritable or off PATH.
3. SECURITY.md 7-day acknowledgement promise.
4. Reconcile the v1.x tags with the pre-release README framing.
5. §4.5 pairwise vs baseline (needed before Stage 4b planning).
6. Free disk on the build host before Stage 2 workers start.
