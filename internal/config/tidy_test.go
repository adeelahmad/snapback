package config

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"
)

func TestTidyKeepsPins(t *testing.T) {
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Skip("go toolchain not found on PATH; go mod tidy -diff needs it")
	}
	cmd := exec.Command(goBin, "mod", "tidy", "-diff")
	cmd.Dir = "../.."
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	runErr := cmd.Run()
	if got := stdout.String(); got != "" || runErr != nil {
		t.Errorf("go mod tidy -diff = %q (err %v, stderr %q), want empty output and nil error", got, runErr, stderr.String())
	}

	lines := requireLines(readRepoFile(t, "go.mod"))
	for _, dep := range pinnedDeps {
		want := dep.module + " " + dep.version
		found := false
		for _, line := range lines {
			fields := strings.Fields(line)
			if len(fields) < 2 || fields[0] != dep.module {
				continue
			}
			found = true
			if got := fields[0] + " " + fields[1]; got != want {
				t.Errorf("go.mod require %q = %q, want %q", dep.module, got, want)
			}
			if strings.Contains(line, "// indirect") {
				t.Errorf("go.mod require %q is indirect, want direct", line)
			}
		}
		if !found {
			t.Errorf("go.mod require %q missing, want %q", dep.module, want)
		}
	}
}

func TestModuleHeaderUnchanged(t *testing.T) {
	gomod := readRepoFile(t, "go.mod")
	counts := map[string]int{}
	for _, raw := range strings.Split(gomod, "\n") {
		line := strings.TrimSpace(raw)
		if line == "replace" || strings.HasPrefix(line, "replace ") || strings.HasPrefix(line, "replace(") {
			t.Errorf("go.mod has replace directive %q, want none", line)
		}
		counts[line]++
	}
	for _, want := range []string{
		"module github.com/adeelahmad/snapback",
		"go 1.27",
		"toolchain go1.27.1",
	} {
		if got := counts[want]; got != 1 {
			t.Errorf("go.mod lines equal to %q = %d, want 1", want, got)
		}
	}
}
