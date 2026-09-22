package web

import (
	"context"
	"sync"

	"github.com/adeelahmad/snapback/internal/config"
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
