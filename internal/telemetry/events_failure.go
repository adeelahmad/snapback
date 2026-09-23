package telemetry

import (
	"fmt"
	"runtime"
	"slices"
	"strings"
	"time"

	"github.com/adeelahmad/snapback/internal/errcode"
)

// doctorChecks is the closed, ordered set of doctor check names DoctorFailed
// accepts. It mirrors doctor.CheckNames(); telemetry cannot import
// internal/doctor, which will later import telemetry, so the list is pinned
// here instead and an external test asserts the two agree.
var doctorChecks = []string{
	"config",
	"restic",
	"rclone",
	"fuse_device",
	"fusermount3",
	"password_file",
	"service_manager",
	"inode_headroom",
	"daemon_socket",
	"on_access",
}

// DoctorChecks returns the closed list of doctor check names DoctorFailed
// accepts, in doctor's registration order. The result is a fresh copy, so a
// caller cannot mutate the backing set.
func DoctorChecks() []string { return slices.Clone(doctorChecks) }

// DoctorFailed returns the doctor.failed event for check at now. check must be
// one of [DoctorChecks]; any other value is an error naming the allowed
// checks.
func DoctorFailed(version, check string, now time.Time) (Event, error) {
	if !slices.Contains(doctorChecks, check) {
		return Event{}, fmt.Errorf("telemetry: doctor check %q is not one of %s", check, strings.Join(doctorChecks, ", "))
	}
	versionAttr, err := NewVersionAttr(version)
	if err != nil {
		return Event{}, err
	}
	pairs := [...][2]string{
		{"os", runtime.GOOS},
		{"arch", runtime.GOARCH},
		{"check", check},
	}
	attrs, err := newAttrs(pairs[:])
	if err != nil {
		return Event{}, err
	}
	attrs = append([]Attr{versionAttr}, attrs...)
	return Event{Name: "doctor.failed", Attrs: attrs, Time: now}, nil
}

// errorCodes is the closed, ordered set of errcode.Code values ErrorEvent
// accepts, pinned against every exported constant in internal/errcode.
var errorCodes = []errcode.Code{
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

// ErrorCodes returns the closed list of errcode.Code values ErrorEvent
// accepts. The result is a fresh copy, so a caller cannot mutate the backing
// set.
func ErrorCodes() []errcode.Code { return slices.Clone(errorCodes) }

// ErrorEvent returns the error event for code at now. code must be one of
// [ErrorCodes]; "" and any other value are errors.
func ErrorEvent(version string, code errcode.Code, now time.Time) (Event, error) {
	if !slices.Contains(errorCodes, code) {
		names := make([]string, len(errorCodes))
		for i, c := range errorCodes {
			names[i] = string(c)
		}
		return Event{}, fmt.Errorf("telemetry: error code %q is not one of %s", code, strings.Join(names, ", "))
	}
	versionAttr, err := NewVersionAttr(version)
	if err != nil {
		return Event{}, err
	}
	pairs := [...][2]string{
		{"os", runtime.GOOS},
		{"arch", runtime.GOARCH},
		{"code", string(code)},
	}
	attrs, err := newAttrs(pairs[:])
	if err != nil {
		return Event{}, err
	}
	attrs = append([]Attr{versionAttr}, attrs...)
	return Event{Name: "error", Attrs: attrs, Time: now}, nil
}

// newAttrs builds an Attr for each key/value pair via NewAttr, in order.
func newAttrs(pairs [][2]string) ([]Attr, error) {
	attrs := make([]Attr, 0, len(pairs))
	for _, p := range pairs {
		a, err := NewAttr(p[0], p[1])
		if err != nil {
			return nil, err
		}
		attrs = append(attrs, a)
	}
	return attrs, nil
}
