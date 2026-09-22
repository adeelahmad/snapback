package fidelity

import "time"

// Observed is the metadata read from one path, plus times that are recorded
// but never asserted.
type Observed struct {
	Meta
	CTime     time.Time
	BirthTime *time.Time
}

// Observe reads the metadata of path without following its final component.
func Observe(path string) (Observed, error) {
	panic("SUB-AGENT-TODO: T3 os.Lstat(path) (never follow final component); fill Meta{Path, Size, Mode, MTime}; os.Readlink for symlinks into LinkTarget; CTime/BirthTime from the per-platform statTimes helper")
}
