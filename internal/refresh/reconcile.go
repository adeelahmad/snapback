package refresh

import "github.com/adeelahmad/snapback/internal/provider"

// Reconcile returns the listed snapshots that are not yet visible in the mount.
func Reconcile(listed []provider.Snapshot, visible []provider.SnapshotID) (pending map[provider.SnapshotID]bool) {
	panic("SUB-AGENT-TODO: Reconcile(listed []provider.Snapshot, visible []provider.SnapshotID) (pending map[provider.SnapshotID]bool) — the snapshots that are listed but not visible are pending (S3-11 T3, plan.md T3 block)")
}

// MountIDs returns a function that reads <backendMountDir>/<repoID>/ids with
// os.ReadDir, keeps only 64-hex entry names, and errors when the directory is
// missing (an absent mount is a failure, not an empty repository).
func MountIDs(backendMountDir string) func(repoID string) ([]provider.SnapshotID, error) {
	panic("SUB-AGENT-TODO: MountIDs(backendMountDir string) func(repoID string) ([]provider.SnapshotID, error) — os.ReadDir(<backendMountDir>/<repoID>/ids), keep 64-hex names, error on missing dir (S3-11 T3, plan.md T3 block)")
}
