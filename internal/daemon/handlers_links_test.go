package daemon

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/ipc"
	"github.com/adeelahmad/snapback/internal/links"
	"github.com/adeelahmad/snapback/internal/rawpath"
)

// roundTrip sends req to the running daemon over its real ipc socket.
func roundTrip(t *testing.T, h *harness, req ipc.Request) ipc.Response {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	sock := h.deps.Listener.Addr().String()
	c, err := ipc.Dial(ctx, sock)
	if err != nil {
		t.Fatalf("ipc.Dial(%q) = %v, want nil", sock, err)
	}
	defer func() { _ = c.Close() }()
	req.V = 1
	resp, err := c.Call(ctx, req)
	if err != nil {
		t.Fatalf("Call(%s) = %v, want nil", req.Op, err)
	}
	return resp
}

func sampleReport(key string) links.RepairReport {
	return links.RepairReport{
		Completed: []links.RepairEntry{{Key: key + "-c", Dir: rawpath.Path("/r/c")}},
		Removed:   []links.RepairEntry{{Key: key + "-r", Dir: rawpath.Path("/r/r")}},
		Preserved: []links.RepairEntry{{Key: key + "-p", Dir: rawpath.Path("/r/p"), Code: errcode.LinkConflict}},
	}
}

func TestLinksListReturnsLinkerRecords(t *testing.T) {
	h := newHarness(t)
	want := []links.Record{
		{Key: "k1", RootID: "home", Rel: rawpath.Path("a"), Dir: rawpath.Path("/r/a"), Target: "/t/a", State: links.StateOwned},
		{Key: "k2", RootID: "home", Rel: rawpath.Path("b"), Dir: rawpath.Path("/r/b"), Target: "/t/b", State: links.StatePending},
	}
	h.linker.records = want
	readyDaemon(t, h)

	resp := roundTrip(t, h, ipc.Request{Op: ipc.OpLinksList})
	if !resp.OK {
		t.Fatalf("links_list = %+v, want OK", resp)
	}
	var got []links.Record
	if err := json.Unmarshal(resp.Data, &got); err != nil {
		t.Fatalf("links_list Data %q: unmarshal: %v", resp.Data, err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("links_list records = %+v, want %+v", got, want)
	}
}

func TestLinksRepairReturnsRepairReport(t *testing.T) {
	h := newHarness(t)
	want := sampleReport("repair")
	h.linker.repair = want
	readyDaemon(t, h)

	resp := roundTrip(t, h, ipc.Request{Op: ipc.OpLinksRepair})
	if !resp.OK {
		t.Fatalf("links_repair = %+v, want OK", resp)
	}
	var got links.RepairReport
	if err := json.Unmarshal(resp.Data, &got); err != nil {
		t.Fatalf("links_repair Data %q: unmarshal: %v", resp.Data, err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("links_repair report = %+v, want %+v", got, want)
	}
}

func TestLinksRemoveManagedReturnsReport(t *testing.T) {
	h := newHarness(t)
	want := sampleReport("remove")
	h.linker.removed = want
	h.linker.repair = sampleReport("wrong")
	readyDaemon(t, h)

	resp := roundTrip(t, h, ipc.Request{Op: ipc.OpLinksRemoveManaged})
	if !resp.OK {
		t.Fatalf("links_remove_managed = %+v, want OK", resp)
	}
	var got links.RepairReport
	if err := json.Unmarshal(resp.Data, &got); err != nil {
		t.Fatalf("links_remove_managed Data %q: unmarshal: %v", resp.Data, err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("links_remove_managed report = %+v, want %+v", got, want)
	}
}

func TestLinksOpErrorMapsToCode(t *testing.T) {
	tests := []struct {
		op   string
		set  func(f *fakeLinker, err error)
		code errcode.Code
	}{
		{ipc.OpLinksList, func(f *fakeLinker, err error) { f.listErr = err }, errcode.PermissionDenied},
		{ipc.OpLinksRepair, func(f *fakeLinker, err error) { f.repairErr = err }, errcode.LinkConflict},
		{ipc.OpLinksRemoveManaged, func(f *fakeLinker, err error) { f.removeErr = err }, errcode.MountFailure},
	}
	for _, tt := range tests {
		t.Run(tt.op, func(t *testing.T) {
			h := newHarness(t)
			tt.set(h.linker, errcode.New(tt.code, "links."+tt.op, errors.New("boom")))
			readyDaemon(t, h)

			resp := roundTrip(t, h, ipc.Request{Op: tt.op})
			if resp.OK || resp.Code != tt.code {
				t.Errorf("%s = {OK:%v Code:%q}, want {OK:false Code:%q}", tt.op, resp.OK, resp.Code, tt.code)
			}
		})
	}
}

func TestLinksUnknownOpStillRejected(t *testing.T) {
	h := newHarness(t)
	readyDaemon(t, h)

	resp := roundTrip(t, h, ipc.Request{Op: "links_explode"})
	if resp.OK || resp.Code != errcode.InvalidConfig {
		t.Errorf("links_explode = {OK:%v Code:%q}, want {OK:false Code:%q}", resp.OK, resp.Code, errcode.InvalidConfig)
	}
}
