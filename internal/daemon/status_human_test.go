package daemon

import (
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/history"
	"github.com/adeelahmad/snapback/internal/readerpolicy"
	"github.com/adeelahmad/snapback/internal/status"
)

func TestRenderHuman(t *testing.T) {
	at := time.Date(2026, 3, 5, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		in   status.Snapshot
		want string
	}{
		{
			name: "populated",
			in: status.Snapshot{
				State: "degraded",
				Repos: []status.Repo{
					{ID: "laptop", State: string(history.StateReady)},
					{ID: "archive", State: "failed", Code: errcode.RepoUnavailable},
					{ID: "photos", State: "failed", Code: errcode.RepoUnavailable},
				},
				Links:         3,
				EligibleCount: map[string]int{"k1": 4, "k2": 0, "k3": 2},
				Discovery:     "seed",
				Generation:    7,
				LastRefresh:   at,
				Prewarm:       status.PrewarmSummary{Warm: 4, Cold: 1, Pending: 2, LastPrewarm: at},
				WebURL:        "http://127.0.0.1:7788/",
				Throttle: []readerpolicy.ThrottleEvent{
					{PID: 1, Process: "rg", Rule: "deny", At: at},
					{PID: 7, Process: "code", Rule: "burst", At: at},
				},
			},
			want: "snapback: degraded\n" +
				"repositories: 1 ready, 2 failed (archive, photos)\n" +
				"links: 3 (2 with history)\n" +
				"discovery: seed\n" +
				"refresh: generation 7 at 2026-03-05T10:00:00Z\n" +
				"prewarm: 4 warm, 2 pending\n" +
				"web: http://127.0.0.1:7788/\n" +
				"throttled: rg(1) deny, code(7) burst\n",
		},
		{
			name: "minimal",
			in:   status.Snapshot{State: "starting"},
			want: "snapback: starting\n" +
				"repositories: 0 ready, 0 failed\n" +
				"links: 0 (0 with history)\n" +
				"discovery: none\n" +
				"refresh: never\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := RenderHuman(tc.in); got != tc.want {
				t.Errorf("RenderHuman(%s) = %q, want %q", tc.name, got, tc.want)
			}
		})
	}
}

func TestRenderHumanSortsFailedRepoIDs(t *testing.T) {
	in := status.Snapshot{
		State: "degraded",
		Repos: []status.Repo{
			{ID: "zulu", State: "failed"},
			{ID: "alpha", State: "failed"},
			{ID: "mike", State: "failed"},
		},
		Discovery: "config",
	}
	want := "snapback: degraded\n" +
		"repositories: 0 ready, 3 failed (alpha, mike, zulu)\n" +
		"links: 0 (0 with history)\n" +
		"discovery: config\n" +
		"refresh: never\n"

	if got := RenderHuman(in); got != want {
		t.Errorf("RenderHuman(unsorted) = %q, want %q", got, want)
	}
}
