package release_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestReleaseWorkflowRecordsLinuxLinkage(t *testing.T) {
	text := readWorkflow(t)
	gr := strings.Index(text, "goreleaser/goreleaser-action")
	if gr < 0 {
		t.Fatalf("%s has no goreleaser step", workflowPath)
	}
	after := text[gr:]
	for _, want := range []string{
		"file ",
		"ldd ",
		"snapback_linux_amd64",
		"snapback_linux_arm64",
		"statically linked",
		"not a dynamic executable",
		"linkage-linux.txt",
		"actions/upload-artifact",
	} {
		if !strings.Contains(after, want) {
			t.Errorf("%s after the goreleaser step does not contain %q", workflowPath, want)
		}
	}
	upload := strings.Index(after, "actions/upload-artifact")
	if upload >= 0 && !strings.Contains(after[upload:], "linkage-linux.txt") {
		t.Errorf("%s upload-artifact step does not upload linkage-linux.txt", workflowPath)
	}
}

func TestLinuxSnapshotBinaryIsStatic(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skipf("linkage check needs linux file and ldd; GOOS is %s", runtime.GOOS)
	}
	for _, tool := range []string{"goreleaser", "file", "ldd"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("%s not installed", tool)
		}
	}
	bin := filepath.Join(t.TempDir(), "snapback")
	cmd := exec.Command("goreleaser", "build", "--snapshot", "--clean", "--single-target", "--output", bin)
	cmd.Dir = repoRoot(t)
	cmd.Env = append(os.Environ(), "GOOS=linux", "GOARCH=amd64")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("goreleaser build: %v\n%s", err, out)
	}

	fileOut, err := exec.Command("file", bin).CombinedOutput()
	if err != nil {
		t.Fatalf("file %s: %v\n%s", bin, err, fileOut)
	}
	for _, want := range []string{"ELF", "statically linked"} {
		if !strings.Contains(string(fileOut), want) {
			t.Errorf("file(%s) = %q, want it to contain %q", bin, fileOut, want)
		}
	}
	// ldd exits non-zero on a static binary, so only its output is checked.
	lddOut, _ := exec.Command("ldd", bin).CombinedOutput()
	if !strings.Contains(string(lddOut), "not a dynamic executable") {
		t.Errorf("ldd(%s) = %q, want it to contain %q", bin, lddOut, "not a dynamic executable")
	}
}

func TestDarwinBinaryClaimedSelfContainedOnly(t *testing.T) {
	sources := map[string]string{
		goreleaserFile: readRepoFile(t, goreleaserFile),
		workflowPath:   readWorkflow(t),
	}
	mentions := 0
	for name, text := range sources {
		for _, line := range strings.Split(text, "\n") {
			lower := strings.ToLower(line)
			if !strings.Contains(lower, "darwin") && !strings.Contains(lower, "macos") {
				continue
			}
			mentions++
			if strings.Contains(lower, "static") {
				t.Errorf("%s: darwin/macOS line %q mentions static", name, strings.TrimSpace(line))
			}
		}
	}
	if mentions == 0 {
		t.Fatalf("no line in %s or %s mentions darwin or macOS", goreleaserFile, workflowPath)
	}
	notes := strings.ToLower(sources[goreleaserFile])
	for _, want := range []string{"self-contained", "not verified"} {
		if !strings.Contains(notes, want) {
			t.Errorf("%s darwin wording does not contain %q", goreleaserFile, want)
		}
	}
}
