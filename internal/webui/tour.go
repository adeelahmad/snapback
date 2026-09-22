package webui

import "html/template"

// TourStep is one step of the first-run tour: the setup control it points at,
// its heading and its body copy.
type TourStep struct {
	For   string
	Title string
	Body  string
}

// SetupTour returns the steps of the first-run tour over the Setup page.
func SetupTour() []TourStep {
	return nil
}

// RenderTour renders steps as the ordered tour list partial.
func RenderTour(steps []TourStep) (template.HTML, error) {
	return "", nil
}
