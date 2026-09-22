package webui

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
func Sections(fields []Field) []Section {
	return nil
}
