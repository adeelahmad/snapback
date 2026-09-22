package daemon

import (
	"testing"

	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/status"
)

func TestExplainState(t *testing.T) {
	tests := []struct {
		name string
		snap status.Snapshot
		want string
	}{
		{
			name: "all ready agrees",
			snap: status.Snapshot{State: "ready", Repos: []status.Repo{
				{ID: "archive", State: "ready"},
				{ID: "laptop", State: "ready"},
			}},
			want: "",
		},
		{
			name: "degraded with a failed repository agrees",
			snap: status.Snapshot{State: "degraded", Repos: []status.Repo{
				{ID: "archive", State: "failed", Code: errcode.RepoUnavailable},
				{ID: "laptop", State: "ready"},
			}},
			want: "",
		},
		{
			name: "ready but a repository failed",
			snap: status.Snapshot{State: "ready", Repos: []status.Repo{
				{ID: "archive", State: "failed", Code: errcode.MountFailure},
				{ID: "laptop", State: "ready"},
			}},
			want: "state: ready but repository archive is failed (mount failed; run: snapback doctor)",
		},
		{
			name: "ready but a repository stopped without a code",
			snap: status.Snapshot{State: "ready", Repos: []status.Repo{
				{ID: "laptop", State: "stopped"},
			}},
			want: "state: ready but repository laptop is stopped " +
				"(no error code reported; run: snapback doctor)",
		},
		{
			name: "degraded but every repository is ready",
			snap: status.Snapshot{State: "degraded", Repos: []status.Repo{
				{ID: "archive", State: "ready"},
				{ID: "laptop", State: "ready"},
			}},
			want: "state: degraded but every repository is ready (no repository reports a problem)",
		},
		{
			name: "ready without repositories",
			snap: status.Snapshot{State: "ready"},
			want: "state: ready but no repositories are configured",
		},
		{
			name: "degraded without repositories agrees",
			snap: status.Snapshot{State: "degraded"},
			want: "",
		},
		{
			name: "starting is transitional",
			snap: status.Snapshot{State: "starting", Repos: []status.Repo{
				{ID: "archive", State: "failed", Code: errcode.MountFailure},
			}},
			want: "",
		},
		{
			name: "stopping is transitional",
			snap: status.Snapshot{State: "stopping", Repos: []status.Repo{
				{ID: "archive", State: "ready"},
			}},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := explainState(tt.snap); got != tt.want {
				t.Errorf("explainState(%s) = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}
