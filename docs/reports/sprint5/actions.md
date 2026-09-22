# Sprint 5 action counts

This report records how many commands a user types to get from an installed
binary to a listable `.snapshot`. Every figure comes from an acceptance test
that runs the counted path end to end against a disposable Restic repository:
`TestActionCountLinux` and `TestActionCountDarwin` in
`test/acceptance/actioncount_linux_test.go` and
`test/acceptance/actioncount_darwin_test.go`. The report records results; it
makes no claim the tests do not check.

## Counting rule

One counted action is one command the user types. Each counted command sits on
a script line ending in `# action`; the test counts those lines and compares
the total with the `actions: N` line the script prints last, so the scripts
cannot drift from the number without failing.

Not counted: creating the repository, exporting `RESTIC_*`, pointing `XDG_*` at
a scratch home, taking the first backup — that is harness preparation. Reading
output is not an action, and neither is waiting for a mount or a service to
come up.

Substitution: action 1 of the real path is
`curl -fsSL https://snapback.run/install.sh | sh`. The installer is proven
separately by the installer acceptance tests, so in these scripts it is
represented by `snapback version`, the one command that shows the binary is
installed and on `PATH`. It stands in for the install step and counts as one.

## Preconditions

The counted path assumes the reader already has:

| Precondition | Why |
| --- | --- |
| FUSE installed (`fuse3` on Linux, macFUSE on macOS) | the `.snapshot` view is a FUSE mount |
| the `restic` CLI on `PATH` | snapback reads repositories through it |
| a Restic repository holding at least one snapshot of the directory | there is nothing to browse otherwise |
| `RESTIC_REPOSITORY` and `RESTIC_PASSWORD_FILE` set | `snapback setup` reads the machine instead of asking |

## Counted actions per platform

Linux — `test/acceptance/testdata/actions-linux.sh`:

| # | Script line | What it stands for |
| --- | --- | --- |
| 1 | `"$SNAPBACK_BIN" version # action` | the install step (installer proven separately) |
| 2 | `cd "$PROJECT" && "$SNAPBACK_BIN" setup $SNAPBACK_SETUP_FLAGS # action` | one `snapback setup` pass: config, seed, service |
| 3 | `entries=$(ls "$PROJECT/.snapshot") # action` | the verification that the entry resolves |

macOS — `test/acceptance/testdata/actions-darwin.sh`:

| # | Script line | What it stands for |
| --- | --- | --- |
| 1 | `"$SNAPBACK_BIN" version # action` | the install step (installer proven separately) |
| 2 | `cd "$PROJECT" && "$SNAPBACK_BIN" setup $SNAPBACK_SETUP_FLAGS # action` | one `snapback setup` pass: config and seed |
| 3 | `"$SNAPBACK_BIN" run & # action` | macOS has no launchd adapter in v0.1, so the user starts the daemon |
| 4 | `entries=$(ls "$PROJECT/.snapshot") # action` | the verification that the entry resolves |

`SNAPBACK_SETUP_FLAGS` is `--no-service` where no user service manager is
available. The bar counts `setup` alone, so those flags never add an action.
The `status` poll loop in the macOS script is the harness waiting for the mount,
not the user typing, so it is not counted.

## Measured numbers

| Platform | Counted actions | Bar | Evidence |
| --- | --- | --- | --- |
| Linux (ubuntu-latest CI, fuse3) | 3 | `linuxActionBar = 3` | CI run [35751379913](https://github.com/adeelahmad/snapback/actions/runs/35751379913), job "FUSE integration (linux)", PR #12 |
| macOS (Apple Silicon, macFUSE, local host) | 4 | `darwinActionBar = 4` | local run; evidence line `testdata/actions-darwin.sh counted 4 actions to a listable .snapshot` |

Both tests assert twice: the counted `# action` lines must not exceed the bar,
and the script's own final `actions: N` line must equal it. A script that
reached `.snapshot` in fewer or more commands than the bar fails.

## Baseline

`docs/agents/sprint3/validation/adoption-linux-path.md` walked the pre-sprint
Linux adoption path step by step and recorded **18 numbered actions** — 7 of
them automatable but not automated, and 1 outright broken (the web setup form
always failed to save). That report set 3 as the target count. The measured
Linux path is now 3. macOS is 4, one more than Linux, because there is no
launchd adapter in v0.1 and the user starts the daemon with `snapback run`.

The 18 was counted on Linux only; there is no macOS baseline to compare
against, so no before/after is claimed for macOS.

## Reproduce

```sh
SNAPBACK_FUSE_TESTS=1 go test -tags integration -run TestActionCount ./test/acceptance/
```

The tests skip when FUSE, restic or a writable home is missing, and each names
the missing prerequisite in its skip reason.
