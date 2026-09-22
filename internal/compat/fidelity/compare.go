package fidelity

import (
	"io/fs"
	"time"
)

// MTimeTolerance is the allowed mtime difference between expected and
// observed metadata. Changing it is a human decision.
const MTimeTolerance time.Duration = 0

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

// Compare reports whether observed matches expected in size, mode and mtime,
// with mtime allowed to differ by at most tol.
func Compare(expected, observed Meta, tol time.Duration) Result {
	panic("SUB-AGENT-TODO: T1 pure compare; Path from expected; SizeOK on Size; ModeOK on Mode&(fs.ModeType|fs.ModePerm); MTimeDelta = observed.MTime.Sub(expected.MTime); MTimeOK when |delta| <= tol")
}

// CompareAll compares every expected file against its observed metadata,
// returning one result per expected file in input order.
func CompareAll(expected []Meta, observed map[string]Meta, tol time.Duration) ([]Result, error) {
	panic("SUB-AGENT-TODO: T1 error on empty expected (M-002) and on any expected path missing from observed; else one Compare result per expected file, in input order")
}

// Precision returns the coarsest of 1ns, 1µs, 1ms and 1s that divides the
// nanosecond part of every timestamp in ts.
func Precision(ts []time.Time) time.Duration {
	panic("SUB-AGENT-TODO: T1 coarsest of 1s, 1ms, 1µs, 1ns dividing every t.Nanosecond()")
}
