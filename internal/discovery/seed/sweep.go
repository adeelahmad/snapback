package seed

import (
	"context"
	"time"
)

// Spec describes one root to sweep.
type Spec struct {
	Root     string
	SeedPath string
	MaxDepth int
	Excludes []string
}

// Sweep builds a plan for each spec and runs it, merging the reports.
func Sweep(ctx context.Context, l Linker, specs []Spec) (Report, error) {
	panic("SUB-AGENT-TODO: T3 — for each Spec build its Plan (Root, SeedPath, MaxDepth, Excludes) and Run it with l; merge counts and Failures into one Report with total Elapsed/DirsPerSec; stop on ctx cancellation")
}

// SweepLoop runs Sweep every interval until ctx is done, passing each report to onReport.
func SweepLoop(ctx context.Context, interval time.Duration, l Linker, specs []Spec, onReport func(Report)) {
	panic("SUB-AGENT-TODO: T3 — ticker at interval; on each tick call Sweep(ctx, l, specs) and onReport(report) if non-nil; return promptly when ctx is done; stop the ticker")
}
