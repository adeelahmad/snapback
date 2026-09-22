// agentic:shim

// Package providertest provides an in-memory provider fake for tests.
package providertest

import (
	"context"
	"errors"

	"github.com/adeelahmad/snapback/internal/provider"
)

var errShim = errors.New("shim: not implemented")

// Fake is a compile shim; its bodies are deliberately wrong.
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
}

// Validate is a shim.
func (f *Fake) Validate(context.Context) (provider.Identity, error) {
	return provider.Identity{}, errShim
}

// List is a shim.
func (f *Fake) List(context.Context) ([]provider.Snapshot, error) { return nil, errShim }

// StartMount is a shim.
func (f *Fake) StartMount(context.Context, string) (provider.MountHandle, error) {
	return &FakeMount{}, errShim
}

// SnapshotRoot is a shim.
func (f *Fake) SnapshotRoot(string, provider.SnapshotID) string { return "" }

// Probe is a shim.
func (f *Fake) Probe(context.Context, string, provider.SnapshotID, string) (provider.ProbeResult, error) {
	return 0, errShim
}

// Snap is a shim.
func (f *Fake) Snap(context.Context, provider.SnapRequest) (provider.SnapshotID, error) {
	return "", errShim
}

// SnapRequests is a shim.
func (f *Fake) SnapRequests() []provider.SnapRequest { return nil }

// Prewarm is a shim.
func (f *Fake) Prewarm(context.Context, []provider.SnapshotID, int) []provider.PrewarmResult {
	return nil
}

// FakeMount is a compile shim; its bodies are deliberately wrong.
type FakeMount struct{}

// Dir is a shim.
func (m *FakeMount) Dir() string { return "" }

// Ready is a shim.
func (m *FakeMount) Ready(context.Context) error { return errShim }

// Done is a shim.
func (m *FakeMount) Done() <-chan struct{} { return nil }

// Stop is a shim.
func (m *FakeMount) Stop(context.Context) error { return errShim }

// Die is a shim.
func (m *FakeMount) Die() {}
