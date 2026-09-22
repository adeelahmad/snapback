// Package providertest provides an in-memory provider fake for tests.
package providertest

import (
	"context"
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
	panic("SUB-AGENT-TODO: T2 Validate: under f.mu return f.Identity, f.ValidateErr (zero Identity when the error is set)")
}

// List returns a copy of f.Snapshots, or f.ListErr when set.
func (f *Fake) List(context.Context) ([]provider.Snapshot, error) {
	panic("SUB-AGENT-TODO: T2 List: under f.mu return f.ListErr if set, else a copy of f.Snapshots so callers cannot mutate it")
}

// StartMount returns a *FakeMount for dir, or f.MountErr when set.
func (f *Fake) StartMount(ctx context.Context, dir string) (provider.MountHandle, error) {
	panic("SUB-AGENT-TODO: T2 StartMount: return f.MountErr if set, else &FakeMount{dir: dir, done: make(chan struct{})}")
}

// SnapshotRoot returns <mountDir>/ids/<id>, matching the real provider.
func (f *Fake) SnapshotRoot(mountDir string, id provider.SnapshotID) string {
	panic("SUB-AGENT-TODO: T2 SnapshotRoot: return filepath.Join(mountDir, \"ids\", string(id))")
}

// Probe returns the scripted result for id and treePath; a missing entry is provider.ProbeAbsent.
func (f *Fake) Probe(ctx context.Context, mountDir string, id provider.SnapshotID, treePath string) (provider.ProbeResult, error) {
	panic("SUB-AGENT-TODO: T2 Probe: under f.mu return f.ProbeErr if set, else f.Probes[id][treePath] or provider.ProbeAbsent when missing")
}

// Snap records req and returns f.SnapID, or f.SnapErr when set.
func (f *Fake) Snap(ctx context.Context, req provider.SnapRequest) (provider.SnapshotID, error) {
	panic("SUB-AGENT-TODO: T2 Snap: under f.mu return f.SnapErr if set, else append a deep copy of req (Tags/Excludes cloned) to f.snapReqs and return f.SnapID")
}

// SnapRequests returns a copy of the requests passed to Snap, in call order.
func (f *Fake) SnapRequests() []provider.SnapRequest {
	panic("SUB-AGENT-TODO: T2 SnapRequests: under f.mu return a copy of f.snapReqs (clone Tags/Excludes) so callers cannot mutate it")
}

// Prewarm marks each ID warm, or fails each with f.PrewarmErr when set.
func (f *Fake) Prewarm(ctx context.Context, ids []provider.SnapshotID, concurrency int) []provider.PrewarmResult {
	panic("SUB-AGENT-TODO: T2 Prewarm: return one PrewarmResult per id in input order: {ID, Warm: true} or {ID, Err: f.PrewarmErr} when set")
}

// FakeMount is the provider.MountHandle returned by Fake.StartMount.
type FakeMount struct {
	dir      string
	done     chan struct{}
	stopOnce sync.Once
}

// Dir returns the mount directory.
func (m *FakeMount) Dir() string {
	panic("SUB-AGENT-TODO: T2 Dir: return m.dir")
}

// Ready returns nil immediately.
func (m *FakeMount) Ready(context.Context) error {
	panic("SUB-AGENT-TODO: T2 Ready: return nil")
}

// Done returns a channel closed after Die or Stop.
func (m *FakeMount) Done() <-chan struct{} {
	panic("SUB-AGENT-TODO: T2 Done: return m.done")
}

// Stop closes Done; it is idempotent.
func (m *FakeMount) Stop(context.Context) error {
	panic("SUB-AGENT-TODO: T2 Stop: m.stopOnce.Do(close(m.done)); return nil (idempotent)")
}

// Die simulates the mount process exiting unexpectedly by closing Done.
func (m *FakeMount) Die() {
	panic("SUB-AGENT-TODO: T2 Die: m.stopOnce.Do(close(m.done))")
}
