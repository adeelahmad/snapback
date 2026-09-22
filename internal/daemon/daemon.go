package daemon

import (
	"context"
	"net"
	"time"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/history"
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

// Daemon is a running Snapback daemon.
type Daemon struct{}

// New returns a Daemon for cfg and deps.
func New(cfg *config.Config, deps Deps) *Daemon {
	panic("SUB-AGENT-TODO: store cfg and deps; default Clock to time.Now and ShutdownTimeout per Decisions; initial state starting")
}

// Run starts the daemon and blocks until ctx is done.
func (d *Daemon) Run(ctx context.Context) error {
	panic("SUB-AGENT-TODO: reject nil cfg or empty Roots with invalid_configuration before any side effect; then Lock(stateDir) (trace lock), serve ipc on Listener (trace ipc), Recover, Supervisor.Start, Refresh (failed repos -> degraded, not fatal), Discovery.Start, Prewarm; set ready/degraded; never walk roots; block until ctx done")
}

// Status returns the daemon status.
func (d *Daemon) Status() status.Snapshot {
	panic("SUB-AGENT-TODO: build status.Snapshot from daemon state, Supervisor.States and last RefreshResult: State starting/ready/degraded, per-repo Code (repository_unavailable for failed), Generation, LastRefresh from Clock")
}

// Run builds a Daemon and runs it.
func Run(ctx context.Context, cfg *config.Config, deps Deps) error {
	return New(cfg, deps).Run(ctx)
}
