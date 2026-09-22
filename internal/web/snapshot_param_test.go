package web

import (
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSnapshotParamMustBeFullID wants every handler that reads a snapshot
// parameter to reject anything but a full 64-hex ID with 400
// invalid_configuration, and to let a full ID past validation.
func TestSnapshotParamMustBeFullID(t *testing.T) {
	const msg = "use the full snapshot id"
	tests := []struct {
		name     string
		snapshot string
		want400  bool
	}{
		{name: "short", snapshot: "560c465d", want400: true},
		{name: "traversal", snapshot: "../x", want400: true},
		{name: "uppercase", snapshot: strings.ToUpper(string(idA)), want400: true},
		{name: "full", snapshot: string(idA), want400: false},
	}
	for _, endpoint := range []string{"/history", "/api/history", "/api/download", "/api/restore"} {
		for _, tt := range tests {
			t.Run(endpoint+"/"+tt.name, func(t *testing.T) {
				h := newFakeHistory(t)
				writeFile(t, filepath.Join(h.snapDir(idA), "docs", "report.docx"), "v1")
				if err := os.MkdirAll(filepath.Join(h.live(), "docs"), 0o755); err != nil {
					t.Fatalf("os.MkdirAll() error = %v", err)
				}
				srv, cookie, csrf := newTestServer(t, Options{Backend: &fakeBackend{}, History: h})

				var code int
				var body string
				if endpoint == "/api/restore" {
					var b []byte
					code, _, b = postRestore(t, srv, cookie, csrf,
						map[string]string{"root": "home", "snapshot": tt.snapshot, "file": "docs/report.docx"})
					body = string(b)
				} else {
					q := url.Values{"root": {"home"}, "path": {"docs"}, "snapshot": {tt.snapshot}}
					if endpoint == "/api/download" {
						q.Set("path", "docs/report.docx")
					}
					w := do(t, srv, http.MethodGet, endpoint+"?"+q.Encode(), nil, apiHeader(cookie, ""))
					code, body = w.Code, w.Body.String()
				}

				if !tt.want400 {
					if code == http.StatusBadRequest || code >= http.StatusInternalServerError {
						t.Errorf("%s snapshot=%q: status = %d, want neither 400 nor 5xx (body %q)",
							endpoint, tt.snapshot, code, body)
					}
					return
				}
				if got, want := code, http.StatusBadRequest; got != want {
					t.Fatalf("%s snapshot=%q: status = %d, want %d (body %q)", endpoint, tt.snapshot, got, want, body)
				}
				for _, want := range []string{"invalid_configuration", msg} {
					if !strings.Contains(body, want) {
						t.Errorf("%s snapshot=%q: body %q does not contain %q", endpoint, tt.snapshot, body, want)
					}
				}
			})
		}
	}
}
