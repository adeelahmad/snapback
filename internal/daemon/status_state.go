package daemon

import (
	"fmt"

	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/history"
	"github.com/adeelahmad/snapback/internal/status"
)

// stateReasons phrases the error codes a repository reports in its state.
var stateReasons = map[errcode.Code]string{
	errcode.MountFailure:    "mount failed",
	errcode.RepoUnavailable: "repository unavailable",
}

// stateReason phrases code, and always names the command that diagnoses it.
func stateReason(code errcode.Code) string {
	switch {
	case code == "":
		return "no error code reported; run: snapback doctor"
	case stateReasons[code] != "":
		return stateReasons[code] + "; run: snapback doctor"
	default:
		return string(code) + "; run: snapback doctor"
	}
}

// explainState reports a disagreement between the top-level state and the
// per-repository states, or "" while the two agree. The transitional states
// "starting" and "stopping" describe the daemon, not the repositories, so
// they never disagree.
func explainState(s status.Snapshot) string {
	if s.State == "starting" || s.State == "stopping" {
		return ""
	}
	ready := s.State == string(history.StateReady)
	for _, repo := range s.Repos {
		if repo.State == string(history.StateReady) {
			continue
		}
		if ready {
			return fmt.Sprintf("state: ready but repository %s is %s (%s)",
				repo.ID, repo.State, stateReason(repo.Code))
		}
		return ""
	}
	if !ready {
		if len(s.Repos) == 0 {
			return ""
		}
		return fmt.Sprintf("state: %s but every repository is ready "+
			"(no repository reports a problem)", s.State)
	}
	if len(s.Repos) == 0 {
		return "state: ready but no repositories are configured"
	}
	return ""
}
