// Package providertest provides an in-memory provider fake for tests.
package providertest

import (
	"context"
	"path/filepath"
	"slices"
	"sync"

	"github.com/adeelahmad/snapback/internal/provider"
)

// Fake is an in-memory provider.SnapshotProvider with scripted results and errors.
// It is safe for concurrent use.
type Fake struct {
	Identity    provider.Identity
	Snapshots   []provider.Snapshot
	Probes      map[provider.SnapshotID]map[string]provider.ProbeResult
	SnapID      provider.SnapshotID
	ValidateErr error
	ListErr     error
	MountErr    error
	ProbeErr    error
	SnapErr     error
	PrewarmErr  error

	mu       sync.Mutex
	snapReqs []provider.SnapRequest
}

// Validate returns f.Identity, or f.ValidateErr when set.
func (f *Fake) Validate(context.Context) (provider.Identity, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.ValidateErr != nil {
		return provider.Identity{}, f.ValidateErr
	}
	return f.Identity, nil
}

// List returns a copy of f.Snapshots, or f.ListErr when set.
func (f *Fake) List(context.Context) ([]provider.Snapshot, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.ListErr != nil {
		return nil, f.ListErr
	}
	return slices.Clone(f.Snapshots), nil
}

// StartMount returns a *FakeMount for dir, or f.MountErr when set.
func (f *Fake) StartMount(ctx context.Context, dir string) (provider.MountHandle, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.MountErr != nil {
		return nil, f.MountErr
	}
	return &FakeMount{dir: dir, done: make(chan struct{})}, nil
}

// SnapshotRoot returns <mountDir>/ids/<id>, matching the real provider.
func (f *Fake) SnapshotRoot(mountDir string, id provider.SnapshotID) string {
	return filepath.Join(mountDir, "ids", string(id))
}

// Probe returns the scripted result for id and treePath; a missing entry is provider.ProbeAbsent.
func (f *Fake) Probe(ctx context.Context, mountDir string, id provider.SnapshotID, treePath string) (provider.ProbeResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.ProbeErr != nil {
		return provider.ProbeAbsent, f.ProbeErr
	}
	if r, ok := f.Probes[id][treePath]; ok {
		return r, nil
	}
	return provider.ProbeAbsent, nil
}

// Snap records req and returns f.SnapID, or f.SnapErr when set.
func (f *Fake) Snap(ctx context.Context, req provider.SnapRequest) (provider.SnapshotID, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.SnapErr != nil {
		return "", f.SnapErr
	}
	f.snapReqs = append(f.snapReqs, cloneSnapRequest(req))
	return f.SnapID, nil
}

// SnapRequests returns a copy of the requests passed to Snap, in call order.
func (f *Fake) SnapRequests() []provider.SnapRequest {
	f.mu.Lock()
	defer f.mu.Unlock()
	reqs := make([]provider.SnapRequest, len(f.snapReqs))
	for i, r := range f.snapReqs {
		reqs[i] = cloneSnapRequest(r)
	}
	return reqs
}

// Prewarm marks each ID warm, or fails each with f.PrewarmErr when set.
func (f *Fake) Prewarm(ctx context.Context, ids []provider.SnapshotID, concurrency int) []provider.PrewarmResult {
	f.mu.Lock()
	err := f.PrewarmErr
	f.mu.Unlock()
	results := make([]provider.PrewarmResult, len(ids))
	for i, id := range ids {
		results[i] = provider.PrewarmResult{ID: id, Warm: err == nil, Err: err}
	}
	return results
}

func cloneSnapRequest(req provider.SnapRequest) provider.SnapRequest {
	req.Tags = slices.Clone(req.Tags)
	req.Excludes = slices.Clone(req.Excludes)
	return req
}

// FakeMount is the provider.MountHandle returned by Fake.StartMount.
type FakeMount struct {
	dir      string
	done     chan struct{}
	stopOnce sync.Once
}

// Dir returns the mount directory.
func (m *FakeMount) Dir() string {
	return m.dir
}

// Ready returns nil immediately.
func (m *FakeMount) Ready(context.Context) error {
	return nil
}

// Done returns a channel closed after Die or Stop.
func (m *FakeMount) Done() <-chan struct{} {
	return m.done
}

// Stop closes Done; it is idempotent.
func (m *FakeMount) Stop(context.Context) error {
	m.stopOnce.Do(func() { close(m.done) })
	return nil
}

// Die simulates the mount process exiting unexpectedly by closing Done.
func (m *FakeMount) Die() {
	m.stopOnce.Do(func() { close(m.done) })
}
