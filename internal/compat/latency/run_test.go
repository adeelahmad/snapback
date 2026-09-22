package latency

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"
)

const testSnapshotID = "4f1d2c3b4a5968778695a4b3c2d1e0f00112233445566778899aabbccddeeff"

// resticValueFlags are restic global flags that consume the following argument.
var resticValueFlags = map[string]bool{"-r": true, "--repo": true, "--password-file": true, "--cache-dir": true}

type runCall struct {
	label string
	name  string
	args  []string
}

// runFake is a scripted Runner and Mounter that records every operation in order.
type runFake struct {
	mu     sync.Mutex
	calls  []runCall
	script map[string][]fakeReply
	seen   map[string]int
	onCall func(runCall)
}

func newRunFake() *runFake {
	return &runFake{script: map[string][]fakeReply{}, seen: map[string]int{}}
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

func defaultReply(label string) fakeReply {
	switch label {
	case "rclone lsf":
		return fakeReply{err: errors.New("rclone: exit status 3: directory not found")}
	case "restic version":
		return fakeReply{out: []byte("restic 0.19.0 compiled with go1.25.1 on darwin/arm64\n")}
	case "rclone version":
		return fakeReply{out: []byte("rclone v1.71.0\n- os/version: darwin\n")}
	case "restic snapshots":
		return fakeReply{out: fmt.Appendf(nil, `[{"id":%q,"time":"2026-09-22T10:00:00Z","hostname":"h","paths":["/d"]}]`, testSnapshotID)}
	case "list":
		return fakeReply{out: []byte("file-000\nfile-001\n")}
	case "read":
		return fakeReply{out: []byte("generated")}
	}
	return fakeReply{}
}

func (f *runFake) record(label, name string, args []string) ([]byte, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	c := runCall{label: label, name: name, args: append([]string(nil), args...)}
	f.calls = append(f.calls, c)
	if f.onCall != nil {
		f.onCall(c)
	}
	n := f.seen[label]
	f.seen[label] = n + 1
	replies := f.script[label]
	if len(replies) == 0 {
		rep := defaultReply(label)
		return rep.out, rep.err
	}
	rep := replies[min(n, len(replies)-1)]
	return rep.out, rep.err
}

func (f *runFake) Run(_ context.Context, name string, args []string) ([]byte, error) {
	return f.record(labelOf(name, args), name, args)
}

func (f *runFake) Mount(_ context.Context, name string, args []string) (Mounted, error) {
	if _, err := f.record(labelOf(name, args), name, args); err != nil {
		return nil, err
	}
	return fakeMounted{f: f}, nil
}

func (f *runFake) labels() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]string, len(f.calls))
	for i, c := range f.calls {
		out[i] = c.label
	}
	return out
}

func (f *runFake) snapshot() []runCall {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Clone(f.calls)
}

type fakeMounted struct{ f *runFake }

func (m fakeMounted) List(_ context.Context, dir string) ([]string, error) {
	out, err := m.f.record("list", "", []string{dir})
	if err != nil {
		return nil, err
	}
	return strings.Fields(string(out)), nil
}

func (m fakeMounted) Read(_ context.Context, path string) ([]byte, error) {
	return m.f.record("read", "", []string{path})
}

func (m fakeMounted) Unmount(_ context.Context) error {
	_, err := m.f.record("unmount", "", nil)
	return err
}

// stepClock advances a fixed step on every Now call.
type stepClock struct {
	mu   sync.Mutex
	t    time.Time
	step time.Duration
}

func (c *stepClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.t = c.t.Add(c.step)
	return c.t
}

func newConfig(f *runFake) Config {
	return Config{
		Remote:  wantRemote,
		Runner:  f,
		Mounter: f,
		Clock:   &stepClock{t: time.Date(2026, 9, 22, 10, 0, 0, 0, time.UTC), step: 7 * time.Millisecond},
	}
}

func flagValue(args []string, flag string) (string, bool) {
	i := slices.Index(args, flag)
	if i < 0 || i+1 >= len(args) {
		return "", false
	}
	return args[i+1], true
}

// repoCalls returns the restic calls that open the repository (every restic call except version).
func repoCalls(calls []runCall) []runCall {
	var out []runCall
	for _, c := range calls {
		if c.name == "restic" && c.label != "restic version" {
			out = append(out, c)
		}
	}
	return out
}

func indexOfLabel(labels []string, label string) int {
	return slices.Index(labels, label)
}

func repeatLabel(label string, n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = label
	}
	return out
}

func TestRunCommandSequence(t *testing.T) {
	f := newRunFake()

	_, err := Run(t.Context(), newConfig(f))

	if err != nil {
		t.Errorf("Run error = %v, want nil", err)
	}
	want := []string{"rclone lsf", "restic version", "rclone version", "restic init", "restic backup", "restic snapshots", "restic mount", "list", "restic ls"}
	want = append(want, repeatLabel("list", Samples)...)
	want = append(want, "read", "unmount", "restic mount")
	want = append(want, repeatLabel("list", Samples)...)
	want = append(want, "unmount", "rclone purge", "rclone lsf")
	if got := f.labels(); !slices.Equal(got, want) {
		t.Errorf("operation order:\n got  %q\n want %q", got, want)
	}
}

func TestRunUsesFreshTempCacheDir(t *testing.T) {
	f := newRunFake()

	if _, err := Run(t.Context(), newConfig(f)); err != nil {
		t.Errorf("Run error = %v, want nil", err)
	}

	calls := repoCalls(f.snapshot())
	if len(calls) == 0 {
		t.Fatal("no restic repository calls recorded")
	}
	var dir string
	mounts := 0
	for _, c := range calls {
		d, ok := flagValue(c.args, "--cache-dir")
		if !ok || d == "" {
			t.Errorf("%s call %q lacks --cache-dir <dir>", c.label, c.args)
			continue
		}
		if dir == "" {
			dir = d
		} else if d != dir {
			t.Errorf("%s uses cache dir %q, want the same %q as earlier calls", c.label, d, dir)
		}
		if c.label == "restic mount" {
			mounts++
		}
	}
	if mounts != 2 {
		t.Errorf("recorded %d mounts with --cache-dir, want 2 sharing one dir", mounts)
	}
	if dir == "" {
		t.Fatal("no cache dir captured")
	}
	if !isUnder(dir, os.TempDir()) {
		t.Errorf("cache dir %q is not under os.TempDir() %q", dir, os.TempDir())
	}
	if userCache, err := os.UserCacheDir(); err == nil && isUnder(dir, userCache) {
		t.Errorf("cache dir %q is the user cache dir %q or below it", dir, userCache)
	}
}

func TestRunPrewarmUsesFullSnapshotID(t *testing.T) {
	f := newRunFake()

	if _, err := Run(t.Context(), newConfig(f)); err != nil {
		t.Errorf("Run error = %v, want nil", err)
	}

	var lsCalls, mountCalls []runCall
	for _, c := range f.snapshot() {
		switch c.label {
		case "restic ls":
			lsCalls = append(lsCalls, c)
		case "restic mount":
			mountCalls = append(mountCalls, c)
		}
	}
	if len(lsCalls) != 1 {
		t.Fatalf("recorded %d restic ls calls, want 1 prewarm", len(lsCalls))
	}
	args := lsCalls[0].args
	i := slices.Index(args, "ls")
	if i < 0 || !slices.Equal(args[i:], []string{"ls", "--json", testSnapshotID}) {
		t.Errorf("prewarm args %q, want to end in ls --json %s", args, testSnapshotID)
	}
	if len(mountCalls) == 0 {
		t.Fatal("no restic mount call recorded")
	}
	for _, c := range mountCalls {
		if v, ok := flagValue(c.args, "--path-template"); !ok || v != "ids/%I" {
			t.Errorf("mount args %q lack --path-template ids/%%I", c.args)
		}
	}
}

func TestRunPasswordNeverInArgsOrResult(t *testing.T) {
	f := newRunFake()
	var password string
	f.onCall = func(c runCall) {
		if password != "" {
			return
		}
		if p, ok := flagValue(c.args, "--password-file"); ok {
			if b, err := os.ReadFile(p); err == nil {
				password = strings.TrimSpace(string(b))
			}
		}
	}

	res, err := Run(t.Context(), newConfig(f))
	if err != nil {
		t.Errorf("Run error = %v, want nil", err)
	}

	if password == "" {
		t.Fatal("password never captured: no call carried a readable --password-file")
	}
	calls := f.snapshot()
	for _, c := range repoCalls(calls) {
		if _, ok := flagValue(c.args, "--password-file"); !ok {
			t.Errorf("%s call %q lacks --password-file", c.label, c.args)
		}
	}
	for _, c := range calls {
		for _, a := range c.args {
			if strings.Contains(a, password) {
				t.Errorf("%s call arg %q contains the password", c.label, a)
			}
		}
	}
	b, err := Encode(res)
	if err != nil {
		t.Fatalf("Encode(result): %v", err)
	}
	if bytes.Contains(b, []byte(password)) {
		t.Error("encoded result contains the password")
	}
}

func TestRunRecordsSamplesFromClock(t *testing.T) {
	f := newRunFake()

	res, err := Run(t.Context(), newConfig(f))
	if err != nil {
		t.Errorf("Run error = %v, want nil", err)
	}

	wantCount := map[string]int{
		"cold_listing":               1,
		"warm_prewarmed_listing":     Samples,
		"cold_first_file_read":       1,
		"warm_listing_after_restart": Samples,
	}
	for _, name := range measurementNames {
		st, ok := res.Measurements[name]
		if !ok {
			t.Errorf("measurement %q missing", name)
			continue
		}
		if len(st.SamplesMS) != wantCount[name] {
			t.Errorf("%s has %d samples, want %d", name, len(st.SamplesMS), wantCount[name])
		}
		durations := make([]time.Duration, len(st.SamplesMS))
		for i, ms := range st.SamplesMS {
			if ms != 7 {
				t.Errorf("%s sample %d = %vms, want 7ms (one clock step)", name, i, ms)
			}
			durations[i] = time.Duration(ms * float64(time.Millisecond))
		}
		if len(durations) == 0 {
			continue
		}
		want, err := Summarize(durations)
		if err != nil {
			t.Fatalf("Summarize(%s): %v", name, err)
		}
		if st.MedianMS != want.MedianMS || st.MinMS != want.MinMS || st.MaxMS != want.MaxMS {
			t.Errorf("%s stats median/min/max = %v/%v/%v, want %v/%v/%v",
				name, st.MedianMS, st.MinMS, st.MaxMS, want.MedianMS, want.MinMS, want.MaxMS)
		}
	}
}

func TestRunCleanupAfterMidRunFailure(t *testing.T) {
	f := newRunFake()
	f.script["restic ls"] = []fakeReply{{err: errors.New("restic ls: exit status 1")}}

	res, err := Run(t.Context(), newConfig(f))

	if err == nil {
		t.Error("Run error = nil, want the prewarm failure")
	}
	labels := f.labels()
	i := indexOfLabel(labels, "restic ls")
	if i < 0 {
		t.Fatalf("prewarm restic ls never called; operations = %q", labels)
	}
	rest := labels[i+1:]
	pos := 0
	for _, want := range []string{"unmount", "rclone purge", "rclone lsf"} {
		j := indexOfLabel(rest[pos:], want)
		if j < 0 {
			t.Errorf("%q not called (in order) after the failed prewarm; after = %q", want, rest)
			continue
		}
		pos += j + 1
	}
	if !res.RemoteDeleted {
		t.Error("Result.RemoteDeleted = false, want true when lsf reports directory not found")
	}
}

func TestRunReportsFailedDelete(t *testing.T) {
	f := newRunFake()
	f.script["rclone lsf"] = []fakeReply{
		{err: errors.New("rclone: exit status 3: directory not found")},
		{out: []byte("data/\n")},
	}

	res, err := Run(t.Context(), newConfig(f))

	if indexOfLabel(f.labels(), "rclone purge") < 0 {
		t.Errorf("purge never called; operations = %q", f.labels())
	}
	if res.RemoteDeleted {
		t.Error("Result.RemoteDeleted = true, want false when lsf still lists data/")
	}
	if err == nil {
		t.Fatal("Run error = nil, want an error reporting the undeleted remote")
	}
	for _, want := range []string{wantRemote, "not deleted"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("Run error %q does not mention %q", err, want)
		}
	}
}

func TestRunAbortsOnNonEmptyRemoteBeforeWrite(t *testing.T) {
	f := newRunFake()
	f.script["rclone lsf"] = []fakeReply{{out: []byte("config\n")}}

	_, err := Run(t.Context(), newConfig(f))

	if err == nil {
		t.Error("Run error = nil, want refusal of a non-empty remote")
	}
	labels := f.labels()
	if len(labels) == 0 || labels[0] != "rclone lsf" {
		t.Fatalf("first operation is not the lsf pre-check; operations = %q", labels)
	}
	for _, banned := range []string{"restic init", "restic backup", "restic mount", "rclone purge"} {
		if slices.Contains(labels, banned) {
			t.Errorf("%q recorded after a non-empty pre-check; operations = %q", banned, labels)
		}
	}
}

func TestRunRefusesOtherRemoteBeforeAnyCommand(t *testing.T) {
	f := newRunFake()
	cfg := newConfig(f)
	cfg.Remote = "gdrive:other"

	_, err := Run(t.Context(), cfg)

	if err == nil {
		t.Fatal("Run error = nil, want refusal of gdrive:other")
	}
	if !strings.Contains(err.Error(), wantRemote) {
		t.Errorf("Run error %q does not name %q", err, wantRemote)
	}
	if calls := f.snapshot(); len(calls) != 0 {
		t.Errorf("recorded %d operations before refusal, want 0: %+v", len(calls), calls)
	}
}

func TestRunRemovesScratch(t *testing.T) {
	cases := []struct {
		name   string
		script map[string][]fakeReply
	}{
		{"success", nil},
		{"failure", map[string][]fakeReply{"restic ls": {{err: errors.New("restic ls: exit status 1")}}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newRunFake()
			for k, v := range tc.script {
				f.script[k] = v
			}
			var pwFile string
			f.onCall = func(c runCall) {
				if p, ok := flagValue(c.args, "--password-file"); ok && pwFile == "" {
					pwFile = p
				}
			}

			_, _ = Run(t.Context(), newConfig(f))

			if pwFile == "" {
				t.Fatal("scratch root never captured: no call carried --password-file")
			}
			root := filepath.Dir(pwFile)
			if isUnder(root, os.TempDir()) && filepath.Clean(root) != filepath.Clean(os.TempDir()) {
				t.Cleanup(func() { _ = os.RemoveAll(root) })
			}
			if _, err := os.Stat(root); !os.IsNotExist(err) {
				t.Errorf("scratch root %q still present after Run (stat err = %v)", root, err)
			}
		})
	}
}
