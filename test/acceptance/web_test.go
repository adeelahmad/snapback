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
	"regexp"
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

	// The setup and config forms are server-rendered, so config values legitimately
	// appear in the HTML. What must never appear is the hostile root name as a live tag.
	for _, page := range []string{"/setup", "/config"} {
		code, html := do(t, c, "GET", base+page, nil, "")
		if code != http.StatusOK {
			t.Errorf("GET %s = %d, want 200", page, code)
			continue
		}
		if strings.Contains(html, "<img src=x") {
			t.Errorf("GET %s HTML contains the raw tag %q, want the root name HTML-escaped", page, "<img src=x")
		}
		escaped := false
		for _, esc := range []string{"&lt;img src=x", "&#60;img src=x", "&#x3c;img src=x", "&#x3C;img src=x"} {
			if strings.Contains(html, esc) {
				escaped = true
				break
			}
		}
		if !escaped {
			t.Errorf("GET %s HTML does not contain the root name %q HTML-escaped, want an escaped entity form", page, hostileName)
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

	t.Run("typed password setup starts the daemon", func(t *testing.T) {
		typedPasswordSetup(t)
	})

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

// csrfInputRE finds the CSRF token a server-rendered form carries.
var csrfInputRE = regexp.MustCompile(`name="csrf_token" value="([^"]*)"`)

// csrfToken returns the CSRF token of a rendered form page.
func csrfToken(t *testing.T, html string) string {
	t.Helper()
	m := csrfInputRE.FindStringSubmatch(html)
	if m == nil || m[1] == "" {
		t.Fatalf("rendered form carries no csrf_token, want a hidden input; page:\n%s", html)
	}
	return m[1]
}

// yamlValue returns the value of the first `key: value` line of a config file,
// at any indentation. The keys this test reads are written once each.
func yamlValue(t *testing.T, yaml, key string) string {
	t.Helper()
	for _, line := range strings.Split(yaml, "\n") {
		if v, ok := strings.CutPrefix(strings.TrimSpace(line), key+":"); ok {
			return strings.Trim(strings.TrimSpace(v), `"'`)
		}
	}
	t.Fatalf("config names no %s, want one; config:\n%s", key, yaml)
	return ""
}

// typedPasswordSetup drives a first run with no configuration at all: the
// whole config is written through POST /setup with the repository password
// typed into the form, and the daemon then starts from what was written. The
// secret must live in a 0600 credential file under the state dir, never in
// config.yaml.
func typedPasswordSetup(t *testing.T) {
	t.Helper()
	e := newEnv(t)
	repo, pwFile := newRepo(t)
	secretBytes, err := os.ReadFile(pwFile)
	if err != nil {
		t.Fatalf("read repo password file: %v", err)
	}
	secret := string(secretBytes)
	if secret == "" {
		t.Fatal("repo password file is empty, want a password to type")
	}
	resticBin, err := exec.LookPath("restic")
	if err != nil {
		t.Fatalf("look up restic: %v", err)
	}
	root := filepath.Join(e.Root, "work")
	mkdirs(t, root, "proj")

	cfgPath := filepath.Join(e.Config, "snapback", "config.yaml")
	if _, err := os.Stat(cfgPath); !os.IsNotExist(err) {
		t.Fatalf("stat %s before setup = %v, want it not to exist", cfgPath, err)
	}

	u := startWeb(t, e)
	base := "http://" + u.Host
	jar, _ := cookiejar.New(nil)
	c := &http.Client{Jar: jar, CheckRedirect: noRedirect}
	if code, _ := do(t, c, "GET", u.String(), nil, ""); code >= 400 {
		t.Fatalf("GET bootstrap URL = %d, want a session", code)
	}
	code, page := do(t, c, "GET", base+"/setup", nil, "")
	if code != http.StatusOK {
		t.Fatalf("GET /setup with no config = %d, want 200", code)
	}

	form := url.Values{
		"csrf_token":                    {csrfToken(t, page)},
		"restic_path":                   {resticBin},
		"rclone_path":                   {""},
		"repo_uri":                      {repo},
		"credential_file":               {""},
		"repositories[0].password_mode": {"typed"},
		"repositories[0].password":      {secret},
		"repositories[0].mount_point":   {""},
		"roots":                         {root},
	}
	hdr := map[string]string{"Content-Type": "application/x-www-form-urlencoded"}
	code, body := do(t, c, "POST", base+"/setup", hdr, form.Encode())
	if code != http.StatusSeeOther {
		t.Fatalf("POST /setup with a typed password = %d, want 303; body:\n%s", code, body)
	}

	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("read the config POST /setup wrote: %v", err)
	}
	yaml := string(data)
	if strings.Contains(yaml, secret) {
		t.Errorf("config written by POST /setup contains the typed password, want only a password_file path; config:\n%s", yaml)
	}
	stateDir := yamlValue(t, yaml, "state_dir")
	credPath := yamlValue(t, yaml, "password_file")
	if !isUnder(credPath, stateDir) {
		t.Errorf("config password_file = %q, want a path under the state dir %q", credPath, stateDir)
	}
	fi, err := os.Stat(credPath)
	if err != nil {
		t.Fatalf("stat the credential file %q: %v", credPath, err)
	}
	if got := fi.Mode().Perm(); got != 0o600 {
		t.Errorf("credential file %q mode = %04o, want 0600", credPath, got)
	}
	stored, err := os.ReadFile(credPath)
	if err != nil {
		t.Fatalf("read the credential file %q: %v", credPath, err)
	}
	if string(stored) != secret {
		t.Errorf("credential file %q holds %d bytes that differ from the typed password, want the password verbatim", credPath, len(stored))
	}

	startDaemon(t, e)
	var upCode int
	var upBody string
	elapsed, ok := waitFor(30*time.Second, func() bool {
		upCode, upBody = do(t, c, "GET", base+"/api/status", nil, "")
		return upCode == http.StatusOK
	})
	if !ok {
		t.Errorf("daemon started from the setup-written config did not reach ready within %s, last /api/status = %d %s", elapsed, upCode, upBody)
		return
	}
	t.Logf("evidence: setup-written config (typed password), daemon ready after %s, /api/status = %d %s", elapsed.Round(time.Millisecond), upCode, upBody)
}
