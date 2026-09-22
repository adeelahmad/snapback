package config

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"

	"github.com/adeelahmad/snapback/internal/errcode"
)

// ErrRevisionConflict reports that the stored configuration changed since the
// caller read it, so Save refused to overwrite it.
var ErrRevisionConflict = errcode.New(errcode.StaleState, "config.save", errors.New("config revision conflict"))

// rename is the atomic-replace step of Save; tests swap it to inject failures.
var rename = os.Rename

// revisionOf returns the lowercase hex SHA-256 of data.
func revisionOf(data []byte) Revision {
	sum := sha256.Sum256(data)
	return Revision(hex.EncodeToString(sum[:]))
}

// Load reads and validates the configuration file at path and returns its revision.
func Load(path string) (*Config, Revision, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", err
	}
	rev := revisionOf(data)
	c, err := Parse(data)
	if err != nil {
		return nil, rev, err
	}
	return c, rev, nil
}

// currentRevision returns the revision of the file at path, or "" if absent.
func currentRevision(path string) (Revision, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return revisionOf(data), nil
}

// Save validates c and atomically replaces the file at path if its current
// revision equals expected ("" means absent), returning the new revision.
func Save(path string, c *Config, expected Revision) (rev Revision, err error) {
	if err := Validate(c); err != nil {
		return "", err
	}
	data, err := Marshal(c)
	if err != nil {
		return "", fmt.Errorf("config.save: marshal: %w", err)
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("config.save: %w", err)
	}

	lock, err := os.OpenFile(path+".lock", os.O_RDWR|os.O_CREATE, 0o600)
	if err != nil {
		return "", fmt.Errorf("config.save: %w", err)
	}
	// Closing the lock file releases the flock.
	defer func() {
		if cerr := lock.Close(); cerr != nil && err == nil {
			rev, err = "", fmt.Errorf("config.save: %w", cerr)
		}
	}()
	if err := unix.Flock(int(lock.Fd()), unix.LOCK_EX); err != nil {
		return "", fmt.Errorf("config.save: lock: %w", err)
	}

	cur, err := currentRevision(path)
	if err != nil {
		return "", fmt.Errorf("config.save: %w", err)
	}
	if cur != expected {
		return "", ErrRevisionConflict
	}

	if err := writeReplace(dir, path, data); err != nil {
		return "", fmt.Errorf("config.save: %w", err)
	}
	return revisionOf(data), nil
}

// writeReplace writes data to a 0600 temp file in dir, syncs it and renames it
// over path, removing the temp file on any failure.
func writeReplace(dir, path string, data []byte) error {
	tmp, err := os.CreateTemp(dir, filepath.Base(path)+".tmp*")
	if err != nil {
		return err
	}
	if err := writeSync(tmp, data); err != nil {
		return errors.Join(err, os.Remove(tmp.Name()))
	}
	if err := rename(tmp.Name(), path); err != nil {
		return errors.Join(err, os.Remove(tmp.Name()))
	}
	return syncDir(dir)
}

// writeSync sets f to 0600, writes data, fsyncs and closes it.
func writeSync(f *os.File, data []byte) error {
	err := f.Chmod(0o600)
	if err == nil {
		_, err = f.Write(data)
	}
	if err == nil {
		err = f.Sync()
	}
	return errors.Join(err, f.Close())
}

// syncDir fsyncs dir so a completed rename is durable.
func syncDir(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	return errors.Join(d.Sync(), d.Close())
}
