package web

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/adeelahmad/snapback/internal/webui"
)

// newTestServer builds a server on 127.0.0.1:0 from opts, fills Listen,
// StateDir, Token and Pages, and logs in through a real /auth exchange. It
// returns the server, the session cookie and the session's CSRF token.
func newTestServer(t *testing.T, opts Options) (*Server, *http.Cookie, string) {
	t.Helper()
	pages, err := webui.Load("")
	if err != nil {
		t.Fatalf("webui.Load(%q) error = %v", "", err)
	}
	opts.Listen = "127.0.0.1:0"
	opts.StateDir = t.TempDir()
	opts.Token = "tok"
	opts.Pages = pages
	if opts.Stdout == nil {
		opts.Stdout = &bytes.Buffer{}
	}
	srv, err := New(opts)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	t.Cleanup(func() { _ = srv.Close() })
	cookie, csrf := login(t, srv, hostOf(t, srv))
	return srv, cookie, csrf
}

// do serves one request against srv with the Host set to the server's own
// host and hdr copied onto the request.
func do(t *testing.T, srv *Server, method, target string, body io.Reader, hdr http.Header) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(method, target, body)
	r.Host = hostOf(t, srv)
	for k, vs := range hdr {
		for _, v := range vs {
			r.Header.Add(k, v)
		}
	}
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, r)
	return w
}
