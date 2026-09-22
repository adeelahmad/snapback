package shellhook

import (
	"os"
	"os/exec"
	"path/filepath"
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
	script := `source $SNIP; emit fish_prompt; cd a; cd '../` + meta + `'; false; emit fish_prompt; echo "rc=$status"`
	res := h.runShell(t, "fish", "--no-config", "-c", script)
	if !strings.HasSuffix(res.stdout, "rc=1\n") || res.stderr != "" {
		t.Errorf("hook stdout = %q, stderr = %q, want suffix %q and empty", res.stdout, res.stderr, "rc=1\n")
	}
	recs := waitLog(t, h.log, 3)
	for i, want := range []string{h.dir, a, m} {
		if got := recs[i][len(recs[i])-1]; got != want {
			t.Errorf("notify %d dir = %q, want %q", i, got, want)
		}
	}
	for _, p := range []string{filepath.Join(h.dir, "PWNED"), filepath.Join(a, "PWNED"), filepath.Join(m, "PWNED")} {
		if _, err := os.Lstat(p); err == nil {
			t.Errorf("%s exists: directory name was executed", p)
		}
	}
}
