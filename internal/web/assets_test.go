package web

import (
	"net/http"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var assetRef = regexp.MustCompile(`(?:href|src)="(/(?:assets|fonts)/[^"]+)"`)

func TestProductionServesLayoutAssets(t *testing.T) {
	opts := productionOptions(r2bConfig(t), filepath.Join(t.TempDir(), "config.yaml"))
	srv, cookie, _ := newTestServer(t, opts)
	withSession := http.Header{"Cookie": {cookie.String()}}

	page := do(t, srv, http.MethodGet, "/", nil, withSession)
	if page.Code != http.StatusOK {
		t.Fatalf("GET / status = %d, want %d", page.Code, http.StatusOK)
	}
	targets := []string{"/fonts/IBMPlexSans-Regular.woff2", "/assets/icons/activity.svg"}
	for _, m := range assetRef.FindAllStringSubmatch(page.Body.String(), -1) {
		targets = append(targets, m[1])
	}
	wantType := map[string]string{
		".css":   "text/css",
		".js":    "text/javascript",
		".svg":   "image/svg+xml",
		".woff2": "font/woff2",
	}
	for _, hdr := range []http.Header{nil, withSession} {
		for _, target := range targets {
			w := do(t, srv, http.MethodGet, target, nil, hdr)
			if w.Code != http.StatusOK {
				t.Errorf("GET %s (session %t) status = %d, want %d", target, hdr != nil, w.Code, http.StatusOK)
				continue
			}
			got := w.Header().Get("Content-Type")
			if want := wantType[path.Ext(target)]; !strings.HasPrefix(got, want) {
				t.Errorf("GET %s (session %t) Content-Type = %q, want %q", target, hdr != nil, got, want)
			}
		}
	}
}
