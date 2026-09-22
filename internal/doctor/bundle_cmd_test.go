package doctor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// bundleStdoutLines returns the non-empty lines of a doctor --bundle run.
func bundleStdoutLines(out string) []string {
	return strings.Split(strings.TrimRight(out, "\n"), "\n")
}

func TestDoctorBundleWritesArchiveAndNextStep(t *testing.T) {
	f := healthyProbes(t)
	dir := t.TempDir()

	code, stdout, stderr := runCommand(t, f, []string{"--bundle", dir})
	if code != 0 {
		t.Fatalf("doctor --bundle %s = exit %d, want 0; stdout %q stderr %q", dir, code, stdout, stderr)
	}

	matches, err := filepath.Glob(filepath.Join(dir, "snapback-bundle-*.tar.gz"))
	if err != nil {
		t.Fatalf("Glob(snapback-bundle-*.tar.gz) = %v, want nil error", err)
	}
	if len(matches) != 1 {
		t.Fatalf("doctor --bundle %s wrote %v, want exactly 1 snapback-bundle-*.tar.gz", dir, matches)
	}
	path := matches[0]

	if want := "bundle: " + path; !strings.Contains(stdout, want) {
		t.Errorf("doctor --bundle stdout = %q, want it to contain %q", stdout, want)
	}
	lines := bundleStdoutLines(stdout)
	last := lines[len(lines)-1]
	if !strings.HasPrefix(last, "next: attach ") {
		t.Errorf("doctor --bundle last line = %q, want it to start with %q", last, "next: attach ")
	}
	if !strings.Contains(last, path) {
		t.Errorf("doctor --bundle last line = %q, want it to name %q", last, path)
	}
}

func TestDoctorBundleDefaultsToCurrentDirectory(t *testing.T) {
	f := healthyProbes(t)
	dir := t.TempDir()
	t.Chdir(dir)

	code, stdout, stderr := runCommand(t, f, []string{"--bundle"})
	if code != 0 {
		t.Fatalf("doctor --bundle = exit %d, want 0; stdout %q stderr %q", code, stdout, stderr)
	}
	matches, err := filepath.Glob(filepath.Join(dir, "snapback-bundle-*.tar.gz"))
	if err != nil {
		t.Fatalf("Glob(snapback-bundle-*.tar.gz) = %v, want nil error", err)
	}
	if len(matches) != 1 {
		t.Fatalf("doctor --bundle wrote %v in the current directory, want exactly 1", matches)
	}
}

func TestDoctorBundleMakesNoNetworkCalls(t *testing.T) {
	src, err := os.ReadFile("bundle_cmd.go")
	if err != nil {
		t.Fatalf("ReadFile(bundle_cmd.go) = %v, want nil error", err)
	}
	for _, pkg := range []string{"net/http", "net/url", `"net"`} {
		if strings.Contains(string(src), pkg) {
			t.Errorf("bundle_cmd.go imports %s, want no network package", pkg)
		}
	}
}

func TestDoctorUsageMentionsBundle(t *testing.T) {
	f := healthyProbes(t)
	code, stdout, stderr := runCommand(t, f, []string{"-h"})
	if code != 0 {
		t.Fatalf("doctor -h = exit %d, want 0", code)
	}
	if help := stdout + stderr; !strings.Contains(help, "-bundle") {
		t.Errorf("doctor -h = %q, want it to mention %q", help, "-bundle")
	}
}
