// Package telemetry builds the opt-in usage events Snapback may send.
package telemetry

import "time"

// Bucket maps a duration onto one of the coarse labels returned by [Buckets].
// Exact timings can fingerprint a repository, so durations never leave the
// process as numbers.
func Bucket(d time.Duration) string {
	_ = d
	return ""
}

// Buckets returns the closed set of duration labels, shortest first.
func Buckets() []string {
	return nil
}
