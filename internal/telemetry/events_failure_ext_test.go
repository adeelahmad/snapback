// Package telemetry_test holds the one telemetry test that is allowed to
// import internal/doctor: doctor will later import telemetry, so telemetry's
// own package (and its internal tests) must not import doctor, but an
// external test binary may import both to pin them against each other.
package telemetry_test

import (
	"slices"
	"testing"

	"github.com/adeelahmad/snapback/internal/doctor"
	"github.com/adeelahmad/snapback/internal/telemetry"
)

func TestDoctorChecksMatchesDoctorPackage(t *testing.T) {
	got := telemetry.DoctorChecks()
	want := doctor.CheckNames()
	if !slices.Equal(got, want) {
		t.Fatalf("telemetry.DoctorChecks() = %v, want doctor.CheckNames() = %v", got, want)
	}
}
