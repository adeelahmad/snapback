package daemon

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/history"
	"github.com/adeelahmad/snapback/internal/ipc"
	"github.com/adeelahmad/snapback/internal/mount"
	"github.com/adeelahmad/snapback/internal/recovery"
	"github.com/adeelahmad/snapback/internal/status"
)

func TestLockSecondInstanceIsStaleState(t *testing.T) {
	dir := t.TempDir()

	unlock, err := Lock(dir)
	if err != nil {
		t.Fatalf("Lock(%q) first = %v, want nil", dir, err)
	}

	unlock2, err := Lock(dir)
	if got := errcode.Of(err); got != errcode.StaleState {
		if unlock2 != nil {
			unlock2()
		}
		unlock()
		t.Fatalf("Lock(%q) second: errcode.Of(%v) = %q, want %q", dir, err, got, errcode.StaleState)
	}
	if !strings.Contains(err.Error(), "another daemon") {
		t.Errorf("Lock(%q) second = %q, want message naming another daemon", dir, err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "daemon.pid"))
	if err != nil {
		t.Errorf("read daemon.pid: %v", err)
	} else if got, want := strings.TrimSpace(string(data)), strconv.Itoa(os.Getpid()); got != want {
		t.Errorf("daemon.pid = %q, want %q", got, want)
	}

	unlock()
	unlock3, err := Lock(dir)
	if err != nil {
		t.Fatalf("Lock(%q) after unlock = %v, want nil", dir, err)
	}
	unlock3()
}

func TestRunStartupOrder(t *testing.T) {
	h := newHarness(t)
	d := New(h.cfg, h.deps)
	start(t, d)

	waitState(t, d, 2*time.Second, func(s string) bool { return s == "ready" })

	want := []string{"lock", "ipc", "recover", "supervisor.start", "refresh", "discovery.start", "prewarm"}
	got := h.rec.list()
	if len(got) < len(want) || !slices.Equal(got[:len(want)], want) {
		t.Errorf("startup steps = %q, want prefix %q", got, want)
	}
}

func TestRunRejectsInvalidConfigFirst(t *testing.T) {
	h := newHarness(t)
	noRoots := &config.Config{
		StateDir:     t.TempDir(),
		Repositories: h.cfg.Repositories,
	}
	for _, tc := range []struct {
		name string
		cfg  *config.Config
	}{
		{"nil", nil},
		{"no roots", noRoots},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			err := Run(ctx, tc.cfg, h.deps)
			if got := errcode.Of(err); got != errcode.InvalidConfig {
				t.Errorf("Run(%s): errcode.Of(%v) = %q, want %q", tc.name, err, got, errcode.InvalidConfig)
			}
			if got := h.rec.list(); len(got) != 0 {
				t.Errorf("Run(%s) steps = %q, want none", tc.name, got)
			}
		})
	}
	if _, err := os.Stat(filepath.Join(noRoots.StateDir, "daemon.lock")); !os.IsNotExist(err) {
		t.Errorf("stat daemon.lock = %v, want not exist (no lock taken)", err)
	}
}

func TestOfflineRepoIsDegradedNotEmpty(t *testing.T) {
	h := newHarness(t)
	h.sup.states = map[string]history.RepoState{
		"repoA": history.StateReady,
		"repoB": history.StateFailed,
	}
	h.ref.result = RefreshResult{Generation: 1, At: fixedNow, Failed: []string{"repoB"}}
	h.ref.err = errcode.New(errcode.RepoUnavailable, "refresh repoB", nil)
	d := New(h.cfg, h.deps)
	errc := start(t, d)

	got := waitState(t, d, 2*time.Second, func(s string) bool { return s != "starting" })
	if got != "degraded" {
		t.Errorf("Status().State = %q, want %q", got, "degraded")
	}

	want := []status.Repo{
		{ID: "repoA", State: string(history.StateReady)},
		{ID: "repoB", State: string(history.StateFailed), Code: errcode.RepoUnavailable},
	}
	if repos := d.Status().Repos; !slices.Equal(repos, want) {
		t.Errorf("Status().Repos = %+v, want %+v", repos, want)
	}

	select {
	case err := <-errc:
		t.Errorf("Run returned %v while a repository is offline, want still running", err)
	default:
	}

	steps := h.rec.list()
	for _, s := range []string{"discovery.start", "prewarm"} {
		if !slices.Contains(steps, s) {
			t.Errorf("startup steps = %q, want %q (degraded still starts it)", steps, s)
		}
	}
}

func TestStartupDoesNotWalkRoots(t *testing.T) {
	h := newHarness(t)
	root := h.cfg.Roots[0].LocalPath
	locked := filepath.Join(root, "locked")
	if err := os.Mkdir(locked, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(locked, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(locked, 0o700) })
	if err := syscall.Mkfifo(filepath.Join(root, "fifo"), 0o600); err != nil {
		t.Fatalf("mkfifo: %v", err)
	}

	d := New(h.cfg, h.deps)
	start(t, d)

	waitState(t, d, 2*time.Second, func(s string) bool { return s == "ready" })
	if got := h.linker.total(); got != 0 {
		t.Errorf("Linker.Ensure calls during startup = %d, want 0", got)
	}
}

func TestStartupRecoveryReportedInStatus(t *testing.T) {
	h := newHarness(t)
	h.deps.Recoverer = &fakeRecoverer{rec: h.rec, report: recovery.Report{
		Unmounted: []string{"a", "b"},
		Foreign:   []string{"c"},
	}}
	d := New(h.cfg, h.deps)
	start(t, d)
	waitState(t, d, 2*time.Second, func(s string) bool { return s == "ready" })

	s := decodeSnapshot(t, call(d, ipc.Request{Op: ipc.OpStatus}))
	want := &status.RecoverySummary{Unmounted: []string{"a", "b"}, Foreign: []string{"c"}}
	if !reflect.DeepEqual(s.Recovery, want) {
		t.Errorf("status op Recovery = %+v, want %+v", s.Recovery, want)
	}
}

// recordingPublisher records every catalog published to it.
type recordingPublisher struct {
	mu   sync.Mutex
	cats []mount.Catalog
}

func (p *recordingPublisher) Publish(cat mount.Catalog) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.cats = append(p.cats, cat)
}

func (p *recordingPublisher) published() []mount.Catalog {
	p.mu.Lock()
	defer p.mu.Unlock()
	return slices.Clone(p.cats)
}

// oneEntryCatalog resolves a single name under the root inode.
type oneEntryCatalog struct {
	name string
	ino  uint64
}

func (c *oneEntryCatalog) Lookup(parent uint64, name string) (uint64, mount.Kind, bool) {
	if parent != rootIno || name != c.name {
		return 0, 0, false
	}
	return c.ino, mount.KindDir, true
}

func (c *oneEntryCatalog) ReadDir(dir uint64) ([]string, bool) {
	if dir != rootIno {
		return nil, false
	}
	return []string{c.name}, true
}

func (c *oneEntryCatalog) Readlink(uint64) (string, bool) { return "", false }

func (c *oneEntryCatalog) ReadFile(uint64) ([]byte, bool) { return nil, false }

// rootIno is the inode the fake catalogs list their entry under.
const rootIno = 1

// lookupResult carries a Lookup made from another goroutine.
type lookupResult struct {
	ino   uint64
	found bool
}

func TestReadyPublisherWaitsForFirstPublish(t *testing.T) {
	rec := &recordingPublisher{}
	pub := NewReadyPublisher(rec)
	served := pub.cat

	got := make(chan lookupResult, 1)
	go func() {
		ino, _, found := served.Lookup(rootIno, "daily")
		got <- lookupResult{ino, found}
	}()
	select {
	case r := <-got:
		t.Fatalf("Lookup(1, daily) = %d, %t before the first publish, want a wait", r.ino, r.found)
	case <-time.After(50 * time.Millisecond):
	}

	pub.Publish(&oneEntryCatalog{name: "daily", ino: 7})
	select {
	case r := <-got:
		if r.ino != 7 || !r.found {
			t.Errorf("Lookup(1, daily) = %d, %t, want 7, true", r.ino, r.found)
		}
	case <-time.After(time.Second):
		t.Fatal("Lookup(1, daily) did not return after the first publish")
	}
	if !pub.ready.IsReady() {
		t.Error("IsReady() = false after the first publish, want true")
	}
}

func TestReadyPublisherServesLaterGenerations(t *testing.T) {
	rec := &recordingPublisher{}
	pub := NewReadyPublisher(rec)

	pub.Publish(&oneEntryCatalog{name: "daily", ino: 7})
	pub.Publish(&oneEntryCatalog{name: "weekly", ino: 9})

	cats := rec.published()
	if len(cats) != 2 {
		t.Fatalf("published() = %d catalogs, want 2", len(cats))
	}
	if cats[0] != cats[1] {
		t.Errorf("published() = two different catalogs, want the same gated catalog")
	}
	if !pub.ready.IsReady() {
		t.Error("IsReady() = false after two publishes, want true")
	}
	ino, _, found := cats[1].Lookup(rootIno, "weekly")
	if ino != 9 || !found {
		t.Errorf("Lookup(1, weekly) = %d, %t, want 9, true", ino, found)
	}
}
