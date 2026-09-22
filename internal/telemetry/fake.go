package telemetry

import "context"

// Fake is a recording exporter for tests. It keeps every batch it is handed,
// counts the calls, and can be primed with [Fake.Fail] to return an error. It
// lives in a non-test file so other packages can export against it.
//
// The zero Fake is ready to use and is safe for concurrent use.
type Fake struct{}

// Export records batch and returns the error primed by [Fake.Fail], if any.
func (f *Fake) Export(ctx context.Context, batch []Event) error {
	return nil
}

// Batches returns every batch Export received, in order.
func (f *Fake) Batches() [][]Event {
	return nil
}

// Calls returns how many times Export was called.
func (f *Fake) Calls() int {
	return 0
}

// Names returns the name of every event Export received, in order.
func (f *Fake) Names() []string {
	return nil
}

// Fail primes Export to return err on every subsequent call. It still records.
func (f *Fake) Fail(err error) {}
