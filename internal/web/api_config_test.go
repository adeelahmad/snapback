package web

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/errcode"
)

// apiHeader returns the headers of a logged-in JSON request, carrying csrf
// when it is not empty.
func apiHeader(cookie *http.Cookie, csrf string) http.Header {
	h := http.Header{}
	h.Set("Cookie", cookie.String())
	h.Set("Content-Type", "application/json")
	if csrf != "" {
		h.Set("X-CSRF-Token", csrf)
	}
	return h
}

// decodeJSON decodes w's body into v, failing the test when it is not JSON.
func decodeJSON(t *testing.T, w *httptest.ResponseRecorder, v any) {
	t.Helper()
	if got := w.Header().Get("Content-Type"); !strings.HasPrefix(got, "application/json") {
		t.Fatalf("Content-Type = %q, want application/json (status %d, body %q)", got, w.Code, w.Body.String())
	}
	if err := json.Unmarshal(w.Body.Bytes(), v); err != nil {
		t.Fatalf("json.Unmarshal(%q) error = %v", w.Body.String(), err)
	}
}

type configBody struct {
	Revision config.Revision `json:"revision"`
	Config   *config.Config  `json:"config"`
}

type errorBody struct {
	Code  string `json:"code"`
	Error string `json:"error"`
}

func TestAPIStatusJSON(t *testing.T) {
	b := &fakeBackend{snap: snapshot{State: "degraded", Generation: 4}, cfg: &config.Config{}}
	srv, cookie, _ := newTestServer(t, Options{Backend: b})

	w := do(t, srv, http.MethodGet, "/api/status", nil, apiHeader(cookie, ""))
	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf("GET /api/status: status = %d, want %d", got, want)
	}
	if got, want := w.Header().Get("Cache-Control"), "no-store"; got != want {
		t.Errorf("GET /api/status: Cache-Control = %q, want %q", got, want)
	}
	var got snapshot
	decodeJSON(t, w, &got)
	if want := b.snap; got != want {
		t.Errorf("GET /api/status = %+v, want %+v", got, want)
	}
}

func TestAPIConfigGetAndSaveWithRevision(t *testing.T) {
	cfg := &config.Config{Version: 1, Roots: []config.Root{{ID: "home", LocalPath: "/home/u"}}}
	b := &fakeBackend{cfg: cfg, rev: "r1", newRev: "r2"}
	srv, cookie, csrf := newTestServer(t, Options{Backend: b})

	w := do(t, srv, http.MethodGet, "/api/config", nil, apiHeader(cookie, ""))
	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf("GET /api/config: status = %d, want %d", got, want)
	}
	var got configBody
	decodeJSON(t, w, &got)
	if got.Revision != "r1" {
		t.Errorf("GET /api/config revision = %q, want %q", got.Revision, "r1")
	}
	if got.Config == nil || len(got.Config.Roots) != 1 || got.Config.Roots[0].ID != "home" {
		t.Errorf("GET /api/config config = %+v, want roots [home]", got.Config)
	}

	edited := *cfg
	edited.Roots = append([]config.Root{}, cfg.Roots...)
	edited.Roots = append(edited.Roots, config.Root{ID: "work", LocalPath: "/srv/work"})
	body, err := json.Marshal(configBody{Revision: "r1", Config: &edited})
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	w = do(t, srv, http.MethodPut, "/api/config", strings.NewReader(string(body)), apiHeader(cookie, csrf))
	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf("PUT /api/config: status = %d, want %d (body %q)", got, want, w.Body.String())
	}
	var put configBody
	decodeJSON(t, w, &put)
	if put.Revision != "r2" {
		t.Errorf("PUT /api/config revision = %q, want %q", put.Revision, "r2")
	}
	if len(b.saved) != 1 {
		t.Fatalf("SaveConfig calls = %d, want 1", len(b.saved))
	}
	if got := b.saved[0].Roots; len(got) != 2 || got[1].ID != "work" {
		t.Errorf("SaveConfig config roots = %+v, want home and work", got)
	}
	if got, want := b.gotRevs[0], config.Revision("r1"); got != want {
		t.Errorf("SaveConfig rev = %q, want %q", got, want)
	}
}

func TestAPIConfigStaleRevisionIs409(t *testing.T) {
	b := &fakeBackend{cfg: &config.Config{}, rev: "r2", saveErr: config.ErrRevisionConflict}
	srv, cookie, csrf := newTestServer(t, Options{Backend: b})

	w := do(t, srv, http.MethodPut, "/api/config", strings.NewReader(`{"revision":"r1","config":{}}`), apiHeader(cookie, csrf))
	if got, want := w.Code, http.StatusConflict; got != want {
		t.Errorf("PUT /api/config stale: status = %d, want %d", got, want)
	}
	var e errorBody
	decodeJSON(t, w, &e)
	if got, want := e.Code, string(errcode.StaleState); got != want {
		t.Errorf("PUT /api/config stale: code = %q, want %q", got, want)
	}

	calls := len(b.saved)
	w = do(t, srv, http.MethodPut, "/api/config", strings.NewReader(`{not json`), apiHeader(cookie, csrf))
	if got, want := w.Code, http.StatusBadRequest; got != want {
		t.Errorf("PUT /api/config bad JSON: status = %d, want %d", got, want)
	}
	e = errorBody{}
	decodeJSON(t, w, &e)
	if got, want := e.Code, string(errcode.InvalidConfig); got != want {
		t.Errorf("PUT /api/config bad JSON: code = %q, want %q", got, want)
	}
	if got := len(b.saved); got != calls {
		t.Errorf("SaveConfig calls after bad JSON = %d, want %d", got, calls)
	}
}

func TestAPISetupValidateAndIntegrations(t *testing.T) {
	v := &fakeValidator{errs: []error{errcode.New(errcode.RepoUnavailable, "restic.validate", errors.New("repository not found"))}}
	opts := Options{Backend: &fakeBackend{cfg: &config.Config{}, rev: "r1"}}
	opts.setValidator(v)
	srv, cookie, csrf := newTestServer(t, opts)

	type result struct {
		OK   bool   `json:"ok"`
		Code string `json:"code"`
	}
	wants := []result{{OK: false, Code: string(errcode.RepoUnavailable)}, {OK: true}}
	for i, want := range wants {
		w := do(t, srv, http.MethodPost, "/api/setup/validate", strings.NewReader(`{"config":{}}`), apiHeader(cookie, csrf))
		if got, wantCode := w.Code, http.StatusOK; got != wantCode {
			t.Fatalf("POST /api/setup/validate #%d: status = %d, want %d (body %q)", i+1, got, wantCode, w.Body.String())
		}
		var got result
		decodeJSON(t, w, &got)
		if got != want {
			t.Errorf("POST /api/setup/validate #%d = %+v, want %+v", i+1, got, want)
		}
	}

	w := do(t, srv, http.MethodGet, "/api/integrations", nil, apiHeader(cookie, ""))
	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf("GET /api/integrations: status = %d, want %d", got, want)
	}
	var integrations map[string]any
	decodeJSON(t, w, &integrations)

	bare, bareCookie, bareCSRF := newTestServer(t, Options{Backend: &fakeBackend{cfg: &config.Config{}}})
	w = do(t, bare, http.MethodPost, "/api/setup/validate", strings.NewReader(`{"config":{}}`), apiHeader(bareCookie, bareCSRF))
	if got, want := w.Code, http.StatusServiceUnavailable; got != want {
		t.Errorf("POST /api/setup/validate without Validator: status = %d, want %d", got, want)
	}
	var e errorBody
	decodeJSON(t, w, &e)
	if got, want := e.Code, string(errcode.PrereqMissing); got != want {
		t.Errorf("POST /api/setup/validate without Validator: code = %q, want %q", got, want)
	}
}
