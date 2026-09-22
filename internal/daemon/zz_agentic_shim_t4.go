// agentic:shim

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

// RefreshResult stands in for status.RefreshResult until S3-10 T2b ships it;
// the scaffolder should turn it into an alias of status.RefreshResult.
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
func New(cfg *config.Config, deps Deps) *Daemon { return &Daemon{} }

// Run starts the daemon and blocks until ctx is done.
func (d *Daemon) Run(ctx context.Context) error { return nil }

// Status returns the daemon status.
func (d *Daemon) Status() status.Snapshot { return status.Snapshot{State: "starting"} }

// Run builds a Daemon and runs it.
func Run(ctx context.Context, cfg *config.Config, deps Deps) error {
	return New(cfg, deps).Run(ctx)
}

// Lock takes the single-instance lock in stateDir.
func Lock(stateDir string) (unlock func(), err error) { return func() {}, nil }
