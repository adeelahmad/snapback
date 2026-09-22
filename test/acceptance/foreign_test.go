//go:build integration

package acceptance

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestAcc08ForeignEntriesSurvive(t *testing.T) {
	recordEvidence(t, "acc08")
	requireLinkPrereqs(t)
	e := newEnv(t)
	root := filepath.Join(e.Root, "work")
	proj := filepath.Join(root, "proj")
	mkdirs(t, root, "proj/a", "proj/b", "proj/c", "proj/d")
	hist := writeLinkConfig(t, e, linkConfig{Root: root, SeedPath: "proj", MaxDepth: 2})

	fileBody := []byte("foreign bytes\n")
	foreignFile := filepath.Join(proj, "a", ".snapshot")
	if err := os.WriteFile(foreignFile, fileBody, 0o644); err != nil {
		t.Fatalf("write foreign file: %v", err)
	}
	const foreignTarget = "/tmp/elsewhere"
	foreignLink := filepath.Join(proj, "b", ".snapshot")
	if err := os.Symlink(foreignTarget, foreignLink); err != nil {
		t.Fatalf("symlink foreign link: %v", err)
	}

	checkForeign := func(step string) {
		t.Helper()
		got, err := os.ReadFile(foreignFile)
		if err != nil {
			t.Errorf("after %s: read proj/a/.snapshot: %v, want the foreign file intact", step, err)
		} else if !bytes.Equal(got, fileBody) {
			t.Errorf("after %s: proj/a/.snapshot = %q, want %q", step, got, fileBody)
		}
		if target, err := os.Readlink(foreignLink); err != nil || target != foreignTarget {
			t.Errorf("after %s: readlink proj/b/.snapshot = %q, %v, want %q", step, target, err, foreignTarget)
		}
	}

	steps := []struct {
		name      string
		args      []string
		wantOwned bool
		service   bool
	}{
		{"seed", []string{"seed", proj}, true, false},
		{"links repair", []string{"links", "repair"}, true, false},
		{"service uninstall", []string{"service", "uninstall"}, true, true},
		{"links remove --managed", []string{"links", "remove", "--managed"}, false, false},
	}
	for _, s := range steps {
		if s.service {
			if reason := serviceManagerMissing(); reason != "" {
				t.Logf("skipping step %q: %s", s.name, reason)
				continue
			}
		}
		_, stderr, code := runSnapback(t, e, s.args...)
		if s.service && code != 0 && strings.Contains(stderr, "unsupported_service_manager") {
			t.Logf("skipping step %q: snapback reports no supported service manager (stderr: %s)", s.name, stderr)
			continue
		}
		if code != 0 {
			t.Errorf("snapback %s exit = %d, want 0 (stderr: %s)", s.name, code, stderr)
		}
		checkForeign(s.name)
		owned := ownedLinks(t, root, hist)
		switch {
		case s.wantOwned && len(owned) == 0:
			t.Errorf("after %s: owned links = none, want > 0", s.name)
		case !s.wantOwned && len(owned) != 0:
			t.Errorf("after %s: owned links = %v, want none", s.name, owned)
		}
	}
}

// serviceManagerMissing names why no supported service manager (systemd, the
// only one v0.1 manages) is available here, or returns "" when one is.
func serviceManagerMissing() string {
	if runtime.GOOS != "linux" {
		return "no supported service manager: v0.1 manages systemd only and GOOS is " + runtime.GOOS
	}
	if _, err := exec.LookPath("systemctl"); err != nil {
		return "no supported service manager: systemctl not found on PATH"
	}
	return ""
}
