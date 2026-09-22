package telemetry

import "time"

// DaemonStarted returns the daemon.started event for version at now.
func DaemonStarted(version string, now time.Time) (Event, error) { return Event{}, nil }

// MountReady returns the mount.ready event for version at now, reporting d as a
// coarse duration bucket.
func MountReady(version string, d time.Duration, now time.Time) (Event, error) {
	return Event{}, nil
}
