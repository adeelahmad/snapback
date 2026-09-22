package history

import (
	"context"
	"time"

	"github.com/adeelahmad/snapback/internal/provider"
)

// RepoState is a repository mount's lifecycle state.
type RepoState string

// Repository mount states.
const (
	StateStarting RepoState = "starting"
	StateReady    RepoState = "ready"
	StateFailed   RepoState = "failed"
	StateStopped  RepoState = "stopped"
)

// Backoff configures restart delays.
type Backoff struct {
	Initial time.Duration
	Max     time.Duration
	After   func(time.Duration) <-chan time.Time
}

// Supervisor runs one private mount per repository.
type Supervisor struct{}

// NewSupervisor returns a supervisor for mounts, each mounted under baseDir.
func NewSupervisor(mounts map[string]provider.Mounter, baseDir string, backoff Backoff) *Supervisor {
	panic("SUB-AGENT-TODO: tasks.md T6 - store mounts, baseDir and backoff (default After to time.After); init the states map and mutex")
}

// Start mounts every repository and returns once each is ready or failed.
func (s *Supervisor) Start(ctx context.Context) error {
	panic("SUB-AGENT-TODO: tasks.md T6 - per repo in sorted order StartMount(context without deadline, baseDir/<repo>) and wait for Ready; a failed first start is recorded as StateFailed, not returned; spawn a watcher goroutine per repo that restarts on Done() with exponential backoff (reset after Ready) until Stop")
}

// States returns a copy of each repository's current state.
func (s *Supervisor) States() map[string]RepoState {
	panic("SUB-AGENT-TODO: tasks.md T6 - return a copy of the states map under the mutex")
}

// Stop stops every mount in reverse start order and prevents restarts.
func (s *Supervisor) Stop(ctx context.Context) error {
	panic("SUB-AGENT-TODO: tasks.md T6 - mark stopping so watchers do not restart, stop handles in reverse start order, wait for watchers, set StateStopped, join errors")
}
