package service

import (
	"context"
	"errors"
	"fmt"
	"os/exec"

	"github.com/adeelahmad/snapback/internal/errcode"
)

// Manager names a service manager: "systemd", "launchd" or "openrc".
type Manager string

// Probe reads the running init system.
type Probe struct {
	PID1Comm func() (string, error)
	Exists   func(path string) bool
}

// Installer installs and controls the Snapback service.
type Installer interface {
	Install(ctx context.Context, o UnitOptions) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Status(ctx context.Context) (string, error)
	Uninstall(ctx context.Context) error
}

// ErrUnsupportedManager reports a service manager Snapback cannot install for.
var ErrUnsupportedManager = errors.New("service: unsupported service manager")

// Detect returns the running service manager.
func Detect(probe Probe) (Manager, error) {
	comm, err := probe.PID1Comm()
	if err != nil {
		return "", errcode.New(errcode.UnsupportedServiceManager, "service detect", err)
	}
	switch {
	case comm == "systemd" || probe.Exists("/run/systemd/system"):
		return "systemd", nil
	case comm == "launchd":
		return "launchd", nil
	case probe.Exists("/run/openrc"):
		return "openrc", nil
	}
	return "", errcode.New(errcode.UnsupportedServiceManager, "service detect", fmt.Errorf("%w: pid 1 is %q", ErrUnsupportedManager, comm))
}

// ForManager returns the Installer for m.
func ForManager(m Manager, unitDir, config string) (Installer, error) {
	if m == "systemd" {
		return &Systemd{UnitDir: unitDir, Run: execRunner}, nil
	}
	return nil, errcode.New(errcode.UnsupportedServiceManager, "service install",
		fmt.Errorf("%w %q; run in the foreground instead: snapback run --config %s", ErrUnsupportedManager, m, config))
}

// execRunner runs name with args directly (no shell) and returns its stdout.
func execRunner(ctx context.Context, name string, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, name, args...).Output()
}
