// agentic:shim

package fidelity

import "time"

// Observed is the metadata read from one path, plus times that are recorded
// but never asserted.
type Observed struct {
	Meta
	CTime     time.Time
	BirthTime *time.Time
}

// Fixtures is a compile shim with a deliberately wrong body.
func Fixtures() []Meta {
	return []Meta{{Path: "shim.txt", Size: -1}}
}

// Observe is a compile shim with a deliberately wrong body.
func Observe(path string) (Observed, error) {
	return Observed{Meta: Meta{Path: path, Size: -1}}, nil
}
