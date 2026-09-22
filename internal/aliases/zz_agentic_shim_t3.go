// agentic:shim

package aliases

import (
	"time"

	"github.com/adeelahmad/snapback/internal/provider"
)

func rsnapshotView(_ []provider.Snapshot, _ *time.Location, _ Keep) map[string]provider.SnapshotID {
	return map[string]provider.SnapshotID{"shim.0": "shim"}
}
