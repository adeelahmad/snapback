package shellhook

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestScriptFishGolden(t *testing.T) {
	src := checkGolden(t, "fish", []string{"--on-variable PWD", "fish_prompt", "pwd -P", "disown", "/dev/null"})
	needTool(t, "fish")
	file := filepath.Join(t.TempDir(), "snapback.fish")
	if err := os.WriteFile(file, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("fish", "--no-execute", file).CombinedOutput(); err != nil {
		t.Errorf("fish --no-execute = %v, output:\n%s\nwant exit 0", err, out)
	}
}

func TestFishNotifiesOnCdAndPreservesStatus(t *testing.T) {
	h := newShellEnv(t, "fish")
	const meta = "$(touch PWNED)"
	a, m := filepath.Join(h.dir, "a"), filepath.Join(h.dir, meta)
	for _, d := range []string{a, m} {
		if err := os.Mkdir(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	script := `source $SNIP; emit fish_prompt; cd a; cd '../` + meta + `'; false; __snapback_hook; echo rc=$status`
	res := h.runShell(t, "fish", "--no-config", "-c", script)
	if !strings.HasSuffix(res.stdout, "rc=1\n") || res.stderr != "" {
		t.Errorf("hook stdout = %q, stderr = %q, want suffix %q and empty", res.stdout, res.stderr, "rc=1\n")
	}
	recs := waitLog(t, h.log, 3)
	// Each notify runs detached, so the records can reach the log in any
	// order; compare the notified directories as a sorted set instead of by
	// position.
	got := make([]string, len(recs))
	for i, rec := range recs {
		got[i] = rec[len(rec)-1]
	}
	slices.Sort(got)
	want := []string{h.dir, a, m}
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Errorf("notify dirs = %q, want %q (each dir exactly once)", got, want)
	}
	for _, p := range []string{filepath.Join(h.dir, "PWNED"), filepath.Join(a, "PWNED"), filepath.Join(m, "PWNED")} {
		if _, err := os.Lstat(p); err == nil {
			t.Errorf("%s exists: directory name was executed", p)
		}
	}
}
