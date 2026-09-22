package webui

import (
	"io/fs"
	"regexp"
	"slices"
	"strings"
	"testing"
)

// appJS returns the embedded assets/app.js or fails the test when it is
// missing or empty (M-002).
func appJS(t *testing.T) string {
	t.Helper()
	b, err := fs.ReadFile(embedded, "assets/app.js")
	if err != nil {
		t.Fatalf("fs.ReadFile(embedded, %q) error = %v, want the app.js module", "assets/app.js", err)
	}
	if len(strings.TrimSpace(string(b))) == 0 {
		t.Fatal("assets/app.js is empty, want the app.js module")
	}
	return string(b)
}

func TestAppJSIsPlainModule(t *testing.T) {
	js := appJS(t)
	if got, want := strings.Count(js, "\n")+1, 20; got < want {
		t.Errorf("assets/app.js lines = %d, want at least %d", got, want)
	}
	banned := []struct {
		name string
		re   *regexp.Regexp
	}{
		{"import statement", regexp.MustCompile(`(?m)^\s*import[\s{*'"]`)},
		{"dynamic import", regexp.MustCompile(`\bimport\(`)},
		{"require call", regexp.MustCompile(`\brequire\(`)},
		{"type annotation", regexp.MustCompile(`:\s*string\b`)},
		{"interface declaration", regexp.MustCompile(`\binterface\s`)},
	}
	for _, b := range banned {
		if m := b.re.FindString(js); m != "" {
			t.Errorf("assets/app.js contains %s %q, want a plain dependency-free ES module", b.name, m)
		}
	}
	layout, err := fs.ReadFile(embedded, "templates/layout.html")
	if err != nil {
		t.Fatalf("fs.ReadFile(embedded, %q) error = %v", "templates/layout.html", err)
	}
	script := regexp.MustCompile(`<script\b[^>]*\btype="module"[^>]*\bsrc="/assets/app\.js"|<script\b[^>]*\bsrc="/assets/app\.js"[^>]*\btype="module"`)
	if !script.Match(layout) {
		t.Error(`templates/layout.html does not load /assets/app.js with type="module", want <script type="module" src="/assets/app.js">`)
	}
}

func TestAppJSNoHTMLInjection(t *testing.T) {
	js := appJS(t)
	if !strings.Contains(js, "textContent") {
		t.Error("assets/app.js never uses textContent, want text set with textContent")
	}
	for _, sink := range []string{"innerHTML", "outerHTML", "insertAdjacentHTML", "document.write", "eval("} {
		if strings.Contains(js, sink) {
			t.Errorf("assets/app.js contains %q, want no HTML string sinks", sink)
		}
	}
}

func TestAppJSFetchSendsCSRF(t *testing.T) {
	js := appJS(t)
	fetches := regexp.MustCompile(`\bfetch\(`).FindAllStringIndex(js, -1)
	if len(fetches) == 0 {
		t.Fatal("assets/app.js fetch( calls = 0, want at least one")
	}
	if !regexp.MustCompile(`meta\[name=["']?csrf-token["']?\]`).MatchString(js) {
		t.Error(`assets/app.js does not read meta[name="csrf-token"], want the token taken from the layout meta tag`)
	}
	if got, want := len(fetches), 1; got != want {
		t.Fatalf("assets/app.js raw fetch( calls = %d, want %d inside one CSRF wrapper", got, want)
	}

	defs := regexp.MustCompile(`(?m)^(?:export\s+)?(?:async\s+)?function\s+(\w+)\s*\(|^(?:export\s+)?(?:const|let)\s+(\w+)\s*=\s*(?:async\s*)?(?:function\b|\()`).FindAllStringSubmatchIndex(js, -1)
	start, name := -1, ""
	for _, d := range defs {
		if d[0] > fetches[0][0] {
			break
		}
		start = d[0]
		if d[2] >= 0 {
			name = js[d[2]:d[3]]
		} else {
			name = js[d[4]:d[5]]
		}
	}
	if start < 0 {
		t.Fatal("assets/app.js fetch( is not inside a top-level function, want one CSRF fetch wrapper")
	}
	body := js[start:]
	if end := strings.Index(body, "\n}"); end >= 0 {
		body = body[:end]
	}
	for _, want := range []string{"X-CSRF-Token", "credentials: 'same-origin'"} {
		if !strings.Contains(body, want) {
			t.Errorf("fetch wrapper %s does not set %q, want every request to carry it", name, want)
		}
	}

	const quote = `['"` + "`" + `]`
	call := regexp.MustCompile(`\b` + regexp.QuoteMeta(name) + `\(\s*` + quote + `([^'"` + "`" + `]*)`)
	calls := call.FindAllStringSubmatch(js[start+len(body):], -1)
	var urls []string
	for _, c := range calls {
		urls = append(urls, c[1])
	}
	if len(urls) == 0 {
		t.Fatalf("assets/app.js calls to wrapper %s with a URL literal = 0, want at least one", name)
	}
	for _, u := range urls {
		if !strings.HasPrefix(u, "/") {
			t.Errorf("%s(%q) URL literal does not start with /, want a same-origin path", name, u)
		}
	}
}

func TestAppJSHooksMatchTemplates(t *testing.T) {
	js := appJS(t)
	history, err := fs.ReadFile(embedded, "templates/history.html")
	if err != nil {
		t.Fatalf("fs.ReadFile(embedded, %q) error = %v, want the history page template", "templates/history.html", err)
	}
	var hooks []string
	for _, m := range regexp.MustCompile(`data-js=["']?([A-Za-z0-9_-]+)["']?`).FindAllStringSubmatch(js, -1) {
		if !slices.Contains(hooks, m[1]) {
			hooks = append(hooks, m[1])
		}
	}
	if got, want := len(hooks), 3; got < want {
		t.Errorf("assets/app.js data-js hooks = %v (%d), want at least %d", hooks, got, want)
	}
	for _, want := range []string{"timeline", "versions", "restore"} {
		if !slices.Contains(hooks, want) {
			t.Errorf("assets/app.js data-js hooks = %v, want %q among them", hooks, want)
		}
	}
	for _, h := range hooks {
		if !regexp.MustCompile(`data-js=["']?` + regexp.QuoteMeta(h) + `["'\s>]`).Match(history) {
			t.Errorf("templates/history.html lacks data-js=%q, want every hook app.js queries present", h)
		}
	}
}
