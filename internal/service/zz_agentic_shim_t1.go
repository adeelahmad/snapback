// agentic:shim

package service

import "errors"

// Manager names a service manager.
type Manager string

// Probe reads the running init system.
type Probe struct {
	PID1Comm func() (string, error)
	Exists   func(path string) bool
}

// Installer installs and controls the Snapback service.
type Installer interface{}

// ErrUnsupportedManager reports a service manager Snapback cannot install for.
var ErrUnsupportedManager = errors.New("shim")

// Detect returns the running service manager.
func Detect(probe Probe) (Manager, error) {
	return "shim", nil
}

// ForManager returns the Installer for m.
func ForManager(m Manager, unitDir, config string) (Installer, error) {
	return nil, nil
}
