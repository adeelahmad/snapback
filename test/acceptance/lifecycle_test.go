//go:build integration

package acceptance

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"
)

// readyCap bounds how long a restarted daemon may take to report ready.
const readyCap = 30 * time.Second

// recoveryLinuxOnly is why the crash-recovery step is skipped off Linux.
const recoveryLinuxOnly = "recovery reads /proc/self/mountinfo; Linux only in v0.1 (macOS follow-up, SPEC §22.1)"

// liveChecksums returns sha256 of every regular file under root, skipping
// .snapshot entries so history content never counts as live data.
func liveChecksums(t *testing.T, root string) map[string]string {
	t.Helper()
	sums := map[string]string{}
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Name() == ".snapshot" {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if !d.Type().IsRegular() {
			return nil
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(b)
		sums[p] = hex.EncodeToString(sum[:])
		return nil
	})
	if err != nil {
		t.Fatalf("checksum live tree %s: %v", root, err)
	}
	return sums
}

// mountsUnder lists mount points at or below dir, from /proc/self/mountinfo
// on Linux and from mount(8) elsewhere.
func mountsUnder(dir string) []string {
	rd := resolve(dir)
	under := func(p string) bool { return strings.HasPrefix(p, rd) || strings.HasPrefix(p, dir) }
	var points []string
	if runtime.GOOS == "linux" {
		b, err := os.ReadFile("/proc/self/mountinfo")
		if err != nil {
			return nil
		}
		for _, line := range strings.Split(string(b), "\n") {
			if f := strings.Fields(line); len(f) > 4 && under(f[4]) {
				points = append(points, f[4])
			}
		}
		return points
	}
	out, err := exec.Command("mount").Output()
	if err != nil {
		return nil
	}
	sc := bufio.NewScanner(bytes.NewReader(out))
	for sc.Scan() {
		if _, rest, ok := strings.Cut(sc.Text(), " on "); ok {
			if p, _, _ := strings.Cut(rest, " ("); under(p) {
				points = append(points, p)
			}
		}
	}
	return points
}

// waitReady polls status until the daemon reports ready or readyCap passes.
func waitReady(t *testing.T, e env) statusEnvelope {
	t.Helper()
	deadline := time.Now().Add(readyCap)
	for {
		s, raw := readStatus(t, e)
		if s.OK && s.Data.State == "ready" {
			return s
		}
		if time.Now().After(deadline) {
			t.Errorf("daemon status = %s, want state ready within %s", raw, readyCap)
			return s
		}
		time.Sleep(statusPoll)
	}
}

func TestAcc15StopCrashRestartUninstall(t *testing.T) {
	recordEvidence(t, "acc-15")
	requireFUSE(t)

	h := newHistRepo(t)
	writeFiles(t, h.proj, map[string]string{"a.txt": "alpha v1\n", "sub/b.txt": "beta v1\n"})
	backup(t, h.fx, "", histHost, time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC), "daily", h.fx.Root)
	writeFiles(t, h.proj, map[string]string{"a.txt": "alpha live\n", "sub/b.txt": "beta live\n"})

	foreignDir := filepath.Join(h.fx.Root, "foreign")
	foreignBody := []byte("foreign bytes\n")
	if err := os.MkdirAll(foreignDir, 0o755); err != nil {
		t.Fatalf("mkdir foreign: %v", err)
	}
	foreignFile := filepath.Join(foreignDir, ".snapshot")
	if err := os.WriteFile(foreignFile, foreignBody, 0o644); err != nil {
		t.Fatalf("write foreign .snapshot: %v", err)
	}

	e := writeHistConfig(t, h)
	state := filepath.Join(e.Root, "state")
	historyMount := filepath.Join(state, "mounts", "history")
	want := liveChecksums(t, h.fx.Root)

	check := func(step string) {
		t.Helper()
		if got := liveChecksums(t, h.fx.Root); !maps.Equal(got, want) {
			t.Errorf("after %s: live checksums = %v, want %v", step, got, want)
		}
		if got, err := os.ReadFile(foreignFile); err != nil || !bytes.Equal(got, foreignBody) {
			t.Errorf("after %s: foreign/.snapshot = %q, %v, want %q", step, got, err, foreignBody)
		}
	}

	d := startDaemon(t, e)
	waitReady(t, e)
	if _, stderr, code := runSnapback(t, e, "link", h.proj); code != 0 {
		t.Errorf("snapback link proj exit = %d, want 0; stderr: %s", code, stderr)
	}
	if owned := ownedLinks(t, h.fx.Root, historyMount); len(owned) == 0 {
		t.Errorf("after link: owned links = none, want proj linked")
	}
	check("start")

	// No CLI shutdown verb exists; SIGTERM drives the same graceful path.
	if err := d.Stop(); err != nil {
		t.Errorf("graceful stop: %v, want clean exit", err)
	}
	if m := mountsUnder(state); len(m) != 0 {
		t.Errorf("after stop: mounts under state = %v, want none", m)
	}
	check("stop")

	d = startDaemon(t, e)
	waitReady(t, e)
	check("restart")

	if err := d.Kill(); err != nil {
		t.Fatalf("kill -9 daemon: %v", err)
	}
	stale := mountsUnder(state)
	check("crash")

	if runtime.GOOS != "linux" {
		t.Logf("skipping step %q: %s", "crash recovery", recoveryLinuxOnly)
		// The SIGKILLed daemon left its mounts; clear them so later steps
		// start clean.
		unmountUnder(state)
		d = startDaemon(t, e)
		waitReady(t, e)
	} else {
		d = startDaemon(t, e)
		s := waitReady(t, e)
		for _, m := range mountsUnder(state) {
			if _, err := os.ReadDir(m); err != nil {
				t.Errorf("after crash restart: mount %s unreadable (%v), want no stale mount", m, err)
			}
		}
		if len(stale) > 0 && (s.Data.Recovery == nil || len(s.Data.Recovery.Unmounted) == 0) {
			t.Errorf("after crash restart: status recovery = %+v, want it to list the stale mounts %v", s.Data.Recovery, stale)
		}
		names, err := listNames(filepath.Join(h.proj, ".snapshot"))
		if err != nil || len(aliasesOnly(names)) == 0 {
			t.Errorf("after crash restart: list proj/.snapshot = %v, %v, want one alias", names, err)
		}
		check("crash restart")
	}

	if err := d.Stop(); err != nil {
		t.Errorf("graceful stop before uninstall: %v", err)
	}
	if reason := serviceManagerMissing(); reason != "" {
		t.Logf("skipping step %q: %s", "service uninstall", reason)
	} else if _, stderr, code := runSnapback(t, e, "service", "uninstall"); code != 0 && !strings.Contains(stderr, "unsupported_service_manager") {
		t.Errorf("snapback service uninstall exit = %d, want 0; stderr: %s", code, stderr)
	}
	if _, stderr, code := runSnapback(t, e, "links", "remove", "--managed"); code != 0 {
		t.Errorf("snapback links remove --managed exit = %d, want 0; stderr: %s", code, stderr)
	}
	if owned := ownedLinks(t, h.fx.Root, historyMount); len(owned) != 0 {
		t.Errorf("after uninstall: owned links = %v, want none", owned)
	}
	if m := mountsUnder(state); len(m) != 0 {
		t.Errorf("after uninstall: mounts under state = %v, want none", slices.Sorted(slices.Values(m)))
	}
	check("uninstall")
}
