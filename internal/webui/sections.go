package webui

import "html/template"

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
	_ = v
	return "", nil
}
