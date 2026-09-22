package main

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/adeelahmad/snapback/internal/cli"
	"github.com/adeelahmad/snapback/internal/service"
	"github.com/adeelahmad/snapback/internal/setup"
)

// The setup seams take these adapters, so the compiler pins their shapes.
var (
	_ setup.ServiceInstaller = setupInstaller{}
	_                        = cli.Deps{
		ServiceInstaller: setupInstaller{},
		ServiceSupported: func() bool { return serviceSupported(runtime.GOOS, service.RealProbe()) },
	}
)

// fakeProbe reads a named PID 1 and finds no /run marker, so Detect decides
// on the command name alone and no systemd has to be running.
func fakeProbe(comm string) service.Probe {
	return service.Probe{
		PID1Comm: func() (string, error) { return comm, nil },
		Exists:   func(string) bool { return false },
	}
}

func TestServiceSupportedOnlyOnLinuxSystemd(t *testing.T) {
	tests := []struct {
		name string
		goos string
		comm string
		want bool
	}{
		{"linux systemd", "linux", "systemd", true},
		{"linux without systemd", "linux", "openrc", false},
		{"darwin launchd", "darwin", "launchd", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := serviceSupported(tt.goos, fakeProbe(tt.comm)); got != tt.want {
				t.Errorf("serviceSupported(%q, %q) = %v, want %v", tt.goos, tt.comm, got, tt.want)
			}
		})
	}
}

func TestSetupUnitDirPrefersXDG(t *testing.T) {
	env := map[string]string{"XDG_CONFIG_HOME": "/cfg", "HOME": "/home/a"}
	getenv := func(k string) string { return env[k] }
	if got, want := setupUnitDir(getenv), filepath.Join("/cfg", "systemd", "user"); got != want {
		t.Errorf("setupUnitDir = %q, want %q", got, want)
	}
	delete(env, "XDG_CONFIG_HOME")
	if got, want := setupUnitDir(getenv), filepath.Join("/home/a", ".config", "systemd", "user"); got != want {
		t.Errorf("setupUnitDir without XDG = %q, want %q", got, want)
	}
}
