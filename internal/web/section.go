package web

import "github.com/adeelahmad/snapback/internal/webui"

// Group is one half of the configuration form: the fields it shows and how
// many they are, so a template can label a collapsed group without counting.
type Group struct {
	Fields []webui.Field
	Count  int
}

// Split partitions the configuration fields into the few a first run has to
// answer and everything else, keeping the order of fields.
func Split(fields []webui.Field) (basic, advanced Group) {
	return Group{}, Group{}
}
