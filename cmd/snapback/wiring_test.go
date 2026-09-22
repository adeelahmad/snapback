package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/cli"
	"github.com/adeelahmad/snapback/internal/ipc"
	"github.com/adeelahmad/snapback/internal/links"
	"github.com/adeelahmad/snapback/internal/provider"
	"github.com/adeelahmad/snapback/internal/status"
	"github.com/adeelahmad/snapback/internal/web"
)

var allNames = []string{
	"config", "doctor", "install", "link", "links", "notify", "open", "refresh",
	"run", "seed", "service", "shell-hook", "snap", "status", "version", "web",
}

const wireID1 = provider.SnapshotID("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")

func TestAllCommandNames(t *testing.T) {
	cmds := allCommands(cli.Deps{})

	seen := map[string]bool{}
	var got []string
	for _, c := range cmds {
		if seen[c.Name] {
			t.Errorf("allCommands(cli.Deps{}) has duplicate %q", c.Name)
		}
		seen[c.Name] = true
		got = append(got, c.Name)
	}
	slices.Sort(got)
	if !slices.Equal(got, allNames) {
		t.Errorf("allCommands(cli.Deps{}) names = %v, want %v", got, allNames)
	}
}

func TestConfigFallbackIsWebSetup(t *testing.T) {
	got, want := configFallback(), web.ConfigCommand()

	if got.Name != want.Name {
		t.Errorf("configFallback().Name = %q, want %q", got.Name, want.Name)
	}
	if got.Summary != want.Summary {
		t.Errorf("configFallback().Summary = %q, want %q", got.Summary, want.Summary)
	}

	f := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(f, []byte("version: 1\nbogus_field: true\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := run([]string{"--config", f, "config", "validate", "--json"}, &stdout, &stderr)
	if code != 1 {
		t.Errorf("run(--config f config validate --json) = %d, want 1 (stderr %q)", code, stderr.String())
	}
	var env struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatalf("run(--config f config validate --json) stdout = %q, not JSON: %v", stdout.String(), err)
	}
	if got, want := env.Code, "invalid_configuration"; got != want {
		t.Errorf("run(--config f config validate --json) code = %q, want %q", got, want)
	}
}

// responder is an in-test NDJSON daemon that records every request and
// answers each with reply.
type responder struct {
	mu    sync.Mutex
	reqs  []ipc.Request
	reply func(n int, req ipc.Request) ipc.Response
}

func (r *responder) requests() []ipc.Request {
	r.mu.Lock()
	defer r.mu.Unlock()
	return slices.Clone(r.reqs)
}

func (r *responder) serveConn(conn net.Conn) {
	defer func() { _ = conn.Close() }()
	sc := bufio.NewScanner(conn)
	for sc.Scan() {
		var req ipc.Request
		if err := json.Unmarshal(sc.Bytes(), &req); err != nil {
			return
		}
		r.mu.Lock()
		r.reqs = append(r.reqs, req)
		resp := r.reply(len(r.reqs), req)
		r.mu.Unlock()
		b, err := json.Marshal(resp)
		if err != nil {
			return
		}
		if _, err := conn.Write(append(b, '\n')); err != nil {
			return
		}
	}
}

// startResponder listens on a short socket path (macOS caps sun_path at 104
// bytes) and serves r until the test ends.
func startResponder(t *testing.T, r *responder) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "sb")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	sock := filepath.Join(dir, "d.sock")
	l, err := ipc.Listen(sock)
	if err != nil {
		t.Fatalf("ipc.Listen(%q) = %v", sock, err)
	}
	t.Cleanup(func() { _ = l.Close() })
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			go r.serveConn(conn)
		}
	}()
	return sock
}

// statusResp encodes s the way the daemon's OpStatus handler does: a
// lower-case "state" key over status.Snapshot's Go-name keys.
func statusResp(t *testing.T, s status.Snapshot) ipc.Response {
	t.Helper()
	b, err := json.Marshal(struct {
		status.Snapshot
		State string `json:"state"`
	}{s, s.State})
	if err != nil {
		t.Errorf("json.Marshal(status) = %v", err)
	}
	return ipc.Response{OK: true, Data: b}
}

func TestIPCDaemonHistoryAvailable(t *testing.T) {
	tests := []struct {
		state string
		want  bool
	}{
		{"ready", true},
		{"unavailable", false},
	}
	for _, tc := range tests {
		t.Run(tc.state, func(t *testing.T) {
			r := &responder{reply: func(int, ipc.Request) ipc.Response {
				return statusResp(t, status.Snapshot{State: tc.state})
			}}
			sock := startResponder(t, r)
			dial := dialDaemon(sock)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			d, err := dial(ctx)
			if err != nil {
				t.Fatalf("dialDaemon(%q)(ctx) error = %v, want nil", sock, err)
			}
			got, err := d.HistoryAvailable(ctx)

			if err != nil {
				t.Errorf("HistoryAvailable() error = %v, want nil", err)
			}
			if got != tc.want {
				t.Errorf("HistoryAvailable() with state %q = %v, want %v", tc.state, got, tc.want)
			}
			reqs := r.requests()
			if len(reqs) != 1 {
				t.Fatalf("daemon got %d requests, want 1", len(reqs))
			}
			if reqs[0].Op != ipc.OpStatus || reqs[0].V != 1 {
				t.Errorf("request = {Op:%q V:%d}, want {Op:%q V:1}", reqs[0].Op, reqs[0].V, ipc.OpStatus)
			}
		})
	}
}

func TestIPCDaemonSnapSubmittedAndVisible(t *testing.T) {
	statuses := 0
	r := &responder{reply: func(_ int, req ipc.Request) ipc.Response {
		if req.Op != ipc.OpStatus {
			return ipc.Response{OK: true}
		}
		statuses++
		if statuses == 1 {
			return statusResp(t, status.Snapshot{State: "ready", Pending: []provider.SnapshotID{wireID1}})
		}
		return statusResp(t, status.Snapshot{State: "ready", Pending: []provider.SnapshotID{}})
	}}
	sock := startResponder(t, r)
	dial := dialDaemon(sock)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	d, err := dial(ctx)
	if err != nil {
		t.Fatalf("dialDaemon(%q)(ctx) error = %v, want nil", sock, err)
	}

	if err := d.SnapSubmitted(ctx, "r1", wireID1); err != nil {
		t.Errorf("SnapSubmitted(r1, id1) error = %v, want nil", err)
	}
	first, err1 := d.Visible(ctx, wireID1)
	second, err2 := d.Visible(ctx, wireID1)

	if err1 != nil || err2 != nil {
		t.Errorf("Visible(id1) errors = %v, %v, want nil, nil", err1, err2)
	}
	if first || !second {
		t.Errorf("Visible(id1) twice = %v, %v, want false, true", first, second)
	}
	var submitted []ipc.Request
	for _, req := range r.requests() {
		if req.Op == ipc.OpSnapSubmitted {
			submitted = append(submitted, req)
		}
	}
	if len(submitted) != 1 {
		t.Fatalf("daemon got %d %q requests, want 1 (all: %+v)", len(submitted), ipc.OpSnapSubmitted, r.requests())
	}
	if submitted[0].ID != wireID1 {
		t.Errorf("%q request ID = %q, want %q", ipc.OpSnapSubmitted, submitted[0].ID, wireID1)
	}
}

// ensureLinker is a cli.Linker whose Ensure reports a created link.
type ensureLinker struct{}

func (ensureLinker) Ensure(_ context.Context, dir string) (links.Result, error) {
	return links.Result{Key: "k1", Created: true, Path: dir + "/.snapshot"}, nil
}

func (ensureLinker) List() ([]links.Record, error) { return nil, nil }

func (ensureLinker) Repair(context.Context) (links.RepairReport, error) {
	return links.RepairReport{}, nil
}

func (ensureLinker) RemoveManaged(context.Context) (links.RepairReport, error) {
	return links.RepairReport{}, nil
}

func TestIPCDaemonDialFailure(t *testing.T) {
	dir, err := os.MkdirTemp("", "sb")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	sock := filepath.Join(dir, "missing.sock")
	factory := dialDaemon(sock)

	if _, err := factory(context.Background()); err == nil {
		t.Errorf("dialDaemon(%q)(ctx) error = nil, want non-nil", sock)
	}

	deps := cli.Deps{
		Linker:   ensureLinker{},
		Daemon:   factory,
		Getwd:    func() (string, error) { return "/w", nil },
		LookPath: func(string) (string, error) { return "/usr/bin/xdg-open", nil },
		Exec: func(context.Context, string, []string) error {
			t.Error("Exec called, want no opener run")
			return nil
		},
	}
	var stdout, stderr bytes.Buffer
	env := cli.Env{Stdout: &stdout, Stderr: &stderr, Getenv: func(string) string { return "" }}
	code := cli.OpenCommand(deps).Run(context.Background(), env, []string{"--json", "d"})
	if code != 1 {
		t.Errorf("open --json d with dial failure = %d, want 1 (stderr %q)", code, stderr.String())
	}
	var out struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
		t.Fatalf("open --json d stdout = %q, not JSON: %v", stdout.String(), err)
	}
	if got, want := out.Code, "repository_unavailable"; got != want {
		t.Errorf("open --json d code = %q, want %q", got, want)
	}
}

func TestBinaryHelpListsAllCommands(t *testing.T) {
	bin := buildBinary(t, "")

	stdout, stderr, err := execBinary(bin, "help")
	if err != nil {
		t.Fatalf("snapback help: err = %v (stderr %q), want exit 0", err, stderr)
	}
	for _, name := range allNames {
		if !strings.Contains(stdout, name) {
			t.Errorf("snapback help stdout = %q, want it to contain %q", stdout, name)
		}
	}

	stdout, _, err = execBinary(bin, "version")
	if err != nil {
		t.Fatalf("snapback version: err = %v, want exit 0", err)
	}
	if !strings.HasPrefix(stdout, "snapback dev (commit none, target ") {
		t.Errorf("snapback version stdout = %q, want prefix %q", stdout, "snapback dev (commit none, target ")
	}
}
