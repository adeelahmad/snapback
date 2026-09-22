package daemon

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/ipc"
	"github.com/adeelahmad/snapback/internal/provider"
	"github.com/adeelahmad/snapback/internal/status"
)

func TestStatusReportsPrewarmSummary(t *testing.T) {
	h := newHarness(t)
	h.ref.result.Pending = []provider.SnapshotID{"p1"}
	h.deps.Prewarmer = &fakePrewarmer{rec: h.rec, results: []provider.PrewarmResult{
		{ID: "a", Warm: true},
		{ID: "b", Warm: true},
		{ID: "c", Err: errors.New("timeout")},
	}}
	d := New(h.cfg, h.deps)
	start(t, d)
	waitState(t, d, 2*time.Second, func(s string) bool { return s == "ready" })

	got := d.Status().Prewarm
	want := status.PrewarmSummary{Warm: 2, Cold: 1, Pending: 1, LastPrewarm: fixedNow}
	if got != want {
		t.Errorf("Status().Prewarm = %+v, want %+v", got, want)
	}
}

func TestStatusOpJSONHasPrewarm(t *testing.T) {
	h := newHarness(t)
	h.deps.Prewarmer = &fakePrewarmer{rec: h.rec, results: []provider.PrewarmResult{{ID: "a", Warm: true}}}
	d := New(h.cfg, h.deps)
	start(t, d)
	waitState(t, d, 2*time.Second, func(s string) bool { return s == "ready" })

	resp := call(d, ipc.Request{Op: ipc.OpStatus})
	var got map[string]json.RawMessage
	if err := json.Unmarshal(resp.Data, &got); err != nil {
		t.Fatalf("json.Unmarshal(status Data %s) = %v", resp.Data, err)
	}
	raw, ok := got["prewarm"]
	if !ok {
		t.Fatalf("status op JSON = %s, want key %q", resp.Data, "prewarm")
	}
	var sum struct {
		Warm int `json:"warm"`
	}
	if err := json.Unmarshal(raw, &sum); err != nil || sum.Warm != 1 {
		t.Errorf("status op prewarm = %s (err %v), want warm=1", raw, err)
	}
}
