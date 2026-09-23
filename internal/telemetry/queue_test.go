package telemetry_test

import (
	"context"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/telemetry"
)

// sleepExporter wraps another [telemetry.Exporter] and sleeps before
// delegating, so tests can simulate a slow backend without touching Fake.
type sleepExporter struct {
	inner telemetry.Exporter
	delay time.Duration
}

func (s *sleepExporter) Export(ctx context.Context, events []telemetry.Event) error {
	time.Sleep(s.delay)
	return s.inner.Export(ctx, events)
}

func TestQueue_NonBlockingEmitWithDropCounter(t *testing.T) {
	fake := &telemetry.Fake{}
	exp := &sleepExporter{inner: fake, delay: 10 * time.Millisecond}
	q := telemetry.NewQueue(exp, 10)

	start := time.Now()
	for i := 0; i < 1000; i++ {
		q.Emit(telemetry.Event{Name: "error"})
	}
	elapsed := time.Since(start)
	if elapsed > 50*time.Millisecond {
		t.Fatalf("1000 Emits took %s, want <= 50ms", elapsed)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := q.Close(ctx); err != nil {
		t.Fatalf("Close: %v", err)
	}

	delivered := 0
	for _, batch := range fake.Batches() {
		delivered += len(batch)
	}
	if got := delivered + q.Dropped(); got != 1000 {
		t.Fatalf("delivered(%d) + Dropped(%d) = %d, want 1000", delivered, q.Dropped(), got)
	}
}

func TestQueue_CloseFlushesQueuedWithinDeadline(t *testing.T) {
	fake := &telemetry.Fake{}
	q := telemetry.NewQueue(fake, 10)

	want := []telemetry.Event{
		{Name: "setup.completed"},
		{Name: "daemon.started"},
		{Name: "mount.ready"},
	}
	for _, ev := range want {
		q.Emit(ev)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	if err := q.Close(ctx); err != nil {
		t.Fatalf("Close: %v", err)
	}

	var got []string
	for _, batch := range fake.Batches() {
		for _, ev := range batch {
			got = append(got, ev.Name)
		}
	}
	if len(got) != len(want) {
		t.Fatalf("exporter received %d events %v, want %d", len(got), got, len(want))
	}
}

func TestQueue_CloseTwiceIsNoOp(t *testing.T) {
	fake := &telemetry.Fake{}
	q := telemetry.NewQueue(fake, 4)
	q.Emit(telemetry.Event{Name: "error"})

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	if err := q.Close(ctx); err != nil {
		t.Fatalf("first Close: %v", err)
	}
	if err := q.Close(ctx); err != nil {
		t.Fatalf("second Close: %v", err)
	}
	if fake.Calls() == 0 {
		t.Fatalf("exporter never received the event queued before Close")
	}
}

func TestQueue_EmitAfterCloseIsNoOp(t *testing.T) {
	fake := &telemetry.Fake{}
	q := telemetry.NewQueue(fake, 4)
	q.Emit(telemetry.Event{Name: "error"})

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	if err := q.Close(ctx); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if fake.Calls() == 0 {
		t.Fatalf("exporter never received the event queued before Close")
	}

	before := q.Dropped()
	done := make(chan struct{})
	go func() {
		q.Emit(telemetry.Event{Name: "error"})
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(50 * time.Millisecond):
		t.Fatalf("Emit after Close blocked")
	}
	if q.Dropped() != before {
		t.Fatalf("Emit after Close incremented Dropped: got %d, want %d", q.Dropped(), before)
	}
}
