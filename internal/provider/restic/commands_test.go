package restic

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/provider"
)

const lockStderr = "unable to create lock in backend: repository is already locked by PID 1234 on host by user\n"

var errExit = errors.New("exit status 1")

// newCmdProvider returns a provider with default options driven by fr.
func newCmdProvider(t *testing.T, fr *fakeRunner) *Provider {
	t.Helper()
	opts := validOptions()
	opts.Runner = fr
	return mustNew(t, opts)
}

// readTestdata returns the contents of testdata/name.
func readTestdata(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("os.ReadFile(%q) error = %v, want nil", name, err)
	}
	return b
}

// onlyRunCall returns the single recorded Run call, failing otherwise.
func onlyRunCall(t *testing.T, fr *fakeRunner) fakeCall {
	t.Helper()
	calls := fr.runCalls()
	if len(calls) != 1 {
		t.Fatalf("recorded %d Run calls, want 1", len(calls))
	}
	return calls[0]
}

func TestValidateParsesIdentity(t *testing.T) {
	fr := &fakeRunner{results: map[string]fakeResult{"cat": {stdout: readTestdata(t, "cat_config.json")}}}
	p := newCmdProvider(t, fr)

	got, err := p.Validate(context.Background())
	if err != nil {
		t.Fatalf("Validate() error = %v, want nil", err)
	}
	want := provider.Identity{RepoID: "5e1f0c3a9b8d7e6f5a4b3c2d1e0f9a8b7c6d5e4f3a2b1c0d9e8f7a6b5c4d3e2f", Version: 2}
	if got != want {
		t.Errorf("Validate() = %+v, want %+v", got, want)
	}
	c := onlyRunCall(t, fr)
	if c.name != bin {
		t.Errorf("Validate() ran %q, want %q", c.name, bin)
	}
	if want := p.validateArgs(); !slices.Equal(c.args, want) {
		t.Errorf("Validate() argv = %q, want %q", c.args, want)
	}
	if !slices.Contains(c.env, "RESTIC_REPOSITORY="+repo) {
		t.Errorf("Validate() env = %q, want it to contain RESTIC_REPOSITORY=<repo>", c.env)
	}
	if !c.hasDeadline {
		t.Error("Validate() ctx has no deadline, want metadataTimeout applied")
	}
}

func TestValidateRejectsShortRepoID(t *testing.T) {
	fr := &fakeRunner{results: map[string]fakeResult{"cat": {stdout: []byte(`{"version":2,"id":"250dc51f"}`)}}}
	p := newCmdProvider(t, fr)

	got, err := p.Validate(context.Background())
	if err == nil {
		t.Fatalf("Validate() = %+v, nil, want an error for a short repository id", got)
	}
	if got != (provider.Identity{}) {
		t.Errorf("Validate() identity = %+v, want zero on error", got)
	}
}

func TestListParsesTwoHostsAndRelativePaths(t *testing.T) {
	fr := &fakeRunner{results: map[string]fakeResult{"snapshots": {stdout: readTestdata(t, "snapshots_two_hosts.json")}}}
	p := newCmdProvider(t, fr)

	got, err := p.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	want := []provider.Snapshot{
		{
			ID:       "1111111111111111111111111111111111111111111111111111111111111111",
			Time:     time.Date(2026, 9, 20, 9, 15, 2, 123456789, time.FixedZone("", 10*3600)),
			Hostname: "laptop",
			Tags:     []string{"daily"},
			Paths:    []string{"/home/alex/project"},
		},
		{
			ID:       "2222222222222222222222222222222222222222222222222222222222222222",
			Time:     time.Date(2026, 9, 21, 22, 40, 0, 0, time.FixedZone("", -4*3600)),
			Hostname: "nas",
			Tags:     []string{"weekly", "offsite"},
			Paths:    []string{"home/alex/project"},
		},
		{
			ID:       "3333333333333333333333333333333333333333333333333333333333333333",
			Time:     time.Date(2026, 9, 22, 5, 12, 44, 0, time.UTC),
			Hostname: "nas",
			Paths:    []string{"/srv/data", "/etc"},
		},
	}
	if len(got) != len(want) {
		t.Fatalf("List() returned %d snapshots, want %d: %+v", len(got), len(want), got)
	}
	for i := range want {
		g, w := got[i], want[i]
		if g.ID != w.ID {
			t.Errorf("List()[%d].ID = %q, want %q", i, g.ID, w.ID)
		}
		if !g.Time.Equal(w.Time) {
			t.Errorf("List()[%d].Time = %v, want %v", i, g.Time, w.Time)
		}
		if g.Hostname != w.Hostname {
			t.Errorf("List()[%d].Hostname = %q, want %q", i, g.Hostname, w.Hostname)
		}
		if !slices.Equal(g.Tags, w.Tags) {
			t.Errorf("List()[%d].Tags = %q, want %q", i, g.Tags, w.Tags)
		}
		if !slices.Equal(g.Paths, w.Paths) {
			t.Errorf("List()[%d].Paths = %q, want %q", i, g.Paths, w.Paths)
		}
	}
	c := onlyRunCall(t, fr)
	if want := p.listArgs(); !slices.Equal(c.args, want) {
		t.Errorf("List() argv = %q, want %q", c.args, want)
	}
	if !c.hasDeadline {
		t.Error("List() ctx has no deadline, want metadataTimeout applied")
	}
}

func TestListIgnoresStderrWarnings(t *testing.T) {
	fr := &fakeRunner{results: map[string]fakeResult{"snapshots": {
		stdout: readTestdata(t, "snapshots_two_hosts.json"),
		stderr: []byte("warning: cache is old\nLoad(<lock/abc>) returned error, retrying\n"),
	}}}
	p := newCmdProvider(t, fr)

	got, err := p.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	if len(got) != 3 {
		t.Errorf("List() returned %d snapshots, want 3", len(got))
	}
}

func TestListRejectsShortIDs(t *testing.T) {
	fr := &fakeRunner{results: map[string]fakeResult{"snapshots": {
		stdout: []byte(`[{"id":"e0bef23a","time":"2026-09-22T05:12:44Z","hostname":"h","paths":["/p"]}]`),
	}}}
	p := newCmdProvider(t, fr)

	got, err := p.List(context.Background())
	if err == nil {
		t.Fatalf("List() = %+v, nil, want an error for a short snapshot id", got)
	}
	if !strings.Contains(err.Error(), "64") {
		t.Errorf("List() error = %q, want it to mention 64", err)
	}
	if len(got) != 0 {
		t.Errorf("List() returned %d snapshots on error, want 0", len(got))
	}
}

func TestListNullIsEmpty(t *testing.T) {
	fr := &fakeRunner{results: map[string]fakeResult{"snapshots": {stdout: []byte("null")}}}
	p := newCmdProvider(t, fr)

	got, err := p.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	if len(got) != 0 {
		t.Errorf("List() = %+v, want empty", got)
	}
	if len(fr.runCalls()) != 1 {
		t.Errorf("List() made %d Run calls, want 1", len(fr.runCalls()))
	}
}

func TestListLockErrorNotRetried(t *testing.T) {
	fr := &fakeRunner{results: map[string]fakeResult{"snapshots": {stderr: []byte(lockStderr), err: errExit}}}
	p := newCmdProvider(t, fr)

	_, err := p.List(context.Background())
	if got, want := errcode.Of(err), errcode.RepoUnavailable; got != want {
		t.Errorf("errcode.Of(List() error) = %q, want %q", got, want)
	}
	calls := fr.runCalls()
	if len(calls) != 1 {
		t.Fatalf("List() made %d Run calls, want exactly 1", len(calls))
	}
	for _, c := range calls {
		if slices.Contains(c.args, "--no-lock") {
			t.Errorf("List() argv = %q, want no --no-lock", c.args)
		}
	}
}

func TestSnapReturnsSummaryID(t *testing.T) {
	fr := &fakeRunner{results: map[string]fakeResult{"backup": {stdout: readTestdata(t, "backup_summary.jsonl")}}}
	p := newCmdProvider(t, fr)
	req := provider.SnapRequest{Path: "/home/alex/project/docs", Host: "laptop", Tags: []string{"pre"}, Excludes: []string{".snapshot"}}

	got, err := p.Snap(context.Background(), req)
	if err != nil {
		t.Fatalf("Snap(%+v) error = %v, want nil", req, err)
	}
	if got != id2 {
		t.Errorf("Snap(%+v) = %q, want %q", req, got, id2)
	}
	c := onlyRunCall(t, fr)
	if want := p.snapArgs(req); !slices.Equal(c.args, want) {
		t.Errorf("Snap(%+v) argv = %q, want %q", req, c.args, want)
	}
	if c.hasDeadline {
		t.Error("Snap() ctx has a deadline, want only the caller's ctx (no internal timeout)")
	}
}

func TestSnapValidatesRequest(t *testing.T) {
	tests := []struct {
		name string
		req  provider.SnapRequest
	}{
		{name: "relative path", req: provider.SnapRequest{Path: "home/alex", Host: "h"}},
		{name: "empty path", req: provider.SnapRequest{Path: "", Host: "h"}},
		{name: "empty host", req: provider.SnapRequest{Path: "/home/alex", Host: ""}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fr := &fakeRunner{results: map[string]fakeResult{"backup": {stdout: readTestdata(t, "backup_summary.jsonl")}}}
			p := newCmdProvider(t, fr)

			got, err := p.Snap(context.Background(), tt.req)
			if err == nil {
				t.Errorf("Snap(%+v) = %q, nil, want an error", tt.req, got)
			}
			if n := len(fr.runCalls()); n != 0 {
				t.Errorf("Snap(%+v) made %d Run calls, want 0", tt.req, n)
			}
		})
	}
}

func TestSnapMissingSummaryErrors(t *testing.T) {
	status := `{"message_type":"status","percent_done":1,"total_files":3}` + "\n"
	tests := []struct {
		name   string
		stdout string
	}{
		{name: "status only", stdout: status + status},
		{name: "short snapshot id", stdout: status + `{"message_type":"summary","snapshot_id":"e0bef23a"}` + "\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fr := &fakeRunner{results: map[string]fakeResult{"backup": {stdout: []byte(tt.stdout)}}}
			p := newCmdProvider(t, fr)
			req := provider.SnapRequest{Path: "/home/alex", Host: "h"}

			if got, err := p.Snap(context.Background(), req); err == nil {
				t.Errorf("Snap(%+v) = %q, nil, want an error", req, got)
			}
		})
	}
}

func TestErrorsNeverLeakRepository(t *testing.T) {
	fr := &fakeRunner{reply: func(fakeCall) fakeResult {
		return fakeResult{stderr: []byte("Fatal: unable to open repository at " + repo + ": access denied\n"), err: errExit}
	}}
	p := newCmdProvider(t, fr)
	ctx := context.Background()

	_, validateErr := p.Validate(ctx)
	_, listErr := p.List(ctx)
	_, snapErr := p.Snap(ctx, provider.SnapRequest{Path: "/home/alex", Host: "h"})

	for name, err := range map[string]error{"Validate": validateErr, "List": listErr, "Snap": snapErr} {
		if err == nil {
			t.Errorf("%s() error = nil, want a repository failure", name)
			continue
		}
		if strings.Contains(err.Error(), "user:secret") {
			t.Errorf("%s() error = %q, want the repository redacted", name, err)
		}
	}
}
