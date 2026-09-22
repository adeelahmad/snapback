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

// Install writes the unit, reloads and enables it, and waits for readiness.
func (s *Systemd) Install(ctx context.Context, o UnitOptions) error {
	panic("SUB-AGENT-TODO: write the rendered unit to UnitDir/snapback.service (refusing to overwrite a foreign unit with errcode.StaleState), run systemctl --user daemon-reload only on first install, run systemctl --user enable --now snapback.service, then poll Ready until it reports \"ready\" or ReadyTimeout elapses, returning an errcode.StaleState error naming both states and the unit path plus a journalctl --user -u snapback hint on timeout")
}

// Start starts the installed unit.
func (s *Systemd) Start(ctx context.Context) error {
	panic("SUB-AGENT-TODO: run systemctl --user start snapback.service via Run")
}

// Stop stops the installed unit.
func (s *Systemd) Stop(ctx context.Context) error {
	panic("SUB-AGENT-TODO: run systemctl --user stop snapback.service via Run")
}

// Status reports the unit's current systemctl state.
func (s *Systemd) Status(ctx context.Context) (string, error) {
	panic("SUB-AGENT-TODO: run systemctl --user is-active snapback.service via Run and return the trimmed stdout")
}

// Uninstall stops, disables and removes the unit, refusing to touch a foreign one.
func (s *Systemd) Uninstall(ctx context.Context) error {
	panic("SUB-AGENT-TODO: refuse with errcode.StaleState if the unit at UnitDir/snapback.service is foreign, else run systemctl --user stop, disable, daemon-reload via Run and remove only the unit file, leaving sibling state files untouched")
}
