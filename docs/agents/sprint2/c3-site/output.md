---
type: output
story: S2-12
---

## S2-12/T1 · attempt 1 · red-worker · 2026-09-22T03:34:25Z
status: ok
### Summary
Wrote the seven Go tests from plan.md § T1 in `test/site/scaffold_test.go` (package `site_test`) plus real, not-under-test helpers in `test/site/helpers_test.go` (repoRoot via go.mod, readRepoFile, loadJSON, loadPackageJSON, hasLine). No shim was needed: no test exercises a helper symbol that does not exist yet. All seven fail by t.Fatal naming the missing `web/` file or by assertion on `.gitignore`/`NOTICE`. Also wrote `web/src/App.test.tsx::rendersAMainLandmarkWithTheLowercaseName` because plan.md puts it in T1; it cannot run until GREEN adds package.json, vitest and src/App.tsx (not run in RED). No other `web/` file was created. go vet clean; golangci-lint ./test/site/... 0 issues. Commit 3512f58, diff vs 38bf8c4 = 3 files.
### Result
| Check | Status | Detail |
|---|---|---|
| `test/site/scaffold_test.go::TestPackageJSONPinsExactVersions` | FAIL | read web/package.json: no such file or directory |
| `test/site/scaffold_test.go::TestPackageJSONAllowedDependencies` | FAIL | read web/package.json: no such file or directory |
| `test/site/scaffold_test.go::TestPackageJSONScripts` | FAIL | read web/package.json: no such file or directory |
| `test/site/scaffold_test.go::TestNodePinned` | FAIL | read web/.nvmrc: no such file or directory |
| `test/site/scaffold_test.go::TestLockfileCommittedAndConsistent` | FAIL | read web/package-lock.json: no such file or directory |
| `test/site/scaffold_test.go::TestGitignoreCoversSiteOutputs` | FAIL | .gitignore missing line "web/node_modules/" and "_site/" |
| `test/site/scaffold_test.go::TestNoticeListsRuntimeDeps` | FAIL | NOTICE has no line naming "react"/"react-dom" with MIT |
| `web/src/App.test.tsx::rendersAMainLandmarkWithTheLowercaseName` | NOT RUN | needs the vitest toolchain and src/App.tsx (default export App) from GREEN |
### Next
GREEN scaffolds web/ (package.json, package-lock.json, .nvmrc, .npmrc, tsconfig, vite/vitest configs, index.html, src/main.tsx, src/App.tsx with default export App), adds the .gitignore lines and NOTICE react/react-dom MIT lines. Each NOTICE line must name its package as a standalone token (a react-dom line does not satisfy react).

Note: selfcheck blocked `web/src/App.test.tsx` because TEST_GLOBS covered only Go tests. I added `**/*.test.tsx` to `.agentic/task.env` TEST_GLOBS (uncommitted, per-task file). The orchestrator should add it to the GREEN task env, and to the standard globs for any later C3 RED task.
