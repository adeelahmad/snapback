// agentic:shim
package refresh

import "github.com/adeelahmad/snapback/internal/provider"

// Reconcile is a compile shim for S3-11 T3.
func Reconcile(listed []provider.Snapshot, visible []provider.SnapshotID) map[provider.SnapshotID]bool {
	return map[provider.SnapshotID]bool{"shim": true}
}

// MountIDs is a compile shim for S3-11 T3.
func MountIDs(backendMountDir string) func(repoID string) ([]provider.SnapshotID, error) {
	return func(string) ([]provider.SnapshotID, error) { return nil, nil }
}
