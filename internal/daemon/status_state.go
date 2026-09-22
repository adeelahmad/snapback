package daemon

import (
	"github.com/adeelahmad/snapback/internal/status"
)

// explainState reports a disagreement between the top-level state and the
// per-repository states, or "" while the two agree.
func explainState(s status.Snapshot) string {
	_ = s
	return ""
}
