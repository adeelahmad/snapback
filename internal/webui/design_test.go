package webui

import (
	"bytes"
	"io/fs"
	"os"
	"path"
	"regexp"
	"slices"
	"strings"
	"testing"
)

var (
	designSupportsRE  = regexp.MustCompile(`@supports\s+not\s*\(\s*backdrop-filter\s*:\s*blur\(\s*1px\s*\)\s*\)\s*\{`)
	designPrintRE     = regexp.MustCompile(`@media\s+print\s*\{`)
	designGradientRE  = regexp.MustCompile(`(?i)\b((?:repeating-)?(?:linear|radial|conic)-gradient)\(`)
	designHeaderRE    = regexp.MustCompile(`(?s)<header[^>]*>(.*?)</header>`)
	designIconRefRE   = regexp.MustCompile(`icons/([A-Za-z0-9_-]+)\.svg`)
	designStrokeRE    = regexp.MustCompile(`stroke-width\s*=\s*"([^"]*)"`)
	designExternalRE  = regexp.MustCompile(`(?i)(https?:)?//[a-z0-9.-]+\.[a-z]{2,}/`)
	designFocusVarRE  = regexp.MustCompile(`--focus\s*:\s*var\(\s*--blue\s*\)`)
	designLockupSrcRE = regexp.MustCompile(`/assets/lockup-horizontal\.svg`)
	designInverseRE   = regexp.MustCompile(`(?s)prefers-color-scheme:\s*dark[^>]*/assets/lockup-horizontal-inverse\.svg|/assets/lockup-horizontal-inverse\.svg[^>]*prefers-color-scheme:\s*dark`)
)

// forbiddenIconWords are the hyphen-separated parts of an icon name the brand
// book rules out: hats, shields, padlocks, clouds, drives, clocks, refresh
// arrows and shutters.
var forbiddenIconWords = []string{
	"hat", "shield", "lock", "padlock", "cloud", "drive", "hard", "clock", "timer",
	"history", "refresh", "rotate", "aperture", "shutter",
}

// atRuleBody returns the body of the first at-rule opened by re in css, with
// braces matched, and whether it was found.
func atRuleBody(css string, re *regexp.Regexp) (string, bool) {
	loc := re.FindStringIndex(css)
	if loc == nil {
		return "", false
	}
	depth := 1
	for i := loc[1]; i < len(css); i++ {
		switch css[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return css[loc[1]:i], true
			}
		}
	}
	return "", false
}

// cssGradients returns every gradient call in css as name and full argument
// text, with nested parentheses matched.
func cssGradients(css string) [][2]string {
	var out [][2]string
	for _, m := range designGradientRE.FindAllStringSubmatchIndex(css, -1) {
		depth := 1
		for i := m[1]; i < len(css); i++ {
			if css[i] == '(' {
				depth++
			} else if css[i] == ')' {
				depth--
				if depth == 0 {
					out = append(out, [2]string{strings.ToLower(css[m[2]:m[3]]), css[m[1]:i]})
					break
				}
			}
		}
	}
	return out
}

// declValues joins the values of every declaration of prop (or of every
// declaration when prop is empty) in the rules whose selector contains sel.
func declValues(rules []cssRule, sel, prop string) string {
	var b strings.Builder
	for _, r := range rules {
		if !strings.Contains(r.selector, sel) {
			continue
		}
		for _, d := range r.decls {
			if prop == "" || d.prop == prop {
				b.WriteString(d.prop + ": " + d.value + ";\n")
			}
		}
	}
	return b.String()
}

// readTemplate reads templates/<name> from the embedded FS.
func readTemplate(t *testing.T, name string) string {
	t.Helper()
	b, err := fs.ReadFile(embedded, "templates/"+name)
	if err != nil {
		t.Fatalf("fs.ReadFile(embedded, %q) error = %v, want the template", "templates/"+name, err)
	}
	return string(b)
}

// allTemplates returns every embedded template keyed by file name.
func allTemplates(t *testing.T) map[string]string {
	t.Helper()
	names, err := fs.Glob(embedded, "templates/*.html")
	if err != nil || len(names) == 0 {
		t.Fatalf("fs.Glob(embedded, templates/*.html) = %v, %v, want the templates", names, err)
	}
	out := make(map[string]string, len(names))
	for _, n := range names {
		out[path.Base(n)] = readTemplate(t, path.Base(n))
	}
	return out
}

func TestDesignGlassPanel(t *testing.T) {
	css := appCSS(t)
	panel := declValues(parseRules(css), ".panel", "")
	for _, want := range []string{
		"var(--glass-fill)",
		"blur(24px) saturate(1.4)",
		"var(--glass-stroke)",
		"0.5px",
		"var(--glass-edge)",
		"var(--glass-highlight)",
		"45%",
		"var(--shadow-glass)",
		"var(--radius-md)",
	} {
		if !strings.Contains(panel, want) {
			t.Errorf("app.css .panel rules lack %q, want the Liquid Glass panel rule; got:\n%s", want, panel)
		}
	}
	if !strings.Contains(panel, "backdrop-filter:") {
		t.Errorf("app.css .panel rules have no backdrop-filter declaration, want blur(24px) saturate(1.4)")
	}

	supports, ok := atRuleBody(css, designSupportsRE)
	if !ok {
		t.Error("app.css has no @supports not (backdrop-filter: blur(1px)) block, want the flat fallback")
	} else if bg := declValues(parseRules(supports), ".panel", "background"); !strings.Contains(bg, "var(--surface)") {
		t.Errorf("@supports not (backdrop-filter) .panel background = %q, want var(--surface)", bg)
	}

	printBody, ok := atRuleBody(css, designPrintRE)
	if !ok {
		t.Fatal("app.css has no @media print block, want the flat print fallback")
	}
	pr := parseRules(printBody)
	if bg := declValues(pr, ".panel", "background"); !strings.Contains(bg, "var(--surface)") {
		t.Errorf("@media print .panel background = %q, want var(--surface)", bg)
	}
	if border := declValues(pr, ".panel", "border"); !strings.Contains(border, "var(--line)") {
		t.Errorf("@media print .panel border = %q, want a var(--line) border", border)
	}
	if shadow := declValues(pr, ".panel", "box-shadow"); !strings.Contains(shadow, "none") {
		t.Errorf("@media print .panel box-shadow = %q, want none", shadow)
	}
}

func TestDesignRadialTintGround(t *testing.T) {
	css := appCSS(t)
	var ground []string
	for _, r := range parseRules(css) {
		if !strings.Contains(r.selector, "body") && !strings.Contains(r.selector, "html") {
			continue
		}
		for _, g := range cssGradients(declValues([]cssRule{r}, "", "")) {
			if g[0] == "radial-gradient" {
				ground = append(ground, g[1])
			}
		}
	}
	if len(ground) == 0 || len(ground) > 3 {
		t.Fatalf("body/html radial-gradient discs = %d, want 1 to 3 tint discs", len(ground))
	}
	var green, blue bool
	for _, g := range ground {
		green = green || strings.Contains(g, "var(--glass-tint-green)")
		blue = blue || strings.Contains(g, "var(--glass-tint-blue)")
		if !strings.Contains(g, "transparent") {
			t.Errorf("radial-gradient(%s) does not fade to transparent, want a soft edge", g)
		}
	}
	if !green || !blue {
		t.Errorf("ground discs use glass-tint-green = %t, glass-tint-blue = %t, want both", green, blue)
	}
}

func TestDesignFocusRing(t *testing.T) {
	css := appCSS(t)
	tokens := readAsset(t, "tokens.css")
	blue := func(v string) bool {
		if strings.Contains(v, "var(--blue)") {
			return true
		}
		return strings.Contains(v, "var(--focus)") &&
			len(designFocusVarRE.FindAllString(tokens, -1)) >= 2
	}
	found := false
	for _, r := range parseRules(css) {
		if !strings.Contains(r.selector, ":focus-visible") {
			continue
		}
		var outline, offset string
		for _, d := range r.decls {
			switch d.prop {
			case "outline":
				outline = d.value
			case "outline-offset":
				offset = d.value
			}
		}
		if strings.HasPrefix(outline, "2px solid ") && blue(outline) && offset == "2px" {
			found = true
		}
	}
	if !found {
		t.Error("app.css has no :focus-visible { outline: 2px solid var(--blue); outline-offset: 2px }, want the blue focus ring")
	}
}

func TestDesignIconsLocalLucide(t *testing.T) {
	icons, err := fs.Glob(embedded, "assets/icons/*.svg")
	if err != nil {
		t.Fatalf("fs.Glob(embedded, assets/icons/*.svg) error = %v", err)
	}
	if len(icons) == 0 {
		t.Fatal("no embedded assets/icons/*.svg, want local Lucide icon SVGs")
	}
	for _, n := range icons {
		b, err := fs.ReadFile(embedded, n)
		if err != nil {
			t.Fatalf("fs.ReadFile(embedded, %q) error = %v", n, err)
		}
		svg := string(b)
		widths := designStrokeRE.FindAllStringSubmatch(svg, -1)
		if len(widths) == 0 {
			t.Errorf("%s has no stroke-width attribute, want stroke-width=\"1.5\"", n)
		}
		for _, w := range widths {
			if w[1] != "1.5" {
				t.Errorf("%s stroke-width = %q, want \"1.5\"", n, w[1])
			}
		}
		if strings.Contains(svg, "<script") || designExternalRE.MatchString(strings.ReplaceAll(svg, "http://www.w3.org/2000/svg", "")) {
			t.Errorf("%s has a script or external reference, want a self-contained icon", n)
		}
	}
	for name, src := range allTemplates(t) {
		if m := designExternalRE.FindAllString(src, -1); len(m) > 0 {
			t.Errorf("template %s references external URLs %v, want only local assets", name, m)
		}
		for _, ref := range designIconRefRE.FindAllStringSubmatch(src, -1) {
			if !slices.Contains(icons, "assets/icons/"+ref[1]+".svg") {
				t.Errorf("template %s references icons/%s.svg, which is not embedded", name, ref[1])
			}
		}
	}
}

func TestDesignNoForbiddenIcons(t *testing.T) {
	icons, err := fs.Glob(embedded, "assets/icons/*.svg")
	if err != nil {
		t.Fatalf("fs.Glob(embedded, assets/icons/*.svg) error = %v", err)
	}
	names := make(map[string]string)
	for _, n := range icons {
		names[strings.TrimSuffix(path.Base(n), ".svg")] = n
	}
	for tmpl, src := range allTemplates(t) {
		for _, ref := range designIconRefRE.FindAllStringSubmatch(src, -1) {
			names[ref[1]] = "templates/" + tmpl
		}
	}
	for name, where := range names {
		for _, part := range strings.Split(strings.ToLower(name), "-") {
			if slices.Contains(forbiddenIconWords, part) {
				t.Errorf("icon %q (in %s) is a forbidden motif %q, want none of %v", name, where, part, forbiddenIconWords)
			}
		}
	}
}

func TestDesignHeaderLockup(t *testing.T) {
	m := designHeaderRE.FindStringSubmatch(readTemplate(t, "layout.html"))
	if m == nil {
		t.Fatal("layout.html has no <header>, want the site header")
	}
	header := m[1]
	if !designLockupSrcRE.MatchString(header) {
		t.Errorf("layout.html header does not reference /assets/lockup-horizontal.svg, want the brand lockup; got:\n%s", header)
	}
	if !designInverseRE.MatchString(header) {
		t.Errorf("layout.html header has no dark-scheme /assets/lockup-horizontal-inverse.svg source, want the inverse on dark")
	}
	for _, name := range []string{"lockup-horizontal.svg", "lockup-horizontal-inverse.svg"} {
		got, err := fs.ReadFile(embedded, "assets/"+name)
		if err != nil {
			t.Errorf("fs.ReadFile(embedded, %q) error = %v, want the brand lockup", "assets/"+name, err)
			continue
		}
		want, err := os.ReadFile("../../docs/brand/" + name)
		if err != nil {
			t.Fatalf("os.ReadFile(docs/brand/%s) error = %v", name, err)
		}
		if !bytes.Equal(got, want) {
			t.Errorf("assets/%s differs from docs/brand/%s, want the unmodified brand file", name, name)
		}
	}
}

func TestDesignGradientsOnlyGlassAndTint(t *testing.T) {
	for _, g := range cssGradients(appCSS(t)) {
		switch g[0] {
		case "linear-gradient":
			if !strings.Contains(g[1], "var(--glass-highlight)") {
				t.Errorf("app.css linear-gradient(%s) is not the glass highlight, want only var(--glass-highlight)", g[1])
			}
		case "radial-gradient":
			if !strings.Contains(g[1], "var(--glass-tint-green)") && !strings.Contains(g[1], "var(--glass-tint-blue)") {
				t.Errorf("app.css radial-gradient(%s) is not a tint disc, want var(--glass-tint-*)", g[1])
			}
		default:
			t.Errorf("app.css uses %s(%s), want only the glass highlight and tint discs", g[0], g[1])
		}
	}
}

func TestDesignNoRawColours(t *testing.T) {
	css := appCSS(t)
	stripped := cssVarCallRE.ReplaceAllString(css, "")
	for _, re := range literalColourPatterns {
		if m := re.FindAllString(stripped, -1); len(m) > 0 {
			t.Errorf("app.css contains raw colour %v outside var(--...), want tokens only", m)
		}
	}
	for _, r := range parseRules(css) {
		for _, d := range r.decls {
			for _, w := range cssWordRE.FindAllString(cssVarCallRE.ReplaceAllString(d.value, ""), -1) {
				if slices.Contains(namedColours, strings.ToLower(w)) {
					t.Errorf("app.css %s { %s: %s } uses named colour %q, want var(--...)", r.selector, d.prop, d.value, w)
				}
			}
		}
	}
}
