package daemon

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"golang.org/x/sys/unix"

	"github.com/adeelahmad/snapback/internal/errcode"
)

// lockFunc is the seam tests use to count lock acquisitions.
var lockFunc = Lock

// Lock takes the single-instance lock in stateDir.
func Lock(stateDir string) (unlock func(), err error) {
	f, err := os.OpenFile(filepath.Join(stateDir, "daemon.lock"), os.O_RDWR|os.O_CREATE, 0o600)
	if err != nil {
		return nil, fmt.Errorf("daemon: open lock: %w", err)
	}
	if err := unix.Flock(int(f.Fd()), unix.LOCK_EX|unix.LOCK_NB); err != nil {
		_ = f.Close()
		if errors.Is(err, unix.EWOULDBLOCK) {
			return nil, errcode.New(errcode.StaleState, "daemon lock", errors.New("another daemon is running"))
		}
		return nil, fmt.Errorf("daemon: flock: %w", err)
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
