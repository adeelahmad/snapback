package daemon

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"golang.org/x/sys/unix"

	"github.com/adeelahmad/snapback/internal/errcode"
)

// lockFunc is the seam tests use to count lock acquisitions.
var lockFunc = Lock

const (
	// lockRetryWindow bounds how long Lock waits out a contended lock, so a
	// CLI command's momentary probe does not read as a running daemon.
	lockRetryWindow = 2 * time.Second
	// lockRetryInterval is the poll interval within lockRetryWindow.
	lockRetryInterval = 25 * time.Millisecond
)

// Lock takes the single-instance lock in stateDir, retrying a contended lock
// for lockRetryWindow before reporting that another daemon is running.
func Lock(stateDir string) (unlock func(), err error) {
	return lockWith(stateDir, lockRetryWindow, lockRetryInterval)
}

// lockWith is Lock with an explicit retry window and poll interval.
func lockWith(stateDir string, window, interval time.Duration) (unlock func(), err error) {
	f, err := os.OpenFile(filepath.Join(stateDir, "daemon.lock"), os.O_RDWR|os.O_CREATE, 0o600)
	if err != nil {
		return nil, fmt.Errorf("daemon: open lock: %w", err)
	}
	if err := flockRetry(f, window, interval); err != nil {
		_ = f.Close()
		return nil, err
	}
	pid := []byte(strconv.Itoa(os.Getpid()) + "\n")
	if err := os.WriteFile(filepath.Join(stateDir, "daemon.pid"), pid, 0o600); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("daemon: write pidfile: %w", err)
	}
	return func() {
		_ = unix.Flock(int(f.Fd()), unix.LOCK_UN)
		_ = f.Close()
	}, nil
}

// flockRetry takes LOCK_EX on f, polling every interval until window elapses.
// It reports StaleState only once the window is spent, so that a momentary
// shared probe by a CLI command does not look like a second daemon.
func flockRetry(f *os.File, window, interval time.Duration) error {
	deadline := time.Now().Add(window)
	for {
		err := unix.Flock(int(f.Fd()), unix.LOCK_EX|unix.LOCK_NB)
		if err == nil {
			return nil
		}
		if !errors.Is(err, unix.EWOULDBLOCK) {
			return fmt.Errorf("daemon: flock: %w", err)
		}
		if !time.Now().Before(deadline) {
			return errcode.New(errcode.StaleState, "daemon lock", errors.New("another daemon is running"))
		}
		time.Sleep(interval)
	}
}

// Running reports whether a daemon holds the single-instance lock in
// stateDir. It probes with a shared, non-blocking lock and releases it at
// once if the probe acquires it; it never creates the lock file or writes the pidfile.
func Running(stateDir string) bool {
	f, err := os.OpenFile(filepath.Join(stateDir, "daemon.lock"), os.O_RDWR, 0)
	if err != nil {
		return false
	}
	defer func() { _ = f.Close() }()
	if err := unix.Flock(int(f.Fd()), unix.LOCK_SH|unix.LOCK_NB); err != nil {
		return errors.Is(err, unix.EWOULDBLOCK)
	}
	_ = unix.Flock(int(f.Fd()), unix.LOCK_UN)
	return false
}
