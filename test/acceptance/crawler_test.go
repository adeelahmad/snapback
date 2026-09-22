//go:build integration

package acceptance

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// openCounter counts IN_OPEN events under a directory via inotifywait.
type openCounter struct {
	n   atomic.Int64
	cmd *exec.Cmd
}

func watchOpens(t *testing.T, dir string) *openCounter {
	t.Helper()
	c := &openCounter{cmd: exec.Command("inotifywait", "-m", "-r", "-q", "-e", "open", dir)}
	out, err := c.cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("inotifywait stdout: %v", err)
	}
	if err := c.cmd.Start(); err != nil {
		t.Fatalf("start inotifywait: %v", err)
	}
	t.Cleanup(func() { _ = c.cmd.Process.Kill(); _ = c.cmd.Wait() })
	go func() {
		s := bufio.NewScanner(out)
		for s.Scan() {
			c.n.Add(1)
		}
	}()
	time.Sleep(500 * time.Millisecond) // let inotifywait establish its watches
	return c
}

// settle waits for pending inotify events to be counted and returns the count.
func (c *openCounter) settle() int64 {
	time.Sleep(500 * time.Millisecond)
	return c.n.Load()
}

func TestAcc12CrawlerZeroReadsAndThrottle(t *testing.T) {
	requireFUSE(t)
	if runtime.GOOS != "linux" {
		skip(t, "missing prerequisite: inotify IN_OPEN counter needs linux, have "+runtime.GOOS)
	}
	for _, tool := range []string{"inotifywait", "rg", "fd", "rsync", "find"} {
		if _, err := exec.LookPath(tool); err != nil {
			skip(t, "missing prerequisite: "+tool+" not found on PATH")
		}
	}
	e := newEnv(t)
	root := filepath.Join(e.Root, "work")
	mkdirs(t, root, "proj/src/deep")
	writeFiles(t, root, map[string]string{"proj/src/deep/a.txt": "x\n"})
	hist := writeLinkConfig(t, e, linkConfig{Root: root, SeedPath: "proj", MaxDepth: 3})
	repo := resticRepoFromConfig(t, e)
	startDaemon(t, e)
	proj := filepath.Join(root, "proj")
	if _, stderr, code := runSnapback(t, e, "seed", proj); code != 0 {
		t.Fatalf("snapback seed exit = %d, want 0 (stderr %q)", code, stderr)
	}
	if got := ownedLinks(t, root, hist); len(got) == 0 {
		t.Fatalf("owned links after seed = %v, want at least one", got)
	}

	c := watchOpens(t, repo)
	dst := filepath.Join(e.Root, "dst")
	crawlers := [][]string{
		{"rg", "x", proj},
		{"fd", ".", proj},
		{"find", proj},
		{"rsync", "-a", proj + "/", dst + "/"},
	}
	for _, argv := range crawlers {
		cmd := exec.Command(argv[0], argv[1:]...)
		cmd.Env = e.environ()
		_ = cmd.Run() // rg exits 1 on no match; only repo reads matter
	}
	if got := c.settle(); got != 0 {
		t.Errorf("repo IN_OPEN count after default crawlers = %d, want 0", got)
	}

	// M-002: the counter must see a direct .snapshot read, or 0 above proves nothing.
	if _, err := os.ReadDir(filepath.Join(proj, ".snapshot")); err != nil {
		t.Errorf("ReadDir(%s/.snapshot) error = %v, want nil (history mount up?)", proj, err)
	}
	if got := c.settle(); got == 0 {
		t.Errorf("repo IN_OPEN count after direct .snapshot read = 0, want > 0")
	}

	rg := exec.Command("rg", "-L", "x", proj)
	rg.Env = e.environ()
	_ = rg.Run()
	stdout, stderr, code := runSnapback(t, e, "status", "--json")
	if code != 0 {
		t.Fatalf("snapback status --json exit = %d, want 0 (stderr %q)", code, stderr)
	}
	lower := strings.ToLower(stdout)
	hasEvent := strings.Contains(lower, "throttl") || strings.Contains(lower, "deny")
	if !hasEvent || !strings.Contains(stdout, `"rg"`) {
		t.Errorf("status --json = %q, want a throttle/deny event naming \"rg\"", stdout)
	}
}

// resticRepoFromConfig returns the repository path from the written config.
func resticRepoFromConfig(t *testing.T, e env) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(e.Config, "snapback", "config.yaml"))
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	for _, line := range strings.Split(string(b), "\n") {
		if v, ok := strings.CutPrefix(strings.TrimSpace(line), "repository: "); ok {
			return v
		}
	}
	t.Fatalf("config has no repository line")
	return ""
}
