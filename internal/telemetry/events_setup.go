package telemetry

import (
	"fmt"
	"runtime"
	"time"
)

// setupOutcomes is the closed set of setup.completed outcome values.
var setupOutcomes = [...]string{"ok", "failed", "abandoned"}

// SetupCompleted returns the setup.completed event for one finished setup run.
// The duration is recorded as a [Bucket] label, never as a number, and outcome
// must be one of ok, failed or abandoned.
func SetupCompleted(version, outcome string, d time.Duration, now time.Time) (Event, error) {
	if !isSetupOutcome(outcome) {
		return Event{}, fmt.Errorf("telemetry: setup outcome %q is not one of ok, failed, abandoned", outcome)
	}
	pairs := [...][2]string{
		{"version", version},
		{"os", runtime.GOOS},
		{"arch", runtime.GOARCH},
		{"outcome", outcome},
		{"duration", Bucket(d)},
	}
	attrs := make([]Attr, 0, len(pairs))
	for _, p := range pairs {
		a, err := NewAttr(p[0], p[1])
		if err != nil {
			return Event{}, err
		}
		attrs = append(attrs, a)
	}
	return Event{Name: "setup.completed", Attrs: attrs, Time: now}, nil
}

// isSetupOutcome reports whether outcome is in the closed outcome set.
func isSetupOutcome(outcome string) bool {
	for _, o := range setupOutcomes {
		if outcome == o {
			return true
		}
	}
	return false
}
