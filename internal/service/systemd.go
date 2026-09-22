package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/adeelahmad/snapback/internal/errcode"
)

const (
	unitName     = "snapback.service"
	ownedMarker  = "# Managed by snapback"
	readyPollGap = 50 * time.Millisecond
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

func (s *Systemd) unitPath() string {
	return filepath.Join(s.UnitDir, unitName)
}

func (s *Systemd) systemctl(ctx context.Context, args ...string) ([]byte, error) {
	return s.Run(ctx, "systemctl", append([]string{"--user"}, args...)...)
}

// existingUnit returns the current unit bytes (nil when absent) and refuses a
// unit Snapback did not write.
func (s *Systemd) existingUnit(op string) ([]byte, error) {
	path := s.unitPath()
	cur, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("%s: read unit: %w", op, err)
	}
	if !bytes.HasPrefix(cur, []byte(ownedMarker)) {
		return nil, errcode.New(errcode.StaleState, op, fmt.Errorf("refusing to touch foreign unit %s", path))
	}
	return cur, nil
}

// Install writes the unit, reloads and enables it, and waits for readiness.
func (s *Systemd) Install(ctx context.Context, o UnitOptions) error {
	const op = "service install"
	unit, err := SystemdUnit(o)
	if err != nil {
		return err
	}
	cur, err := s.existingUnit(op)
	if err != nil {
		return err
	}
	if cur == nil || string(cur) != unit {
		if err := os.MkdirAll(s.UnitDir, 0o755); err != nil {
			return fmt.Errorf("%s: create unit dir: %w", op, err)
		}
		if err := os.Chmod(s.UnitDir, 0o755); err != nil {
			return fmt.Errorf("%s: create unit dir: %w", op, err)
		}
		if err := writeAtomic(s.unitPath(), []byte(unit)); err != nil {
			return fmt.Errorf("%s: write unit: %w", op, err)
		}
		if _, err := s.systemctl(ctx, "daemon-reload"); err != nil {
			return fmt.Errorf("%s: daemon-reload: %w", op, err)
		}
	}
	if _, err := s.systemctl(ctx, "enable", "--now", unitName); err != nil {
		return fmt.Errorf("%s: enable: %w", op, err)
	}
	return s.waitReady(ctx, op)
}

// waitReady polls Ready until it reports "ready" or "degraded", or
// ReadyTimeout elapses. A daemon never reached gives prerequisite_missing; one
// reached but not ready gives stale_state.
func (s *Systemd) waitReady(ctx context.Context, op string) error {
	deadline := time.Now().Add(s.ReadyTimeout)
	state := ""
	up := false
	for {
		if st, err := s.Ready(ctx); err == nil {
			up = true
			state = st
			if st == "ready" || st == "degraded" {
				return nil
			}
		}
		if time.Now().Add(readyPollGap).After(deadline) {
			break
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(readyPollGap):
		}
	}
	if !up {
		return errcode.New(errcode.PrereqMissing, op, fmt.Errorf(
			"service installed and enabled at %s but daemon socket not reachable after %s; see journalctl --user -u snapback",
			s.unitPath(), s.ReadyTimeout))
	}
	return errcode.New(errcode.StaleState, op, fmt.Errorf(
		"service installed and enabled at %s but daemon not ready (state %q) after %s; see journalctl --user -u snapback",
		s.unitPath(), state, s.ReadyTimeout))
}

// writeAtomic writes data to a temp file beside path and renames it into place.
func writeAtomic(path string, data []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".tmp*")
	if err != nil {
		return err
	}
	_, werr := tmp.Write(data)
	cerr := tmp.Chmod(0o644)
	err = errors.Join(werr, cerr, tmp.Close())
	if err == nil {
		err = os.Rename(tmp.Name(), path)
	}
	if err != nil {
		return errors.Join(err, os.Remove(tmp.Name()))
	}
	return nil
}

// Start starts the installed unit.
func (s *Systemd) Start(ctx context.Context) error {
	_, err := s.systemctl(ctx, "start", unitName)
	return err
}

// Stop stops the installed unit.
func (s *Systemd) Stop(ctx context.Context) error {
	_, err := s.systemctl(ctx, "stop", unitName)
	return err
}

// Status reports the unit's current systemctl state.
func (s *Systemd) Status(ctx context.Context) (string, error) {
	out, err := s.systemctl(ctx, "is-active", unitName)
	return strings.TrimSpace(string(out)), err
}

// Uninstall stops, disables and removes the unit, refusing to touch a foreign one.
func (s *Systemd) Uninstall(ctx context.Context) error {
	const op = "service uninstall"
	if _, err := s.existingUnit(op); err != nil {
		return err
	}
	for _, args := range [][]string{{"stop", unitName}, {"disable", unitName}} {
		if _, err := s.systemctl(ctx, args...); err != nil {
			return fmt.Errorf("%s: %s: %w", op, args[0], err)
		}
	}
	if err := os.Remove(s.unitPath()); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("%s: remove unit: %w", op, err)
	}
	if _, err := s.systemctl(ctx, "daemon-reload"); err != nil {
		return fmt.Errorf("%s: daemon-reload: %w", op, err)
	}
	return nil
}
