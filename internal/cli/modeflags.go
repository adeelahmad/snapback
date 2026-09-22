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
	fs.StringVar(&m.umask, "umask", "", "umask")
	fs.StringVar(&m.dirMode, "dir-mode", "", "directory mode")
	fs.StringVar(&m.fileMode, "file-mode", "", "file mode")
	return m
}

// Resolve folds the flags over the files section of the config: a flag
// overrides the matching config key, and a flag umask replaces both explicit
// config modes.
func (m *ModeFlags) Resolve(cfg config.Files) (fsmode.Modes, error) {
	return fsmode.Modes{}, nil
}
