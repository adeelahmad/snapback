package daemon

import (
	"slices"
	"testing"

	"github.com/adeelahmad/snapback/internal/ipc"
	"github.com/adeelahmad/snapback/internal/readerpolicy"
)

// TestStatusReportsThrottleEvents pins that the status snapshot carries the
// reader policy's events, in order, over both the direct call and the IPC
// status op: Acc 12 saw "throttle": null while the policy was denying reads.
func TestStatusReportsThrottleEvents(t *testing.T) {
	want := []readerpolicy.ThrottleEvent{
		{PID: 42, Process: "find", Rule: "deny", At: fixedNow},
		{PID: 43, Process: "mds", Rule: "burst", At: fixedNow},
	}
	h := newHarness(t)
	h.deps.Throttle = func() []readerpolicy.ThrottleEvent { return want }
	d := readyDaemon(t, h)

	if got := d.Status().Throttle; !slices.Equal(got, want) {
		t.Errorf("Status().Throttle = %+v, want %+v", got, want)
	}
	got := decodeSnapshot(t, call(d, ipc.Request{Op: ipc.OpStatus})).Throttle
	if !slices.Equal(got, want) {
		t.Errorf("status op Throttle = %+v, want %+v", got, want)
	}
}

// TestStatusThrottleEmptyWithoutPolicy pins that a daemon without a reader
// policy reports no events instead of panicking.
func TestStatusThrottleEmptyWithoutPolicy(t *testing.T) {
	h := newHarness(t)
	h.deps.Throttle = nil
	d := readyDaemon(t, h)

	if got := d.Status().Throttle; len(got) != 0 {
		t.Errorf("Status().Throttle = %+v, want empty", got)
	}
}
