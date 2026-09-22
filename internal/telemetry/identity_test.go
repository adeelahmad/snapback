package telemetry

import (
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func TestIdentityHasOnlyTheBuildTriple(t *testing.T) {
	typ := reflect.TypeOf(Identity{})
	var got []string
	for i := range typ.NumField() {
		field := typ.Field(i)
		if field.Type.Kind() != reflect.String {
			t.Errorf("Identity.%s is %s, want string", field.Name, field.Type)
		}
		got = append(got, field.Name)
	}
	want := []string{"Version", "OS", "Arch"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Identity fields = %q, want exactly %q", got, want)
	}
}

func TestNewIdentityUsesTheRuntimeConstants(t *testing.T) {
	for _, version := range []string{"1.4.1", "0.0.0-dev", ""} {
		got := NewIdentity(version)
		want := Identity{Version: version, OS: runtime.GOOS, Arch: runtime.GOARCH}
		if got != want {
			t.Errorf("NewIdentity(%q) = %+v, want %+v", version, got, want)
		}
	}
}

func TestNewIdentityIgnoresTheEnvironment(t *testing.T) {
	const sentinel = "sentinel-must-not-appear"
	for _, key := range []string{"HOSTNAME", "HOST", "USER", "LOGNAME", "SNAPBACK_VERSION", "SNAPBACK_OS", "SNAPBACK_ARCH"} {
		t.Setenv(key, sentinel)
	}

	got := NewIdentity("1.4.1")
	for _, field := range []string{got.Version, got.OS, got.Arch} {
		if strings.Contains(field, sentinel) {
			t.Fatalf("NewIdentity read the environment: %+v", got)
		}
	}
	if want := (Identity{Version: "1.4.1", OS: runtime.GOOS, Arch: runtime.GOARCH}); got != want {
		t.Fatalf("NewIdentity(%q) = %+v, want %+v", "1.4.1", got, want)
	}
}

func TestIdentityAttrsAreTheOrderedTriple(t *testing.T) {
	got, err := NewIdentity("1.4.1").Attrs()
	if err != nil {
		t.Fatalf("Attrs() error = %v, want nil", err)
	}
	want := []Attr{
		{Key: "version", Value: "1.4.1"},
		{Key: "os", Value: runtime.GOOS},
		{Key: "arch", Value: runtime.GOARCH},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Attrs() = %+v, want %+v", got, want)
	}
}

func TestIdentityAttrsRejectAnUnsafeVersion(t *testing.T) {
	got, err := NewIdentity("1.4.1 dev").Attrs()
	if err == nil {
		t.Fatalf("Attrs() = %+v, want an error for a version NewAttr rejects", got)
	}
	if got != nil {
		t.Errorf("Attrs() = %+v on error, want nil", got)
	}
	if !strings.Contains(err.Error(), "version") {
		t.Errorf("Attrs() error = %q, want it to name the version key", err)
	}
}
