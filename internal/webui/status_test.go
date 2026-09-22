package webui

import (
	"regexp"
	"strings"
	"testing"
)

func statusPtr[T any](v T) *T { return &v }

// measuredStatus returns a StatusView with every metric measured.
func measuredStatus() StatusView {
	return StatusView{
		Chrome: Chrome{Title: "Status", Active: "status", CSRFToken: "tok"},
		Mounts: []MountStatus{
			{Name: "home", State: "ready"},
			{Name: "media", State: "degraded"},
		},
		LastRefresh:       statusPtr("2026-09-20 10:00"),
		EligibleSnapshots: statusPtr(12),
		ManagedLinks:      statusPtr(3),
		ThrottleEvents:    statusPtr(0),
		Prewarm:           "warm 3/3",
		DiscoveryMode:     "seeded",
	}
}

func TestStatusRendersMeasuredMetrics(t *testing.T) {
	p := mustLoad(t, "")
	got := render(t, p, "status", measuredStatus())
	for _, want := range []string{
		"home", "ready", "media", "degraded",
		"2026-09-20 10:00", "12", "3", "warm 3/3", "seeded",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("Render(status) output missing %q", want)
		}
	}
	nav := regexp.MustCompile(`<a href="/status"[^>]*aria-current="page"`)
	if !nav.MatchString(got) {
		t.Errorf(`Render(status) Status nav link lacks aria-current="page"`)
	}
}

func TestStatusUnmeasuredNotZero(t *testing.T) {
	p := mustLoad(t, "")
	v := StatusView{
		Chrome:        Chrome{Title: "Status", Active: "status"},
		Mounts:        []MountStatus{{Name: "home", State: "ready"}},
		Prewarm:       "warm 1/1",
		DiscoveryMode: "seeded",
	}
	got := render(t, p, "status", v)
	if n := strings.Count(got, "not measured yet"); n != 4 {
		t.Errorf(`Render(status) has %d "not measured yet", want 4`, n)
	}
	zero := regexp.MustCompile(`>\s*0\s*<`)
	if m := zero.FindString(got); m != "" {
		t.Errorf("Render(status) renders an unmeasured metric as zero: %q", m)
	}
}

func TestStatusEscapesHostileText(t *testing.T) {
	p := mustLoad(t, "")
	v := measuredStatus()
	v.Mounts = []MountStatus{{Name: hostile, State: "ready"}}
	v.Errors = []string{hostile}
	got := render(t, p, "status", v)
	if want := "&lt;script&gt;alert(1)&lt;/script&gt;"; !strings.Contains(got, want) {
		t.Errorf("Render(status) output missing escaped %q", want)
	}
	if strings.Contains(got, "<script>alert(1)") {
		t.Errorf("Render(status) output contains raw <script>alert(1)")
	}
}

func TestStatusWorksWithoutJS(t *testing.T) {
	p := mustLoad(t, "")
	got := render(t, p, "status", measuredStatus())
	for _, want := range []string{"2026-09-20 10:00", ">12<", "warm 3/3", "seeded"} {
		if !strings.Contains(got, want) {
			t.Errorf("Render(status) server HTML missing metric %q", want)
		}
	}
	noscript := regexp.MustCompile(`(?is)<noscript[^>]*>(.*?)</noscript>`)
	for _, m := range noscript.FindAllStringSubmatch(got, -1) {
		body := strings.ToLower(m[1])
		if strings.Contains(body, "javascript") || strings.Contains(body, "js") {
			t.Errorf("Render(status) has a JavaScript-required notice: %q", m[0])
		}
	}
}

// statusHookTexts are the shell hook snippets fed to the Integrations page.
var statusHookTexts = map[string]string{
	"bash": `eval "$(snapback hook bash)"`,
	"zsh":  `eval "$(snapback hook zsh)"`,
	"fish": `snapback hook fish | source`,
}

// statusPreTexts returns the inner text of every <pre> element in html.
func statusPreTexts(html string) []string {
	re := regexp.MustCompile(`(?s)<pre[^>]*>(.*?)</pre>`)
	var out []string
	for _, m := range re.FindAllStringSubmatch(html, -1) {
		out = append(out, m[1])
	}
	return out
}

// statusEscape mirrors html/template's escaping of text content.
func statusEscape(s string) string {
	return strings.NewReplacer(
		"&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&#34;", "'", "&#39;", "+", "&#43;",
	).Replace(s)
}

func TestIntegrationsHonestCopy(t *testing.T) {
	p := mustLoad(t, "")
	v := IntegrationsView{
		Chrome:       Chrome{Title: "Integrations", Active: "integrations"},
		ShellHook:    statusHookTexts,
		ServiceState: "active",
	}
	got := render(t, p, "integrations", v)
	for _, want := range []string{"Finder", "not available on Linux", "not included in v0.1", "active"} {
		if !strings.Contains(got, want) {
			t.Errorf("Render(integrations) output missing %q", want)
		}
	}
	pres := statusPreTexts(got)
	for shell, text := range statusHookTexts {
		want := statusEscape(text)
		found := false
		for _, b := range pres {
			if strings.Contains(b, want) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Render(integrations) %s hook %q not inside a <pre>", shell, want)
		}
	}
	for _, banned := range []string{"Finder-integrated", "Finder integrated", "Finder integration"} {
		if strings.Contains(got, banned) {
			t.Errorf("Render(integrations) claims %q", banned)
		}
	}
}

func TestIntegrationsEscapesHookText(t *testing.T) {
	p := mustLoad(t, "")
	v := IntegrationsView{
		Chrome: Chrome{Title: "Integrations", Active: "integrations"},
		ShellHook: map[string]string{
			"bash": statusHookTexts["bash"],
			"zsh":  hostile,
			"fish": statusHookTexts["fish"],
		},
		ServiceState: "active",
	}
	got := render(t, p, "integrations", v)
	if want := "&lt;script&gt;alert(1)&lt;/script&gt;"; !strings.Contains(got, want) {
		t.Errorf("Render(integrations) output missing escaped %q", want)
	}
	if strings.Contains(got, "<script>alert(1)") {
		t.Errorf("Render(integrations) output contains raw <script>alert(1)")
	}
}
