package installer

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const (
	planPrefix    = "installing snapback "
	changeDirHint = "set SNAPBACK_INSTALL_DIR or pass --dir to change the location"
	verifiedLine  = "checksum verified: " + linuxAmd64Asset
)

// dryRunOnLocalServer runs the installer in dry-run mode against a local
// server that answers 404, so no request can ever reach the network.
func dryRunOnLocalServer(t *testing.T, installDir string) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	env := dryRunEnv("Linux", "x86_64")
	env["SNAPBACK_BASE_URL"] = srv.URL
	env["SNAPBACK_INSTALL_DIR"] = installDir

	stdout, stderr, code := runInstaller(t, env)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%q", code, stderr)
	}
	return stdout
}

// lineIndex returns the index of the first line carrying want, or -1.
func lineIndex(lines []string, want string) int {
	for i, line := range lines {
		if strings.Contains(line, want) {
			return i
		}
	}
	return -1
}

// TestPlanLinePrecedesDownloadPlan pins that the installer says what it is
// about to do — version, platform and destination — before it names any URL
// it would fetch.
func TestPlanLinePrecedesDownloadPlan(t *testing.T) {
	installDir := t.TempDir()
	stdout := dryRunOnLocalServer(t, installDir)
	lines := strings.Split(stdout, "\n")

	plan := lineIndex(lines, planPrefix)
	if plan < 0 {
		t.Fatalf("stdout has no %q line; stdout=%q", planPrefix, stdout)
	}
	want := planPrefix + "latest for linux/amd64 to " + installDir
	if got := strings.TrimSpace(lines[plan]); got != want {
		t.Errorf("plan line = %q, want %q", got, want)
	}
	if url := lineIndex(lines, "url: "); url < 0 || plan >= url {
		t.Errorf("plan line at %d does not precede the download line at %d; stdout=%q", plan, url, stdout)
	}
	if hint := lineIndex(lines, changeDirHint); hint < 0 {
		t.Errorf("stdout has no %q line; stdout=%q", changeDirHint, stdout)
	}
}

// TestDryRunAnnouncesPlanOnly pins that the dry run announces the plan exactly
// once and never claims a verification it did not perform.
func TestDryRunAnnouncesPlanOnly(t *testing.T) {
	stdout := dryRunOnLocalServer(t, t.TempDir())

	if got := strings.Count(stdout, planPrefix); got != 1 {
		t.Errorf("stdout carries %d plan lines, want exactly 1; stdout=%q", got, stdout)
	}
	if strings.Contains(stdout, "checksum verified") {
		t.Errorf("dry run claims a checksum verification; stdout=%q", stdout)
	}
}

// TestInstallConfirmsChecksum pins that a successful verification is stated
// positively rather than passing silently.
func TestInstallConfirmsChecksum(t *testing.T) {
	archive := fakeArchive(t)
	srv := newReleaseServer(t, map[string][]byte{
		linuxAmd64Asset: archive,
		"checksums.txt": []byte(sha256Hex(archive) + "  " + linuxAmd64Asset + "\n"),
	})
	installDir := t.TempDir()

	stdout, stderr, code := runInstaller(t, installEnv(srv.URL, installDir))

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%q", code, stderr)
	}
	lines := strings.Split(stdout, "\n")
	verified := lineIndex(lines, verifiedLine)
	if verified < 0 {
		t.Fatalf("stdout has no %q line; stdout=%q", verifiedLine, stdout)
	}
	plan := lineIndex(lines, planPrefix)
	if plan < 0 || plan >= verified {
		t.Errorf("plan line at %d does not precede the checksum line at %d; stdout=%q", plan, verified, stdout)
	}
	if installed := lineIndex(lines, "installed snapback to "); installed >= 0 && verified >= installed {
		t.Errorf("checksum line at %d does not precede the install line at %d; stdout=%q", verified, installed, stdout)
	}
}
