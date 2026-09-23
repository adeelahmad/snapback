package telemetry

import (
	"reflect"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/errcode"
)

// failureNow is a fixed instant, so the events under test carry no wall clock.
var failureNow = time.Date(2026, time.March, 4, 9, 30, 0, 0, time.UTC)

// wantDoctorChecks mirrors doctor.CheckNames(); events_failure_ext_test.go
// asserts the two packages agree, since telemetry cannot import doctor.
var wantDoctorChecks = []string{
	"config", "restic", "rclone", "fuse_device", "fusermount3", "password_file",
	"service_manager", "inode_headroom", "daemon_socket", "on_access",
}

func TestDoctorChecksIsTheClosedCheckList(t *testing.T) {
	if got := DoctorChecks(); !slices.Equal(got, wantDoctorChecks) {
		t.Fatalf("DoctorChecks() = %v, want %v", got, wantDoctorChecks)
	}
}

func TestDoctorChecksReturnsAFreshCopy(t *testing.T) {
	first := DoctorChecks()
	if len(first) == 0 {
		t.Fatal("DoctorChecks() is empty, want the registered check names")
	}
	first[0] = "mutated"
	if second := DoctorChecks(); second[0] == "mutated" {
		t.Fatal("DoctorChecks() shares backing storage across calls")
	}
}

func TestDoctorFailedAcceptsEveryRegisteredCheck(t *testing.T) {
	for _, check := range wantDoctorChecks {
		got, err := DoctorFailed("1.4.1", check, failureNow)
		if err != nil {
			t.Errorf("DoctorFailed(%q) error = %v, want nil", check, err)
			continue
		}
		if got.Name != "doctor.failed" {
			t.Errorf("DoctorFailed(%q).Name = %q, want %q", check, got.Name, "doctor.failed")
		}
		if !got.Time.Equal(failureNow) {
			t.Errorf("DoctorFailed(%q).Time = %v, want %v", check, got.Time, failureNow)
		}
		want := []Attr{
			{Key: "version", Value: "1.4.1"},
			{Key: "os", Value: runtime.GOOS},
			{Key: "arch", Value: runtime.GOARCH},
			{Key: "check", Value: check},
		}
		if !reflect.DeepEqual(got.Attrs, want) {
			t.Errorf("DoctorFailed(%q).Attrs = %v, want exactly %v", check, got.Attrs, want)
		}
	}
}

func TestDoctorFailedRejectsAnUnregisteredCheck(t *testing.T) {
	for _, check := range []string{"", "not_a_real_check", "repository:repoA"} {
		got, err := DoctorFailed("1.4.1", check, failureNow)
		if err == nil {
			t.Errorf("DoctorFailed(%q) = %+v, nil; want an error", check, got)
			continue
		}
		msg := err.Error()
		for _, want := range wantDoctorChecks {
			if !strings.Contains(msg, want) {
				t.Errorf("DoctorFailed(%q) error %q does not name allowed check %q", check, msg, want)
			}
		}
	}
}

// wantErrorCodes is pinned against every exported constant in
// internal/errcode; a new constant there must be added here too.
var wantErrorCodes = []errcode.Code{
	errcode.InvalidConfig,
	errcode.PrereqMissing,
	errcode.PermissionDenied,
	errcode.LinkConflict,
	errcode.RepoUnavailable,
	errcode.MappingAbsent,
	errcode.MountFailure,
	errcode.UnsupportedServiceManager,
	errcode.InodeBudgetExceeded,
	errcode.OnAccessUnavailable,
	errcode.StaleState,
}

func TestErrorCodesIsPinnedToEveryErrcodeConstant(t *testing.T) {
	if got := ErrorCodes(); !slices.Equal(got, wantErrorCodes) {
		t.Fatalf("ErrorCodes() = %v, want %v", got, wantErrorCodes)
	}
}

func TestErrorCodesReturnsAFreshCopy(t *testing.T) {
	first := ErrorCodes()
	if len(first) == 0 {
		t.Fatal("ErrorCodes() is empty, want every errcode constant")
	}
	first[0] = errcode.Code("mutated")
	if second := ErrorCodes(); second[0] == errcode.Code("mutated") {
		t.Fatal("ErrorCodes() shares backing storage across calls")
	}
}

func TestErrorEventAcceptsEveryKnownCode(t *testing.T) {
	for _, code := range wantErrorCodes {
		got, err := ErrorEvent("1.4.1", code, failureNow)
		if err != nil {
			t.Errorf("ErrorEvent(%q) error = %v, want nil", code, err)
			continue
		}
		if got.Name != "error" {
			t.Errorf("ErrorEvent(%q).Name = %q, want %q", code, got.Name, "error")
		}
		if !got.Time.Equal(failureNow) {
			t.Errorf("ErrorEvent(%q).Time = %v, want %v", code, got.Time, failureNow)
		}
		want := []Attr{
			{Key: "version", Value: "1.4.1"},
			{Key: "os", Value: runtime.GOOS},
			{Key: "arch", Value: runtime.GOARCH},
			{Key: "code", Value: string(code)},
		}
		if !reflect.DeepEqual(got.Attrs, want) {
			t.Errorf("ErrorEvent(%q).Attrs = %v, want exactly %v", code, got.Attrs, want)
		}
	}
}

func TestErrorEventRejectsEmptyAndUnknownCodes(t *testing.T) {
	for _, code := range []errcode.Code{"", "not_a_real_code"} {
		got, err := ErrorEvent("1.4.1", code, failureNow)
		if err == nil {
			t.Errorf("ErrorEvent(%q) = %+v, nil; want an error", code, got)
			continue
		}
		msg := err.Error()
		for _, want := range wantErrorCodes {
			if !strings.Contains(msg, string(want)) {
				t.Errorf("ErrorEvent(%q) error %q does not name allowed code %q", code, msg, want)
			}
		}
	}
}

// TestFailureEventSignaturesTakeNoMessage pins the parameter types: neither
// constructor can accept a message, an op, an argv or a wrapped error.
func TestFailureEventSignaturesTakeNoMessage(t *testing.T) {
	tests := []struct {
		name  string
		fn    any
		types []reflect.Type
	}{
		{
			name:  "DoctorFailed",
			fn:    DoctorFailed,
			types: []reflect.Type{reflect.TypeOf(""), reflect.TypeOf(""), reflect.TypeOf(time.Time{})},
		},
		{
			name:  "ErrorEvent",
			fn:    ErrorEvent,
			types: []reflect.Type{reflect.TypeOf(""), reflect.TypeOf(errcode.Code("")), reflect.TypeOf(time.Time{})},
		},
	}
	for _, tt := range tests {
		ft := reflect.TypeOf(tt.fn)
		if ft.NumIn() != len(tt.types) {
			t.Errorf("%s takes %d parameters, want exactly %d", tt.name, ft.NumIn(), len(tt.types))
			continue
		}
		for i, want := range tt.types {
			if got := ft.In(i); got != want {
				t.Errorf("%s parameter %d is %v, want %v", tt.name, i, got, want)
			}
		}
	}
}
