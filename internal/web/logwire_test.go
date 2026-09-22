package web

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// logWireWait bounds how long a test waits, while the server is up, for the
// record the command is supposed to write.
const logWireWait = 2 * time.Second

// logWireLines splits s into its non-blank lines.
func logWireLines(s string) []string {
	var out []string
	for _, line := range strings.Split(s, "\n") {
		if strings.TrimSpace(line) != "" {
			out = append(out, line)
		}
	}
	return out
}

// waitFor polls want until it reports true or logWireWait passes.
func waitFor(want func() bool) {
	deadline := time.Now().Add(logWireWait)
	for time.Now().Before(deadline) {
		if want() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// TestLogWireWebJSONFileSink pins that `snapback web` carries the same three
// log flags as the daemon and that the logger they resolve to is the one the
// server writes through: every record lands in --log-file as JSON, and the
// server's own startup record is among them.
func TestLogWireWebJSONFileSink(t *testing.T) {
	forceGOOS(t, "linux")
	r := newCmdRun(t, map[string]string{})
	logPath := filepath.Join(t.TempDir(), "w.log")
	args := []string{"--log-level", "debug", "--log-format", "json", "--log-file", logPath}

	code := r.run(Command(), args, func(string) {
		waitFor(func() bool {
			b, err := os.ReadFile(logPath)
			return err == nil && strings.Contains(string(b), `"msg":"web listening"`)
		})
	})

	if code != 0 {
		t.Errorf("Command().Run(%v) = %d, want 0 (stderr %q)", args, code, r.stderr.String())
	}
	b, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) = %v, want the log file written by --log-file", logPath, err)
	}
	lines := logWireLines(string(b))
	if len(lines) == 0 {
		t.Fatalf("log file %q = %q, want at least one JSON record", logPath, b)
	}
	var sawListening bool
	for _, line := range lines {
		var rec map[string]any
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			t.Errorf("json.Unmarshal(log line %q) = %v, want a JSON record", line, err)
			continue
		}
		if level, _ := rec["level"].(string); level == "" {
			t.Errorf("log line %q has level %v, want a non-empty level", line, rec["level"])
		}
		msg, _ := rec["msg"].(string)
		if msg == "" {
			t.Errorf("log line %q has msg %v, want a non-empty msg", line, rec["msg"])
		}
		if msg != "web listening" {
			continue
		}
		sawListening = true
		if addr, _ := rec["addr"].(string); addr == "" {
			t.Errorf("record %q has addr %v, want the listen address it bound", line, rec["addr"])
		}
	}
	if !sawListening {
		t.Errorf("log file %q = %q, want the server's %q record", logPath, b, "web listening")
	}
}

// TestLogWireWebDefaultsToTextOnStderr pins the default sink: with no log
// flags the server's records stay on stderr in slog's text format at info.
func TestLogWireWebDefaultsToTextOnStderr(t *testing.T) {
	forceGOOS(t, "linux")
	r := newCmdRun(t, map[string]string{})

	code := r.run(Command(), nil, func(string) {
		waitFor(func() bool { return strings.Contains(r.stderr.String(), `msg="web listening"`) })
	})

	if code != 0 {
		t.Errorf("Command().Run(nil) = %d, want 0 (stderr %q)", code, r.stderr.String())
	}
	for _, want := range []string{"level=INFO", `msg="web listening"`} {
		if got := r.stderr.String(); !strings.Contains(got, want) {
			t.Errorf("Command().Run(nil) stderr = %q, want it to contain %q", got, want)
		}
	}
}

// TestLogWireWebBadLevelIsAUsageError pins that a bad --log-level is a usage
// error, reported before the listener opens: exit 2 and no web.url.
func TestLogWireWebBadLevelIsAUsageError(t *testing.T) {
	forceGOOS(t, "linux")
	r := newCmdRun(t, map[string]string{})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []string{"--log-level=verbose"}
	code := Command().Run(ctx, r.env, args)

	if code != 2 {
		t.Errorf("Command().Run(%v) = %d, want 2 (usage error)", args, code)
	}
	if got := r.stderr.String(); !strings.Contains(got, "--log-level") {
		t.Errorf("Command().Run(%v) stderr = %q, want it to name %q", args, got, "--log-level")
	}
	if r.served() {
		t.Errorf("Command().Run(%v) opened a listener (stdout %q), want none: a bad value is caught before the listener opens",
			args, r.stdout.String())
	}
}

// TestLogWireWebHelpListsLogFlags pins that `snapback web -h` documents the
// three log flags.
func TestLogWireWebHelpListsLogFlags(t *testing.T) {
	r := newCmdRun(t, map[string]string{})

	if code := Command().Run(t.Context(), r.env, []string{"-h"}); code != 0 {
		t.Errorf("Command().Run([-h]) = %d, want 0", code)
	}
	out := r.stderr.String()
	for _, tt := range []struct{ flag, want string }{
		{"--log-level", "debug"},
		{"--log-format", "json"},
		{"--log-file", "file"},
	} {
		usage := flagUsage(out, tt.flag)
		if usage == "" {
			t.Errorf("web -h usage for %s = %q, want a non-empty description\nfull usage:\n%s", tt.flag, usage, out)
			continue
		}
		if !strings.Contains(usage, tt.want) {
			t.Errorf("web -h usage for %s = %q, want it to mention %q", tt.flag, usage, tt.want)
		}
	}
}
