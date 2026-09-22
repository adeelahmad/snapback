package main

import (
	"context"
	"os"
	"path/filepath"

	"github.com/adeelahmad/snapback/internal/service"
)

// setupInstaller installs the login service the way "install service" does —
// detect the manager, then write the user unit — behind the narrow seam
// setup turns the service on through.
type setupInstaller struct{}

func (setupInstaller) Install(ctx context.Context, exe, configPath string) error {
	m, err := service.Detect(service.RealProbe())
	if err != nil {
		return err
	}
	inst, err := service.ForManager(m, setupUnitDir(os.Getenv), configPath)
	if err != nil {
		return err
	}
	return inst.Install(ctx, service.UnitOptions{Exe: exe, Config: configPath, Scope: "user"})
}

// serviceSupported reports whether this host can hold the login service:
// Linux running systemd. Everything else — macOS launchd included — leaves
// setup to name the reason and skip the install.
func serviceSupported(goos string, probe service.Probe) bool {
	if goos != "linux" {
		return false
	}
	m, err := service.Detect(probe)
	return err == nil && m == "systemd"
}

// setupUnitDir returns $XDG_CONFIG_HOME/systemd/user, or ~/.config/systemd/user:
// the same user unit directory "install service" writes into.
func setupUnitDir(getenv func(string) string) string {
	if xdg := getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "systemd", "user")
	}
	return filepath.Join(getenv("HOME"), ".config", "systemd", "user")
}
