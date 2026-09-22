package logging

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// Open returns the writer log records are written to. A non-empty path is a log
// file: its missing parent directories are created with dirMode, the file is
// created with fileMode and opened for appending so an existing log is never
// truncated, and the returned io.Closer closes it. An empty path means "no log
// file": fallback is returned unchanged with a closer that does nothing.
func Open(path string, dirMode, fileMode fs.FileMode, fallback io.Writer) (io.Writer, io.Closer, error) {
	if path == "" {
		return fallback, nopCloser{}, nil
	}
	// Only a directory this call creates is chmod'ed: an existing one keeps the
	// permissions its owner chose, and MkdirAll's modes are subject to the umask.
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); err != nil {
		if err := os.MkdirAll(dir, dirMode); err != nil {
			return nil, nil, fmt.Errorf("open log file %s: %w", path, err)
		}
		if err := os.Chmod(dir, dirMode); err != nil {
			return nil, nil, fmt.Errorf("open log file %s: %w", path, err)
		}
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, fileMode)
	if err != nil {
		return nil, nil, fmt.Errorf("open log file %s: %w", path, err)
	}
	if err := f.Chmod(fileMode); err != nil {
		_ = f.Close()
		return nil, nil, fmt.Errorf("open log file %s: %w", path, err)
	}
	return f, f, nil
}

// nopCloser is the io.Closer returned when there is no log file to close.
type nopCloser struct{}

func (nopCloser) Close() error { return nil }
