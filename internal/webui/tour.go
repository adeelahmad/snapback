package webui

import (
	"bytes"
	"fmt"
	"html/template"
)

// TourStep is one step of the first-run tour: the setup control it points at,
// its heading and its body copy.
type TourStep struct {
	For   string
	Title string
	Body  string
}

// tourView is one step as the tour partial renders it, with its 1-based number.
type tourView struct {
	TourStep
	N int
}

// SetupTour returns the steps of the first-run tour over the Setup page.
func SetupTour() []TourStep {
	return []TourStep{
		{
			For:   "repo-uri",
			Title: "Your backup repository",
			Body:  "Where your restic backups live, for example a local path or a remote like s3:... — Snapback only ever reads it.",
		},
		{
			For:   "credential-file",
			Title: "Repository password file",
			Body:  "A file holding the password that unlocks the repository, so Snapback can read snapshots without asking every time.",
		},
		{
			For:   "restic-path",
			Title: "The restic binary",
			Body:  "The restic program Snapback runs. Leave it as it is unless restic lives somewhere off your PATH.",
		},
		{
			For:   "roots",
			Title: "Directories to cover",
			Body:  "The directories that get a .snapshot entry, one per line. Start with the folders you would most want to restore.",
		},
	}
}

// RenderTour renders steps as the ordered tour list partial.
func RenderTour(steps []TourStep) (template.HTML, error) {
	t, err := template.ParseFS(embedded, "templates/tour.html")
	if err != nil {
		return "", fmt.Errorf("webui: parse tour: %w", err)
	}
	views := make([]tourView, len(steps))
	for i, s := range steps {
		views[i] = tourView{TourStep: s, N: i + 1}
	}
	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, "tour", views); err != nil {
		return "", fmt.Errorf("webui: render tour: %w", err)
	}
	return template.HTML(buf.String()), nil // #nosec G203 -- built from the embedded template, which escapes the step text
}
