package installer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const genericFuseHint = "1. Install the fuse3 package with your distribution package manager"

// writeOSRelease writes an os-release body to a temp file and returns its path,
// so a test can pin the hint for a distribution it is not running on.
func writeOSRelease(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "os-release")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write os-release: %v", err)
	}
	return path
}

// fuseHint dry-runs the installer and returns the first "Next steps:" line,
// which is the one that tells the user how to install FUSE.
func fuseHint(t *testing.T, osName, osRelease string) string {
	t.Helper()
	env := dryRunEnv(osName, "x86_64")
	if osRelease != "" {
		env["SNAPBACK_OS_RELEASE"] = osRelease
	}
	stdout, stderr, code := runInstaller(t, env)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%q", code, stderr)
	}
	_, steps, ok := strings.Cut(stdout, "Next steps:")
	if !ok {
		t.Fatalf("output has no %q section; stdout=%q", "Next steps:", stdout)
	}
	for _, line := range strings.Split(steps, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			return line
		}
	}
	t.Fatalf("next steps section is empty; stdout=%q", stdout)
	return ""
}

// TestFuseHintNamesDistroCommand pins that the installer names the command the
// user's own distribution accepts, the way `snapback doctor` already does.
func TestFuseHintNamesDistroCommand(t *testing.T) {
	for _, tc := range []struct{ name, body, want string }{
		{"debian", "ID=debian\n", "sudo apt install fuse3"},
		{"ubuntu", "ID=ubuntu\nVERSION_ID=\"24.04\"\n", "sudo apt install fuse3"},
		{"raspbian", "ID=raspbian\n", "sudo apt install fuse3"},
		{"debian_like", "ID=pop\nID_LIKE=\"ubuntu debian\"\n", "sudo apt install fuse3"},
		{"fedora", "ID=fedora\n", "sudo dnf install fuse3"},
		{"rocky", "ID=rocky\nID_LIKE=\"rhel centos fedora\"\n", "sudo dnf install fuse3"},
		{"arch", "ID=arch\n", "sudo pacman -S fuse3"},
		{"manjaro", "ID=manjaro\n", "sudo pacman -S fuse3"},
		{"alpine", "ID=alpine\n", "sudo apk add fuse3"},
		{"opensuse", "ID=\"opensuse-leap\"\n", "sudo zypper install fuse3"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			want := "1. Install fuse3: " + tc.want
			if got := fuseHint(t, "Linux", writeOSRelease(t, tc.body)); got != want {
				t.Errorf("fuse hint = %q, want %q", got, want)
			}
		})
	}
}

// TestFuseHintUnknownDistroStaysGeneric pins that an unrecognised or missing
// os-release keeps the honest generic line instead of guessing a command.
func TestFuseHintUnknownDistroStaysGeneric(t *testing.T) {
	for _, tc := range []struct{ name, osRelease string }{
		{"unknown_id", writeOSRelease(t, "ID=plan9\n")},
		{"no_id", writeOSRelease(t, "PRETTY_NAME=\"Something\"\n")},
		{"missing_file", filepath.Join(t.TempDir(), "absent")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := fuseHint(t, "Linux", tc.osRelease); got != genericFuseHint {
				t.Errorf("fuse hint = %q, want %q", got, genericFuseHint)
			}
		})
	}
}

// TestFuseHintDarwinNamesMacFUSE pins that macOS is unaffected: there is no
// distribution package manager there, only macFUSE.
func TestFuseHintDarwinNamesMacFUSE(t *testing.T) {
	want := "1. Install macFUSE: https://macfuse.github.io/"
	if got := fuseHint(t, "Darwin", writeOSRelease(t, "ID=debian\n")); got != want {
		t.Errorf("fuse hint = %q, want %q", got, want)
	}
}

// TestUsageDocumentsOSReleaseOverride pins that the override the tests rely on
// is documented, so it is a supported seam rather than a hidden variable.
func TestUsageDocumentsOSReleaseOverride(t *testing.T) {
	stdout, stderr, code := runInstallerArgs(t, map[string]string{}, "--help")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%q", code, stderr)
	}
	for _, want := range []string{"SNAPBACK_DRY_RUN", "SNAPBACK_OS_RELEASE"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("usage does not document %q; stdout=%q", want, stdout)
		}
	}
}
