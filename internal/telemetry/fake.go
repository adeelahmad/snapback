package telemetry

import (
	"context"
	"slices"
	"sync"
)

// Fake is a recording exporter for tests. It keeps every batch it is handed,
// counts the calls, and can be primed with [Fake.Fail] to return an error. It
// lives in a non-test file so other packages can export against it.
//
// The zero Fake is ready to use and is safe for concurrent use.
type Fake struct {
	mu      sync.Mutex
	batches [][]Event
	calls   int
	err     error
}

// Export records batch and returns the error primed by [Fake.Fail], if any.
func (f *Fake) Export(_ context.Context, batch []Event) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.batches = append(f.batches, slices.Clone(batch))
	f.calls++
	return f.err
}

// Batches returns every batch Export received, in order. The result and its
// batches are copies, so callers cannot disturb what the Fake recorded.
func (f *Fake) Batches() [][]Event {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([][]Event, len(f.batches))
	for i, batch := range f.batches {
		out[i] = slices.Clone(batch)
	}
	return out
}

// Calls returns how many times Export was called.
func (f *Fake) Calls() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls
}

// Names returns the name of every event Export received, in order, flattened
// across batches.
func (f *Fake) Names() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []string
	for _, batch := range f.batches {
		for _, e := range batch {
			out = append(out, e.Name)
		}
	}
	return out
}

// Fail primes Export to return err on every subsequent call. It still records.
func (f *Fake) Fail(err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.err = err
}
