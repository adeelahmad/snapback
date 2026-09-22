package telemetry

import (
	"runtime"
	"time"
)

// DaemonStarted returns the daemon.started event for version at now.
func DaemonStarted(version string, now time.Time) (Event, error) {
	attrs, err := runtimeAttrs(version)
	if err != nil {
		return Event{}, err
	}
	return Event{Name: "daemon.started", Attrs: attrs, Time: now}, nil
}

// MountReady returns the mount.ready event for version at now, reporting d as a
// coarse duration bucket.
func MountReady(version string, d time.Duration, now time.Time) (Event, error) {
	attrs, err := runtimeAttrs(version)
	if err != nil {
		return Event{}, err
	}
	duration, err := NewAttr("duration", Bucket(d))
	if err != nil {
		return Event{}, err
	}
	return Event{Name: "mount.ready", Attrs: append(attrs, duration), Time: now}, nil
}

// runtimeAttrs returns the version, os and arch attributes every runtime event
// carries, in schema order.
func runtimeAttrs(version string) ([]Attr, error) {
	pairs := []struct{ key, value string }{
		{"version", version},
		{"os", runtime.GOOS},
		{"arch", runtime.GOARCH},
	}
	attrs := make([]Attr, 0, len(pairs)+1)
	for _, p := range pairs {
		attr, err := NewAttr(p.key, p.value)
		if err != nil {
			return nil, err
		}
		attrs = append(attrs, attr)
	}
	return attrs, nil
}
