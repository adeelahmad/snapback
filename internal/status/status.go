package status

import (
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
	ID    string
	State string
	Code  errcode.Code
}

// Snapshot is the daemon status model.
type Snapshot struct {
	State         string
	Repos         []Repo
	LastRefresh   time.Time
	Generation    uint64
	EligibleCount map[string]int
	Links         int
	Warm          map[provider.SnapshotID]bool
	Pending       []provider.SnapshotID
	Discovery     string
	Throttle      []readerpolicy.ThrottleEvent
	WebURL        string
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
	panic("SUB-AGENT-TODO: deep copy of refresh.Result: copy slices and maps")
}
