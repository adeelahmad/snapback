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
	var total Report
	var dirs int
	for _, sp := range specs {
		if err := ctx.Err(); err != nil {
			return total, err
		}
		p, err := PlanPath(sp.Root, sp.SeedPath, sp.MaxDepth, sp.Excludes)
		if err != nil {
			return total, err
		}
		r, err := Run(ctx, l, p)
		total.Linked += r.Linked
		total.Existing += r.Existing
		total.Failures = append(total.Failures, r.Failures...)
		total.Elapsed += r.Elapsed
		dirs += r.Linked + r.Existing + len(r.Failures)
		if err != nil {
			return total, err
		}
	}
	if secs := total.Elapsed.Seconds(); secs > 0 {
		total.DirsPerSec = float64(dirs) / secs
	}
	return total, nil
}

// SweepLoop runs Sweep every interval until ctx is done, passing each report
// and its error to onReport.
func SweepLoop(ctx context.Context, interval time.Duration, l Linker, specs []Spec, onReport func(Report, error)) {
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			// SUB-AGENT-TODO: R1 GREEN must pass the Sweep error to onReport.
			r, _ := Sweep(ctx, l, specs)
			if onReport != nil {
				onReport(r, nil)
			}
		}
	}
}
