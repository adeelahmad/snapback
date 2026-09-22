package telemetry

import (
	"reflect"
	"testing"
	"time"
)

// wantNames is the schema, spelled out here so that adding a sixth event, or
// reordering the five, cannot happen without editing this test.
var wantNames = []string{
	"setup.completed",
	"daemon.started",
	"mount.ready",
	"doctor.failed",
	"error",
}

func TestNamesReturnsTheClosedSetInOrder(t *testing.T) {
	got := Names()
	if !reflect.DeepEqual(got, wantNames) {
		t.Errorf("Names() = %q, want %q", got, wantNames)
	}
}

func TestNamesIsAClosedSet(t *testing.T) {
	got := Names()
	if len(got) != len(wantNames) {
		t.Fatalf("Names() has %d names, want exactly %d: %q", len(got), len(wantNames), got)
	}

	inSchema := make(map[string]bool, len(wantNames))
	for _, name := range wantNames {
		inSchema[name] = true
	}
	for _, name := range got {
		if !inSchema[name] {
			t.Errorf("Names() contains %q, which is not in the schema", name)
		}
	}

	emitted := make(map[string]bool, len(got))
	for _, name := range got {
		if emitted[name] {
			t.Errorf("Names() repeats %q", name)
		}
		emitted[name] = true
	}
	for _, name := range wantNames {
		if !emitted[name] {
			t.Errorf("Names() is missing schema name %q", name)
		}
	}
}

// TestEventNamesVarIsRangeable pins the package-level var the schema is built
// from, so other schema tests can range over it directly.
func TestEventNamesVarIsRangeable(t *testing.T) {
	var ranged []string
	ranged = append(ranged, eventNames...)
	if !reflect.DeepEqual(ranged, wantNames) {
		t.Errorf("ranging eventNames gave %q, want %q", ranged, wantNames)
	}
}

func TestIsName(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"setup.completed", true},
		{"daemon.started", true},
		{"mount.ready", true},
		{"doctor.failed", true},
		{"error", true},

		{"setup.started", false},
		{"", false},
		{"Error", false},
		{"ERROR", false},
		{"errors", false},
		{"error ", false},
		{" error", false},
		{"snapback.error", false},
		{"error.code", false},
		{"mount", false},
		{"mount.ready.ok", false},
		{"doctor.passed", false},
		{"daemon.stopped", false},
		{"setup.completed\n", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsName(tt.name); got != tt.want {
				t.Errorf("IsName(%q) = %t, want %t", tt.name, got, tt.want)
			}
		})
	}
}

// TestIsNameAgreesWithNames is the closed-set guard: a name can be added to
// IsName only by also adding it to Names, and to wantNames above.
func TestIsNameAgreesWithNames(t *testing.T) {
	for _, name := range wantNames {
		if !IsName(name) {
			t.Errorf("IsName(%q) = false, want true for a schema name", name)
		}
	}
	for _, name := range Names() {
		if !IsName(name) {
			t.Errorf("Names() reports %q but IsName(%q) = false", name, name)
		}
	}
}

func TestEventCarriesNameAttrsAndTime(t *testing.T) {
	ev := Event{
		Name:  "daemon.started",
		Attrs: []Attr{{Key: "os", Value: "linux"}},
		Time:  time.Unix(0, 0).UTC(),
	}
	if !IsName(ev.Name) {
		t.Errorf("IsName(%q) = false, want true for the event's own name", ev.Name)
	}
	if got := reflect.TypeOf(Event{}).NumField(); got != 3 {
		t.Errorf("Event has %d fields, want exactly 3 (Name, Attrs, Time)", got)
	}
}
