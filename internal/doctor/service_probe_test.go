package doctor

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/service"
)

// TestDoctorServiceManagerMatchesService checks that doctor's production
// service-manager detection agrees with the service package's on this host.
func TestDoctorServiceManagerMatchesService(t *testing.T) {
	got, gotErr := realProbes().Detect()
	want, wantErr := service.Detect(service.RealProbe())
	if got != want || (gotErr == nil) != (wantErr == nil) {
		t.Errorf("realProbes().Detect() on %s = (%q, %v), want (%q, %v) as service.Detect(service.RealProbe())",
			runtime.GOOS, got, gotErr, want, wantErr)
	}
}

// TestDoctorReusesServiceProbe fails while internal/doctor wires its own
// PID 1 probe instead of calling service.RealProbe.
func TestDoctorReusesServiceProbe(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("filepath.Glob(%q) = %v", "*.go", err)
	}
	var callsRealProbe bool
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("os.ReadFile(%q) = %v", f, err)
		}
		src := string(b)
		for _, dup := range []string{"/proc/1/comm", "service.Probe{"} {
			if strings.Contains(src, dup) {
				t.Errorf("%s contains %q, want doctor to reuse service.RealProbe()", f, dup)
			}
		}
		if strings.Contains(src, "service.RealProbe()") {
			callsRealProbe = true
		}
	}
	if !callsRealProbe {
		t.Errorf("internal/doctor never calls service.RealProbe(), want its Detect probe to use it")
	}
}
