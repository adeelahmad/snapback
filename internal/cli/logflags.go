package cli

import (
	"flag"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/logging"
)

// LogFlags holds the values of the shared --log-level, --log-format and
// --log-file flags. An empty field means "not given on the command line", so
// the configured value, and then the default, wins.
type LogFlags struct {
	Level  string
	Format string
	File   string
}

// AddLogFlags registers --log-level, --log-format and --log-file on fs and
// returns the values they parse into. Every flag defaults to the empty string,
// which means "use the logging section of the config, then the defaults".
func AddLogFlags(fs *flag.FlagSet) *LogFlags {
	return &LogFlags{}
}

// Resolve merges f over cfg field by field: a flag that was given wins over
// the configured value, and a field neither gives is left empty for
// [logging.Options.Resolve] to default. An invalid flag value is a
// [UsageError] naming the flag and the accepted words.
func (f *LogFlags) Resolve(cfg config.Logging) (logging.Options, error) {
	return logging.Options{}, nil
}
