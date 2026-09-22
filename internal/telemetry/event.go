// Package telemetry defines the closed, opt-in telemetry event schema.
//
// The schema is exhaustive by design: the five names in [Names] are the only
// events Snapback can ever emit, and every attribute they carry is an enum, a
// counter or a coarse duration bucket. Nothing in this package can represent a
// path, a repository URI, a hostname or a username.
package telemetry

import "time"

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
var eventNames = []string{}

// Names returns the closed set of telemetry event names, in schema order.
func Names() []string { return nil }

// IsName reports whether s is one of the closed set of event names.
func IsName(s string) bool { return false }
