// Package fsmode parses and resolves the file and directory modes Snapback uses
// when it creates its own state. Modes are explicit: the process umask is only a
// convenience spelling that computes the two modes.
package fsmode

import "io/fs"

// Modes is the pair of modes Snapback creates its own directories and files with.
type Modes struct {
	Dir  fs.FileMode
	File fs.FileMode
}

// Parse reads an octal mode such as "0750", "750" or "0o750".
func Parse(s string) (fs.FileMode, error) {
	return 0, nil // SUB-AGENT-TODO
}

// FromUmask computes the two modes from an octal umask such as "027".
func FromUmask(s string) (Modes, error) {
	return Modes{}, nil // SUB-AGENT-TODO
}

// OrDefault fills the zero-valued fields with Snapback's defaults.
func (m Modes) OrDefault() Modes {
	return Modes{} // SUB-AGENT-TODO
}
