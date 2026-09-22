package web

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func sessionCookies(w interface{ Result() *http.Response }) []*http.Cookie {
	var got []*http.Cookie
	for _, c := range w.Result().Cookies() {
		if c.Name == "snapback_session" {
			got = append(got, c)
		}
	}
	return got
}

func TestBootstrapTokenExchange(t *testing.T) {
	srv := newGuardServer(t, &guardBackend{})

	w := do(t, srv, http.MethodGet, "/auth?token=tok", nil, nil)

	if got, want := w.Code, http.StatusSeeOther; got != want {
		t.Errorf("GET /auth?token=tok: status = %d, want %d", got, want)
	}
	if got, want := w.Header().Get("Location"), "/"; got != want {
		t.Errorf("GET /auth?token=tok: Location = %q, want %q", got, want)
	}
	if loc := w.Header().Get("Location"); strings.Contains(loc, "token") {
		t.Errorf("GET /auth?token=tok: Location = %q, want no token in it", loc)
	}
	cookies := sessionCookies(w)
	if got, want := len(cookies), 1; got != want {
		t.Fatalf("GET /auth?token=tok: snapback_session cookies = %d, want %d", got, want)
	}
	c := cookies[0]
	if !c.HttpOnly {
		t.Errorf("snapback_session HttpOnly = false, want true")
	}
	if got, want := c.SameSite, http.SameSiteStrictMode; got != want {
		t.Errorf("snapback_session SameSite = %v, want %v", got, want)
	}
	if got, want := c.Path, "/"; got != want {
		t.Errorf("snapback_session Path = %q, want %q", got, want)
	}
	if c.Value == "" || c.Value == "tok" {
		t.Errorf("snapback_session value = %q, want a non-empty value other than the bootstrap token", c.Value)
	}
}

func TestBootstrapTokenSingleUseAndWrongToken(t *testing.T) {
	srv := newGuardServer(t, &guardBackend{})
	first := do(t, srv, http.MethodGet, "/auth?token=tok", nil, nil)
	if got, want := first.Code, http.StatusSeeOther; got != want {
		t.Fatalf("first GET /auth?token=tok: status = %d, want %d", got, want)
	}

	second := do(t, srv, http.MethodGet, "/auth?token=tok", nil, nil)
	if got, want := second.Code, http.StatusForbidden; got != want {
		t.Errorf("second GET /auth?token=tok: status = %d, want %d", got, want)
	}
	if got := second.Header().Values("Set-Cookie"); len(got) != 0 {
		t.Errorf("second GET /auth?token=tok: Set-Cookie = %q, want none", got)
	}

	fresh := newGuardServer(t, &guardBackend{})
	wrong := do(t, fresh, http.MethodGet, "/auth?token=nope", nil, nil)
	if got, want := wrong.Code, http.StatusForbidden; got != want {
		t.Errorf("GET /auth?token=nope: status = %d, want %d", got, want)
	}
	if got := wrong.Header().Values("Set-Cookie"); len(got) != 0 {
		t.Errorf("GET /auth?token=nope: Set-Cookie = %q, want none", got)
	}
}

func TestRoutesRequireSession(t *testing.T) {
	b := &guardBackend{}
	srv := newGuardServer(t, b)

	cookies := []struct {
		name string
		hdr  http.Header
	}{
		{"no cookie", nil},
		{"forged cookie", http.Header{"Cookie": {"snapback_session=abc"}}},
	}
	for _, c := range cookies {
		t.Run(c.name, func(t *testing.T) {
			api := do(t, srv, http.MethodGet, "/api/status", nil, c.hdr)
			if got, want := api.Code, http.StatusUnauthorized; got != want {
				t.Errorf("GET /api/status: status = %d, want %d", got, want)
			}
			if got := api.Header().Get("Content-Type"); !strings.HasPrefix(got, "application/json") {
				t.Errorf("GET /api/status: Content-Type = %q, want application/json", got)
			}
			if got, want := api.Body.String(), `"code":"permission_denied"`; !strings.Contains(got, want) {
				t.Errorf("GET /api/status: body = %q, want it to contain %q", got, want)
			}

			page := do(t, srv, http.MethodGet, "/history", nil, c.hdr)
			if got, want := page.Code, http.StatusUnauthorized; got != want {
				t.Errorf("GET /history: status = %d, want %d", got, want)
			}
			if got := page.Header().Get("Content-Type"); !strings.HasPrefix(got, "text/html") {
				t.Errorf("GET /history: Content-Type = %q, want text/html", got)
			}
			if got := page.Body.String(); strings.Contains(got, "csrf-token") {
				t.Errorf("GET /history: body = %q, want no csrf-token", got)
			}

			asset := do(t, srv, http.MethodGet, "/assets/app.css", nil, c.hdr)
			if got, want := asset.Code, http.StatusOK; got != want {
				t.Errorf("GET /assets/app.css: status = %d, want %d", got, want)
			}
		})
	}
	if b.calls != 0 {
		t.Errorf("Backend calls = %d, want 0", b.calls)
	}
}

func TestWritesRequireCSRF(t *testing.T) {
	b := &guardBackend{}
	srv, cookie, csrf := newTestServer(t, Options{Backend: b})

	writes := []struct {
		method string
		target string
	}{
		{http.MethodPost, "/api/restore"},
		{http.MethodPut, "/api/config"},
	}
	for _, wr := range writes {
		bad := []struct {
			name  string
			token string
		}{
			{"missing token", ""},
			{"wrong token", "not-" + csrf},
		}
		for _, tt := range bad {
			hdr := http.Header{
				"Cookie":       {cookie.String()},
				"Content-Type": {"application/x-www-form-urlencoded"},
			}
			if tt.token != "" {
				hdr.Set("X-CSRF-Token", tt.token)
			}
			w := do(t, srv, wr.method, wr.target, strings.NewReader("root=home"), hdr)
			if got, want := w.Code, http.StatusForbidden; got != want {
				t.Errorf("%s %s %s: status = %d, want %d", wr.method, wr.target, tt.name, got, want)
			}
			if got, want := w.Body.String(), `"code":"permission_denied"`; !strings.Contains(got, want) {
				t.Errorf("%s %s %s: body = %q, want it to contain %q", wr.method, wr.target, tt.name, got, want)
			}
		}

		form := url.Values{"csrf_token": {csrf}, "root": {"home"}}
		hdr := http.Header{
			"Cookie":       {cookie.String()},
			"Content-Type": {"application/x-www-form-urlencoded"},
		}
		w := do(t, srv, wr.method, wr.target, strings.NewReader(form.Encode()), hdr)
		if w.Code == http.StatusForbidden {
			t.Errorf("%s %s with csrf_token form field: status = %d, want not %d", wr.method, wr.target, w.Code, http.StatusForbidden)
		}
	}

	if len(b.saved) != 0 {
		t.Errorf("SaveConfig calls = %d, want 0", len(b.saved))
	}
}
