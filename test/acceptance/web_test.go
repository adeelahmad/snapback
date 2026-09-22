//go:build integration

package acceptance

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const hostileName = `<img src=x onerror=alert(1)>`

// startWeb starts `snapback web` and returns the bootstrap URL it prints.
func startWeb(t *testing.T, e env) *url.URL {
	t.Helper()
	cmd := exec.Command(snapbackBin, "web")
	cmd.Env = append(e.environ(), "DISPLAY=", "WAYLAND_DISPLAY=", "INVOCATION_ID=acc16")
	out, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		t.Fatalf("start snapback web: %v", err)
	}
	done := make(chan struct{})
	go func() { _ = cmd.Wait(); close(done) }()
	t.Cleanup(func() { _ = cmd.Process.Kill(); <-done })
	lines := make(chan string, 1)
	go func() {
		sc := bufio.NewScanner(out)
		if sc.Scan() {
			lines <- sc.Text()
		}
		_, _ = io.Copy(io.Discard, out)
	}()
	select {
	case line := <-lines:
		u, err := url.Parse(strings.TrimSpace(line))
		if err != nil || u.Host == "" {
			t.Fatalf("snapback web printed %q, want a bootstrap URL", line)
		}
		return u
	case <-done:
		t.Fatal("snapback web exited before printing its URL")
	case <-time.After(10 * time.Second):
		t.Fatal("snapback web printed no URL within 10s")
	}
	return nil
}

func noRedirect(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }

func do(t *testing.T, c *http.Client, method, u string, hdr map[string]string, body string) (int, string) {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), method, u, strings.NewReader(body))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	for k, v := range hdr {
		if k == "Host" {
			req.Host = v
			continue
		}
		req.Header.Set(k, v)
	}
	resp, err := c.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, u, err)
	}
	defer func() { _ = resp.Body.Close() }()
	b, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(b)
}

func TestAcc16WebCLIParityAndSecurity(t *testing.T) {
	recordEvidence(t, "acc-16")
	requireLinkPrereqs(t)
	e := newEnv(t)
	root := filepath.Join(e.Root, "work", hostileName)
	mkdirs(t, root, "proj")
	writeLinkConfig(t, e, linkConfig{Root: root, SeedPath: "proj", MaxDepth: 3})
	u := startWeb(t, e)
	base := "http://" + u.Host

	host, _, err := net.SplitHostPort(u.Host)
	if err != nil || !net.ParseIP(host).IsLoopback() {
		t.Errorf("web bound to %q, want a loopback address", u.Host)
	}
	if b, err := os.ReadFile(filepath.Join(e.Root, "state", "web.url")); err != nil || !strings.Contains(string(b), u.Host) {
		t.Errorf("state web.url = %q (err %v), want it to name %s", b, err, u.Host)
	}

	anon := &http.Client{CheckRedirect: noRedirect}
	if code, _ := do(t, anon, "GET", base+"/api/status", nil, ""); code != http.StatusUnauthorized && code != http.StatusForbidden {
		t.Errorf("GET /api/status without session = %d, want 401 or 403", code)
	}
	if code, _ := do(t, anon, "GET", base+"/auth?token=wrong", nil, ""); code != http.StatusForbidden {
		t.Errorf("GET /auth with a wrong token = %d, want 403", code)
	}

	jar, _ := cookiejar.New(nil)
	c := &http.Client{Jar: jar, CheckRedirect: noRedirect}
	if code, _ := do(t, c, "GET", u.String(), nil, ""); code >= 400 {
		t.Fatalf("GET bootstrap URL = %d, want a session", code)
	}
	if code, _ := do(t, c, "GET", u.String(), nil, ""); code != http.StatusForbidden {
		t.Errorf("reusing the bootstrap token = %d, want 403", code)
	}

	if code, _ := do(t, c, "GET", base+"/api/config", map[string]string{"Host": "evil.example"}, ""); code < 400 {
		t.Errorf("GET /api/config with Host evil.example = %d, want 4xx", code)
	}
	cross := map[string]string{"Origin": "http://evil.example", "Content-Type": "application/json"}
	if code, _ := do(t, c, "PUT", base+"/api/config", cross, `{}`); code != http.StatusForbidden {
		t.Errorf("cross-origin PUT /api/config = %d, want 403", code)
	}
	if code, _ := do(t, c, "POST", base+"/api/restore", map[string]string{"Content-Type": "application/json"}, `{}`); code != http.StatusForbidden {
		t.Errorf("POST /api/restore without CSRF token = %d, want 403", code)
	}

	for _, p := range []string{"../../etc/passwd", "/etc/passwd"} {
		q := base + "/api/history?root=work&path=" + url.QueryEscape(p)
		if code, body := do(t, c, "GET", q, nil, ""); code != http.StatusBadRequest {
			t.Errorf("GET /api/history path=%q = %d %q, want 400", p, code, body)
		}
	}

	code, body := do(t, c, "GET", base+"/api/status", nil, "")
	if code != http.StatusServiceUnavailable || !strings.Contains(body, `"prerequisite_missing"`) {
		t.Errorf("GET /api/status with daemon down = %d %q, want 503 prerequisite_missing", code, body)
	}

	for _, page := range []string{"/setup", "/config"} {
		code, html := do(t, c, "GET", base+page, nil, "")
		if code != http.StatusOK {
			t.Errorf("GET %s = %d, want 200", page, code)
			continue
		}
		for _, v := range []string{"img src=x", "img%20src", filepath.Join(e.Root, "work")} {
			if strings.Contains(html, v) {
				t.Errorf("GET %s HTML contains config value %q, want config served only by /api/config", page, v)
			}
		}
	}

	req, err := http.NewRequestWithContext(t.Context(), "GET", base+"/api/config", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	resp, err := c.Do(req)
	if err != nil {
		t.Fatalf("GET /api/config: %v", err)
	}
	cfgBody, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Errorf("GET /api/config Content-Type = %q, want application/json", ct)
	}
	if got := resp.Header.Get("X-Content-Type-Options"); got != "nosniff" {
		t.Errorf("GET /api/config X-Content-Type-Options = %q, want nosniff", got)
	}
	var cfg any
	if err := json.Unmarshal(cfgBody, &cfg); err != nil {
		t.Errorf("GET /api/config body %q is not JSON: %v", cfgBody, err)
	}
	if !jsonStringContains(cfg, hostileName) {
		t.Errorf("GET /api/config body %q, want the root name %q inside a JSON string", cfgBody, hostileName)
	}
	if bytes.Contains(cfgBody, []byte("&lt;img")) {
		t.Errorf("GET /api/config body %q HTML-escapes the root name, want it JSON-encoded only", cfgBody)
	}

	startDaemon(t, e)
	var upCode int
	var upBody string
	_, ok := waitFor(20*time.Second, func() bool {
		upCode, upBody = do(t, c, "GET", base+"/api/status", nil, "")
		return upCode == http.StatusOK
	})
	if !ok {
		skip(t, "missing prerequisite: daemon did not reach ready within 20s, last status "+upBody)
	}
	t.Logf("evidence: /api/status with daemon up = %d %s", upCode, upBody)
}

// jsonStringContains reports whether any string in the decoded JSON value v contains sub.
func jsonStringContains(v any, sub string) bool {
	switch x := v.(type) {
	case string:
		return strings.Contains(x, sub)
	case []any:
		for _, e := range x {
			if jsonStringContains(e, sub) {
				return true
			}
		}
	case map[string]any:
		for _, e := range x {
			if jsonStringContains(e, sub) {
				return true
			}
		}
	}
	return false
}
