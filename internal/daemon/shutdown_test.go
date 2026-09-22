package daemon

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/ipc"
	"github.com/adeelahmad/snapback/internal/links"
	"github.com/adeelahmad/snapback/internal/resolver"
)

// awaitRun waits up to timeout for Run's result.
func awaitRun(t *testing.T, errc <-chan error, timeout time.Duration) (error, bool) {
	t.Helper()
	select {
	case err := <-errc:
		return err, true
	case <-time.After(timeout):
		return nil, false
	}
}

// runCancelable runs d with a context the test cancels itself.
func runCancelable(t *testing.T, d *Daemon) (context.CancelFunc, <-chan error) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	errc := make(chan error, 1)
	go func() { errc <- d.Run(ctx) }()
	t.Cleanup(cancel)
	return cancel, errc
}

func dialShutdown(t *testing.T, sock string) ipc.Response {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	c, err := ipc.Dial(ctx, sock)
	if err != nil {
		t.Fatalf("ipc.Dial(%q) = %v, want nil", sock, err)
	}
	defer func() { _ = c.Close() }()
	resp, err := c.Call(ctx, ipc.Request{V: 1, Op: ipc.OpShutdown})
	if err != nil {
		t.Fatalf("Call(shutdown) = %v, want nil", err)
	}
	return resp
}

func TestShutdownOrder(t *testing.T) {
	h := newHarness(t)
	sock := h.deps.Listener.Addr().String()
	d := New(h.cfg, h.deps)
	_, errc := runCancelable(t, d)
	waitState(t, d, 2*time.Second, func(s string) bool { return s == "ready" })

	resp := dialShutdown(t, sock)
	if !resp.OK {
		t.Errorf("Call(shutdown) = %+v, want OK", resp)
	}

	err, ok := awaitRun(t, errc, 3*time.Second)
	if !ok {
		t.Fatal("Run did not return after the shutdown op")
	}
	if err != nil {
		t.Errorf("Run() = %v, want nil", err)
	}

	want := []string{"listener.close", "discovery.stop", "history.unmount", "supervisor.stop"}
	got := h.rec.list()
	if len(got) < len(want) || !slices.Equal(got[len(got)-len(want):], want) {
		t.Errorf("shutdown steps = %q, want suffix %q", got, want)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if c, err := ipc.Dial(ctx, sock); err == nil {
		_ = c.Close()
		t.Errorf("ipc.Dial(%q) after shutdown = nil error, want failure", sock)
	}
}

func TestShutdownReportsBusyMount(t *testing.T) {
	h := newHarness(t)
	h.hist.unmountErr = syscall.EBUSY
	d := New(h.cfg, h.deps)
	cancel, errc := runCancelable(t, d)
	waitState(t, d, 2*time.Second, func(s string) bool { return s == "ready" })

	cancel()
	err, ok := awaitRun(t, errc, 3*time.Second)
	if !ok {
		t.Fatal("Run did not return after cancel")
	}
	if got := errcode.Of(err); got != errcode.MountFailure {
		t.Fatalf("errcode.Of(Run() = %v) = %q, want %q", err, got, errcode.MountFailure)
	}
	msg := err.Error()
	if !strings.Contains(msg, "history") || !strings.Contains(msg, "busy") {
		t.Errorf("Run() = %q, want text naming the history mount and busy", msg)
	}

	steps := h.rec.list()
	hi, si := slices.Index(steps, "history.unmount"), slices.Index(steps, "supervisor.stop")
	if hi < 0 || si < hi {
		t.Errorf("shutdown steps = %q, want supervisor.stop attempted after history.unmount", steps)
	}
	if slices.Contains(steps, "kill") {
		t.Errorf("shutdown steps = %q, want no kill", steps)
	}
}

func TestShutdownIsBounded(t *testing.T) {
	h := newHarness(t)
	h.hist.block = make(chan struct{})
	h.deps.ShutdownTimeout = 200 * time.Millisecond
	d := New(h.cfg, h.deps)
	cancel, errc := runCancelable(t, d)
	waitState(t, d, 2*time.Second, func(s string) bool { return s == "ready" })

	cancel()
	err, ok := awaitRun(t, errc, time.Second)
	if !ok {
		t.Fatal("Run did not return within 1s of cancel")
	}
	if got := errcode.Of(err); got != errcode.MountFailure {
		t.Fatalf("errcode.Of(Run() = %v) = %q, want %q", err, got, errcode.MountFailure)
	}
	if !strings.Contains(err.Error(), "timeout") && !strings.Contains(err.Error(), "timed out") {
		t.Errorf("Run() = %q, want text naming a timeout", err)
	}
}

// blockingRefresher blocks every Refresh until its ctx is done and reports
// the ctx error on done.
type blockingRefresher struct {
	started chan struct{}
	done    chan error
}

func (b *blockingRefresher) Refresh(ctx context.Context) (RefreshResult, error) {
	select {
	case b.started <- struct{}{}:
	default:
	}
	<-ctx.Done()
	select {
	case b.done <- ctx.Err():
	default:
	}
	return RefreshResult{}, ctx.Err()
}

func TestShutdownLeavesLinksAndCancelsFiniteOps(t *testing.T) {
	h := newHarness(t)
	sock := h.deps.Listener.Addr().String()

	root := h.cfg.Roots[0].LocalPath
	hist := t.TempDir()
	reg, err := links.OpenRegistry(filepath.Join(t.TempDir(), "links.db"))
	if err != nil {
		t.Fatalf("links.OpenRegistry: %v", err)
	}
	t.Cleanup(func() { _ = reg.Close() })
	eng := links.NewEngine(reg, links.Policy{
		LinkName:     ".snapshot",
		HistoryMount: hist,
		Roots:        []resolver.RootSpec{{ID: h.cfg.Roots[0].ID, LocalPath: root}},
	})
	dir := filepath.Join(root, "docs")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatalf("Mkdir(%q) = %v", dir, err)
	}
	if _, err := eng.Ensure(context.Background(), dir); err != nil {
		t.Fatalf("Ensure(%q) = %v, want nil", dir, err)
	}
	link := filepath.Join(dir, ".snapshot")
	before, err := os.Readlink(link)
	if err != nil {
		t.Fatalf("Readlink(%q) = %v", link, err)
	}

	br := &blockingRefresher{started: make(chan struct{}, 1), done: make(chan error, 1)}
	h.deps.Refresher = br
	h.deps.Linker = eng
	d := New(h.cfg, h.deps)
	_, errc := runCancelable(t, d)

	select {
	case <-br.started:
	case <-time.After(2 * time.Second):
		t.Fatal("refresh never started")
	}

	resp := dialShutdown(t, sock)
	if !resp.OK {
		t.Errorf("Call(shutdown) = %+v, want OK", resp)
	}

	select {
	case got := <-br.done:
		if !errors.Is(got, context.Canceled) {
			t.Errorf("blocked refresh returned %v, want %v", got, context.Canceled)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("blocked refresh was not cancelled by shutdown")
	}
	if _, ok := awaitRun(t, errc, 3*time.Second); !ok {
		t.Fatal("Run did not return after the shutdown op")
	}

	after, err := os.Readlink(link)
	if err != nil {
		t.Fatalf("Readlink(%q) after shutdown = %v, want the link kept", link, err)
	}
	if after != before {
		t.Errorf("Readlink(%q) = %q, want %q", link, after, before)
	}
	recs, err := eng.List()
	if err != nil {
		t.Fatalf("List() = %v", err)
	}
	if len(recs) != 1 {
		t.Errorf("len(List()) = %d, want 1", len(recs))
	}
}

func TestFailedStartupUnmountsStartedMounts(t *testing.T) {
	h := newHarness(t)
	refreshErr := errcode.New(errcode.InvalidConfig, "refresh", errors.New("bad catalog"))
	h.ref.err = refreshErr
	d := New(h.cfg, h.deps)
	_, errc := runCancelable(t, d)

	err, ok := awaitRun(t, errc, 3*time.Second)
	if !ok {
		t.Fatal("Run did not return after the failed first refresh")
	}
	if !errors.Is(err, refreshErr) {
		t.Errorf("Run() = %v, want %v", err, refreshErr)
	}

	steps := h.rec.list()
	si := slices.Index(steps, "supervisor.start")
	hi := slices.Index(steps, "history.unmount")
	bi := slices.Index(steps, "supervisor.stop")
	if si < 0 || hi < si || bi < hi {
		t.Errorf("startup steps = %q, want supervisor.start, then history.unmount, then supervisor.stop", steps)
	}
}
