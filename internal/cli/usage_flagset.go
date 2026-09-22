package cli

import (
	"errors"
	"flag"
	"fmt"
)

// NewFlagSet returns a ContinueOnError flag set whose usage prints u followed
// by the flag defaults to env.Stderr.
func NewFlagSet(env Env, u Usage) *flag.FlagSet {
	fs := flag.NewFlagSet(u.Synopsis, flag.ContinueOnError)
	fs.SetOutput(env.Stderr)
	fs.Usage = func() {
		_, _ = fmt.Fprintf(env.Stderr, "%s\nFlags:\n", u)
		fs.PrintDefaults()
	}
	return fs
}

// ParseWithUsage parses args with fs, which must come from [NewFlagSet] so
// that the usage is printed on -h and on a bad flag. It reports help for -h,
// for which the caller exits 0, and wraps any other parse failure in a
// [UsageError], for which the caller exits 2.
func ParseWithUsage(fs *flag.FlagSet, args []string) (help bool, err error) {
	switch err := fs.Parse(args); {
	case err == nil:
		return false, nil
	case errors.Is(err, flag.ErrHelp):
		return true, nil
	default:
		return false, &UsageError{Msg: err.Error()}
	}
}
