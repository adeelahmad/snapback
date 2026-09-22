package webui

import (
	"bytes"
	"io"
	"io/fs"
	"net/http"
	"regexp"
	"slices"
	"strings"
	"testing"
)

// tourJS returns the embedded assets/tour.js or fails the test when it is
// missing or empty (M-002).
func tourJS(t *testing.T) string {
	t.Helper()
	b, err := fs.ReadFile(embedded, "assets/tour.js")
	if err != nil {
		t.Fatalf("fs.ReadFile(embedded, %q) error = %v, want the tour script", "assets/tour.js", err)
	}
	if len(strings.TrimSpace(string(b))) == 0 {
		t.Fatal("assets/tour.js is empty, want the tour dismiss script")
	}
	return string(b)
}

func TestTourJSRemembersDismissal(t *testing.T) {
	js := tourJS(t)
	for _, want := range []string{"snapback.tour.dismissed", "localStorage.getItem(", "localStorage.setItem("} {
		if !strings.Contains(js, want) {
			t.Errorf("assets/tour.js does not contain %q, want the dismissal read from and written to localStorage", want)
		}
	}
	guard := regexp.MustCompile(`try\s*\{[^{}]*localStorage[\s\S]*?\}\s*catch\s*[({]`)
	if !guard.MatchString(js) {
		t.Error("assets/tour.js has no try { ... localStorage ... } catch around storage access, want every storage call guarded")
	}
	for _, m := range regexp.MustCompile(`localStorage\.\w+\(`).FindAllStringIndex(js, -1) {
		before := js[:m[0]]
		open := strings.LastIndex(before, "try {")
		if open < 0 {
			t.Errorf("assets/tour.js has localStorage access at offset %d outside any try block, want it wrapped", m[0])
			continue
		}
		if catchAt := strings.Index(js[open:], "catch"); catchAt >= 0 && open+catchAt < m[0] {
			t.Errorf("assets/tour.js has localStorage access at offset %d after the catch of its try block, want it inside the try", m[0])
		}
	}
}

func TestTourJSHidesTheTourOnDismiss(t *testing.T) {
	js := tourJS(t)
	for _, want := range []string{`[data-js="tour-dismiss"]`, "addEventListener('click'", "[data-tour]"} {
		if !strings.Contains(js, want) {
			t.Errorf("assets/tour.js does not contain %q, want the dismiss button wired to the tour list", want)
		}
	}
	if !strings.Contains(js, "hidden = true") && !strings.Contains(js, "setAttribute('hidden'") {
		t.Error("assets/tour.js never hides the tour, want hidden = true or setAttribute('hidden', ...) on [data-tour]")
	}
	tour, err := fs.ReadFile(embedded, "templates/tour.html")
	if err != nil {
		t.Fatalf("fs.ReadFile(embedded, %q) error = %v, want the tour partial", "templates/tour.html", err)
	}
	var hooks []string
	for _, m := range regexp.MustCompile(`data-js=["']?([A-Za-z0-9_-]+)["']?`).FindAllStringSubmatch(js, -1) {
		if !slices.Contains(hooks, m[1]) {
			hooks = append(hooks, m[1])
		}
	}
	if !slices.Contains(hooks, "tour-dismiss") {
		t.Errorf("assets/tour.js data-js hooks = %v, want %q among them", hooks, "tour-dismiss")
	}
	for _, h := range hooks {
		if !regexp.MustCompile(`data-js=["']?` + regexp.QuoteMeta(h) + `["'\s>]`).Match(tour) {
			t.Errorf("templates/tour.html lacks data-js=%q, want every hook tour.js queries present", h)
		}
	}
}

func TestTourJSServedAssetMatchesEmbed(t *testing.T) {
	want := []byte(tourJS(t))
	resp := serve(t, mustLoad(t, "").Static(), "/assets/tour.js", nil)
	if got, want := resp.StatusCode, http.StatusOK; got != want {
		t.Fatalf("Static() GET /assets/tour.js = %d, want %d", got, want)
	}
	if got, want := resp.Header.Get("Content-Type"), "text/javascript"; !strings.HasPrefix(got, want) {
		t.Errorf("Static() GET /assets/tour.js Content-Type = %q, want %q", got, want)
	}
	got, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("io.ReadAll(/assets/tour.js) error = %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("served /assets/tour.js = %d bytes, want the %d embedded bytes", len(got), len(want))
	}
	served := string(got)
	for _, banned := range []string{"import ", "require(", "export ", "<script", "innerHTML", "eval("} {
		if strings.Contains(served, banned) {
			t.Errorf("served /assets/tour.js contains %q, want a bundler-free injection-free plain script", banned)
		}
	}
	for _, want := range []string{"snapback.tour.dismissed", `[data-js="tour-dismiss"]`, "[data-tour]"} {
		if !strings.Contains(served, want) {
			t.Errorf("served /assets/tour.js does not contain %q, want the dismiss handler served to the browser", want)
		}
	}
}

func TestTourJSLoadedWhereTheTourRenders(t *testing.T) {
	var hosts []string
	for _, f := range embeddedFiles(t) {
		if !strings.HasPrefix(f, "templates/") || f == "templates/tour.html" {
			continue
		}
		b, err := fs.ReadFile(embedded, f)
		if err != nil {
			t.Fatalf("fs.ReadFile(embedded, %q) error = %v", f, err)
		}
		if regexp.MustCompile(`\{\{\s*template\s+"tour"`).Match(b) {
			hosts = append(hosts, f)
		}
	}
	if len(hosts) == 0 {
		t.Skip(`no template includes {{template "tour"}} yet, so nothing must load /assets/tour.js`)
	}
	script := regexp.MustCompile(`<script\b[^>]*\bsrc="/assets/tour\.js"`)
	for _, f := range append(hosts, "templates/layout.html") {
		b, err := fs.ReadFile(embedded, f)
		if err != nil {
			t.Fatalf("fs.ReadFile(embedded, %q) error = %v", f, err)
		}
		if script.Match(b) {
			return
		}
	}
	t.Errorf(`templates including the tour (%v) and templates/layout.html lack <script src="/assets/tour.js">, want the tour script loaded`, hosts)
}
