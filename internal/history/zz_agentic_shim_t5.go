// agentic:shim
package history

import (
	"github.com/adeelahmad/snapback/internal/aliases"
	"github.com/adeelahmad/snapback/internal/projection"
	"github.com/adeelahmad/snapback/internal/provider"
	"github.com/adeelahmad/snapback/internal/resolver"
)

// Dir is one registered directory and its eligible snapshots.
type Dir struct {
	Key, RootID, Rel, RepoID string
	Eligible                 []resolver.Eligible
	Aliases                  aliases.Set
	Pending                  []provider.SnapshotID
}

// Input is everything Build needs.
type Input struct {
	BackendMountDir string
	Dirs            []Dir
	Stale           bool
	Repos           map[string]RepoState
}

// Build is a deliberately wrong compile shim.
func Build(_ Input) (projection.Spec, error) {
	return projection.Spec{Links: []projection.Link{{Name: "shim", Target: "/shim"}}}, nil
}
