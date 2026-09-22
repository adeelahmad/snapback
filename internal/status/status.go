package status

import (
	"maps"
	"slices"
	"sort"
	"time"

	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/history"
	"github.com/adeelahmad/snapback/internal/provider"
	"github.com/adeelahmad/snapback/internal/readerpolicy"
	"github.com/adeelahmad/snapback/internal/refresh"
)

// RepoState is a repository mount's lifecycle state.
type RepoState = history.RepoState

// Repo is one repository's status.
type Repo struct {
	ID    string       `json:"id"`
	State string       `json:"state"`
	Code  errcode.Code `json:"code"`
}

// Snapshot is the daemon status model.
type Snapshot struct {
	State         string                       `json:"state"`
	Repos         []Repo                       `json:"repos"`
	LastRefresh   time.Time                    `json:"last_refresh"`
	Generation    uint64                       `json:"generation"`
	EligibleCount map[string]int               `json:"eligible_count"`
	Links         int                          `json:"links"`
	Warm          map[provider.SnapshotID]bool `json:"warm"`
	Prewarm       PrewarmSummary               `json:"prewarm"`
	Pending       []provider.SnapshotID        `json:"pending"`
	Discovery     string                       `json:"discovery"`
	Throttle      []readerpolicy.ThrottleEvent `json:"throttle"`
	WebURL        string                       `json:"web_url"`
	Recovery      *RecoverySummary             `json:"recovery,omitempty"`
}

// PrewarmSummary counts snapshots by pre-warm state and records when the
// daemon last pre-warmed.
type PrewarmSummary struct {
	Warm        int       `json:"warm"`
	Cold        int       `json:"cold"`
	Pending     int       `json:"pending"`
	LastPrewarm time.Time `json:"last_prewarm"`
}

// SummarizePrewarm summarizes one pre-warm pass: successful results count
// as warm and every other result as cold.
func SummarizePrewarm(results []provider.PrewarmResult, pending int, at time.Time) PrewarmSummary {
	sum := PrewarmSummary{Pending: pending, LastPrewarm: at}
	for _, r := range results {
		if r.Warm && r.Err == nil {
			sum.Warm++
		} else {
			sum.Cold++
		}
	}
	return sum
}

// RecoverySummary lists the paths the startup crash recovery cleaned up.
type RecoverySummary struct {
	Unmounted []string `json:"unmounted"`
	Foreign   []string `json:"foreign"`
}

// Derive computes the overall daemon state and the sorted per-repository
// status from the daemon's phase and each repository's current state.
func Derive(phase string, repos map[string]RepoState) (string, []Repo) {
	out := make([]Repo, 0, len(repos))
	for id, state := range repos {
		out = append(out, Repo{ID: id, State: string(state)})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })

	if phase == "starting" || phase == "stopping" {
		return phase, out
	}

	state := "ready"
	if len(out) == 0 {
		state = "degraded"
	}
	for i := range out {
		if out[i].State != string(history.StateReady) {
			state = "degraded"
			out[i].Code = errcode.RepoUnavailable
		}
	}
	return state, out
}

// FromRefresh returns a deep copy of r, so the status model never shares
// slices or maps with the refresh loop that produced it.
func FromRefresh(r refresh.Result) refresh.Result {
	r.Failed = slices.Clone(r.Failed)
	r.Pending = slices.Clone(r.Pending)
	r.EligibleCount = maps.Clone(r.EligibleCount)
	r.Warm = maps.Clone(r.Warm)
	return r
}
