package setup

import (
	"context"
	"errors"
	"testing"
)

// countingInstaller records every Install call instead of touching the host,
// so a test never starts a real service.
type countingInstaller struct {
	calls  int
	exe    string
	config string
	err    error
}

func (c *countingInstaller) Install(_ context.Context, exe, configPath string) error {
	c.calls++
	c.exe, c.config = exe, configPath
	return c.err
}

func TestInstallService(t *testing.T) {
	errInstall := errors.New("unit write failed")

	tests := []struct {
		name       string
		goos       string
		supported  bool
		noService  bool
		installErr error
		wantCalls  int
		want       Outcome
		wantErr    error
	}{
		{
			name:      "--no-service skips without calling the installer",
			goos:      "linux",
			supported: true,
			noService: true,
			wantCalls: 0,
			want:      Outcome{Reason: "skipped (--no-service)"},
		},
		{
			name:      "darwin has no login service yet",
			goos:      "darwin",
			supported: true,
			wantCalls: 0,
			want:      Outcome{Reason: "login service is Linux-only for now"},
		},
		{
			name:      "linux without a manager names the reason",
			goos:      "linux",
			supported: false,
			wantCalls: 0,
			want:      Outcome{Reason: "no supported service manager detected"},
		},
		{
			name:      "linux with a manager installs once",
			goos:      "linux",
			supported: true,
			wantCalls: 1,
			want:      Outcome{Installed: true, Removal: "snapback service uninstall"},
		},
		{
			name:       "a failing install is reported",
			goos:       "linux",
			supported:  true,
			installErr: errInstall,
			wantCalls:  1,
			wantErr:    errInstall,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inst := &countingInstaller{err: tt.installErr}
			got, err := InstallService(context.Background(), tt.goos, tt.supported, tt.noService, inst, "/usr/bin/snapback", "/etc/snapback.toml")

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("InstallService(%q, %t, %t) error = %v, want %v", tt.goos, tt.supported, tt.noService, err, tt.wantErr)
			}
			if tt.wantErr == nil && got != tt.want {
				t.Errorf("InstallService(%q, %t, %t) = %+v, want %+v", tt.goos, tt.supported, tt.noService, got, tt.want)
			}
			if inst.calls != tt.wantCalls {
				t.Errorf("InstallService(%q, %t, %t) made %d Install calls, want %d", tt.goos, tt.supported, tt.noService, inst.calls, tt.wantCalls)
			}
			if tt.wantCalls == 1 && (inst.exe != "/usr/bin/snapback" || inst.config != "/etc/snapback.toml") {
				t.Errorf("Install(exe, config) = (%q, %q), want (%q, %q)", inst.exe, inst.config, "/usr/bin/snapback", "/etc/snapback.toml")
			}
		})
	}
}
