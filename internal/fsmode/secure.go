package fsmode

import "io/fs"

// SecureFile and SecureDir are the fixed modes for state that is never widened by
// configuration: the credential store, password files, daemon.lock, daemon.pid,
// the ipc socket directory, the saved config and the diagnostic bundle.
const (
	SecureFile fs.FileMode = 0 // SUB-AGENT-TODO: always 0o600.
	SecureDir  fs.FileMode = 0 // SUB-AGENT-TODO: always 0o700.
)

// Secure is the mode pair those paths use whatever files.* is configured to.
func Secure() Modes {
	return Modes{} // SUB-AGENT-TODO: {Dir: SecureDir, File: SecureFile}.
}
