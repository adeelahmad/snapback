package webui

import (
	"fmt"
	"io/fs"
	"regexp"
	"strings"
	"testing"
)

var (
	tourOLRE      = regexp.MustCompile(`(?is)<ol\b[^>]*\bdata-tour\b[^>]*>`)
	tourLIRE      = regexp.MustCompile(`(?is)<li\b[^>]*>`)
	tourStepRE    = regexp.MustCompile(`(?i)\bdata-tour="([^"]*)"`)
	tourForRE     = regexp.MustCompile(`(?i)\bdata-tour-for="([^"]*)"`)
	tourDismissRE = regexp.MustCompile(`(?is)<button\b[^>]*\bdata-js="tour-dismiss"[^>]*>`)
	scriptSrcRE   = regexp.MustCompile(`(?is)<script\b[^>]*\bsrc=`)
	setupAnchorRE = regexp.MustCompile(`(?i)\b(?:id|name)="([^"{}]+)"`)
)

// setupAnchors returns the literal id and name attribute values the Setup page
// template declares, which are the only anchors a tour step may point at.
func setupAnchors(t *testing.T) map[string]bool {
	t.Helper()
	b, err := fs.ReadFile(embedded, "templates/setup.html")
	if err != nil {
		t.Fatalf("fs.ReadFile(embedded, %q) error = %v", "templates/setup.html", err)
	}
	anchors := make(map[string]bool)
	for _, m := range setupAnchorRE.FindAllStringSubmatch(string(b), -1) {
		anchors[m[1]] = true
	}
	if len(anchors) == 0 {
		t.Fatal("templates/setup.html declares no id= or name= attributes, want the setup form controls")
	}
	return anchors
}

// renderTour renders steps or fails the test.
func renderTour(t *testing.T, steps []TourStep) string {
	t.Helper()
	out, err := RenderTour(steps)
	if err != nil {
		t.Fatalf("RenderTour(%d steps) error = %v", len(steps), err)
	}
	return string(out)
}

func TestSetupTourAnchorsRealSetupFields(t *testing.T) {
	steps := SetupTour()
	if got, want := len(steps), 3; got < want {
		t.Fatalf("SetupTour() steps = %d, want at least %d", got, want)
	}
	anchors := setupAnchors(t)
	seen := make(map[string]bool, len(steps))
	for i, s := range steps {
		if strings.TrimSpace(s.For) == "" {
			t.Errorf("SetupTour()[%d].For = %q, want a setup field path", i, s.For)
			continue
		}
		if strings.TrimSpace(s.Title) == "" {
			t.Errorf("SetupTour()[%d].Title = %q, want a step heading", i, s.Title)
		}
		if strings.TrimSpace(s.Body) == "" {
			t.Errorf("SetupTour()[%d].Body = %q, want step copy", i, s.Body)
		}
		if seen[s.For] {
			t.Errorf("SetupTour()[%d].For = %q, want each step to anchor a different field", i, s.For)
		}
		seen[s.For] = true
		if !anchors[s.For] {
			t.Errorf("SetupTour()[%d].For = %q, want an id or name declared in templates/setup.html", i, s.For)
		}
	}
}

func TestRenderTourGoldenShape(t *testing.T) {
	steps := SetupTour()
	if len(steps) == 0 {
		t.Fatal("SetupTour() = no steps, want the first-run tour steps to render")
	}
	out := renderTour(t, steps)
	if got, want := len(tourOLRE.FindAllString(out, -1)), 1; got != want {
		t.Errorf("RenderTour() <ol data-tour> count = %d, want %d\nrendered: %s", got, want, out)
	}
	lis := tourLIRE.FindAllString(out, -1)
	if got, want := len(lis), len(steps); got != want {
		t.Fatalf("RenderTour() <li> count = %d, want %d (one per step)\nrendered: %s", got, want, out)
	}
	for i, li := range lis {
		wantStep := fmt.Sprint(i + 1)
		m := tourStepRE.FindStringSubmatch(li)
		if m == nil {
			t.Errorf("RenderTour() <li>[%d] = %q, want data-tour=%q", i, li, wantStep)
		} else if m[1] != wantStep {
			t.Errorf("RenderTour() <li>[%d] data-tour = %q, want %q (steps numbered 1..n in order)", i, m[1], wantStep)
		}
		f := tourForRE.FindStringSubmatch(li)
		if f == nil {
			t.Errorf("RenderTour() <li>[%d] = %q, want data-tour-for=%q", i, li, steps[i].For)
		} else if f[1] != steps[i].For {
			t.Errorf("RenderTour() <li>[%d] data-tour-for = %q, want %q", i, f[1], steps[i].For)
		}
	}
	if got, want := len(tourDismissRE.FindAllString(out, -1)), 1; got != want {
		t.Errorf(`RenderTour() <button data-js="tour-dismiss"> count = %d, want %d`+"\nrendered: %s", got, want, out)
	}
	if m := scriptSrcRE.FindString(out); m != "" {
		t.Errorf("RenderTour() contains %q, want no <script src> in the server-rendered partial", m)
	}
}

func TestRenderTourEscapesStepText(t *testing.T) {
	steps := []TourStep{
		{For: "repo-uri", Title: "Repository", Body: "Point at <b>your</b> restic repo " + hostile},
		{For: "roots", Title: "Roots", Body: "List the directories to watch"},
		{For: "restic-path", Title: "Restic", Body: "Pick the restic binary"},
	}
	out := renderTour(t, steps)
	for _, raw := range []string{"<b>your</b>", "<script>alert(1)</script>"} {
		if strings.Contains(out, raw) {
			t.Errorf("RenderTour() contains raw %q, want it escaped as text", raw)
		}
	}
	for _, want := range []string{"&lt;b&gt;your&lt;/b&gt;", "&lt;script&gt;alert(1)&lt;/script&gt;"} {
		if !strings.Contains(out, want) {
			t.Errorf("RenderTour() missing escaped %q\nrendered: %s", want, out)
		}
	}
}
