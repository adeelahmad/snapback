package telemetry_test

import (
	"sort"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/telemetry"
)

// TestSchemaClosedGuardKeysHaveProducers is a closed-schema guard: it builds
// one sample event from each of the five event constructors and asserts that
// the union of every attribute key they produce is exactly telemetry.Keys(),
// as sets. If a key is ever added to Keys() without a constructor that emits
// it, or a constructor emits a key that Keys() no longer lists, this test
// fails.
func TestSchemaClosedGuardKeysHaveProducers(t *testing.T) {
	now := time.Unix(0, 0).UTC()

	setupCompleted, err := telemetry.SetupCompleted("1.4.1", "ok", time.Second, now)
	if err != nil {
		t.Fatalf("SetupCompleted() error = %v, want nil", err)
	}
	daemonStarted, err := telemetry.DaemonStarted("1.4.1", now)
	if err != nil {
		t.Fatalf("DaemonStarted() error = %v, want nil", err)
	}
	mountReady, err := telemetry.MountReady("1.4.1", time.Second, now)
	if err != nil {
		t.Fatalf("MountReady() error = %v, want nil", err)
	}
	doctorFailed, err := telemetry.DoctorFailed("1.4.1", telemetry.DoctorChecks()[0], now)
	if err != nil {
		t.Fatalf("DoctorFailed() error = %v, want nil", err)
	}
	errorEvent, err := telemetry.ErrorEvent("1.4.1", telemetry.ErrorCodes()[0], now)
	if err != nil {
		t.Fatalf("ErrorEvent() error = %v, want nil", err)
	}

	events := []telemetry.Event{setupCompleted, daemonStarted, mountReady, doctorFailed, errorEvent}

	got := map[string]bool{}
	for _, e := range events {
		for _, a := range e.Attrs {
			got[a.Key] = true
		}
	}

	// want is the closed key set, spelled out here as a literal (not derived
	// from telemetry.Keys()) so this assertion has independent teeth: it
	// would fail if a constructor's attrs and the package's Keys() ever drift
	// apart, in either direction.
	want := map[string]bool{
		"version":  true,
		"os":       true,
		"arch":     true,
		"check":    true,
		"code":     true,
		"duration": true,
		"outcome":  true,
	}

	if len(got) != len(want) {
		t.Fatalf("union of attr keys across all five events = %v, want %v", sortedKeys(got), sortedKeys(want))
	}
	for k := range want {
		if !got[k] {
			t.Errorf("union of attr keys across all five events = %v, missing want key %q", sortedKeys(got), k)
		}
	}
	for k := range got {
		if !want[k] {
			t.Errorf("union of attr keys across all five events = %v, has extra key %q not in want", sortedKeys(got), k)
		}
	}
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
