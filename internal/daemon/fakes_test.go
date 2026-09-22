package daemon

import (
	"context"
	"net"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/history"
	"github.com/adeelahmad/snapback/internal/ipc"
	"github.com/adeelahmad/snapback/internal/links"
	"github.com/adeelahmad/snapback/internal/provider"
	"github.com/adeelahmad/snapback/internal/provider/providertest"
	"github.com/adeelahmad/snapback/internal/recovery"
)

const (
	idA provider.SnapshotID = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	idB provider.SnapshotID = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
)

var fixedNow = time.Date(2026, 9, 22, 3, 0, 0, 0, time.UTC)

func fixedClock() time.Time { return fixedNow }

// sockPath keeps macOS sun_path under 104 bytes.
func sockPath(t *testing.T) string {
	t.Helper()
	t.Setenv("TMPDIR", "/tmp")
	return filepath.Join(t.TempDir(), "d.sock")
}

// recorder collects step names in call order.
type recorder struct {
	mu    sync.Mutex
	steps []string
}

func (r *recorder) add(step string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.steps = append(r.steps, step)
}

func (r *recorder) list() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.steps...)
}

type fakeSupervisor struct {
	rec      *recorder
	states   map[string]history.RepoState
	startErr error
	stopErr  error
}

func (f *fakeSupervisor) Start(context.Context) error {
	f.rec.add("supervisor.start")
	return f.startErr
}

func (f *fakeSupervisor) States() map[string]history.RepoState {
	out := make(map[string]history.RepoState, len(f.states))
	for k, v := range f.states {
		out[k] = v
	}
	return out
}

func (f *fakeSupervisor) Stop(context.Context) error {
	f.rec.add("supervisor.stop")
	return f.stopErr
}

type fakeHistory struct {
	rec        *recorder
	unmountErr error
	block      chan struct{}
}

func (f *fakeHistory) Unmount(ctx context.Context) error {
	f.rec.add("history.unmount")
	if f.block != nil {
		select {
		case <-f.block:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return f.unmountErr
}

type fakeRefresher struct {
	rec    *recorder
	mu     sync.Mutex
	result RefreshResult
	err    error
	calls  int
}

func (f *fakeRefresher) Refresh(context.Context) (RefreshResult, error) {
	f.rec.add("refresh")
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	return f.result, f.err
}

type fakeLinker struct {
	mu     sync.Mutex
	calls  map[string]int
	result map[string]links.Result
	errs   map[string]error

	records   []links.Record
	repair    links.RepairReport
	removed   links.RepairReport
	listErr   error
	repairErr error
	removeErr error
}

func (f *fakeLinker) List() ([]links.Record, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.records, f.listErr
}

func (f *fakeLinker) Repair(context.Context) (links.RepairReport, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.repair, f.repairErr
}

func (f *fakeLinker) RemoveManaged(context.Context) (links.RepairReport, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.removed, f.removeErr
}

func (f *fakeLinker) Ensure(_ context.Context, dir string) (links.Result, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.calls == nil {
		f.calls = map[string]int{}
	}
	f.calls[dir]++
	return f.result[dir], f.errs[dir]
}

func (f *fakeLinker) total() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	n := 0
	for _, c := range f.calls {
		n += c
	}
	return n
}

type fakeRecoverer struct {
	rec    *recorder
	report recovery.Report
}

func (f *fakeRecoverer) Recover(context.Context) (recovery.Report, error) {
	f.rec.add("recover")
	return f.report, nil
}

type fakeDiscovery struct {
	rec *recorder
}

func (f *fakeDiscovery) Start(context.Context) error {
	f.rec.add("discovery.start")
	return nil
}

func (f *fakeDiscovery) Stop() { f.rec.add("discovery.stop") }

type fakePrewarmer struct {
	providertest.Fake
	rec     *recorder
	results []provider.PrewarmResult
}

func (f *fakePrewarmer) Prewarm(context.Context) []provider.PrewarmResult {
	f.rec.add("prewarm")
	return f.results
}

// recordingListener records "listener.close" when closed.
type recordingListener struct {
	net.Listener
	rec *recorder
}

func (l *recordingListener) Close() error {
	l.rec.add("listener.close")
	return l.Listener.Close()
}

// harness bundles a config, healthy fakes and the shared recorder.
type harness struct {
	rec    *recorder
	cfg    *config.Config
	sup    *fakeSupervisor
	hist   *fakeHistory
	ref    *fakeRefresher
	linker *fakeLinker
	deps   Deps
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	rec := &recorder{}
	l, err := ipc.Listen(sockPath(t))
	if err != nil {
		t.Fatalf("ipc.Listen: %v", err)
	}
	t.Cleanup(func() { _ = l.Close() })
	h := &harness{
		rec: rec,
		cfg: &config.Config{
			StateDir: t.TempDir(),
			Repositories: []config.Repository{
				{ID: "repoA", Repository: "/r/a", PasswordFile: "/pw"},
				{ID: "repoB", Repository: "/r/b", PasswordFile: "/pw"},
			},
			Roots: []config.Root{{ID: "home", LocalPath: t.TempDir(), RepositoryID: "repoA"}},
		},
		sup: &fakeSupervisor{rec: rec, states: map[string]history.RepoState{
			"repoA": history.StateReady,
			"repoB": history.StateReady,
		}},
		hist:   &fakeHistory{rec: rec},
		ref:    &fakeRefresher{rec: rec, result: RefreshResult{Generation: 1, At: fixedNow}},
		linker: &fakeLinker{},
	}
	h.deps = Deps{
		Supervisor:      h.sup,
		History:         h.hist,
		Refresher:       h.ref,
		Linker:          h.linker,
		Recoverer:       &fakeRecoverer{rec: rec},
		Discovery:       &fakeDiscovery{rec: rec},
		Prewarmer:       &fakePrewarmer{rec: rec},
		Listener:        &recordingListener{Listener: l, rec: rec},
		Clock:           fixedClock,
		ShutdownTimeout: 2 * time.Second,
		Trace:           rec.add,
	}
	return h
}

// start runs d in a goroutine and returns its result channel; the run is
// cancelled at cleanup.
func start(t *testing.T, d *Daemon) <-chan error {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	errc := make(chan error, 1)
	go func() { errc <- d.Run(ctx) }()
	t.Cleanup(func() {
		cancel()
		select {
		case <-errc:
		case <-time.After(5 * time.Second):
			t.Error("Run did not return after cancel")
		}
	})
	return errc
}

// waitState polls d.Status() until pred holds or the timeout expires.
func waitState(t *testing.T, d *Daemon, timeout time.Duration, pred func(string) bool) string {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for {
		got := d.Status().State
		if pred(got) {
			return got
		}
		if time.Now().After(deadline) {
			t.Fatalf("Status().State = %q after %v, want the awaited state", got, timeout)
		}
		time.Sleep(5 * time.Millisecond)
	}
}
