package seed

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

const benchDirs = 2000

// mkBenchTree creates a depth-3 tree under root holding benchDirs directories
// in total, root included.
func mkBenchTree(b *testing.B, root string) {
	b.Helper()
	made := 1
	mkdir := func(rel string) {
		if made == benchDirs {
			return
		}
		if err := os.Mkdir(filepath.Join(root, rel), 0o755); err != nil {
			b.Fatalf("os.Mkdir(%q) error = %v, want nil", rel, err)
		}
		made++
	}
	for i := range 10 {
		mkdir(fmt.Sprintf("d%d", i))
	}
	for i := range 10 {
		for j := range 10 {
			mkdir(filepath.Join(fmt.Sprintf("d%d", i), fmt.Sprintf("e%d", j)))
		}
	}
	for i := range 10 {
		for j := range 10 {
			for k := range 20 {
				mkdir(filepath.Join(fmt.Sprintf("d%d", i), fmt.Sprintf("e%d", j), fmt.Sprintf("f%d", k)))
			}
		}
	}
	if made != benchDirs {
		b.Fatalf("mkBenchTree made %d dirs, want %d", made, benchDirs)
	}
}

func BenchmarkSeedRun(b *testing.B) {
	root := b.TempDir()
	mkBenchTree(b, root)
	p, err := PlanPath(root, "", 3, nil)
	if err != nil {
		b.Fatalf("PlanPath(%q, \"\", 3, nil) error = %v, want nil", root, err)
	}
	if p.Count != benchDirs {
		b.Fatalf("PlanPath(%q, \"\", 3, nil).Count = %d, want %d", root, p.Count, benchDirs)
	}
	l := &fakeLinker{}
	ctx := context.Background()
	var dirs int
	var elapsed time.Duration

	b.ResetTimer()
	for range b.N {
		r, err := Run(ctx, l, p)
		if err != nil {
			b.Fatalf("Run(plan) error = %v, want nil", err)
		}
		got := r.Linked + r.Existing
		if got != benchDirs {
			b.Fatalf("Run(plan).Linked+Existing = %d, want %d", got, benchDirs)
		}
		dirs += got
		elapsed += r.Elapsed
	}
	b.StopTimer()
	if elapsed > 0 {
		b.ReportMetric(float64(dirs)/elapsed.Seconds(), "dirs/s")
	}
}
