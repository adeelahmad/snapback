package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestDaemonBuilderLogsToStderr checks that the production builder wires the
// daemon's logger to the process stderr.
func TestDaemonBuilderLogsToStderr(t *testing.T) {
	tmp := shortTempDir(t)
	bin := filepath.Join(tmp, "bin")
	if err := os.MkdirAll(bin, 0o755); err != nil {
		t.Fatalf("os.MkdirAll(%s) = %v", bin, err)
	}
	restic := filepath.Join(bin, "restic")
	if err := os.WriteFile(restic, []byte(fakeRestic), 0o755); err != nil {
		t.Fatalf("os.WriteFile(%s) = %v", restic, err)
	}
	t.Setenv("PATH", bin)
	cfg := daemonDepsConfig(t, tmp, restic)
	ln := listenUnix(t, tmp)

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() = %v", err)
	}
	saved := os.Stderr
	os.Stderr = w
	deps, buildErr := daemonBuilder(t.Context(), cfg, ln)
	if deps.Log != nil {
		deps.Log.Info("wiring probe", "marker", "snapback-log-wiring")
	}
	os.Stderr = saved
	if err := w.Close(); err != nil {
		t.Fatalf("w.Close() = %v", err)
	}
	out, readErr := io.ReadAll(r)
	if readErr != nil {
		t.Fatalf("io.ReadAll(stderr pipe) = %v", readErr)
	}

	if buildErr != nil {
		t.Fatalf("daemonBuilder(ctx, cfg, ln) = %v, want nil error", buildErr)
	}
	if deps.Log == nil {
		t.Fatalf("daemonBuilder(ctx, cfg, ln).Log = nil, want a logger wired to stderr")
	}
	if got := string(out); !strings.Contains(got, "snapback-log-wiring") {
		t.Errorf("daemonBuilder(ctx, cfg, ln).Log wrote %q to stderr, want it to contain %q",
			got, "snapback-log-wiring")
	}
}
