package daemon

import (
	"context"
	"errors"
	"fmt"
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
	// Unlock, when set, releases a daemon lock the caller already holds for
	// the state dir. Run then uses that lock instead of taking its own and
	// calls Unlock once on return.
	Unlock func()
}

// defaultShutdownTimeout bounds shutdown when Deps leaves it unset.
const defaultShutdownTimeout = 10 * time.Second

// Daemon is a running Snapback daemon.
type Daemon struct {
	cfg  *config.Config
	deps Deps

	dedups *dedupSet

	// ops is cancelled by shutdown to stop finite operations such as a
	// refresh started over IPC.
	ops     context.Context
	stopOps context.CancelFunc

	mu          sync.Mutex
	phase       string
	refresh     RefreshResult
	lastRefresh time.Time
	recovery    *status.RecoverySummary
	cancel      context.CancelFunc // stops Run; nil until Run starts
}

// New returns a Daemon for cfg and deps.
func New(cfg *config.Config, deps Deps) *Daemon {
	if deps.Clock == nil {
		deps.Clock = time.Now
	}
	if deps.ShutdownTimeout <= 0 {
		deps.ShutdownTimeout = defaultShutdownTimeout
	}
	ops, stopOps := context.WithCancel(context.Background())
	return &Daemon{cfg: cfg, deps: deps, phase: "starting", dedups: newDedupSet(), ops: ops, stopOps: stopOps}
}

// onceListener closes its Listener at most once, so the shutdown sequence
// and ipc.Serve can both close it.
type onceListener struct {
	net.Listener
	once sync.Once
	err  error
}

func (l *onceListener) Close() error {
	l.once.Do(func() { l.err = l.Listener.Close() })
	return l.err
}

func (d *Daemon) trace(step string) {
	if d.deps.Trace != nil {
		d.deps.Trace(step)
	}
}

// Run starts the daemon and blocks until ctx is done.
func (d *Daemon) Run(ctx context.Context) error {
	unlock := d.deps.Unlock
	if d.cfg == nil || len(d.cfg.Repositories) == 0 || len(d.cfg.Roots) == 0 {
		if unlock != nil {
			unlock()
		}
		return errcode.New(errcode.InvalidConfig, "daemon run", errors.New("config needs repositories and roots"))
	}

	if unlock == nil {
		var err error
		unlock, err = lockFunc(d.cfg.StateDir)
		if err != nil {
			return err
		}
	}
	defer unlock()
	d.trace("lock")

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	d.mu.Lock()
	d.cancel = cancel
	d.mu.Unlock()

	// IPC is served on a context Run's cancel does not reach, so the
	// shutdown op can still reply; the shutdown sequence closes l instead.
	l := &onceListener{Listener: d.deps.Listener}
	defer func() { _ = l.Close() }()
	go func() {
		_ = ipc.Serve(context.WithoutCancel(ctx), l, d.handle, ipc.ServeOptions{UID: uint32(os.Getuid())})
	}()
	d.trace("ipc")

	rep, err := d.deps.Recoverer.Recover(ctx)
	if err != nil {
		return err
	}
	d.mu.Lock()
	d.recovery = &status.RecoverySummary{Unmounted: rep.Unmounted, Foreign: rep.Foreign}
	d.mu.Unlock()
	if err := d.deps.Supervisor.Start(ctx); err != nil {
		return err
	}
	res, err := d.deps.Refresher.Refresh(ctx)
	if err != nil && errcode.Of(err) != errcode.RepoUnavailable {
		d.unmountStarted(context.WithoutCancel(ctx))
		return err
	}
	d.mu.Lock()
	d.refresh = res
	d.lastRefresh = d.deps.Clock()
	d.mu.Unlock()

	if err := d.deps.Discovery.Start(ctx); err != nil {
		d.unmountStarted(context.WithoutCancel(ctx))
		return err
	}
	d.deps.Prewarmer.Prewarm(ctx)

	d.mu.Lock()
	d.phase = "ready"
	d.mu.Unlock()

	<-ctx.Done()
	return d.shutdown(context.WithoutCancel(ctx), l)
}

// shutdown stops the daemon in order: stop answering IPC, cancel finite
// operations, then unmount the history view before the backend mounts. A
// busy mount is reported as errcode.MountFailure and never forced. Managed
// links are left in place.
func (d *Daemon) shutdown(ctx context.Context, l net.Listener) error {
	d.mu.Lock()
	d.phase = "stopping"
	d.mu.Unlock()

	_ = l.Close()
	d.stopOps()
	d.deps.Discovery.Stop()

	ctx, cancel := context.WithTimeout(ctx, d.deps.ShutdownTimeout)
	defer cancel()
	var errs []error
	if err := d.deps.History.Unmount(ctx); err != nil {
		errs = append(errs, mountErr(ctx, "history", err))
	}
	if err := d.deps.Supervisor.Stop(ctx); err != nil {
		errs = append(errs, mountErr(ctx, "backend", err))
	}
	return errors.Join(errs...)
}

// unmountStarted best-effort unmounts the mounts Supervisor.Start already
// brought up when a later startup step fails: the history catalog first,
// then the backend mounts, bounded by the shutdown timeout. It never kills
// a process and discards its own errors; the caller returns the original
// startup error.
func (d *Daemon) unmountStarted(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, d.deps.ShutdownTimeout)
	defer cancel()
	_ = d.deps.History.Unmount(ctx)
	_ = d.deps.Supervisor.Stop(ctx)
}

// mountErr reports a failed unmount of the named mount, naming a timeout
// when the shutdown deadline passed.
func mountErr(ctx context.Context, mount string, err error) error {
	if ctx.Err() != nil {
		err = fmt.Errorf("timed out: %w", err)
	}
	return errcode.New(errcode.MountFailure, "daemon shutdown", fmt.Errorf("unmount %s: %w", mount, err))
}

// Status returns the daemon status.
func (d *Daemon) Status() status.Snapshot {
	d.mu.Lock()
	phase, res, last, rec := d.phase, d.refresh, d.lastRefresh, d.recovery
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
		Recovery:      rec,
	}
}

// Run builds a Daemon and runs it.
func Run(ctx context.Context, cfg *config.Config, deps Deps) error {
	return New(cfg, deps).Run(ctx)
}
