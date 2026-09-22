package cli

import "flag"

// NewFlagSet returns a flag set for u.
func NewFlagSet(env Env, u Usage) *flag.FlagSet {
	return flag.NewFlagSet(u.Synopsis, flag.ContinueOnError)
}

// ParseWithUsage parses args with fs.
func ParseWithUsage(fs *flag.FlagSet, args []string) (help bool, err error) {
	return false, nil
}
