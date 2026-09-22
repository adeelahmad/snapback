---
type: init
story: S1-03
---

## S1-03/T1 · attempt 1 · red-worker · 2026-09-22T00:45:58Z

### Mandate
Write every T1 test bullet in `plan.md` § T1 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. install.sh does not exist yet; runInstaller runs `sh install.sh` and the assertions on output/exit code fail. runInstaller is test infra (not itself under test), so it may be real.

### Scope
#### May
- Create the test files named for T1 in `tasks.md`: `test/installer/helpers_test.go`, `test/installer/detect_test.go`
- Test helpers that a T1 test itself exercises (e.g. `repoRoot`, `readRepoFile`, `yamlBlock`, `runInstaller`) go in a marked shim `zz_agentic_shim_test.go` in the same test package (first line `// agentic:shim`) with deliberately WRONG bodies so those tests fail by assertion; helpers NOT under test may be real and live in the named helpers file.
#### May Not
- Create or edit any non-test artifact (workflow YAML, JSON config, install.sh, .goreleaser.yaml, SPEC.md/README.md, go.mod) — those are GREEN's job; touch other stories' files; suppress/skip tests (except the tool-absent `t.Skip` cases plan.md explicitly allows).

### Inputs
- `plan.md` § T1, `tasks.md` § T1 (contracts), `validate.md` § T1. Chain base: `stage-0` @ b0f2304 (S1-01 merged: module `github.com/adeelahmad/snapback`, `internal/version`).

### Acceptance
`go test ./test/...` for this story's package compiles; every new test FAILS by assertion; `go vet ./...` clean; diff vs BASE_REF = only this task's test files (+ shim); output.md block appended; selfcheck PASS.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md in first); orchestrator relays output.md back.
- red-worker: make shim bodies differ so comparison tests cannot pass by accident.
- all: errcheck flags unchecked writes — use explicit `_, _ =` discard, never nolint.

## S1-03/T2 · attempt 1 · red-worker · 2026-09-22T00:48:04Z

### Mandate
Write every T2 test bullet in `plan.md` § T2 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. install.sh is still absent; add the T2 tests to detect_test.go reusing runInstaller.

### Scope
#### May
- Create the test files named for T2 in `tasks.md`: `test/installer/detect_test.go` (append)
- Test helpers that a T2 test itself exercises (e.g. `repoRoot`, `readRepoFile`, `yamlBlock`, `runInstaller`) go in a marked shim `zz_agentic_shim_test.go` in the same test package (first line `// agentic:shim`) with deliberately WRONG bodies so those tests fail by assertion; helpers NOT under test may be real and live in the named helpers file.
#### May Not
- Create or edit any non-test artifact (workflow YAML, JSON config, install.sh, .goreleaser.yaml, SPEC.md/README.md, go.mod) — those are GREEN's job; touch other stories' files; suppress/skip tests (except the tool-absent `t.Skip` cases plan.md explicitly allows).

### Inputs
- `plan.md` § T2, `tasks.md` § T2 (contracts), `validate.md` § T2. Chain base: the story chain branch named in your prompt (T1 tests + shim present; stage-0 S1-01 merged underneath).

### Acceptance
`go test ./test/...` for this story's package compiles; every new test FAILS by assertion; `go vet ./...` clean; diff vs BASE_REF = only this task's test files (+ shim); output.md block appended; selfcheck PASS.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md in first); orchestrator relays output.md back.
- red-worker: make shim bodies differ so comparison tests cannot pass by accident.
- all: errcheck flags unchecked writes — use explicit `_, _ =` discard, never nolint.

## S1-03/T3 · attempt 1 · red-worker · 2026-09-22T00:51:15Z

### Mandate
Write every T3 test bullet in `plan.md` § T3 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. install.sh is still absent; reuse runInstaller.

### Scope
#### May
- Create the test files named for T3 in `tasks.md`: `test/installer/dryrun_test.go`
- Test helpers that a T3 test itself exercises (e.g. `repoRoot`, `readRepoFile`, `yamlBlock`, `runInstaller`) go in a marked shim `zz_agentic_shim_test.go` in the same test package (first line `// agentic:shim`) with deliberately WRONG bodies so those tests fail by assertion; helpers NOT under test may be real and live in the named helpers file.
#### May Not
- Create or edit any non-test artifact (workflow YAML, JSON config, install.sh, .goreleaser.yaml, SPEC.md/README.md, go.mod) — those are GREEN's job; touch other stories' files; suppress/skip tests (except the tool-absent `t.Skip` cases plan.md explicitly allows).

### Inputs
- `plan.md` § T3, `tasks.md` § T3 (contracts), `validate.md` § T3. Chain base: the story chain branch named in your prompt (T1 tests + shim present; stage-0 S1-01 merged underneath).

### Acceptance
`go test ./test/...` for this story's package compiles; every new test FAILS by assertion; `go vet ./...` clean; diff vs BASE_REF = only this task's test files (+ shim); output.md block appended; selfcheck PASS.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md in first); orchestrator relays output.md back.
- red-worker: make shim bodies differ so comparison tests cannot pass by accident.
- all: errcheck flags unchecked writes — use explicit `_, _ =` discard, never nolint.

## S1-03/T4 · attempt 1 · red-worker · 2026-09-22T00:53:56Z

### Mandate
Write every T4 test bullet in `plan.md` § T4 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. install.sh still absent. Tests use an httptest server serving a fake archive + checksums.txt (good vs bad checksum) per plan.md; reuse runInstaller.

### Scope
#### May
- Create the test files named for T4 in `tasks.md`: `test/installer/install_test.go`
- Test helpers that a T4 test itself exercises (e.g. `repoRoot`, `readRepoFile`, `yamlBlock`, `runInstaller`) go in a marked shim `zz_agentic_shim_test.go` in the same test package (first line `// agentic:shim`) with deliberately WRONG bodies so those tests fail by assertion; helpers NOT under test may be real and live in the named helpers file.
#### May Not
- Create or edit any non-test artifact (workflow YAML, JSON config, install.sh, .goreleaser.yaml, SPEC.md/README.md, go.mod) — those are GREEN's job; touch other stories' files; suppress/skip tests (except the tool-absent `t.Skip` cases plan.md explicitly allows).

### Inputs
- `plan.md` § T4, `tasks.md` § T4 (contracts), `validate.md` § T4. Chain base: the story chain branch named in your prompt (T1 tests + shim present; stage-0 S1-01 merged underneath).

### Acceptance
`go test ./test/...` for this story's package compiles; every new test FAILS by assertion; `go vet ./...` clean; diff vs BASE_REF = only this task's test files (+ shim); output.md block appended; selfcheck PASS.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md in first); orchestrator relays output.md back.
- red-worker: make shim bodies differ so comparison tests cannot pass by accident.
- all: errcheck flags unchecked writes — use explicit `_, _ =` discard, never nolint.

## S1-03/T5 · attempt 1 · red-worker · 2026-09-22T00:56:06Z

### Mandate
Write every T5 test bullet in `plan.md` § T5 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. install.sh absent; append the T5 install-dir fallback tests; reuse the T4 httptest fixtures.

### Scope
#### May
- Create the test files named for T5 in `tasks.md`: `test/installer/install_test.go` (append)
- Test helpers that a T5 test itself exercises (e.g. `repoRoot`, `readRepoFile`, `yamlBlock`, `runInstaller`) go in a marked shim `zz_agentic_shim_test.go` in the same test package (first line `// agentic:shim`) with deliberately WRONG bodies so those tests fail by assertion; helpers NOT under test may be real and live in the named helpers file.
#### May Not
- Create or edit any non-test artifact (workflow YAML, JSON config, install.sh, .goreleaser.yaml, SPEC.md/README.md, go.mod) — those are GREEN's job; touch other stories' files; suppress/skip tests (except the tool-absent `t.Skip` cases plan.md explicitly allows).

### Inputs
- `plan.md` § T5, `tasks.md` § T5 (contracts), `validate.md` § T5. Chain base: the story chain branch named in your prompt (T1 tests + shim present; stage-0 S1-01 merged underneath).

### Acceptance
`go test ./test/...` for this story's package compiles; every new test FAILS by assertion; `go vet ./...` clean; diff vs BASE_REF = only this task's test files (+ shim); output.md block appended; selfcheck PASS.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md in first); orchestrator relays output.md back.
- red-worker: make shim bodies differ so comparison tests cannot pass by accident.
- all: errcheck flags unchecked writes — use explicit `_, _ =` discard, never nolint.

## S1-03/T6 · attempt 1 · red-worker · 2026-09-22T00:58:19Z

### Mandate
Write every T6 test bullet in `plan.md` § T6 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. install.sh absent; static text checks on install.sh fail via t.Fatal on the missing file. "Must not contain" checks that cannot fail today must be written as planned and reported PASS-ON-RED.

### Scope
#### May
- Create the test files named for T6 in `tasks.md`: `test/installer/static_test.go`
- Test helpers that a T6 test itself exercises (e.g. `repoRoot`, `readRepoFile`, `yamlBlock`, `runInstaller`) go in a marked shim `zz_agentic_shim_test.go` in the same test package (first line `// agentic:shim`) with deliberately WRONG bodies so those tests fail by assertion; helpers NOT under test may be real and live in the named helpers file.
#### May Not
- Create or edit any non-test artifact (workflow YAML, JSON config, install.sh, .goreleaser.yaml, SPEC.md/README.md, go.mod) — those are GREEN's job; touch other stories' files; suppress/skip tests (except the tool-absent `t.Skip` cases plan.md explicitly allows).

### Inputs
- `plan.md` § T6, `tasks.md` § T6 (contracts), `validate.md` § T6. Chain base: the story chain branch named in your prompt (T1 tests + shim present; stage-0 S1-01 merged underneath).

### Acceptance
`go test ./test/...` for this story's package compiles; every new test FAILS by assertion; `go vet ./...` clean; diff vs BASE_REF = only this task's test files (+ shim); output.md block appended; selfcheck PASS.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md in first); orchestrator relays output.md back.
- red-worker: make shim bodies differ so comparison tests cannot pass by accident.
- all: errcheck flags unchecked writes — use explicit `_, _ =` discard, never nolint.

## S1-03/T7 · attempt 1 · red-worker · 2026-09-22T01:00:49Z

### Mandate
Write every T7 test bullet in `plan.md` § T7 at the exact path::fn so each FAILS BY ASSERTION (compiles, runs, assertion or t.Fatal on a missing/incorrect artifact) — never by a compile error or missing symbol. shellcheck: skip only when absent (check `command -v shellcheck`); install.sh absent so a present shellcheck fails on the missing file.

### Scope
#### May
- Create the test files named for T7 in `tasks.md`: `test/installer/shellcheck_test.go`
- Test helpers that a T7 test itself exercises (e.g. `repoRoot`, `readRepoFile`, `yamlBlock`, `runInstaller`) go in a marked shim `zz_agentic_shim_test.go` in the same test package (first line `// agentic:shim`) with deliberately WRONG bodies so those tests fail by assertion; helpers NOT under test may be real and live in the named helpers file.
#### May Not
- Create or edit any non-test artifact (workflow YAML, JSON config, install.sh, .goreleaser.yaml, SPEC.md/README.md, go.mod) — those are GREEN's job; touch other stories' files; suppress/skip tests (except the tool-absent `t.Skip` cases plan.md explicitly allows).

### Inputs
- `plan.md` § T7, `tasks.md` § T7 (contracts), `validate.md` § T7. Chain base: the story chain branch named in your prompt (T1 tests + shim present; stage-0 S1-01 merged underneath).

### Acceptance
`go test ./test/...` for this story's package compiles; every new test FAILS by assertion; `go vet ./...` clean; diff vs BASE_REF = only this task's test files (+ shim); output.md block appended; selfcheck PASS.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md in first); orchestrator relays output.md back.
- red-worker: make shim bodies differ so comparison tests cannot pass by accident.
- all: errcheck flags unchecked writes — use explicit `_, _ =` discard, never nolint.

## S1-03/T1 · attempt 1 · green-worker · 2026-09-22T01:02:30Z

### Mandate
Implement S1-03 T1 per `tasks.md` § T1 with the least change that makes exactly the T1 tests in test/installer/detect_test.go (TestAssetNameAllTargets, TestArchAliases) pass. SCAFFOLD vacuous for S1-03 (no Go symbols, no shims). Later T2-T7 tests may still fail. Keep the script POSIX and shellcheck-clean from the start (shellcheck is installed) so T6/T7 do not force rewrites. Default base URL host must be example.invalid.

### Scope
#### May
- `install.sh` (create: #!/bin/sh, set -eu, detect_os, detect_arch, asset_name, SNAPBACK_OS/SNAPBACK_ARCH overrides, dry-run prints asset name) only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T1, `plan-ready.md` § T1, `validate.md` § T1. Chain base `chain/s1-03` @ d6da6b4.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`install.sh`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S1-03/T2 · attempt 1 · green-worker · 2026-09-22T01:05:12Z

### Mandate
Implement S1-03 T2 per `tasks.md` § T2 with the least change that makes exactly the T2 tests in test/installer/detect_test.go (TestUnsupportedPlatformFails, TestUnverifiedWarning) pass. T1 tests must stay passing; keep shellcheck clean.

### Scope
#### May
- `install.sh` (edit: reject unsupported OS/arch combos incl. darwin/non-amd64-arm64 with "unsupported" on stderr, non-zero exit, install dir untouched; warn "unverified" for arm/mips/mipsle) only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T2, `plan-ready.md` § T2, `validate.md` § T2. Chain base `chain/s1-03` @ 03fc9b9.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`install.sh`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S1-03/T3 · attempt 1 · green-worker · 2026-09-22T01:06:44Z

### Mandate
Implement S1-03 T3 per `tasks.md` § T3 with the least change that makes exactly the T3 tests in test/installer/dryrun_test.go pass. T1/T2 tests must stay passing; shellcheck clean; never install FUSE or call a package manager.

### Scope
#### May
- `install.sh` (edit: dry-run prints the plan — asset, URL, install dir honouring SNAPBACK_INSTALL_DIR — with no network; next steps: `snapback config`, macFUSE on darwin / fuse3 on linux as instructions only) only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T3, `plan-ready.md` § T3, `validate.md` § T3. Chain base `chain/s1-03` @ ab3184b.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`install.sh`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S1-03/T4 · attempt 1 · green-worker · 2026-09-22T01:08:51Z

### Mandate
Implement S1-03 T4 per `tasks.md` § T4 with the least change that makes exactly the 4 T4 tests in test/installer/install_test.go pass. T1-T3 tests must stay passing; shellcheck clean; honour SNAPBACK_BASE_URL and SNAPBACK_INSTALL_DIR.

### Scope
#### May
- `install.sh` (edit: real download via curl or wget, SHA-256 verify against checksums.txt via sha256sum or shasum -a 256, tar extract, install binary; refuse on bad/missing checksum leaving install dir empty; clear stderr on download failure) only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T4, `plan-ready.md` § T4, `validate.md` § T4. Chain base `chain/s1-03` @ c603fa8.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`install.sh`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S1-03/T5 · attempt 1 · green-worker · 2026-09-22T01:10:17Z

### Mandate
Implement S1-03 T5 per `tasks.md` § T5 with the least change that makes exactly TestInstallDirFallbackHomeLocalBin in test/installer/install_test.go pass. T1-T4 tests must stay passing; shellcheck clean. Print $HOME/.local/bin as-is, not a resolved real path (macOS /var vs /private/var).

### Scope
#### May
- `install.sh` (edit: when SNAPBACK_INSTALL_DIR is unset and the system dir is not writable/appropriate, fall back to $HOME/.local/bin (print it unresolved), create it, and warn when it is not on PATH) only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T5, `plan-ready.md` § T5, `validate.md` § T5. Chain base `chain/s1-03` @ 5622689.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`install.sh`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S1-03/T6 · attempt 1 · green-worker · 2026-09-22T01:11:47Z

### Mandate
Implement S1-03 T6 per `tasks.md` § T6 with the least change that makes exactly TestStaticNoPackageManagerInstall (TestStaticPosixHeader, TestStaticNoBashisms, TestShellcheck already pass and must stay passing) pass. This is S1-03's LAST task: FULL standards matrix runs; all 17 S1-03 tests must pass. The FUSE next-steps must remain instructions for the user (SPEC §17: never installs FUSE; prints how) — reword so no line is an executable package-manager install invocation per the test's rule, but still tell users how to get macFUSE / fuse3.

### Scope
#### May
- `install.sh` (edit: make static safety checks pass without losing behaviour) only.
#### May Not
- Write/edit tests; implement later tasks; touch other files; add secrets; suppress anything.

### Inputs
- `tasks.md` § T6, `plan-ready.md` § T6, `validate.md` § T6. Chain base `chain/s1-03` @ 125f24c.

### Acceptance
Target tests PASS under `go test -race`; previously passing tests still pass; actionlint clean on any workflow touched (installed); diff within SCOPE_GLOBS=`install.sh`; output.md block; selfcheck PASS. GATE_RUN_MATRIX=0 unless this is the story's last task.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree (copy init.md/output.md/plan-ready.md in first); orchestrator relays output.md back.
- green-worker: pin actions/tools to released versions that really exist; never `latest`.

## S1-03+S1-05+S1-07/review · attempt 1 · structural-reviewer · 2026-09-22T01:16:08Z

### Mandate
Read-only structural review of what S1-03 (install.sh, test/installer), S1-05 (community files, test/community) and S1-07 (mkdocs.yml, docs-site/, requirements-docs.txt, .github/workflows/docs.yml, test/docs) added to stage-0 @ 796e05a on top of d3954c3. Check: orphans, parallel implementations, duplicates WITHIN a package (cross-package test-helper duplication is already logged as tech debt — don't re-report it as poisoning), workflow drift between docs.yml and ci/commitlint/release workflows (pinning, permissions), installer↔release naming contract (S1-03 vs S1-06 .goreleaser.yaml), leftover markers, any non-Restic backend or banned honesty word in user-facing files (README, docs-site, community files).

### Scope
#### May
- Read; ctx-symbols/grep/go/actionlint/shellcheck read-only; append review block.
#### May Not
- Edit anything.

### Acceptance
verdict clean|isolated|foundation-poisoning; findings file:line; selfcheck (known _test.go false positive may be reported, not fixed). Finish within 5 minutes.

### Memory
- all: harness locks workers to their worktree — STORY_DIR inside your worktree.
- structural-reviewer: gate-structural-integrity misreports Go _test.go duplicates (function-local and cross-package) as HIGH — verify scope first; don't spend time re-proving it.
