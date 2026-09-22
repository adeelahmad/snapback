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

// Build turns the registered directories into one projection spec.
func Build(in Input) (projection.Spec, error) {
	panic("SUB-AGENT-TODO: tasks.md T5 - validate keys and root IDs (wrapped error, no spec on duplicate or empty key/root ID or HistoryTarget error); per dir emit roots/<root>/dirs/<key>/ with snapshots/<id> links into BackendMountDir, alias, by-date and latest links relative to snapshots/, skip Pending, no latest when nothing is eligible, only info.json with state unavailable when the repo is not ready; info.json (info.go) per plan Decisions: id/alias/time, rel as b64 object for invalid UTF-8, stale, pending, no credentials; deterministic and projectable")
}
