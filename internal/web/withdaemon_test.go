package web

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// daemonFactory records every newDaemon call and hands out one fakeDaemon, so
// a test can tell "never constructed" from "constructed but never started".
type daemonFactory struct {
	mu       sync.Mutex
	stateDir []string
	daemon   *fakeDaemon
}

// installDaemonFactory swaps the newDaemon seam for a recorder whose daemon
// fails Start with startErr when that is not nil.
func installDaemonFactory(t *testing.T, startErr error) *daemonFactory {
	t.Helper()
	f := &daemonFactory{daemon: &fakeDaemon{startErr: startErr}}
	prev := newDaemon
	newDaemon = func(stateDir string) DaemonControl {
		f.mu.Lock()
		defer f.mu.Unlock()
		f.stateDir = append(f.stateDir, stateDir)
		return f.daemon
	}
	t.Cleanup(func() { newDaemon = prev })
	return f
}

// calls returns the state dirs newDaemon was asked for, plus the Start and
// Stop counts of the daemon it handed out.
func (f *daemonFactory) calls() (dirs []string, starts, stops int) {
	f.mu.Lock()
	dirs = append([]string(nil), f.stateDir...)
	f.mu.Unlock()
	f.daemon.mu.Lock()
	defer f.daemon.mu.Unlock()
	return dirs, f.daemon.starts, f.daemon.stops
}

func TestWebWithDaemonStartsTheDaemonBeforeServing(t *testing.T) {
	f := installDaemonFactory(t, nil)
	r := newCmdRun(t, nil)

	var servingDirs []string
	var servingStarts, servingStops int
	code := r.run(Command(), []string{"--with-daemon"}, func(string) {
		servingDirs, servingStarts, servingStops = f.calls()
	})
	if code != 0 {
		t.Fatalf("Command().Run([--with-daemon]) = %d, want 0 (stderr %q)", code, r.stderr.String())
	}
	if want := []string{r.stateDir}; len(servingDirs) != 1 || servingDirs[0] != want[0] {
		t.Errorf("newDaemon calls once the server accepts = %q, want exactly %q", servingDirs, want)
	}
	if servingStarts != 1 {
		t.Errorf("Start calls once the server accepts = %d, want 1", servingStarts)
	}
	if servingStops != 0 {
		t.Errorf("Stop calls once the server accepts = %d, want 0", servingStops)
	}
}

func TestWebWithDaemonStopsTheDaemonOnContextCancel(t *testing.T) {
	f := installDaemonFactory(t, nil)
	r := newCmdRun(t, nil)

	if code := r.run(Command(), []string{"--with-daemon"}, nil); code != 0 {
		t.Fatalf("Command().Run([--with-daemon]) = %d, want 0 (stderr %q)", code, r.stderr.String())
	}
	_, starts, stops := f.calls()
	if starts != 1 {
		t.Errorf("Start calls after the command returned = %d, want 1", starts)
	}
	if stops != 1 {
		t.Errorf("Stop calls after the command returned = %d, want 1", stops)
	}
}

func TestWebWithoutWithDaemonStartsNoDaemon(t *testing.T) {
	f := installDaemonFactory(t, nil)
	r := newCmdRun(t, nil)

	if code := r.run(Command(), []string{}, nil); code != 0 {
		t.Fatalf("Command().Run([]) = %d, want 0 (stderr %q)", code, r.stderr.String())
	}
	dirs, starts, stops := f.calls()
	if len(dirs) != 0 {
		t.Errorf("newDaemon calls without --with-daemon = %q, want none", dirs)
	}
	if starts != 0 || stops != 0 {
		t.Errorf("Start/Stop calls without --with-daemon = %d/%d, want 0/0", starts, stops)
	}
}

func TestWebHelpDocumentsWithDaemon(t *testing.T) {
	r := newCmdRun(t, nil)

	code := Command().Run(context.Background(), r.env, []string{"-h"})

	if code != 0 {
		t.Errorf("Command().Run([-h]) = %d, want 0", code)
	}
	out := r.stdout.String() + r.stderr.String()
	if !strings.Contains(out, "-with-daemon") {
		t.Fatalf("Command().Run([-h]) output = %q, want it to name -with-daemon", out)
	}
	desc := flagDescription(out, "-with-daemon")
	if desc == "" || !strings.Contains(desc, "daemon") {
		t.Errorf("-with-daemon usage line = %q, want a description saying it runs a daemon (output %q)", desc, out)
	}
}

// flagDescription returns the trimmed usage line printed under flag in the
// output of flag.FlagSet.PrintDefaults, or "" when there is none.
func flagDescription(out, flag string) string {
	_, rest, ok := strings.Cut(out, flag)
	if !ok {
		return ""
	}
	lines := strings.Split(rest, "\n")
	if len(lines) < 2 {
		return ""
	}
	return strings.TrimSpace(lines[1])
}

func TestWebWithDaemonStartErrorFailsBeforeListening(t *testing.T) {
	startErr := errors.New("daemon already running elsewhere")
	f := installDaemonFactory(t, startErr)
	r := newCmdRun(t, nil)
	ctx, cancel := context.WithTimeout(context.Background(), runTimeout)
	defer cancel()

	start := time.Now()
	code := Command().Run(ctx, r.env, []string{"--with-daemon"})

	if code == 0 {
		t.Errorf("Command().Run([--with-daemon]) with a failing Start = 0, want non-zero")
	}
	if elapsed := time.Since(start); elapsed >= runTimeout {
		t.Fatalf("Command().Run([--with-daemon]) with a failing Start took %v, want it to fail fast", elapsed)
	}
	if !strings.Contains(r.stderr.String(), startErr.Error()) {
		t.Errorf("Command().Run([--with-daemon]) stderr = %q, want it to contain %q", r.stderr.String(), startErr.Error())
	}
	if _, _, stops := f.calls(); stops != 0 {
		t.Errorf("Stop calls after a failing Start = %d, want 0", stops)
	}
	if _, err := os.Stat(filepath.Join(r.stateDir, "web.url")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("os.Stat(web.url) error = %v, want %v: the server must never listen", err, os.ErrNotExist)
	}
}
