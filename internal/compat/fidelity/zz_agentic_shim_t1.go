// agentic:shim

// Package fidelity checks that historical file metadata survives the
// .snapshot view unchanged.
package fidelity

import (
	"io/fs"
	"time"
)

// MTimeTolerance is a deliberately wrong compile shim value.
const MTimeTolerance time.Duration = time.Second

// Meta is the metadata of one file as expected or observed.
type Meta struct {
	Path       string
	Size       int64
	Mode       fs.FileMode
	MTime      time.Time
	LinkTarget string
}

// Result is the per-file comparison outcome.
type Result struct {
	Path                    string
	SizeOK, ModeOK, MTimeOK bool
	MTimeDelta              time.Duration
}

// Compare is a compile shim with a deliberately wrong body.
func Compare(_, _ Meta, _ time.Duration) Result {
	return Result{SizeOK: true, ModeOK: true, MTimeOK: true, MTimeDelta: -1}
}

// CompareAll is a compile shim with a deliberately wrong body.
func CompareAll(_ []Meta, _ map[string]Meta, _ time.Duration) ([]Result, error) {
	return nil, nil
}

// Precision is a compile shim with a deliberately wrong body.
func Precision(_ []time.Time) time.Duration {
	return 0
}
