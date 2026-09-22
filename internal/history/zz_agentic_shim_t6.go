// agentic:shim
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

// NewSupervisor is a compile shim.
func NewSupervisor(mounts map[string]provider.Mounter, baseDir string, backoff Backoff) *Supervisor {
	return &Supervisor{}
}

// Start is a compile shim that starts nothing.
func (s *Supervisor) Start(ctx context.Context) error { return nil }

// States is a compile shim that reports nothing.
func (s *Supervisor) States() map[string]RepoState { return map[string]RepoState{} }

// Stop is a compile shim that stops nothing.
func (s *Supervisor) Stop(ctx context.Context) error { return nil }
