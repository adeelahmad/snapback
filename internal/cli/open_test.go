package cli

import (
	"context"
	"errors"
	"os/exec"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/links"
	"github.com/adeelahmad/snapback/internal/provider"
)

const xdgOpenPath = "/usr/bin/xdg-open"

// openRecorder records the order of engine, daemon and Exec calls.
type openRecorder struct {
	mu        sync.Mutex
	order     []string
	ensured   []string
	execName  string
	execArgs  []string
	execs     int
	available bool
}

func (r *openRecorder) log(s string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.order = append(r.order, s)
}

// Ensure implements Linker.
func (r *openRecorder) Ensure(_ context.Context, dir string) (links.Result, error) {
	r.log("Ensure(" + dir + ")")
	r.mu.Lock()
	r.ensured = append(r.ensured, dir)
	r.mu.Unlock()
	return links.Result{Key: "k", Path: dir + "/.snapshot"}, nil
}

func (r *openRecorder) List() ([]links.Record, error) { return nil, nil }

func (r *openRecorder) Repair(context.Context) (links.RepairReport, error) {
	return links.RepairReport{}, nil
}

func (r *openRecorder) RemoveManaged(context.Context) (links.RepairReport, error) {
	return links.RepairReport{}, nil
}

// HistoryAvailable implements Daemon.
func (r *openRecorder) HistoryAvailable(context.Context) (bool, error) {
	r.log("HistoryAvailable")
	return r.available, nil
}

func (r *openRecorder) SnapSubmitted(context.Context, string, provider.SnapshotID) error {
	return nil
}

func (r *openRecorder) Visible(context.Context, provider.SnapshotID) (bool, error) {
	return false, nil
}

// openDeps returns Deps over r with working directory /w, a LookPath that
// finds xdg-open and an Exec that records its call.
func openDeps(r *openRecorder) Deps {
	return Deps{
		Linker: r,
		Daemon: func(context.Context) (Daemon, error) { return r, nil },
		Getwd:  func() (string, error) { return "/w", nil },
		LookPath: func(file string) (string, error) {
			if file != "xdg-open" {
				return "", exec.ErrNotFound
			}
			return xdgOpenPath, nil
		},
		Exec: func(_ context.Context, name string, args []string) error {
			r.log("Exec")
			r.mu.Lock()
			defer r.mu.Unlock()
			r.execs++
			r.execName = name
			r.execArgs = args
			return nil
		},
		OpenTimeout: 10 * time.Second,
	}
}

func TestOpenEnsuresThenRunsXdgOpen(t *testing.T) {
	r := &openRecorder{available: true}
	env, _, _ := newEnv(nil)

	got := OpenCommand(openDeps(r)).Run(context.Background(), env, []string{"d"})

	if got != 0 {
		t.Fatalf("open d = %d, want 0", got)
	}
	if want := []string{"Ensure(/w/d)", "HistoryAvailable", "Exec"}; !reflect.DeepEqual(r.order, want) {
		t.Errorf("call order = %q, want %q", r.order, want)
	}
	if r.execName != xdgOpenPath {
		t.Errorf("Exec name = %q, want %q", r.execName, xdgOpenPath)
	}
	if want := []string{"/w/d/.snapshot"}; !reflect.DeepEqual(r.execArgs, want) {
		t.Errorf("Exec args = %q, want %q", r.execArgs, want)
	}
}

func TestOpenHistoryUnavailable(t *testing.T) {
	r := &openRecorder{available: false}
	env, out, _ := newEnv(nil)

	got := OpenCommand(openDeps(r)).Run(context.Background(), env, []string{"--json", "d"})

	if got != 1 {
		t.Fatalf("open --json d = %d, want 1", got)
	}
	e := decodeEnvelope(t, out.Bytes())
	if want := string(errcode.RepoUnavailable); e.Code != want {
		t.Errorf("code = %q, want %q", e.Code, want)
	}
	if !strings.Contains(e.Fix, "snapback status") {
		t.Errorf("fix = %q, want it to mention %q", e.Fix, "snapback status")
	}
	if r.execs != 0 {
		t.Errorf("Exec calls = %d, want 0", r.execs)
	}
}

func TestOpenDaemonUnreachable(t *testing.T) {
	r := &openRecorder{available: true}
	d := openDeps(r)
	d.Daemon = func(context.Context) (Daemon, error) {
		return nil, errors.New("dial unix /run/snapback.sock: connection refused")
	}
	env, out, _ := newEnv(nil)

	got := OpenCommand(d).Run(context.Background(), env, []string{"--json", "d"})

	if got != 1 {
		t.Fatalf("open --json d = %d, want 1", got)
	}
	e := decodeEnvelope(t, out.Bytes())
	if want := string(errcode.RepoUnavailable); e.Code != want {
		t.Errorf("code = %q, want %q", e.Code, want)
	}
	if r.execs != 0 {
		t.Errorf("Exec calls = %d, want 0", r.execs)
	}
	if len(r.ensured) != 1 {
		t.Errorf("Ensure calls = %d, want 1", len(r.ensured))
	}
}

func TestOpenXdgOpenMissing(t *testing.T) {
	r := &openRecorder{available: true}
	d := openDeps(r)
	d.LookPath = func(string) (string, error) { return "", exec.ErrNotFound }
	env, out, _ := newEnv(nil)

	got := OpenCommand(d).Run(context.Background(), env, []string{"--json", "d"})

	if got != 1 {
		t.Fatalf("open --json d = %d, want 1", got)
	}
	e := decodeEnvelope(t, out.Bytes())
	if want := string(errcode.PrereqMissing); e.Code != want {
		t.Errorf("code = %q, want %q", e.Code, want)
	}
	if r.execs != 0 {
		t.Errorf("Exec calls = %d, want 0", r.execs)
	}
}

func TestOpenBoundedByTimeout(t *testing.T) {
	r := &openRecorder{available: true}
	d := openDeps(r)
	d.OpenTimeout = 50 * time.Millisecond
	var (
		mu          sync.Mutex
		hadDeadline bool
	)
	d.Exec = func(ctx context.Context, _ string, _ []string) error {
		_, ok := ctx.Deadline()
		mu.Lock()
		hadDeadline = ok
		mu.Unlock()
		<-ctx.Done()
		return ctx.Err()
	}
	env, out, _ := newEnv(nil)

	done := make(chan int, 1)
	go func() {
		done <- OpenCommand(d).Run(context.Background(), env, []string{"--json", "d"})
	}()

	var got int
	select {
	case got = <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("open --json d did not return within 2s")
	}
	if got != 1 {
		t.Fatalf("open --json d = %d, want 1", got)
	}
	e := decodeEnvelope(t, out.Bytes())
	if want := string(errcode.PrereqMissing); e.Code != want {
		t.Errorf("code = %q, want %q", e.Code, want)
	}
	mu.Lock()
	defer mu.Unlock()
	if !hadDeadline {
		t.Errorf("Exec ctx had no deadline, want one from OpenTimeout")
	}
}

func TestOpenMetacharacterPath(t *testing.T) {
	r := &openRecorder{available: true}
	env, _, _ := newEnv(nil)

	OpenCommand(openDeps(r)).Run(context.Background(), env, []string{"$(touch x)"})

	if r.execName != xdgOpenPath {
		t.Errorf("Exec name = %q, want %q (the LookPath result, not a shell)", r.execName, xdgOpenPath)
	}
	if len(r.execArgs) != 1 {
		t.Fatalf("Exec args = %q, want exactly 1 element", r.execArgs)
	}
	if want := "/w/$(touch x)/.snapshot"; r.execArgs[0] != want {
		t.Errorf("Exec args[0] = %q, want %q", r.execArgs[0], want)
	}
}
