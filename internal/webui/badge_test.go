package webui

import (
	"regexp"
	"strings"
	"testing"
)

var (
	badgeSelectorRE = regexp.MustCompile(`^\.badge(-[a-z]+)?$`)
	tokenValueRE    = regexp.MustCompile(`^var\(\s*(--[A-Za-z0-9_-]+)\s*\)$`)
	buttonRE        = regexp.MustCompile(`(?s)<button[^>]*>.*?</button>`)
)

// TestAppCSSNoDashedBorders wants no rule but a focus rule to draw a dashed
// border or outline, so nothing else looks like the focus ring.
func TestAppCSSNoDashedBorders(t *testing.T) {
	for _, r := range parseRules(appCSS(t)) {
		if strings.Contains(r.selector, ":focus") {
			continue
		}
		for _, d := range r.decls {
			if strings.Contains(d.value, "dashed") {
				t.Errorf("rule %q %s: %s, want no dashed treatment outside the focus rule", r.selector, d.prop, d.value)
			}
		}
	}
}

// TestBadgeContrast wants every badge class to set its own colour and
// background from tokens, with at least 4.5:1 contrast in both themes.
func TestBadgeContrast(t *testing.T) {
	tokens := cssCommentRE.ReplaceAllString(readAsset(t, "tokens.css"), "")
	light := cssLightRootRE.FindStringSubmatch(tokens)
	dark := cssDarkRootRE.FindStringSubmatch(tokens)
	if light == nil || dark == nil {
		t.Fatalf("tokens.css light :root found = %v, dark :root found = %v, want both", light != nil, dark != nil)
	}
	themes := map[string]map[string]string{"light": hexTokens(light[1]), "dark": hexTokens(dark[1])}

	found := 0
	for _, r := range parseRules(appCSS(t)) {
		if !badgeSelectorRE.MatchString(r.selector) {
			continue
		}
		found++
		var fg, bg string
		for _, d := range r.decls {
			m := tokenValueRE.FindStringSubmatch(d.value)
			switch {
			case d.prop == "color" && m != nil:
				fg = m[1]
			case d.prop == "background" && m != nil:
				bg = m[1]
			}
		}
		if fg == "" || bg == "" {
			t.Errorf("rule %q color = %q, background = %q, want both set from tokens", r.selector, fg, bg)
			continue
		}
		for name, hex := range themes {
			got, err := contrastRatio(hex[fg], hex[bg])
			if err != nil {
				t.Errorf("%s: contrastRatio(%s, %s) error = %v", name, fg, bg, err)
				continue
			}
			if got < minTextContrast {
				t.Errorf("%s: %q contrast(%s on %s) = %.2f, want >= %.1f", name, r.selector, fg, bg, got, minTextContrast)
			}
		}
	}
	if found == 0 {
		t.Fatal("app.css has no .badge rules, want badge styles")
	}
}

// TestHistoryBadgesNotInsideButtons wants no badge inside a button, since
// buttons are filled with the accent colour.
func TestHistoryBadgesNotInsideButtons(t *testing.T) {
	out := render(t, mustLoad(t, ""), "history", historyView())

	for _, b := range buttonRE.FindAllString(out, -1) {
		if strings.Contains(b, `class="badge`) {
			t.Errorf("render(history) button = %q, want no badge inside an accent-filled button", b)
		}
	}
}
