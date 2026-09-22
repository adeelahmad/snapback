package telemetry

import (
	"context"
	"reflect"
	"testing"
	"time"
)

// threeEvents builds a fixed batch large enough to prove the no-op exporter
// neither inspects nor retains what it is handed.
func threeEvents() []Event {
	base := time.Date(2026, time.September, 23, 12, 0, 0, 0, time.UTC)
	return []Event{
		{Name: "daemon.started", Time: base},
		{Name: "mount.ready", Time: base.Add(time.Second)},
		{Name: "doctor.failed", Time: base.Add(2 * time.Second)},
	}
}

func TestNopExporterReturnsNilForNilBatch(t *testing.T) {
	if err := (Nop{}).Export(context.Background(), nil); err != nil {
		t.Errorf("Nop.Export(ctx, nil) = %v, want nil", err)
	}
}

func TestNopExporterReturnsNilForThreeEvents(t *testing.T) {
	if err := (Nop{}).Export(context.Background(), threeEvents()); err != nil {
		t.Errorf("Nop.Export(ctx, 3 events) = %v, want nil", err)
	}
}

func TestNopExporterAllocatesNothingPerCall(t *testing.T) {
	ctx := context.Background()
	events := threeEvents()
	nop := Nop{}

	got := testing.AllocsPerRun(100, func() {
		_ = nop.Export(ctx, events)
	})
	if got != 0 {
		t.Errorf("testing.AllocsPerRun(Nop.Export) = %v, want 0", got)
	}
}

func TestExporterInterfaceHasExactlyOneMethod(t *testing.T) {
	typ := reflect.TypeOf((*Exporter)(nil)).Elem()

	if got := typ.NumMethod(); got != 1 {
		names := make([]string, 0, got)
		for i := range got {
			names = append(names, typ.Method(i).Name)
		}
		t.Fatalf("Exporter.NumMethod() = %d %v, want 1 [Export]", got, names)
	}

	m := typ.Method(0)
	if m.Name != "Export" {
		t.Fatalf("Exporter method 0 = %q, want %q", m.Name, "Export")
	}

	want := reflect.TypeOf(func(context.Context, []Event) error { return nil })
	if m.Type != want {
		t.Errorf("Export signature = %v, want %v", m.Type, want)
	}
}
