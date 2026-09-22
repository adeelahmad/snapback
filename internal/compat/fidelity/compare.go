package fidelity

import (
	"errors"
	"fmt"
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
	const modeMask = fs.ModeType | fs.ModePerm
	delta := observed.MTime.Sub(expected.MTime)
	return Result{
		Path:       expected.Path,
		SizeOK:     expected.Size == observed.Size,
		ModeOK:     expected.Mode&modeMask == observed.Mode&modeMask,
		MTimeOK:    delta <= tol && -delta <= tol,
		MTimeDelta: delta,
	}
}

// CompareAll compares every expected file against its observed metadata,
// returning one result per expected file in input order.
func CompareAll(expected []Meta, observed map[string]Meta, tol time.Duration) ([]Result, error) {
	if len(expected) == 0 {
		return nil, errors.New("no expected files to compare")
	}
	results := make([]Result, 0, len(expected))
	for _, e := range expected {
		o, ok := observed[e.Path]
		if !ok {
			return nil, fmt.Errorf("no observed metadata for %q", e.Path)
		}
		results = append(results, Compare(e, o, tol))
	}
	return results, nil
}

// Precision returns the coarsest of 1ns, 1µs, 1ms and 1s that divides the
// nanosecond part of every timestamp in ts.
func Precision(ts []time.Time) time.Duration {
	for _, p := range []time.Duration{time.Second, time.Millisecond, time.Microsecond} {
		divides := true
		for _, t := range ts {
			if time.Duration(t.Nanosecond())%p != 0 {
				divides = false
				break
			}
		}
		if divides {
			return p
		}
	}
	return time.Nanosecond
}
