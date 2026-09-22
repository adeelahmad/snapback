package daemon

import (
	"bytes"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/links"
	"github.com/adeelahmad/snapback/internal/rawpath"
)

// logBuffer returns a logger writing text records into buf, with the time
// attribute removed so assertions stay stable.
func logBuffer() (*slog.Logger, *bytes.Buffer) {
	var buf bytes.Buffer
	h := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
	})
	return slog.New(h), &buf
}

// linesAt returns the captured lines logged at the given level.
func linesAt(buf *bytes.Buffer, level slog.Level) []string {
	var out []string
	for _, line := range strings.Split(buf.String(), "\n") {
		if strings.Contains(line, "level="+level.String()) {
			out = append(out, line)
		}
	}
	return out
}

// TestNewNilLogDiscards checks that a Deps without a logger still builds and
// runs a daemon: the daemon discards its log instead of panicking.
func TestNewNilLogDiscards(t *testing.T) {
	h := newHarness(t)
	h.cfg.Catalog.RefreshInterval = time.Hour
	h.deps.Log = nil

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("New(cfg, Deps{Log: nil}) panicked with %v, want no panic", r)
		}
	}()
	d := New(h.cfg, h.deps)
	errc := start(t, d)
	if got := awaitRunning(t, d, errc); got != "ready" {
		t.Errorf("Status().State with a nil Log = %q, want %q", got, "ready")
	}
}

// TestRefreshLogsOneInfoLine pins the single operational line the daemon
// logs after a refresh: generation, per-repository state, the eligible
// directory count against the linked total, and the failed repositories.
func TestRefreshLogsOneInfoLine(t *testing.T) {
	h := newHarness(t)
	h.cfg.Catalog.RefreshInterval = time.Hour
	log, buf := logBuffer()
	h.deps.Log = log
	h.ref.result = RefreshResult{
		Generation:    7,
		At:            fixedNow,
		Failed:        []string{"repoB"},
		EligibleCount: map[string]int{"/work/one": 3, "/work/two": 0},
	}
	h.linker.records = []links.Record{
		{Key: "one", Dir: rawpath.Path("/work/one"), State: links.StateOwned},
		{Key: "two", Dir: rawpath.Path("/work/two"), State: links.StateOwned},
	}

	d := New(h.cfg, h.deps)
	errc := start(t, d)
	awaitRunning(t, d, errc)

	got := linesAt(buf, slog.LevelInfo)
	if len(got) != 1 {
		t.Fatalf("Info lines after one refresh = %d (%q), want 1", len(got), buf.String())
	}
	line := got[0]
	for _, want := range []string{
		"generation=7",
		"repoA=ready",
		"repoB=failed",
		"eligible=1/2",
		"failed=repoB",
	} {
		if !strings.Contains(line, want) {
			t.Errorf("refresh log line = %q, want it to contain %q", line, want)
		}
	}
}

// TestMountFailureLogsErrorLine pins the one Error line the daemon logs when
// a repository's mount fails: it names the repository and the error text.
func TestMountFailureLogsErrorLine(t *testing.T) {
	h, _ := perRepoHarness(t)
	h.cfg.Catalog.RefreshInterval = time.Hour
	log, buf := logBuffer()
	h.deps.Log = log
	h.sup.startErr = errcode.New(errcode.MountFailure, "mount repoA", errors.New("mountpoint unusable"))

	d := New(h.cfg, h.deps)
	errc := start(t, d)
	awaitRunning(t, d, errc)

	got := linesAt(buf, slog.LevelError)
	if len(got) != 1 {
		t.Fatalf("Error lines after a mount failure = %d (%q), want 1", len(got), buf.String())
	}
	for _, want := range []string{"repoA", "mountpoint unusable"} {
		if !strings.Contains(got[0], want) {
			t.Errorf("mount failure log line = %q, want it to contain %q", got[0], want)
		}
	}
}
