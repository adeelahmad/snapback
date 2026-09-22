package web

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/cli"
	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/errcode"
)

// firstRun is a `snapback config` invocation on a machine with no config file.
type firstRun struct {
	tmp            string
	cfgPath        string
	env            cli.Env
	stdout, stderr *syncBuffer
}

// newFirstRun returns an env whose ConfigPath is tmp/fresh/sub/config.yaml,
// which does not exist and whose parent dirs do not exist either.
func newFirstRun(t *testing.T) *firstRun {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_STATE_HOME", filepath.Join(tmp, "xdgstate"))
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmp, "xdgconfig"))
	vars := map[string]string{"DISPLAY": ":0", "HOME": tmp}
	r := &firstRun{
		tmp:     tmp,
		cfgPath: filepath.Join(tmp, "fresh", "sub", "config.yaml"),
		stdout:  &syncBuffer{},
		stderr:  &syncBuffer{},
	}
	r.env = cli.Env{
		Stdout:     r.stdout,
		Stderr:     r.stderr,
		Getenv:     func(k string) string { return vars[k] },
		ConfigPath: r.cfgPath,
	}
	return r
}

// run executes ConfigCommand with args. Once the server prints its URL it
// calls during with the URL the browser was sent to, then cancels ctx.
func (r *firstRun) run(rec *browserRecorder, args []string, during func(openURL string)) int {
	ctx, cancel := context.WithTimeout(context.Background(), runTimeout)
	defer cancel()
	done := make(chan struct{})
	watched := make(chan struct{})
	go func() {
		defer close(watched)
		for {
			select {
			case <-done:
				return
			case <-ctx.Done():
				return
			case <-time.After(20 * time.Millisecond):
			}
			if !urlRE.MatchString(r.stdout.String()) {
				continue
			}
			if calls := rec.waitCalls(); len(calls) > 0 && during != nil {
				during(calls[0])
			}
			cancel()
			return
		}
	}()
	code := ConfigCommand().Run(ctx, r.env, args)
	close(done)
	<-watched
	return code
}

// validConfigYAML returns config/testdata/example.yaml with its @TMP@ paths
// pointing at a fresh dir that holds the state, work and password files.
func validConfigYAML(t *testing.T) []byte {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "config", "testdata", "example.yaml"))
	if err != nil {
		t.Fatalf("os.ReadFile(example.yaml) error = %v", err)
	}
	tmp := t.TempDir()
	for _, d := range []string{"state", "work"} {
		if err := os.MkdirAll(filepath.Join(tmp, d), 0o755); err != nil {
			t.Fatalf("os.MkdirAll(%q) error = %v", d, err)
		}
	}
	if err := os.WriteFile(filepath.Join(tmp, "password"), []byte("pw\n"), 0o600); err != nil {
		t.Fatalf("os.WriteFile(password) error = %v", err)
	}
	return bytes.ReplaceAll(raw, []byte("@TMP@"), []byte(tmp))
}

// session logs a jar client in through openURL and returns it with the
// server's base URL and the session's CSRF token.
func session(t *testing.T, openURL string) (c *http.Client, base, csrf string) {
	t.Helper()
	c = jarClient(t)
	base = openURL[:strings.Index(openURL, "/auth")+1]
	get(t, c, openURL)
	_, page := get(t, c, base+"config")
	m := csrfMeta.FindStringSubmatch(page)
	if m == nil {
		t.Errorf("GET /config: no csrf-token meta")
		return c, base, ""
	}
	return c, base, m[1]
}

// putConfig sends body to PUT /api/config and returns the status and body.
func putConfig(t *testing.T, c *http.Client, base, csrf string, body []byte) (int, []byte) {
	t.Helper()
	req, err := http.NewRequest(http.MethodPut, base+"api/config", bytes.NewReader(body))
	if err != nil {
		t.Errorf("http.NewRequest() error = %v", err)
		return 0, nil
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF-Token", csrf)
	resp, err := c.Do(req)
	if err != nil {
		t.Errorf("PUT /api/config error = %v", err)
		return 0, nil
	}
	defer func() { _ = resp.Body.Close() }()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("reading PUT /api/config body error = %v", err)
	}
	return resp.StatusCode, b
}

func TestConfigCommandFirstRunServesSetup(t *testing.T) {
	forceGOOS(t, "linux")
	rec := recordBrowser(t)
	r := newFirstRun(t)

	var gotRev config.Revision = "unset"
	var gotStatus int
	code := r.run(rec, nil, func(openURL string) {
		c, base, _ := session(t, openURL)
		resp, err := c.Get(base + "api/config")
		if err != nil {
			t.Errorf("GET /api/config error = %v", err)
			return
		}
		defer func() { _ = resp.Body.Close() }()
		gotStatus = resp.StatusCode
		var body configBody
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			t.Errorf("decoding GET /api/config error = %v", err)
		}
		gotRev = body.Revision
	})

	if code != 0 {
		t.Errorf("ConfigCommand().Run() with no config file = %d, want 0 (stderr %q)", code, r.stderr.String())
	}
	if !urlRE.MatchString(r.stdout.String()) {
		t.Errorf("ConfigCommand() stdout = %q, want the server URL", r.stdout.String())
	}
	calls := rec.calls()
	if len(calls) != 1 || !strings.Contains(calls[0], "next=/setup") {
		t.Errorf("openBrowser calls = %q, want one URL with next=/setup", calls)
	}
	if gotStatus != http.StatusOK {
		t.Errorf("GET /api/config on first run: status = %d, want %d", gotStatus, http.StatusOK)
	}
	if gotRev != "" {
		t.Errorf("GET /api/config on first run: revision = %q, want the zero revision", gotRev)
	}
}

func TestConfigCommandFirstRunSaveCreatesFile(t *testing.T) {
	forceGOOS(t, "linux")
	rec := recordBrowser(t)
	r := newFirstRun(t)
	cfg, err := config.Parse(validConfigYAML(t))
	if err != nil {
		t.Fatalf("config.Parse(example.yaml) error = %v", err)
	}
	body, err := json.Marshal(configBody{Revision: "", Config: cfg})
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var gotStatus int
	var gotBody []byte
	r.run(rec, nil, func(openURL string) {
		c, base, csrf := session(t, openURL)
		gotStatus, gotBody = putConfig(t, c, base, csrf, body)
	})

	if gotStatus != http.StatusOK {
		t.Fatalf("PUT /api/config at revision 0: status = %d, want %d (body %q, stderr %q)",
			gotStatus, http.StatusOK, gotBody, r.stderr.String())
	}
	fi, err := os.Stat(r.cfgPath)
	if err != nil {
		t.Fatalf("os.Stat(config) after PUT error = %v", err)
	}
	if got, want := fi.Mode().Perm(), os.FileMode(0o600); got != want {
		t.Errorf("config file mode = %v, want %v", got, want)
	}
	for _, d := range []string{filepath.Dir(r.cfgPath), filepath.Dir(filepath.Dir(r.cfgPath))} {
		di, err := os.Stat(d)
		if err != nil {
			t.Fatalf("os.Stat(%q) error = %v", d, err)
		}
		if got, want := di.Mode().Perm(), os.FileMode(0o700); got != want {
			t.Errorf("config parent %q mode = %v, want %v", d, got, want)
		}
	}
	saved, _, err := config.Load(r.cfgPath)
	if err != nil {
		t.Fatalf("config.Load(saved) error = %v", err)
	}
	if err := config.Validate(saved); err != nil {
		t.Errorf("config.Validate(saved) = %v, want nil", err)
	}
}

func TestConfigCommandFirstRunInvalidSaveIs400(t *testing.T) {
	forceGOOS(t, "linux")
	rec := recordBrowser(t)
	r := newFirstRun(t)

	var gotStatus int
	var gotBody []byte
	r.run(rec, nil, func(openURL string) {
		c, base, csrf := session(t, openURL)
		gotStatus, gotBody = putConfig(t, c, base, csrf, []byte(`{"revision":"","config":{"version":99}}`))
	})

	if gotStatus != http.StatusBadRequest {
		t.Fatalf("PUT /api/config invalid: status = %d, want %d (body %q)", gotStatus, http.StatusBadRequest, gotBody)
	}
	var e errorBody
	if err := json.Unmarshal(gotBody, &e); err != nil {
		t.Errorf("json.Unmarshal(%q) error = %v", gotBody, err)
	}
	if got, want := e.Code, string(errcode.InvalidConfig); got != want {
		t.Errorf("PUT /api/config invalid: code = %q, want %q", got, want)
	}
	if !strings.Contains(e.Error, "version") {
		t.Errorf("PUT /api/config invalid: error = %q, want a field error naming version", e.Error)
	}
	if _, err := os.Stat(r.cfgPath); !os.IsNotExist(err) {
		t.Errorf("os.Stat(config) after invalid PUT error = %v, want not-exist", err)
	}
}

func TestConfigCommandFileNonInteractive(t *testing.T) {
	forceGOOS(t, "linux")
	rec := recordBrowser(t)
	r := newFirstRun(t)
	src := filepath.Join(r.tmp, "candidate.yaml")
	if err := os.WriteFile(src, validConfigYAML(t), 0o600); err != nil {
		t.Fatalf("os.WriteFile(candidate) error = %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if got := ConfigCommand().Run(ctx, r.env, []string{"--file", src}); got != 0 {
		t.Errorf("ConfigCommand().Run(--file valid) = %d, want 0 (stderr %q)", got, r.stderr.String())
	}
	if _, _, err := config.Load(r.cfgPath); err != nil {
		t.Errorf("config.Load(config dir) after --file error = %v, want the written config", err)
	}
	if calls := rec.calls(); len(calls) != 0 {
		t.Errorf("openBrowser calls with --file = %q, want none", calls)
	}

	bad := filepath.Join(r.tmp, "bad.yaml")
	if err := os.WriteFile(bad, []byte("version: 99\n"), 0o600); err != nil {
		t.Fatalf("os.WriteFile(bad) error = %v", err)
	}
	inv := newFirstRun(t)
	if err := os.MkdirAll(filepath.Dir(inv.cfgPath), 0o700); err != nil {
		t.Fatalf("os.MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(inv.cfgPath, validConfigYAML(t), 0o600); err != nil {
		t.Fatalf("os.WriteFile(existing config) error = %v", err)
	}
	before, err := os.ReadFile(inv.cfgPath)
	if err != nil {
		t.Fatalf("os.ReadFile(existing config) error = %v", err)
	}
	ctx2, cancel2 := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel2()
	if got := ConfigCommand().Run(ctx2, inv.env, []string{"--file", bad}); got != 1 {
		t.Errorf("ConfigCommand().Run(--file invalid) = %d, want 1", got)
	}
	if got, want := inv.stderr.String(), string(errcode.InvalidConfig); !strings.Contains(got, want) {
		t.Errorf("ConfigCommand().Run(--file invalid) stderr = %q, want it to contain %q", got, want)
	}
	after, err := os.ReadFile(inv.cfgPath)
	if err != nil {
		t.Fatalf("os.ReadFile(existing config) after invalid --file error = %v", err)
	}
	if !bytes.Equal(before, after) {
		t.Errorf("config after invalid --file changed, want it untouched")
	}
}
