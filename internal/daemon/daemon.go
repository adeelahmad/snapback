package daemon

import (
	"context"
	"errors"
	"fmt"
	"maps"
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
	"github.com/adeelahmad/snapback/internal/refresh"
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

// Linker ensures a directory's managed link and serves link maintenance.
type Linker interface {
	Ensure(ctx context.Context, dir string) (links.Result, error)
	List() ([]links.Record, error)
	Repair(ctx context.Context) (links.RepairReport, error)
	RemoveManaged(ctx context.Context) (links.RepairReport, error)
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
	prewarmSum  status.PrewarmSummary
	// warm is the single source of truth for per-snapshot warm state: the
	// pre-warm pass and every refresh result merge into it.
	warm  map[provider.SnapshotID]bool
	links int // owned link records, recounted on refresh and link ops
	// mountFailed holds the repos whose mount failed; each carries
	// errcode.MountFailure until it is ready again or a refresh succeeds.
	mountFailed map[string]bool
	cancel      context.CancelFunc // stops Run; nil until Run starts
	loop        *refresh.Loop      // periodic refresh loop; nil until it starts
	linkQueued  bool               // a refresh for new links is scheduled
	linkTimer   *time.Timer        // fires the scheduled link refresh; nil when none
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
	mountFailed := make(map[string]bool)
	if err := d.deps.Supervisor.Start(ctx); err != nil {
		if errcode.Of(err) != errcode.MountFailure {
			return err
		}
		if pr, ok := d.deps.Supervisor.(interface{ MountFailures() []string }); ok {
			for _, id := range pr.MountFailures() {
				mountFailed[id] = true
			}
		} else {
			markAll(mountFailed, d.deps.Supervisor.States())
		}
	}
	res, err := d.deps.Refresher.Refresh(ctx)
	switch errcode.Of(err) {
	case errcode.MountFailure:
		markAll(mountFailed, d.deps.Supervisor.States())
	case errcode.RepoUnavailable:
	default:
		if err != nil {
			d.unmountStarted(context.WithoutCancel(ctx))
			return err
		}
	}
	d.mu.Lock()
	d.mountFailed = mountFailed
	d.refresh = res
	d.mergeWarm(res.Warm)
	d.lastRefresh = d.deps.Clock()
	d.mu.Unlock()

	if err := d.deps.Discovery.Start(ctx); err != nil {
		d.unmountStarted(context.WithoutCancel(ctx))
		return err
	}
	d.prewarm(ctx)
	d.countLinks()

	d.mu.Lock()
	d.phase = "ready"
	d.mu.Unlock()

	d.refreshEvery(ctx, d.cfg.Catalog.RefreshInterval)
	return d.shutdown(context.WithoutCancel(ctx), l)
}

// refreshEvery runs a refresh.Loop every interval until ctx is done, so a
// snapshot that was pending while restic's mount had not yet reloaded gets
// its links once it becomes visible. A non-positive interval only waits.
func (d *Daemon) refreshEvery(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		<-ctx.Done()
		return
	}
	loop := refresh.NewLoop(loopTarget{d}, interval, time.After)
	d.mu.Lock()
	d.loop = loop
	d.mu.Unlock()
	_ = loop.Run(ctx)
}

// linkRefreshDelay coalesces a burst of new links, such as seed linking many
// directories, into one refresh.
const linkRefreshDelay = 500 * time.Millisecond

// refreshForLinks schedules one refresh of the periodic loop linkRefreshDelay
// from now, so new links get their history without waiting for the next
// interval. It never blocks; calls made while one is scheduled coalesce.
func (d *Daemon) refreshForLinks() {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.linkQueued {
		return
	}
	d.linkQueued = true
	d.linkTimer = time.AfterFunc(linkRefreshDelay, func() {
		d.mu.Lock()
		stopping := d.phase == "stopping"
		d.linkQueued = false
		d.linkTimer = nil
		loop := d.loop
		d.mu.Unlock()
		if loop != nil && !stopping {
			loop.Trigger()
		}
	})
}

// dropLinkRefresh cancels a scheduled link refresh, so a burst still pending
// at shutdown never triggers a loop whose Run has returned.
func (d *Daemon) dropLinkRefresh() {
	d.mu.Lock()
	timer := d.linkTimer
	d.linkTimer = nil
	d.linkQueued = false
	d.mu.Unlock()
	if timer != nil {
		timer.Stop()
	}
}

// loopTarget adapts a Daemon to refresh.Target, recording each result for
// status through runRefresh.
type loopTarget struct{ d *Daemon }

func (t loopTarget) Refresh(ctx context.Context) (refresh.Result, error) {
	return refresh.Result{}, t.d.runRefresh(ctx)
}

func (t loopTarget) Prewarm(ctx context.Context) []provider.PrewarmResult {
	return t.d.prewarm(ctx)
}

// prewarm runs one pre-warm pass and records its summary for status.
func (d *Daemon) prewarm(ctx context.Context) []provider.PrewarmResult {
	results := d.deps.Prewarmer.Prewarm(ctx)
	d.mu.Lock()
	d.prewarmSum = status.SummarizePrewarm(results, len(d.refresh.Pending), d.deps.Clock())
	for _, r := range results {
		d.setWarm(r.ID, r.Warm && r.Err == nil)
	}
	d.mu.Unlock()
	return results
}

// mergeWarm folds a refresh result's warm state into d.warm. The caller
// holds d.mu.
func (d *Daemon) mergeWarm(m map[provider.SnapshotID]bool) {
	for id, warm := range m {
		d.setWarm(id, warm)
	}
}

// setWarm records the warm state of one snapshot. The caller holds d.mu.
func (d *Daemon) setWarm(id provider.SnapshotID, warm bool) {
	if d.warm == nil {
		d.warm = make(map[provider.SnapshotID]bool)
	}
	d.warm[id] = warm
}

// shutdown stops the daemon in order: stop answering IPC, cancel finite
// operations, then unmount the history view before the backend mounts. A
// busy mount is reported as errcode.MountFailure and never forced. Managed
// links are left in place.
func (d *Daemon) shutdown(ctx context.Context, l net.Listener) error {
	d.mu.Lock()
	d.phase = "stopping"
	d.mu.Unlock()
	d.dropLinkRefresh()

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
	defer d.mu.Unlock()
	phase, res, last, rec, pre := d.phase, d.refresh, d.lastRefresh, d.recovery, d.prewarmSum

	repos := d.deps.Supervisor.States()
	for _, id := range res.Failed {
		repos[id] = history.StateFailed
	}
	state, out := status.Derive(phase, repos)
	for i := range out {
		if !d.mountFailed[out[i].ID] {
			continue
		}
		if out[i].Code == errcode.RepoUnavailable {
			out[i].Code = errcode.MountFailure
		} else if out[i].State == string(history.StateReady) {
			delete(d.mountFailed, out[i].ID)
		}
	}
	return status.Snapshot{
		State:         state,
		Repos:         out,
		LastRefresh:   last,
		Generation:    res.Generation,
		EligibleCount: res.EligibleCount,
		Links:         d.links,
		Warm:          maps.Clone(d.warm),
		Prewarm:       pre,
		Pending:       res.Pending,
		Discovery:     d.cfg.Discovery.Mode,
		Recovery:      rec,
	}
}

// countLinks caches the number of owned link records for Status. A List
// error keeps the previous count.
func (d *Daemon) countLinks() {
	recs, err := d.deps.Linker.List()
	if err != nil {
		return
	}
	n := 0
	for _, r := range recs {
		if r.State == links.StateOwned {
			n++
		}
	}
	d.mu.Lock()
	d.links = n
	d.mu.Unlock()
}

// Run builds a Daemon and runs it.
func Run(ctx context.Context, cfg *config.Config, deps Deps) error {
	return New(cfg, deps).Run(ctx)
}

// markAll adds every repo in repos to set.
func markAll(set map[string]bool, repos map[string]history.RepoState) {
	for id := range repos {
		set[id] = true
	}
}
