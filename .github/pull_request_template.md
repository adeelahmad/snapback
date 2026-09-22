## Summary

What does this change do, and why is it needed?

## Title

Use a conventional commit title in the form `type(scope): subject`,
for example `fix(restore): keep file mode on restore`. Allowed types:
feat, fix, refactor, docs, test, chore, perf, ci.

## Checklist

- [ ] RED first: a failing test was written before the implementation
- [ ] `go test -race ./...` passes locally
- [ ] docs updated (README, CONTRIBUTING, or inline) where behaviour changed
- [ ] Commits follow the conventional commit format

## Related issues

Closes #
