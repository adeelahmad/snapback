package webui

import (
	"regexp"
	"slices"
	"strings"
	"testing"
)

var (
	// controlClassAttrRE matches a literal class attribute in a template; the
	// {} exclusion skips any attribute built from a template action.
	controlClassAttrRE = regexp.MustCompile(`class="([^"{}]*)"`)

	// controlStatePatterns are the interaction states every form control must
	// style, keyed by the selector shape app.css has to carry.
	controlStatePatterns = []struct {
		name string
		re   *regexp.Regexp
		want string
	}{
		{
			name: "focus",
			re:   regexp.MustCompile(`(\.field__control|\.field__switch|\.chip__remove)[^,]*:focus-visible`),
			want: `a rule such as .field__control:focus-visible`,
		},
		{
			name: "error",
			re:   regexp.MustCompile(`\[aria-invalid="true"\]|\.is-invalid`),
			want: `a rule such as .field__control[aria-invalid="true"]`,
		},
		{
			name: "disabled",
			re:   regexp.MustCompile(`:disabled`),
			want: `a rule such as .field__control:disabled`,
		},
	}
)

// coreControlClasses are the class names the control partials are built from;
// they guard controlClasses against a template that stops emitting them.
var coreControlClasses = []string{
	"chip", "chip__remove", "chip__text", "field", "field--chips", "field--rows",
	"field--toggle", "field__chips", "field__control", "field__error", "field__help",
	"field__label", "field__row", "field__switch", "field__switch-track",
}

// controlClasses returns every class name the control partials emit, sorted
// and deduplicated.
func controlClasses(t *testing.T) []string {
	t.Helper()
	src := readTemplate(t, "controls.html")
	seen := make(map[string]bool)
	var out []string
	for _, m := range controlClassAttrRE.FindAllStringSubmatch(src, -1) {
		for _, c := range strings.Fields(m[1]) {
			if seen[c] {
				continue
			}
			seen[c] = true
			out = append(out, c)
		}
	}
	slices.Sort(out)
	for _, want := range coreControlClasses {
		if !slices.Contains(out, want) {
			t.Fatalf("controls.html emits %v, want it to include %q", out, want)
		}
	}
	return out
}

// classSelectorRE matches a selector that targets the whole class name, so
// .field does not match .field__label or .field--rows.
func classSelectorRE(class string) *regexp.Regexp {
	return regexp.MustCompile(`\.` + regexp.QuoteMeta(class) + `($|[^A-Za-z0-9_-])`)
}

// classDecls returns every declaration of every rule targeting class.
func classDecls(rules []cssRule, class string) []cssDecl {
	re := classSelectorRE(class)
	var out []cssDecl
	for _, r := range rules {
		for sel := range strings.SplitSeq(r.selector, ",") {
			if re.MatchString(strings.TrimSpace(sel)) {
				out = append(out, r.decls...)
				break
			}
		}
	}
	return out
}

// TestControlClassesStyledFromTokens wants every class the control partials
// emit to have a rule in app.css whose values name colours only through
// var(--...) tokens.
func TestControlClassesStyledFromTokens(t *testing.T) {
	rules := parseRules(appCSS(t))
	for _, class := range controlClasses(t) {
		decls := classDecls(rules, class)
		if len(decls) == 0 {
			t.Errorf("app.css has no rule for .%s, want the control partial styled", class)
			continue
		}
		for _, d := range decls {
			value := cssVarCallRE.ReplaceAllString(d.value, "")
			for _, re := range literalColourPatterns {
				if m := re.FindAllString(value, -1); len(m) > 0 {
					t.Errorf("app.css .%s { %s: %s } uses literal colour %v, want var(--...) from tokens.css", class, d.prop, d.value, m)
				}
			}
			for _, w := range cssWordRE.FindAllString(value, -1) {
				if slices.Contains(namedColours, strings.ToLower(w)) {
					t.Errorf("app.css .%s { %s: %s } uses named colour %q, want var(--...) from tokens.css", class, d.prop, d.value, w)
				}
			}
		}
	}
}

// TestControlStateRules wants the focus, error and disabled states of the
// form controls each to carry their own rule.
func TestControlStateRules(t *testing.T) {
	rules := parseRules(appCSS(t))
	for _, state := range controlStatePatterns {
		styled := false
		for _, r := range rules {
			if state.re.MatchString(r.selector) && len(r.decls) > 0 {
				styled = true
				break
			}
		}
		if !styled {
			t.Errorf("app.css has no %s-state rule for the form controls, want %s", state.name, state.want)
		}
	}
}

// TestControlErrorContrast wants the field error message to set its own
// colour and background from tokens, with at least 4.5:1 contrast in both
// themes, exactly as the badges do.
func TestControlErrorContrast(t *testing.T) {
	tokens := cssCommentRE.ReplaceAllString(readAsset(t, "tokens.css"), "")
	light := cssLightRootRE.FindStringSubmatch(tokens)
	dark := cssDarkRootRE.FindStringSubmatch(tokens)
	if light == nil || dark == nil {
		t.Fatalf("tokens.css light :root found = %v, dark :root found = %v, want both", light != nil, dark != nil)
	}
	themes := map[string]map[string]string{"light": hexTokens(light[1]), "dark": hexTokens(dark[1])}

	var fg, bg string
	for _, d := range classDecls(parseRules(appCSS(t)), "field__error") {
		m := tokenValueRE.FindStringSubmatch(d.value)
		if m == nil {
			continue
		}
		switch d.prop {
		case "color":
			fg = m[1]
		case "background", "background-color":
			bg = m[1]
		}
	}
	if fg == "" || bg == "" {
		t.Fatalf("app.css .field__error color = %q, background = %q, want both set from tokens", fg, bg)
	}
	for name, hex := range themes {
		got, err := contrastRatio(hex[fg], hex[bg])
		if err != nil {
			t.Errorf("%s: contrastRatio(%s, %s) error = %v", name, fg, bg, err)
			continue
		}
		if got < minTextContrast {
			t.Errorf("%s: .field__error contrast(%s on %s) = %.2f, want >= %.1f", name, fg, bg, got, minTextContrast)
		}
	}
}
