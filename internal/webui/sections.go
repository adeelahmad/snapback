package webui

import (
	"bytes"
	"fmt"
	"html/template"
)

// FormSections is the basic/advanced split of a configuration form: Basic is
// always visible, Advanced is collapsed behind a disclosure whose badge shows
// AdvancedCount, the number of fields hidden inside it.
type FormSections struct {
	Basic         []ConfigSection
	Advanced      []ConfigSection
	AdvancedCount int
}

// RenderFormSections renders v as the open basic block followed by the
// collapsed advanced disclosure.
func RenderFormSections(v FormSections) (template.HTML, error) {
	t, err := template.ParseFS(embedded, "templates/sections.html", "templates/controls.html")
	if err != nil {
		return "", fmt.Errorf("webui: parse form sections: %w", err)
	}
	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, "form-sections", v); err != nil {
		return "", fmt.Errorf("webui: render form sections: %w", err)
	}
	return template.HTML(buf.String()), nil // #nosec G203 -- built from the embedded templates, which escape the control text
}
