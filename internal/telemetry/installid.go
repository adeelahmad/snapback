package telemetry

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/adeelahmad/snapback/internal/fsmode"
)

// installIDFile is the name of the file holding the per-install identifier.
const installIDFile = "install_id"

// installIDBytes is how much randomness an identifier carries, in bytes.
const installIDBytes = 16

// InstallID returns the random per-install identifier stored in <dir>/install_id,
// creating the directory (0700) and the file (0600) on first use.
//
// The identifier is 32 lowercase hex characters drawn from crypto/rand, so two
// installs on one host never share one. A stored value of any other shape is an
// error naming the file: a corrupt identifier is never silently replaced.
func InstallID(dir string) (string, error) {
	path := filepath.Join(dir, installIDFile)

	raw, err := os.ReadFile(path)
	switch {
	case err == nil:
		id := strings.TrimSpace(string(raw))
		if !isInstallID(id) {
			return "", fmt.Errorf("telemetry: %s does not hold 32 lowercase hex characters", path)
		}
		return id, nil
	case !errors.Is(err, fs.ErrNotExist):
		return "", fmt.Errorf("telemetry: read install id: %w", err)
	}

	var buf [installIDBytes]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", fmt.Errorf("telemetry: generate install id: %w", err)
	}
	id := hex.EncodeToString(buf[:])

	if err := fsmode.MkdirAll(dir, fsmode.Secure()); err != nil {
		return "", fmt.Errorf("telemetry: store install id: %w", err)
	}
	if err := fsmode.WriteFile(path, []byte(id+"\n"), fsmode.Secure()); err != nil {
		return "", fmt.Errorf("telemetry: store install id: %w", err)
	}
	return id, nil
}

// ForgetInstallID removes <dir>/install_id, so the next InstallID mints a fresh
// one. A missing file, or a missing directory, is not an error.
func ForgetInstallID(dir string) error {
	path := filepath.Join(dir, installIDFile)
	if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("telemetry: forget install id: %w", err)
	}
	return nil
}

// isInstallID reports whether s is exactly 32 lowercase hex characters.
func isInstallID(s string) bool {
	if len(s) != hex.EncodedLen(installIDBytes) {
		return false
	}
	for _, r := range s {
		switch {
		case r >= '0' && r <= '9', r >= 'a' && r <= 'f':
		default:
			return false
		}
	}
	return true
}
