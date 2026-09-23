package mount

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/telemetry"
)

// fakeMountAdapter is a minimal Adapter whose Mount outcome is fixed by err.
type fakeMountAdapter struct {
	err         error
	mountCalls  int
	unmountCall int
}

func (f *fakeMountAdapter) Mount(dir string, cat Catalog) error {
	f.mountCalls++
	return f.err
}

func (f *fakeMountAdapter) Unmount() error {
	f.unmountCall++
	return nil
}

// sequenceClock returns each time in times in order, then repeats the last
// one, so a test can control the elapsed duration a caller observes.
func sequenceClock(times ...time.Time) func() time.Time {
	i := 0
	return func() time.Time {
		t := times[i]
		if i < len(times)-1 {
			i++
		}
		return t
	}
}

func newTestClient(exp telemetry.Exporter) *telemetry.Client {
	return telemetry.New(telemetry.Options{Enabled: true, Endpoint: "http://collector.example", Exporter: exp})
}

func TestTelemetryAdapterEmitsMountReadyOnFirstSuccess(t *testing.T) {
	var fake telemetry.Fake
	start := time.Date(2026, time.March, 4, 9, 30, 0, 0, time.UTC)
	ready := start.Add(2 * time.Second)
	a := &TelemetryAdapter{
		Adapter: &fakeMountAdapter{},
		Client:  newTestClient(&fake),
		Version: "1.4.1",
		Now:     sequenceClock(start, ready),
	}

	if err := a.Mount("/mnt/snapshot", newFakeCatalog()); err != nil {
		t.Fatalf("Mount() error = %v, want nil", err)
	}

	names := fake.Names()
	if len(names) != 1 || names[0] != "mount.ready" {
		t.Fatalf("Fake.Names() = %v, want exactly one %q", names, "mount.ready")
	}
	batches := fake.Batches()
	ev := batches[0][0]
	want := []telemetry.Attr{
		{Key: "version", Value: "1.4.1"},
		{Key: "os", Value: attrOS(t)},
		{Key: "arch", Value: attrArch(t)},
		{Key: "duration", Value: telemetry.Bucket(2 * time.Second)},
	}
	if len(ev.Attrs) != len(want) {
		t.Fatalf("Attrs = %v, want %v", ev.Attrs, want)
	}
	for i, wa := range want {
		if ev.Attrs[i] != wa {
			t.Errorf("Attrs[%d] = %v, want %v", i, ev.Attrs[i], wa)
		}
	}
}

func TestTelemetryAdapterEmitsMountReadyOnlyOnceForFirstSuccess(t *testing.T) {
	var fake telemetry.Fake
	now := time.Date(2026, time.March, 4, 9, 30, 0, 0, time.UTC)
	a := &TelemetryAdapter{
		Adapter: &fakeMountAdapter{},
		Client:  newTestClient(&fake),
		Version: "1.4.1",
		Now:     func() time.Time { return now },
	}

	if err := a.Mount("/mnt/one", newFakeCatalog()); err != nil {
		t.Fatalf("first Mount() error = %v, want nil", err)
	}
	if err := a.Mount("/mnt/two", newFakeCatalog()); err != nil {
		t.Fatalf("second Mount() error = %v, want nil", err)
	}

	if got := fake.Calls(); got != 1 {
		t.Fatalf("Fake.Calls() = %d, want exactly 1 (only the first successful mount reports)", got)
	}
}

func TestTelemetryAdapterEmitsErrorAndNoMountReadyOnFailure(t *testing.T) {
	var fake telemetry.Fake
	now := time.Date(2026, time.March, 4, 9, 30, 0, 0, time.UTC)
	a := &TelemetryAdapter{
		Adapter: &fakeMountAdapter{err: errors.New("mount /mnt/snapshot: permission denied")},
		Client:  newTestClient(&fake),
		Version: "1.4.1",
		Now:     func() time.Time { return now },
	}

	err := a.Mount("/mnt/snapshot", newFakeCatalog())
	if err == nil {
		t.Fatal("Mount() error = nil, want the wrapped adapter's error")
	}

	names := fake.Names()
	if len(names) != 1 || names[0] != "error" {
		t.Fatalf("Fake.Names() = %v, want exactly one %q and no %q", names, "error", "mount.ready")
	}
	ev := fake.Batches()[0][0]
	var gotCode string
	for _, attr := range ev.Attrs {
		if attr.Key == "code" {
			gotCode = attr.Value
		}
	}
	if gotCode != string(errcode.MountFailure) {
		t.Errorf("error event code = %q, want %q", gotCode, errcode.MountFailure)
	}
}

func TestTelemetryAdapterWithNilClientMountsWithoutPanicking(t *testing.T) {
	a := &TelemetryAdapter{Adapter: &fakeMountAdapter{}, Version: "1.4.1"}
	if err := a.Mount("/mnt/snapshot", newFakeCatalog()); err != nil {
		t.Fatalf("Mount() error = %v, want nil", err)
	}

	failing := &TelemetryAdapter{Adapter: &fakeMountAdapter{err: errors.New("boom")}, Version: "1.4.1"}
	if err := failing.Mount("/mnt/snapshot", newFakeCatalog()); err == nil {
		t.Fatal("Mount() error = nil, want the wrapped adapter's error")
	}
}

// TestTelemetryAdapterEventsCarryNoMountIdentifiers proves the events this
// call site emits, over both outcomes, carry none of the identifiers a mount
// attempt naturally has in hand: the mount point, a repository id or the
// live root path. It drives the adapter with a realistic-looking mount point,
// repo id and root path and scans every attribute the Fake exporter recorded,
// mirroring internal/telemetry's noleak_events_test.go technique.
func TestTelemetryAdapterEventsCarryNoMountIdentifiers(t *testing.T) {
	const (
		mountPoint = "/home/jdoe/project/.snapshot"
		repoID     = "s3:backups.example.com/jdoe-project"
		rootPath   = "/home/jdoe/project"
	)
	var fake telemetry.Fake
	now := time.Date(2026, time.March, 4, 9, 30, 0, 0, time.UTC)

	ok := &TelemetryAdapter{Adapter: &fakeMountAdapter{}, Client: newTestClient(&fake), Version: "1.4.1", Now: func() time.Time { return now }}
	if err := ok.Mount(mountPoint, newFakeCatalog()); err != nil {
		t.Fatalf("Mount() error = %v, want nil", err)
	}

	failing := &TelemetryAdapter{Adapter: &fakeMountAdapter{err: errors.New("mount " + rootPath + " via " + repoID + ": denied")}, Client: newTestClient(&fake), Version: "1.4.1", Now: func() time.Time { return now }}
	if err := failing.Mount(mountPoint, newFakeCatalog()); err == nil {
		t.Fatal("Mount() error = nil, want the wrapped adapter's error")
	}

	for _, batch := range fake.Batches() {
		for _, ev := range batch {
			b, err := json.Marshal(ev.Attrs)
			if err != nil {
				t.Fatalf("json.Marshal(%+v) = %v, want no error", ev.Attrs, err)
			}
			for _, needle := range []string{mountPoint, repoID, rootPath} {
				if strings.Contains(string(b), needle) {
					t.Errorf("event %q attrs %s contain %q", ev.Name, b, needle)
				}
			}
			if findings := telemetry.ScanProhibited(b); len(findings) != 0 {
				t.Errorf("ScanProhibited(%s) = %+v, want zero findings", b, findings)
			}
		}
	}
}

// TestMountEventConstructorSignaturesTakeNoIdentifier mirrors
// internal/telemetry's TestRuntimeEventSignaturesTakeNoIdentifier: it pins
// the parameter types of the two event constructors this call site uses, so
// nothing at the call site could pass a mount point, a repo id or a root
// path through a spare parameter slot -- the constructors have none.
func TestMountEventConstructorSignaturesTakeNoIdentifier(t *testing.T) {
	requireOneStringParam(t, "telemetry.MountReady", telemetry.MountReady)
	requireOneStringParam(t, "telemetry.ErrorEvent", telemetry.ErrorEvent)
}

// requireOneStringParam fails the test unless fn's function type has exactly
// one parameter of the plain string type: enough room for a version, and (by
// exact type, not kind, so a distinct enum type such as errcode.Code does not
// count) never enough for a second identifier alongside it.
func requireOneStringParam(t *testing.T, name string, fn any) {
	t.Helper()
	ft := reflect.TypeOf(fn)
	stringType := reflect.TypeOf("")
	stringParams := 0
	for i := range ft.NumIn() {
		if ft.In(i) == stringType {
			stringParams++
		}
	}
	if stringParams != 1 {
		t.Errorf("%s takes %d plain-string parameters, want exactly 1 (the version)", name, stringParams)
	}
}

func attrOS(t *testing.T) string {
	t.Helper()
	ev, err := telemetry.DaemonStarted("1.4.1", time.Now())
	if err != nil {
		t.Fatalf("telemetry.DaemonStarted(...) error = %v, want nil", err)
	}
	for _, a := range ev.Attrs {
		if a.Key == "os" {
			return a.Value
		}
	}
	t.Fatal("DaemonStarted(...) attrs have no os key")
	return ""
}

func attrArch(t *testing.T) string {
	t.Helper()
	ev, err := telemetry.DaemonStarted("1.4.1", time.Now())
	if err != nil {
		t.Fatalf("telemetry.DaemonStarted(...) error = %v, want nil", err)
	}
	for _, a := range ev.Attrs {
		if a.Key == "arch" {
			return a.Value
		}
	}
	t.Fatal("DaemonStarted(...) attrs have no arch key")
	return ""
}
