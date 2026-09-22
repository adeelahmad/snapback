---
type: tasks
story: S2-02
---

# S2-02 tasks — Pure `projection` catalog model

Story intent and owned files: `docs/agents/sprint2/stories.md` (S2-02). Every task touches only `internal/projection/**`. Tests per task are in `plan.md`; gate commands are in `validate.md`. Sources: SPEC.md §4 (`projection`: pure catalog model, immutable generations, no FUSE library), §6 (each refresh publishes a new immutable generation atomically; preserve filename bytes), §7 (full IDs under `snapshots` are canonical; stable inode identities; EROFS for mutations).

Sequencing: T1 -> T2 -> T3 -> T4 -> T5 -> T6, strictly serial. T4, T5 and T6 share `internal/projection/catalog.go`; T4 creates `fixture_test.go`, which T5 and T6 reuse without editing it. No task may run in parallel with another.

Matrix rule (M-005): T1 to T5 GREEN run with `GATE_RUN_MATRIX=0` plus the package-scoped commands in `validate.md`. Only T6, the story's last task, runs the full standards matrix.

Every test is a pure in-package unit test (`package projection`) or a `go list` subprocess. No test mounts anything, touches the filesystem outside `t.TempDir()`, or needs FUSE.

Decisions fixed here so workers do not guess:

- Package `github.com/adeelahmad/snapback/internal/projection`, standard library only. It must not import `internal/mount` or any `hanwen/go-fuse` package.
- **Cross-story contract (binding, from `docs/agents/sprint2/s2-01-mount-adapter/tasks.md`).** `*Generation` exposes exactly these exported methods, with built-in types only, so it satisfies `mount.Catalog` structurally without importing `mount`:
  - `Lookup(parent uint64, name string) (ino uint64, isDir bool, found bool)`
  - `ReadDir(dir uint64) (names []string, found bool)` — names in ascending byte order (`sort.Strings`); `found` is false for a missing inode or a non-directory.
  - `Readlink(ino uint64) (target string, found bool)` — target bytes verbatim; `found` is false for a missing inode or a non-symlink.
  - `const RootIno uint64 = 1`, declared locally in `projection` (same value as `mount.RootIno`; not imported).
- Spec types: `type Spec struct { Dirs []Dir; Links []Link }` (the root's contents), `type Dir struct { Name string; Dirs []Dir; Links []Link }`, `type Link struct { Name, Target string }`. `func Build(spec Spec) (*Generation, error)`.
- Errors: exported sentinels `ErrInvalidName` and `ErrDuplicateName`. Build wraps them with `fmt.Errorf("projection: %q: %w", path, err)`, where `path` is the slash-joined position of the offending node (for example `docs/a\x00b`). On error Build returns a nil `*Generation`.
- Valid names: non-empty, not `.` or `..`, no `/`, no NUL byte. Everything else (Unicode, upper case, leading dot, spaces, 64-hex snapshot IDs) is accepted byte-for-byte, with no case folding or Unicode normalization (§6).
- Duplicate detection covers the union of a directory's `Dirs` and `Links` names (a dir and a link with the same name collide). The same name in different parents is allowed.
- Inodes: root is `RootIno`. Other nodes are numbered from 2 upward in depth-first preorder, visiting each directory's children (dirs and links together) in ascending byte order of name. So the same spec always yields the same inode per path (coherent enumeration across a refresh, §7) and inodes never change within a generation.
- Immutability: Build deep-copies the spec. A `*Generation` has no exported fields and no mutating methods. `ReadDir` returns a fresh slice on every call. A `*Generation` is safe for concurrent reads with no locks, because nothing writes to it after Build returns. The only way to change the namespace is to Build a new generation (§6).
- EROFS at the model level: the model has no mutation operation to fail. The adapter (S2-01 `attr.go`, S2-04 `fs.go`) returns `EROFS` for every mutation. T6's contract test pins the read-only surface by asserting that the exported method set of `*Generation` is exactly the three catalog methods.
- Full snapshot IDs: names are matched exactly. A 64-hex ID entry is never found by a short prefix, and never truncated in `ReadDir` or `Readlink`.
- Files stay under 400 lines and functions under 50 lines.

## T1 — Package skeleton and dependency guard

Create `internal/projection/doc.go` containing only the package doc comment and `package projection`. Add the deps test to `deps_test.go`. Tests (RED): the deps test only. RED shims: none. With no non-test Go file in the directory, `go list -deps ./internal/projection` fails, and the test reports that through `t.Fatalf` with the command output (the S2-01 T1 pattern).

Files: `internal/projection/doc.go`, `internal/projection/deps_test.go`.

## T2 — Name validation

Add unexported `func validateName(name string) error` and the exported sentinel `ErrInvalidName`. It returns `fmt.Errorf("%q: %w", name, ErrInvalidName)` for the empty string, `.`, `..`, any name containing `/`, and any name containing NUL. Otherwise it returns nil. RED shims (`name.go`): `var ErrInvalidName = errors.New("invalid name")` and `validateName` returning nil, so the rejection rows fail by assertion.

Files: `internal/projection/name.go`, `internal/projection/name_test.go`.

## T3 — Spec types and Build validation

Add `Spec`, `Dir`, `Link`, `RootIno`, `ErrDuplicateName`, the unexported node representation, the `Generation` struct (no exported fields), and `Build`. Build walks the spec recursively (split the walk into small helpers to stay under 50 lines), calls `validateName` on every dir and link name, rejects duplicate sibling names, deep-copies the tree, and assigns inodes as fixed above. Errors are wrapped with the node's path. RED shims (`spec.go`, `build.go`): the types and constants as declared, `ErrDuplicateName = errors.New("duplicate name")`, and `Build` returning `&Generation{}, nil`, so every rejection test fails by assertion.

Files: `internal/projection/spec.go`, `internal/projection/build.go`, `internal/projection/build_test.go`.

## T4 — Lookup, inode identity and generation independence

Implement `(*Generation).Lookup` in `catalog.go`: it returns the child's inode and whether it is a directory; `found` is false for an unknown parent inode, a parent that is a symlink, or a missing name. Create the shared fixture helper `fixtureSpec()` in `fixture_test.go`. It holds: root `docs/` (containing symlink `readme-link` -> `../target`), `snapshots/` (containing symlink `<fullID>` -> `../ids/<fullID>`), `empty/` (no children), and symlinks `rel` -> `a/b`, `up` -> `../outside`, `abs` -> `/var/tmp/x`, `latest` -> `snapshots/<fullID>`. `const fullID` is 64 lowercase hex characters. Declare root entries in non-sorted order so that sorting is really exercised later. RED shims (`catalog.go`): `Lookup` returning `0, false, false`.

Files: `internal/projection/catalog.go`, `internal/projection/fixture_test.go`, `internal/projection/lookup_test.go`.

## T5 — Ordered directory listing

Implement `(*Generation).ReadDir` in `catalog.go`: for a directory inode it returns a new slice of child names in ascending byte order and `found=true`, including an empty non-nil slice for an empty directory. For a symlink, an unknown inode, or 0 it returns `nil, false`. Store children pre-sorted at Build time and copy them on each call, so map iteration order can never leak. RED shims (`catalog.go`): `ReadDir` returning `nil, false`.

Files: `internal/projection/catalog.go`, `internal/projection/readdir_test.go`.

## T6 — Readlink, the catalog contract, and the full matrix

Implement `(*Generation).Readlink` in `catalog.go`: it returns the stored target unchanged (no `filepath.Clean`, no resolution, no trimming) with `found=true` for a symlink inode, and `"", false` for a directory, an unknown inode, or 0. Add `contract_test.go` with a test-local interface `catalog` holding the three exact signatures, plus the compile-time assertion `var _ catalog = (*Generation)(nil)`. That interface is not imported from `mount`. RED shims (`catalog.go`): `Readlink` returning `"", false`. This is the story's last task: after GREEN, run the full standards matrix and the package coverage check (at least 80%) from `validate.md` Final sign-off.

Files: `internal/projection/catalog.go`, `internal/projection/readlink_test.go`, `internal/projection/contract_test.go`.
