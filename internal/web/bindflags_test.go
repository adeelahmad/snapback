package web

import (
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

// remoteBind is a non-loopback bind address on an ephemeral port: it needs
// --allow-remote and carries the operator warning.
const remoteBind = "0.0.0.0:0"

// bindCmdRun returns a cmdRun whose config carries web.bind: bind.
func bindCmdRun(t *testing.T, bind string) *cmdRun {
	t.Helper()
	r := newCmdRun(t, map[string]string{})
	raw, err := os.ReadFile(r.env.ConfigPath)
	if err != nil {
		t.Fatalf("os.ReadFile(config) error = %v", err)
	}
	const listen = "  listen: 127.0.0.1:0\n"
	if !strings.Contains(string(raw), listen) {
		t.Fatalf("config = %q, want it to contain %q", raw, listen)
	}
	out := strings.Replace(string(raw), listen, listen+"  bind: "+bind+"\n", 1)
	if err := os.WriteFile(r.env.ConfigPath, []byte(out), 0o600); err != nil {
		t.Fatalf("os.WriteFile(config) error = %v", err)
	}
	return r
}

// served reports whether the run published a listener: web.url in the state
// dir and the same URL on stdout are the only ready signals the command has.
func (r *cmdRun) served() bool {
	_, err := os.Stat(r.stateDir + "/web.url")
	return err == nil || strings.Contains(r.stdout.String(), "http://")
}

func TestBindFlagsRemoteBindRefusedWithoutAllowRemote(t *testing.T) {
	forceGOOS(t, "linux")
	r := bindCmdRun(t, remoteBind)

	code := r.run(Command(), []string{}, func(string) { time.Sleep(100 * time.Millisecond) })

	if code == 0 {
		t.Errorf("Command().Run([]) with web.bind %q = %d, want a non-zero exit", remoteBind, code)
	}
	if got := r.stderr.String(); !strings.Contains(got, "--allow-remote") {
		t.Errorf("Command().Run([]) with web.bind %q stderr = %q, want it to name --allow-remote", remoteBind, got)
	}
	if r.served() {
		t.Errorf("Command().Run([]) with web.bind %q opened a listener (stdout %q), want none", remoteBind, r.stdout.String())
	}
}

func TestBindFlagsRemoteWarningPrecedesFirstRequest(t *testing.T) {
	forceGOOS(t, "linux")
	r := bindCmdRun(t, remoteBind)

	var atReady string
	code := r.run(Command(), []string{"--allow-remote"}, func(string) { atReady = r.stderr.String() })

	if code != 0 {
		t.Errorf("Command().Run([--allow-remote]) with web.bind %q = %d, want 0 (stderr %q)", remoteBind, code, r.stderr.String())
	}
	if !strings.Contains(atReady, "reachable from other machines") {
		t.Errorf("stderr when the listener became ready = %q, want the remote-bind warning before the first request", atReady)
	}
}

func TestBindFlagsBindOverridesConfigBind(t *testing.T) {
	forceGOOS(t, "linux")

	// The config bind governs on its own: without that, --bind overrides
	// nothing.
	plain := bindCmdRun(t, remoteBind)
	if code := plain.run(Command(), []string{}, func(string) { time.Sleep(100 * time.Millisecond) }); code == 0 {
		t.Errorf("Command().Run([]) with web.bind %q = 0, want a non-zero exit so --bind has something to override", remoteBind)
	}

	over := bindCmdRun(t, remoteBind)
	var printed string
	code := over.run(Command(), []string{"--bind", "127.0.0.1:0"}, func(url string) { printed = url })
	if code != 0 {
		t.Errorf("Command().Run([--bind 127.0.0.1:0]) with web.bind %q = %d, want 0 (stderr %q)", remoteBind, code, over.stderr.String())
	}
	if !strings.HasPrefix(printed, "http://127.0.0.1:") {
		t.Errorf("published URL with --bind 127.0.0.1:0 = %q, want a http://127.0.0.1: URL", printed)
	}
	if got := over.stderr.String(); strings.Contains(got, "reachable from other machines") {
		t.Errorf("stderr with --bind 127.0.0.1:0 = %q, want no remote-bind warning", got)
	}
}

func TestBindFlagsAllowOriginRepeats(t *testing.T) {
	forceGOOS(t, "linux")
	const a, b = "https://a.example", "https://b.example"
	r := newCmdRun(t, map[string]string{})

	var got map[string]int
	code := r.run(Command(), []string{"--allow-origin", a, "--allow-origin", b}, func(url string) {
		got = map[string]int{
			a:                   writeStatus(t, url, a),
			b:                   writeStatus(t, url, b),
			"https://c.example": writeStatus(t, url, "https://c.example"),
		}
	})
	if code != 0 {
		t.Errorf("Command().Run([--allow-origin ...]) = %d, want 0 (stderr %q)", code, r.stderr.String())
	}
	// The guard rejects a cross-origin write with 403 before the session
	// check, so an allowed origin reaches the session check (401) instead.
	for _, origin := range []string{a, b} {
		if got[origin] == http.StatusForbidden {
			t.Errorf("PUT /api/config Origin %q with --allow-origin %s --allow-origin %s: status = %d, want it past the cross-origin guard",
				origin, a, b, got[origin])
		}
	}
	if want := http.StatusForbidden; got["https://c.example"] != want {
		t.Errorf("PUT /api/config Origin %q with --allow-origin %s --allow-origin %s: status = %d, want %d",
			"https://c.example", a, b, got["https://c.example"], want)
	}
}

// writeStatus sends an unauthenticated cross-origin PUT to the server's
// /api/config and returns its status.
func writeStatus(t *testing.T, authURL, origin string) int {
	t.Helper()
	base, _, ok := strings.Cut(authURL, "auth?")
	if !ok {
		t.Fatalf("published URL = %q, want it to contain %q", authURL, "auth?")
	}
	req, err := http.NewRequest(http.MethodPut, base+"api/config", strings.NewReader(`{"revision":"","config":{}}`))
	if err != nil {
		t.Fatalf("http.NewRequest() error = %v", err)
	}
	req.Header.Set("Origin", origin)
	req.Header.Set("Content-Type", "application/json")
	resp, err := (&http.Client{Timeout: 5 * time.Second}).Do(req)
	if err != nil {
		t.Fatalf("PUT %q error = %v", base+"api/config", err)
	}
	defer func() { _ = resp.Body.Close() }()
	return resp.StatusCode
}

func TestBindFlagsHelpDocumentsThem(t *testing.T) {
	r := newCmdRun(t, map[string]string{})
	if code := Command().Run(t.Context(), r.env, []string{"-h"}); code != 0 {
		t.Errorf("Command().Run([-h]) = %d, want 0", code)
	}
	out := r.stderr.String()
	tests := []struct{ flag, want string }{
		{"--bind", "address"},
		{"--allow-origin", "origin"},
		{"--allow-remote", "loopback"},
	}
	for _, tt := range tests {
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

// flagUsage returns the description the shared flag printer printed for flag
// name in out, or "" when the flag is absent or undocumented.
func flagUsage(out, name string) string {
	lines := strings.Split(out, "\n")
	for i, line := range lines {
		if !strings.HasPrefix(strings.TrimRight(line, " \t"), "  "+name) {
			continue
		}
		var desc []string
		for _, next := range lines[i+1:] {
			if strings.HasPrefix(next, "  -") {
				break
			}
			desc = append(desc, strings.TrimSpace(next))
		}
		return strings.TrimSpace(strings.Join(desc, " "))
	}
	return ""
}
