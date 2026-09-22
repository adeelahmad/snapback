package telemetry

import (
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

// setupNow is a fixed instant, so a test never depends on the wall clock.
var setupNow = time.Date(2026, time.March, 4, 9, 30, 0, 0, time.UTC)

func TestSetupCompletedName(t *testing.T) {
	got, err := SetupCompleted("1.4.1", "ok", 250*time.Millisecond, setupNow)
	if err != nil {
		t.Fatalf("SetupCompleted() error = %v, want nil", err)
	}
	if want := "setup.completed"; got.Name != want {
		t.Errorf("Name = %q, want %q", got.Name, want)
	}
	if !IsName(got.Name) {
		t.Errorf("IsName(%q) = false, want true", got.Name)
	}
}

func TestSetupCompletedTimeIsNow(t *testing.T) {
	got, err := SetupCompleted("1.4.1", "ok", time.Second, setupNow)
	if err != nil {
		t.Fatalf("SetupCompleted() error = %v, want nil", err)
	}
	if !got.Time.Equal(setupNow) {
		t.Errorf("Time = %v, want %v", got.Time, setupNow)
	}
}

func TestSetupCompletedAttrsAreTheSchemaInOrder(t *testing.T) {
	got, err := SetupCompleted("1.4.1", "failed", 5*time.Second, setupNow)
	if err != nil {
		t.Fatalf("SetupCompleted() error = %v, want nil", err)
	}
	want := []Attr{
		{Key: "version", Value: "1.4.1"},
		{Key: "os", Value: runtime.GOOS},
		{Key: "arch", Value: runtime.GOARCH},
		{Key: "outcome", Value: "failed"},
		{Key: "duration", Value: Bucket(5 * time.Second)},
	}
	if !reflect.DeepEqual(got.Attrs, want) {
		t.Fatalf("Attrs = %+v, want %+v", got.Attrs, want)
	}
}

func TestSetupCompletedAcceptsEveryOutcome(t *testing.T) {
	for _, outcome := range []string{"ok", "failed", "abandoned"} {
		got, err := SetupCompleted("1.4.1", outcome, time.Minute, setupNow)
		if err != nil {
			t.Errorf("SetupCompleted(outcome=%q) error = %v, want nil", outcome, err)
			continue
		}
		if len(got.Attrs) != 5 {
			t.Errorf("SetupCompleted(outcome=%q) has %d attrs, want 5", outcome, len(got.Attrs))
			continue
		}
		if got.Attrs[3] != (Attr{Key: "outcome", Value: outcome}) {
			t.Errorf("outcome attr = %+v, want {outcome %s}", got.Attrs[3], outcome)
		}
	}
}

func TestSetupCompletedRejectsUnknownOutcome(t *testing.T) {
	for _, outcome := range []string{"", "OK", "success", "cancelled"} {
		got, err := SetupCompleted("1.4.1", outcome, time.Second, setupNow)
		if err == nil {
			t.Errorf("SetupCompleted(outcome=%q) = %+v, want an error", outcome, got)
			continue
		}
		msg := err.Error()
		for _, want := range []string{"ok", "failed", "abandoned"} {
			if !strings.Contains(msg, want) {
				t.Errorf("SetupCompleted(outcome=%q) error %q does not name %q", outcome, msg, want)
			}
		}
	}
}

func TestSetupCompletedDurationIsABucketLabel(t *testing.T) {
	for _, d := range []time.Duration{0, 50 * time.Millisecond, 500 * time.Millisecond, 5 * time.Second, 30 * time.Second, 2 * time.Minute} {
		got, err := SetupCompleted("1.4.1", "ok", d, setupNow)
		if err != nil {
			t.Errorf("SetupCompleted(d=%v) error = %v, want nil", d, err)
			continue
		}
		if len(got.Attrs) != 5 {
			t.Errorf("SetupCompleted(d=%v) has %d attrs, want 5", d, len(got.Attrs))
			continue
		}
		if want := (Attr{Key: "duration", Value: Bucket(d)}); got.Attrs[4] != want {
			t.Errorf("duration attr for %v = %+v, want %+v", d, got.Attrs[4], want)
		}
	}
}

func TestSetupCompletedCarriesNoNumericValue(t *testing.T) {
	got, err := SetupCompleted("1.4.1", "ok", 1234*time.Millisecond, setupNow)
	if err != nil {
		t.Fatalf("SetupCompleted() error = %v, want nil", err)
	}
	if len(got.Attrs) == 0 {
		t.Fatalf("Attrs = %+v, want the five schema attrs", got.Attrs)
	}
	for _, attr := range got.Attrs {
		if _, err := strconv.ParseFloat(attr.Value, 64); err == nil {
			t.Errorf("attr %q value %q parses as a number, want a label", attr.Key, attr.Value)
		}
	}
}
