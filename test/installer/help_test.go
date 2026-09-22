package installer

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

// runInstallerArgs runs `sh install.sh <args...>` the way runInstaller does,
// but with command-line arguments, so argument handling can be pinned.
func runInstallerArgs(t *testing.T, env map[string]string, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()

	script, err := filepath.Abs(installScript)
	if err != nil {
		t.Fatalf("resolve install.sh: %v", err)
	}

	cmd := exec.Command("sh", append([]string{script}, args...)...)
	cmd.Dir = filepath.Dir(script)
	cmd.Env = []string{
		"PATH=" + os.Getenv("PATH"),
		"HOME=" + t.TempDir(),
		"TMPDIR=" + t.TempDir(),
	}
	for k, v := range env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	runErr := cmd.Run()
	var exitErr *exec.ExitError
	switch {
	case runErr == nil:
		exitCode = 0
	case errors.As(runErr, &exitErr):
		exitCode = exitErr.ExitCode()
	default:
		t.Fatalf("run sh install.sh: %v", runErr)
	}
	return outBuf.String(), errBuf.String(), exitCode
}

// sandbox returns an install directory the installer must not write into and a
// base URL served by a handler that counts requests and always fails, so a
// test can prove the installer did nothing.
func sandbox(t *testing.T) (installDir string, baseURL string, requests *atomic.Int64) {
	t.Helper()
	requests = &atomic.Int64{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)
	return t.TempDir(), srv.URL, requests
}

func assertNoSideEffect(t *testing.T, installDir string, requests *atomic.Int64) {
	t.Helper()
	entries, err := os.ReadDir(installDir)
	if err != nil {
		t.Fatalf("read install dir: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("install dir has %d entries, want 0", len(entries))
	}
	if n := requests.Load(); n != 0 {
		t.Errorf("server saw %d requests, want 0", n)
	}
}

// TestHelpPrintsUsageWithoutSideEffect pins that argument parsing happens
// before any download or filesystem write: `-h`/`--help` prints usage, exits 0
// and leaves the sandboxed install dir untouched.
func TestHelpPrintsUsageWithoutSideEffect(t *testing.T) {
	for _, flag := range []string{"-h", "--help"} {
		t.Run(flag, func(t *testing.T) {
			installDir, baseURL, requests := sandbox(t)
			env := map[string]string{
				"SNAPBACK_INSTALL_DIR": installDir,
				"SNAPBACK_BASE_URL":    baseURL,
			}

			stdout, stderr, code := runInstallerArgs(t, env, flag)

			if code != 0 {
				t.Fatalf("exit code = %d, want 0; stderr=%q", code, stderr)
			}
			for _, want := range []string{
				"Usage",
				"--dry-run",
				"--dir",
				"--version",
				"--yes",
				"SNAPBACK_INSTALL_DIR",
				"SNAPBACK_DRY_RUN",
			} {
				if !strings.Contains(stdout, want) {
					t.Errorf("usage does not contain %q; stdout=%q", want, stdout)
				}
			}
			assertNoSideEffect(t, installDir, requests)
		})
	}
}

// TestUnknownArgumentIsRejected pins that an unrecognised argument names
// itself on stderr, exits 2 and changes nothing.
func TestUnknownArgumentIsRejected(t *testing.T) {
	installDir, baseURL, requests := sandbox(t)
	env := map[string]string{
		"SNAPBACK_INSTALL_DIR": installDir,
		"SNAPBACK_BASE_URL":    baseURL,
	}

	stdout, stderr, code := runInstallerArgs(t, env, "--bogus")

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stdout=%q stderr=%q", code, stdout, stderr)
	}
	for _, want := range []string{"unknown argument: --bogus", "Usage"} {
		if !strings.Contains(stderr, want) {
			t.Errorf("stderr does not contain %q; stderr=%q", want, stderr)
		}
	}
	assertNoSideEffect(t, installDir, requests)
}

// TestDryRunAndDirFlagsMatchEnvVars pins that `--dry-run --dir DIR` behaves
// exactly like SNAPBACK_DRY_RUN=1 with SNAPBACK_INSTALL_DIR.
func TestDryRunAndDirFlagsMatchEnvVars(t *testing.T) {
	installDir, baseURL, requests := sandbox(t)
	env := map[string]string{
		"SNAPBACK_BASE_URL": baseURL,
		"SNAPBACK_OS":       "Linux",
		"SNAPBACK_ARCH":     "x86_64",
	}

	stdout, stderr, code := runInstallerArgs(t, env, "--dry-run", "--dir", installDir)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%q", code, stderr)
	}
	for _, want := range []string{
		baseURL + "/snapback_linux_amd64.tar.gz",
		baseURL + "/checksums.txt",
		installDir,
	} {
		if !strings.Contains(stdout, want) {
			t.Errorf("stdout does not contain %q; stdout=%q", want, stdout)
		}
	}
	assertNoSideEffect(t, installDir, requests)
}

// TestVersionFlagSelectsReleaseTag pins that `--version VER` asks for that
// release's assets instead of the latest ones.
func TestVersionFlagSelectsReleaseTag(t *testing.T) {
	installDir := t.TempDir()
	// The tag only shows in the default base URL, so this test cannot point
	// SNAPBACK_BASE_URL at a local server. An empty PATH hides curl and wget
	// instead, so the installer can never reach the network from here.
	env := map[string]string{
		"PATH":                 "",
		"SNAPBACK_OS":          "Linux",
		"SNAPBACK_ARCH":        "x86_64",
		"SNAPBACK_INSTALL_DIR": installDir,
	}

	stdout, stderr, code := runInstallerArgs(t, env, "--dry-run", "--version", "v9.9.9")

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%q", code, stderr)
	}
	if want := "/releases/download/v9.9.9/snapback_linux_amd64.tar.gz"; !strings.Contains(stdout, want) {
		t.Errorf("stdout does not contain %q; stdout=%q", want, stdout)
	}
	if strings.Contains(stdout, "/releases/latest/") {
		t.Errorf("stdout still points at the latest release; stdout=%q", stdout)
	}
	entries, err := os.ReadDir(installDir)
	if err != nil {
		t.Fatalf("read install dir: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("install dir has %d entries, want 0", len(entries))
	}
}

// TestYesFlagIsAccepted pins that `--yes` is accepted today so a later
// confirmation step can use it.
func TestYesFlagIsAccepted(t *testing.T) {
	installDir, baseURL, requests := sandbox(t)
	env := map[string]string{
		"SNAPBACK_BASE_URL": baseURL,
		"SNAPBACK_OS":       "Linux",
		"SNAPBACK_ARCH":     "x86_64",
	}

	stdout, stderr, code := runInstallerArgs(t, env, "--yes", "--dry-run", "--dir", installDir)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stdout=%q stderr=%q", code, stdout, stderr)
	}
	assertNoSideEffect(t, installDir, requests)
}
