package service

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/errcode"
)

func fakeProbe(comm string, present ...string) Probe {
	return Probe{
		PID1Comm: func() (string, error) { return comm, nil },
		Exists: func(path string) bool {
			for _, p := range present {
				if p == path {
					return true
				}
			}
			return false
		},
	}
}

func TestDetectByRunningInit(t *testing.T) {
	tests := []struct {
		name    string
		probe   Probe
		want    Manager
		wantErr errcode.Code
	}{
		{name: "systemd comm", probe: fakeProbe("systemd"), want: "systemd"},
		{name: "init with systemd run dir", probe: fakeProbe("init", "/run/systemd/system"), want: "systemd"},
		{name: "launchd comm", probe: fakeProbe("launchd"), want: "launchd"},
		{name: "init with openrc run dir", probe: fakeProbe("init", "/run/openrc"), want: "openrc"},
		{name: "busybox with nothing", probe: fakeProbe("busybox"), wantErr: errcode.UnsupportedServiceManager},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Detect(tt.probe)
			if tt.wantErr != "" {
				if code := errcode.Of(err); code != tt.wantErr {
					t.Fatalf("Detect(%s) error code = %q (err %v), want %q", tt.name, code, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Detect(%s) error = %v, want nil", tt.name, err)
			}
			if got != tt.want {
				t.Errorf("Detect(%s) = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

func TestDetectIgnoresDistributionName(t *testing.T) {
	var asked []string
	probe := Probe{
		PID1Comm: func() (string, error) { return "init", nil },
		Exists: func(path string) bool {
			asked = append(asked, path)
			return false
		},
	}

	_, err := Detect(probe)

	if code := errcode.Of(err); code != errcode.UnsupportedServiceManager {
		t.Errorf("Detect(init, nothing present) error code = %q (err %v), want %q", code, err, errcode.UnsupportedServiceManager)
	}
	for _, p := range asked {
		for _, bad := range []string{"os-release", "lsb-release", "/etc/"} {
			if strings.Contains(p, bad) {
				t.Errorf("Detect asked Exists(%q), want no path containing %q", p, bad)
			}
		}
	}
}

func TestForManagerUnsupportedWritesNothing(t *testing.T) {
	const config = "/home/u/.config/snapback/config.yaml"
	const wantHint = "snapback run --config " + config
	for _, m := range []Manager{"launchd", "openrc"} {
		t.Run(string(m), func(t *testing.T) {
			unitDir := t.TempDir()

			_, err := ForManager(m, unitDir, config)

			if !errors.Is(err, ErrUnsupportedManager) {
				t.Errorf("ForManager(%q) error = %v, want errors.Is ErrUnsupportedManager", m, err)
			}
			if code := errcode.Of(err); code != errcode.UnsupportedServiceManager {
				t.Errorf("ForManager(%q) error code = %q, want %q", m, code, errcode.UnsupportedServiceManager)
			}
			if err == nil || !strings.Contains(err.Error(), wantHint) {
				t.Errorf("ForManager(%q) error = %v, want message containing %q", m, err, wantHint)
			}
			entries, rerr := os.ReadDir(unitDir)
			if rerr != nil {
				t.Fatalf("os.ReadDir(%q) error = %v", unitDir, rerr)
			}
			if len(entries) != 0 {
				t.Errorf("ForManager(%q) left %d entries in unit dir, want 0", m, len(entries))
			}
		})
	}
}
