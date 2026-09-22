// agentic:shim
package restic

import (
	"context"
	"errors"

	"github.com/adeelahmad/snapback/internal/provider"
)

// StartMount is a T6 compile shim with a deliberately wrong body.
func (p *Provider) StartMount(ctx context.Context, dir string) (provider.MountHandle, error) {
	return nil, errors.New("shim: StartMount not implemented")
}

// SnapshotRoot is a T6 compile shim with a deliberately wrong body.
func (p *Provider) SnapshotRoot(mountDir string, id provider.SnapshotID) string {
	return ""
}

// Probe is a T6 compile shim with a deliberately wrong body.
func (p *Provider) Probe(ctx context.Context, mountDir string, id provider.SnapshotID, treePath string) (provider.ProbeResult, error) {
	return 0, nil
}
