package daemon

import (
	"encoding/json"
	"testing"

	"github.com/adeelahmad/snapback/internal/ipc"
	"github.com/adeelahmad/snapback/internal/links"
	"github.com/adeelahmad/snapback/internal/rawpath"
)

func TestStatusReportsLinksAndDiscovery(t *testing.T) {
	h := newHarness(t)
	h.cfg.Discovery.Mode = "seed"
	h.linker.records = []links.Record{
		{Key: "k1", Dir: rawpath.Path("/r/a"), State: links.StateOwned},
		{Key: "k2", Dir: rawpath.Path("/r/b"), State: links.StateOwned},
		{Key: "k3", Dir: rawpath.Path("/r/c"), State: links.StatePending},
	}
	d := readyDaemon(t, h)

	s := d.Status()
	if s.Links != 2 || s.Discovery != "seed" {
		t.Errorf("Status() Links, Discovery = %d, %q, want 2, %q", s.Links, s.Discovery, "seed")
	}

	resp := roundTrip(t, h, ipc.Request{Op: ipc.OpStatus})
	if !resp.OK {
		t.Fatalf("status = %+v, want OK", resp)
	}
	var got struct {
		Links     int
		Discovery string
	}
	if err := json.Unmarshal(resp.Data, &got); err != nil {
		t.Fatalf("status Data %q: unmarshal: %v", resp.Data, err)
	}
	if got.Links != 2 || got.Discovery != "seed" {
		t.Errorf("status op Links, Discovery = %d, %q, want 2, %q", got.Links, got.Discovery, "seed")
	}
}
