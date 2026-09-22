// agentic:shim

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
func (d *daemon) Stop() error { return nil }

// Kill kills the daemon without cleanup.
func (d *daemon) Kill() error { return nil }

func newEnv(t *testing.T) env {
	t.Helper()
	return env{Root: "/nonexistent-sandbox", Home: realHome, Config: realHome}
}

func runSnapback(t *testing.T, e env, args ...string) (stdout, stderr string, code int) {
	t.Helper()
	return "", "shim: runSnapback not implemented", -1
}

func recordEvidence(t *testing.T, acc string) {
	t.Helper()
}

func newRepo(t *testing.T) (repo, passwordFile string) {
	t.Helper()
	return "", ""
}

func buildFixture(t *testing.T) fixture {
	t.Helper()
	return fixture{}
}

func startDaemon(t *testing.T, e env) *daemon {
	t.Helper()
	return &daemon{}
}
