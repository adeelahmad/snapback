package refresh

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/adeelahmad/snapback/internal/provider"
)

// Reconcile returns the listed snapshots that are not yet visible in the mount.
func Reconcile(listed []provider.Snapshot, visible []provider.SnapshotID) (pending map[provider.SnapshotID]bool) {
	visibleSet := make(map[provider.SnapshotID]bool, len(visible))
	for _, id := range visible {
		visibleSet[id] = true
	}
	pending = make(map[provider.SnapshotID]bool)
	for _, s := range listed {
		if !visibleSet[s.ID] {
			pending[s.ID] = true
		}
	}
	return pending
}

// MountIDs returns a function that reads <backendMountDir>/<repoID>/ids with
// os.ReadDir, keeps only 64-hex entry names, and errors when the directory is
// missing (an absent mount is a failure, not an empty repository).
func MountIDs(backendMountDir string) func(repoID string) ([]provider.SnapshotID, error) {
	return func(repoID string) ([]provider.SnapshotID, error) {
		dir := filepath.Join(backendMountDir, repoID, "ids")
		entries, err := os.ReadDir(dir)
		if err != nil {
			return nil, fmt.Errorf("read ids dir %s: %w", dir, err)
		}
		ids := make([]provider.SnapshotID, 0, len(entries))
		for _, e := range entries {
			id := provider.SnapshotID(e.Name())
			if id.Valid() {
				ids = append(ids, id)
			}
		}
		return ids, nil
	}
}
