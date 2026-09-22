package daemon

import (
	"context"
	"errors"
	"net"
	"os"
	"sync"
	"time"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/history"
	"github.com/adeelahmad/snapback/internal/ipc"
	"github.com/adeelahmad/snapback/internal/links"
	"github.com/adeelahmad/snapback/internal/provider"
	"github.com/adeelahmad/snapback/internal/readerpolicy"
	"github.com/adeelahmad/snapback/internal/recovery"
	"github.com/adeelahmad/snapback/internal/status"
)

// RefreshResult stands in for the refresh result type until S3-10 T2b swaps
// it for refresh.Result.
type RefreshResult struct {
	Generation    uint64
	At            time.Time
	Stale         bool
	Failed        []string
	Pending       []provider.SnapshotID
	EligibleCount map[string]int
	Warm          map[provider.SnapshotID]bool
}

// Supervisor runs the repository mounts.
type Supervisor interface {
	Start(ctx context.Context) error
	States() map[string]history.RepoState
	Stop(ctx context.Context) error
}

// HistoryMount is the published history view.
type HistoryMount interface {
	Unmount(ctx context.Context) error
}

// Refresher rebuilds the catalog.
type Refresher interface {
	Refresh(ctx context.Context) (RefreshResult, error)
}

// Linker ensures a directory's managed link.
type Linker interface {
	Ensure(ctx context.Context, dir string) (links.Result, error)
}

// Recoverer cleans up after a crashed daemon.
type Recoverer interface {
	Recover(ctx context.Context) (recovery.Report, error)
}

// Discovery watches roots for new directories.
type Discovery interface {
	Start(ctx context.Context) error
	Stop()
}

// Prewarmer warms eligible snapshots.
type Prewarmer interface {
	Prewarm(ctx context.Context) []provider.PrewarmResult
}

// Deps are the daemon's collaborators.
type Deps struct {
	Supervisor      Supervisor
	History         HistoryMount
	Refresher       Refresher
	Linker          Linker
	Recoverer       Recoverer
	Discovery       Discovery
	Prewarmer       Prewarmer
	Listener        net.Listener
	Throttle        func() []readerpolicy.ThrottleEvent
	Clock           func() time.Time
	ShutdownTimeout time.Duration
	// Trace, when set, is called with the daemon's own startup steps
	// ("lock", "ipc") so tests can check their order.
	Trace func(step string)
}

// defaultShutdownTimeout bounds shutdown when Deps leaves it unset.
const defaultShutdownTimeout = 10 * time.Second

// Daemon is a running Snapback daemon.
type Daemon struct {
	cfg  *config.Config
	deps Deps

	mu          sync.Mutex
	phase       string
	refresh     RefreshResult
	lastRefresh time.Time
}

// New returns a Daemon for cfg and deps.
func New(cfg *config.Config, deps Deps) *Daemon {
	if deps.Clock == nil {
		deps.Clock = time.Now
	}
	if deps.ShutdownTimeout <= 0 {
		deps.ShutdownTimeout = defaultShutdownTimeout
	}
	return &Daemon{cfg: cfg, deps: deps, phase: "starting"}
}

func (d *Daemon) trace(step string) {
	if d.deps.Trace != nil {
		d.deps.Trace(step)
	}
}

// Run starts the daemon and blocks until ctx is done.
func (d *Daemon) Run(ctx context.Context) error {
	if d.cfg == nil || len(d.cfg.Repositories) == 0 || len(d.cfg.Roots) == 0 {
		return errcode.New(errcode.InvalidConfig, "daemon run", errors.New("config needs repositories and roots"))
	}

	unlock, err := Lock(d.cfg.StateDir)
	if err != nil {
		return err
	}
	defer unlock()
	d.trace("lock")

	go func() {
		_ = ipc.Serve(ctx, d.deps.Listener, d.handle, ipc.ServeOptions{UID: uint32(os.Getuid())})
	}()
	d.trace("ipc")

	if _, err := d.deps.Recoverer.Recover(ctx); err != nil {
		return err
	}
	if err := d.deps.Supervisor.Start(ctx); err != nil {
		return err
	}
	res, err := d.deps.Refresher.Refresh(ctx)
	if err != nil && errcode.Of(err) != errcode.RepoUnavailable {
		return err
	}
	d.mu.Lock()
	d.refresh = res
	d.lastRefresh = d.deps.Clock()
	d.mu.Unlock()

	if err := d.deps.Discovery.Start(ctx); err != nil {
		return err
	}
	d.deps.Prewarmer.Prewarm(ctx)

	d.mu.Lock()
	d.phase = "ready"
	d.mu.Unlock()

	<-ctx.Done()
	return nil
}

// handle answers IPC requests until the op handlers land.
func (d *Daemon) handle(context.Context, ipc.Request) ipc.Response {
	return ipc.Response{Code: errcode.InvalidConfig, Error: "unknown op"}
}

// Status returns the daemon status.
func (d *Daemon) Status() status.Snapshot {
	d.mu.Lock()
	phase, res, last := d.phase, d.refresh, d.lastRefresh
	d.mu.Unlock()

	repos := d.deps.Supervisor.States()
	for _, id := range res.Failed {
		repos[id] = history.StateFailed
	}
	state, out := status.Derive(phase, repos)
	return status.Snapshot{
		State:         state,
		Repos:         out,
		LastRefresh:   last,
		Generation:    res.Generation,
		EligibleCount: res.EligibleCount,
		Warm:          res.Warm,
		Pending:       res.Pending,
	}
}

// Run builds a Daemon and runs it.
func Run(ctx context.Context, cfg *config.Config, deps Deps) error {
	return New(cfg, deps).Run(ctx)
}
