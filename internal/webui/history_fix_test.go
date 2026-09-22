package webui

import (
	"strings"
	"testing"
)

func TestHistoryStateCellRendersValue(t *testing.T) {
	v := historyView()
	out := render(t, mustLoad(t, ""), "history", v)

	row, ok := smallestElement(out, []string{"tr"}, ">notes.txt<")
	if !ok {
		t.Fatalf("render(history) has no <tr> for entry %q", "notes.txt")
	}
	if want := `<td><span class="badge-ok">ok</span></td>`; !strings.Contains(row.html, want) {
		t.Errorf("row %q = %q, want it to contain %q", "notes.txt", row.html, want)
	}
}

func TestHistorySnapshotFormCarriesRootAndPath(t *testing.T) {
	v := historyView()
	v.Root, v.Path = "home", "docs/sub"
	out := render(t, mustLoad(t, ""), "history", v)

	for _, tick := range v.Timeline {
		form, ok := smallestElement(out, []string{"form"}, `value="`+tick.ID+`"`)
		if !ok {
			t.Errorf("render(history) has no <form> for snapshot %q", tick.ID)
			continue
		}
		for _, want := range []string{
			`<input type="hidden" name="root" value="home">`,
			`<input type="hidden" name="path" value="docs/sub">`,
		} {
			if !strings.Contains(form.html, want) {
				t.Errorf("snapshot %q form = %q, want it to contain %q", tick.ID, form.html, want)
			}
		}
	}
}

// TestHistoryPanelsKeepPadding wants every rule that styles the roots, files
// or versions panels to leave the panel padding at var(--space-4).
func TestHistoryPanelsKeepPadding(t *testing.T) {
	for _, r := range parseRules(appCSS(t)) {
		for _, sel := range strings.Split(r.selector, ",") {
			sel = strings.TrimSpace(sel)
			if sel != ".roots" && sel != ".files" && sel != ".versions" {
				continue
			}
			for _, d := range r.decls {
				if d.prop == "padding" && d.value != "var(--space-4)" {
					t.Errorf("rule %q padding = %q, want %q", r.selector, d.value, "var(--space-4)")
				}
			}
		}
	}
}
