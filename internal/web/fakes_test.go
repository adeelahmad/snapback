package web

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/provider"
)

// snapshot mirrors the fields of status.Snapshot the API tests read; it is
// the stand-in Backend.Status value until S3-10 merges.
type snapshot struct {
	State      string
	Generation int
}

// fakeBackend is a Backend with a fixed status and config that records
// SaveConfig calls.
type fakeBackend struct {
	mu      sync.Mutex
	snap    snapshot
	cfg     *config.Config
	rev     config.Revision
	newRev  config.Revision
	saveErr error
	saved   []*config.Config
	gotRevs []config.Revision
}

func (b *fakeBackend) Status() any { return b.snap }

func (b *fakeBackend) Config() (*config.Config, config.Revision, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.cfg, b.rev, nil
}

func (b *fakeBackend) SaveConfig(c *config.Config, rev config.Revision) (config.Revision, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.saved = append(b.saved, c)
	b.gotRevs = append(b.gotRevs, rev)
	if b.saveErr != nil {
		return "", b.saveErr
	}
	return b.newRev, nil
}

// fakeValidator returns errs in order, one per Validate call, then nil.
type fakeValidator struct {
	mu   sync.Mutex
	errs []error
}

func (v *fakeValidator) Validate(ctx context.Context, c *config.Config) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	if len(v.errs) == 0 {
		return nil
	}
	err := v.errs[0]
	v.errs = v.errs[1:]
	return err
}

// Snapshot IDs used by the history fixtures.
const (
	idA provider.SnapshotID = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	idB provider.SnapshotID = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
)

// fakeHistory serves one root, home, over a temp tree: the live root at
// base/live and snapshots idA and idB at base/snaps/<id>. It records List
// calls.
type fakeHistory struct {
	base     string
	entries  []Entry
	versions []Version

	mu        sync.Mutex
	listCalls int
}

// newFakeHistory creates the live and snapshot directories under a temp dir.
func newFakeHistory(t *testing.T) *fakeHistory {
	t.Helper()
	h := &fakeHistory{base: t.TempDir()}
	for _, d := range []string{h.live(), h.snapDir(idA), h.snapDir(idB)} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatalf("os.MkdirAll(%q) error = %v", d, err)
		}
	}
	return h
}

func (h *fakeHistory) live() string { return filepath.Join(h.base, "live") }

func (h *fakeHistory) snapDir(id provider.SnapshotID) string {
	return filepath.Join(h.base, "snaps", string(id))
}

func (h *fakeHistory) Roots() []Root {
	return []Root{{ID: "home", Path: h.live(), State: "ok"}}
}

func (h *fakeHistory) SnapshotDir(root string, id provider.SnapshotID) (string, time.Time, error) {
	switch {
	case root != "home":
		return "", time.Time{}, fmt.Errorf("unknown root %q", root)
	case id == idA:
		return h.snapDir(idA), time.Date(2026, 9, 20, 10, 0, 0, 0, time.UTC), nil
	case id == idB:
		return h.snapDir(idB), time.Date(2026, 9, 21, 10, 0, 0, 0, time.UTC), nil
	}
	return "", time.Time{}, fmt.Errorf("unknown snapshot %q", id)
}

func (h *fakeHistory) List(ctx context.Context, root, dir string, id provider.SnapshotID) ([]Entry, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.listCalls++
	return h.entries, nil
}

func (h *fakeHistory) Versions(ctx context.Context, root, file string) ([]Version, error) {
	return h.versions, nil
}

// lists returns the number of List calls so far.
func (h *fakeHistory) lists() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.listCalls
}
