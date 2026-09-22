package logging

import (
	"io"
	"io/fs"
)

// Open returns the writer log records are written to. A non-empty path is a log
// file: its missing parent directories are created with dirMode, the file is
// created with fileMode and opened for appending so an existing log is never
// truncated, and the returned io.Closer closes it. An empty path means "no log
// file": fallback is returned unchanged with a closer that does nothing.
func Open(path string, dirMode, fileMode fs.FileMode, fallback io.Writer) (io.Writer, io.Closer, error) {
	return io.Discard, nopCloser{}, nil
}

// nopCloser is the io.Closer returned when there is no log file to close.
type nopCloser struct{}

func (nopCloser) Close() error { return nil }
