package latency

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"
)

// scratch is a disposable temp tree holding the run's password file and working dirs.
type scratch struct {
	root         string
	passwordFile string
	cacheDir     string
	dataDir      string
	mountDir     string
}

func newScratch() (scratch, error) {
	root, err := os.MkdirTemp(os.TempDir(), "snapback-latency-")
	if err != nil {
		return scratch{}, err
	}
	s := scratch{
		root:         root,
		passwordFile: filepath.Join(root, "password"),
		cacheDir:     filepath.Join(root, "cache"),
		dataDir:      filepath.Join(root, "data"),
		mountDir:     filepath.Join(root, "mount"),
	}
	if err := s.populate(); err != nil {
		_ = os.RemoveAll(root)
		return scratch{}, err
	}
	return s, nil
}

func (s scratch) populate() error {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return err
	}
	if err := os.WriteFile(s.passwordFile, []byte(hex.EncodeToString(secret)+"\n"), 0o600); err != nil {
		return err
	}
	for _, dir := range []string{s.cacheDir, s.dataDir, s.mountDir} {
		if err := os.Mkdir(dir, 0o700); err != nil {
			return err
		}
	}
	return nil
}

// Close removes the scratch root and everything under it.
func (s scratch) Close() error {
	return os.RemoveAll(s.root)
}
