package cli

import (
	"flag"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/fsmode"
)

// ModeFlags holds the values of the shared -umask, -dir-mode and -file-mode
// flags a command registers with [AddModeFlags].
type ModeFlags struct {
	umask    string
	dirMode  string
	fileMode string
}

// AddModeFlags registers -umask, -dir-mode and -file-mode on fs and returns
// the values they parse into.
func AddModeFlags(fs *flag.FlagSet) *ModeFlags {
	m := &ModeFlags{}
	fs.StringVar(&m.umask, "umask", "", "octal umask for created state, sugar for -dir-mode and -file-mode")
	fs.StringVar(&m.dirMode, "dir-mode", "", "octal mode for directories Snapback creates, e.g. 0750")
	fs.StringVar(&m.fileMode, "file-mode", "", "octal mode for files Snapback creates, e.g. 0640; "+
		"credential files stay 0600 and credential directories 0700 regardless")
	return m
}

// Resolve folds the flags over the files section of the config: a flag
// overrides the matching config key, and a flag umask replaces both explicit
// config modes.
func (m *ModeFlags) Resolve(cfg config.Files) (fsmode.Modes, error) {
	if m.umask != "" {
		if m.dirMode != "" || m.fileMode != "" {
			return fsmode.Modes{}, &UsageError{
				Msg: "-umask cannot be combined with -dir-mode or -file-mode: pick one",
			}
		}
		modes, err := fsmode.FromUmask(m.umask)
		if err != nil {
			return fsmode.Modes{}, &UsageError{Msg: "-umask: " + err.Error()}
		}
		return modes, nil
	}

	for _, f := range []struct{ name, value string }{
		{"-dir-mode", m.dirMode},
		{"-file-mode", m.fileMode},
	} {
		if f.value == "" {
			continue
		}
		if _, err := fsmode.Parse(f.value); err != nil {
			return fsmode.Modes{}, &UsageError{Msg: f.name + ": " + err.Error()}
		}
		cfg.Umask = "" // a flag mode wins over the config umask
	}
	if m.dirMode != "" {
		cfg.DirMode = m.dirMode
	}
	if m.fileMode != "" {
		cfg.FileMode = m.fileMode
	}

	modes, err := cfg.Modes()
	if err != nil {
		return fsmode.Modes{}, &UsageError{Msg: err.Error()}
	}
	return modes, nil
}
