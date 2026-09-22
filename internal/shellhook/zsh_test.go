package shellhook

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// zshChain runs every precmd function the way zsh does before a prompt.
const zshChain = `for f in "${precmd_functions[@]}"; do $f; done`

// newShellEnv is newHookEnv for shell: it skips when shell is absent and
// writes Script(shell) to <tmp>/snapback.<shell>.
func newShellEnv(t *testing.T, shell string) *hookEnv {
	t.Helper()
	needTool(t, shell)
	bin := fakeSnapback(t)
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatalf("filepath.EvalSymlinks() = %v", err)
	}
	h := &hookEnv{
		dir:     dir,
		home:    filepath.Join(dir, "home"),
		log:     filepath.Join(dir, "snapback.log"),
		snippet: filepath.Join(dir, "snapback."+shell),
	}
	if err := os.MkdirAll(h.home, 0o755); err != nil {
		t.Fatalf("os.MkdirAll(%q) = %v", h.home, err)
	}
	src, err := Script(shell)
	if err != nil {
		t.Fatalf("Script(%q) error = %v, want nil", shell, err)
	}
	if err := os.WriteFile(h.snippet, []byte(src), 0o644); err != nil {
		t.Fatalf("os.WriteFile(%q) = %v", h.snippet, err)
	}
	h.env = []string{
		"PATH=" + bin + string(os.PathListSeparator) + os.Getenv("PATH"),
		"HOME=" + h.home,
		"SNAPBACK_TEST_LOG=" + h.log,
		"SNIP=" + h.snippet,
		"LC_ALL=C",
	}
	return h
}

// runShell executes argv in h.dir and fails the test when it exits non-zero.
func (h *hookEnv) runShell(t *testing.T, argv ...string) bashResult {
	t.Helper()
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Dir = h.dir
	cmd.Env = h.env
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	err := cmd.Run()
	res := bashResult{stdout: out.String(), stderr: errb.String()}
	if err != nil {
		t.Fatalf("%q: %v (stdout %q, stderr %q)", argv, err, res.stdout, res.stderr)
	}
	return res
}

// checkGolden compares Script(shell) with testdata/golden/<shell>.golden,
// requires each of wants in it and returns the snippet.
func checkGolden(t *testing.T, shell string, wants []string) string {
	t.Helper()
	got, err := Script(shell)
	if err != nil {
		t.Fatalf("Script(%q) error = %v, want nil", shell, err)
	}
	golden := filepath.Join("testdata", "golden", shell+".golden")
	if *update {
		if err := os.MkdirAll(filepath.Dir(golden), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(golden, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("golden %s missing (run with -update): %v", golden, err)
	}
	if got != string(want) {
		t.Errorf("Script(%q) = %q, want golden %q", shell, got, want)
	}
	for _, s := range wants {
		if !strings.Contains(got, s) {
			t.Errorf("Script(%q) lacks %q", shell, s)
		}
	}
	return got
}

func TestScriptZshGolden(t *testing.T) {
	src := checkGolden(t, "zsh", []string{"add-zsh-hook precmd", "emulate -L zsh", "pwd -P", "&!", "/dev/null"})
	needTool(t, "zsh")
	file := filepath.Join(t.TempDir(), "snapback.zsh")
	if err := os.WriteFile(file, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := exec.Command("zsh", "-n", file).CombinedOutput()
	if err != nil || len(out) != 0 {
		t.Errorf("zsh -n = %v, output:\n%s\nwant exit 0 and no output", err, out)
	}
}

func TestZshChainsPrecmdAndStatus(t *testing.T) {
	h := newShellEnv(t, "zsh")
	script := `prior() { print prior >> $HOME/prior }; precmd_functions=(prior)
source "$SNIP"
print -rn -- "${(j: :)precmd_functions}" > $HOME/state
(exit 5); for f in $precmd_functions; do $f; done; print "rc=$?"`
	res := h.runShell(t, "zsh", "-f", "-c", script)
	if res.stdout != "rc=5\n" || res.stderr != "" {
		t.Errorf("hook stdout = %q, stderr = %q, want %q and empty", res.stdout, res.stderr, "rc=5\n")
	}
	if got, want := readFile(t, filepath.Join(h.home, "state")), "prior __snapback_hook"; got != want {
		t.Errorf("precmd_functions = (%s), want (%s)", got, want)
	}
	if got := readFile(t, filepath.Join(h.home, "prior")); !strings.Contains(got, "prior") {
		t.Errorf("prior precmd output = %q, want it to contain %q", got, "prior")
	}
	recs := waitLog(t, h.log, 1)
	if got := recs[0][len(recs[0])-1]; got != h.dir {
		t.Errorf("notify dir = %q, want %q", got, h.dir)
	}
}

func TestZshOptionsUntouched(t *testing.T) {
	h := newShellEnv(t, "zsh")
	script := `setopt ksh_arrays nounset
setopt > $HOME/before
source "$SNIP"
` + zshChain + `; ` + zshChain + `
setopt > $HOME/after`
	res := h.runShell(t, "zsh", "-f", "-c", script)
	if res.stderr != "" {
		t.Errorf("hook stderr = %q, want empty", res.stderr)
	}
	before, after := readFile(t, filepath.Join(h.home, "before")), readFile(t, filepath.Join(h.home, "after"))
	if before != after {
		t.Errorf("setopt after hook = %q, want unchanged %q", after, before)
	}
	waitLog(t, h.log, 1)
}

func TestZshMetacharacterDirs(t *testing.T) {
	names := []string{"sp ace", "new\nline", "trail\n", "$(touch PWNED)", "semi;colon", `quote'"`, "-dash", `back\slash`}
	for _, name := range names {
		t.Run(strconv.Quote(name), func(t *testing.T) {
			h := newShellEnv(t, "zsh")
			dir := filepath.Join(h.dir, name)
			if err := os.Mkdir(dir, 0o755); err != nil {
				t.Fatal(err)
			}
			res := h.runShell(t, "zsh", "-f", "-c", `cd -- "$1" && source "$SNIP" && `+zshChain, "zsh", name)
			if res.stderr != "" {
				t.Errorf("hook stderr = %q, want empty", res.stderr)
			}
			recs := waitLog(t, h.log, 1)
			last := recs[len(recs)-1]
			if got := last[len(last)-1]; got != dir {
				t.Errorf("notify dir = %q, want %q", got, dir)
			}
			for _, p := range []string{filepath.Join(dir, "PWNED"), filepath.Join(h.dir, "PWNED")} {
				if _, err := os.Lstat(p); err == nil {
					t.Errorf("%s exists: directory name was executed", p)
				}
			}
		})
	}
}
