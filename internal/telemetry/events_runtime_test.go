package telemetry

import (
	"reflect"
	"runtime"
	"testing"
	"time"
)

// runtimeNow is a fixed instant, so the events under test carry no wall clock.
var runtimeNow = time.Date(2026, time.March, 4, 9, 30, 0, 0, time.UTC)

func TestDaemonStartedCarriesVersionOSAndArch(t *testing.T) {
	got, err := DaemonStarted("1.4.1", runtimeNow)
	if err != nil {
		t.Fatalf("DaemonStarted(%q, %v) error = %v, want nil", "1.4.1", runtimeNow, err)
	}
	if got.Name != "daemon.started" {
		t.Errorf("DaemonStarted(...).Name = %q, want %q", got.Name, "daemon.started")
	}
	if !got.Time.Equal(runtimeNow) {
		t.Errorf("DaemonStarted(...).Time = %v, want %v", got.Time, runtimeNow)
	}
	want := []Attr{
		{Key: "version", Value: "1.4.1"},
		{Key: "os", Value: runtime.GOOS},
		{Key: "arch", Value: runtime.GOARCH},
	}
	if !reflect.DeepEqual(got.Attrs, want) {
		t.Errorf("DaemonStarted(...).Attrs = %v, want exactly %v", got.Attrs, want)
	}
}

func TestDaemonStartedRejectsAVersionThatCouldCarryAnIdentifier(t *testing.T) {
	got, err := DaemonStarted("/var/lib/snapback", runtimeNow)
	if err == nil {
		t.Fatalf("DaemonStarted(%q, %v) = %v, nil; want an error", "/var/lib/snapback", runtimeNow, got)
	}
}

func TestMountReadyCarriesVersionOSArchAndDurationBucket(t *testing.T) {
	const d = 2 * time.Second
	got, err := MountReady("1.4.1", d, runtimeNow)
	if err != nil {
		t.Fatalf("MountReady(%q, %v, %v) error = %v, want nil", "1.4.1", d, runtimeNow, err)
	}
	if got.Name != "mount.ready" {
		t.Errorf("MountReady(...).Name = %q, want %q", got.Name, "mount.ready")
	}
	if !got.Time.Equal(runtimeNow) {
		t.Errorf("MountReady(...).Time = %v, want %v", got.Time, runtimeNow)
	}
	want := []Attr{
		{Key: "version", Value: "1.4.1"},
		{Key: "os", Value: runtime.GOOS},
		{Key: "arch", Value: runtime.GOARCH},
		{Key: "duration", Value: Bucket(d)},
	}
	if !reflect.DeepEqual(got.Attrs, want) {
		t.Errorf("MountReady(...).Attrs = %v, want exactly %v", got.Attrs, want)
	}
}

func TestMountReadyReportsTheDurationOnlyAsABucket(t *testing.T) {
	for _, d := range []time.Duration{0, 50 * time.Millisecond, 500 * time.Millisecond, 5 * time.Second, 30 * time.Second, 10 * time.Minute} {
		got, err := MountReady("1.4.1", d, runtimeNow)
		if err != nil {
			t.Fatalf("MountReady(%q, %v, %v) error = %v, want nil", "1.4.1", d, runtimeNow, err)
		}
		if len(got.Attrs) != 4 {
			t.Fatalf("MountReady(_, %v, _).Attrs = %v, want exactly 4 attributes", d, got.Attrs)
		}
		last := got.Attrs[len(got.Attrs)-1]
		if last.Key != "duration" || last.Value != Bucket(d) {
			t.Errorf("MountReady(_, %v, _) last attribute = %v, want {duration %s}", d, last, Bucket(d))
		}
	}
}

// TestRuntimeEventSignaturesTakeNoIdentifier pins the parameter types of both
// constructors. Only one string may cross the boundary, and it is the version:
// a second string parameter could hold a repository id, a root path or a mount
// point, so the schema forbids one.
func TestRuntimeEventSignaturesTakeNoIdentifier(t *testing.T) {
	tests := []struct {
		name  string
		fn    any
		types []reflect.Type
	}{
		{
			name:  "DaemonStarted",
			fn:    DaemonStarted,
			types: []reflect.Type{reflect.TypeOf(""), reflect.TypeOf(time.Time{})},
		},
		{
			name:  "MountReady",
			fn:    MountReady,
			types: []reflect.Type{reflect.TypeOf(""), reflect.TypeOf(time.Duration(0)), reflect.TypeOf(time.Time{})},
		},
	}
	for _, tt := range tests {
		ft := reflect.TypeOf(tt.fn)
		if ft.NumIn() != len(tt.types) {
			t.Errorf("%s takes %d parameters, want exactly %d", tt.name, ft.NumIn(), len(tt.types))
			continue
		}
		for i, want := range tt.types {
			if got := ft.In(i); got != want {
				t.Errorf("%s parameter %d is %v, want %v", tt.name, i, got, want)
			}
		}
		stringParams := 0
		for i := range ft.NumIn() {
			if ft.In(i).Kind() == reflect.String {
				stringParams++
			}
		}
		if stringParams != 1 {
			t.Errorf("%s takes %d string parameters, want exactly 1 (the version)", tt.name, stringParams)
		}
	}
}
