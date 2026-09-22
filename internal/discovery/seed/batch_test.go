package seed

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/adeelahmad/snapback/internal/links"
)

// batchLinker fakes a Linker that also implements BatchLinker. A dir whose
// base name starts with "fail-" fails and one starting with "old-" exists.
type batchLinker struct {
	mu          sync.Mutex
	batches     [][]string
	ensureCalls int
	afterBatch  func(n int)
}

var _ BatchLinker = (*batchLinker)(nil)

func (b *batchLinker) Ensure(_ context.Context, dir string) (links.Result, error) {
	b.mu.Lock()
	b.ensureCalls++
	b.mu.Unlock()
	return links.Result{Created: true, Path: dir}, nil
}

func (b *batchLinker) EnsureBatch(_ context.Context, dirs []string) ([]links.Result, error) {
	b.mu.Lock()
	b.batches = append(b.batches, append([]string(nil), dirs...))
	n := len(b.batches)
	b.mu.Unlock()
	res := make([]links.Result, len(dirs))
	var errs []error
	for i, d := range dirs {
		base := d[strings.LastIndex(d, "/")+1:]
		switch {
		case strings.HasPrefix(base, "fail-"):
			errs = append(errs, &links.EnsureError{Dir: d, Err: errors.New("boom")})
		case strings.HasPrefix(base, "old-"):
			res[i] = links.Result{Path: d}
		default:
			res[i] = links.Result{Created: true, Path: d}
		}
	}
	if b.afterBatch != nil {
		b.afterBatch(n)
	}
	return res, errors.Join(errs...)
}

func planOf(n int, name func(i int) string) Plan {
	dirs := make([]string, n)
	for i := range dirs {
		dirs[i] = "/root/" + name(i)
	}
	return Plan{Dirs: dirs, Count: n}
}

func plainName(i int) string { return fmt.Sprintf("d%d", i) }

func TestRunEnsuresInBatches(t *testing.T) {
	l := &batchLinker{}
	p := planOf(600, plainName)

	r, err := Run(context.Background(), l, p)
	if err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}
	if got, want := len(l.batches), 3; got != want {
		t.Errorf("Run(600 dirs) EnsureBatch calls = %d, want %d", got, want)
	}
	if got := l.ensureCalls; got != 0 {
		t.Errorf("Run(600 dirs) Ensure calls = %d, want 0", got)
	}
	var seen []string
	for i, b := range l.batches {
		if len(b) > 256 {
			t.Errorf("batch %d size = %d, want <= 256", i, len(b))
		}
		seen = append(seen, b...)
	}
	if got, want := strings.Join(seen, ","), strings.Join(p.Dirs, ","); got != want {
		t.Errorf("Run(600 dirs) batched dirs are not the plan dirs in order")
	}
	if got, want := r.Linked, 600; got != want {
		t.Errorf("Run(600 dirs).Linked = %d, want %d", got, want)
	}
}

func TestRunBatchReportCounts(t *testing.T) {
	l := &batchLinker{}
	p := planOf(300, func(i int) string {
		switch i % 3 {
		case 0:
			return fmt.Sprintf("new-%d", i)
		case 1:
			return fmt.Sprintf("old-%d", i)
		default:
			return fmt.Sprintf("fail-%d", i)
		}
	})

	r, err := Run(context.Background(), l, p)
	if err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}
	if got := l.ensureCalls; got != 0 {
		t.Errorf("Run() Ensure calls = %d, want 0", got)
	}
	if r.Linked != 100 || r.Existing != 100 || len(r.Failures) != 100 {
		t.Errorf("Run() = Linked %d, Existing %d, Failures %d, want 100, 100, 100",
			r.Linked, r.Existing, len(r.Failures))
	}
	for _, f := range r.Failures {
		if !strings.Contains(f.Dir, "/fail-") {
			t.Errorf("Failure.Dir = %q, want a fail- dir", f.Dir)
		}
		var ee *links.EnsureError
		if !errors.As(f.Err, &ee) || ee.Dir != f.Dir {
			t.Errorf("Failure{%q}.Err = %v, want *links.EnsureError for that dir", f.Dir, f.Err)
		}
	}
}

// plainLinker has no EnsureBatch.
type plainLinker struct{ calls int }

func (p *plainLinker) Ensure(_ context.Context, dir string) (links.Result, error) {
	p.calls++
	return links.Result{Created: true, Path: dir}, nil
}

func TestRunFallsBackToEnsure(t *testing.T) {
	l := &plainLinker{}
	p := planOf(600, plainName)

	r, err := Run(context.Background(), l, p)
	if err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}
	if got, want := l.calls, 600; got != want {
		t.Errorf("Run(600 dirs) Ensure calls = %d, want %d", got, want)
	}
	if got, want := r.Linked, 600; got != want {
		t.Errorf("Run(600 dirs).Linked = %d, want %d", got, want)
	}
}

func TestRunBatchStopsOnCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	l := &batchLinker{afterBatch: func(n int) {
		if n == 1 {
			cancel()
		}
	}}
	p := planOf(600, plainName)

	r, err := Run(ctx, l, p)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Run() error = %v, want context.Canceled", err)
	}
	if got, want := len(l.batches), 1; got != want {
		t.Errorf("Run() EnsureBatch calls after cancel = %d, want %d", got, want)
	}
	if got := l.ensureCalls; got != 0 {
		t.Errorf("Run() Ensure calls = %d, want 0", got)
	}
	if got, want := r.Linked, 256; got != want {
		t.Errorf("Run().Linked = %d, want %d", got, want)
	}
}
