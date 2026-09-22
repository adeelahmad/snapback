---
type: stories
sprint: 2
---

# Sprint 2 stories — Snapback Stage 1: Compatibility milestone

Source of intent: `docs/agents/sprint2/intake.md`. Rules: `docs/agents/sprint2/standards.md` (gate matrix plus the integration-test gating section). Spec anchors: SPEC.md §4, §5, §7, §11, §20, §21, §22 row 1. Lessons applied: `docs/agents/memory.md` M-002, M-005, M-006, M-007, M-010, M-012.

## Sprint goal

Deliver Stage 1 of the spec: pin `github.com/hanwen/go-fuse/v2`, put a `mount` adapter seam and a pure `projection` catalog model in place, verify `restic mount --path-template ids/%I` against the installed restic 0.19.0 using a disposable repo, mount a tiny directory/symlink history catalog through go-fuse on macOS (local macFUSE) and Linux (new `fuse3` CI job), check metadata fidelity (mtime, size, mode) through `.snapshot/<timestamp>/`, measure the four §11 latencies over the real rclone remote `gdrive:snapback-stage1`, run the crawler hit-count test, and record every number in committed files. Exit evidence (§22 row 1): numbers recorded in `docs/reports/stage1-measurements.md` plus per-measurement JSON under `docs/reports/stage1/`, and a latency go/no-go item left open for the human.

## Sprint demo

1. `go list -m github.com/hanwen/go-fuse/v2` prints the exact pinned version (v2.11.0 unless Stage 2 planning finds a reason to change it), and `go list -deps ./internal/projection` contains no go-fuse package.
2. `SNAPBACK_FUSE_TESTS=1 go test -race -tags=integration ./internal/...` on this Mac mounts the catalog, lists it, reads a symlink, gets `EROFS` on a write, unmounts, and writes `*-darwin.json` evidence files.
3. The GitHub Actions `fuse-linux` job on `master` is green and uploads the matching `*-linux.json` evidence artifact.
4. `docs/reports/stage1/pathtemplate-darwin.json` shows `ids/<64-hex full snapshot ID>` resolving under a `restic mount --path-template ids/%I` of a disposable repo, with restic version `0.19.0`.
5. `docs/reports/stage1/latency.json` holds four numbers (cold listing, warm pre-warmed listing, cold first-file read, warm listing after restart) measured over `gdrive:snapback-stage1`, and records that the remote repo was deleted afterwards.
6. `docs/reports/stage1-measurements.md` shows every number, the three-way requirement matrix, and the open item "Latency go/no-go: PENDING — human decision". No threshold appears.

## Definition of Done

- Every story below is merged, and the full gate matrix in `docs/agents/sprint2/standards.md` § Cross-cutting gates passes on the final GREEN of each story. Intermediate GREENs run a package-scoped matrix (M-005).
- The integration suite (`-tags=integration`) passes on macOS locally with `SNAPBACK_FUSE_TESTS=1` and on Linux in the `fuse-linux` CI job. Any prerequisite-gated skip names the missing prerequisite, and the report classifies it as **implemented but not tested here**.
- Every evidence JSON file named in the owned-file lists exists, is non-empty, and is produced by running the real harness. The orchestrator copies the Linux files from the CI artifact of a green `master` run.
- The disposable Google Drive repo `gdrive:snapback-stage1` has been deleted, and `latency.json` records that deletion as verified.
- The report ends with the requirement matrix (implemented and tested / implemented but not tested here / not implemented, with reasons) and an open human go/no-go item. No latency threshold is invented anywhere.
- No docs text uses "production-ready", "cross-platform", "static" or "Finder-integrated" without linked evidence (§1, §18).
- No two stories own the same file. No Stage 2+ code exists: no `config`, `resolver`, `links`, `discovery`, `daemon`, `web`, or `provider/restic` package.
- Every external-tool config change (the CI workflow) is validated with the real tool (`actionlint`, then a real CI run) as well as tests (M-007). No tool install uses `@latest` (M-006).

## Out of scope

- Everything in Stage 2+: config schema, `SnapshotProvider` interface and its Restic implementation, resolver, per-directory snapshot selection, links registry, discovery modes, daemon, pre-warm scheduling, IPC, web UI, service installers.
- `catalog.reader_policy` enforcement (deny lists, rate limiting, status surfacing, §12). Stage 1 measures crawler hits and asserts nothing about throttling.
- Timestamp aliases, `by-date`, `info.json`, `DIRECTORY_KEY` hashing and the full §7 layout. The Stage 1 catalog holds only directories and symlinks.
- The full §20 fixture set (three snapshots, two hosts, Unicode, collisions). Stage 1 uses a minimal generated tree.
- Implementing the §23 local hot-cache repository. Stage 1 only feeds the decision on whether to pull it forward.
- A macOS CI runner with macFUSE. macOS proof in this sprint is the local host (Apple Silicon); the Stage 5 macOS proof stays separate.
- Wiring the catalog into the `snapback` CLI. `cmd/snapback` is untouched this sprint.
- Permanently excluded: overlay, union or passthrough filesystems; any backend other than Restic; Windows; importing `restic/internal/...` or forking Restic.

## User stories

### S2-01 — go-fuse pin and `mount` adapter seam (foundation)

As a maintainer, I want go-fuse pinned and a small `mount` adapter interface defined so that the FUSE library sits behind a seam that a documented substitute (e.g. cgofuse) could replace without touching any other package.

**Owned files:** `go.mod`, `go.sum`, `internal/mount/mount.go`, `internal/mount/mount_test.go`, `internal/mount/gofuse/attr.go`, `internal/mount/gofuse/attr_test.go`.

**Intent**

- **What's wanted:** `go.mod`/`go.sum` pin `github.com/hanwen/go-fuse/v2` at an exact version (v2.11.0, the newest tagged release found by `go list -m -versions`). Package `internal/mount` defines, where it is used: a read-only `Catalog` interface (lookup a child by name, list a directory, read a symlink target, stable inode per node) that `projection` satisfies structurally without importing `mount`; an `Adapter` interface (mount a `Catalog` at a directory, unmount); an `Observer` hook that receives one event per catalog operation (op kind and node path) so tests can count catalog hits; and `Entry`/`Kind` types (directory, symlink only). `internal/mount/gofuse/attr.go` holds the pure translation from `mount.Entry` to go-fuse attributes (read-only mode bits, the daemon user's uid/gid, bounded entry/attr timeouts as named constants) and the `EROFS` errno used for every mutation.
- **Constraints:** Interfaces have 1–3 methods each, following the Go coding-style rule. `internal/mount` (the interface file) imports no go-fuse package; only `internal/mount/gofuse` does. `attr.go` imports both `github.com/hanwen/go-fuse/v2/fs` and `github.com/hanwen/go-fuse/v2/fuse` so that `go mod tidy` leaves a complete `go.sum` and S2-04 never has to touch it. `CGO_ENABLED=0` builds pass. No magic numbers for timeouts or modes. Invalidation relies on bounded lifetimes, not kernel notifications (§5).
- **Failure scenarios:** go-fuse is pinned but nothing imports it, so `go mod tidy` drops it. `go.sum` is incomplete, forcing a later story to edit a file it does not own. The interface leaks go-fuse types (`fuse.Attr`, `fs.Inode`), which destroys the seam. A version range or pseudo-version is used instead of a tag. Write permission bits appear in the translated attributes.
- **Success scenarios:** A test asserts `go.mod` requires `github.com/hanwen/go-fuse/v2` at an exact `vX.Y.Z` tag. A test runs `go list -deps ./internal/mount` and asserts that it contains no `hanwen/go-fuse` path. Before asserting that, it checks that the deps list is non-empty (M-002). Table-driven tests on `attr.go` cover directory and symlink entries (mode has no write bits, correct type bits, uid/gid, timeouts equal the named constants) and assert that the mutation errno is `syscall.EROFS`. A compile-time assertion checks that a test fake satisfies `Catalog`, `Adapter` and `Observer`.
- **Connections:** Hard prerequisite for S2-04 (it implements `Adapter` using `attr.go`). S2-02 must satisfy `Catalog` structurally. That is a cross-story contract, and each side tests its own half: S2-02 asserts it against a local copy of the method set, and S2-04 asserts `projection` values pass as `mount.Catalog`. S2-07 consumes `Observer` for hit counts.

### S2-02 — Pure `projection` catalog model (foundation)

As a maintainer, I want an immutable, FUSE-independent catalog model of directories and symlinks so that the mount adapter only renders it and Stage 2 can extend it without touching FUSE code.

**Owned files:** `internal/projection/**`.

**Intent**

- **What's wanted:** Package `internal/projection` builds an immutable catalog generation from a declarative spec: a tree of named directories and symlinks with raw target strings. It supports lookup by parent and name, ordered directory listing, readlink, and a stable inode per node within a generation. The method set matches S2-01's `mount.Catalog` contract. Building a new generation never mutates an existing one (§6: generations publish atomically).
- **Constraints:** Standard library only. Importing go-fuse is forbidden (§4 module table: "Pure catalog model independent of any FUSE library"). Only directories and symlinks exist; there are no regular files and no timestamp or alias logic (Stage 2). Names containing `/`, `.`/`..`, NUL or empty strings are rejected at build time with a wrapped error. Duplicate sibling names are rejected. Symlink targets are stored byte-for-byte, never resolved or cleaned. Functions stay under 50 lines, and files stay under 400 lines.
- **Failure scenarios:** A lookup on generation N returns nodes added in generation N+1, meaning shared mutable state. Inode numbers change between two lookups of the same node in one generation. Listing order is nondeterministic (map iteration). A symlink target is `filepath.Clean`ed or resolved. A malformed name is accepted and later breaks the FUSE layer.
- **Success scenarios:** Table-driven unit tests cover valid trees and each rejected name class; lookup hit, miss, and lookup on a non-directory; sorted listings; that readlink returns the exact target bytes (including `..` and absolute targets); inode stability within a generation; and that building generation 2 from a modified spec leaves generation 1's lookups unchanged. A deps test asserts that `go list -deps ./internal/projection` is non-empty and contains no `hanwen/go-fuse` (M-002). Package coverage is at least 80%.
- **Connections:** Parallel with S2-01. Its structural contract with `mount.Catalog` is described there. S2-04, S2-06 and S2-07 build their test catalogs with it.

### S2-03 — Disposable Restic fixture and `--path-template ids/%I` verification

As a maintainer, I want a safe helper that creates and destroys a throwaway Restic repo of generated data so that I can verify `restic mount --path-template ids/%I` against restic 0.19.0 and reuse the same helper for fidelity and latency work.

**Owned files:** `internal/compat/resticfx/**`, `docs/reports/stage1/pathtemplate-darwin.json`, `docs/reports/stage1/pathtemplate-linux.json`.

**Intent**

- **What's wanted:** Package `internal/compat/resticfx` provides the following. A pinned-version check parses `restic version` and fails or skips when it is not `0.19.0`. A generated-tree writer creates files with chosen sizes, modes and mtimes and returns the expected metadata. Other functions init a repo (with a random password written to a 0600 temp file and passed via `--password-file`), back up the generated tree, list snapshots with `--json` (returning the full 64-hex IDs), start `restic mount --path-template ids/%I <dir>` as a supervised long-running process, wait for readiness, unmount, and destroy the repo. The helper also records the tool versions (restic, and rclone when used). An integration test (`//go:build integration`, gated on `SNAPBACK_FUSE_TESTS=1`, `restic` on PATH and a FUSE device or macFUSE present) runs the path-template flow. When `SNAPBACK_EVIDENCE_DIR` is set, it writes `pathtemplate-<goos>.json` containing the restic version, the full snapshot ID, the directory name observed under `ids/`, and the pass/fail result.
- **Constraints:** Every subprocess uses `exec.CommandContext` with an argument array and never a shell (§5). Timeouts apply only to finite commands; the mount process is not given a timeout (§5). The repository guard accepts only (a) a path under `os.TempDir()`/`t.TempDir()` or (b) exactly `rclone:gdrive:snapback-stage1` (human decision). It refuses anything under `$HOME` outside temp, any other remote, and any existing repo it did not create (§20). Passwords never appear in arguments, logs or evidence files (§12). There is no `restic/internal` import. Full snapshot IDs only, never short prefixes (§5). Pure logic (version parsing, guard, argument building, JSON evidence encoding) is unit-tested with a fake command runner so default-build coverage stays at 80% or higher. Integration-only code lives in `_test.go` files.
- **Failure scenarios:** The path-template check passes because `ids/` contains a short ID or any directory at all. It must match the exact 64-hex ID from `snapshots --json`. The test runs against a restic version other than 0.19.0 without saying so. The guard lets through `$HOME/…`, `gdrive:` root, or another remote. A leaked mount survives a failed test (the cleanup must unmount even on failure). The password shows up in `ps` output or evidence JSON. A skip message does not name the missing prerequisite.
- **Success scenarios:** Unit tests cover version parsing (0.19.0 accepted; 0.18.1 and garbage rejected), a guard table (temp path allowed, `rclone:gdrive:snapback-stage1` allowed, `$HOME/x`, `rclone:gdrive:`, `rclone:gdrive:other` and `/` refused), the argument arrays for init, backup, snapshots, mount and ls (asserting `--path-template`, `ids/%I`, `--password-file`, and no password value), and the evidence JSON round-trip. The integration test asserts that `<mnt>/ids/<fullID>` exists, that its listing contains the generated files, and that no other entry under `ids/` is a prefix-only name. The macOS evidence file is produced locally. The Linux file comes from the S2-05 CI artifact.
- **Connections:** Independent of S2-01 and S2-02 (stdlib only). Hard prerequisite for S2-06 (fidelity) and S2-08 (latency). It also defines the `SNAPBACK_EVIDENCE_DIR` contract, which the S2-05 job sets and uploads. It foreshadows the Stage 2 `provider/restic`, but it is test-harness code and is not that package.

### S2-04 — Tiny go-fuse history catalog (directories and symlinks) behind the adapter

As a maintainer, I want a go-fuse `mount.Adapter` that serves a `projection` catalog read-only so that there is real mount, browse and unmount evidence on macOS and Linux before any feature work.

**Owned files:** `internal/mount/gofuse/fs.go`, `internal/mount/gofuse/fs_test.go`, `internal/mount/gofuse/adapter.go`, `internal/mount/gofuse/adapter_test.go`, `internal/mount/gofuse/mount_integration_test.go`, `docs/reports/stage1/catalog-darwin.json`, `docs/reports/stage1/catalog-linux.json`.

**Intent**

- **What's wanted:** `fs.go` implements go-fuse node types for directories and symlinks over a `mount.Catalog`. It serves only lookup, getattr, readdir, readlink, statfs and lifecycle, and returns `EROFS` (from `attr.go`) for create, mkdir, unlink, rename, setattr, write and symlink creation (§7). It fires the `mount.Observer` for every lookup, readdir and readlink. `adapter.go` implements `mount.Adapter`: mount at a caller-owned empty directory as a separate mount (never over a live directory, §1), and unmount cleanly. An integration test (`//go:build integration`, gated per standards) mounts a small projection, lists it with `os.ReadDir`, reads symlinks with `os.Readlink`, asserts `EROFS` on write and mkdir, checks that inode numbers are stable across two stats, unmounts, and verifies that the mountpoint is empty afterwards. When `SNAPBACK_EVIDENCE_DIR` is set, it writes `catalog-<goos>.json` with the platform, the go-fuse version, the operations exercised, and the result.
- **Constraints:** Only `internal/mount/gofuse` imports go-fuse. There is no read path for regular-file content (the catalog has none). Mount options are read-only, and `allow_other` is off, so the mount is private to the user (§7). The mount lifetime is not tied to a timeout context (§5). Unit tests exercise node methods directly through go-fuse's in-process APIs, without a kernel mount, so default-build coverage holds. S2-04 does not edit `go.mod`/`go.sum`. If a new module requirement appears, that is an S2-01 fix task, not an edit here. If go-fuse cannot pass on macFUSE, the task stops and reports; it does not silently swap in cgofuse (§5).
- **Failure scenarios:** The mount is placed on or inside a live directory. A mutation returns `EPERM`/`ENOSYS` instead of `EROFS`. The Observer is not called on some path, which undercounts crawler hits. A failed test leaves a stale mount behind. The integration test passes on an empty listing (M-002: assert the expected entries exist first). macOS is claimed without the local macFUSE run.
- **Success scenarios:** Unit tests (default build) check that lookup, readdir and readlink return the projection's data; that each mutation method returns `EROFS`; that the Observer receives exactly one event per op; and that attributes come from `attr.go`. The integration test passes locally on macOS (Apple Silicon, macFUSE) and in the `fuse-linux` CI job, producing both evidence files. A compile-time check confirms that `*Adapter` satisfies `mount.Adapter` and that a `projection` generation passes as a `mount.Catalog`.
- **Connections:** Depends on S2-01 (interface, `attr.go`, pin) and S2-02 (catalog model). Prerequisite for S2-06 and S2-07. The Linux evidence depends on S2-05's job being merged.

### S2-05 — Linux FUSE CI job

As a maintainer, I want a CI job on `ubuntu-latest` that installs `fuse3` and a pinned restic and runs the real integration suite so that Linux mount evidence comes from an actual kernel mount, not a skip.

**Owned files:** `.github/workflows/ci.yml`, `test/ci/fuse_job_test.go`.

**Intent**

- **What's wanted:** A new `fuse-linux` job in `ci.yml` runs on push to `master` and on PRs. It installs `fuse3` via apt, and installs restic `0.19.0` from its GitHub release asset with a SHA-256 check. It also installs `ripgrep`, `fd-find` and `rsync` for the crawler test, and pins rclone only if a later story needs it (none does in CI). It sets `SNAPBACK_FUSE_TESTS=1` and `SNAPBACK_EVIDENCE_DIR`, verifies `/dev/fuse` exists, runs `go test -race -tags=integration ./...`, and uploads the evidence directory as an artifact named `stage1-evidence-linux`. It never sets `SNAPBACK_RCLONE_REMOTE`, so latency tests skip in CI. The existing `test` and `cross-compile` jobs stay unchanged.
- **Constraints:** Every action is pinned (`actions/checkout@v4`, `actions/setup-go@v5`, `actions/upload-artifact@v4`, matching the existing file). No `@latest` (M-006). The restic download URL and checksum are pinned literals. The job is validated with `actionlint` locally (it is in the gate matrix) and then proven by a real green run on `master` (M-007). There are no secrets in the job. Existing sprint-1 CI story (1-02) tests in `test/ci/` (`ci_test.go`, `matrix_test.go`, `lint_test.go`, `helpers_test.go`) must keep passing unmodified. This story adds a new test file only and reuses the existing helpers without editing them.
- **Failure scenarios:** The job installs fuse3 but never sets `SNAPBACK_FUSE_TESTS`, so every mount test silently skips and the job is green with no evidence. The job installs the distro's restic (not 0.19.0), so the path-template result does not apply to the pinned version. The artifact upload runs only on success, which hides failure evidence. It uses `if: always()` with a missing directory, and the step fails confusingly. An edit breaks an existing sprint-1 CI story (1-02) assertion. The job passes `actionlint` but fails on GitHub because of an apt package name (`fd-find` provides `fdfind`, not `fd`).
- **Success scenarios:** `test/ci/fuse_job_test.go` asserts that the `fuse-linux` job exists on `ubuntu-latest`, installs `fuse3`, downloads restic `0.19.0` and checks a 64-hex SHA-256, sets `SNAPBACK_FUSE_TESTS: "1"` and `SNAPBACK_EVIDENCE_DIR`, runs `go test` with `-race` and `-tags=integration`, uploads `stage1-evidence-linux` with `if: always()`, contains no `@latest`, and does not set `SNAPBACK_RCLONE_REMOTE`. Each negative check first asserts that the job block is non-empty (M-002). `actionlint` is clean locally. **Exit evidence:** a green `fuse-linux` run on `master` whose artifact contains the Linux JSON files, verified by the orchestrator.
- **Connections:** Independent of the code stories for its files, so it can run in wave 1. Its value depends on S2-03, S2-04, S2-06 and S2-07 integration tests landing. The orchestrator re-runs or waits for the job after those merge and copies the artifact's JSON files into the owning stories' evidence paths.

### S2-06 — Metadata-fidelity check through `.snapshot/<timestamp>/`

As a user, I want proof that files reached through a `.snapshot` alias show the mtime, size and mode recorded at backup so that "historical metadata is truthful" (§7) is evidence, not a claim.

**Owned files:** `internal/compat/fidelity/**`, `docs/reports/stage1/fidelity-darwin.json`, `docs/reports/stage1/fidelity-linux.json`.

**Intent**

- **What's wanted:** An integration test (`//go:build integration`, gated on `SNAPBACK_FUSE_TESTS=1` plus restic and FUSE probes) does the following. It uses `resticfx` to generate a tree with deliberately distinct mtimes (including sub-second and pre-2000 values), sizes (0 bytes, small, multi-MiB) and modes (0644, 0600, 0755, a read-only file). It backs the tree up to a disposable repo and starts `restic mount --path-template ids/%I`. It mounts the S2-04 catalog with a projection containing `.snapshot/<timestamp>` as a symlink to `<resticmnt>/ids/<fullID>`. Then it stats every file through `.snapshot/<timestamp>/…`. It compares mtime, size and mode against both the generation-time values and `restic ls --json`, and records the observed ctime and birth time (where the platform exposes them) as documented-not-claimed. When `SNAPBACK_EVIDENCE_DIR` is set, it writes `fidelity-<goos>.json` with a per-file expected/observed pair, a per-attribute pass/fail, the platform, and the mtime precision observed.
- **Constraints:** Only mtime, size and mode are asserted (§5). ctime and birth time are recorded and never asserted as equal. The report text calls them FUSE-approximated. Files are served by restic's mount. The catalog never proxies or restamps them (§7), and the test asserts that the resolved path lands inside the restic mount. Any mtime precision loss restic or FUSE introduces (e.g. sub-second truncation) is recorded as observed data, not hidden by widening the comparison silently. A tolerance, if one is needed, is a named constant that also appears in the evidence JSON. Pure comparison and evidence-encoding logic is unit-tested in the default build.
- **Failure scenarios:** The comparison runs against the live generated tree instead of the path through `.snapshot`, which proves nothing. The test compares against values it read back from the mount itself (circular). ctime or birth-time equality is asserted or reported as accurate. The test passes on an empty file list (M-002). The tolerance is widened to hide a real mismatch without recording it.
- **Success scenarios:** Unit tests cover the comparator (exact match passes, 1 s mtime drift fails, mode drift fails, precision recorded) and the evidence JSON round-trip. The integration test asserts that the number of files compared equals the number generated (at least 6) and that every file passes mtime, size and mode on macOS locally and on Linux in CI. The two evidence files are committed. A mismatch on a platform fails the test and is carried into the report as "measured and failing". It is never skipped.
- **Connections:** Depends on S2-03 (resticfx) and S2-04 (catalog adapter), plus S2-02 for the projection. S2-05 provides the Linux run. Its results feed S2-09.

### S2-07 — Crawler-safety hit-count test

As a maintainer, I want to measure how many catalog operations common crawlers trigger over a seeded tree so that the §7 crawler risk is quantified before reader-policy work in Stage 3.

**Owned files:** `internal/compat/crawler/**`, `docs/reports/stage1/crawler-darwin.json`, `docs/reports/stage1/crawler-linux.json`.

**Intent**

- **What's wanted:** An integration test (`//go:build integration`, gated on `SNAPBACK_FUSE_TESTS=1` plus a FUSE probe) seeds a temp tree of at least 20 directories, 3 levels deep, with regular files. Each directory has a `.snapshot` symlink into a mounted S2-04 catalog, whose snapshot entries symlink to temp "history" directories with files. An Observer counts catalog hits. The test runs each tool from a fixed table with a timeout per tool: `rg`, `rg -L`, `fd`, `fd -L` (`fdfind` on Debian/Ubuntu), `find`, `find -L`, and `rsync -a` into a temp destination. It resets the counter between tools and records hits per tool plus the tool's version. VS Code search with `search.followSymlinks` is recorded as `not-tested-here` with the reason. When `SNAPBACK_EVIDENCE_DIR` is set, it writes `crawler-<goos>.json`.
- **Constraints:** It asserts zero catalog hits for the non-following tools (`rg`, `fd`, `find`, `rsync -a`), which is §20 acceptance 12, first half. It records hit counts for the following tools (`rg -L`, `fd -L`, `find -L`) without asserting throttling or denial, because `catalog.reader_policy` does not exist until Stage 3. A missing tool binary skips that row only, with a message naming the tool, and the JSON marks it `not-tested-here` (for example, `fd` is not installed on this Mac today). Tools run via `exec.CommandContext` argument arrays inside temp dirs only, never in `$HOME` (§20). Pure pieces (tool table, fdfind fallback resolution, evidence encoding) are unit-tested in the default build.
- **Failure scenarios:** The seeded tree has no `.snapshot` links, or the links point at nothing, so zero hits is vacuous. The test first asserts that the tree has the expected link count and that `ls -L` of one link produces at least one hit (M-002). The counter is not reset between tools, which inflates later rows. A following tool loops forever through a symlink cycle and hangs the suite (hence the per-tool timeout). VS Code is reported as tested when it was not. A skipped tool is folded into a zero-hit "pass".
- **Success scenarios:** Unit tests cover the tool table, the fdfind/fd resolution, and evidence encoding. The integration test asserts the positive control (hits > 0 after a deliberate dereference), zero hits for every present non-following tool, and a recorded count for every present following tool. Evidence files for darwin (local) and linux (CI) are committed.
- **Connections:** Depends on S2-04 (catalog plus Observer) and S2-02 (projection). It needs no restic, so it is independent of S2-03. S2-05 installs `ripgrep`, `fd-find` and `rsync` for the Linux row. Results feed S2-09 and later the Stage 3 reader-policy design.

### S2-08 — rclone/Google Drive latency harness and measurements

As a maintainer, I want the four §11 latencies measured over the real `gdrive:snapback-stage1` remote so that the human can decide go/no-go on the local hot-cache roadmap item with real numbers.

**Owned files:** `internal/compat/latency/**`, `tools/stage1latency/**`, `docs/reports/stage1/latency.json`.

**Intent**

- **What's wanted:** `internal/compat/latency` orchestrates one measurement run through `resticfx`. It generates a test tree (generated data only; size and file count are named constants recorded in the JSON), creates a disposable repo at `rclone:gdrive:snapback-stage1` with a fresh empty `--cache-dir` under temp, and backs up. It then measures the following, each with `time.Since` around a single operation, repeated a small named number of times where the measurement allows, recording every sample:
  1. **Cold listing:** first `ReadDir` of the snapshot root through `restic mount --path-template ids/%I` with an empty cache.
  2. **Warm (pre-warmed) listing:** listing after `restic ls --json <fullID>` has populated the cache.
  3. **Cold first-file read:** first full read of one file's content.
  4. **Warm listing after restart:** unmount, remount with the same cache dir, list again.

  It then deletes the remote repo with `rclone purge gdrive:snapback-stage1`, verifies absence with `rclone lsf`, and writes `latency.json` with the samples, per-measurement median/min/max in milliseconds, restic and rclone versions, host OS/arch, UTC timestamp, data size, and `remote_deleted: true|false`. `tools/stage1latency` is a thin `main` that runs it and writes JSON to a path flag. There is no gated network integration test. The real run is an orchestrator-run command, `SNAPBACK_RCLONE_REMOTE=gdrive:snapback-stage1 go run ./tools/stage1latency -out docs/reports/stage1/latency.json`, executed once, and only after the human explicitly confirms that run (see sprint plan "Decisions requiring human action").
- **Constraints:** The remote is exactly `gdrive:snapback-stage1` (human decision). Any other value of `SNAPBACK_RCLONE_REMOTE` makes the harness refuse with an error naming the allowed value, and the `resticfx` guard enforces it too. The run refuses to start if the remote path already exists and is non-empty, because it only creates and deletes its own repo. Deletion runs in a deferred cleanup even when a measurement fails, and a failed delete is recorded as `remote_deleted: false` and surfaced loudly. No threshold, verdict or pass/fail on latency is computed anywhere in the code or the JSON (human decision). Every command uses `exec.CommandContext` argument arrays. The password lives only in a 0600 temp file, never in args, logs or JSON. The credential config for rclone is never read or printed. It never runs in default CI or the default matrix. `main` stays thin, and the orchestration logic is unit-tested with a fake runner and fake clock so coverage holds.
- **Failure scenarios:** The harness runs against a local path or another remote and labels the result "Google Drive". The "cold" listing is actually warm because the default user cache dir was reused. "After restart" does not restart the mount. The remote repo is left on Drive after a failure. A latency threshold or "GO/NO-GO" field is invented. The rclone `purge` target is built from user input without the guard and could delete something else. Only one sample is recorded while the file claims a median.
- **Success scenarios:** Unit tests (fake runner and clock) cover the command sequence order (init, backup, mount, list, ls-prewarm, list, read, unmount, remount, list, unmount, purge, lsf), the use of a fresh temp `--cache-dir`, that cleanup purges after an injected mid-run failure, that a refused remote (`gdrive:`, `gdrive:other`, `s3:x`) is rejected, that a non-empty pre-existing remote aborts before any write, the median/min/max math, and the JSON schema (no threshold or verdict key; `remote_deleted` present). After explicit human confirmation, the orchestrator runs `go run ./tools/stage1latency -out docs/reports/stage1/latency.json` once on this Mac and commits the file with four numbers and `remote_deleted: true`.
- **Connections:** Depends on S2-03 (resticfx, guard, versions). Independent of the FUSE catalog stories: it lists through restic's own mount, which needs macFUSE locally. Results feed S2-09 and the human go/no-go. It relates to §21 ("Cold browsing latency over `rclone:gdrive`") and §23 (the local hot-cache repository).

### S2-09 — Stage 1 measurements report and requirement matrix

As the human deciding go/no-go, I want one committed report that shows every Stage 1 number from the evidence files, with an honest requirement matrix, so that the §22 row 1 exit evidence is a file, not a chat message.

**Owned files:** `docs/reports/stage1-measurements.md`, `test/reports/**`.

**Intent**

- **What's wanted:** `docs/reports/stage1-measurements.md` has these sections: pinned versions (go-fuse from `go.mod`, restic 0.19.0, rclone v1.75.0, Go toolchain); path-template verification per platform; catalog mount/browse/unmount per platform; metadata fidelity per platform (mtime, size and mode results, with ctime and birth time explicitly marked FUSE-approximated and documented, not claimed); the four latency numbers (median plus min/max, sample count, data size, remote deleted); crawler hit counts per tool per platform, with VS Code marked not-tested-here; the three-way requirement matrix for every Stage 1 item in §22 row 1, with exact reasons; and a final open item, "Latency go/no-go: PENDING — human decision", listing the options (proceed, deeper pre-warm, or pull the §23 local hot-cache repository forward) without a threshold or recommendation presented as a verdict. A Go test package `test/reports` cross-checks the markdown against the JSON evidence files.
- **Constraints:** Every number in the report is copied from a committed JSON file, and the test enforces equality, so no hand-typed figures drift. Markdown only, placed outside `docs-site/`, so the `mkdocs build --strict` gate is unaffected. The honesty gate applies: no "production-ready", "cross-platform", "static" or "Finder-integrated". A macOS result is labelled "macOS (Apple Silicon, macFUSE, local host)", not "macOS support". Any platform or tool with missing evidence is listed as "implemented but not tested here" or "not implemented", with the reason, and is never omitted. The test package reads evidence files; it does not edit them (they are owned by S2-03, S2-04, S2-06, S2-07 and S2-08).
- **Failure scenarios:** The report narrates results without numbers. A number in the report disagrees with its JSON. A GO/NO-GO verdict or threshold is written by an agent. VS Code, `fd` or Linux fidelity is shown as tested when its evidence says `not-tested-here` or is missing. ctime is described as accurate. The test passes vacuously because it found no JSON files or no numbers (M-002).
- **Success scenarios:** `test/reports` asserts the following: every expected evidence file exists and parses, and there are at least 9 (pathtemplate, catalog, fidelity and crawler for 2 platforms, plus latency), unless a file is explicitly listed as not-tested-here in the matrix. The go-fuse version in the report equals `go.mod`. Each of the four latency medians in the markdown equals `latency.json`. Crawler counts match per tool and platform. Fidelity pass/fail matches. The matrix contains all §22 row 1 items. `remote_deleted: true` is shown. The pending human-decision line is present. No numeric threshold or verdict word ("GO", "NO-GO" as a decision, "acceptable") appears outside the options list. Banned honesty words are absent. Positive-content checks run before negative ones.
- **Connections:** Depends on every other story's evidence (S2-03, S2-04, S2-06, S2-07, S2-08) and on the S2-05 CI run for the Linux files. It is the §22 row 1 exit artifact. The human go/no-go recorded after it gates Stage 2 planning.

## Story dependency graph

```
Wave 1 (parallel, no shared files):
  S2-01 go-fuse pin + mount seam ----+
  S2-02 projection model ------------+--> S2-04 go-fuse catalog adapter (wave 2) --+--> S2-06 metadata fidelity (wave 3) --+
  S2-03 resticfx + path-template ----+----------------------------------------------+                                      |
                                     +--> S2-08 latency over gdrive (wave 2) ------------------------------------------------+
                                            S2-04 --> S2-07 crawler hit counts (wave 3) --------------------------------------+
  S2-05 fuse-linux CI job ...... (soft: runs every integration test above; supplies *-linux.json) ..........................+
                                                                                                                             v
                                                                               S2-09 Stage 1 report (wave 4) --> HUMAN go/no-go
```

Hard edges: S2-01 -> S2-04; S2-02 -> S2-04; S2-03 -> S2-06; S2-04 -> S2-06; S2-04 -> S2-07; S2-03 -> S2-08; {S2-03, S2-04, S2-06, S2-07, S2-08} -> S2-09.

Soft edges (textual contracts, no shared file): S2-01 `mount.Catalog` and S2-02's method set; the `SNAPBACK_EVIDENCE_DIR` contract, defined by S2-03 and used by S2-04, S2-06 and S2-07, which S2-05 sets and uploads; S2-05 -> the Linux evidence files of S2-03, S2-04, S2-06 and S2-07 (the orchestrator copies them from the green `master` artifact).

File ownership is disjoint: `internal/mount/gofuse/` is split by exact file between S2-01 (`attr.go`, `attr_test.go`) and S2-04 (`fs.go`, `fs_test.go`, `adapter.go`, `adapter_test.go`, `mount_integration_test.go`). `docs/reports/stage1/` is split by exact file. `test/ci/` gains only `fuse_job_test.go` (S2-05), and the sprint-1 CI story (1-02) files there are read-only for this sprint.
