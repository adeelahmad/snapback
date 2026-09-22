package restic

import "github.com/adeelahmad/snapback/internal/provider"

var _ provider.SnapshotProvider = (*Provider)(nil)
