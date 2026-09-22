package fsmode

import "os"

// Placeholder modes for the compile shim below. They are deliberately neither
// the configured modes nor any mode a caller asks for, so every mode assertion
// in create_test.go fails until the real helpers land.
const (
	shimDirMode  = 0o711
	shimFileMode = 0o644
)

// MkdirAll creates path and every missing parent with m.Dir.
//
// SUB-AGENT-TODO: the real implementation must chmod every directory it creates
// so an inherited process umask cannot clear bits, must leave an existing
// directory's mode alone, and must return an error naming the path.
func MkdirAll(path string, m Modes) error {
	_ = m
	_ = os.MkdirAll(path, shimDirMode)
	_ = os.Chmod(path, shimDirMode)
	return nil
}

// WriteFile writes data to path, creating it with m.File.
//
// SUB-AGENT-TODO: the real implementation must chmod the file it creates, leave
// an existing file's mode alone, and return an error naming the path.
func WriteFile(path string, data []byte, m Modes) error {
	_ = m
	_ = os.WriteFile(path, data, shimFileMode)
	_ = os.Chmod(path, shimFileMode)
	return nil
}
