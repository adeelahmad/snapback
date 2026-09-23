package fsmode

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// MkdirAll creates path and every missing parent with m.Dir.
//
// Each directory it creates is chmod-ed after the fact, because the process
// umask clears bits from the mode os.Mkdir is given. Directories that already
// exist, including one a concurrent creator won the race to make, keep the
// mode they have.
func MkdirAll(path string, m Modes) error {
	var missing []string
	for p := path; ; {
		if _, err := os.Lstat(p); err == nil {
			break
		}
		missing = append(missing, p)
		parent := filepath.Dir(p)
		if parent == p {
			break
		}
		p = parent
	}
	for i := len(missing) - 1; i >= 0; i-- {
		if err := os.Mkdir(missing[i], m.Dir); err != nil {
			if errors.Is(err, fs.ErrExist) {
				continue
			}
			return fmt.Errorf("create directory %s: %w", path, err)
		}
		if err := os.Chmod(missing[i], m.Dir); err != nil {
			return fmt.Errorf("create directory %s: %w", path, err)
		}
	}
	return nil
}

// WriteFile writes data to path, creating it with m.File.
//
// A file it creates is chmod-ed after the fact so the umask cannot clear bits.
// A file that already exists keeps the mode it has.
func WriteFile(path string, data []byte, m Modes) error {
	flag := os.O_WRONLY | os.O_TRUNC
	created := false
	if _, err := os.Lstat(path); err != nil {
		flag |= os.O_CREATE
		created = true
	}
	f, err := os.OpenFile(path, flag, m.File)
	if err != nil {
		return fmt.Errorf("write file %s: %w", path, err)
	}
	if created {
		if err := f.Chmod(m.File); err != nil {
			_ = f.Close()
			return fmt.Errorf("write file %s: %w", path, err)
		}
	}
	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		return fmt.Errorf("write file %s: %w", path, err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("write file %s: %w", path, err)
	}
	return nil
}
