package shellhook

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// needTool skips the test when name is not on PATH.
func needTool(t *testing.T, name string) {
	t.Helper()
	if _, err := exec.LookPath(name); err != nil {
		t.Skipf("%s not installed: real-shell check skipped", name)
	}
}

// fakeSnapback writes <tmp>/bin/snapback, which appends each argument
// NUL-terminated to $SNAPBACK_TEST_LOG and then sleeps
// ${SNAPBACK_TEST_SLEEP:-0} seconds. It returns the bin directory.
func fakeSnapback(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "bin")
	if err := os.MkdirAll(bin, 0o755); err != nil {
		t.Fatalf("os.MkdirAll(%q) = %v", bin, err)
	}
	script := "#!/bin/sh\nprintf '%s\\0' \"$@\" >> \"$SNAPBACK_TEST_LOG\"\nsleep \"${SNAPBACK_TEST_SLEEP:-0}\"\n"
	if err := os.WriteFile(filepath.Join(bin, "snapback"), []byte(script), 0o755); err != nil {
		t.Fatalf("os.WriteFile(snapback) = %v", err)
	}
	return bin
}

// parseLog splits the NUL-terminated argument stream into one argv record
// per snapback invocation; a record starts at each "notify" argument.
func parseLog(data []byte) [][]string {
	var recs [][]string
	for _, tok := range bytes.SplitAfter(data, []byte{0}) {
		if len(tok) == 0 || tok[len(tok)-1] != 0 {
			continue
		}
		arg := string(tok[:len(tok)-1])
		if arg == "notify" || len(recs) == 0 {
			recs = append(recs, nil)
		}
		recs[len(recs)-1] = append(recs[len(recs)-1], arg)
	}
	return recs
}

// waitLog polls path for up to 5 s until it holds n argv records and
// returns them.
func waitLog(t *testing.T, path string, n int) [][]string {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	var recs [][]string
	for {
		data, _ := os.ReadFile(path)
		recs = parseLog(data)
		if len(recs) >= n || time.Now().After(deadline) {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if len(recs) != n {
		t.Fatalf("snapback invocations = %d %q, want %d", len(recs), recs, n)
	}
	return recs
}

// hookEnv is a sandbox for running the bash snippet in a real shell.
type hookEnv struct {
	dir     string // physical temp dir, also the shell's start dir
	home    string
	log     string
	snippet string
	env     []string
}

func newHookEnv(t *testing.T) *hookEnv {
	t.Helper()
	needTool(t, "bash")
	bin := fakeSnapback(t)
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatalf("filepath.EvalSymlinks() = %v", err)
	}
	h := &hookEnv{
		dir:     dir,
		home:    filepath.Join(dir, "home"),
		log:     filepath.Join(dir, "snapback.log"),
		snippet: filepath.Join(dir, "snapback.bash"),
	}
	if err := os.MkdirAll(h.home, 0o755); err != nil {
		t.Fatalf("os.MkdirAll(%q) = %v", h.home, err)
	}
	src, err := Script("bash")
	if err != nil {
		t.Fatalf(`Script("bash") error = %v, want nil`, err)
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

type bashResult struct {
	stdout, stderr string
	pid            int
	elapsed        time.Duration
}

// run executes `bash --noprofile --norc -c script bash args...` in h.dir.
func (h *hookEnv) run(t *testing.T, script string, extraEnv []string, args ...string) bashResult {
	t.Helper()
	cmd := exec.Command("bash", append([]string{"--noprofile", "--norc", "-c", script, "bash"}, args...)...)
	cmd.Dir = h.dir
	cmd.Env = append(append([]string{}, h.env...), extraEnv...)
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	start := time.Now()
	err := cmd.Run()
	res := bashResult{stdout: out.String(), stderr: errb.String(), elapsed: time.Since(start)}
	if cmd.Process != nil {
		res.pid = cmd.Process.Pid
	}
	if err != nil {
		t.Fatalf("bash -c %q: %v (stderr %q)", script, err, res.stderr)
	}
	return res
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) = %v", path, err)
	}
	return string(b)
}
