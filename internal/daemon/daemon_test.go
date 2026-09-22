package daemon

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/history"
	"github.com/adeelahmad/snapback/internal/ipc"
	"github.com/adeelahmad/snapback/internal/recovery"
	"github.com/adeelahmad/snapback/internal/status"
)

func TestLockSecondInstanceIsStaleState(t *testing.T) {
	dir := t.TempDir()

	unlock, err := Lock(dir)
	if err != nil {
		t.Fatalf("Lock(%q) first = %v, want nil", dir, err)
	}

	unlock2, err := Lock(dir)
	if got := errcode.Of(err); got != errcode.StaleState {
		if unlock2 != nil {
			unlock2()
		}
		unlock()
		t.Fatalf("Lock(%q) second: errcode.Of(%v) = %q, want %q", dir, err, got, errcode.StaleState)
	}
	if !strings.Contains(err.Error(), "another daemon") {
		t.Errorf("Lock(%q) second = %q, want message naming another daemon", dir, err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "daemon.pid"))
	if err != nil {
		t.Errorf("read daemon.pid: %v", err)
	} else if got, want := strings.TrimSpace(string(data)), strconv.Itoa(os.Getpid()); got != want {
		t.Errorf("daemon.pid = %q, want %q", got, want)
	}

	unlock()
	unlock3, err := Lock(dir)
	if err != nil {
		t.Fatalf("Lock(%q) after unlock = %v, want nil", dir, err)
	}
	unlock3()
}

func TestRunStartupOrder(t *testing.T) {
	h := newHarness(t)
	d := New(h.cfg, h.deps)
	start(t, d)

	waitState(t, d, 2*time.Second, func(s string) bool { return s == "ready" })

	want := []string{"lock", "ipc", "recover", "supervisor.start", "refresh", "discovery.start", "prewarm"}
	got := h.rec.list()
	if len(got) < len(want) || !slices.Equal(got[:len(want)], want) {
		t.Errorf("startup steps = %q, want prefix %q", got, want)
	}
}

func TestRunRejectsInvalidConfigFirst(t *testing.T) {
	h := newHarness(t)
	noRoots := &config.Config{
		StateDir:     t.TempDir(),
		Repositories: h.cfg.Repositories,
	}
	for _, tc := range []struct {
		name string
		cfg  *config.Config
	}{
		{"nil", nil},
		{"no roots", noRoots},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			err := Run(ctx, tc.cfg, h.deps)
			if got := errcode.Of(err); got != errcode.InvalidConfig {
				t.Errorf("Run(%s): errcode.Of(%v) = %q, want %q", tc.name, err, got, errcode.InvalidConfig)
			}
			if got := h.rec.list(); len(got) != 0 {
				t.Errorf("Run(%s) steps = %q, want none", tc.name, got)
			}
		})
	}
	if _, err := os.Stat(filepath.Join(noRoots.StateDir, "daemon.lock")); !os.IsNotExist(err) {
		t.Errorf("stat daemon.lock = %v, want not exist (no lock taken)", err)
	}
}

func TestOfflineRepoIsDegradedNotEmpty(t *testing.T) {
	h := newHarness(t)
	h.sup.states = map[string]history.RepoState{
		"repoA": history.StateReady,
		"repoB": history.StateFailed,
	}
	h.ref.result = RefreshResult{Generation: 1, At: fixedNow, Failed: []string{"repoB"}}
	h.ref.err = errcode.New(errcode.RepoUnavailable, "refresh repoB", nil)
	d := New(h.cfg, h.deps)
	errc := start(t, d)

	got := waitState(t, d, 2*time.Second, func(s string) bool { return s != "starting" })
	if got != "degraded" {
		t.Errorf("Status().State = %q, want %q", got, "degraded")
	}

	want := []status.Repo{
		{ID: "repoA", State: string(history.StateReady)},
		{ID: "repoB", State: string(history.StateFailed), Code: errcode.RepoUnavailable},
	}
	if repos := d.Status().Repos; !slices.Equal(repos, want) {
		t.Errorf("Status().Repos = %+v, want %+v", repos, want)
	}

	select {
	case err := <-errc:
		t.Errorf("Run returned %v while a repository is offline, want still running", err)
	default:
	}

	steps := h.rec.list()
	for _, s := range []string{"discovery.start", "prewarm"} {
		if !slices.Contains(steps, s) {
			t.Errorf("startup steps = %q, want %q (degraded still starts it)", steps, s)
		}
	}
}

func TestStartupDoesNotWalkRoots(t *testing.T) {
	h := newHarness(t)
	root := h.cfg.Roots[0].LocalPath
	locked := filepath.Join(root, "locked")
	if err := os.Mkdir(locked, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(locked, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(locked, 0o700) })
	if err := syscall.Mkfifo(filepath.Join(root, "fifo"), 0o600); err != nil {
		t.Fatalf("mkfifo: %v", err)
	}

	d := New(h.cfg, h.deps)
	start(t, d)

	waitState(t, d, 2*time.Second, func(s string) bool { return s == "ready" })
	if got := h.linker.total(); got != 0 {
		t.Errorf("Linker.Ensure calls during startup = %d, want 0", got)
	}
}

func TestStartupRecoveryReportedInStatus(t *testing.T) {
	h := newHarness(t)
	h.deps.Recoverer = &fakeRecoverer{rec: h.rec, report: recovery.Report{
		Unmounted: []string{"a", "b"},
		Foreign:   []string{"c"},
	}}
	d := New(h.cfg, h.deps)
	start(t, d)
	waitState(t, d, 2*time.Second, func(s string) bool { return s == "ready" })

	s := decodeSnapshot(t, call(d, ipc.Request{Op: ipc.OpStatus}))
	want := &status.RecoverySummary{Unmounted: []string{"a", "b"}, Foreign: []string{"c"}}
	if !reflect.DeepEqual(s.Recovery, want) {
		t.Errorf("status op Recovery = %+v, want %+v", s.Recovery, want)
	}
}
