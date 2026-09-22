package site_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"unicode/utf8"
)

const (
	indexHTML    = "web/index.html"
	siteOrigin   = "https://snapback.sh/"
	ogImageURL   = "https://snapback.sh/og-card-light.png"
	minDescLen   = 50
	maxDescLen   = 160
	lightSchemeQ = "(prefers-color-scheme: light)"
	darkSchemeQ  = "(prefers-color-scheme: dark)"
)

var (
	headTagRE = regexp.MustCompile(`(?is)<(meta|link|html)\b[^>]*>`)
	attrRE    = regexp.MustCompile(`(?s)([a-zA-Z_:-]+)\s*=\s*"([^"]*)"`)
	titleRE   = regexp.MustCompile(`(?is)<title>(.*?)</title>`)
	urlAttrRE = regexp.MustCompile(`(?i)\b(src|href|content)\s*=\s*"([^"]*)"`)
)

// headTag is one <meta>, <link> or <html> tag with its attributes.
type headTag struct {
	name  string
	attrs map[string]string
}

// headTags returns every <meta>, <link> and <html> tag in html in order.
func headTags(html string) []headTag {
	var tags []headTag
	for _, m := range headTagRE.FindAllStringSubmatch(html, -1) {
		attrs := map[string]string{}
		for _, a := range attrRE.FindAllStringSubmatch(m[0], -1) {
			attrs[strings.ToLower(a[1])] = a[2]
		}
		tags = append(tags, headTag{name: strings.ToLower(m[1]), attrs: attrs})
	}
	return tags
}

// findTag returns the first tag named name whose attributes include every
// key/value in match.
func findTag(tags []headTag, name string, match map[string]string) (headTag, bool) {
	for _, tg := range tags {
		if tg.name != name {
			continue
		}
		ok := true
		for k, v := range match {
			if tg.attrs[k] != v {
				ok = false
				break
			}
		}
		if ok {
			return tg, true
		}
	}
	return headTag{}, false
}

// metaContent returns the content of the meta whose name or property is key.
func metaContent(tags []headTag, key string) string {
	for _, tg := range tags {
		if tg.name != "meta" {
			continue
		}
		if tg.attrs["name"] == key || tg.attrs["property"] == key {
			return tg.attrs["content"]
		}
	}
	return ""
}

func pageTitle(html string) string {
	m := titleRE.FindStringSubmatch(html)
	if m == nil {
		return ""
	}
	return strings.TrimSpace(m[1])
}

func TestHeadMeta(t *testing.T) {
	html := readRepoFile(t, indexHTML)
	tags := headTags(html)

	if got := pageTitle(html); !strings.HasPrefix(got, "snapback") {
		t.Errorf("title = %q, want prefix %q", got, "snapback")
	}

	tests := []struct {
		desc  string
		tag   string
		match map[string]string
		attr  string
		want  string
	}{
		{"html[lang]", "html", nil, "lang", "en"},
		{"link[rel=canonical]", "link", map[string]string{"rel": "canonical"}, "href", siteOrigin},
		{"meta[name=color-scheme]", "meta", map[string]string{"name": "color-scheme"}, "content", "light dark"},
		{"og:url", "meta", map[string]string{"property": "og:url"}, "content", siteOrigin},
		{"og:type", "meta", map[string]string{"property": "og:type"}, "content", "website"},
		{"og:image", "meta", map[string]string{"property": "og:image"}, "content", ogImageURL},
		{"twitter:card", "meta", map[string]string{"name": "twitter:card"}, "content", "summary_large_image"},
		{"link[rel=icon][type=image/svg+xml]", "link", map[string]string{"rel": "icon", "type": "image/svg+xml"}, "href", "/favicon.svg"},
		{"link[rel=apple-touch-icon]", "link", map[string]string{"rel": "apple-touch-icon"}, "href", "/app-icon-180.png"},
	}
	for _, tc := range tests {
		tg, ok := findTag(tags, tc.tag, tc.match)
		if !ok {
			t.Errorf("%s: tag missing from %s", tc.desc, indexHTML)
			continue
		}
		if got := tg.attrs[tc.attr]; got != tc.want {
			t.Errorf("%s %s = %q, want %q", tc.desc, tc.attr, got, tc.want)
		}
	}
}

// canvasValues returns the light and dark values of the canvas colour token.
// A token value is either one string used by every theme or an object keyed
// by theme id, so values are decoded only for the canvas token.
func canvasValues(t *testing.T) (light, dark string) {
	t.Helper()
	var tokens struct {
		Color struct {
			Tokens []struct {
				Name  string          `json:"name"`
				Value json.RawMessage `json:"value"`
			} `json:"tokens"`
		} `json:"color"`
	}
	loadJSON(t, "web/tokens.json", &tokens)
	for _, tok := range tokens.Color.Tokens {
		if tok.Name != "canvas" {
			continue
		}
		var single string
		if err := json.Unmarshal(tok.Value, &single); err == nil {
			return single, single
		}
		var perTheme map[string]string
		if err := json.Unmarshal(tok.Value, &perTheme); err != nil {
			t.Fatalf("web/tokens.json canvas value %s: want string or {light,dark} object: %v", tok.Value, err)
		}
		return perTheme["light"], perTheme["dark"]
	}
	t.Fatalf("web/tokens.json has no canvas colour token")
	return "", ""
}

func TestThemeColorPerScheme(t *testing.T) {
	light, dark := canvasValues(t)
	if light == "" || dark == "" {
		t.Fatalf("canvas token values = (%q, %q), want both non-empty", light, dark)
	}
	tags := headTags(readRepoFile(t, indexHTML))

	var themeColors []headTag
	for _, tg := range tags {
		if tg.name == "meta" && tg.attrs["name"] == "theme-color" {
			themeColors = append(themeColors, tg)
		}
	}
	if got, want := len(themeColors), 2; got != want {
		t.Errorf("theme-color meta count = %d, want %d", got, want)
	}

	tests := []struct {
		media string
		want  string
	}{
		{lightSchemeQ, light},
		{darkSchemeQ, dark},
	}
	for _, tc := range tests {
		tg, ok := findTag(tags, "meta", map[string]string{"name": "theme-color", "media": tc.media})
		if !ok {
			t.Errorf("theme-color with media %q missing", tc.media)
			continue
		}
		if got := tg.attrs["content"]; !strings.EqualFold(got, tc.want) {
			t.Errorf("theme-color %q content = %q, want %q", tc.media, got, tc.want)
		}
	}
}

func isEmoji(r rune) bool {
	return (r >= 0x1F000 && r <= 0x1FAFF) || (r >= 0x2600 && r <= 0x27BF)
}

func TestMetaVoice(t *testing.T) {
	html := readRepoFile(t, indexHTML)
	tags := headTags(html)

	fields := []struct {
		name  string
		value string
	}{
		{"title", pageTitle(html)},
		{"description", metaContent(tags, "description")},
		{"og:title", metaContent(tags, "og:title")},
		{"og:description", metaContent(tags, "og:description")},
	}
	for _, f := range fields {
		if f.value == "" {
			t.Errorf("%s is empty or missing", f.name)
			continue
		}
		if strings.Contains(f.value, "!") {
			t.Errorf("%s = %q, want no %q", f.name, f.value, "!")
		}
		if strings.Contains(f.value, "Snapback") {
			t.Errorf("%s = %q, want lowercase name, not %q", f.name, f.value, "Snapback")
		}
		for _, r := range f.value {
			if isEmoji(r) {
				t.Errorf("%s = %q, contains emoji %U", f.name, f.value, r)
				break
			}
		}
	}

	desc := metaContent(tags, "description")
	if n := utf8.RuneCountInString(desc); n < minDescLen || n > maxDescLen {
		t.Errorf("description length = %d, want %d-%d", n, minDescLen, maxDescLen)
	}
}

// headSection returns the text between <head> and </head>, or "" if absent.
func headSection(html string) string {
	lower := strings.ToLower(html)
	start := strings.Index(lower, "<head>")
	end := strings.Index(lower, "</head>")
	if start < 0 || end < start {
		return ""
	}
	return html[start:end]
}

func TestMetaIconsExist(t *testing.T) {
	head := headSection(readRepoFile(t, indexHTML))
	if head == "" {
		t.Fatalf("%s has no <head> section", indexHTML)
	}
	var paths []string
	for _, m := range urlAttrRE.FindAllStringSubmatch(head, -1) {
		if strings.HasPrefix(m[2], "/") && !strings.HasPrefix(m[2], "//") {
			paths = append(paths, m[2])
		}
	}
	if len(paths) == 0 {
		t.Fatalf("%s head has no local asset paths, want favicon and icon links", indexHTML)
	}
	public := filepath.Join(repoRoot(t), "web", "public")
	for _, p := range paths {
		if _, err := os.Stat(filepath.Join(public, filepath.FromSlash(p))); err != nil {
			t.Errorf("head path %q: web/public%s missing: %v", p, p, err)
		}
	}
}

func TestNoExternalOrigins(t *testing.T) {
	html := readRepoFile(t, indexHTML)
	tags := headTags(html)
	allowed := map[string]bool{}
	for _, key := range []string{"og:url", "og:image", "twitter:image"} {
		if v := metaContent(tags, key); v != "" {
			allowed[v] = true
		}
	}
	if tg, ok := findTag(tags, "link", map[string]string{"rel": "canonical"}); ok {
		allowed[tg.attrs["href"]] = true
	}
	if !allowed[siteOrigin] {
		t.Fatalf("canonical/og:url %q not found in %s, want it present before checking origins", siteOrigin, indexHTML)
	}
	for v := range allowed {
		if !strings.HasPrefix(v, siteOrigin) {
			t.Errorf("allowed URL %q is not on %s", v, siteOrigin)
		}
	}

	for _, m := range urlAttrRE.FindAllStringSubmatch(html, -1) {
		attr, v := strings.ToLower(m[1]), m[2]
		if attr == "content" {
			continue
		}
		if attr == "src" && v == gtagLoaderURL {
			continue
		}
		if strings.HasPrefix(v, "//") || (strings.HasPrefix(strings.ToLower(v), "http") && !allowed[v]) {
			t.Errorf("%s=%q is an external origin, want only %s (as the gtag loader)", attr, v, gtagOrigin)
		}
	}

	// The one allowed external origin must be used only as the gtag loader.
	var gtagSrcs int
	for _, tg := range scriptTags(html) {
		if tg.attrs["src"] == gtagLoaderURL {
			gtagSrcs++
		}
	}
	if n := strings.Count(html, gtagOrigin); n != 1 || gtagSrcs != 1 {
		t.Errorf("%s: %s occurrences = %d, gtag loader scripts = %d, want 1 and 1", indexHTML, gtagOrigin, n, gtagSrcs)
	}
}
