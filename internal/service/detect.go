package service

import (
	"context"
	"errors"
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
	panic("SUB-AGENT-TODO: T1 read probe.PID1Comm and probe.Exists markers to pick systemd/launchd/openrc; nothing recognised -> errcode.UnsupportedServiceManager")
}

// ForManager returns the Installer for m.
func ForManager(m Manager, unitDir, config string) (Installer, error) {
	panic("SUB-AGENT-TODO: T1 unsupported managers return ErrUnsupportedManager with errcode.UnsupportedServiceManager and foreground instructions, writing nothing to unitDir")
}
