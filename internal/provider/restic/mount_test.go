package restic

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/provider"
)

// mountFixture builds a provider over f and starts a mount in a temp dir.
func mountFixture(t *testing.T, f *fakeRunner) (provider.MountHandle, string) {
	t.Helper()
	opts := validOptions()
	opts.Runner = f
	p := mustNew(t, opts)
	dir := t.TempDir()
	h, err := p.StartMount(context.Background(), dir)
	if err != nil {
		t.Fatalf("StartMount(%q) error = %v, want nil", dir, err)
	}
	if h == nil {
		t.Fatalf("StartMount(%q) = nil handle, want non-nil", dir)
	}
	return h, dir
}

// onlyProcess returns the single process handed out by f.
func onlyProcess(t *testing.T, f *fakeRunner) *fakeProcess {
	t.Helper()
	procs := f.processes()
	if len(procs) != 1 {
		t.Fatalf("Start calls = %d, want 1", len(procs))
	}
	return procs[0]
}

// closedWithin reports whether ch closes within d.
func closedWithin(ch <-chan struct{}, d time.Duration) bool {
	select {
	case <-ch:
		return true
	case <-time.After(d):
		return false
	}
}

func TestStartMountArgvAndNoContext(t *testing.T) {
	f := &fakeRunner{}
	opts := validOptions()
	opts.Runner = f
	p := mustNew(t, opts)
	dir := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h, err := p.StartMount(ctx, dir)
	if err != nil {
		t.Fatalf("StartMount(%q) error = %v, want nil", dir, err)
	}

	calls := f.recorded()
	if len(calls) != 1 || !calls[0].start {
		t.Fatalf("recorded calls = %+v, want exactly one Start", calls)
	}
	if got, want := calls[0].name, bin; got != want {
		t.Errorf("Start name = %q, want %q", got, want)
	}
	if got, want := calls[0].args, p.mountArgs(dir); !slices.Equal(got, want) {
		t.Errorf("Start argv = %q, want %q", got, want)
	}
	if got, want := calls[0].env, p.childEnv(); !slices.Equal(got, want) {
		t.Errorf("Start env = %q, want %q", got, want)
	}
	cancel()
	if closedWithin(h.Done(), 100*time.Millisecond) {
		t.Errorf("Done() closed after caller ctx cancelled, want open (the mount outlives ctx)")
	}
	onlyProcess(t, f).exit(nil)
}

func TestMountReadyWhenIdsListable(t *testing.T) {
	f := &fakeRunner{}
	h, dir := mountFixture(t, f)
	defer onlyProcess(t, f).exit(nil)
	go func() {
		time.Sleep(100 * time.Millisecond)
		_ = os.Mkdir(filepath.Join(dir, "ids"), 0o755)
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := h.Ready(ctx); err != nil {
		t.Errorf("Ready() = %v, want nil", err)
	}
	if got := h.Dir(); got != dir {
		t.Errorf("Dir() = %q, want %q", got, dir)
	}
}

func TestMountReadyFailsWhenProcessExits(t *testing.T) {
	f := &fakeRunner{}
	h, _ := mountFixture(t, f)
	onlyProcess(t, f).exit(errors.New("fusermount: failed"))
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := h.Ready(ctx)

	if got, want := errcode.Of(err), errcode.MountFailure; got != want {
		t.Errorf("errcode.Of(Ready()) = %q (err %v), want %q", got, err, want)
	}
	if !closedWithin(h.Done(), time.Second) {
		t.Errorf("Done() open after process exit, want closed")
	}
}

func TestMountReadyHonoursContext(t *testing.T) {
	f := &fakeRunner{}
	h, _ := mountFixture(t, f)
	defer onlyProcess(t, f).exit(nil)
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	if err := h.Ready(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Ready() = %v, want context.DeadlineExceeded", err)
	}
}

func TestMountDoneOnSilentDeath(t *testing.T) {
	f := &fakeRunner{}
	h, dir := mountFixture(t, f)
	if err := os.Mkdir(filepath.Join(dir, "ids"), 0o755); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := h.Ready(ctx); err != nil {
		t.Fatalf("Ready() = %v, want nil", err)
	}

	onlyProcess(t, f).exit(errors.New("killed"))

	if !closedWithin(h.Done(), time.Second) {
		t.Errorf("Done() still open 1s after process death, want closed")
	}
}

func TestMountStopGraceful(t *testing.T) {
	f := &fakeRunner{newProcess: func() *fakeProcess {
		p := newFakeProcess()
		p.exitOnInterrupt = true
		return p
	}}
	h, _ := mountFixture(t, f)
	proc := onlyProcess(t, f)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	first := h.Stop(ctx)

	if got, want := proc.signalled(), []os.Signal{os.Interrupt}; !slices.Equal(got, want) {
		t.Errorf("signals = %v, want %v", got, want)
	}
	if got := len(f.runCalls()); got != 0 {
		t.Errorf("unmount Run calls = %d, want 0", got)
	}
	if got := proc.killCount(); got != 0 {
		t.Errorf("Kill calls = %d, want 0", got)
	}
	if second := h.Stop(ctx); second != first {
		t.Errorf("second Stop() = %v, want %v", second, first)
	}
	if got := len(proc.signalled()); got != 1 {
		t.Errorf("signals after second Stop = %d, want 1", got)
	}
	if got := len(f.recorded()); got != 1 {
		t.Errorf("runner calls after second Stop = %d, want 1 (the Start)", got)
	}
}

func TestMountStopEscalates(t *testing.T) {
	tests := []struct {
		goos     string
		wantName string
		wantArgs func(dir string) []string
	}{
		{goos: "linux", wantName: "fusermount3", wantArgs: func(dir string) []string { return []string{"-u", dir} }},
		{goos: "darwin", wantName: "umount", wantArgs: func(dir string) []string { return []string{dir} }},
	}
	for _, tt := range tests {
		t.Run(tt.goos, func(t *testing.T) {
			f := &fakeRunner{}
			killsAtUnmount := -1
			f.reply = func(fakeCall) fakeResult {
				if procs := f.processes(); len(procs) == 1 {
					killsAtUnmount = procs[0].killCount()
				}
				return fakeResult{}
			}
			opts := validOptions()
			opts.Runner = f
			p := mustNew(t, opts)
			p.goos = tt.goos
			dir := t.TempDir()
			h, err := p.StartMount(context.Background(), dir)
			if err != nil {
				t.Fatalf("StartMount(%q) error = %v, want nil", dir, err)
			}
			proc := onlyProcess(t, f)
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			_ = h.Stop(ctx)

			runs := f.runCalls()
			if len(runs) != 1 {
				t.Fatalf("unmount Run calls = %d, want 1", len(runs))
			}
			if got := filepath.Base(runs[0].name); got != tt.wantName {
				t.Errorf("unmount command = %q, want %q", got, tt.wantName)
			}
			if got, want := runs[0].args, tt.wantArgs(dir); !slices.Equal(got, want) {
				t.Errorf("unmount args = %q, want %q", got, want)
			}
			if killsAtUnmount != 0 {
				t.Errorf("Kill calls before unmount = %d, want 0 (unmount then Kill)", killsAtUnmount)
			}
			if got := proc.killCount(); got != 1 {
				t.Errorf("Kill calls = %d, want 1", got)
			}
			if !closedWithin(h.Done(), time.Second) {
				t.Errorf("Done() open after Stop, want closed")
			}
		})
	}
}

func TestSnapshotRootFullID(t *testing.T) {
	p := mustNew(t, validOptions())
	if got, want := p.SnapshotRoot("/m", id1), "/m/ids/"+id1; got != want {
		t.Errorf("SnapshotRoot(%q, %q) = %q, want %q", "/m", id1, got, want)
	}
}
