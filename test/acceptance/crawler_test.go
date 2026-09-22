//go:build integration

package acceptance

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
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
	recordEvidence(t, "acc-12")
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
	hist := writeLinkConfig(t, e, linkConfig{Root: root, SeedPath: "proj", MaxDepth: 3, Host: histHost})
	repo := configValue(t, e, "repository")
	proj := filepath.Join(root, "proj")
	// newRepo only runs restic init. Without a snapshot of the root the history
	// view has no entry for proj, so proj/.snapshot dangles (ENOENT) and no
	// crawler reaches FUSE, which makes the zero-reads and throttle checks
	// meaningless. The host and the backed-up path must match the config's
	// prefix_map, the shape aliases_test.go proves on macOS and Linux.
	resticRun(t, root, repo, configValue(t, e, "password_file"), "backup", "--host", histHost, root)
	startDaemon(t, e)
	if _, stderr, code := runSnapback(t, e, "seed", proj); code != 0 {
		t.Fatalf("snapback seed exit = %d, want 0 (stderr %q)", code, stderr)
	}
	if got := ownedLinks(t, root, hist); len(got) == 0 {
		t.Fatalf("owned links after seed = %v, want at least one", got)
	}
	// restic reloads its snapshot list on about a 60s window, so the refresher
	// may report no eligible snapshots for a while after the backup.
	if got, err := pollEligible(t, e); err != nil {
		t.Fatalf("EligibleCount within %v: %v (last %v)", histPollCap, err, got)
	}

	// Watch only the pack files. The daemon's own restic processes open config,
	// keys, locks, snapshots and index in the background; that is not a crawler
	// read. Any file or directory content served through .snapshot comes from
	// data/ packs, so opens there are the ones a crawler could cause.
	c := watchOpens(t, filepath.Join(repo, "data"))
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
	// Read file content, not just the listing: only content comes from data/ packs.
	var read bool
	err := filepath.WalkDir(filepath.Join(proj, ".snapshot")+"/", func(p string, d os.DirEntry, err error) error {
		if err != nil || read || !d.Type().IsRegular() {
			return err
		}
		_, err = os.ReadFile(p)
		read = err == nil
		return err
	})
	if err != nil || !read {
		t.Errorf("read a file under %s/.snapshot: read = %t, error = %v, want a read (history mount up?)", proj, read, err)
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

// pollEligible polls status --json until some directory has eligible
// snapshots, and returns the last counts seen.
func pollEligible(t *testing.T, e env) (map[string]int, error) {
	t.Helper()
	deadline := time.Now().Add(histPollCap)
	var last map[string]int
	for {
		stdout, stderr, code := runSnapback(t, e, "status", "--json")
		if code != 0 {
			return last, fmt.Errorf("status --json exit = %d (stderr %q)", code, stderr)
		}
		var st struct {
			EligibleCount map[string]int
		}
		if err := json.Unmarshal([]byte(stdout), &st); err != nil {
			return last, fmt.Errorf("decode status --json %q: %w", stdout, err)
		}
		last = st.EligibleCount
		for _, n := range last {
			if n > 0 {
				return last, nil
			}
		}
		if time.Now().After(deadline) {
			return last, errors.New("no directory has an eligible snapshot")
		}
		time.Sleep(histPoll)
	}
}

// configValue returns the first "key: value" entry for key in the written config.
func configValue(t *testing.T, e env, key string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(e.Config, "snapback", "config.yaml"))
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	for _, line := range strings.Split(string(b), "\n") {
		if v, ok := strings.CutPrefix(strings.TrimSpace(line), key+": "); ok {
			return v
		}
	}
	t.Fatalf("config has no %s line", key)
	return ""
}
