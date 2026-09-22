// Package fsmode parses and resolves the file and directory modes Snapback uses
// when it creates its own state.
//
// The ruling this package encodes: explicit directory and file modes are the
// primary knob, an umask is only sugar that computes that same pair, and
// credential files and their directories always stay 0600 and 0700 regardless of
// either setting.
package fsmode

import (
	"fmt"
	"io/fs"
	"strconv"
	"strings"
)

// Default modes for state Snapback creates when no mode is configured.
const (
	defaultDirMode  fs.FileMode = 0o700
	defaultFileMode fs.FileMode = 0o600

	maxMode     = 0o7777
	specialBits = 0o7000 // set-uid, set-gid and sticky
	otherWrite  = 0o002
	ownerRW     = 0o600
)

// Modes is the pair of modes Snapback creates its own directories and files with.
type Modes struct {
	Dir  fs.FileMode
	File fs.FileMode
}

// Parse reads an octal mode such as "0750", "750" or "0o750".
func Parse(s string) (fs.FileMode, error) {
	v, err := strconv.ParseUint(strings.TrimPrefix(s, "0o"), 8, 32)
	if err != nil {
		return 0, fmt.Errorf("mode %q is not an octal number", s)
	}
	switch {
	case v > maxMode:
		return 0, fmt.Errorf("mode %q is out of range", s)
	case v&specialBits != 0:
		return 0, fmt.Errorf("mode %q sets a set-uid, set-gid or sticky bit", s)
	case v&otherWrite != 0:
		return 0, fmt.Errorf("mode %q is world writable", s)
	case v&ownerRW != ownerRW:
		return 0, fmt.Errorf("mode %q does not give the owner read and write", s)
	}
	return fs.FileMode(v), nil
}

// FromUmask computes the two modes from an octal umask such as "027".
func FromUmask(s string) (Modes, error) {
	u, err := strconv.ParseUint(strings.TrimPrefix(s, "0o"), 8, 32)
	if err != nil || u > 0o777 {
		return Modes{}, fmt.Errorf("umask %q is not an octal value between 0 and 0777", s)
	}
	return Modes{
		Dir:  fs.FileMode(0o777 &^ u),
		File: fs.FileMode(0o666 &^ u),
	}, nil
}

// OrDefault fills the zero-valued fields with Snapback's defaults.
func (m Modes) OrDefault() Modes {
	if m.Dir == 0 {
		m.Dir = defaultDirMode
	}
	if m.File == 0 {
		m.File = defaultFileMode
	}
	return m
}
