package cli

import (
	"flag"
	"fmt"

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
	f := &LogFlags{}
	fs.StringVar(&f.Level, "log-level", "", "log level, one of debug, info, warn, error; overrides logging.level")
	fs.StringVar(&f.Format, "log-format", "", "log format, one of text, json; overrides logging.format")
	fs.StringVar(&f.File, "log-file", "", "write logs to this file; overrides logging.file")
	return f
}

// Resolve merges f over cfg field by field: a flag that was given wins over
// the configured value, and a field neither gives is left empty for
// [logging.Options.Resolve] to default. An invalid flag value is a
// [UsageError] naming the flag and the accepted words.
func (f *LogFlags) Resolve(cfg config.Logging) (logging.Options, error) {
	opts := logging.Options{
		Level:  merge(f.Level, cfg.Level),
		Format: merge(f.Format, cfg.Format),
		File:   merge(f.File, cfg.File),
	}
	if opts.Level != "" {
		if _, err := logging.Parse(opts.Level); err != nil {
			return logging.Options{}, flagValueError("--log-level", f.Level != "", err)
		}
	}
	if opts.Format != "" {
		if _, err := logging.ParseFormat(opts.Format); err != nil {
			return logging.Options{}, flagValueError("--log-format", f.Format != "", err)
		}
	}
	return opts, nil
}

// merge picks the flag value when it was given, else the configured one.
func merge(flagValue, cfgValue string) string {
	if flagValue != "" {
		return flagValue
	}
	return cfgValue
}

// flagValueError blames the flag when the bad value came from the command
// line, and reports a plain error when it came from the config file.
func flagValueError(name string, fromFlag bool, err error) error {
	if fromFlag {
		return &UsageError{Msg: fmt.Sprintf("%s: %v", name, err)}
	}
	return err
}
