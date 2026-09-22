package provider

import (
	"context"
	"errors"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"
)

var (
	id1 = SnapshotID(strings.Repeat("a", 64))
	id2 = strings.Repeat("b", 64)
)

func TestSnapshotIDValid(t *testing.T) {
	tests := []struct {
		name string
		id   SnapshotID
		want bool
	}{
		{"64 lowercase hex", id1, true},
		{"64 uppercase hex", SnapshotID(strings.Repeat("A", 64)), false},
		{"63 hex", SnapshotID(strings.Repeat("a", 63)), false},
		{"65 hex", SnapshotID(strings.Repeat("a", 65)), false},
		{"8 hex short id", "aaaaaaaa", false},
		{"empty", "", false},
		{"64 chars with g", SnapshotID(strings.Repeat("a", 63) + "g"), false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.id.Valid(); got != tc.want {
				t.Errorf("SnapshotID(%q).Valid() = %v, want %v", tc.id, got, tc.want)
			}
		})
	}
}

func TestProbeResultValuesDistinctAndNonZero(t *testing.T) {
	values := []ProbeResult{ProbeDir, ProbeNotDir, ProbeAbsent}
	for i, v := range values {
		if v == ProbeResult(0) {
			t.Errorf("ProbeResult value %d = 0, want non-zero (zero means unknown)", i)
		}
		for j := i + 1; j < len(values); j++ {
			if v == values[j] {
				t.Errorf("ProbeResult values %d and %d both = %d, want distinct", i, j, v)
			}
		}
	}

	tests := []struct {
		r    ProbeResult
		want string
	}{
		{ProbeDir, "dir"},
		{ProbeNotDir, "not_dir"},
		{ProbeAbsent, "absent"},
		{ProbeResult(0), "unknown"},
	}
	for _, tc := range tests {
		if got := tc.r.String(); got != tc.want {
			t.Errorf("ProbeResult(%d).String() = %q, want %q", uint8(tc.r), got, tc.want)
		}
	}
}

type stubProvider struct{}

func (stubProvider) Validate(context.Context) (Identity, error) { return Identity{}, nil }

func (stubProvider) List(context.Context) ([]Snapshot, error) { return nil, nil }

func (stubProvider) StartMount(context.Context, string) (MountHandle, error) {
	return nil, errors.New("stub")
}

func (stubProvider) SnapshotRoot(string, SnapshotID) string { return "" }

func (stubProvider) Probe(context.Context, string, SnapshotID, string) (ProbeResult, error) {
	return 0, nil
}

func (stubProvider) Snap(context.Context, SnapRequest) (SnapshotID, error) { return "", nil }

func (stubProvider) Prewarm(context.Context, []SnapshotID, int) []PrewarmResult { return nil }

func TestSnapshotProviderComposesNarrowInterfaces(t *testing.T) {
	var s stubProvider
	var (
		_ SnapshotProvider = s
		_ Validator        = s
		_ Lister           = s
		_ Mounter          = s
		_ Prober           = s
		_ Snapper          = s
		_ Prewarmer        = s
	)

	typ := reflect.TypeFor[SnapshotProvider]()
	if got, want := typ.NumMethod(), 7; got != want {
		t.Errorf("SnapshotProvider NumMethod() = %d, want %d", got, want)
	}
	var got []string
	for i := range typ.NumMethod() {
		got = append(got, typ.Method(i).Name)
	}
	slices.Sort(got)
	want := []string{"List", "Prewarm", "Probe", "Snap", "SnapshotRoot", "StartMount", "Validate"}
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Errorf("SnapshotProvider methods = %v, want %v", got, want)
	}
}

// fieldNames returns the declared field names of struct value v in order.
func fieldNames(v any) []string {
	typ := reflect.TypeOf(v)
	names := make([]string, 0, typ.NumField())
	for i := range typ.NumField() {
		names = append(names, typ.Field(i).Name)
	}
	return names
}

func TestSnapshotFieldsMatchContract(t *testing.T) {
	ts := time.Date(2026, 9, 22, 5, 0, 0, 0, time.UTC)
	errWarm := errors.New("warm failed")

	snap := Snapshot{ID: id1, Time: ts, Hostname: "h", Tags: []string{"x"}, Paths: []string{"rel/p"}}
	if snap.ID != id1 || !snap.Time.Equal(ts) || snap.Hostname != "h" ||
		!slices.Equal(snap.Tags, []string{"x"}) || !slices.Equal(snap.Paths, []string{"rel/p"}) {
		t.Errorf("Snapshot fields read back = %+v, want ID %q Time %v Hostname h Tags [x] Paths [rel/p]", snap, id1, ts)
	}

	ident := Identity{RepoID: id2, Version: 2}
	if ident.RepoID != id2 || ident.Version != 2 {
		t.Errorf("Identity fields read back = %+v, want RepoID %q Version 2", ident, id2)
	}

	req := SnapRequest{Path: "/p", Host: "h", Tags: []string{"t"}, Excludes: []string{".snapshot"}}
	if req.Path != "/p" || req.Host != "h" ||
		!slices.Equal(req.Tags, []string{"t"}) || !slices.Equal(req.Excludes, []string{".snapshot"}) {
		t.Errorf("SnapRequest fields read back = %+v, want Path /p Host h Tags [t] Excludes [.snapshot]", req)
	}

	res := PrewarmResult{ID: id1, Warm: true, Err: errWarm}
	if res.ID != id1 || !res.Warm || !errors.Is(res.Err, errWarm) {
		t.Errorf("PrewarmResult fields read back = %+v, want ID %q Warm true Err %v", res, id1, errWarm)
	}

	// The binding contract fixes each struct's exact field set; extra or
	// renamed fields would break S3-03 and S3-04, which build these literals.
	fields := []struct {
		name string
		v    any
		want []string
	}{
		{"Snapshot", Snapshot{}, []string{"ID", "Time", "Hostname", "Tags", "Paths"}},
		{"Identity", Identity{}, []string{"RepoID", "Version"}},
		{"SnapRequest", SnapRequest{}, []string{"Path", "Host", "Tags", "Excludes"}},
		{"PrewarmResult", PrewarmResult{}, []string{"ID", "Warm", "Err"}},
	}
	for _, f := range fields {
		if got := fieldNames(f.v); !slices.Equal(got, f.want) {
			t.Errorf("%s fields = %v, want %v", f.name, got, f.want)
		}
	}
}
