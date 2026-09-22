// agentic:shim

// Package status is a compile shim for S3-10 T2a; the scaffolder replaces it.
package status

import (
	"time"

	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/provider"
	"github.com/adeelahmad/snapback/internal/readerpolicy"
)

// RepoState stands in for history.RepoState until S3-05 lands on the chain.
type RepoState string

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

// Derive is deliberately wrong.
func Derive(phase string, repos map[string]RepoState) (string, []Repo) {
	return "shim", nil
}
