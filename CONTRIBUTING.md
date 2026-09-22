# Contributing to Snapback

Thanks for your interest in improving Snapback. This guide covers how changes
are developed, how commits are written, and how to run the checks CI runs.

## Your first change

New here? A small, self-contained change is the best way in.

1. Browse the issue tracker for the `good first issue` label; those issues carry
   enough context to start without reading the whole codebase.
2. No issue fits? Open one first from the templates in `.github/ISSUE_TEMPLATE/`
   (`bug.yml` for a defect, `feature.yml` for a concrete request) and say what
   you plan to change, so nobody duplicates the work.
3. Questions belong in Issues for now; `.github/ISSUE_TEMPLATE/config.yml` also
   links a Discussions page, which works once Discussions is enabled on the
   repository. Security reports go through a private advisory, never an issue —
   see SECURITY.md.

## Development workflow (TDD)

1. Write a failing test that describes the behaviour you want (RED).
2. Write the smallest change that makes it pass (GREEN).
3. Clean up without changing behaviour, keeping every test green (REFACTOR).
4. Open a pull request against `master` once the gate matrix passes locally.

## Commit messages

Commits follow [Conventional Commits](https://www.conventionalcommits.org/):
`<type>: <description>`. Allowed types:

- `feat` — a new feature
- `fix` — a bug fix
- `refactor` — a code change that neither fixes a bug nor adds a feature
- `docs` — documentation only
- `test` — adding or correcting tests
- `chore` — maintenance and tooling
- `perf` — a performance improvement
- `ci` — changes to CI configuration

Keep the subject line at 100 characters or fewer, and wrap body lines at 100
characters too. Explain in the body why the change is needed, not just what it
does.

## Pull requests

Pull requests target `master`. Group related commits into one candidate pull
request rather than several overlapping ones, fill in the pull request template,
and link the issue it closes. A pull request merges only when CI is green: every
job in the gate matrix has to pass on the branch as reviewers see it. If a job is
red, push a fix rather than re-running until it passes.

## Running the gate matrix locally

Run these from the repository root before pushing:

```sh
go build ./...
go test -race ./...
go vet ./...
golangci-lint run
go test ./test/installer/
mkdocs build --strict
```

`go build ./...` compiles every package, `go test -race ./...` runs the unit
suite under the race detector, and `go test ./test/installer/` checks `install.sh`
against the installer suite. `mkdocs build --strict` builds the docs site and
fails on a broken link or a page missing from the navigation; it needs MkDocs
installed (`pip install -r requirements-docs.txt`).

### Acceptance tests

The acceptance suite mounts a real FUSE filesystem, so it is behind a build tag
and an opt-in variable, and it runs on Linux with FUSE available (the `fuse3`
package and a usable `/dev/fuse`). Restic must be on `PATH`:

```sh
SNAPBACK_FUSE_TESTS=1 go test -race -tags=integration ./test/acceptance/...
```

Without `SNAPBACK_FUSE_TESTS=1` the suite skips and names the prerequisite it is
missing, so a skip is never mistaken for a pass.

## Make targets

The repo-root `Makefile` wraps the common tasks. Run `make` to list them.

```sh
make build                      # bin/snapback with version info
make install PREFIX=/usr/local  # install bin/snapback (honours DESTDIR)
make ci                         # every gate CI runs, in order
```

## Honesty in documentation

Snapback's docs describe only what the code does today. A claim in README.md,
SPEC.md, the docs site or a release note needs evidence behind it: a test, a
command a reader can run, or a recorded acceptance run. Features that are planned
are labelled as planned. If a change makes a document's claim untrue, update the
document in the same pull request.

## Contact

For anything that does not belong in a public issue, write to
`adeel [at] adeelahmad.net`.
