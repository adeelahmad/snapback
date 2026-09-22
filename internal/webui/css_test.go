package webui

import (
	"fmt"
	"io/fs"
	"math"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// cssRule is one innermost ruleset: its selector and its declarations in
// source order.
type cssRule struct {
	selector string
	decls    []cssDecl
}

// cssDecl is one property: value declaration with both sides trimmed and the
// property lower-cased.
type cssDecl struct {
	prop, value string
}

var (
	cssCommentRE   = regexp.MustCompile(`(?s)/\*.*?\*/`)
	cssRuleRE      = regexp.MustCompile(`([^{}]*)\{([^{}]*)\}`)
	cssVarRE       = regexp.MustCompile(`var\(\s*(--[A-Za-z0-9_-]+)`)
	cssVarCallRE   = regexp.MustCompile(`var\([^()]*\)`)
	cssTokenDefRE  = regexp.MustCompile(`(--[A-Za-z0-9_-]+)\s*:`)
	cssHexTokenRE  = regexp.MustCompile(`(--[A-Za-z0-9_-]+)\s*:\s*#([0-9a-fA-F]{6}|[0-9a-fA-F]{3})\s*;`)
	cssWordRE      = regexp.MustCompile(`[A-Za-z]+`)
	cssLightRootRE = regexp.MustCompile(`(?ms)^:root\s*\{(.*?)\}`)
	cssDarkRootRE  = regexp.MustCompile(`(?s)@media\s*\(\s*prefers-color-scheme\s*:\s*dark\s*\)\s*\{\s*:root\s*\{(.*?)\}`)
	cssReducedRE   = regexp.MustCompile(`@media\s*\(\s*prefers-reduced-motion\s*:\s*reduce\s*\)\s*\{`)
)

// literalColourPatterns are the colour literals app.css must never contain.
var literalColourPatterns = []*regexp.Regexp{
	regexp.MustCompile(`#[0-9a-fA-F]{3,8}\b`),
	regexp.MustCompile(`(?i)rgb\(`),
	regexp.MustCompile(`(?i)rgba\(`),
	regexp.MustCompile(`(?i)hsl\(`),
}

// namedColours is the CSS named-colour keyword list, minus the allowed
// keywords transparent and currentColor.
var namedColours = []string{
	"aliceblue", "antiquewhite", "aqua", "aquamarine", "azure", "beige", "bisque", "black",
	"blanchedalmond", "blue", "blueviolet", "brown", "burlywood", "cadetblue", "chartreuse",
	"chocolate", "coral", "cornflowerblue", "cornsilk", "crimson", "cyan", "darkblue",
	"darkcyan", "darkgoldenrod", "darkgray", "darkgreen", "darkgrey", "darkkhaki",
	"darkmagenta", "darkolivegreen", "darkorange", "darkorchid", "darkred", "darksalmon",
	"darkseagreen", "darkslateblue", "darkslategray", "darkslategrey", "darkturquoise",
	"darkviolet", "deeppink", "deepskyblue", "dimgray", "dimgrey", "dodgerblue", "firebrick",
	"floralwhite", "forestgreen", "fuchsia", "gainsboro", "ghostwhite", "gold", "goldenrod",
	"gray", "green", "greenyellow", "grey", "honeydew", "hotpink", "indianred", "indigo",
	"ivory", "khaki", "lavender", "lavenderblush", "lawngreen", "lemonchiffon", "lightblue",
	"lightcoral", "lightcyan", "lightgoldenrodyellow", "lightgray", "lightgreen", "lightgrey",
	"lightpink", "lightsalmon", "lightseagreen", "lightskyblue", "lightslategray",
	"lightslategrey", "lightsteelblue", "lightyellow", "lime", "limegreen", "linen", "magenta",
	"maroon", "mediumaquamarine", "mediumblue", "mediumorchid", "mediumpurple",
	"mediumseagreen", "mediumslateblue", "mediumspringgreen", "mediumturquoise",
	"mediumvioletred", "midnightblue", "mintcream", "mistyrose", "moccasin", "navajowhite",
	"navy", "oldlace", "olive", "olivedrab", "orange", "orangered", "orchid", "palegoldenrod",
	"palegreen", "paleturquoise", "palevioletred", "papayawhip", "peachpuff", "peru", "pink",
	"plum", "powderblue", "purple", "rebeccapurple", "red", "rosybrown", "royalblue",
	"saddlebrown", "salmon", "sandybrown", "seagreen", "seashell", "sienna", "silver",
	"skyblue", "slateblue", "slategray", "slategrey", "snow", "springgreen", "steelblue",
	"tan", "teal", "thistle", "tomato", "turquoise", "violet", "wheat", "white", "whitesmoke",
	"yellow", "yellowgreen",
}

// allowedTextColours are the only values a color: declaration may take.
var allowedTextColours = []string{
	"var(--ink)", "var(--ink-body)", "var(--ink-muted)", "var(--accent-text)",
	"var(--yellow-text)", "var(--on-accent)", "inherit", "currentColor",
}

// allowedFontFamilies are the only values a font-family: declaration may take.
var allowedFontFamilies = []string{
	"var(--font-sans)", "var(--font-mono)", "var(--font-mono-display)", "inherit",
}

// surfaceOnlyTextTokens are text tokens only ever placed on --surface, so
// their contrast is checked against --surface alone.
var surfaceOnlyTextTokens = []string{"--yellow-text"}

// minTextContrast is the WCAG AA contrast ratio for normal text.
const minTextContrast = 4.5

// readAsset reads assets/<name> from the embedded FS and fails the test if it
// is missing or empty.
func readAsset(t *testing.T, name string) string {
	t.Helper()
	b, err := fs.ReadFile(embedded, "assets/"+name)
	if err != nil {
		t.Fatalf("fs.ReadFile(embedded, %q) error = %v, want the embedded stylesheet", "assets/"+name, err)
	}
	if len(b) == 0 {
		t.Fatalf("embedded assets/%s is empty, want a stylesheet", name)
	}
	return string(b)
}

// appCSS returns the embedded app.css with comments stripped.
func appCSS(t *testing.T) string {
	t.Helper()
	return cssCommentRE.ReplaceAllString(readAsset(t, "app.css"), "")
}

// parseRules returns every innermost ruleset of css; at-rule wrappers such as
// @media contribute only their inner rules.
func parseRules(css string) []cssRule {
	var rules []cssRule
	for _, m := range cssRuleRE.FindAllStringSubmatch(css, -1) {
		r := cssRule{selector: strings.TrimSpace(m[1])}
		for _, d := range strings.Split(m[2], ";") {
			prop, value, ok := strings.Cut(d, ":")
			if !ok {
				continue
			}
			r.decls = append(r.decls, cssDecl{
				prop:  strings.ToLower(strings.TrimSpace(prop)),
				value: strings.TrimSpace(value),
			})
		}
		rules = append(rules, r)
	}
	return rules
}

// hexTokens maps each --name: #hex declaration in block to its hex digits.
func hexTokens(block string) map[string]string {
	m := make(map[string]string)
	for _, g := range cssHexTokenRE.FindAllStringSubmatch(block, -1) {
		m[g[1]] = g[2]
	}
	return m
}

// relativeLuminance returns the WCAG 2 relative luminance of a #rgb or
// #rrggbb colour given without the leading #.
func relativeLuminance(hex string) (float64, error) {
	if len(hex) == 3 {
		hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
	}
	if len(hex) != 6 {
		return 0, fmt.Errorf("bad hex colour %q", hex)
	}
	var lin [3]float64
	for i := range 3 {
		v, err := strconv.ParseUint(hex[2*i:2*i+2], 16, 8)
		if err != nil {
			return 0, fmt.Errorf("bad hex colour %q: %w", hex, err)
		}
		c := float64(v) / 255
		if c <= 0.03928 {
			lin[i] = c / 12.92
		} else {
			lin[i] = math.Pow((c+0.055)/1.055, 2.4)
		}
	}
	return 0.2126*lin[0] + 0.7152*lin[1] + 0.0722*lin[2], nil
}

// contrastRatio returns the WCAG contrast ratio between two hex colours.
func contrastRatio(a, b string) (float64, error) {
	la, err := relativeLuminance(a)
	if err != nil {
		return 0, err
	}
	lb, err := relativeLuminance(b)
	if err != nil {
		return 0, err
	}
	hi, lo := math.Max(la, lb), math.Min(la, lb)
	return (hi + 0.05) / (lo + 0.05), nil
}

func TestAppCSSUsesOnlyTokenColours(t *testing.T) {
	css := appCSS(t)
	if n := strings.Count(css, "var(--"); n < 10 {
		t.Fatalf("app.css has %d var(-- references, want at least 10", n)
	}
	for _, re := range literalColourPatterns {
		if m := re.FindAllString(css, -1); len(m) > 0 {
			t.Errorf("app.css contains literal colour %v (pattern %s), want only var(--...)", m, re)
		}
	}
	for _, r := range parseRules(css) {
		for _, d := range r.decls {
			value := cssVarCallRE.ReplaceAllString(d.value, "")
			for _, w := range cssWordRE.FindAllString(value, -1) {
				if slices.Contains(namedColours, strings.ToLower(w)) {
					t.Errorf("app.css %s { %s: %s } uses named colour %q, want var(--...)", r.selector, d.prop, d.value, w)
				}
			}
		}
	}
}

func TestAppCSSReferencesDefinedTokens(t *testing.T) {
	css := appCSS(t)
	tokens := readAsset(t, "tokens.css")
	defined := make(map[string]bool)
	for _, m := range cssTokenDefRE.FindAllStringSubmatch(tokens, -1) {
		defined[m[1]] = true
	}
	if len(defined) == 0 {
		t.Fatal("tokens.css defines no --name: tokens, want the design-system tokens")
	}
	used := cssVarRE.FindAllStringSubmatch(css, -1)
	if len(used) == 0 {
		t.Fatal("app.css references no var(--name), want token references")
	}
	for _, m := range used {
		if !defined[m[1]] {
			t.Errorf("app.css references var(%s), which tokens.css does not define", m[1])
		}
	}
}

func TestAppCSSTextColoursAllowed(t *testing.T) {
	rules := parseRules(appCSS(t))
	colours := 0
	for _, r := range rules {
		usesOnAccent := false
		usesYellowText := false
		background := ""
		for _, d := range r.decls {
			if strings.Contains(d.value, "var(--on-accent)") {
				usesOnAccent = true
			}
			if strings.Contains(d.value, "var(--yellow-text)") {
				usesYellowText = true
			}
			if d.prop == "background" || d.prop == "background-color" {
				background = d.value
			}
			if d.prop != "color" {
				continue
			}
			colours++
			if !slices.Contains(allowedTextColours, d.value) {
				t.Errorf("app.css %s { color: %s }, want one of %v", r.selector, d.value, allowedTextColours)
			}
		}
		if usesOnAccent && background != "var(--accent)" {
			t.Errorf("app.css %s uses var(--on-accent) with background %q, want background var(--accent)", r.selector, background)
		}
		if usesYellowText && background != "var(--surface)" {
			t.Errorf("app.css %s uses var(--yellow-text) with background %q, want background var(--surface)", r.selector, background)
		}
	}
	if colours == 0 {
		t.Fatal("app.css has no color: declarations, want text colours on the tokens")
	}
}

func TestTextTokenContrast(t *testing.T) {
	tokens := cssCommentRE.ReplaceAllString(readAsset(t, "tokens.css"), "")
	lightBlock := cssLightRootRE.FindStringSubmatch(tokens)
	darkBlock := cssDarkRootRE.FindStringSubmatch(tokens)
	if lightBlock == nil || darkBlock == nil {
		t.Fatalf("tokens.css light :root found = %v, dark :root found = %v, want both", lightBlock != nil, darkBlock != nil)
	}
	schemes := []struct {
		name  string
		block string
	}{
		{"light", lightBlock[1]},
		{"dark", darkBlock[1]},
	}
	textTokens := []string{"--ink", "--ink-body", "--ink-muted", "--accent-text", "--yellow-text"}
	for _, s := range schemes {
		hex := hexTokens(s.block)
		if len(hex) == 0 {
			t.Fatalf("tokens.css %s block has no hex tokens, want a colour map", s.name)
		}
		type pair struct{ fg, bg string }
		var pairs []pair
		for _, fg := range textTokens {
			pairs = append(pairs, pair{fg, "--surface"})
			if !slices.Contains(surfaceOnlyTextTokens, fg) {
				pairs = append(pairs, pair{fg, "--canvas"})
			}
		}
		pairs = append(pairs, pair{"--on-accent", "--accent"})
		for _, p := range pairs {
			fg, okFG := hex[p.fg]
			bg, okBG := hex[p.bg]
			if !okFG || !okBG {
				t.Errorf("%s: %s or %s has no hex value in tokens.css", s.name, p.fg, p.bg)
				continue
			}
			got, err := contrastRatio(fg, bg)
			if err != nil {
				t.Errorf("%s: contrastRatio(%s, %s) error = %v", s.name, p.fg, p.bg, err)
				continue
			}
			if got < minTextContrast {
				t.Errorf("%s: contrast(%s #%s on %s #%s) = %.2f, want >= %.1f", s.name, p.fg, fg, p.bg, bg, got, minTextContrast)
			}
		}
	}
}

func TestAppCSSFocusAndReducedMotion(t *testing.T) {
	css := appCSS(t)
	focus := false
	for _, r := range parseRules(css) {
		if !strings.Contains(r.selector, ":focus-visible") {
			continue
		}
		for _, d := range r.decls {
			if strings.HasPrefix(d.prop, "outline") && strings.Contains(d.value, "var(--focus)") {
				focus = true
			}
		}
	}
	if !focus {
		t.Error("app.css has no :focus-visible rule with an outline using var(--focus), want one")
	}
	if !cssReducedRE.MatchString(css) {
		t.Error("app.css has no @media (prefers-reduced-motion: reduce) block, want one")
	}
}

func TestAppCSSFontsFromTokens(t *testing.T) {
	css := appCSS(t)
	if strings.Contains(strings.ToLower(css), "@font-face") {
		t.Error("app.css declares @font-face, want fonts only from tokens.css")
	}
	families := 0
	for _, r := range parseRules(css) {
		for _, d := range r.decls {
			if d.prop != "font-family" {
				continue
			}
			families++
			if !slices.Contains(allowedFontFamilies, d.value) {
				t.Errorf("app.css %s { font-family: %s }, want one of %v", r.selector, d.value, allowedFontFamilies)
			}
		}
	}
	if families == 0 {
		t.Fatal("app.css has no font-family: declarations, want the token font stacks")
	}
}
