package telemetry_test

import (
	"context"
	"runtime"
	"slices"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/telemetry"
)

// testEvent is a valid schema event the client tests emit.
func testEvent() telemetry.Event {
	return telemetry.Event{Name: "daemon.started", Time: time.Unix(0, 0).UTC()}
}

func TestClientDisabledEmitsNothing(t *testing.T) {
	fake := &telemetry.Fake{}
	ctx := context.Background()

	before := runtime.NumGoroutine()
	client := telemetry.New(telemetry.Options{Enabled: false, Endpoint: "http://collector.invalid", Exporter: fake})
	client.Emit(ctx, testEvent())
	err := client.Close(ctx)
	after := runtime.NumGoroutine()

	if client.Enabled() {
		t.Errorf("Enabled() = true, want false for Options{Enabled: false}")
	}
	if got := fake.Calls(); got != 0 {
		t.Errorf("fake.Calls() = %d, want 0 for a disabled client", got)
	}
	if err != nil {
		t.Errorf("Close() = %v, want nil", err)
	}
	if after != before {
		t.Errorf("runtime.NumGoroutine() = %d after New+Emit+Close, want %d", after, before)
	}
}

func TestClientEnabledWithoutEndpointEmitsNothing(t *testing.T) {
	fake := &telemetry.Fake{}
	ctx := context.Background()

	before := runtime.NumGoroutine()
	client := telemetry.New(telemetry.Options{Enabled: true, Endpoint: "", Exporter: fake})
	client.Emit(ctx, testEvent())
	err := client.Close(ctx)
	after := runtime.NumGoroutine()

	if client.Enabled() {
		t.Errorf("Enabled() = true, want false for an empty Endpoint")
	}
	if got := fake.Calls(); got != 0 {
		t.Errorf("fake.Calls() = %d, want 0 for an empty Endpoint", got)
	}
	if err != nil {
		t.Errorf("Close() = %v, want nil", err)
	}
	if after != before {
		t.Errorf("runtime.NumGoroutine() = %d after New+Emit+Close, want %d", after, before)
	}
}

func TestClientEnabledWithoutExporterEmitsNothing(t *testing.T) {
	ctx := context.Background()

	client := telemetry.New(telemetry.Options{Enabled: true, Endpoint: "http://collector.invalid", Exporter: nil})
	client.Emit(ctx, testEvent())

	if client.Enabled() {
		t.Errorf("Enabled() = true, want false for a nil Exporter")
	}
	if err := client.Close(ctx); err != nil {
		t.Errorf("Close() = %v, want nil", err)
	}
}

func TestClientEnabledExportsOneEventBatch(t *testing.T) {
	fake := &telemetry.Fake{}
	ctx := context.Background()

	client := telemetry.New(telemetry.Options{Enabled: true, Endpoint: "http://collector.invalid", Exporter: fake})
	t.Cleanup(func() { _ = client.Close(ctx) })

	if !client.Enabled() {
		t.Errorf("Enabled() = false, want true when Enabled, Endpoint and Exporter are all set")
	}

	ev := testEvent()
	client.Emit(ctx, ev)

	if got := fake.Calls(); got != 1 {
		t.Errorf("fake.Calls() = %d, want 1 after one Emit", got)
	}
	want := []string{ev.Name}
	if got := fake.Names(); !slices.Equal(got, want) {
		t.Errorf("fake.Names() = %v, want %v", got, want)
	}
	batches := fake.Batches()
	if len(batches) != 1 || len(batches[0]) != 1 {
		t.Fatalf("fake.Batches() = %v, want one batch of one event", batches)
	}
}

func TestClientEmitAfterCloseIsDropped(t *testing.T) {
	fake := &telemetry.Fake{}
	ctx := context.Background()

	client := telemetry.New(telemetry.Options{Enabled: true, Endpoint: "http://collector.invalid", Exporter: fake})
	client.Emit(ctx, testEvent())
	if err := client.Close(ctx); err != nil {
		t.Fatalf("Close() = %v, want nil", err)
	}

	client.Emit(ctx, testEvent())

	if got := fake.Calls(); got != 1 {
		t.Errorf("fake.Calls() = %d, want 1: Emit after Close must be dropped", got)
	}
}

func TestClientCloseTwiceReturnsNil(t *testing.T) {
	fake := &telemetry.Fake{}
	ctx := context.Background()

	client := telemetry.New(telemetry.Options{Enabled: true, Endpoint: "http://collector.invalid", Exporter: fake})

	if err := client.Close(ctx); err != nil {
		t.Errorf("first Close() = %v, want nil", err)
	}
	if err := client.Close(ctx); err != nil {
		t.Errorf("second Close() = %v, want nil", err)
	}
}
