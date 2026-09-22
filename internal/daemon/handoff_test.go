package daemon

import (
	"context"
	"log/slog"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/recovery"
)

// probingRecoverer tries the daemon lock from inside Daemon.Run.
type probingRecoverer struct {
	stateDir string

	mu     sync.Mutex
	err    error
	probed bool
}

func (r *probingRecoverer) Recover(context.Context) (recovery.Report, error) {
	err := tryLock(r.stateDir)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.err, r.probed = err, true
	return recovery.Report{}, nil
}

// tryLock takes and at once releases stateDir's lock, returning Lock's error.
func tryLock(stateDir string) error {
	unlock, err := Lock(stateDir)
	if err != nil {
		return err
	}
	unlock()
	return nil
}

func TestRunKeepsLockThroughHandoff(t *testing.T) {
	cfgPath, stateDir := writeValidConfig(t)
	h := newHarness(t)

	var (
		mu        sync.Mutex
		lockCalls int
	)
	orig := lockFunc
	lockFunc = func(dir string) (func(), error) {
		mu.Lock()
		lockCalls++
		mu.Unlock()
		return orig(dir)
	}
	t.Cleanup(func() { lockFunc = orig })

	probe := &probingRecoverer{stateDir: stateDir}
	var buildErr error
	built := make(chan struct{})
	build := func(_ context.Context, _ *config.Config, ln net.Listener, _ *slog.Logger) (Deps, error) {
		buildErr = tryLock(stateDir)
		close(built)
		deps := h.deps
		deps.Listener = ln
		deps.Recoverer = probe
		deps.Trace = nil
		return deps, nil
	}
	env, _, stderr := cmdEnv("", cfgPath)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		select {
		case <-built:
			time.Sleep(100 * time.Millisecond)
		case <-time.After(2 * time.Second):
		}
		cancel()
	}()

	code, panicked := runRecovering(t, 5*time.Second, func() int { return Command(build).Run(ctx, env, nil) })

	if panicked != nil {
		t.Fatalf("Command(build).Run panicked: %v, want exit 0", panicked)
	}
	if code != 0 {
		t.Errorf("Command(build).Run(valid config) = %d, want 0 (stderr %q)", code, stderr.String())
	}
	if got := errcode.Of(buildErr); got != errcode.StaleState {
		t.Errorf("Lock(%q) during Builder = %v (code %q), want code %q", stateDir, buildErr, got, errcode.StaleState)
	}
	probe.mu.Lock()
	probed, runErr := probe.probed, probe.err
	probe.mu.Unlock()
	if !probed {
		t.Errorf("Lock(%q) inside Daemon.Run not attempted, want a probe from Recover", stateDir)
	} else if got := errcode.Of(runErr); got != errcode.StaleState {
		t.Errorf("Lock(%q) inside Daemon.Run = %v (code %q), want code %q", stateDir, runErr, got, errcode.StaleState)
	}
	mu.Lock()
	gotCalls := lockCalls
	mu.Unlock()
	if gotCalls != 1 {
		t.Errorf("lockFunc calls = %d, want 1 (lock taken once and handed to Daemon.Run)", gotCalls)
	}
	assertUnlocked(t, stateDir)
}
