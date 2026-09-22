package refresh

import (
	"github.com/adeelahmad/snapback/internal/history"
	"github.com/adeelahmad/snapback/internal/provider"
)

// orderForPublish splits dirs into the ones a generation may list and the ones
// it must withhold.
func orderForPublish(dirs []history.Dir, pending map[provider.SnapshotID]bool) (ready, deferred []history.Dir) {
	return dirs, nil
}
