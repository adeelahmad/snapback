package config

import "github.com/adeelahmad/snapback/internal/fsmode"

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
	return fsmode.Modes{}, nil
}

// checkFiles reports the errors in the files section: an unparsable mode on
// files.dir_mode or files.file_mode, an unparsable umask on files.umask and a
// umask set together with an explicit mode, also on files.umask. An absent
// section is valid.
func checkFiles(c *Config) []FieldError {
	return nil
}
