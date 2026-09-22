package restic

import (
	"context"

	"github.com/adeelahmad/snapback/internal/provider"
)

// Probe reports whether treePath exists inside snapshot id under mountDir.
func (p *Provider) Probe(ctx context.Context, mountDir string, id provider.SnapshotID, treePath string) (provider.ProbeResult, error) {
	panic("SUB-AGENT-TODO: per plan decisions; invalid ID or a treePath escaping the snapshot root after cleaning is an error; otherwise stat under SnapshotRoot and map to provider.ProbeResult")
}
