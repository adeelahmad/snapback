package web

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/cli"
)

// runTimeout bounds one command run so a broken command cannot hang the test.
const runTimeout = 10 * time.Second

var urlRE = regexp.MustCompile(`http://127\.0\.0\.1:\d+/\S*`)

// syncBuffer is a bytes.Buffer safe for concurrent writes and reads.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

// browserRecorder replaces openBrowser for the test and records each URL.
type browserRecorder struct {
	mu   sync.Mutex
	urls []string
}

func recordBrowser(t *testing.T) *browserRecorder {
	t.Helper()
	rec := &browserRecorder{}
	prev := openBrowser
	openBrowser = func(ctx context.Context, url string) error {
		rec.mu.Lock()
		defer rec.mu.Unlock()
		rec.urls = append(rec.urls, url)
		return nil
	}
	t.Cleanup(func() { openBrowser = prev })
	return rec
}

func (r *browserRecorder) calls() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.urls...)
}

// waitCalls polls rec until it has at least one call or a second passes.
func (r *browserRecorder) waitCalls() []string {
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if c := r.calls(); len(c) > 0 {
			return c
		}
		time.Sleep(10 * time.Millisecond)
	}
	return r.calls()
}

func forceGOOS(t *testing.T, v string) {
	t.Helper()
	prev := goos
	goos = v
	t.Cleanup(func() { goos = prev })
}

// cmdRun is one command invocation with its own config and state dir.
type cmdRun struct {
	env            cli.Env
	stateDir       string
	stdout, stderr *syncBuffer
}

// newCmdRun writes a valid config (config/testdata/example.yaml, which
// listens on 127.0.0.1:0) into a fresh temp dir and returns an env whose
// Getenv serves vars.
func newCmdRun(t *testing.T, vars map[string]string) *cmdRun {
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
	cfgPath := filepath.Join(tmp, "config.yaml")
	if err := os.WriteFile(cfgPath, bytes.ReplaceAll(raw, []byte("@TMP@"), []byte(tmp)), 0o600); err != nil {
		t.Fatalf("os.WriteFile(config) error = %v", err)
	}
	r := &cmdRun{stateDir: filepath.Join(tmp, "state"), stdout: &syncBuffer{}, stderr: &syncBuffer{}}
	r.env = cli.Env{
		Stdout:     r.stdout,
		Stderr:     r.stderr,
		Getenv:     func(k string) string { return vars[k] },
		ConfigPath: cfgPath,
	}
	return r
}

// run executes cmd with args. Once web.url appears it calls during (if not
// nil) with the published URL while the server is up, then cancels ctx.
func (r *cmdRun) run(cmd cli.Command, args []string, during func(url string)) int {
	ctx, cancel := context.WithTimeout(context.Background(), runTimeout)
	defer cancel()
	done := make(chan struct{})
	watched := make(chan struct{})
	go func() {
		defer close(watched)
		urlFile := filepath.Join(r.stateDir, "web.url")
		for {
			select {
			case <-done:
				return
			case <-ctx.Done():
				return
			case <-time.After(20 * time.Millisecond):
			}
			data, err := os.ReadFile(urlFile)
			if err != nil || len(bytes.TrimSpace(data)) == 0 {
				continue
			}
			if during != nil {
				during(strings.TrimSpace(string(data)))
			}
			cancel()
			return
		}
	}()
	code := cmd.Run(ctx, r.env, args)
	close(done)
	<-watched
	return code
}

// jarClient is an HTTP client that keeps cookies and follows redirects.
func jarClient(t *testing.T) *http.Client {
	t.Helper()
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("cookiejar.New() error = %v", err)
	}
	return &http.Client{Jar: jar, Timeout: 5 * time.Second}
}

// get fetches url with c and returns the final request path and the body.
func get(t *testing.T, c *http.Client, url string) (path, body string) {
	t.Helper()
	resp, err := c.Get(url)
	if err != nil {
		t.Errorf("GET %q error = %v", url, err)
		return "", ""
	}
	defer func() { _ = resp.Body.Close() }()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("reading GET %q body error = %v", url, err)
	}
	return resp.Request.URL.Path, string(b)
}

func TestWebCommandPrintsURLAndOpensOnlyWhenAsked(t *testing.T) {
	forceGOOS(t, "linux")
	rec := recordBrowser(t)
	vars := map[string]string{"DISPLAY": ":0"}

	if got := Command().Name; got != "web" {
		t.Errorf("Command().Name = %q, want %q", got, "web")
	}

	plain := newCmdRun(t, vars)
	if code := plain.run(Command(), []string{}, nil); code != 0 {
		t.Errorf("Command().Run([]) = %d, want 0 (stderr %q)", code, plain.stderr.String())
	}
	if !strings.Contains(plain.stdout.String(), "http://127.0.0.1:") {
		t.Errorf("Command().Run([]) stdout = %q, want an http://127.0.0.1: URL", plain.stdout.String())
	}
	if got := rec.calls(); len(got) != 0 {
		t.Errorf("openBrowser calls without --open = %q, want none", got)
	}

	open := newCmdRun(t, vars)
	var calls []string
	code := open.run(Command(), []string{"--open"}, func(string) { calls = rec.waitCalls() })
	if code != 0 {
		t.Errorf("Command().Run([--open]) = %d, want 0 (stderr %q)", code, open.stderr.String())
	}
	printed := urlRE.FindString(open.stdout.String())
	if printed == "" {
		t.Fatalf("Command().Run([--open]) stdout = %q, want an http://127.0.0.1: URL", open.stdout.String())
	}
	if len(calls) != 1 || calls[0] != printed {
		t.Errorf("openBrowser calls with --open = %q, want exactly [%q]", calls, printed)
	}
}

func TestNoBrowserWithoutDesktopOrUnderService(t *testing.T) {
	forceGOOS(t, "linux")
	tests := []struct {
		name string
		vars map[string]string
	}{
		{name: "no desktop session", vars: map[string]string{"DISPLAY": "", "WAYLAND_DISPLAY": ""}},
		{name: "under a service", vars: map[string]string{"DISPLAY": ":0", "INVOCATION_ID": "abc"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := recordBrowser(t)
			r := newCmdRun(t, tt.vars)
			code := r.run(Command(), []string{"--open"}, func(string) { time.Sleep(100 * time.Millisecond) })
			if code != 0 {
				t.Errorf("Command().Run([--open]) = %d, want 0 (stderr %q)", code, r.stderr.String())
			}
			if got := rec.calls(); len(got) != 0 {
				t.Errorf("openBrowser calls = %q, want none", got)
			}
			printed := urlRE.FindString(r.stdout.String())
			if printed == "" || !strings.Contains(r.stderr.String(), printed) {
				t.Errorf("Command().Run([--open]) stderr = %q, want it to name the URL %q to open manually", r.stderr.String(), printed)
			}
		})
	}
}

// overrideAssets copies the webui templates and assets into a temp dir and
// marks status.html with OVERRIDE-MARK.
func overrideAssets(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, sub := range []string{"templates", "assets"} {
		src := filepath.Join("..", "webui", sub)
		if err := os.CopyFS(filepath.Join(dir, sub), os.DirFS(src)); err != nil {
			t.Fatalf("os.CopyFS(%q) error = %v", src, err)
		}
	}
	status := filepath.Join(dir, "templates", "status.html")
	data, err := os.ReadFile(status)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) error = %v", status, err)
	}
	marker := []byte(`{{define "content"}}`)
	i := bytes.Index(data, marker)
	if i < 0 {
		t.Fatalf("status.html missing %q", marker)
	}
	data = append(data[:i+len(marker):i+len(marker)], append([]byte("\nOVERRIDE-MARK\n"), data[i+len(marker):]...)...)
	if err := os.WriteFile(status, data, 0o600); err != nil {
		t.Fatalf("os.WriteFile(%q) error = %v", status, err)
	}
	return dir
}

func TestConfigCommandOpensSetupAndAssetsFlag(t *testing.T) {
	forceGOOS(t, "linux")
	vars := map[string]string{"DISPLAY": ":0"}

	if got := ConfigCommand().Name; got != "config" {
		t.Errorf("ConfigCommand().Name = %q, want %q", got, "config")
	}

	rec := recordBrowser(t)
	cfg := newCmdRun(t, vars)
	var opened, finalPath string
	code := cfg.run(ConfigCommand(), nil, func(string) {
		if calls := rec.waitCalls(); len(calls) == 1 {
			opened = calls[0]
			finalPath, _ = get(t, jarClient(t), opened)
		}
	})
	if code != 0 {
		t.Errorf("ConfigCommand().Run(nil) = %d, want 0 (stderr %q)", code, cfg.stderr.String())
	}
	if !strings.Contains(opened, "auth?token=") || !strings.Contains(opened, "next=/setup") {
		t.Errorf("ConfigCommand().Run(nil) opened %q, want an auth?token= URL with next=/setup", opened)
	}
	if finalPath != "/setup" {
		t.Errorf("following %q ends at path %q, want %q", opened, finalPath, "/setup")
	}

	assets := overrideAssets(t)
	ov := newCmdRun(t, vars)
	var body string
	code = ov.run(Command(), []string{"--assets", assets}, func(url string) {
		c := jarClient(t)
		get(t, c, url)
		i := strings.Index(url, "/auth")
		if i < 0 {
			t.Errorf("web.url = %q, want an /auth URL", url)
			return
		}
		_, body = get(t, c, url[:i]+"/")
	})
	if code != 0 {
		t.Errorf("Command().Run([--assets %s]) = %d, want 0 (stderr %q)", assets, code, ov.stderr.String())
	}
	if !strings.Contains(body, "OVERRIDE-MARK") {
		t.Errorf("logged-in GET / with --assets = %q, want it to contain OVERRIDE-MARK", body)
	}

	bad := newCmdRun(t, vars)
	code = bad.run(Command(), []string{"--assets", "/nonexistent"}, nil)
	if code == 0 {
		t.Errorf("Command().Run([--assets /nonexistent]) = 0, want non-zero")
	}
	if !strings.Contains(bad.stderr.String(), "invalid_configuration") {
		t.Errorf("Command().Run([--assets /nonexistent]) stderr = %q, want invalid_configuration", bad.stderr.String())
	}
}

func TestWebAndConfigUsage(t *testing.T) {
	tests := []struct {
		name string
		cmd  cli.Command
		want []string
	}{
		{
			name: "web",
			cmd:  Command(),
			want: []string{"Usage: snapback web [flags]", "Args:", "Example:", "snapback web", "-assets", "-open"},
		},
		{
			name: "config",
			cmd:  ConfigCommand(),
			want: []string{"Usage: snapback config [flags]", "Args:", "Example:", "-file"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, arg := range []string{"-h", "--help"} {
				r := newCmdRun(t, nil)
				if code := r.run(tt.cmd, []string{arg}, nil); code != 0 {
					t.Errorf("%s %s = exit %d, want 0; stderr %q", tt.name, arg, code, r.stderr.String())
				}
				if out := r.stdout.String(); out != "" {
					t.Errorf("%s %s stdout = %q, want it empty", tt.name, arg, out)
				}
				for _, want := range tt.want {
					if !strings.Contains(r.stderr.String(), want) {
						t.Errorf("%s %s stderr = %q, want it to contain %q", tt.name, arg, r.stderr.String(), want)
					}
				}
			}

			bad := newCmdRun(t, nil)
			if code := bad.run(tt.cmd, []string{"--nope"}, nil); code != 2 {
				t.Errorf("%s --nope = exit %d, want 2; stderr %q", tt.name, code, bad.stderr.String())
			}
			if !strings.Contains(bad.stderr.String(), "Usage: snapback "+tt.name+" [flags]") {
				t.Errorf("%s --nope stderr = %q, want it to contain the usage", tt.name, bad.stderr.String())
			}
		})
	}
}
