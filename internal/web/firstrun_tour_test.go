package web

import (
	"net/http"
	"regexp"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/webui"
)

// tourList matches the tour partial's ordered list, however its attributes
// are ordered.
var tourList = regexp.MustCompile(`(?s)<ol\b[^>]*\bdata-tour\b[^>]*>`)

// tourScript matches the <script src="/assets/tour.js"> the tour needs to be
// dismissable.
var tourScript = regexp.MustCompile(`<script\b[^>]*\bsrc="/assets/tour\.js"`)

// tourStepFor captures the field each tour step points at.
var tourStepFor = regexp.MustCompile(`data-tour-for="([^"]*)"`)

// getPage serves a GET for path against srv as the logged-in session and
// returns the body, failing the test unless the page rendered.
func getPage(t *testing.T, srv *Server, cookie *http.Cookie, path string) string {
	t.Helper()
	w := do(t, srv, http.MethodGet, path, nil, http.Header{"Cookie": {cookie.String()}})
	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf("GET %s: status = %d, want %d", path, got, want)
	}
	return w.Body.String()
}

// tourServer returns a server whose Backend reports rev as the stored
// config's revision; the zero revision is a machine with no config yet.
func tourServer(t *testing.T, rev config.Revision) (*Server, *http.Cookie) {
	t.Helper()
	srv, cookie, _ := newTestServer(t, Options{Backend: &fakeBackend{cfg: &config.Config{}, rev: rev}})
	return srv, cookie
}

func TestFirstRunTourOnSetupBeforeAnyConfig(t *testing.T) {
	srv, cookie := tourServer(t, "")

	body := getPage(t, srv, cookie, "/setup")

	if !tourList.MatchString(body) {
		t.Errorf("GET /setup on first run: body has no <ol ... data-tour ...>, want the first-run tour list")
	}
	if !strings.Contains(body, `data-js="tour-dismiss"`) {
		t.Errorf("GET /setup on first run: body has no data-js=%q control, want the tour dismiss button", "tour-dismiss")
	}
	if !tourScript.MatchString(body) {
		t.Errorf(`GET /setup on first run: body has no <script src="/assets/tour.js">, want the dismiss script loaded`)
	}
}

func TestFirstRunTourVisibility(t *testing.T) {
	cases := []struct {
		name string
		rev  config.Revision
		path string
		want bool
	}{
		{name: "setup with no config", rev: "", path: "/setup", want: true},
		{name: "setup once a config exists", rev: "r7", path: "/setup", want: false},
		{name: "config page on first run", rev: "", path: "/config", want: false},
		{name: "config page once a config exists", rev: "r7", path: "/config", want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv, cookie := tourServer(t, tc.rev)

			body := getPage(t, srv, cookie, tc.path)

			if got := tourList.MatchString(body); got != tc.want {
				t.Errorf("GET %s at revision %q: tour present = %t, want %t", tc.path, tc.rev, got, tc.want)
			}
			if got := strings.Contains(body, `data-js="tour-dismiss"`); got != tc.want {
				t.Errorf("GET %s at revision %q: tour dismiss button present = %t, want %t", tc.path, tc.rev, got, tc.want)
			}
		})
	}
}

func TestFirstRunTourStepsAnchorToSetupFields(t *testing.T) {
	srv, cookie := tourServer(t, "")

	body := getPage(t, srv, cookie, "/setup")

	var got []string
	for _, m := range tourStepFor.FindAllStringSubmatch(body, -1) {
		got = append(got, m[1])
	}
	if want := len(webui.SetupTour()); len(got) != want {
		t.Fatalf("GET /setup on first run: data-tour-for values = %v (%d), want the %d steps of webui.SetupTour()", got, len(got), want)
	}
	for _, id := range got {
		if !strings.Contains(body, `id="`+id+`"`) {
			t.Errorf("GET /setup on first run: tour step points at data-tour-for=%q, but the page has no id=%q", id, id)
		}
	}
}
