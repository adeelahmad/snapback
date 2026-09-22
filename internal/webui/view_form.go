package webui

import "strings"

// Kind is the control a configuration field is rendered with.
type Kind string

// The controls the configuration form is built from.
const (
	KindText     Kind = "text"
	KindPassword Kind = "password"
	KindNumber   Kind = "number"
	KindDuration Kind = "duration"
	KindSelect   Kind = "select"
	KindToggle   Kind = "toggle"
	KindChips    Kind = "chips"
	KindRow      Kind = "row"
)

// Option is one choice of a select control.
type Option struct {
	Value string
	Label string
}

// Field is one configuration key as the templates render it. Path is the YAML
// key path, such as repositories[0].lock_mode, and is also the form input name.
// Value carries single-valued controls; Values carries chip lists.
type Field struct {
	Path    string
	Kind    Kind
	Value   string
	Values  []string
	Options []Option
	Label   string
	Help    string
}

// Section is a group of fields sharing a top-level configuration key.
type Section struct {
	Title  string
	Fields []Field
}

// Sections groups fields by their top-level key, keeping the given order.
// Top-level scalars share one "General" section.
func Sections(fields []Field) []Section {
	var out []Section
	at := map[string]int{}
	for _, f := range fields {
		key := sectionKey(f.Path)
		i, ok := at[key]
		if !ok {
			i = len(out)
			at[key] = i
			out = append(out, Section{Title: sectionTitle(key)})
		}
		out[i].Fields = append(out[i].Fields, f)
	}
	return out
}

// sectionKey is the top-level configuration key a field path belongs to.
func sectionKey(path string) string {
	head, _, _ := strings.Cut(path, ".")
	name, _, indexed := strings.Cut(head, "[")
	if name == path && !indexed {
		return "general"
	}
	return name
}

func sectionTitle(key string) string {
	s := strings.ReplaceAll(key, "_", " ")
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
