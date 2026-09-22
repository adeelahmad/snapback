// Package telemetry builds the opt-in usage events Snapback may send.
package telemetry

import "time"

// Duration labels, shortest first. Decision D5: a duration is only ever
// reported as one of these buckets, never as a number, because exact timings
// can fingerprint a repository.
const (
	bucketUnder100ms = "<100ms"
	bucketUnder1s    = "<1s"
	bucketUnder10s   = "<10s"
	bucketUnder60s   = "<60s"
	bucketAtLeast60s = ">=60s"
)

// Bucket maps a duration onto one of the coarse labels returned by [Buckets].
// Exact timings can fingerprint a repository, so durations never leave the
// process as numbers. Bands are half-open, upper bound exclusive; durations
// below 100ms, including zero and every negative duration, fall in the first.
func Bucket(d time.Duration) string {
	switch {
	case d < 100*time.Millisecond:
		return bucketUnder100ms
	case d < time.Second:
		return bucketUnder1s
	case d < 10*time.Second:
		return bucketUnder10s
	case d < time.Minute:
		return bucketUnder60s
	default:
		return bucketAtLeast60s
	}
}

// Buckets returns the closed set of duration labels, shortest first. Each call
// returns a fresh slice, so a caller cannot reorder or rewrite the set.
func Buckets() []string {
	return []string{
		bucketUnder100ms,
		bucketUnder1s,
		bucketUnder10s,
		bucketUnder60s,
		bucketAtLeast60s,
	}
}
