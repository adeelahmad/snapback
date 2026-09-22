package service

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/errcode"
)

const foreignUnit = "[Service]\nExecStart=/usr/bin/other\n"

// fakeRunner records every argv and answers from a canned stdout table keyed
// by the space-joined argv.
type fakeRunner struct {
	mu     sync.Mutex
	calls  [][]string
	stdout map[string]string
}

func (f *fakeRunner) run(_ context.Context, name string, args ...string) ([]byte, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	argv := append([]string{name}, args...)
	f.calls = append(f.calls, argv)
	return []byte(f.stdout[strings.Join(argv, " ")]), nil
}

func (f *fakeRunner) argv() [][]string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Clone(f.calls)
}

// scriptedReady returns its states in order, repeating the last one forever.
type scriptedReady struct {
	mu     sync.Mutex
	states []string
	n      int
}

func (r *scriptedReady) ready(context.Context) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	i := min(r.n, len(r.states)-1)
	r.n++
	return r.states[i], nil
}

func (r *scriptedReady) calls() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.n
}

func systemctl(args ...string) []string {
	return append([]string{"systemctl", "--user"}, args...)
}

func hasArgv(calls [][]string, want []string) bool {
	return slices.ContainsFunc(calls, func(c []string) bool { return slices.Equal(c, want) })
}

func writeUnit(t *testing.T, dir, body string) string {
	t.Helper()
	path := filepath.Join(dir, "snapback.service")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write unit: %v", err)
	}
	return path
}

func TestInstallWritesReloadsEnablesAndWaits(t *testing.T) {
	dir := t.TempDir()
	run := &fakeRunner{}
	ready := &scriptedReady{states: []string{"starting", "ready"}}
	s := &Systemd{UnitDir: dir, Run: run.run, Ready: ready.ready, ReadyTimeout: 5 * time.Second}

	if err := s.Install(t.Context(), userOpts); err != nil {
		t.Fatalf("Install(userOpts) = %v, want nil", err)
	}

	path := filepath.Join(dir, "snapback.service")
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read installed unit: %v", err)
	}
	if want := readGolden(t, "user.service.golden"); string(got) != want {
		t.Errorf("installed unit = %q, want %q", got, want)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat installed unit: %v", err)
	}
	if mode := info.Mode().Perm(); mode != 0o644 {
		t.Errorf("unit mode = %o, want 644", mode)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read unit dir: %v", err)
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	if want := []string{"snapback.service"}; !slices.Equal(names, want) {
		t.Errorf("unit dir entries = %v, want %v (no leftover temp files)", names, want)
	}
	wantCalls := [][]string{systemctl("daemon-reload"), systemctl("enable", "--now", "snapback.service")}
	if calls := run.argv(); !slices.EqualFunc(calls, wantCalls, slices.Equal) {
		t.Errorf("runner argv = %q, want %q", calls, wantCalls)
	}
	if n := ready.calls(); n < 2 {
		t.Errorf("readiness checks = %d, want at least 2", n)
	}
}

func TestInstallIdempotent(t *testing.T) {
	dir := t.TempDir()
	run := &fakeRunner{}
	ready := &scriptedReady{states: []string{"ready"}}
	s := &Systemd{UnitDir: dir, Run: run.run, Ready: ready.ready, ReadyTimeout: 5 * time.Second}

	if err := s.Install(t.Context(), userOpts); err != nil {
		t.Fatalf("first Install(userOpts) = %v, want nil", err)
	}
	path := filepath.Join(dir, "snapback.service")
	before, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat unit after first Install: %v", err)
	}
	beforeBytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read unit after first Install: %v", err)
	}
	firstCalls := len(run.argv())
	firstReady := ready.calls()

	if err := s.Install(t.Context(), userOpts); err != nil {
		t.Fatalf("second Install(userOpts) = %v, want nil", err)
	}

	after, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat unit after second Install: %v", err)
	}
	afterBytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read unit after second Install: %v", err)
	}
	if !after.ModTime().Equal(before.ModTime()) {
		t.Errorf("unit mtime = %v, want unchanged %v", after.ModTime(), before.ModTime())
	}
	if string(afterBytes) != string(beforeBytes) {
		t.Errorf("unit bytes = %q, want unchanged %q", afterBytes, beforeBytes)
	}
	second := run.argv()[firstCalls:]
	if hasArgv(second, systemctl("daemon-reload")) {
		t.Errorf("second Install argv = %q, want no daemon-reload", second)
	}
	if want := systemctl("enable", "--now", "snapback.service"); !hasArgv(second, want) {
		t.Errorf("second Install argv = %q, want %q", second, want)
	}
	if n := ready.calls(); n <= firstReady {
		t.Errorf("readiness checks after second Install = %d, want more than %d", n, firstReady)
	}
}

func TestInstallRefusesForeignUnit(t *testing.T) {
	dir := t.TempDir()
	path := writeUnit(t, dir, foreignUnit)
	run := &fakeRunner{}
	ready := &scriptedReady{states: []string{"ready"}}
	s := &Systemd{UnitDir: dir, Run: run.run, Ready: ready.ready, ReadyTimeout: 5 * time.Second}

	err := s.Install(t.Context(), userOpts)

	if got := errcode.Of(err); got != errcode.StaleState {
		t.Errorf("errcode.Of(Install(foreign)) = %q, want %q (err %v)", got, errcode.StaleState, err)
	}
	if err == nil || !strings.Contains(err.Error(), "foreign") {
		t.Errorf("Install(foreign) error = %v, want it to mention foreign", err)
	}
	got, rerr := os.ReadFile(path)
	if rerr != nil {
		t.Fatalf("read foreign unit: %v", rerr)
	}
	if string(got) != foreignUnit {
		t.Errorf("foreign unit = %q, want unchanged %q", got, foreignUnit)
	}
	if calls := run.argv(); len(calls) != 0 {
		t.Errorf("runner argv = %q, want no calls", calls)
	}
}

func TestInstallNotReadyExplainsBothStates(t *testing.T) {
	dir := t.TempDir()
	run := &fakeRunner{}
	ready := &scriptedReady{states: []string{"starting"}}
	s := &Systemd{UnitDir: dir, Run: run.run, Ready: ready.ready, ReadyTimeout: 300 * time.Millisecond}

	start := time.Now()
	err := s.Install(t.Context(), userOpts)
	elapsed := time.Since(start)

	if elapsed > time.Second {
		t.Errorf("Install(not ready) took %v, want at most 1s", elapsed)
	}
	if got := errcode.Of(err); got != errcode.StaleState {
		t.Errorf("errcode.Of(Install(not ready)) = %q, want %q (err %v)", got, errcode.StaleState, err)
	}
	path := filepath.Join(dir, "snapback.service")
	msg := ""
	if err != nil {
		msg = err.Error()
	}
	for _, want := range []string{"installed and enabled at " + path, "not ready", "journalctl --user -u snapback"} {
		if !strings.Contains(msg, want) {
			t.Errorf("Install(not ready) error = %q, want it to contain %q", msg, want)
		}
	}
	if _, serr := os.Stat(path); serr != nil {
		t.Errorf("unit file after not-ready Install: %v, want it left in place", serr)
	}
}

func TestLifecycleArgv(t *testing.T) {
	dir := t.TempDir()
	writeUnit(t, dir, readGolden(t, "user.service.golden"))
	run := &fakeRunner{stdout: map[string]string{
		"systemctl --user is-active snapback.service": "active\n",
	}}
	s := &Systemd{UnitDir: dir, Run: run.run}

	if err := s.Start(t.Context()); err != nil {
		t.Errorf("Start() = %v, want nil", err)
	}
	if err := s.Stop(t.Context()); err != nil {
		t.Errorf("Stop() = %v, want nil", err)
	}
	got, err := s.Status(t.Context())
	if err != nil {
		t.Errorf("Status() error = %v, want nil", err)
	}
	if got != "active" {
		t.Errorf("Status() = %q, want %q", got, "active")
	}
	want := [][]string{
		systemctl("start", "snapback.service"),
		systemctl("stop", "snapback.service"),
		systemctl("is-active", "snapback.service"),
	}
	if calls := run.argv(); !slices.EqualFunc(calls, want, slices.Equal) {
		t.Errorf("runner argv = %q, want %q", calls, want)
	}
}

func TestUninstallRemovesOnlyOwnedUnit(t *testing.T) {
	t.Run("owned", func(t *testing.T) {
		dir := t.TempDir()
		path := writeUnit(t, dir, readGolden(t, "user.service.golden"))
		stateDir := t.TempDir()
		siblings := []string{"config.yaml", "links.db", "password"}
		for _, name := range siblings {
			if err := os.WriteFile(filepath.Join(stateDir, name), []byte(name), 0o600); err != nil {
				t.Fatalf("write sibling %s: %v", name, err)
			}
		}
		run := &fakeRunner{}
		s := &Systemd{UnitDir: dir, Run: run.run}

		if err := s.Uninstall(t.Context()); err != nil {
			t.Fatalf("Uninstall(owned) = %v, want nil", err)
		}

		calls := run.argv()
		for _, want := range [][]string{
			systemctl("stop", "snapback.service"),
			systemctl("disable", "snapback.service"),
			systemctl("daemon-reload"),
		} {
			if !hasArgv(calls, want) {
				t.Errorf("Uninstall(owned) argv = %q, want %q", calls, want)
			}
		}
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("stat unit after Uninstall(owned) = %v, want not exist", err)
		}
		for _, name := range siblings {
			if _, err := os.Stat(filepath.Join(stateDir, name)); err != nil {
				t.Errorf("sibling %s after Uninstall: %v, want present", name, err)
			}
		}
	})

	t.Run("foreign", func(t *testing.T) {
		dir := t.TempDir()
		path := writeUnit(t, dir, foreignUnit)
		run := &fakeRunner{}
		s := &Systemd{UnitDir: dir, Run: run.run}

		err := s.Uninstall(t.Context())

		if got := errcode.Of(err); got != errcode.StaleState {
			t.Errorf("errcode.Of(Uninstall(foreign)) = %q, want %q (err %v)", got, errcode.StaleState, err)
		}
		got, rerr := os.ReadFile(path)
		if rerr != nil {
			t.Fatalf("read foreign unit after Uninstall: %v", rerr)
		}
		if string(got) != foreignUnit {
			t.Errorf("foreign unit = %q, want unchanged %q", got, foreignUnit)
		}
	})
}
