package history

import (
	"context"
	"errors"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"sync"
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
type Supervisor struct {
	mounts  map[string]provider.Mounter
	baseDir string
	backoff Backoff

	mu       sync.Mutex
	states   map[string]RepoState
	handles  map[string]provider.MountHandle
	order    []string
	stopping bool
	stop     chan struct{}
	wg       sync.WaitGroup
}

// NewSupervisor returns a supervisor for mounts, each mounted under baseDir.
func NewSupervisor(mounts map[string]provider.Mounter, baseDir string, backoff Backoff) *Supervisor {
	if backoff.After == nil {
		backoff.After = time.After
	}
	return &Supervisor{
		mounts:  mounts,
		baseDir: baseDir,
		backoff: backoff,
		states:  make(map[string]RepoState, len(mounts)),
		handles: make(map[string]provider.MountHandle, len(mounts)),
		stop:    make(chan struct{}),
	}
}

// Start mounts every repository and returns once each is ready or failed.
func (s *Supervisor) Start(ctx context.Context) error {
	// Mounts outlive Start, so they must not inherit its cancellation or deadline.
	mountCtx := context.WithoutCancel(ctx)
	for _, repo := range slices.Sorted(maps.Keys(s.mounts)) {
		s.mu.Lock()
		s.states[repo] = StateStarting
		s.order = append(s.order, repo)
		s.mu.Unlock()

		h, err := s.mount(mountCtx, repo)
		s.mu.Lock()
		if err != nil {
			s.states[repo] = StateFailed
		} else {
			s.states[repo] = StateReady
			s.handles[repo] = h
		}
		s.mu.Unlock()

		s.wg.Add(1)
		go s.watch(mountCtx, repo, h)
	}
	return nil
}

// mount creates the repository's mount directory, starts the mount and waits for it to serve.
func (s *Supervisor) mount(ctx context.Context, repo string) (provider.MountHandle, error) {
	dir := filepath.Join(s.baseDir, repo)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	h, err := s.mounts[repo].StartMount(ctx, dir)
	if err != nil {
		return nil, err
	}
	if err := h.Ready(ctx); err != nil {
		_ = h.Stop(context.WithoutCancel(ctx))
		return nil, err
	}
	return h, nil
}

// watch restarts repo's mount with exponential backoff whenever it ends, until Stop.
func (s *Supervisor) watch(ctx context.Context, repo string, h provider.MountHandle) {
	defer s.wg.Done()
	delay := s.backoff.Initial
	for {
		if h != nil {
			select {
			case <-h.Done():
			case <-s.stop:
				return
			}
		}
		s.mu.Lock()
		if s.stopping {
			s.mu.Unlock()
			return
		}
		s.states[repo] = StateFailed
		s.mu.Unlock()

		select {
		case <-s.backoff.After(delay):
		case <-s.stop:
			return
		}
		delay = min(2*delay, s.backoff.Max)

		var err error
		if h, err = s.mount(ctx, repo); err != nil {
			h = nil
			continue
		}
		s.mu.Lock()
		if s.stopping {
			s.mu.Unlock()
			_ = h.Stop(ctx)
			return
		}
		s.states[repo] = StateReady
		s.handles[repo] = h
		s.mu.Unlock()
		delay = s.backoff.Initial
	}
}

// States returns a copy of each repository's current state.
func (s *Supervisor) States() map[string]RepoState {
	s.mu.Lock()
	defer s.mu.Unlock()
	return maps.Clone(s.states)
}

// Stop stops every mount in reverse start order and prevents restarts.
func (s *Supervisor) Stop(ctx context.Context) error {
	s.mu.Lock()
	if s.stopping {
		s.mu.Unlock()
		return nil
	}
	s.stopping = true
	close(s.stop)
	var handles []provider.MountHandle
	for _, repo := range slices.Backward(s.order) {
		if h := s.handles[repo]; h != nil {
			handles = append(handles, h)
		}
	}
	s.mu.Unlock()

	var errs []error
	for _, h := range handles {
		if err := h.Stop(ctx); err != nil {
			errs = append(errs, err)
		}
	}
	s.wg.Wait()

	s.mu.Lock()
	for _, repo := range s.order {
		s.states[repo] = StateStopped
	}
	s.mu.Unlock()
	return errors.Join(errs...)
}
