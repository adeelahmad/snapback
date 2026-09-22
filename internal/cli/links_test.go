package cli

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/links"
	"github.com/adeelahmad/snapback/internal/rawpath"
)

func TestLinksListJSON(t *testing.T) {
	l := &fakeLinker{records: []links.Record{
		{Key: "k1", RootID: "home", Rel: rawpath.Path("a"), Dir: rawpath.Path("/w/a"), Target: "t1"},
		{Key: "k2", RootID: "home", Rel: rawpath.Path("\xff"), Dir: rawpath.Path("/w/\xff"), Target: "t2"},
	}}
	var execs int
	env, out, _ := newEnv(nil)

	got := LinksCommand(linkDeps(l, "/w", &execs)).Run(context.Background(), env, []string{"list", "--json"})

	if got != 0 {
		t.Fatalf("links list --json = %d, want 0", got)
	}
	e := decodeEnvelope(t, out.Bytes())
	var data []map[string]json.RawMessage
	if err := json.Unmarshal(e.Data, &data); err != nil {
		t.Fatalf("decode data %q: %v", e.Data, err)
	}
	if len(data) != 2 {
		t.Fatalf("len(data) = %d, want 2", len(data))
	}
	if got, want := string(data[0]["dir"]), `"/w/a"`; got != want {
		t.Errorf("data[0].dir = %s, want %s", got, want)
	}
	if got, want := string(data[1]["dir"]), `{"b64":"L3cv/w=="}`; got != want {
		t.Errorf("data[1].dir = %s, want %s", got, want)
	}
}

func TestLinksRepairAndRemoveManaged(t *testing.T) {
	report := links.RepairReport{
		Removed:   []links.RepairEntry{{Key: "k1", Dir: rawpath.Path("/w/gone")}},
		Preserved: []links.RepairEntry{{Key: "k2", Dir: rawpath.Path("/w/kept"), Code: errcode.LinkConflict}},
	}
	tests := []struct {
		args       []string
		wantRepair int
		wantRemove int
	}{
		{args: []string{"repair"}, wantRepair: 1},
		{args: []string{"remove", "--managed"}, wantRemove: 1},
	}
	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			l := &fakeLinker{report: report}
			var execs int
			env, out, _ := newEnv(nil)

			got := LinksCommand(linkDeps(l, "/w", &execs)).Run(context.Background(), env, tt.args)

			if got != 0 {
				t.Fatalf("links %q = %d, want 0", tt.args, got)
			}
			if l.repaired != tt.wantRepair || l.removedAll != tt.wantRemove {
				t.Errorf("Repair, RemoveManaged calls = %d, %d, want %d, %d", l.repaired, l.removedAll, tt.wantRepair, tt.wantRemove)
			}
			if len(l.ensured) != 0 || l.listed != 0 {
				t.Errorf("Ensure, List calls = %d, %d, want 0, 0", len(l.ensured), l.listed)
			}
			if !strings.Contains(out.String(), "/w/kept") {
				t.Errorf("links %q stdout = %q, want it to name the preserved entry /w/kept", tt.args, out.String())
			}
		})
	}
}

func TestLinksRemoveRequiresManaged(t *testing.T) {
	tests := [][]string{
		{"remove"},
		{"remove", "/w/d"},
		{"bogus"},
		{},
	}
	for _, args := range tests {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			l := &fakeLinker{}
			var execs int
			env, _, _ := newEnv(nil)

			got := LinksCommand(linkDeps(l, "/w", &execs)).Run(context.Background(), env, args)

			if got != 2 {
				t.Errorf("links %q = %d, want 2", args, got)
			}
			if n := l.calls(); n != 0 {
				t.Errorf("engine calls = %d, want 0", n)
			}
		})
	}
}
