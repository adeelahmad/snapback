package web

import (
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/config"
)

// daemonStatusPill matches the daemon control's status element and captures
// its text, so the page tests read the same pill the partial renders.
var daemonStatusPill = regexp.MustCompile(`(?is)<[a-z][a-z0-9]*\b[^>]*\bdata-js="daemon-status"[^>]*>([^<]*)</[a-z][a-z0-9]*>`)

// daemonStatusForm captures the body of the status page form that posts to
// action.
func daemonStatusForm(t *testing.T, body, action string) string {
	t.Helper()
	re := regexp.MustCompile(`(?is)<form\b[^>]*\baction="` + regexp.QuoteMeta(action) + `"[^>]*>(.*?)</form>`)
	m := re.FindStringSubmatch(body)
	if m == nil {
		t.Fatalf("GET /status has no <form action=%q>, want the JS-off %s form\npage: %s", action, action, body)
	}
	return m[1]
}

// statusPage fetches /status as the session behind cookie.
func statusPage(t *testing.T, srv *Server, cookie *http.Cookie) string {
	t.Helper()
	w := do(t, srv, http.MethodGet, "/status", nil, http.Header{"Cookie": {cookie.String()}})
	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf("GET /status status = %d, want %d (body %q)", got, want, w.Body.String())
	}
	return w.Body.String()
}

func TestStatusPageDaemonStatusShowsControl(t *testing.T) {
	tests := []struct {
		name    string
		running bool
		want    string
	}{
		{"running", true, "running"},
		{"stopped", false, "stopped"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &fakeDaemon{running: tt.running}
			srv, cookie, _ := newTestServer(t, Options{Backend: &fakeBackend{cfg: &config.Config{}}, Daemon: d})

			body := statusPage(t, srv, cookie)

			pills := daemonStatusPill.FindAllStringSubmatch(body, -1)
			if got, want := len(pills), 1; got != want {
				t.Fatalf(`GET /status with Options.Daemon: elements with data-js="daemon-status" = %d, want %d`+"\npage: %s", got, want, body)
			}
			if got := strings.TrimSpace(html.UnescapeString(pills[0][1])); got != tt.want {
				t.Errorf("GET /status daemon status pill text = %q, want %q (daemon Running = %t)", got, tt.want, tt.running)
			}
		})
	}
}

func TestStatusPageDaemonStatusFormsCarrySessionCSRF(t *testing.T) {
	d := &fakeDaemon{}
	srv, cookie, csrf := newTestServer(t, Options{Backend: &fakeBackend{cfg: &config.Config{}}, Daemon: d})

	body := statusPage(t, srv, cookie)

	for _, action := range []string{"/api/daemon/start", "/api/daemon/stop"} {
		form := daemonStatusForm(t, body, action)
		m := regexp.MustCompile(`(?is)<input\b[^>]*\bname="` + csrfField + `"[^>]*>`).FindString(form)
		if m == "" {
			t.Errorf("GET /status form %s has no <input name=%q>, want the JS-off token\nform: %s", action, csrfField, form)
			continue
		}
		v := regexp.MustCompile(`(?is)\bvalue="([^"]*)"`).FindStringSubmatch(m)
		if v == nil {
			t.Errorf("GET /status form %s csrf input = %q, want a value attribute", action, m)
			continue
		}
		if got := html.UnescapeString(v[1]); got != csrf {
			t.Errorf("GET /status form %s csrf token = %q, want the session token %q", action, got, csrf)
		}
	}
}

func TestStatusPageDaemonStatusAbsentWithoutControl(t *testing.T) {
	// Control: the same page with a daemon seam must carry the pill, so the
	// absence below pins the nil case and not a page that never shows it.
	withD, withCookie, _ := newTestServer(t, Options{Backend: &fakeBackend{cfg: &config.Config{}}, Daemon: &fakeDaemon{}})
	if body := statusPage(t, withD, withCookie); !daemonStatusPill.MatchString(body) {
		t.Errorf(`GET /status with Options.Daemon has no data-js="daemon-status", want the daemon control`+"\npage: %s", body)
	}

	srv, cookie, _ := newTestServer(t, Options{Backend: &fakeBackend{cfg: &config.Config{}}})
	body := statusPage(t, srv, cookie)
	if daemonStatusPill.MatchString(body) {
		t.Errorf(`GET /status with Options.Daemon == nil contains data-js="daemon-status", want no daemon control`+"\npage: %s", body)
	}
	for _, action := range []string{"/api/daemon/start", "/api/daemon/stop"} {
		if strings.Contains(body, `action="`+action+`"`) {
			t.Errorf("GET /status with Options.Daemon == nil contains a form posting to %s, want none", action)
		}
	}
}

func TestDaemonStatusFormPostRedirectsToStatus(t *testing.T) {
	tests := []struct {
		name    string
		target  string
		running bool
		want    bool
		count   func(*fakeDaemon) int
	}{
		{"start", "/api/daemon/start", false, true, func(d *fakeDaemon) int { starts, _ := d.counts(); return starts }},
		{"stop", "/api/daemon/stop", true, false, func(d *fakeDaemon) int { _, stops := d.counts(); return stops }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &fakeDaemon{running: tt.running}
			srv, cookie, csrf := newTestServer(t, Options{Backend: &fakeBackend{cfg: &config.Config{}}, Daemon: d})

			form := url.Values{csrfField: {csrf}}
			w := do(t, srv, http.MethodPost, tt.target, strings.NewReader(form.Encode()), formHeader(cookie))

			if got, want := w.Code, http.StatusSeeOther; got != want {
				t.Errorf("POST %s as an HTML form: status = %d, want %d (the JS-off fallback redirects)\nbody: %s", tt.target, got, want, w.Body.String())
			}
			if got, want := w.Header().Get("Location"), "/status"; got != want {
				t.Errorf("POST %s as an HTML form: Location = %q, want %q", tt.target, got, want)
			}
			if got, want := tt.count(d), 1; got != want {
				t.Errorf("POST %s as an HTML form: control calls = %d, want %d", tt.target, got, want)
			}

			// Control: a JSON caller must still get the 200 state envelope,
			// so the redirect above is the form case and not every POST.
			jd := &fakeDaemon{running: tt.running}
			jsrv, jcookie, jcsrf := newTestServer(t, Options{Backend: &fakeBackend{cfg: &config.Config{}}, Daemon: jd})
			jw := do(t, jsrv, http.MethodPost, tt.target, nil, apiHeader(jcookie, jcsrf))
			if got, want := jw.Code, http.StatusOK; got != want {
				t.Fatalf("POST %s as JSON: status = %d, want %d (JSON callers keep the body)\nbody: %s", tt.target, got, want, jw.Body.String())
			}
			if got := jw.Header().Get("Location"); got != "" {
				t.Errorf("POST %s as JSON: Location = %q, want no redirect", tt.target, got)
			}
			var state daemonState
			decodeJSON(t, jw, &state)
			if state.Running != tt.want {
				t.Errorf("POST %s as JSON = %+v, want running %t", tt.target, state, tt.want)
			}
		})
	}
}

func TestDaemonStatusServedAppJSPolls(t *testing.T) {
	srv, cookie, _ := newTestServer(t, Options{Backend: &fakeBackend{cfg: &config.Config{}}, Daemon: &fakeDaemon{}})

	w := do(t, srv, http.MethodGet, "/assets/app.js", nil, http.Header{"Cookie": {cookie.String()}})
	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf("GET /assets/app.js status = %d, want %d", got, want)
	}
	b, err := io.ReadAll(w.Body)
	if err != nil {
		t.Fatalf("io.ReadAll(/assets/app.js) error = %v", err)
	}
	js := string(b)

	register := regexp.MustCompile(`document\.querySelectorAll\(\s*['"]\[data-js="daemon-status"\]['"]\s*\)\.forEach\(\s*(\w+)\s*\)`)
	m := register.FindStringSubmatch(js)
	if m == nil {
		t.Fatalf(`served /assets/app.js does not register [data-js="daemon-status"], want document.querySelectorAll('[data-js="daemon-status"]').forEach(<handler>)`)
	}
	body := jsFuncSource(js, m[1])
	if body == "" {
		t.Fatalf("served /assets/app.js does not define function %s, want the daemon-status handler", m[1])
	}
	// The poller must go through the CSRF fetch wrapper, so the regex accepts
	// csrfFetch('/api/daemon') as well as a bare fetch('/api/daemon').
	if !regexp.MustCompile(`(?i)fetch\(\s*['"]/api/daemon['"]`).MatchString(body) {
		t.Errorf("%s does not fetch('/api/daemon'), want the pill polled from GET /api/daemon\nhandler: %s", m[1], body)
	}
	for _, want := range []string{"setInterval(", "textContent"} {
		if !strings.Contains(body, want) {
			t.Errorf("%s does not contain %q, want the pill text refreshed on a poll\nhandler: %s", m[1], want, body)
		}
	}
}

// jsFuncSource returns the source of the top-level JS function named name, or
// "" when js does not define it.
func jsFuncSource(js, name string) string {
	start := regexp.MustCompile(`(?m)^(?:async\s+)?function\s+` + regexp.QuoteMeta(name) + `\s*\(`).FindStringIndex(js)
	if start == nil {
		return ""
	}
	body := js[start[0]:]
	if end := strings.Index(body, "\n}"); end >= 0 {
		body = body[:end+2]
	}
	return body
}
