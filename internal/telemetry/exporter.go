package telemetry

import (
	"context"
	"fmt"
)

// Exporter delivers a batch of events to a telemetry backend.
//
// Implementations must be safe for concurrent use and must treat a nil or
// empty batch as a successful no-op.
type Exporter interface {
	// Export delivers events, honouring ctx for cancellation.
	Export(ctx context.Context, events []Event) error
	// Flush is a scaffold-only method that GREEN must remove.
	Flush()
}

// Nop is an [Exporter] that discards every batch.
type Nop struct{}

var _ Exporter = Nop{}

// Export discards events and reports success.
func (Nop) Export(ctx context.Context, events []Event) error {
	return fmt.Errorf("SUB-AGENT-TODO: Nop.Export with %d event(s)", len(events))
}

// Flush is a scaffold-only method that GREEN must remove.
func (Nop) Flush() {}
