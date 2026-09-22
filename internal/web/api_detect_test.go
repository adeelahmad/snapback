package web

import (
	"context"
	"encoding/json"
	"io/fs"
	"net/http"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/setup"
)

// detectBody is the wire shape of GET /api/detect. The tags pin the JSON
// names the setup page's JavaScript reads.
type detectBody struct {
	ResticPath     string              `json:"restic_path"`
	RclonePath     string              `json:"rclone_path"`
	RepoURI        string              `json:"repo_uri"`
	CredentialFile string              `json:"credential_file"`
	Hostname       string              `json:"hostname"`
	Roots          []string            `json:"roots"`
	PrefixMap      []detectPrefixEntry `json:"prefix_map"`
}

// detectPrefixEntry is one derived prefix mapping as the API reports it.
type detectPrefixEntry struct {
	Hostname   string `json:"hostname"`
	SourcePath string `json:"source_path"`
	TreePrefix string `json:"tree_prefix"`
}

// fakeRunner is a setup.Runner that returns out for every call and records
// the argument vectors it was given, so a test can prove the probe read the
// repository and wrote nothing to it.
type fakeRunner struct {
	mu   sync.Mutex
	out  string
	err  error
	arvs [][]string
}

func (f *fakeRunner) run(_ context.Context, name string, args ...string) ([]byte, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.arvs = append(f.arvs, append([]string{name}, args...))
	return []byte(f.out), f.err
}

func (f *fakeRunner) calls() [][]string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Clone(f.arvs)
}

// detectEnv is the environment the detect tests inject. Nothing reads the
// process environment, so no test depends on the machine it runs on.
var detectEnv = map[string]string{
	"RESTIC_REPOSITORY":    "sftp:backup:/srv/repo",
	"RESTIC_PASSWORD_FILE": "/etc/snapback/repo.pw",
}

// detectOptions returns Options whose detection seams are all fakes: a fixed
// environment, a PATH that finds restic and rclone at fixed paths, one root,
// a fixed machine hostname and run as the probe's command runner.
func detectOptions(run setup.Runner) Options {
	opts := Options{Backend: &fakeBackend{cfg: &config.Config{}, rev: "r1"}}
	opts.SetupDeps = setup.Deps{
		Getenv:   func(k string) string { return detectEnv[k] },
		LookPath: func(name string) (string, error) { return "/fake/bin/" + name, nil },
		Getwd:    func() (string, error) { return "/home/u", nil },
		Hostname: func() (string, error) { return "laptop", nil },
		Given:    []string{"/home/u"},
	}
	opts.SetupRunner = run
	return opts
}

// snapshotsJSON is one `restic snapshots --json` reply: the repository holds
// /data/home/u on host probe-host, which is the mapping /home/u needs.
var snapshotsJSON = `[{"id":"` + strings.Repeat("a", 64) + `","hostname":"probe-host",` +
	`"paths":["/data/home/u"],"time":"2026-01-02T03:04:05Z"}]`

func TestAPIDetectReportsEnvironmentBinariesAndPrefixMap(t *testing.T) {
	run := &fakeRunner{out: snapshotsJSON}
	srv, cookie, _ := newTestServer(t, detectOptions(run.run))

	w := do(t, srv, http.MethodGet, "/api/detect", nil, apiHeader(cookie, ""))
	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf("GET /api/detect: status = %d, want %d (body %q)", got, want, w.Body.String())
	}
	if got, want := w.Header().Get("Cache-Control"), "no-store"; got != want {
		t.Errorf("GET /api/detect: Cache-Control = %q, want %q", got, want)
	}
	var got detectBody
	decodeJSON(t, w, &got)

	want := detectBody{
		ResticPath:     "/fake/bin/restic",
		RclonePath:     "/fake/bin/rclone",
		RepoURI:        "sftp:backup:/srv/repo",
		CredentialFile: "/etc/snapback/repo.pw",
		Hostname:       "probe-host",
		Roots:          []string{"/home/u"},
		PrefixMap: []detectPrefixEntry{{
			Hostname:   "probe-host",
			SourcePath: "/data/home/u",
			TreePrefix: "/data/home/u",
		}},
	}
	if !sameDetect(got, want) {
		t.Errorf("GET /api/detect = %+v, want %+v", got, want)
	}
}

// sameDetect compares two detect bodies by their JSON encoding, so the
// failure above can name the whole shape rather than one field.
func sameDetect(a, b detectBody) bool {
	ja, err := json.Marshal(a)
	if err != nil {
		return false
	}
	jb, err := json.Marshal(b)
	if err != nil {
		return false
	}
	return string(ja) == string(jb)
}

func TestAPIDetectProbesReadOnlyAndWritesNothing(t *testing.T) {
	run := &fakeRunner{out: snapshotsJSON}
	opts := detectOptions(run.run)
	backend := opts.Backend.(*fakeBackend)
	srv, cookie, _ := newTestServer(t, opts)
	before := treeOf(t, srv.opts.StateDir)

	w := do(t, srv, http.MethodGet, "/api/detect", nil, apiHeader(cookie, ""))
	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf("GET /api/detect: status = %d, want %d (body %q)", got, want, w.Body.String())
	}

	calls := run.calls()
	if len(calls) != 1 {
		t.Fatalf("restic invocations = %d (%v), want 1", len(calls), calls)
	}
	argv := calls[0]
	if got, want := argv[0], "/fake/bin/restic"; got != want {
		t.Errorf("probe ran %q, want %q", got, want)
	}
	if !slices.Contains(argv, "--no-lock") {
		t.Errorf("probe argv = %v, want it to carry --no-lock", argv)
	}
	for _, write := range []string{"backup", "forget", "prune", "unlock", "init", "restore"} {
		if slices.Contains(argv, write) {
			t.Errorf("probe argv = %v, want no write subcommand, got %q", argv, write)
		}
	}
	if len(backend.saved) != 0 {
		t.Errorf("SaveConfig calls = %d, want 0: detect must not write the config", len(backend.saved))
	}
	if after := treeOf(t, srv.opts.StateDir); !slices.Equal(before, after) {
		t.Errorf("state dir = %v, want it unchanged at %v", after, before)
	}
}

// treeOf lists the paths under dir, relative to it and sorted.
func treeOf(t *testing.T, dir string) []string {
	t.Helper()
	var out []string
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(dir, p)
		if err != nil {
			return err
		}
		out = append(out, rel)
		return nil
	})
	if err != nil {
		t.Fatalf("walk %q error = %v", dir, err)
	}
	slices.Sort(out)
	return out
}

func TestAPIDetectWithoutSnapshotsFallsBackToTheMachineHostname(t *testing.T) {
	run := &fakeRunner{out: `[]`}
	srv, cookie, _ := newTestServer(t, detectOptions(run.run))

	w := do(t, srv, http.MethodGet, "/api/detect", nil, apiHeader(cookie, ""))
	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf("GET /api/detect: status = %d, want %d (body %q)", got, want, w.Body.String())
	}
	var got detectBody
	decodeJSON(t, w, &got)
	if want := "laptop"; got.Hostname != want {
		t.Errorf("GET /api/detect hostname = %q, want %q", got.Hostname, want)
	}
	if len(got.PrefixMap) != 0 {
		t.Errorf("GET /api/detect prefix_map = %+v, want empty without snapshots", got.PrefixMap)
	}
}

func TestAPIDetectIsSessionGuarded(t *testing.T) {
	run := &fakeRunner{out: snapshotsJSON}
	srv, cookie, _ := newTestServer(t, detectOptions(run.run))

	w := do(t, srv, http.MethodGet, "/api/detect", nil, http.Header{})
	if got, want := w.Code, http.StatusUnauthorized; got != want {
		t.Errorf("GET /api/detect without a session: status = %d, want %d", got, want)
	}
	if n := len(run.calls()); n != 0 {
		t.Errorf("restic invocations without a session = %d, want 0", n)
	}
	w = do(t, srv, http.MethodGet, "/api/detect", nil, apiHeader(cookie, ""))
	if got, want := w.Code, http.StatusOK; got != want {
		t.Errorf("GET /api/detect with a session: status = %d, want %d (body %q)", got, want, w.Body.String())
	}
}

func TestSetupPagePrefillsDetectedValues(t *testing.T) {
	run := &fakeRunner{out: snapshotsJSON}
	srv, cookie, _ := newTestServer(t, detectOptions(run.run))

	w := do(t, srv, http.MethodGet, "/setup", nil, http.Header{"Cookie": {cookie.String()}})
	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf("GET /setup: status = %d, want %d", got, want)
	}
	body := w.Body.String()
	for _, want := range []string{
		`/fake/bin/restic`,
		`/fake/bin/rclone`,
		`value="sftp:backup:/srv/repo"`,
		`value="/etc/snapback/repo.pw"`,
		`probe-host`,
		`/data/home/u`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("GET /setup body does not contain %q; want it prefilled from detection", want)
		}
	}
}
