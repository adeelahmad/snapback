package cli

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/provider"
)

// id1 is a full 64-hex snapshot ID.
var id1 = provider.SnapshotID(strings.Repeat("ab", 32))

// fixtureCfg returns a config with repository r1 and root home at r.
func fixtureCfg(r, s, h, b string) config.Config {
	return config.Config{
		LinkName:        ".snapshot",
		StateDir:        s,
		HistoryMount:    h,
		BackendMountDir: b,
		Repositories:    []config.Repository{{ID: "r1", Repository: "/repo"}},
		Roots:           []config.Root{{ID: "home", LocalPath: r, RepositoryID: "r1"}},
	}
}

// snapRig is a recording Snapper, Daemon and fake clock.
type snapRig struct {
	mu           sync.Mutex
	factoryCalls int
	factoryRepos []string
	snapCalls    int
	reqs         []provider.SnapRequest
	snapID       provider.SnapshotID
	snapErr      error
	submitted    []string
	visible      []bool
	daemonErr    error
	sleeps       []time.Duration
	start, now   time.Time
}

func newSnapRig() *snapRig {
	t0 := time.Date(2026, 9, 22, 10, 0, 0, 0, time.UTC)
	return &snapRig{snapID: id1, start: t0, now: t0}
}

func (r *snapRig) Snap(_ context.Context, req provider.SnapRequest) (provider.SnapshotID, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.snapCalls++
	r.reqs = append(r.reqs, req)
	if r.snapErr != nil {
		return "", r.snapErr
	}
	return r.snapID, nil
}

func (r *snapRig) HistoryAvailable(context.Context) (bool, error) { return true, nil }

func (r *snapRig) SnapSubmitted(_ context.Context, repoID string, id provider.SnapshotID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.submitted = append(r.submitted, repoID+":"+string(id))
	return nil
}

// Visible pops the next scripted answer; once exhausted it keeps the last,
// or false when none were scripted.
func (r *snapRig) Visible(context.Context, provider.SnapshotID) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.visible) == 0 {
		return false, nil
	}
	v := r.visible[0]
	if len(r.visible) > 1 {
		r.visible = r.visible[1:]
	}
	return v, nil
}

func (r *snapRig) elapsed() time.Duration {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.now.Sub(r.start)
}

// snapDeps returns Deps over r loading cfg, with host hostA and cwd /w.
func snapDeps(r *snapRig, cfg config.Config) Deps {
	return Deps{
		LoadConfig: func(string) (config.Config, error) { return cfg, nil },
		NewSnapper: func(_ config.Config, repoID string) (provider.Snapper, error) {
			r.mu.Lock()
			defer r.mu.Unlock()
			r.factoryCalls++
			r.factoryRepos = append(r.factoryRepos, repoID)
			return r, nil
		},
		Daemon: func(context.Context) (Daemon, error) {
			if r.daemonErr != nil {
				return nil, r.daemonErr
			}
			return r, nil
		},
		Getwd:    func() (string, error) { return "/w", nil },
		Hostname: func() (string, error) { return "hostA", nil },
		Sleep: func(_ context.Context, d time.Duration) error {
			r.mu.Lock()
			defer r.mu.Unlock()
			r.sleeps = append(r.sleeps, d)
			r.now = r.now.Add(d)
			return nil
		},
		Now: func() time.Time {
			r.mu.Lock()
			defer r.mu.Unlock()
			return r.now
		},
	}
}

// snapFixture returns a config over a fresh root R and the path R/proj.
func snapFixture(t *testing.T) (cfg config.Config, proj string) {
	t.Helper()
	root := t.TempDir()
	state := t.TempDir()
	cfg = fixtureCfg(root, filepath.Join(state, "state"), filepath.Join(state, "history"), filepath.Join(state, "backend"))
	return cfg, filepath.Join(root, "proj")
}

func runSnap(d Deps, args ...string) (code int, stdout, stderr string) {
	env, out, errb := newEnv(nil)
	code = SnapCommand(d).Run(context.Background(), env, args)
	return code, out.String(), errb.String()
}

func snapData(t *testing.T, stdout string) map[string]any {
	t.Helper()
	e := decodeEnvelope(t, []byte(stdout))
	var m map[string]any
	if err := json.Unmarshal(e.Data, &m); err != nil {
		t.Fatalf("decode data %q: %v", e.Data, err)
	}
	return m
}

func TestSnapBuildsRequest(t *testing.T) {
	cfg, proj := snapFixture(t)
	r := newSnapRig()

	code, _, stderr := runSnap(snapDeps(r, cfg), proj, "--tag", "a", "--tag", "b")

	if code != 0 {
		t.Fatalf("snap %s --tag a --tag b exit = %d, want 0 (stderr %q)", proj, code, stderr)
	}
	if want := []string{"r1"}; !slices.Equal(r.factoryRepos, want) {
		t.Errorf("NewSnapper repos = %q, want %q", r.factoryRepos, want)
	}
	want := provider.SnapRequest{
		Path:     proj,
		Host:     "hostA",
		Tags:     []string{"snapback:adhoc", "a", "b"},
		Excludes: []string{".snapshot", cfg.StateDir, cfg.HistoryMount, cfg.BackendMountDir},
	}
	if len(r.reqs) != 1 {
		t.Fatalf("Snap calls = %d, want 1", len(r.reqs))
	}
	got := r.reqs[0]
	if got.Path != want.Path || got.Host != want.Host || !slices.Equal(got.Tags, want.Tags) || !slices.Equal(got.Excludes, want.Excludes) {
		t.Errorf("SnapRequest = %+v, want %+v", got, want)
	}
}

func TestSnapDefaultPathIsCwd(t *testing.T) {
	cfg, proj := snapFixture(t)
	r := newSnapRig()
	d := snapDeps(r, cfg)
	d.Getwd = func() (string, error) { return proj, nil }

	code, _, stderr := runSnap(d)

	if code != 0 {
		t.Fatalf("snap exit = %d, want 0 (stderr %q)", code, stderr)
	}
	if len(r.reqs) != 1 || r.reqs[0].Path != proj {
		t.Errorf("SnapRequest paths = %+v, want one with Path %q", r.reqs, proj)
	}
}

func TestSnapReportsPending(t *testing.T) {
	cfg, proj := snapFixture(t)
	r := newSnapRig()

	code, stdout, _ := runSnap(snapDeps(r, cfg), "--json", proj)

	if code != 0 {
		t.Fatalf("snap --json exit = %d, want 0 (stdout %q)", code, stdout)
	}
	if want := []string{"r1:" + string(id1)}; !slices.Equal(r.submitted, want) {
		t.Errorf("SnapSubmitted = %q, want %q", r.submitted, want)
	}
	data := snapData(t, stdout)
	if data["id"] != string(id1) || data["state"] != "pending" {
		t.Errorf("snap --json data = %v, want id %q state pending", data, id1)
	}

	h := newSnapRig()
	code, stdout, _ = runSnap(snapDeps(h, cfg), proj)
	if code != 0 {
		t.Fatalf("snap exit = %d, want 0", code)
	}
	if !strings.Contains(stdout, "pending") || !strings.Contains(stdout, string(id1)) {
		t.Errorf("snap stdout = %q, want pending and the full ID %q", stdout, id1)
	}
	if strings.Contains(stdout, "ready") {
		t.Errorf("snap stdout = %q, must not report ready", stdout)
	}
}

func TestSnapWaitPendingThenReady(t *testing.T) {
	cfg, proj := snapFixture(t)
	r := newSnapRig()
	r.visible = []bool{false, false, true}

	code, stdout, _ := runSnap(snapDeps(r, cfg), "--wait", "--json", proj)

	if code != 0 {
		t.Fatalf("snap --wait exit = %d, want 0 (stdout %q)", code, stdout)
	}
	if data := snapData(t, stdout); data["state"] != "ready" {
		t.Errorf("snap --wait state = %v, want ready", data["state"])
	}
	if want := []time.Duration{snapPollInterval, snapPollInterval}; !slices.Equal(r.sleeps, want) {
		t.Errorf("Sleep calls = %v, want %v", r.sleeps, want)
	}
	if got := r.elapsed(); got >= 120*time.Second {
		t.Errorf("fake elapsed = %v, want < 120s", got)
	}
}

func TestSnapWaitTimeout(t *testing.T) {
	cfg, proj := snapFixture(t)
	r := newSnapRig()

	code, stdout, _ := runSnap(snapDeps(r, cfg), "--wait", "--timeout", "10s", "--json", proj)

	if code != 1 {
		t.Fatalf("snap --wait --timeout 10s exit = %d, want 1 (stdout %q)", code, stdout)
	}
	e := decodeEnvelope(t, []byte(stdout))
	if e.Code != string(errcode.StaleState) {
		t.Errorf("code = %q, want %q", e.Code, errcode.StaleState)
	}
	if !strings.Contains(e.Error, string(id1)) {
		t.Errorf("error = %q, want it to contain %q", e.Error, id1)
	}
	if len(r.sleeps) != 5 {
		t.Errorf("Sleep count = %d, want 5", len(r.sleeps))
	}
	if strings.Contains(stdout, "ready") {
		t.Errorf("stdout = %q, must not report ready", stdout)
	}
}

func TestSnapOutsideRootRefused(t *testing.T) {
	cfg, _ := snapFixture(t)
	outside := t.TempDir()

	probe := newSnapRig()
	s, err := snapDeps(probe, cfg).NewSnapper(cfg, "r1")
	if err != nil {
		t.Fatalf("probe NewSnapper: %v", err)
	}
	if _, err := s.Snap(context.Background(), provider.SnapRequest{}); err != nil {
		t.Fatalf("probe Snap: %v", err)
	}
	if probe.factoryCalls != 1 || probe.snapCalls != 1 {
		t.Fatalf("probe counts = factory %d snap %d, want 1 and 1 (fake not wired)", probe.factoryCalls, probe.snapCalls)
	}

	tests := []struct {
		name     string
		flags    []string
		wantExit int
		wantCode errcode.Code
	}{
		{name: "no flags", wantExit: 1, wantCode: errcode.MappingAbsent},
		{name: "repository only", flags: []string{"--repository", "r1"}, wantExit: 2},
		{name: "prefix only", flags: []string{"--prefix", "/x"}, wantExit: 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newSnapRig()
			args := append([]string{"--json", outside}, tt.flags...)

			code, stdout, _ := runSnap(snapDeps(r, cfg), args...)

			if code != tt.wantExit {
				t.Fatalf("snap %q exit = %d, want %d (stdout %q)", args, code, tt.wantExit, stdout)
			}
			if tt.wantCode != "" {
				if e := decodeEnvelope(t, []byte(stdout)); e.Code != string(tt.wantCode) {
					t.Errorf("snap %q code = %q, want %q", args, e.Code, tt.wantCode)
				}
			}
			if r.factoryCalls != 0 || r.snapCalls != 0 {
				t.Errorf("snap %q counts = factory %d snap %d, want 0 and 0", args, r.factoryCalls, r.snapCalls)
			}
		})
	}
}

func TestSnapExplicitRepositoryOutsideRoot(t *testing.T) {
	cfg, _ := snapFixture(t)

	t.Run("known repository", func(t *testing.T) {
		r := newSnapRig()
		args := []string{"--json", "/elsewhere", "--repository", "r1", "--prefix", "/srv/x"}

		code, stdout, _ := runSnap(snapDeps(r, cfg), args...)

		if code != 0 {
			t.Fatalf("snap %q exit = %d, want 0 (stdout %q)", args, code, stdout)
		}
		if r.snapCalls != 1 || !slices.Equal(r.factoryRepos, []string{"r1"}) {
			t.Errorf("snap %q: Snap calls %d on %q, want 1 on [r1]", args, r.snapCalls, r.factoryRepos)
		}
		if len(r.submitted) != 0 {
			t.Errorf("snap %q SnapSubmitted = %q, want none", args, r.submitted)
		}
		if data := snapData(t, stdout); data["browsable"] != false {
			t.Errorf("snap %q browsable = %v, want false", args, data["browsable"])
		}
	})

	t.Run("unknown repository", func(t *testing.T) {
		r := newSnapRig()
		args := []string{"--json", "/elsewhere", "--repository", "nope", "--prefix", "/srv/x"}

		code, stdout, _ := runSnap(snapDeps(r, cfg), args...)

		if code != 1 {
			t.Fatalf("snap %q exit = %d, want 1 (stdout %q)", args, code, stdout)
		}
		if e := decodeEnvelope(t, []byte(stdout)); e.Code != string(errcode.InvalidConfig) {
			t.Errorf("snap %q code = %q, want %q", args, e.Code, errcode.InvalidConfig)
		}
		if r.snapCalls != 0 {
			t.Errorf("snap %q Snap calls = %d, want 0", args, r.snapCalls)
		}
	})
}

func TestSnapDaemonUnreachable(t *testing.T) {
	cfg, proj := snapFixture(t)

	t.Run("no wait", func(t *testing.T) {
		r := newSnapRig()
		r.daemonErr = errors.New("dial unix: connection refused")

		code, stdout, stderr := runSnap(snapDeps(r, cfg), "--json", proj)

		if code != 0 {
			t.Fatalf("snap --json exit = %d, want 0 (stdout %q)", code, stdout)
		}
		if data := snapData(t, stdout); data["state"] != "pending" {
			t.Errorf("state = %v, want pending", data["state"])
		}
		if !strings.Contains(stdout+stderr, "snapback run") {
			t.Errorf("stdout %q stderr %q, want a fix naming 'snapback run'", stdout, stderr)
		}
		if r.snapCalls != 1 {
			t.Errorf("Snap calls = %d, want 1", r.snapCalls)
		}
	})

	t.Run("wait", func(t *testing.T) {
		r := newSnapRig()
		r.daemonErr = errors.New("dial unix: connection refused")

		code, stdout, _ := runSnap(snapDeps(r, cfg), "--wait", "--json", proj)

		if code != 1 {
			t.Fatalf("snap --wait --json exit = %d, want 1 (stdout %q)", code, stdout)
		}
		if e := decodeEnvelope(t, []byte(stdout)); e.Code != string(errcode.PrereqMissing) {
			t.Errorf("code = %q, want %q", e.Code, errcode.PrereqMissing)
		}
		if r.snapCalls != 1 {
			t.Errorf("Snap calls = %d, want 1", r.snapCalls)
		}
	})
}

func TestSnapLockErrorNotRetried(t *testing.T) {
	cfg, proj := snapFixture(t)
	r := newSnapRig()
	r.snapErr = errcode.New(errcode.RepoUnavailable, "snap", errors.New("unable to create lock"))

	code, stdout, _ := runSnap(snapDeps(r, cfg), "--json", proj)

	if code != 1 {
		t.Fatalf("snap --json exit = %d, want 1 (stdout %q)", code, stdout)
	}
	e := decodeEnvelope(t, []byte(stdout))
	if e.Code != string(errcode.RepoUnavailable) {
		t.Errorf("code = %q, want %q", e.Code, errcode.RepoUnavailable)
	}
	if !strings.Contains(e.Error, "lock") {
		t.Errorf("error = %q, want it to contain lock", e.Error)
	}
	if r.snapCalls != 1 {
		t.Errorf("Snap calls = %d, want exactly 1", r.snapCalls)
	}
	if len(r.submitted) != 0 {
		t.Errorf("SnapSubmitted = %q, want none", r.submitted)
	}
}

// TestSnapUsesRootHost checks that snap records the matched root's
// snapshots.hostname (SPEC §10), falling back to Deps.Hostname when unset.
func TestSnapUsesRootHost(t *testing.T) {
	tests := []struct {
		name     string
		rootHost string
		want     string
	}{
		{name: "root hostname set", rootHost: "lin", want: "lin"},
		{name: "root hostname unset", rootHost: "", want: "mac"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, proj := snapFixture(t)
			cfg.Roots[0].Snapshots.Hostname = tt.rootHost
			r := newSnapRig()
			d := snapDeps(r, cfg)
			d.Hostname = func() (string, error) { return "mac", nil }

			code, _, stderr := runSnap(d, proj)

			if code != 0 {
				t.Fatalf("snap %s exit = %d, want 0 (stderr %q)", proj, code, stderr)
			}
			if len(r.reqs) != 1 {
				t.Fatalf("Snap calls = %d, want 1", len(r.reqs))
			}
			if got := r.reqs[0].Host; got != tt.want {
				t.Errorf("SnapRequest.Host with root hostname %q = %q, want %q", tt.rootHost, got, tt.want)
			}
		})
	}
}

// TestSnapHelpPrintsUsage checks that -h and --help exit 0 and print the
// synopsis, an example and every flag to stderr without taking a snapshot.
func TestSnapHelpPrintsUsage(t *testing.T) {
	cfg, _ := snapFixture(t)
	for _, arg := range []string{"-h", "--help"} {
		t.Run(arg, func(t *testing.T) {
			r := newSnapRig()

			code, stdout, stderr := runSnap(snapDeps(r, cfg), arg)

			if code != 0 {
				t.Fatalf("snap %s exit = %d, want 0 (stderr %q)", arg, code, stderr)
			}
			for _, want := range []string{"snap [flags] [DIR]", "--tag", "--timeout"} {
				if !strings.Contains(stderr, want) {
					t.Errorf("snap %s stderr = %q, want it to contain %q", arg, stderr, want)
				}
			}
			for _, name := range []string{"tag", "repository", "prefix", "wait", "timeout"} {
				if want := "\n  -" + name; !strings.Contains(stderr, want) {
					t.Errorf("snap %s stderr = %q, want a flag line %q", arg, stderr, want)
				}
			}
			if stdout != "" {
				t.Errorf("snap %s stdout = %q, want it empty", arg, stdout)
			}
			if r.snapCalls != 0 {
				t.Errorf("snap %s Snap calls = %d, want 0", arg, r.snapCalls)
			}
		})
	}
}

// TestSnapEndsWithNextStep checks that a successful non-JSON snap ends by
// naming the directory it snapped, and that --json stays a bare envelope.
func TestSnapEndsWithNextStep(t *testing.T) {
	cfg, proj := snapFixture(t)

	t.Run("explicit dir", func(t *testing.T) {
		r := newSnapRig()

		code, stdout, stderr := runSnap(snapDeps(r, cfg), proj)

		if code != 0 {
			t.Fatalf("snap %s exit = %d, want 0 (stderr %q)", proj, code, stderr)
		}
		want := "next: ls " + filepath.Join(proj, ".snapshot")
		if got := lastLine(stdout); got != want {
			t.Errorf("snap %s last stdout line = %q, want %q", proj, got, want)
		}
	})

	t.Run("default dir", func(t *testing.T) {
		r := newSnapRig()
		d := snapDeps(r, cfg)
		d.Getwd = func() (string, error) { return proj, nil }

		code, stdout, stderr := runSnap(d)

		if code != 0 {
			t.Fatalf("snap exit = %d, want 0 (stderr %q)", code, stderr)
		}
		want := "next: ls " + filepath.Join(proj, ".snapshot")
		if got := lastLine(stdout); got != want {
			t.Errorf("snap last stdout line = %q, want %q", got, want)
		}
	})

	t.Run("json", func(t *testing.T) {
		r := newSnapRig()

		code, stdout, stderr := runSnap(snapDeps(r, cfg), "--json", proj)

		if code != 0 {
			t.Fatalf("snap --json %s exit = %d, want 0 (stderr %q)", proj, code, stderr)
		}
		if strings.Contains(stdout, "next:") {
			t.Errorf("snap --json stdout = %q, want no next-step line", stdout)
		}
	})
}
