package status

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/provider"
	"github.com/adeelahmad/snapback/internal/refresh"
)

func TestDeriveReadyAndDegraded(t *testing.T) {
	tests := []struct {
		name      string
		repos     map[string]RepoState
		wantState string
		wantRepos []Repo
	}{
		{
			name:      "all ready",
			repos:     map[string]RepoState{"repoB": "ready", "repoA": "ready"},
			wantState: "ready",
			wantRepos: []Repo{
				{ID: "repoA", State: "ready"},
				{ID: "repoB", State: "ready"},
			},
		},
		{
			name:      "one failed",
			repos:     map[string]RepoState{"repoB": "failed", "repoA": "ready"},
			wantState: "degraded",
			wantRepos: []Repo{
				{ID: "repoA", State: "ready"},
				{ID: "repoB", State: "failed", Code: errcode.RepoUnavailable},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotState, gotRepos := Derive("ready", tt.repos)
			if gotState != tt.wantState {
				t.Errorf("Derive(ready, %v) state = %q, want %q", tt.repos, gotState, tt.wantState)
			}
			if !reflect.DeepEqual(gotRepos, tt.wantRepos) {
				t.Errorf("Derive(ready, %v) repos = %+v, want %+v", tt.repos, gotRepos, tt.wantRepos)
			}
		})
	}
}

func TestDerivePhaseAndEmptyRepos(t *testing.T) {
	tests := []struct {
		phase string
		repos map[string]RepoState
		want  string
	}{
		{phase: "starting", repos: map[string]RepoState{"repoA": "failed"}, want: "starting"},
		{phase: "stopping", repos: map[string]RepoState{"repoA": "ready"}, want: "stopping"},
		{phase: "ready", repos: map[string]RepoState{}, want: "degraded"},
	}
	for _, tt := range tests {
		got, _ := Derive(tt.phase, tt.repos)
		if got != tt.want {
			t.Errorf("Derive(%q, %v) state = %q, want %q", tt.phase, tt.repos, got, tt.want)
		}
	}
}

func TestSnapshotJSONHasNoCredentials(t *testing.T) {
	cfg := &config.Config{
		Repositories: []config.Repository{{
			ID:           "repoA",
			Repository:   "rclone:gd:repo",
			PasswordFile: "/secret/pw",
			Environment:  map[string]string{"RCLONE_PASS": "hunter2"},
		}},
	}
	repos := make(map[string]RepoState)
	for _, r := range cfg.Repositories {
		repos[r.ID] = "ready"
	}
	state, out := Derive("ready", repos)
	snap := Snapshot{State: state, Repos: out}

	b, err := json.Marshal(snap)
	if err != nil {
		t.Fatalf("json.Marshal(%+v) error = %v", snap, err)
	}
	got := string(b)
	if !strings.Contains(got, `"repoA"`) {
		t.Errorf("json.Marshal(snapshot) = %s, want it to contain %q", got, `"repoA"`)
	}
	for _, secret := range []string{"/secret/pw", "rclone:gd", "hunter2"} {
		if strings.Contains(got, secret) {
			t.Errorf("json.Marshal(snapshot) = %s, want no %q", got, secret)
		}
	}

	var top map[string]json.RawMessage
	if err := json.Unmarshal(b, &top); err != nil {
		t.Fatalf("json.Unmarshal(%s) error = %v", got, err)
	}
	for _, key := range []string{"State", "Repos", "LastRefresh", "Generation", "EligibleCount", "Links", "Warm", "Pending", "Discovery", "Throttle", "WebURL"} {
		if _, ok := top[key]; !ok {
			t.Errorf("json.Marshal(snapshot) = %s, want key %q", got, key)
		}
	}
	var reposJSON []map[string]json.RawMessage
	if err := json.Unmarshal(top["Repos"], &reposJSON); err != nil {
		t.Fatalf("json.Unmarshal(Repos %s) error = %v", top["Repos"], err)
	}
	if len(reposJSON) != 1 {
		t.Fatalf("Repos JSON = %s, want 1 entry", top["Repos"])
	}
	for _, key := range []string{"ID", "State", "Code"} {
		if _, ok := reposJSON[0][key]; !ok {
			t.Errorf("Repos[0] JSON = %s, want key %q", top["Repos"], key)
		}
	}
}

func TestFromRefreshCopiesFields(t *testing.T) {
	idA := provider.SnapshotID(strings.Repeat("a", 64))
	idB := provider.SnapshotID(strings.Repeat("b", 64))
	at := time.Date(2026, 9, 22, 6, 0, 0, 0, time.UTC)
	newInput := func() refresh.Result {
		return refresh.Result{
			Generation:    3,
			At:            at,
			Stale:         true,
			Failed:        []string{"repoA"},
			Pending:       []provider.SnapshotID{idB},
			EligibleCount: map[string]int{"k": 2},
			Warm:          map[provider.SnapshotID]bool{idA: true},
		}
	}
	in := newInput()
	want := newInput()

	got := FromRefresh(in)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("FromRefresh(%+v) = %+v, want %+v", in, got, want)
	}

	in.Failed[0] = "mutated"
	in.Pending[0] = idA
	in.EligibleCount["k"] = 99
	in.EligibleCount["extra"] = 1
	in.Warm[idA] = false
	in.Warm[idB] = true
	if !reflect.DeepEqual(got, want) {
		t.Errorf("FromRefresh result after mutating its input = %+v, want %+v (input not copied)", got, want)
	}
}

func TestSnapshotCarriesRecoverySummary(t *testing.T) {
	b, err := json.Marshal(Snapshot{State: "ready"})
	if err != nil {
		t.Fatalf("json.Marshal(snapshot without recovery) error = %v", err)
	}
	var top map[string]json.RawMessage
	if err := json.Unmarshal(b, &top); err != nil {
		t.Fatalf("json.Unmarshal(%s) error = %v", b, err)
	}
	for _, key := range []string{"recovery", "Recovery"} {
		if _, ok := top[key]; ok {
			t.Errorf("json.Marshal(snapshot with nil Recovery) = %s, want no %q key", b, key)
		}
	}

	want := &RecoverySummary{
		Unmounted: []string{"/run/snapback/mnt/repoA"},
		Foreign:   []string{"/home/u/.snapshot"},
	}
	snap := Snapshot{State: "ready", Recovery: want}
	b, err = json.Marshal(snap)
	if err != nil {
		t.Fatalf("json.Marshal(%+v) error = %v", snap, err)
	}
	top = nil
	if err := json.Unmarshal(b, &top); err != nil {
		t.Fatalf("json.Unmarshal(%s) error = %v", b, err)
	}
	raw, ok := top["recovery"]
	if !ok {
		t.Fatalf("json.Marshal(snapshot with Recovery) = %s, want key %q", b, "recovery")
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatalf("json.Unmarshal(recovery %s) error = %v", raw, err)
	}
	if len(fields) != 2 {
		t.Errorf("recovery JSON = %s, want only the Unmounted and Foreign path lists", raw)
	}
	var got RecoverySummary
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("json.Unmarshal(recovery %s) error = %v", raw, err)
	}
	if !reflect.DeepEqual(&got, want) {
		t.Errorf("recovery JSON round trip = %+v, want %+v", got, *want)
	}
}
