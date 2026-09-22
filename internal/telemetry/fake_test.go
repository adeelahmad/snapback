package telemetry

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
)

// Fake must satisfy the exporter seam. The interface is spelled out here rather
// than named so that this file does not depend on exporter.go.
var _ interface {
	Export(context.Context, []Event) error
} = (*Fake)(nil)

func ev(name string) Event {
	return Event{Name: name}
}

func TestFakeRecordsEveryBatchInOrder(t *testing.T) {
	var f Fake
	want := [][]Event{
		{ev("setup.completed")},
		{ev("daemon.started"), ev("mount.ready")},
		{ev("error")},
	}

	for _, batch := range want {
		if err := f.Export(context.Background(), batch); err != nil {
			t.Fatalf("Export(_, %v) = %v, want nil", batch, err)
		}
	}

	if got := f.Calls(); got != len(want) {
		t.Errorf("Calls() = %d, want %d", got, len(want))
	}
	if got := f.Batches(); !reflect.DeepEqual(got, want) {
		t.Errorf("Batches() = %v, want %v", got, want)
	}
}

func TestFakeNamesAreInReceiptOrder(t *testing.T) {
	var f Fake
	batches := [][]Event{
		{ev("setup.completed")},
		{ev("daemon.started"), ev("mount.ready")},
		{ev("error")},
	}
	for _, batch := range batches {
		if err := f.Export(context.Background(), batch); err != nil {
			t.Fatalf("Export(_, %v) = %v, want nil", batch, err)
		}
	}

	want := []string{"setup.completed", "daemon.started", "mount.ready", "error"}
	if got := f.Names(); !reflect.DeepEqual(got, want) {
		t.Errorf("Names() = %q, want %q", got, want)
	}
}

func TestFakeFailReturnsThePrimedErrorAndStillRecords(t *testing.T) {
	var f Fake
	wantErr := errors.New("exporter is down")
	f.Fail(wantErr)

	batch := []Event{ev("doctor.failed")}
	if err := f.Export(context.Background(), batch); !errors.Is(err, wantErr) {
		t.Errorf("Export(_, %v) = %v, want %v", batch, err, wantErr)
	}
	if got := f.Calls(); got != 1 {
		t.Errorf("Calls() = %d, want 1", got)
	}
	if got, want := f.Batches(), [][]Event{batch}; !reflect.DeepEqual(got, want) {
		t.Errorf("Batches() = %v, want %v", got, want)
	}
}

func TestFakeExportIsSafeForConcurrentUse(t *testing.T) {
	const (
		exporters      = 8
		perExporter    = 100
		wantTotalCalls = exporters * perExporter
	)

	var f Fake
	var wg sync.WaitGroup
	for i := 0; i < exporters; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < perExporter; j++ {
				if err := f.Export(context.Background(), []Event{ev("daemon.started")}); err != nil {
					t.Errorf("Export(_, [daemon.started]) = %v, want nil", err)
					return
				}
			}
		}()
	}
	wg.Wait()

	if got := f.Calls(); got != wantTotalCalls {
		t.Errorf("Calls() = %d, want %d", got, wantTotalCalls)
	}
	if got := len(f.Batches()); got != wantTotalCalls {
		t.Errorf("len(Batches()) = %d, want %d", got, wantTotalCalls)
	}
	if got := len(f.Names()); got != wantTotalCalls {
		t.Errorf("len(Names()) = %d, want %d", got, wantTotalCalls)
	}
}
