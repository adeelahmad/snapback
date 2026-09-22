package installer

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestDryRunPrintsPlanWithoutNetwork(t *testing.T) {
	var requests atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	installDir := t.TempDir()
	env := dryRunEnv("Linux", "x86_64")
	env["SNAPBACK_BASE_URL"] = srv.URL
	env["SNAPBACK_INSTALL_DIR"] = installDir

	stdout, stderr, code := runInstaller(t, env)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%q", code, stderr)
	}
	for _, want := range []string{
		srv.URL + "/snapback_linux_amd64.tar.gz",
		srv.URL + "/checksums.txt",
		installDir,
	} {
		if !strings.Contains(stdout, want) {
			t.Errorf("stdout does not contain %q; stdout=%q", want, stdout)
		}
	}
	if n := requests.Load(); n != 0 {
		t.Errorf("server saw %d requests during dry-run, want 0", n)
	}
}

const defaultBaseURL = "https://github.com/adeelahmad/snapback/releases/latest/download"

func TestDryRunDefaultBaseURLIsGitHubReleases(t *testing.T) {
	stdout, stderr, code := runInstaller(t, dryRunEnv("Linux", "x86_64"))

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%q", code, stderr)
	}
	if want := defaultBaseURL + "/snapback_linux_amd64.tar.gz"; !strings.Contains(stdout, want) {
		t.Errorf("stdout does not contain %q; stdout=%q", want, stdout)
	}
	if strings.Contains(stdout+stderr, "example.invalid") {
		t.Errorf("output contains placeholder %q; stdout=%q", "example.invalid", stdout)
	}
}

func TestNextStepsDarwin(t *testing.T) {
	assertNextSteps(t, "Darwin", "arm64", "macFUSE")
}

func TestNextStepsLinux(t *testing.T) {
	assertNextSteps(t, "Linux", "amd64", "fuse3")
}

func assertNextSteps(t *testing.T, osName, arch, fuseHint string) {
	t.Helper()
	stdout, stderr, code := runInstaller(t, dryRunEnv(osName, arch))

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%q", code, stderr)
	}
	combined := stdout + stderr
	for _, want := range []string{"snapback config", fuseHint} {
		if !strings.Contains(combined, want) {
			t.Errorf("output does not contain %q; output=%q", want, combined)
		}
	}
}
