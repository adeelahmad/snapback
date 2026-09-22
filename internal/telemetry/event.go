// Package telemetry defines the closed, opt-in telemetry event schema.
//
// The schema is exhaustive by design: the five names in [Names] are the only
// events Snapback can ever emit, and every attribute they carry is an enum, a
// counter or a coarse duration bucket. Nothing in this package can represent a
// path, a repository URI, a hostname or a username.
//
// The set is closed: a sixth name, or a different order, cannot be added here
// alone. The schema is spelled out again in the package's tests, so every
// addition or reordering has to change that test too, which is the review gate
// on what Snapback is allowed to report.
package telemetry

import (
	"slices"
	"time"
)

// Event is one counted occurrence of a schema name with its attributes.
type Event struct {
	// Name is one of the closed set reported by [Names].
	Name string
	// Attrs are the event's attributes, all from the closed key set.
	Attrs []Attr
	// Time is when the event occurred.
	Time time.Time
}

// eventNames is the closed, ordered set of telemetry event names.
var eventNames = []string{
	"setup.completed",
	"daemon.started",
	"mount.ready",
	"doctor.failed",
	"error",
}

// Names returns the closed set of telemetry event names, in schema order.
// The result is a fresh copy, so callers cannot reorder or extend the schema.
func Names() []string { return slices.Clone(eventNames) }

// IsName reports whether s is one of the closed set of event names. The match
// is exact and case-sensitive.
func IsName(s string) bool { return slices.Contains(eventNames, s) }
