package shellhook

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"
)

const hookCmd = `eval "$PROMPT_COMMAND"`

func TestBashSnippetShellcheck(t *testing.T) {
	needTool(t, "shellcheck")
	src, err := Script("bash")
	if err != nil {
		t.Fatalf(`Script("bash") error = %v, want nil`, err)
	}
	file := filepath.Join(t.TempDir(), "snapback.bash")
	if err := os.WriteFile(file, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := exec.Command("shellcheck", "-s", "bash", "-S", "style", file).CombinedOutput()
	if err != nil || len(out) != 0 {
		t.Errorf("shellcheck -s bash -S style = %v, output:\n%s\nwant exit 0 and no output", err, out)
	}
}

func notifyArgv(pid int, dir string) []string {
	return []string{"notify", "--timeout", "200ms", "--session", strconv.Itoa(pid), "--", dir}
}

func TestBashPreservesStringPromptCommand(t *testing.T) {
	h := newHookEnv(t)
	script := `PROMPT_COMMAND='echo prior >> "$HOME/prior"'
source "$SNIP"
(exit 7); ` + hookCmd + `; echo "rc=$?"`
	res := h.run(t, script, nil)
	if res.stdout != "rc=7\n" || res.stderr != "" {
		t.Errorf("hook stdout = %q, stderr = %q, want %q and empty", res.stdout, res.stderr, "rc=7\n")
	}
	if got := readFile(t, filepath.Join(h.home, "prior")); !strings.Contains(got, "prior") {
		t.Errorf("prior PROMPT_COMMAND output = %q, want it to contain %q", got, "prior")
	}
	recs := waitLog(t, h.log, 1)
	if want := notifyArgv(res.pid, h.dir); !slices.Equal(recs[0], want) {
		t.Errorf("snapback argv = %q, want %q", recs[0], want)
	}
}

func TestBashPreservesArrayPromptCommand(t *testing.T) {
	h := newHookEnv(t)
	ver := h.run(t, `echo "${BASH_VERSINFO[0]} ${BASH_VERSINFO[1]}"`, nil)
	var major, minor int
	fields := strings.Fields(ver.stdout)
	if len(fields) == 2 {
		major, _ = strconv.Atoi(fields[0])
		minor, _ = strconv.Atoi(fields[1])
	}
	if major < 5 || (major == 5 && minor < 1) {
		t.Skipf("bash %s lacks array PROMPT_COMMAND (needs 5.1)", strings.TrimSpace(ver.stdout))
	}
	const prior = `echo prior >> "$HOME/prior"`
	script := `PROMPT_COMMAND=('` + prior + `')
source "$SNIP"
(exit 3); for c in "${PROMPT_COMMAND[@]}"; do eval "$c"; done; echo "rc=$?"
printf '%s\n%s' "${#PROMPT_COMMAND[@]}" "${PROMPT_COMMAND[0]}" > "$HOME/state"`
	res := h.run(t, script, nil)
	if res.stdout != "rc=0\n" || res.stderr != "" {
		t.Errorf("hook stdout = %q, stderr = %q, want %q and empty", res.stdout, res.stderr, "rc=0\n")
	}
	if got, want := readFile(t, filepath.Join(h.home, "state")), "2\n"+prior; got != want {
		t.Errorf("PROMPT_COMMAND length and element 0 = %q, want %q", got, want)
	}
	if got := readFile(t, filepath.Join(h.home, "prior")); !strings.Contains(got, "prior") {
		t.Errorf("prior PROMPT_COMMAND output = %q, want it to contain %q", got, "prior")
	}
	waitLog(t, h.log, 1)
}

func TestBashDedupAndPhysicalDir(t *testing.T) {
	h := newHookEnv(t)
	a, b := filepath.Join(h.dir, "a"), filepath.Join(h.dir, "b")
	for _, d := range []string{a, b} {
		if err := os.Mkdir(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink(a, filepath.Join(h.dir, "link")); err != nil {
		t.Fatal(err)
	}
	script := `set -u
source "$SNIP"
` + hookCmd + `; ` + hookCmd + `; cd link; ` + hookCmd + `; cd ../b; ` + hookCmd
	res := h.run(t, script, nil)
	if res.stderr != "" {
		t.Errorf("hook under set -u stderr = %q, want empty", res.stderr)
	}
	recs := waitLog(t, h.log, 3)
	physA, _ := filepath.EvalSymlinks(a)
	physB, _ := filepath.EvalSymlinks(b)
	for i, want := range []string{h.dir, physA, physB} {
		if got := recs[i][len(recs[i])-1]; got != want {
			t.Errorf("notify %d dir = %q, want %q", i, got, want)
		}
	}
}

func TestBashMetacharacterDirs(t *testing.T) {
	names := []string{"sp ace", "new\nline", "trail\n", "$(touch PWNED)", "semi;colon", `quote'"`, "-dash", `back\slash`}
	for _, name := range names {
		t.Run(strconv.Quote(name), func(t *testing.T) {
			h := newHookEnv(t)
			dir := filepath.Join(h.dir, name)
			if err := os.Mkdir(dir, 0o755); err != nil {
				t.Fatal(err)
			}
			res := h.run(t, `cd -- "$1" && source "$SNIP" && `+hookCmd, nil, name)
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

func TestBashHookNeverBlocks(t *testing.T) {
	h := newHookEnv(t)
	res := h.run(t, `source "$SNIP"; `+hookCmd+`; echo done`, []string{"SNAPBACK_TEST_SLEEP=5"})
	if !strings.Contains(res.stdout, "done") {
		t.Errorf("hook stdout = %q, want it to contain %q", res.stdout, "done")
	}
	if res.elapsed >= time.Second {
		t.Errorf("hook took %v, want < 1s", res.elapsed)
	}
	if strings.Contains(res.stdout, "[1]") || res.stderr != "" {
		t.Errorf("hook stdout = %q, stderr = %q, want no job-control line and empty stderr", res.stdout, res.stderr)
	}
	waitLog(t, h.log, 1)
}
