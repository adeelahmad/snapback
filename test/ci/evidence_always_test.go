package ci_test

import (
	"strings"
	"testing"
)

// fuseLinuxJob returns the text of the fuse-linux job in ci.yml.
func fuseLinuxJob(t *testing.T) string {
	t.Helper()
	lines := strings.Split(readCI(t), "\n")
	start := -1
	for i, l := range lines {
		if strings.TrimSpace(l) == "fuse-linux:" {
			start = i
			break
		}
	}
	if start < 0 {
		t.Fatalf("%s: fuse-linux job not found", ciWorkflowPath)
	}
	base := indentOf(lines[start])
	end := start + 1
	for end < len(lines) && (strings.TrimSpace(lines[end]) == "" || indentOf(lines[end]) > base) {
		end++
	}
	return strings.Join(lines[start:end], "\n")
}

// namedStep returns the fuse-linux step whose name line is `- name: <name>`.
func namedStep(t *testing.T, job, name string) string {
	t.Helper()
	for _, s := range stepsContaining(job, "name: "+name) {
		if strings.Contains(strings.SplitN(s, "\n", 2)[0], "name: "+name) {
			return s
		}
	}
	t.Fatalf("fuse-linux step %q not found", name)
	return ""
}

// Acceptance evidence must be collected even when the integration step fails,
// so the acceptance step and the steps it depends on run unless cancelled.
func TestFuseLinuxAcceptanceRunsUnlessCancelled(t *testing.T) {
	job := fuseLinuxJob(t)
	for _, name := range []string{"Build snapback for acceptance", "Start the user systemd manager", "Acceptance tests"} {
		if step := namedStep(t, job, name); !strings.Contains(step, "!cancelled()") {
			t.Errorf("fuse-linux step %q has no !cancelled() condition:\n%s", name, step)
		}
	}
	if step := namedStep(t, job, "Upload v0.1 evidence"); !strings.Contains(step, "always()") {
		t.Errorf("fuse-linux step %q lost its always() condition:\n%s", "Upload v0.1 evidence", step)
	}
}

// The job must still fail when any step fails.
func TestFuseLinuxNoContinueOnError(t *testing.T) {
	if job := fuseLinuxJob(t); strings.Contains(job, "continue-on-error") {
		t.Errorf("fuse-linux job contains continue-on-error; a failing step must fail the job")
	}
}
