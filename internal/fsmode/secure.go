package fsmode

import "io/fs"

// SecureFile and SecureDir are the fixed modes for state that is never widened by
// configuration: the credential store, password files, daemon.lock, daemon.pid,
// the ipc socket directory, the saved config and the diagnostic bundle. Those paths
// always use 0o600 and 0o700 whatever files.dirMode and files.fileMode are set to.
const (
	SecureFile fs.FileMode = 0o600
	SecureDir  fs.FileMode = 0o700
)

// Secure is the mode pair those paths use whatever files.* is configured to.
func Secure() Modes {
	return Modes{Dir: SecureDir, File: SecureFile}
}
