# Contributing to Snapback

Thanks for your interest in improving Snapback. This guide covers how changes
are developed, how commits are written, and how to run the checks CI runs.

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

## Running the gate matrix locally

Run these from the repository root before pushing:

```sh
go test -race ./...
go vet ./...
golangci-lint run
```
