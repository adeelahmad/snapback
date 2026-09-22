package daemon

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/cli"
	"github.com/adeelahmad/snapback/internal/config"
)

// logWireWait bounds how long a test waits for the daemon's first log record
// before it cancels the run and asserts on what was written.
const logWireWait = 2 * time.Second

// syncBuf is a bytes.Buffer safe for a test goroutine to poll while the
// daemon's logger writes into it.
type syncBuf struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuf) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuf) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

// logWireEnv returns a cli.Env whose streams are safe to read while the
// command runs, with configPath as the default config file.
func logWireEnv(configPath string) (cli.Env, *syncBuf, *syncBuf) {
	stdout, stderr := &syncBuf{}, &syncBuf{}
	env := cli.Env{
		Stdout:     stdout,
		Stderr:     stderr,
		Getenv:     func(string) string { return "" },
		ConfigPath: configPath,
	}
	return env, stdout, stderr
}

// runLogWire runs "snapback run" with args against a fake builder serving the
// harness deps, cancels the run as soon as ready reports a log record landed
// (or after logWireWait), and returns the exit code with both streams. ready
// is given everything written to stderr so far.
func runLogWire(t *testing.T, cfgPath string, args []string, ready func(stderr string) bool) (code int, stdout, stderr string) {
	t.Helper()
	h := newHarness(t)
	build := func(_ context.Context, _ *config.Config, ln net.Listener, _ *slog.Logger) (Deps, error) {
		deps := h.deps
		deps.Listener = ln
		deps.Trace = nil
		return deps, nil
	}
	env, outBuf, errBuf := logWireEnv(cfgPath)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		deadline := time.Now().Add(logWireWait)
		for time.Now().Before(deadline) {
			if ready(errBuf.String()) {
				break
			}
			time.Sleep(10 * time.Millisecond)
		}
		cancel()
	}()

	code = runWithin(t, logWireWait+5*time.Second, func() int { return Command(build).Run(ctx, env, args) })
	return code, outBuf.String(), errBuf.String()
}

// TestLogWireJSONFileSink pins that --log-level, --log-format and --log-file
// send the daemon's own records to that file as JSON, and that nothing is
// left on stderr once a log file is given.
func TestLogWireJSONFileSink(t *testing.T) {
	cfgPath, _ := writeValidConfig(t)
	logPath := filepath.Join(t.TempDir(), "d.log")
	args := []string{"--log-level", "debug", "--log-format", "json", "--log-file", logPath}

	code, stdout, stderr := runLogWire(t, cfgPath, args, func(string) bool {
		b, err := os.ReadFile(logPath)
		return err == nil && bytes.Contains(b, []byte(`"msg":"refresh"`))
	})

	if code != 0 {
		t.Errorf("run %v = %d, want 0 (stderr %q)", args, code, stderr)
	}
	if stdout != "" {
		t.Errorf("run %v stdout = %q, want empty", args, stdout)
	}
	if stderr != "" {
		t.Errorf("run %v stderr = %q, want empty: --log-file takes every record", args, stderr)
	}
	b, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) = %v, want the log file written by --log-file", logPath, err)
	}
	lines := nonEmptyLines(string(b))
	if len(lines) == 0 {
		t.Fatalf("log file %q = %q, want at least one JSON record", logPath, b)
	}
	var sawRefresh bool
	for _, line := range lines {
		var rec map[string]any
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			t.Errorf("json.Unmarshal(log line %q) = %v, want a JSON record", line, err)
			continue
		}
		level, _ := rec["level"].(string)
		if level == "" {
			t.Errorf("log line %q has level %v, want a non-empty level", line, rec["level"])
		}
		msg, _ := rec["msg"].(string)
		if msg == "" {
			t.Errorf("log line %q has msg %v, want a non-empty msg", line, rec["msg"])
		}
		if msg == "refresh" {
			sawRefresh = true
		}
	}
	if !sawRefresh {
		t.Errorf("log file %q = %q, want the daemon's %q record", logPath, b, "refresh")
	}
}

// TestLogWireDefaultsToTextOnStderr pins today's default (LOG-1): with no log
// flags the daemon's records stay on stderr in slog's text format at info.
func TestLogWireDefaultsToTextOnStderr(t *testing.T) {
	cfgPath, _ := writeValidConfig(t)

	code, stdout, stderr := runLogWire(t, cfgPath, nil, func(stderr string) bool {
		return strings.Contains(stderr, "msg=refresh")
	})

	if code != 0 {
		t.Errorf("run (no log flags) = %d, want 0 (stderr %q)", code, stderr)
	}
	if stdout != "" {
		t.Errorf("run (no log flags) stdout = %q, want empty", stdout)
	}
	for _, want := range []string{"level=INFO", "msg=refresh"} {
		if !strings.Contains(stderr, want) {
			t.Errorf("run (no log flags) stderr = %q, want it to contain %q", stderr, want)
		}
	}
}

// TestLogWireBadLevelIsAUsageError pins that a bad --log-level is rejected as
// a usage error before the daemon is built: no builder call, no lock file.
func TestLogWireBadLevelIsAUsageError(t *testing.T) {
	cfgPath, stateDir := writeValidConfig(t)
	var calls int
	build := func(context.Context, *config.Config, net.Listener, *slog.Logger) (Deps, error) {
		calls++
		return Deps{}, errors.New("builder must not run on a bad --log-level")
	}
	env, stdout, stderr := logWireEnv(cfgPath)

	args := []string{"--log-level=verbose"}
	code := runWithin(t, 5*time.Second, func() int {
		return Command(build).Run(context.Background(), env, args)
	})

	if code != 2 {
		t.Errorf("run %v = %d, want 2 (usage error)", args, code)
	}
	if calls != 0 {
		t.Errorf("Builder calls on %v = %d, want 0", args, calls)
	}
	if got := stdout.String(); got != "" {
		t.Errorf("run %v stdout = %q, want empty", args, got)
	}
	if got := stderr.String(); !strings.Contains(got, "--log-level") {
		t.Errorf("run %v stderr = %q, want it to name %q", args, got, "--log-level")
	}
	lock := filepath.Join(stateDir, "daemon.lock")
	if _, err := os.Stat(lock); err == nil {
		t.Errorf("os.Stat(%q) = nil, want not exist: no lock is taken on a usage error", lock)
	}
}

// TestLogWireHelpListsLogFlags pins that "snapback run -h" documents the three
// log flags.
func TestLogWireHelpListsLogFlags(t *testing.T) {
	env, stdout, stderr := logWireEnv("")

	code := runWithin(t, 2*time.Second, func() int {
		return Command(unusedBuilder(t)).Run(context.Background(), env, []string{"-h"})
	})

	if code != 0 {
		t.Errorf("run -h = %d, want 0", code)
	}
	if got := stdout.String(); got != "" {
		t.Errorf("run -h stdout = %q, want empty", got)
	}
	for _, want := range []string{"-log-level", "-log-format", "-log-file"} {
		if got := stderr.String(); !strings.Contains(got, want) {
			t.Errorf("run -h stderr = %q, want it to contain %q", got, want)
		}
	}
}

// nonEmptyLines splits s into its non-blank lines.
func nonEmptyLines(s string) []string {
	var out []string
	for _, line := range strings.Split(s, "\n") {
		if strings.TrimSpace(line) != "" {
			out = append(out, line)
		}
	}
	return out
}
