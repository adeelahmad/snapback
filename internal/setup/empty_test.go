package setup

import (
	"context"
	"slices"
	"testing"

	"github.com/adeelahmad/snapback/internal/config"
)

const planRoot = "/srv/media"

// planResult is the detection result the plan tests probe from.
func planResult() Result {
	return Result{
		RepoURI:        probeRepo,
		CredentialFile: probePassword,
		ResticPath:     probeRestic,
		Roots:          []string{planRoot},
		Hostname:       "local",
	}
}

func TestPlanAdvisesSnapWhenRepositoryIsEmpty(t *testing.T) {
	fr := &fakeProbeRunner{stdout: []byte("[]")}

	got, adv, err := Plan(context.Background(), fr.run, planResult())
	if err != nil {
		t.Fatalf("Plan() error = %v, want nil", err)
	}

	if want := "snapback snap " + planRoot; adv.Next != want {
		t.Errorf("Plan() next = %q, want %q", adv.Next, want)
	}
	want := []string{"repository has no snapshots yet"}
	if !slices.Equal(adv.Notes, want) {
		t.Errorf("Plan() notes = %q, want %q", adv.Notes, want)
	}
	if got.Hostname != "local" {
		t.Errorf("Plan() hostname = %q, want %q", got.Hostname, "local")
	}
	if len(got.PrefixMappings) != 0 {
		t.Errorf("Plan() prefix mappings = %+v, want none", got.PrefixMappings)
	}
}

func TestPlanTakesHostnameFromTheProbeAndAdvisesRun(t *testing.T) {
	const snapshots = `[
  {"id":"1111111111111111111111111111111111111111111111111111111111111111",
   "time":"2026-01-02T03:04:05Z","hostname":"h","paths":["/srv/media"]},
  {"id":"2222222222222222222222222222222222222222222222222222222222222222",
   "time":"2026-03-04T05:06:07Z","hostname":"h","paths":["/srv/media"]}
]`
	fr := &fakeProbeRunner{stdout: []byte(snapshots)}

	got, adv, err := Plan(context.Background(), fr.run, planResult())
	if err != nil {
		t.Fatalf("Plan() error = %v, want nil", err)
	}

	if len(fr.calls) != 1 {
		t.Errorf("Plan() made %d calls, want 1: %+v", len(fr.calls), fr.calls)
	}
	if got.Hostname != "h" {
		t.Errorf("Plan() hostname = %q, want %q", got.Hostname, "h")
	}
	if adv.Next != nextRun {
		t.Errorf("Plan() next = %q, want %q", adv.Next, nextRun)
	}
	if len(adv.Notes) != 0 {
		t.Errorf("Plan() notes = %q, want none", adv.Notes)
	}
	var want []config.PrefixMapping
	if !slices.Equal(got.PrefixMappings, want) {
		t.Errorf("Plan() prefix mappings = %+v, want %+v", got.PrefixMappings, want)
	}
}
