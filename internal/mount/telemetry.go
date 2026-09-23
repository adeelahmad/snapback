package mount

import (
	"context"
	"sync"
	"time"

	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/telemetry"
)

// TelemetryAdapter wraps an Adapter and reports the process's first
// successful mount as mount.ready, and every failed mount as an error event,
// to Client. Client may be nil, in which case Mount behaves exactly like the
// wrapped Adapter and reports nothing.
type TelemetryAdapter struct {
	Adapter
	// Client delivers the events. A nil Client disables reporting.
	Client *telemetry.Client
	// Version is the release string every event carries.
	Version string
	// Now reports the current time. A nil Now uses time.Now.
	Now func() time.Time

	once sync.Once
}

// Mount times the wrapped Adapter's mount, starting the clock at the call
// (mount start, not process start). On the process's first success it
// reports mount.ready with the elapsed duration as a bucket; a later success
// reports nothing more. On failure it reports an error event with
// errcode.MountFailure and never mount.ready.
func (a *TelemetryAdapter) Mount(dir string, cat Catalog) error {
	start := a.now()
	err := a.Adapter.Mount(dir, cat)
	if err != nil {
		a.emit(mountFailureEvent(a.Version, a.now()))
		return err
	}
	a.once.Do(func() {
		a.emit(mountReadyEvent(a.Version, a.now().Sub(start), a.now()))
	})
	return nil
}

func (a *TelemetryAdapter) now() time.Time {
	if a.Now != nil {
		return a.Now()
	}
	return time.Now()
}

func (a *TelemetryAdapter) emit(ev telemetry.Event, err error) {
	if a.Client == nil || err != nil {
		return
	}
	a.Client.Emit(context.Background(), ev)
}

// mountReadyEvent builds the mount.ready event for this call site. It takes
// only a version, a duration and a timestamp, so nothing here can pass a
// mount point, a repo id or a root path through to the event.
func mountReadyEvent(version string, d time.Duration, now time.Time) (telemetry.Event, error) {
	return telemetry.MountReady(version, d, now)
}

// mountFailureEvent builds the error event for a failed mount at this call
// site. It takes only a version and a timestamp, so nothing here can pass a
// mount point, a repo id or a root path through to the event.
func mountFailureEvent(version string, now time.Time) (telemetry.Event, error) {
	return telemetry.ErrorEvent(version, errcode.MountFailure, now)
}
