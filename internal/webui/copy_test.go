package webui

import (
	"regexp"
	"strings"
	"testing"
)

// templateActionRE matches a {{...}} template action.
var templateActionRE = regexp.MustCompile(`\{\{.*?\}\}`)

// assetPathRE matches served asset paths, which may contain words the copy
// rules otherwise ban.
var assetPathRE = regexp.MustCompile(`/(?:assets|fonts)/[^"'\s)]*`)

// uiText returns every rendered fixture page plus app.js.
func uiText(t *testing.T) string {
	t.Helper()
	var b strings.Builder
	for _, out := range renderedPages(t) {
		b.WriteString(out)
		b.WriteString("\n")
	}
	b.WriteString(appJS(t))
	if strings.TrimSpace(b.String()) == "" {
		t.Fatal("uiText() = empty, want rendered pages and app.js")
	}
	return b.String()
}

func TestNoBannedWordsInUI(t *testing.T) {
	text := strings.ToLower(assetPathRE.ReplaceAllString(uiText(t), ""))
	banned := []*regexp.Regexp{
		regexp.MustCompile(`production-ready`),
		regexp.MustCompile(`cross-platform`),
		regexp.MustCompile(`\bstatic\b`),
		regexp.MustCompile(`finder[- ]integrated`),
	}
	for _, re := range banned {
		if m := re.FindString(text); m != "" {
			t.Errorf("UI text contains banned %q", m)
		}
	}
}

func TestIdenticalOnlyAsLikely(t *testing.T) {
	sources := map[string]string{"assets/app.js": appJS(t)}
	for _, f := range embeddedFiles(t) {
		if !strings.HasPrefix(f, "templates/") {
			continue
		}
		b, err := embedded.ReadFile(f)
		if err != nil {
			t.Fatalf("embedded.ReadFile(%q) error = %v", f, err)
		}
		// Template actions are code, not copy (e.g. .LikelyIdentical).
		sources[f] = templateActionRE.ReplaceAllString(string(b), "")
	}
	history := render(t, mustLoad(t, ""), "history", historyView())
	if !strings.Contains(history, "likely identical") {
		t.Fatal(`Render(history) with a likely-identical group lacks "likely identical"`)
	}
	sources["history render"] = history
	re := regexp.MustCompile(`(?i)identical`)
	for name, s := range sources {
		for _, loc := range re.FindAllStringIndex(s, -1) {
			if !strings.HasSuffix(strings.ToLower(s[:loc[0]]), "likely ") {
				t.Errorf("%s: %q not preceded by \"likely \"", name, s[max(0, loc[0]-20):loc[1]])
			}
		}
	}
}

func TestNoCachedContentClaims(t *testing.T) {
	p := mustLoad(t, "")
	status := measuredStatus()
	renders := map[string]string{
		// The measured Prewarm value is view data, not template copy.
		"status":  strings.ReplaceAll(render(t, p, "status", status), status.Prewarm, ""),
		"history": render(t, p, "history", historyView()),
	}
	claims := regexp.MustCompile(`(?i)cached locally|available offline|\binstant(?:ly)?\b|\balways\b`)
	badge := regexp.MustCompile(`<span class="badge-[a-z]+">(?:warm|cold)</span>`)
	temp := regexp.MustCompile(`(?i)\b(?:warm|cold)\b`)
	for name, out := range renders {
		if out == "" {
			t.Fatalf("Render(%s) = empty, want the page", name)
		}
		if m := claims.FindString(out); m != "" {
			t.Errorf("Render(%s) contains content claim %q", name, m)
		}
		if m := temp.FindString(badge.ReplaceAllString(out, "")); m != "" {
			t.Errorf("Render(%s) uses %q outside badge text", name, m)
		}
	}
}
