---
type: validate
story: S2-01
---

# S2-01 validate — PASS/FAIL rubric

Run every command from the repo root `/Users/adeelahmad/work/snapback`. A row PASSES only when the output matches the literal expected output. Anything else is a FAIL and cites the rule it breaks (`docs/agents/sprint2/standards.md`, `docs/agents/sprint2/stories.md` S2-01). No row needs FUSE, macFUSE or `/dev/fuse`.

T1 and T2 are intermediate GREENs. They run with `GATE_RUN_MATRIX=0` and only their package-scoped rows (M-005). T3 is the story's last task and runs the full matrix under "Final sign-off".

## Pre-flight

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| Toolchain | `go version` | `go version go1.27.1 <os>/<arch>` | standards.md "Detected stack" (toolchain pin) |
| Gate tools present | `command -v goimports golangci-lint govulncheck actionlint shellcheck mkdocs goreleaser` | seven paths printed, exit 0 | standards.md "Cross-cutting gates" (Local status INSTALLED) |
| Tag exists | `go list -m -versions github.com/hanwen/go-fuse/v2 \| tr ' ' '\n' \| grep -x v2.11.0` | `v2.11.0` | stories.md S2-01 (exact tag, not a pseudo-version) |
| Scope | `git diff --name-only master -- . ':!docs'` | only `go.mod`, `go.sum`, `internal/mount/mount.go`, `internal/mount/mount_test.go`, `internal/mount/gofuse/attr.go`, `internal/mount/gofuse/attr_test.go` | stories.md S2-01 owned files; sprint2/plan.md ("go.mod/go.sum belong to S2-01 only"); M-012 |

## T1 — Pin go-fuse v2.11.0 in go.mod and go.sum

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| Pin line | `grep -E 'github.com/hanwen/go-fuse/v2 ' go.mod` | one line `github.com/hanwen/go-fuse/v2 v2.11.0` (within `require`), with no `// indirect` | stories.md failure scenario (range or pseudo-version) |
| No replace | `grep -c 'replace' go.mod` | `0` | stories.md S2-01 (exact tag) |
| go.sum hashes | `grep -c '^github.com/hanwen/go-fuse/v2 v2.11.0' go.sum` | `2` (`h1:` and `/go.mod h1:`) | stories.md failure scenario (incomplete go.sum) |
| Tidy is a no-op | `go mod tidy -diff; echo "exit=$?"` | only `exit=0` | stories.md failure scenario (tidy drops an unused pin) |
| Static build | `CGO_ENABLED=0 go build ./...; echo "exit=$?"` | only `exit=0` | standards.md Build gate; SPEC.md §5 |
| Tests | `go test -race -run 'TestGoModPinsGoFuseExactTag\|TestGoSumHasGoFuseHashes\|TestGofuseDepsIncludeFsAndFuse' ./internal/mount/...` | `ok  	github.com/adeelahmad/snapback/internal/mount` and `ok  	github.com/adeelahmad/snapback/internal/mount/gofuse` | plan.md T1 |

## T2 — `internal/mount` interfaces and types

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| Tests | `go test -race -run 'TestMountDepsExcludeGoFuse\|TestOpString\|TestKindAndRootConstants\|TestFakesSatisfyInterfaces' ./internal/mount/` | `ok  	github.com/adeelahmad/snapback/internal/mount` | plan.md T2 |
| Seam intact | `go list -deps ./internal/mount \| grep -c hanwen/go-fuse` | `0` (and `go list -deps ./internal/mount \| grep -cx github.com/adeelahmad/snapback/internal/mount` prints `1`, M-002) | stories.md failure scenario (interface leaks go-fuse types); SPEC.md §5 |
| No go-fuse identifiers | `grep -nE 'fuse\.\|fs\.(Inode\|StableAttr)' internal/mount/mount.go` | no output | stories.md failure scenario (`fuse.Attr`, `fs.Inode` leak) |
| Small interfaces | `go doc -all ./internal/mount \| grep -A5 -E '^type (Catalog\|Adapter\|Observer) interface'` | Catalog shows 3 methods (`Lookup`, `ReadDir`, `Readlink`), Adapter shows 2 (`Mount`, `Unmount`), Observer shows 1 (`Observe`) | standards.md "keep interfaces small (1-3 methods)" |
| Predeclared-only Catalog | `go doc ./internal/mount Catalog` | signatures exactly as in tasks.md: `Lookup(parent uint64, name string) (ino uint64, isDir bool, found bool)`, `ReadDir(dir uint64) (names []string, found bool)`, `Readlink(ino uint64) (target string, found bool)` | stories.md S2-01 connections (S2-02 structural contract) |
| Package build and vet | `CGO_ENABLED=0 go build ./internal/mount/... && go vet ./internal/mount/...; echo "exit=$?"` | only `exit=0` | standards.md Build and Vet gates |

## T3 — `gofuse/attr.go` translation, EROFS errno and full matrix

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| Tests | `go test -race -run 'TestStableAttr\|TestAttr\|TestAttrUnknownKindHasNoMode\|TestEntryOutTimeouts\|TestAttrOutTimeout\|TestTimeoutsBounded\|TestReadOnlyErrnoIsEROFS\|TestDaemonOwner\|TestGofuseDepsIncludeFsAndFuse' ./internal/mount/gofuse/` | `ok  	github.com/adeelahmad/snapback/internal/mount/gofuse` | plan.md T3 |
| No blank imports left | `grep -c '^\s*_ "github.com/hanwen' internal/mount/gofuse/attr.go` | `0` | tasks.md T3 (real uses replace the T1 placeholder) |
| Named constants | `grep -nE 'time\.Second\|0o555' internal/mount/gofuse/attr.go` | matches only inside the `const` block defining `EntryTimeout`, `AttrTimeout`, `DirPerm`, `SymlinkPerm` | standards.md "no magic numbers" |
| No write bits | `go test -race -run 'TestAttr$\|TestTimeoutsBounded' -v ./internal/mount/gofuse/ \| grep -E '^(--- PASS\|ok)'` | `--- PASS: TestAttr`, `--- PASS: TestTimeoutsBounded`, `ok  	github.com/adeelahmad/snapback/internal/mount/gofuse` | stories.md failure scenario (write bits in attributes) |
| Tidy is a no-op | `go mod tidy -diff; echo "exit=$?"` | only `exit=0` | stories.md constraint (S2-04 never edits go.sum) |
| Package coverage | `go test -race -cover ./internal/mount/...` | each package `coverage: >=80.0% of statements` (a test-only or declaration-only package may print `[no statements]`) | `~/.claude/rules/common/testing.md` |
| Full matrix | `GATE_RUN_MATRIX=1` GREEN (runs every row of "Final sign-off") | all rows PASS | M-005 (full matrix only on the story's last GREEN) |

## Final sign-off

Run the standards matrix (`docs/agents/sprint2/standards.md` "Cross-cutting gates") verbatim. Every line must exit 0. This runs once, at T3.

| Check | Command | Expected (PASS) | FAIL cites |
| --- | --- | --- | --- |
| gofmt | `test -z "$(gofmt -l .)"` | exit 0, no output | `~/.claude/rules/golang/coding-style.md` |
| goimports | `test -z "$(goimports -l .)"` | exit 0, no output | `~/.claude/rules/golang/coding-style.md` |
| Build | `CGO_ENABLED=0 go build ./...` | exit 0, no output | SPEC.md §5, §17 |
| Vet | `go vet ./...` | exit 0, no output | SPEC.md §18 |
| Lint | `golangci-lint run` | exit 0, no output (test files included, M-004) | SPEC.md §18; `.golangci.yml` |
| Race | `go test -race ./...` | every package `ok` or `[no test files]` | SPEC.md §18, §20; golang/testing.md |
| Coverage | `go test -race -covermode=atomic -coverprofile=coverage.out ./... && go tool cover -func=coverage.out \| awk '/^total:/{gsub("%","",$NF); if ($NF+0 < 80) { print "coverage " $NF "% is below the 80% minimum (common/testing.md)"; exit 1 } else { print "coverage " $NF "% OK (>=80%)" } }'` | `coverage NN.N% OK (>=80%)` | `~/.claude/rules/common/testing.md` |
| Vulns | `govulncheck ./...` | `No vulnerabilities found.` (go-fuse v2.11.0 and its `golang.org/x/sys` pull are now in scope) | SPEC.md §18 |
| Actions lint | `actionlint` | exit 0, no output | standards.md (task brief gate) |
| Shell lint | `shellcheck -s sh install.sh` | exit 0, no output | standards.md (task brief gate) |
| Docs build | `mkdocs build --strict --site-dir site` | exit 0 | SPEC.md §18; `.github/workflows/docs.yml` |
| Release config | `goreleaser check` | exit 0 | SPEC.md §17, §18; M-007 |
| Honesty | `grep -rniE 'production-ready\|cross-platform\|finder-integrated' internal/mount` | no output | standards.md honest reporting (SPEC.md §1, §20, §22) |
| Plan ticked | `grep -c '^- \[ \]' docs/agents/sprint2/s2-01-mount-adapter/plan.md` | `0` after execution | standards.md TDD workflow |
