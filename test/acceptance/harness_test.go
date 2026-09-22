//go:build integration

package acceptance

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHarnessSandboxesHome(t *testing.T) {
	if realHome == "" {
		t.Fatal("realHome is empty; TestMain did not capture the real home")
	}
	env := newEnv(t)

	stdout, stderr, code := runSnapback(t, env, "config", "path")
	if code != 0 {
		t.Fatalf("runSnapback(config path) exit = %d, want 0; stderr: %s", code, stderr)
	}
	got := strings.TrimSpace(stdout)
	if got == "" {
		t.Fatal("runSnapback(config path) printed no path, want a config path")
	}
	if !isUnder(got, env.Root) {
		t.Errorf("config path = %q, want under sandbox root %q", got, env.Root)
	}
	if isUnder(got, realHome) {
		t.Errorf("config path = %q, want not under real home %q", got, realHome)
	}
}

func TestRecordEvidenceWritesJSON(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("SNAPBACK_EVIDENCE_DIR", dir)

	t.Run("skipped", func(st *testing.T) {
		recordEvidence(st, "acc-00")
		skip(st, "no fuse")
	})

	data, err := os.ReadFile(filepath.Join(dir, "acc-00.json"))
	if err != nil {
		t.Fatalf("read evidence: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("json.Unmarshal(acc-00.json) = %v, want valid JSON", err)
	}
	want := map[string]string{"acc": "acc-00", "status": "skip", "skip_reason": "no fuse"}
	for k, w := range want {
		if g, _ := got[k].(string); g != w {
			t.Errorf("evidence[%q] = %q, want %q", k, g, w)
		}
	}
	for _, k := range []string{"kernel", "commit"} {
		if g, _ := got[k].(string); g == "" {
			t.Errorf("evidence[%q] = %q, want non-empty", k, g)
		}
	}
	for _, k := range []string{"restic_version", "gofuse_version", "fuse3_version"} {
		if _, ok := got[k]; !ok {
			t.Errorf("evidence has no %q key, want present", k)
		}
	}
}

// isUnder reports whether p is dir or inside it, after resolving symlinks
// (macOS temp dirs live behind /var -> /private/var).
func isUnder(p, dir string) bool {
	rp, rd := resolve(p), resolve(dir)
	rel, err := filepath.Rel(rd, rp)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

func resolve(p string) string {
	for cur, rest := filepath.Clean(p), ""; ; {
		if r, err := filepath.EvalSymlinks(cur); err == nil {
			return filepath.Join(r, rest)
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return filepath.Clean(p)
		}
		rest = filepath.Join(filepath.Base(cur), rest)
		cur = parent
	}
}
