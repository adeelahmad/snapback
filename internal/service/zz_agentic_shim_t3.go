// agentic:shim
package service

import (
	"context"
	"time"
)

// Runner runs a command and returns its stdout.
type Runner func(ctx context.Context, name string, args ...string) ([]byte, error)

// Systemd installs and controls Snapback through systemctl --user.
type Systemd struct {
	UnitDir      string
	Run          Runner
	Ready        func(ctx context.Context) (string, error)
	ReadyTimeout time.Duration
}

// Install is a compile shim.
func (s *Systemd) Install(ctx context.Context, o UnitOptions) error { return nil }

// Start is a compile shim.
func (s *Systemd) Start(ctx context.Context) error { return nil }

// Stop is a compile shim.
func (s *Systemd) Stop(ctx context.Context) error { return nil }

// Status is a compile shim.
func (s *Systemd) Status(ctx context.Context) (string, error) { return "shim", nil }

// Uninstall is a compile shim.
func (s *Systemd) Uninstall(ctx context.Context) error { return nil }
