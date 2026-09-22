---
name: Good first issue
about: A small, self-contained task a first-time contributor can finish in one sitting
title: 'good first issue: '
labels: ['good first issue']
---

<!-- Maintainers file this one. Fill in all three sections before you apply the label:
an issue that does not say where the code is, is not a good first issue. -->

## What

One or two sentences on the change, written so it makes sense without reading the code
first. Say what a user sees afterwards, not how to implement it.

## Where in the code

The package and the file to start in, and the one function or test that matters. If the
change touches a test as well as the code it covers, name both paths.

## How to verify

The commands that show the change works, ending with the full run:

```sh
go build ./...
go test ./...
```

Say which test should fail before the change and pass after it. Tests come first here:
write the failing test, then the code that makes it pass. `CONTRIBUTING.md` has the
workflow and the commit format, and questions are welcome on the issue itself — no
question about a first contribution is too small.
