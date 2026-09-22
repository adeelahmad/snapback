package telemetry

import "context"

// Exporter delivers a batch of events to a telemetry backend.
//
// Implementations must be safe for concurrent use and must treat a nil or
// empty batch as a successful no-op.
type Exporter interface {
	// Export delivers events, honouring ctx for cancellation.
	Export(ctx context.Context, events []Event) error
}

// Nop is an [Exporter] that discards every batch.
type Nop struct{}

var _ Exporter = Nop{}

// Export discards events and reports success without allocating.
func (Nop) Export(_ context.Context, _ []Event) error { return nil }
