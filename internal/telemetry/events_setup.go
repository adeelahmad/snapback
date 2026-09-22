package telemetry

import "time"

// SetupCompleted returns the setup.completed event for one finished setup run.
func SetupCompleted(version, outcome string, d time.Duration, now time.Time) (Event, error) {
	return Event{}, nil
}
