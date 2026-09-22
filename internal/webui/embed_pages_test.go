package webui

import (
	"net/http"
	"path"
	"regexp"
	"strings"
	"testing"
)

// integrationsFixture is an Integrations view with every field set.
func integrationsFixture() IntegrationsView {
	return IntegrationsView{
		Chrome:         Chrome{Title: "Integrations", Active: "integrations", CSRFToken: "tok-integrations"},
		ShellHook:      map[string]string{"bash": "eval \"$(snapback hook bash)\""},
		ServiceState:   "not installed",
		ServiceInstall: "snapback service install",
	}
}

// renderedPages renders one fixture view per page from the embedded set.
func renderedPages(t *testing.T) map[PageName]string {
	t.Helper()
	p := mustLoad(t, "")
	views := map[PageName]any{
		"setup":        setupFixture(),
		"config":       configFixture(),
		"history":      historyView(),
		"status":       measuredStatus(),
		"integrations": integrationsFixture(),
	}
	out := make(map[PageName]string, len(views))
	for name, v := range views {
		out[name] = render(t, p, name, v)
	}
	return out
}

func TestEmbedFileSetComplete(t *testing.T) {
	files := embeddedFiles(t)
	have := make(map[string]bool, len(files))
	fonts := 0
	for _, f := range files {
		have[f] = true
		if path.Dir(f) == "assets/fonts" && path.Ext(f) == ".woff2" {
			fonts++
		}
	}
	want := []string{
		"templates/layout.html",
		"templates/setup.html",
		"templates/config.html",
		"templates/history.html",
		"templates/status.html",
		"templates/integrations.html",
		"assets/app.css",
		"assets/app.js",
		"assets/tokens.css",
	}
	for _, f := range want {
		if !have[f] {
			t.Errorf("embeddedFiles() missing %q", f)
		}
	}
	if fonts == 0 {
		t.Error("embeddedFiles() has no assets/fonts/*.woff2, want at least one")
	}
}

func TestEmbedRendersEveryPage(t *testing.T) {
	want := []string{
		`lang="en"`,
		`href="#main"`,
		`<main id="main">`,
		`<link rel="stylesheet" href="/assets/tokens.css">`,
		`<link rel="stylesheet" href="/assets/app.css">`,
		`<script type="module" src="/assets/app.js">`,
	}
	for name, out := range renderedPages(t) {
		if !strings.HasPrefix(strings.ToLower(out), "<!doctype html>") {
			t.Errorf("Render(%q) starts %.30q, want <!doctype html>", name, out)
		}
		for _, s := range want {
			if !strings.Contains(out, s) {
				t.Errorf("Render(%q) missing %q", name, s)
			}
		}
		if got := strings.Count(out, "<h1"); got != 1 {
			t.Errorf("Render(%q) has %d <h1>, want 1", name, got)
		}
	}
}

var assetRefRE = regexp.MustCompile(`(?:href|src)="(/(?:assets|fonts)/[^"]+)"`)

func TestEmbedPagesReferenceOnlyServedAssets(t *testing.T) {
	p := mustLoad(t, "")
	refs := map[string]bool{}
	for _, out := range renderedPages(t) {
		for _, m := range assetRefRE.FindAllStringSubmatch(out, -1) {
			refs[m[1]] = true
		}
	}
	if len(refs) == 0 {
		t.Fatal("rendered pages reference no /assets/ or /fonts/ paths, want the layout assets")
	}
	for ref := range refs {
		if got := serve(t, p.Static(), ref, nil).StatusCode; got != http.StatusOK {
			t.Errorf("Static() GET %s = %d, want %d", ref, got, http.StatusOK)
		}
	}
}

// TestHistoryTimelineWorksWithoutJS checks every snapshot on the timeline is
// reachable as a plain link or a GET form, so switching needs no script.
func TestHistoryTimelineWorksWithoutJS(t *testing.T) {
	v := historyView()
	out := render(t, mustLoad(t, ""), "history", v)
	forms := regexp.MustCompile(`(?is)<form[^>]*method="get"[^>]*>.*?</form>`).FindAllString(out, -1)
	for _, tick := range v.Timeline {
		link := regexp.MustCompile(`<a [^>]*href="[^"]*snapshot=` + regexp.QuoteMeta(tick.ID) + `[&"]`)
		if link.MatchString(out) {
			continue
		}
		reachable := false
		for _, f := range forms {
			if strings.Contains(f, `name="snapshot"`) && strings.Contains(f, `value="`+tick.ID+`"`) {
				reachable = true
			}
		}
		if !reachable {
			t.Errorf("Render(history) snapshot %q has no plain link (href ...snapshot=%s) or GET form, want it reachable without JS", tick.ID, tick.ID)
		}
	}
}
