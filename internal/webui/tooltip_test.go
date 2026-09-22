package webui

import (
	"regexp"
	"strings"
	"testing"
)

// controlPartials names every partial defined in templates/controls.html.
var controlPartials = []string{
	"control-input",
	"control-password",
	"control-number",
	"control-duration",
	"control-select",
	"control-toggle",
	"control-chips",
	"control-row",
}

// ariaDescribedBy matches one aria-describedby attribute and captures its
// token list.
var ariaDescribedBy = regexp.MustCompile(`aria-describedby="([^"]*)"`)

// describedByTokens returns the token list of every aria-describedby
// attribute in src.
func describedByTokens(src string) [][]string {
	var out [][]string
	for _, m := range ariaDescribedBy.FindAllStringSubmatch(src, -1) {
		out = append(out, strings.Fields(m[1]))
	}
	return out
}

// hasToken reports whether tokens contains want.
func hasToken(tokens []string, want string) bool {
	for _, tok := range tokens {
		if tok == want {
			return true
		}
	}
	return false
}

// TestControlPartialTooltipsWithHelp pins the tooltip contract: a control
// whose Help is set carries that same text as a title attribute, points at
// its own help element with aria-describedby, and renders that help element
// exactly once. The id scheme is the one controls.html already uses,
// f-<path>-help on a <p class="field__help">, not the plan's <small>.
func TestControlPartialTooltipsWithHelp(t *testing.T) {
	const (
		path = "catalog.max_entries"
		help = "Entries kept in the catalog."
	)
	helpID := "f-" + path + "-help"
	view := controlView{
		Path:    path,
		Label:   "Maximum entries",
		Help:    help,
		Value:   "500",
		Options: []controlOption{{Value: "500", Label: "Five hundred"}},
	}

	for _, name := range controlPartials {
		t.Run(name, func(t *testing.T) {
			got := renderControl(t, name, view)

			if want := `title="` + help + `"`; !strings.Contains(got, want) {
				t.Errorf("%s render does not contain %s, want the Help text as a tooltip; got:\n%s", name, want, got)
			}

			tokens := describedByTokens(got)
			if len(tokens) == 0 {
				t.Fatalf("%s render has no aria-describedby attribute, want one naming %q; got:\n%s", name, helpID, got)
			}
			for i, toks := range tokens {
				if !hasToken(toks, helpID) {
					t.Errorf("%s aria-describedby #%d = %q, want a token %q", name, i, strings.Join(toks, " "), helpID)
				}
			}

			if n := strings.Count(got, `id="`+helpID+`"`); n != 1 {
				t.Errorf("%s render has %d elements with id=%q, want exactly 1; got:\n%s", name, n, helpID, got)
			}
			if !strings.Contains(got, `<p class="field__help" id="`+helpID+`">`+help+`</p>`) {
				t.Errorf("%s render does not contain the field__help paragraph for %q; got:\n%s", name, helpID, got)
			}
		})
	}
}

// TestControlPartialTooltipsWithoutHelp pins the other half: with no Help
// there is no tooltip, no help element and no aria-describedby pointing at
// an element that does not exist.
func TestControlPartialTooltipsWithoutHelp(t *testing.T) {
	const path = "catalog.max_entries"
	helpID := "f-" + path + "-help"
	view := controlView{
		Path:    path,
		Label:   "Maximum entries",
		Value:   "500",
		Options: []controlOption{{Value: "500", Label: "Five hundred"}},
	}

	for _, name := range controlPartials {
		t.Run(name, func(t *testing.T) {
			got := renderControl(t, name, view)

			if strings.Contains(got, "title=") {
				t.Errorf("%s render contains a title attribute with no Help set, want none; got:\n%s", name, got)
			}
			if tokens := describedByTokens(got); len(tokens) != 0 {
				t.Errorf("%s render has %d aria-describedby attributes with no Help set, want none; got:\n%s", name, len(tokens), got)
			}
			if strings.Contains(got, helpID) {
				t.Errorf("%s render still names %q with no Help set, want no help element and no reference to one; got:\n%s", name, helpID, got)
			}
		})
	}
}
