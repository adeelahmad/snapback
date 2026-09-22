package links

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/resolver"
)

func BenchmarkEnsureLink(b *testing.B) {
	r := b.TempDir()
	pol := Policy{
		LinkName:     ".snapshot",
		HistoryMount: b.TempDir(),
		Roots:        []resolver.RootSpec{{ID: "home", LocalPath: r}},
	}
	reg, err := OpenRegistry(filepath.Join(b.TempDir(), "links.db"))
	if err != nil {
		b.Fatalf("OpenRegistry() = %v", err)
	}
	b.Cleanup(func() { _ = reg.Close() })
	e := NewEngine(reg, pol)

	dirs := make([]string, b.N)
	for i := range dirs {
		dirs[i] = filepath.Join(r, fmt.Sprintf("d%d", i))
		if err := os.Mkdir(dirs[i], 0o755); err != nil {
			b.Fatalf("Mkdir(%q) = %v", dirs[i], err)
		}
	}
	ctx := context.Background()
	durs := make([]time.Duration, 0, b.N)

	b.ResetTimer()
	for _, dir := range dirs {
		start := time.Now()
		if _, err := e.Ensure(ctx, dir); err != nil {
			b.Fatalf("Ensure(%q) = %v, want nil", dir, err)
		}
		durs = append(durs, time.Since(start))
	}
	b.StopTimer()

	slices.Sort(durs)
	// Nearest-rank p95: the ceil(0.95*n)-th smallest sample.
	p95 := durs[(len(durs)*95+99)/100-1]
	b.ReportMetric(float64(p95.Nanoseconds()), "p95-ns/op")
}
