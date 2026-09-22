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

	// The daemon's own refresh and pre-warm read tree blobs out of data/ packs,
	// so the window only measures crawlers once that work has finished.
	waitPrewarmed(t, e)

	// Watch only the pack files. The daemon's own restic processes open config,
	// keys, locks, snapshots and index in the background; that is not a crawler
	// read. The repository has a cache_dir, so restic serves tree packs, the
	// index and the snapshot list from the cache and only file content blobs
	// still come out of data/, which makes opens there the crawler's reads.
	c := watchOpens(t, filepath.Join(repo, "data"))
	time.Sleep(crawlerSettle)
	if got := c.n.Load(); got != 0 {
		t.Fatalf("repo IN_OPEN count while idle for %v = %d, want 0", crawlerSettle, got)
	}
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

	// M-002: the counter must see a direct .snapshot read, or 0 above proves
	// nothing. Name the backed-up file: the first entry of the .snapshot listing
	// is the synthetic info.json, which the daemon serves without touching a
	// data/ pack, so walking to "the first regular file" reads no content at all.
	content := filepath.Join(proj, ".snapshot", "latest", "src", "deep", "a.txt")
	if got, err := os.ReadFile(content); err != nil || string(got) != "x\n" {
		t.Errorf("read %s = %q, %v, want \"x\\n\" (history mount up?)", content, got, err)
	}
	if got := c.settle(); got == 0 {
		t.Errorf("repo IN_OPEN count after direct .snapshot read = 0, want > 0")
	}

	// The probe must be a deny-listed process whose name the policy can resolve.
	// readerpolicy.ProcName reads /proc/<pid>/comm and FUSE reports the calling
	// THREAD, so a multi-threaded crawler is only denied when its worker threads
	// keep the process comm; find is single-threaded, so its comm is "find" for
	// every request it makes. -L follows the .snapshot symlink, which the default
	// traversal would only lstat, never entering the history mount.
	probe := exec.Command("find", "-L", filepath.Join(proj, ".snapshot"))
	probe.Env = e.environ()
	_ = probe.Run() // denied entries make find exit non-zero; only the event matters
	stdout, stderr, code := runSnapback(t, e, "status", "--json")
	if code != 0 {
		t.Fatalf("snapback status --json exit = %d, want 0 (stderr %q)", code, stderr)
	}
	lower := strings.ToLower(stdout)
	hasEvent := strings.Contains(lower, "throttl") || strings.Contains(lower, "deny")
	if !hasEvent || !strings.Contains(stdout, `"find"`) {
		t.Errorf("status --json = %q, want a throttle/deny event naming \"find\"", stdout)
	}
}

// crawlerSettle is how long the open counter must stay at zero before the
// crawlers run, so a late daemon read cannot be blamed on them.
const crawlerSettle = 3 * time.Second

// waitPrewarmed polls status --json until the daemon has no pre-warm work
// left, so its own pack reads are over before the crawler window opens.
func waitPrewarmed(t *testing.T, e env) {
	t.Helper()
	deadline := time.Now().Add(histPollCap)
	for {
		st, out := readStatus(t, e)
		var s struct {
			Data struct {
				Prewarm struct {
					Warm    int `json:"warm"`
					Pending int `json:"pending"`
				} `json:"prewarm"`
			} `json:"data"`
		}
		if st.OK && json.Unmarshal([]byte(out), &s) == nil &&
			s.Data.Prewarm.Pending == 0 && s.Data.Prewarm.Warm >= 1 {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("prewarm not quiescent within %v; last status --json = %q", histPollCap, out)
		}
		time.Sleep(time.Second)
	}
}

// pollEligible polls status --json until some directory has eligible
// snapshots, and returns the last counts seen.
func pollEligible(t *testing.T, e env) (map[string]int, error) {
	t.Helper()
	deadline := time.Now().Add(histPollCap)
	var last map[string]int
	for {
		st, out := readStatus(t, e)
		if !st.OK && time.Now().After(deadline) {
			return last, fmt.Errorf("status --json not ok: %s", out)
		}
		last = st.Data.EligibleCount
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
