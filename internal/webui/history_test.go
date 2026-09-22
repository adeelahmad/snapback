package webui

import (
	"html/template"
	"regexp"
	"strings"
	"testing"
)

const historyCSRF = "csrf-t5-token"

// historyView returns a History view with two roots, three ticks, two entries
// and a Versions panel for report.docx: one group of three likely identical
// occurrences and one group of one.
func historyView() HistoryView {
	return HistoryView{
		Chrome: Chrome{Title: "History", Active: "history", CSRFToken: historyCSRF},
		Roots: []RootItem{
			{Name: "home", Path: "/home/a", RepoState: "ok", MountState: "ready"},
			{Name: "media", Path: "/srv/media", RepoState: "offline", MountState: "stopped"},
		},
		Timeline: []SnapshotTick{
			{ID: "a", Time: "2026-09-18 09:00", Warm: true},
			{ID: "b", Time: "2026-09-19 09:00"},
			{ID: "c", Time: "2026-09-20 09:00", Warm: true, Selected: true},
		},
		Entries: []Entry{
			{Name: "notes.txt", Size: "4 KB", Modified: "2026-09-19 08:12", State: "ok"},
			{Name: "photos", Size: "", Modified: "2026-09-17 21:40", State: "ok"},
		},
		Versions: &VersionsPanel{
			File: "report.docx",
			Groups: []VersionGroup{
				{LikelyIdentical: true, Occurrences: []Occurrence{
					occurrence("s1", "2026-09-16 10:00"),
					occurrence("s2", "2026-09-17 10:00"),
					occurrence("s3", "2026-09-18 10:00"),
				}},
				{Occurrences: []Occurrence{occurrence("s4", "2026-09-20 10:00")}},
			},
		},
	}
}

func occurrence(id, when string) Occurrence {
	return Occurrence{
		SnapshotID:  id,
		Time:        when,
		Size:        "18 KB",
		Modified:    "2026-09-15 17:30",
		DownloadURL: "/api/download/" + id + "/report.docx",
		RestoreName: "report (" + when[:10] + ").docx",
	}
}

// htmlElement is one HTML htmlElement located in rendered output.
type htmlElement struct {
	html string
}

// htmlElements returns every tag htmlElement in out, outermost first, with nesting
// of the same tag honoured.
func htmlElements(out, tag string) []htmlElement {
	open := regexp.MustCompile(`(?i)<` + tag + `[\s>]`)
	tok := regexp.MustCompile(`(?i)<` + tag + `[\s>]|</` + tag + `\s*>`)
	var got []htmlElement
	for _, loc := range open.FindAllStringIndex(out, -1) {
		depth := 0
		for _, m := range tok.FindAllStringIndex(out[loc[0]:], -1) {
			if strings.HasPrefix(out[loc[0]+m[0]:], "</") {
				depth--
			} else {
				depth++
			}
			if depth == 0 {
				got = append(got, htmlElement{html: out[loc[0] : loc[0]+m[1]]})
				break
			}
		}
	}
	return got
}

// smallestElement returns the shortest htmlElement of any of tags whose HTML contains
// every one of needles, and whether one was found.
func smallestElement(out string, tags []string, needles ...string) (htmlElement, bool) {
	var best htmlElement
	found := false
	for _, tag := range tags {
		for _, e := range htmlElements(out, tag) {
			if !containsAllOf(e.html, needles) {
				continue
			}
			if !found || len(e.html) < len(best.html) {
				best, found = e, true
			}
		}
	}
	return best, found
}

func containsAllOf(s string, needles []string) bool {
	for _, n := range needles {
		if !strings.Contains(s, n) {
			return false
		}
	}
	return true
}

var historyTagRE = regexp.MustCompile(`<[^>]*>`)

// visibleText strips tags from an HTML fragment, leaving its visible text.
func visibleText(fragment string) string {
	return historyTagRE.ReplaceAllString(fragment, " ")
}

// versionsPanel isolates the htmlElement that holds the Versions heading.
func versionsPanel(t *testing.T, out string) string {
	t.Helper()
	e, ok := smallestElement(out, []string{"section", "aside", "div"}, ">Versions<", "report.docx")
	if !ok {
		t.Fatalf("render(history) has no section, aside or div holding the Versions heading and report.docx:\n%s", out)
	}
	return e.html
}

var (
	historyRowTags   = []string{"tr", "li", "article", "div", "form"}
	historyTickTags  = []string{"li", "button", "div", "a"}
	historyGroupTags = []string{"tbody", "section", "fieldset", "ul", "ol", "li", "article", "div", "table"}
)

func TestHistoryThreePanelLayout(t *testing.T) {
	v := historyView()
	out := render(t, mustLoad(t, ""), "history", v)

	nav, ok := smallestElement(out, []string{"nav"}, "home", "media")
	if !ok {
		t.Fatalf("render(history) has no <nav> listing roots home and media:\n%s", out)
	}
	navText := visibleText(nav.html)
	for _, want := range []string{"home", "media", "ok", "ready", "offline", "stopped"} {
		if !strings.Contains(navText, want) {
			t.Errorf("roots <nav> text = %q, want it to contain %q", navText, want)
		}
	}

	for _, tick := range v.Timeline {
		if !strings.Contains(out, tick.Time) {
			t.Errorf("render(history) lacks timeline tick time %q", tick.Time)
		}
	}
	if !strings.Contains(out, `role="slider"`) {
		for _, tick := range v.Timeline {
			if _, ok := smallestElement(out, []string{"button"}, tick.Time); !ok {
				t.Errorf("render(history) has no role=\"slider\" and no tick <button> carrying %q", tick.Time)
			}
		}
	}

	tables := htmlElements(out, "table")
	var list string
	for _, tb := range tables {
		if strings.Contains(tb.html, "notes.txt") {
			list = tb.html
		}
	}
	if list == "" {
		t.Fatalf("render(history) has no <table> listing the entries:\n%s", out)
	}
	for _, h := range []string{"Name", "Size", "Modified"} {
		if !regexp.MustCompile(`<th[^>]*>\s*` + h + `\s*</th>`).MatchString(list) {
			t.Errorf("file list table lacks header <th>%s</th>", h)
		}
	}
	for _, e := range v.Entries {
		if !strings.Contains(list, e.Name) {
			t.Errorf("file list table lacks entry %q", e.Name)
		}
	}
}

func TestHistoryWarmColdBadges(t *testing.T) {
	v := historyView()
	v.Timeline = []SnapshotTick{
		{ID: "a", Time: "2026-09-18 09:00", Warm: true},
		{ID: "b", Time: "2026-09-19 09:00"},
	}
	out := render(t, mustLoad(t, ""), "history", v)

	tests := []struct {
		time, want, notWant string
	}{
		{time: "2026-09-18 09:00", want: "warm", notWant: "cold"},
		{time: "2026-09-19 09:00", want: "cold", notWant: "warm"},
	}
	for _, tt := range tests {
		e, ok := smallestElement(out, historyTickTags, tt.time, ">"+tt.want)
		if !ok {
			t.Errorf("tick %q has no htmlElement carrying badge text %q", tt.time, tt.want)
			continue
		}
		if got := visibleText(e.html); strings.Contains(got, tt.notWant) {
			t.Errorf("tick %q text = %q, want no %q", tt.time, got, tt.notWant)
		}
	}
}

func TestHistoryAbsentDistinctFromFailed(t *testing.T) {
	v := historyView()
	v.Entries = []Entry{
		{Name: "gone", Size: "1 KB", Modified: "2026-09-01 10:00", State: "absent"},
		{Name: "broken", Size: "2 KB", Modified: "2026-09-02 10:00", State: "failed"},
	}
	out := render(t, mustLoad(t, ""), "history", v)

	tests := []struct {
		name, text, class, otherText, otherClass string
	}{
		{name: "gone", text: "absent", class: "state-absent", otherText: "read failed", otherClass: "state-failed"},
		{name: "broken", text: "read failed", class: "state-failed", otherText: "absent", otherClass: "state-absent"},
	}
	for _, tt := range tests {
		row, ok := smallestElement(out, []string{"tr"}, ">"+tt.name+"<")
		if !ok {
			t.Errorf("render(history) has no <tr> for entry %q", tt.name)
			continue
		}
		rowText := visibleText(row.html)
		if !strings.Contains(rowText, tt.text) {
			t.Errorf("row %q text = %q, want %q", tt.name, rowText, tt.text)
		}
		if !strings.Contains(row.html, tt.class) {
			t.Errorf("row %q = %q, want class %q", tt.name, row.html, tt.class)
		}
		if strings.Contains(rowText, tt.otherText) || strings.Contains(row.html, tt.otherClass) {
			t.Errorf("row %q = %q, want neither %q nor %q", tt.name, row.html, tt.otherText, tt.otherClass)
		}
	}
}

func TestHistoryVersionsListsEveryOccurrence(t *testing.T) {
	v := historyView()
	out := render(t, mustLoad(t, ""), "history", v)
	panel := versionsPanel(t, out)

	const restore = "Restore copy next to original"
	if got := strings.Count(panel, restore); got != 4 {
		t.Errorf("Versions panel has %d %q buttons, want 4", got, restore)
	}
	for _, g := range v.Versions.Groups {
		for _, o := range g.Occurrences {
			row, ok := smallestElement(panel, historyRowTags, o.Time, restore)
			if !ok {
				t.Errorf("Versions panel has no row with time %q and a %q button", o.Time, restore)
				continue
			}
			if !strings.Contains(row.html, `href="`+o.DownloadURL+`"`) || !strings.Contains(visibleText(row.html), "Download") {
				t.Errorf("occurrence %q row = %q, want a Download link to %q", o.SnapshotID, row.html, o.DownloadURL)
			}
			if !regexp.MustCompile(`<button[^>]*>[^<]*` + restore).MatchString(row.html) {
				t.Errorf("occurrence %q row = %q, want a <button> labelled %q", o.SnapshotID, row.html, restore)
			}
		}
	}
}

func TestHistoryVersionsLikelyIdenticalWording(t *testing.T) {
	out := render(t, mustLoad(t, ""), "history", historyView())
	panel := versionsPanel(t, out)
	body := visibleText(panel)

	const label = "likely identical (same size and modified time)"
	group, ok := smallestElement(panel, historyGroupTags, label, "2026-09-16 10:00", "2026-09-18 10:00")
	if !ok {
		t.Fatalf("Versions panel has no group labelled %q holding the three occurrences:\n%s", label, panel)
	}
	if strings.Contains(group.html, "2026-09-20 10:00") {
		t.Errorf("likely identical group = %q, want the single occurrence outside it", group.html)
	}

	lower := strings.ToLower(body)
	for i := strings.Index(lower, "identical"); i >= 0; {
		if !strings.HasSuffix(lower[:i], "likely ") {
			t.Errorf("Versions panel text has %q not preceded by \"likely \": %q", "identical", body)
			break
		}
		next := strings.Index(lower[i+1:], "identical")
		if next < 0 {
			break
		}
		i += next + 1
	}

	heading := strings.Index(body, "Versions")
	if heading < 0 {
		t.Fatalf("Versions panel text = %q, want the heading Versions", body)
	}
	rest := strings.ToLower(body[:heading] + body[heading+len("Versions"):])
	if strings.Contains(rest, "version") {
		t.Errorf("Versions panel body = %q, want no \"version\" besides the heading", body)
	}
}

func TestHistoryRestoreCopyNeverOverwrite(t *testing.T) {
	v := historyView()
	const name = "report (2026-09-20).docx"
	v.Versions.Groups = []VersionGroup{{Occurrences: []Occurrence{{
		SnapshotID:  "s9",
		Time:        "2026-09-20 10:00",
		Size:        "18 KB",
		Modified:    "2026-09-20 09:55",
		DownloadURL: "/api/download/s9/report.docx",
		RestoreName: name,
	}}}}
	out := render(t, mustLoad(t, ""), "history", v)

	form, ok := smallestElement(out, []string{"form"}, name)
	if !ok {
		t.Fatalf("render(history) has no <form> showing restore name %q:\n%s", name, out)
	}
	if !regexp.MustCompile(`(?i)<form[^>]*method="post"`).MatchString(form.html) {
		t.Errorf("restore form = %q, want method=\"post\"", form.html)
	}
	csrf := regexp.MustCompile(`<input[^>]*type="hidden"[^>]*name="csrf_token"[^>]*value="` + historyCSRF + `"` +
		`|<input[^>]*name="csrf_token"[^>]*type="hidden"[^>]*value="` + historyCSRF + `"`)
	if !csrf.MatchString(form.html) {
		t.Errorf("restore form = %q, want a hidden csrf_token input with value %q", form.html, historyCSRF)
	}
	lower := strings.ToLower(out)
	for _, banned := range []string{"overwrite", "replace"} {
		if strings.Contains(lower, banned) {
			t.Errorf("render(history) contains %q, want restore-copy only", banned)
		}
	}
}

func TestHistoryEscapesHostileFilenames(t *testing.T) {
	v := historyView()
	v.Roots[0].Name = hostile
	v.Entries[0].Name = hostile
	v.Versions.File = hostile
	out := render(t, mustLoad(t, ""), "history", v)

	escaped := template.HTMLEscapeString(hostile)
	if got := strings.Count(out, escaped); got < 3 {
		t.Errorf("render(history) has %d escaped hostile names %q, want at least 3", got, escaped)
	}
	if strings.Contains(out, "<script>alert(1)") {
		t.Errorf("render(history) contains raw %q, want it escaped everywhere", "<script>alert(1)")
	}
	for _, m := range regexp.MustCompile(`data-[\w-]+="[^"]*"`).FindAllString(out, -1) {
		if strings.Contains(m, "<script") {
			t.Errorf("data attribute %q holds a raw <script, want it escaped", m)
		}
	}
}

func TestHistoryOpenInFileManager(t *testing.T) {
	const label = "Open in file manager"
	p := mustLoad(t, "")

	v := historyView()
	v.FileManagerPath = "/home/a/docs"
	out := render(t, p, "history", v)
	if _, ok := smallestElement(out, []string{"button"}, label); !ok {
		t.Errorf("render(history) with FileManagerPath %q has no <button> %q", v.FileManagerPath, label)
	}
	if _, ok := smallestElement(out, []string{"form", "button"}, label, v.FileManagerPath); !ok {
		t.Errorf("render(history) has no %q button or form carrying path %q", label, v.FileManagerPath)
	}

	v.FileManagerPath = ""
	out = render(t, p, "history", v)
	if !strings.Contains(out, "report.docx") {
		t.Fatalf("render(history) with empty FileManagerPath lacks the Versions panel; cannot check the button is absent")
	}
	if strings.Contains(out, label) {
		t.Errorf("render(history) with empty FileManagerPath contains %q, want no button", label)
	}
}
