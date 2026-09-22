package config

import (
	"errors"
	"fmt"

	"github.com/adeelahmad/snapback/internal/fsmode"
)

// Files is the `files:` section: the modes Snapback creates its own state with,
// exactly as the user spelled them. Every field is optional; the zero value
// means "the defaults", which fsmode.Modes.OrDefault resolves.
//
// DirMode and FileMode are the primary spelling. Umask is sugar that computes
// the same pair; setting it together with either explicit mode is an error.
type Files struct {
	DirMode  string `yaml:"dir_mode,omitempty"`
	FileMode string `yaml:"file_mode,omitempty"`
	Umask    string `yaml:"umask,omitempty"`
}

// Modes resolves the section into the pair of modes Snapback creates
// directories and files with. An absent section resolves to the defaults.
func (f Files) Modes() (fsmode.Modes, error) {
	if f.Umask != "" {
		if f.DirMode != "" || f.FileMode != "" {
			return fsmode.Modes{}, errors.New(umaskConflict)
		}
		return fsmode.FromUmask(f.Umask)
	}

	var m fsmode.Modes
	var err error
	if f.DirMode != "" {
		if m.Dir, err = fsmode.Parse(f.DirMode); err != nil {
			return fsmode.Modes{}, fmt.Errorf("files.dir_mode: %w", err)
		}
	}
	if f.FileMode != "" {
		if m.File, err = fsmode.Parse(f.FileMode); err != nil {
			return fsmode.Modes{}, fmt.Errorf("files.file_mode: %w", err)
		}
	}
	return m.OrDefault(), nil
}

// checkFiles reports the errors in the files section: an unparsable mode on
// files.dir_mode or files.file_mode, an unparsable umask on files.umask and a
// umask set together with an explicit mode, also on files.umask. An absent
// section is valid.
func checkFiles(c *Config) []FieldError {
	f := c.Files
	var errs []FieldError
	if f.Umask != "" {
		if f.DirMode != "" || f.FileMode != "" {
			return append(errs, FieldError{Path: "files.umask", Msg: umaskConflict})
		}
		if _, err := fsmode.FromUmask(f.Umask); err != nil {
			errs = append(errs, FieldError{Path: "files.umask", Msg: err.Error()})
		}
		return errs
	}
	if f.DirMode != "" {
		if _, err := fsmode.Parse(f.DirMode); err != nil {
			errs = append(errs, FieldError{Path: "files.dir_mode", Msg: err.Error()})
		}
	}
	if f.FileMode != "" {
		if _, err := fsmode.Parse(f.FileMode); err != nil {
			errs = append(errs, FieldError{Path: "files.file_mode", Msg: err.Error()})
		}
	}
	return errs
}

// umaskConflict is the message for a umask set together with an explicit mode.
const umaskConflict = "umask cannot be combined with an explicit dir_mode or file_mode: pick one spelling"
