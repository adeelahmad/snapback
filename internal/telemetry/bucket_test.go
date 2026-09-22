package telemetry_test

import (
	"math"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/telemetry"
)

// bucketCases pins every boundary of the five-label scale from decision D5.
var bucketCases = []struct {
	name string
	in   time.Duration
	want string
}{
	{"negative hour", -time.Hour, "<100ms"},
	{"negative second", -time.Second, "<100ms"},
	{"negative nanosecond", -time.Nanosecond, "<100ms"},
	{"zero", 0, "<100ms"},
	{"one nanosecond", time.Nanosecond, "<100ms"},
	{"one millisecond", time.Millisecond, "<100ms"},
	{"fifty milliseconds", 50 * time.Millisecond, "<100ms"},
	{"ninety-nine milliseconds", 99 * time.Millisecond, "<100ms"},
	{"just under a hundred milliseconds", 100*time.Millisecond - time.Nanosecond, "<100ms"},

	{"a hundred milliseconds", 100 * time.Millisecond, "<1s"},
	{"just over a hundred milliseconds", 100*time.Millisecond + time.Nanosecond, "<1s"},
	{"two hundred fifty milliseconds", 250 * time.Millisecond, "<1s"},
	{"half a second", 500 * time.Millisecond, "<1s"},
	{"nine hundred ninety-nine milliseconds", 999 * time.Millisecond, "<1s"},
	{"just under a second", time.Second - time.Nanosecond, "<1s"},

	{"one second", time.Second, "<10s"},
	{"just over a second", time.Second + time.Nanosecond, "<10s"},
	{"one and a half seconds", 1500 * time.Millisecond, "<10s"},
	{"two seconds", 2 * time.Second, "<10s"},
	{"five seconds", 5 * time.Second, "<10s"},
	{"nine seconds", 9 * time.Second, "<10s"},
	{"nine thousand nine hundred ninety-nine milliseconds", 9999 * time.Millisecond, "<10s"},
	{"just under ten seconds", 10*time.Second - time.Nanosecond, "<10s"},

	{"ten seconds", 10 * time.Second, "<60s"},
	{"just over ten seconds", 10*time.Second + time.Nanosecond, "<60s"},
	{"fifteen seconds", 15 * time.Second, "<60s"},
	{"thirty seconds", 30 * time.Second, "<60s"},
	{"forty-five seconds", 45 * time.Second, "<60s"},
	{"fifty-nine seconds", 59 * time.Second, "<60s"},
	{"fifty-nine thousand nine hundred ninety-nine milliseconds", 59999 * time.Millisecond, "<60s"},
	{"just under sixty seconds", time.Minute - time.Nanosecond, "<60s"},

	{"sixty seconds", 60 * time.Second, ">=60s"},
	{"one minute", time.Minute, ">=60s"},
	{"just over a minute", time.Minute + time.Nanosecond, ">=60s"},
	{"ninety seconds", 90 * time.Second, ">=60s"},
	{"two minutes", 2 * time.Minute, ">=60s"},
	{"five minutes", 5 * time.Minute, ">=60s"},
	{"one hour", time.Hour, ">=60s"},
	{"one day", 24 * time.Hour, ">=60s"},
	{"maximum duration", time.Duration(math.MaxInt64), ">=60s"},
}

func TestBucketBoundaries(t *testing.T) {
	t.Parallel()

	for _, tc := range bucketCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := telemetry.Bucket(tc.in); got != tc.want {
				t.Errorf("Bucket(%v) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestBucketsReturnsTheClosedLabelSetInOrder(t *testing.T) {
	t.Parallel()

	want := []string{"<100ms", "<1s", "<10s", "<60s", ">=60s"}

	got := telemetry.Buckets()
	if len(got) != len(want) {
		t.Fatalf("Buckets() = %q, want %q", got, want)
	}

	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Buckets()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestBucketOnlyEverReturnsALabelFromBuckets(t *testing.T) {
	t.Parallel()

	known := make(map[string]bool, len(telemetry.Buckets()))
	for _, label := range telemetry.Buckets() {
		known[label] = true
	}

	for _, tc := range bucketCases {
		if got := telemetry.Bucket(tc.in); !known[got] {
			t.Errorf("Bucket(%v) = %q, which is not in Buckets() = %q", tc.in, got, telemetry.Buckets())
		}
	}
}
