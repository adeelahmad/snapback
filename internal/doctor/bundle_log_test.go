package doctor

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/cli"
	"github.com/adeelahmad/snapback/internal/config"
)

// bundleLogFixture lays out a state directory holding a daemon log and returns
// the parent directory, the state directory and a configuration naming it.
func bundleLogFixture(t *testing.T, stateLog string) (dir, stateDir string, cfg *config.Config) {
	t.Helper()
	dir = t.TempDir()
	stateDir = filepath.Join(dir, "state")
	if err := os.MkdirAll(stateDir, 0o700); err != nil {
		t.Fatalf("MkdirAll(%q) = %v, want nil error", stateDir, err)
	}
	if err := os.WriteFile(filepath.Join(stateDir, daemonLogName), []byte(stateLog), 0o600); err != nil {
		t.Fatalf("WriteFile(daemon.log) = %v, want nil error", err)
	}
	cfg = &config.Config{
		Version:      1,
		LinkName:     ".snapshot",
		StateDir:     stateDir,
		Repositories: []config.Repository{{ID: "repoA", Repository: bundleLogURI, ResticBinary: "restic"}},
	}
	return dir, stateDir, cfg
}

// bundleLogURI is the repository URI the log lines quote, so every case also
// pins that the log still goes through redactForBundle.
const bundleLogURI = "sftp://user@example.invalid/repo"

// writeBundleLog runs writeBundleFor over cfg and returns the exit code and the
// members of the single bundle it wrote.
func writeBundleLog(t *testing.T, cfg *config.Config, dir string) (int, map[string]string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	env := cli.Env{Stdout: &stdout, Stderr: &stderr, Getenv: func(string) string { return "" }}
	checks := []Check{{Name: "daemon", Status: statusOK, Detail: "running"}}
	code := writeBundleFor(env, checks, cfg, dir)
	if code != 0 {
		return code, nil
	}
	matches, err := filepath.Glob(filepath.Join(dir, "snapback-bundle-*.tar.gz"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("Glob(snapback-bundle-*.tar.gz) = %v, %v, want exactly 1 match", matches, err)
	}
	return code, readBundle(t, matches[0])
}

// TestBundleLogReadsConfiguredLoggingFile pins that a configured logging.file
// is the source of the bundle's daemon.log, that the state-directory file is
// not, and that the log still loses the repository URI.
func TestBundleLogReadsConfiguredLoggingFile(t *testing.T) {
	const stateLine = "state-dir log line\n"
	dir, _, cfg := bundleLogFixture(t, stateLine)

	logDir := filepath.Join(dir, "logs")
	if err := os.MkdirAll(logDir, 0o700); err != nil {
		t.Fatalf("MkdirAll(%q) = %v, want nil error", logDir, err)
	}
	logFile := filepath.Join(logDir, "d.log")
	body := "daemon started\nopening repository " + bundleLogURI + "\n"
	if err := os.WriteFile(logFile, []byte(body), 0o600); err != nil {
		t.Fatalf("WriteFile(%q) = %v, want nil error", logFile, err)
	}
	cfg.Logging = config.Logging{File: logFile}

	code, members := writeBundleLog(t, cfg, dir)
	if code != 0 {
		t.Fatalf("writeBundleFor(logging.file = %q) = %d, want 0", logFile, code)
	}
	got, ok := members["daemon.log"]
	if !ok {
		t.Fatalf("bundle members = %v, want one named %q", memberNames(members), "daemon.log")
	}
	want := "daemon started\nopening repository " + bundleRedactedMarker + "\n"
	if got != want {
		t.Errorf("bundle daemon.log = %q, want %q (the configured logging.file, redacted)", got, want)
	}
	if strings.Contains(got, strings.TrimSuffix(stateLine, "\n")) {
		t.Errorf("bundle daemon.log = %q, want it not to contain the state-dir log %q", got, stateLine)
	}
}

// TestBundleLogFallsBackToStateDirectory pins that without logging.file the
// bundle keeps reading daemon.log from the state directory.
func TestBundleLogFallsBackToStateDirectory(t *testing.T) {
	const stateLine = "state-dir log line\n"
	dir, _, cfg := bundleLogFixture(t, stateLine)

	code, members := writeBundleLog(t, cfg, dir)
	if code != 0 {
		t.Fatalf("writeBundleFor(no logging.file) = %d, want 0", code)
	}
	got, ok := members["daemon.log"]
	if !ok {
		t.Fatalf("bundle members = %v, want one named %q", memberNames(members), "daemon.log")
	}
	if got != stateLine {
		t.Errorf("bundle daemon.log = %q, want %q (the state-directory log)", got, stateLine)
	}
}

// TestBundleLogMissingConfiguredFileIsNoted pins that a logging.file that does
// not exist still writes a valid bundle and states the absence in place of the
// log, rather than dropping the member without a word.
func TestBundleLogMissingConfiguredFileIsNoted(t *testing.T) {
	dir, _, cfg := bundleLogFixture(t, "state-dir log line\n")
	missing := filepath.Join(dir, "logs", "absent.log")
	cfg.Logging = config.Logging{File: missing}

	code, members := writeBundleLog(t, cfg, dir)
	if code != 0 {
		t.Fatalf("writeBundleFor(logging.file = %q) = %d, want 0", missing, code)
	}
	got, ok := members["daemon.log"]
	if !ok {
		t.Fatalf("bundle members = %v, want one named %q stating the log is missing",
			memberNames(members), "daemon.log")
	}
	if !strings.Contains(got, missing) {
		t.Errorf("bundle daemon.log = %q, want it to name the configured path %q", got, missing)
	}
	if !strings.Contains(strings.ToLower(got), "missing") {
		t.Errorf("bundle daemon.log = %q, want it to state that the log is missing", got)
	}
}
