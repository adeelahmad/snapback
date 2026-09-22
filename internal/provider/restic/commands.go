package restic

import (
	"context"

	"github.com/adeelahmad/snapback/internal/provider"
)

// Validate checks that the repository is reachable and returns its identity.
func (p *Provider) Validate(ctx context.Context) (provider.Identity, error) {
	panic("SUB-AGENT-TODO: run cat config via Runner.Run, parse JSON id/version into provider.Identity; non-64-hex id is an error; failures through classify")
}

// List returns every snapshot in the repository.
func (p *Provider) List(ctx context.Context) ([]provider.Snapshot, error) {
	panic("SUB-AGENT-TODO: run snapshots --json, parse into []provider.Snapshot keeping full IDs and relative paths verbatim; null -> empty; non-64-hex id -> error; ignore stderr for parsing; failures through classify")
}

// Snap backs up req.Path and returns the new snapshot's ID.
func (p *Provider) Snap(ctx context.Context, req provider.SnapRequest) (provider.SnapshotID, error) {
	panic("SUB-AGENT-TODO: require absolute req.Path and non-empty req.Host, run snapArgs, return the summary line's snapshot_id; missing summary or bad id -> error; failures through classify")
}
