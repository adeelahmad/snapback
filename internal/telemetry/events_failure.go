package telemetry

import (
	"time"

	"github.com/adeelahmad/snapback/internal/errcode"
)

// DoctorChecks is the closed list of doctor check names DoctorFailed accepts.
// It is a fixed copy of the names doctor.CheckNames() reports; telemetry
// cannot import internal/doctor, which will later import telemetry, so the
// list is pinned here instead and an external test asserts the two agree.
//
// SUB-AGENT-TODO: return the real, fixed check names in doctor's registration
// order, as a fresh copy on every call.
func DoctorChecks() []string { return nil }

// DoctorFailed returns the doctor.failed event for check at now. check must be
// one of [DoctorChecks]; any other value is an error naming the allowed
// checks.
//
// SUB-AGENT-TODO: validate check against DoctorChecks and build the
// version/os/arch/check attrs via NewAttr.
func DoctorFailed(version, check string, now time.Time) (Event, error) {
	return Event{}, nil
}

// ErrorCodes is the closed list of errcode.Code values ErrorEvent accepts,
// pinned against every exported constant in internal/errcode.
//
// SUB-AGENT-TODO: return the real, fixed code list as a fresh copy on every
// call.
func ErrorCodes() []errcode.Code { return nil }

// ErrorEvent returns the error event for code at now. code must be one of
// [ErrorCodes]; "" and any other value are errors.
//
// SUB-AGENT-TODO: validate code against ErrorCodes and build the
// version/os/arch/code attrs via NewAttr.
func ErrorEvent(version string, code errcode.Code, now time.Time) (Event, error) {
	return Event{}, nil
}
