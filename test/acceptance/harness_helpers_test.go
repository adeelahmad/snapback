//go:build integration

package acceptance

import "testing"

// env is a sandboxed environment for one snapback invocation.
type env struct {
	Root   string
	Home   string
	Config string
}

// fixture is the §20 restic repo built by buildFixture.
type fixture struct {
	Repo         string
	PasswordFile string
	Root         string
	IDs          map[string]string
}

// daemon is a running snapback daemon started by startDaemon.
type daemon struct{}

// Stop stops the daemon gracefully.
func (d *daemon) Stop() error {
	panic("SUB-AGENT-TODO: T2 startDaemon handle Stop/Kill per tasks.md T2 — stop the running snapback daemon process gracefully")
}

// Kill kills the daemon without cleanup.
func (d *daemon) Kill() error {
	panic("SUB-AGENT-TODO: T2 startDaemon handle Stop/Kill per tasks.md T2 — kill the running snapback daemon process without cleanup")
}

func newEnv(t *testing.T) env {
	t.Helper()
	panic("SUB-AGENT-TODO: T2 TestMain sandbox HOME per tasks.md T2 — build a sandboxed env{Root, Home, Config} rooted under t.TempDir(), never under the real HOME")
}

func runSnapback(t *testing.T, e env, args ...string) (stdout, stderr string, code int) {
	t.Helper()
	panic("SUB-AGENT-TODO: T2 runSnapback(t, args...) per tasks.md T2 — exec snapbackBin with args, env.Home as HOME, capture stdout/stderr and exit code")
}

func recordEvidence(t *testing.T, acc string) {
	t.Helper()
	panic("SUB-AGENT-TODO: T2 recordEvidence per tasks.md T2 — write SNAPBACK_EVIDENCE_DIR/<acc>.json with acc, status, skip_reason, kernel, commit, restic_version, gofuse_version, fuse3_version")
}

func newRepo(t *testing.T) (repo, passwordFile string) {
	t.Helper()
	panic("SUB-AGENT-TODO: T2 newRepo(t) per tasks.md T2 — restic init with a temp password file, return the repo path and password file path")
}

func buildFixture(t *testing.T) fixture {
	t.Helper()
	panic("SUB-AGENT-TODO: T2 buildFixture(t) per tasks.md T2 — produce the §20 fixture set with restic backup --host/--time/--tag, return fixture{Repo, PasswordFile, Root, IDs}")
}

func startDaemon(t *testing.T, e env) *daemon {
	t.Helper()
	panic("SUB-AGENT-TODO: T2 startDaemon(t) per tasks.md T2 — start the snapback daemon against env e, return a handle with Stop/Kill")
}
