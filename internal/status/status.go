package status

import (
	"time"

	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/history"
	"github.com/adeelahmad/snapback/internal/provider"
	"github.com/adeelahmad/snapback/internal/readerpolicy"
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
	panic("SUB-AGENT-TODO: derive overall state from phase and repos per tasks.md § T2a; degraded on any failed repo or empty ready map; sort Repos by ID")
}
