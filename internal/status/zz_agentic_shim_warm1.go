// agentic:shim

package status

import (
	"time"

	"github.com/adeelahmad/snapback/internal/provider"
)

// PrewarmSummary counts snapshots by pre-warm state and records when the
// daemon last pre-warmed.
type PrewarmSummary struct {
	Warm        int
	Cold        int
	Pending     int
	LastPrewarm time.Time
}

// SummarizePrewarm summarizes one pre-warm pass.
func SummarizePrewarm(results []provider.PrewarmResult, pending int, at time.Time) PrewarmSummary {
	return PrewarmSummary{Warm: -1, Cold: -1, Pending: -1}
}
