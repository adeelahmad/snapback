package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/compat/latency"
)

const (
	wantRemote     = "gdrive:snapback-stage1"
	testSnapshotID = "4f1d2c3b4a5968778695a4b3c2d1e0f00112233445566778899aabbccddeeff"
)

var resticValueFlags = map[string]bool{"-r": true, "--repo": true, "--password-file": true, "--cache-dir": true}

type reply struct {
	out []byte
	err error
}

// depsFake is a scripted latency Runner and Mounter; it never starts a process.
type depsFake struct {
	mu     sync.Mutex
	calls  []string
	script map[string][]reply
	seen   map[string]int
}

func newDepsFake() *depsFake {
	return &depsFake{script: map[string][]reply{}, seen: map[string]int{}}
}

func labelOf(name string, args []string) string {
	switch name {
	case "rclone":
		if len(args) > 0 {
			return "rclone " + args[0]
		}
	case "restic":
		for i := 0; i < len(args); i++ {
			if resticValueFlags[args[i]] {
				i++
				continue
			}
			if !strings.HasPrefix(args[i], "-") {
				return "restic " + args[i]
			}
		}
	}
	return name
}

func defaultReply(label string) reply {
	switch label {
	case "rclone lsf":
		return reply{err: errors.New("exit status 3: directory not found")}
	case "restic version":
		return reply{out: []byte("restic 0.19.0 compiled with go1.25.1 on darwin/arm64\n")}
	case "rclone version":
		return reply{out: []byte("v1.71.0\n")}
	case "restic snapshots":
		return reply{out: fmt.Appendf(nil, `[{"id":%q,"time":"2026-09-22T10:00:00Z","hostname":"h","paths":["/d"]}]`, testSnapshotID)}
	case "list":
		return reply{out: []byte("file-000\nfile-001\n")}
	case "read":
		return reply{out: []byte("generated")}
	}
	return reply{}
}

func (f *depsFake) record(label string) ([]byte, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, label)
	n := f.seen[label]
	f.seen[label] = n + 1
	replies := f.script[label]
	if len(replies) == 0 {
		r := defaultReply(label)
		return r.out, r.err
	}
	r := replies[min(n, len(replies)-1)]
	return r.out, r.err
}

func (f *depsFake) Run(_ context.Context, name string, args []string) ([]byte, error) {
	return f.record(labelOf(name, args))
}

func (f *depsFake) Mount(_ context.Context, name string, args []string) (latency.Mounted, error) {
	if _, err := f.record(labelOf(name, args)); err != nil {
		return nil, err
	}
	return fakeMounted{f: f}, nil
}

func (f *depsFake) callCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.calls)
}

type fakeMounted struct{ f *depsFake }

func (m fakeMounted) List(_ context.Context, _ string) ([]string, error) {
	out, err := m.f.record("list")
	if err != nil {
		return nil, err
	}
	return strings.Fields(string(out)), nil
}

func (m fakeMounted) Read(_ context.Context, _ string) ([]byte, error) {
	return m.f.record("read")
}

func (m fakeMounted) Unmount(_ context.Context) error {
	_, err := m.f.record("unmount")
	return err
}

type stepClock struct {
	mu sync.Mutex
	t  time.Time
}

func (c *stepClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.t = c.t.Add(7 * time.Millisecond)
	return c.t
}

// useFake swaps newConfig for one wired to f and restores it after the test.
func useFake(t *testing.T, f *depsFake) {
	t.Helper()
	orig := newConfig
	newConfig = func(remote string) latency.Config {
		return latency.Config{
			Remote:  remote,
			Runner:  f,
			Mounter: f,
			Clock:   &stepClock{t: time.Date(2026, 9, 22, 10, 0, 0, 0, time.UTC)},
		}
	}
	t.Cleanup(func() { newConfig = orig })
}

func envWith(remote string) func(string) string {
	return func(key string) string {
		if key == "SNAPBACK_RCLONE_REMOTE" {
			return remote
		}
		return ""
	}
}

func readResult(t *testing.T, path string) latency.Result {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var res latency.Result
	if err := json.Unmarshal(b, &res); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	return res
}

func TestRunMissingOutFlagExits2(t *testing.T) {
	f := newDepsFake()
	useFake(t, f)
	var stdout, stderr bytes.Buffer

	code := run([]string{}, envWith(wantRemote), &stdout, &stderr)

	if code != 2 {
		t.Errorf("run exit = %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "-out") {
		t.Errorf("stderr %q does not mention -out", stderr.String())
	}
	if n := f.callCount(); n != 0 {
		t.Errorf("recorded %d runner calls, want 0", n)
	}
}

func TestRunRefusedRemoteExits2(t *testing.T) {
	for _, remote := range []string{"gdrive:other", ""} {
		t.Run(fmt.Sprintf("remote=%q", remote), func(t *testing.T) {
			f := newDepsFake()
			useFake(t, f)
			out := filepath.Join(t.TempDir(), "latency.json")
			var stdout, stderr bytes.Buffer

			code := run([]string{"-out", out}, envWith(remote), &stdout, &stderr)

			if code != 2 {
				t.Errorf("run exit = %d, want 2", code)
			}
			if !strings.Contains(stderr.String(), wantRemote) {
				t.Errorf("stderr %q does not name %q", stderr.String(), wantRemote)
			}
			if _, err := os.Stat(out); !os.IsNotExist(err) {
				t.Errorf("file at -out %s exists after refusal (stat err = %v)", out, err)
			}
			if n := f.callCount(); n != 0 {
				t.Errorf("recorded %d runner calls, want 0", n)
			}
		})
	}
}

func TestRunWritesJSONAndExits0(t *testing.T) {
	f := newDepsFake()
	useFake(t, f)
	out := filepath.Join(t.TempDir(), "latency.json")
	var stdout, stderr bytes.Buffer

	code := run([]string{"-out", out}, envWith(wantRemote), &stdout, &stderr)

	if code != 0 {
		t.Errorf("run exit = %d, want 0; stderr = %q", code, stderr.String())
	}
	res := readResult(t, out)
	if len(res.Measurements) != 4 {
		t.Errorf("result has %d measurements, want 4: %v", len(res.Measurements), res.Measurements)
	}
	for _, k := range []string{"cold_listing", "warm_prewarmed_listing", "cold_first_file_read", "warm_listing_after_restart"} {
		if _, ok := res.Measurements[k]; !ok {
			t.Errorf("measurement %q missing", k)
		}
	}
	if !res.RemoteDeleted {
		t.Error("remote_deleted = false, want true")
	}
	if !strings.Contains(stdout.String(), out) {
		t.Errorf("stdout %q does not print the output path %q", stdout.String(), out)
	}
}

func TestRunRemoteNotDeletedExits1Loudly(t *testing.T) {
	f := newDepsFake()
	f.script["rclone lsf"] = []reply{
		{err: errors.New("exit status 3: directory not found")},
		{out: []byte("data/\n")},
	}
	useFake(t, f)
	out := filepath.Join(t.TempDir(), "latency.json")
	var stdout, stderr bytes.Buffer

	code := run([]string{"-out", out}, envWith(wantRemote), &stdout, &stderr)

	if code != 1 {
		t.Errorf("run exit = %d, want 1", code)
	}
	for _, want := range []string{"NOT DELETED", wantRemote} {
		if !strings.Contains(stderr.String(), want) {
			t.Errorf("stderr %q does not contain %q", stderr.String(), want)
		}
	}
	b, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("JSON not written at -out: %v", err)
	}
	var raw map[string]any
	if err := json.Unmarshal(b, &raw); err != nil {
		t.Fatalf("parse %s: %v", out, err)
	}
	v, ok := raw["remote_deleted"]
	if !ok {
		t.Fatal("JSON lacks remote_deleted")
	}
	if v != false {
		t.Errorf("remote_deleted = %v, want false", v)
	}
}

func TestMainIsThin(t *testing.T) {
	b, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("read main.go: %v", err)
	}
	src := string(b)
	if !strings.Contains(src, "os.Exit(run(") {
		t.Errorf("main.go does not contain os.Exit(run(:\n%s", src)
	}
	nonEmpty := 0
	for line := range strings.SplitSeq(src, "\n") {
		if strings.TrimSpace(line) != "" {
			nonEmpty++
		}
	}
	if nonEmpty > 12 {
		t.Errorf("main.go has %d non-empty lines, want <= 12", nonEmpty)
	}
}
