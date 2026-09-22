---
type: intake
sprint: 2
---

# Intake — Stage 1: Compatibility milestone

## Restated request

Deliver Stage 1 ("Compatibility milestone") of the Snapback spec (`SPEC.md`, Revision 2, §22 row 1), building on the completed Stage 0 scaffolding (CI, docs site, `install.sh` skeleton, `v1.0.1` release and `snapback version` all exist): pin dependencies; stand up a disposable Restic repository and verify `restic mount --path-template ids/%I` against the pinned Restic version; build a tiny directory/symlink FUSE catalog on Linux and macOS using go-fuse behind the `mount` adapter interface (§5); run a metadata-fidelity check confirming mtimes shown through `.snapshot/<timestamp>/` match the backed-up values (§5); measure rclone/Google Drive cold/warm listing and read latency (§11 "Stage 1 measurements"); and run the crawler-safety test (`rg`, `rg -L`, `fd -L`, `fd`, `find`, `rsync -a`, VS Code search over a seeded tree, §7, §20 acceptance 12). Exit evidence per §22 row 1 is "numbers recorded in the report; go/no-go on latency" — the orchestrator rule is that these numbers land in a file, not a prose sentence.

## 1. What's wanted (intent)

- Dependencies pinned: `github.com/hanwen/go-fuse/v2` added to `go.mod`/`go.sum` (§5 "FUSE adapter"); the installed Restic and rclone binary versions recorded and treated as the versions everything else in Stage 1 is verified against (§5 "Restic invocation": "Verify the option against the pinned Restic version").
- A disposable (throwaway, test-only) Restic repository — local, not the human's production backend — used to exercise `restic mount --path-template ids/%I` and confirm `%I` resolves to the full snapshot ID as the spec requires (§5).
- A minimal FUSE catalog: directories and symlinks only, no full projection/resolver/registry logic yet — implemented behind the `mount` adapter interface described in §4 ("OS FUSE adapter for the history catalog (go-fuse behind an interface)") and §5, running on both Linux and macOS.
- A metadata-fidelity check: for files reachable through the tiny catalog and a Restic mount of the disposable repo, verify displayed mtime/size/mode match the values recorded at backup time, per platform (§5: "Verify per platform in stage 1 that displayed mtimes ... match the backed-up values; ctime and birth time may be approximated ... and should be documented, not claimed").
- Real-backend latency numbers: cold listing, warm (pre-warmed) listing, cold first-file read, and warm listing after daemon-equivalent restart, all measured "over the real rclone/Google Drive backend" (§11 "Stage 1 measurements", §21 "Cold browsing latency over `rclone:gdrive`" risk row).
- A crawler-safety test: run `rg`, `fd`, `find`, `rsync -a` (must produce zero catalog hits) and `rg -L`, `fd -L`, VS Code search with follow-symlinks (must be throttled/denied and visible in status, once the reader-policy exists) over a seeded tree, and record hit counts (§7 "Stage 1 includes a test that runs `rg`, `rg -L`, `fd -L` and a VS Code search over a seeded tree and measures catalog hits"; §20 acceptance 12).
- All the above numbers recorded in a durable report artifact (not narrated only in a status update), because the exit evidence is explicitly "numbers recorded in the report."
- A go/no-go decision on latency: whether the measured Google Drive cold/warm numbers are acceptable for the product-feel bar the spec cares about (§21: "If opening `.snapshot/latest/` takes seconds, the product feel collapses"), or whether the fallback — deeper pre-warm, or pulling forward the Stage 1+ "local hot-cache repository" roadmap item (§23) — must be flagged before Stage 2 begins.

## 2. Constraints

Fixed by the spec (must follow as written, cited):

- **No overlay/union/passthrough FUSE layer, ever, over the live tree** (§1 "Non-negotiable design rules"). The Stage 1 catalog is a *separate* FUSE mount (the history catalog), never mounted over or in place of the live directory.
- **`SnapshotProvider`/`mount` seam is architectural, not optional**: the FUSE adapter must sit behind the `mount` adapter interface so a substitute adapter can be swapped in without touching `resolver`, `projection`, `links`, `discovery`, `daemon` or `web` (§4, §5). Restic access goes only through the installed `restic` CLI via `exec.CommandContext` with argument arrays — never a shell string, never `restic/internal/...`, never a fork (§5).
- **go-fuse is the pinned FUSE library** (`github.com/hanwen/go-fuse/v2`), used behind the adapter interface; only if it demonstrably fails to pass on macOS does the spec allow substituting a different tested adapter (cgofuse, noted as adding cgo) behind the same interface — and that failure/substitution must be documented, never silently swapped (§5).
- **A real macOS compatibility test is required** before any macOS support claim: "Require a real macOS compatibility test (macFUSE, Apple Silicon) before claiming macOS support" (§5). This environment has macFUSE installed on Apple Silicon (see Environment facts), so the local macOS half of that test is directly runnable here.
- **`--path-template ids/%I` must be verified against the pinned Restic version**, and `%I` (full snapshot ID) must never be replaced with a short-ID convention anywhere in the design (§5: "Never identify a snapshot by a short prefix alone").
- **Static Linux builds**: `CGO_ENABLED=0` for the Go core on Linux, verified with `ldd`/`file` before calling any binary static (§5 "Static builds") — this already exists from Stage 0's cross-compile job; adding a cgo-dependent FUSE fallback later would need to be reflected in build docs per §5, but is out of scope unless go-fuse fails on macOS.
- **Metadata-fidelity scope is explicit**: only mtime, size and mode are asserted true-to-backup; ctime/birth time may be FUSE-approximated and must be *documented*, never claimed as accurate (§5).
- **Stage 1 measurements are fixed to four specific numbers** — cold listing, warm listing, cold first-file read, warm listing after restart — "all over the real rclone/Google Drive backend" (§11). No other backend (a local path, a different cloud) satisfies this per §21's identical framing ("Cold browsing latency over `rclone:gdrive`").
- **Crawler-safety test is specified precisely**: which tools (`rg`, `rg -L`, `fd -L`, VS Code search) and what is measured (catalog hits) is fixed by §7; the fuller reader-policy enforcement (deny lists, rate limiting, status surfacing) is `catalog.reader_policy` config (§12) which does not exist until Stage 3 — so at Stage 1 the test can *measure* hits but cannot yet assert throttling/denial behavior that depends on unbuilt config.
- **Honesty gate** (§1, §18): no claim of "cross-platform," "production-ready," "static," or "Finder-integrated" without linked evidence — applies to how the Stage 1 report characterizes the FUSE catalog and metadata fidelity.
- **Never a Restic fork or `restic/internal` import** (§5) — pinning go-fuse is fine; pinning/vendoring Restic internals is explicitly forbidden.
- **Fixtures for later acceptance tests come from §20**, but Stage 1 itself only needs enough of a disposable repo/tree to exercise path-template resolution, metadata fidelity and the crawler test — the full §20 fixture set (three snapshots, two hosts, Unicode, symlinks, collisions, etc.) is Stage 2+'s acceptance-test fixture, not a Stage 1 deliverable.

## 3. Failure scenarios (what must not happen)

- The Stage 1 report states latency/metadata/crawler results only in narrative prose (a chat message or PR description) with no artifact in the repo — this fails the orchestrator's explicit "numbers recorded in the report" rule.
- `restic mount --path-template ids/%I` is verified against a Restic version other than the one actually pinned/recorded for this project, or the verification is skipped and simply assumed from documentation.
- The FUSE catalog is implemented as a direct go-fuse dependency inside `daemon`/`resolver`/other packages instead of behind the `mount` adapter interface — this recreates exactly the coupling §4/§5 warn against and makes a later adapter swap a refactor instead of a new package.
- The catalog (even in its tiny Stage 1 form) is mounted as, or effectively behaves as, an overlay on top of the live directory rather than a wholly separate mount reached only via `.snapshot` — violates the Non-negotiable design rule in §1.
- Metadata-fidelity results conflate ctime/birth time with mtime/size/mode, or claim ctime/birth time are accurate when FUSE only approximates them — violates §5's explicit fidelity/documentation split.
- Latency numbers are measured against a local path, a different cloud remote, or a synthetic/mocked "Google Drive" instead of the real configured `gdrive:` rclone remote — this does not satisfy §11/§21's "real rclone/Google Drive backend" requirement and produces numbers the go/no-go decision cannot actually rely on.
- The go/no-go decision on latency is made silently by proceeding to Stage 2 without ever writing the decision down, or is fabricated without real measured numbers behind it.
- The disposable Restic repository is confused with, or accidentally points at, a real/production repository or the human's real Google Drive backup data — must stay throwaway and clearly separated from anything with real user data.
- The crawler test is run without an actually seeded tree of managed `.snapshot`-style symlinks (i.e., testing against an empty or unrepresentative directory), producing a hit count that is not evidence of anything.
- Any Stage 2+ scope creep: building the full resolver, per-directory snapshot selection, links registry, discovery modes, daemon, web UI, or config schema now instead of the narrow Stage 1 pieces — contradicts the mandate's explicit out-of-scope list and §22's staged ordering.
- Windows, an overlay/union filesystem, or a non-Restic backend gets introduced anywhere in the Stage 1 work — all three are permanently out of scope per §2 and the mandate.

## 4. Success scenarios (acceptance / done)

- `go.mod`/`go.sum` pin `github.com/hanwen/go-fuse/v2`; the installed `restic` and `rclone` versions used for verification are recorded (this environment: restic 0.19.0, rclone v1.75.0 — see Environment facts).
- A disposable local Restic repository exists (created and destroyable at will, containing no real user data), and a documented run of `restic [options] mount --path-template ids/%I BACKEND_MOUNT` against it shows `%I` resolving to the mounted directory named by the full snapshot ID, verified against the pinned Restic version.
- A tiny FUSE catalog (directories + symlinks only) runs behind the `mount` adapter interface, demonstrated mounted and browsable on at least: macOS (Apple Silicon, macFUSE — locally verifiable in this environment) and Linux (verifiable via the existing GitHub Actions `ubuntu-latest` runners, pending a `fuse3` install step that does not yet exist in `.github/workflows/ci.yml` — see Connections and Open questions).
- A metadata-fidelity report exists per platform, stating pass/fail for mtime/size/mode against the disposable repo's known backup-time values, and explicitly documenting where ctime/birth time are FUSE-approximated rather than claiming accuracy.
- A report artifact under version control (e.g. `docs/agents/sprint2/...` or wherever the build-phase agent places it) records: cold listing time, warm (pre-warmed) listing time, cold first-file read time, and warm listing time after a restart, each measured against the real configured rclone Google Drive remote, plus the explicit go/no-go verdict on whether that latency is acceptable, or whether the Stage-23 local hot-cache fallback should be pulled forward.
- A crawler-safety report records hit counts for `rg`, `fd`, `find`, `rsync -a` (expected: zero) and `rg -L`, `fd -L`, VS Code search-with-follow-symlinks (expected: nonzero, since throttling/denial logic doesn't exist until Stage 3) run over a seeded tree of catalog-style symlinks.
- The report ends with which of the above are "measured and passing," "measured and failing/no-go," or "not measurable yet because X" — mirroring the spec's final requirement-matrix habit (§22) at Stage-1 scale.

## 5. Connections (dependencies / relations to other work)

- Depends on Stage 0 (done): the empty `cmd/snapback` binary, CI, `go.mod`, and the release pipeline that Stage 1's dependency pins (go-fuse) will flow through on the next CI run.
- The `mount` adapter interface built here is the same seam Stage 5 ("macOS proof") later hardens and Stage 2's `projection`/`resolver` packages will consume — Stage 1 should not pre-build `projection`/`resolver`/`links` logic, only the minimal catalog needed to prove the FUSE adapter and metadata fidelity (§4 module table).
- The existing CI (`.github/workflows/ci.yml`) runs `test` and `cross-compile` jobs on `ubuntu-latest` with no `fuse3` install step and no real mount/browse/unmount job yet — §18's "real mount/browse/unmount tests on Linux runners and macOS runners with macFUSE" is a Stage 1+ addition to that pipeline, not something Stage 0 built. Whether Stage 1's Linux FUSE proof runs as a new CI job (needing `apt-get install fuse3`) or only as a local demonstration is one of the open questions below.
- §11's Stage 1 measurements and §21's "Cold browsing latency" risk row are the same requirement stated twice; the fallback listed if latency is unacceptable is §23's "local hot-cache repository" roadmap item — that item stays out of scope for Stage 1 itself; Stage 1 only decides whether it needs to be pulled forward.
- The crawler test here is a narrower rehearsal of full acceptance criterion 12 (§20), which in its complete form depends on `catalog.reader_policy` (§12) and the `links`/`discovery` packages that don't exist until Stage 3 — Stage 1's version measures hits only, it cannot yet test throttling/denial enforcement.
- The disposable Restic repository and its metadata-fidelity check are also the direct predecessor to Stage 2's per-directory selection and resolver work (§6), which will need a real (not disposable) provider-backed mount to validate against.
- `ORCHESTRATOR.md` already tracks Stage 1 as "planning (Sprint 2)" and records the same local-tool gap noted in Sprint 1 intake (`golangci-lint`, `staticcheck`, `govulncheck` not installed locally) — not a new blocker, carried over from Stage 0.

## In scope

- Pinning `github.com/hanwen/go-fuse/v2` in `go.mod`/`go.sum`; recording pinned Restic/rclone versions used for verification.
- Standing up and tearing down a disposable, local, test-only Restic repository.
- Verifying `restic mount --path-template ids/%I` against that disposable repo and the pinned Restic version.
- A minimal directory/symlink-only FUSE catalog behind the `mount` adapter interface, run on macOS (local) and Linux (local or CI).
- The per-platform metadata-fidelity check (mtime/size/mode) against the disposable repo, with ctime/birth-time caveats documented.
- rclone/Google Drive latency measurements: cold listing, warm listing, cold first read, warm listing after restart — against the real configured `gdrive:` remote.
- The crawler-safety hit-count test (`rg`, `rg -L`, `fd -L`, `fd`, `find`, `rsync -a`, VS Code search) over a seeded tree of catalog-style symlinks.
- A report artifact recording all of the above numbers plus an explicit go/no-go verdict on latency.

## Out of scope (later stages, per mandate)

- Stage 2+: `config` package/schema, full `SnapshotProvider` implementation beyond what compatibility testing needs, `resolver` (per-snapshot prefix mapping, per-directory selection), `links` registry, any discovery mode (seed/on-access/explicit), `daemon`, `web` UI, native service installers.
- Full `catalog.reader_policy` enforcement (deny lists, rate limiting, status surfacing) — Stage 1 measures crawler hits only, it does not build or test throttling/denial logic.
- The full §20 acceptance-test fixture set (multi-snapshot, multi-host, Unicode, collisions, etc.) — Stage 1 needs only enough fixture to exercise path-template resolution, metadata fidelity and the crawler test.
- Any overlay, union or passthrough filesystem, any backend other than Restic, and Windows support — permanently excluded per §1/§2 and the mandate, not just deferred.
- Pulling the §23 "local hot-cache repository" roadmap item forward into real implementation — Stage 1 only decides, via the go/no-go verdict, whether that item needs to move up; it does not build it.
- macOS Finder companion, on-access kernel hooks, and their acceptance tests (§20 items 18–19) — unchanged from Stage 0's out-of-scope list, still not reached.
- Writing real user/production data into any Restic repository, disposable or otherwise, during these tests.

## Open questions (blocking)

1. **Which rclone remote/path should the latency measurements use?** `rclone listremotes` on this host shows a `gdrive:` remote already configured (type `drive`, service-account authentication) among many other remotes (`google:`, `s3:`, `qnap:`, iCloud-related remotes, etc.). Is `gdrive:` the intended remote for the §11 "real rclone/Google Drive backend" measurements, and if so, what path/folder under it should the disposable Restic repository use? This is the human's real credential and Drive account — the intake agent has not read or printed the credential contents (only confirmed the remote's existence and type).
2. **May the Stage 1 measurements write a disposable Restic repository to that Google Drive account?** The spec requires real cold/warm read-and-list numbers against the actual backend (§11, §21), which means real (if small and disposable) data must be written and read back. Given `gdrive:` is a real account, explicit confirmation is needed before any write happens there, including how much data/how many objects, and whether it must be cleaned up afterward.
3. **Is a macOS-only local mount test plus a Linux mount test run in CI (via `fuse3` on GitHub's `ubuntu-latest` runners) an acceptable reading of "verify ... on Linux and macOS"?** There is no Linux machine available locally in this environment; the only Linux compute available is GitHub Actions. The existing `.github/workflows/ci.yml` has no `fuse3` install step or mount job yet — adding one is straightforward, but confirming this satisfies the Stage 1 "Linux and macOS" proof (as distinct from the fuller Stage 5 "macOS proof" and §20 acceptance 17/19 platform criteria) avoids building the wrong kind of evidence.
4. **What are the go/no-go latency thresholds?** §11 and §21 require measuring cold/warm listing and read latency and deciding "before launch" whether the local hot-cache roadmap item (§23) is needed, but the spec states no numeric threshold (e.g., no "under N seconds") anywhere in §11, §20's Performance section, or §21. Should the human set an explicit threshold now (e.g., "cold listing must complete within N seconds to be a go"), or is the go/no-go a qualitative human judgment call made after seeing the Stage 1 numbers?

## Assumptions

- **Restic and rclone versions used for Stage 1 verification are the ones already installed on this host**: `restic 0.19.0` (darwin/arm64) and `rclone v1.75.0` (darwin/arm64) — see Environment facts. Pinning these exact versions (or newer ones discovered during the build phase) into project documentation is a routine build-phase action, not an intake decision, consistent with §24 ("make routine implementation choices autonomously, pin and verify dependencies").
- **The disposable Restic repository for path-template verification, metadata fidelity, and the crawler test can be entirely local** (e.g., a local-path Restic repo in a scratch directory) — only the §11/§21 latency measurements specifically require the real rclone/Google Drive backend. This reading follows §22 row 1's own phrasing, which lists "disposable Restic repo" and "rclone/Google Drive latency measurements" as separate bullet items rather than implying every Stage 1 test needs the cloud backend.
- **go-fuse (`github.com/hanwen/go-fuse/v2`) is assumed usable on this macOS host** because macFUSE (`macfuse.fs`) is present under `/Library/Filesystems` on Apple Silicon (`darwin/arm64`) — satisfying the precondition for the "real macOS compatibility test" §5 requires, pending the build phase actually running it.
- **The crawler-safety test's "seeded tree" can be a small synthetic fixture** (a handful of directories with catalog-style symlinks) built specifically for this test, not the full §20 fixture set — consistent with Stage 1's narrower scope versus the full acceptance criteria reserved for Stage 2+.

## Environment facts (probed)

- Host: macOS, `Darwin ... 25.5.0 ... arm64` (Apple Silicon).
- `/Library/Filesystems` contains `macfuse.fs` — macFUSE is installed.
- `restic version` → `restic 0.19.0 compiled with go1.26.4 on darwin/arm64`.
- `rclone version` → `rclone v1.75.0`, `os/arch: arm64`, built with `cmount` tag.
- `go version` → `go1.27.1 darwin/arm64`; `go.mod` already declares `go 1.27` / `toolchain go1.27.1` with no dependencies pinned yet.
- `rclone listremotes` lists many configured remotes, including `gdrive:` (confirmed `type = drive`, service-account authentication) — i.e., a real Google Drive remote is already configured on this host. No remote credentials, tokens, or full config contents were printed or read beyond the `type` field needed to confirm this is a Drive remote.
- No Linux machine is available locally; `.github/workflows/ci.yml` exists and runs `test` and `cross-compile` jobs on GitHub-hosted `ubuntu-latest` runners, but currently has no `fuse3` installation step or FUSE mount test job.
- `cmd/snapback` currently contains only `main.go`/`run.go` (Stage 0's minimal `snapback version`-only binary) plus their tests; no FUSE, provider, config, or daemon code exists yet, consistent with Stage 0 being scope-limited to scaffolding.
