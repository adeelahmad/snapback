package resticfx

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
)

const passwordBytes = 32

// NewPasswordFile writes a random password to a 0600 file under dir and returns only its path.
func NewPasswordFile(dir string) (string, error) {
	buf := make([]byte, passwordBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate password: %w", err)
	}
	f, err := os.CreateTemp(dir, "restic-password-*")
	if err != nil {
		return "", fmt.Errorf("create password file: %w", err)
	}
	path := f.Name()
	if err := f.Chmod(0o600); err != nil {
		_ = f.Close()
		return "", fmt.Errorf("chmod password file: %w", err)
	}
	if _, err := f.WriteString(hex.EncodeToString(buf)); err != nil {
		_ = f.Close()
		return "", fmt.Errorf("write password file: %w", err)
	}
	if err := f.Close(); err != nil {
		return "", fmt.Errorf("close password file: %w", err)
	}
	return path, nil
}
