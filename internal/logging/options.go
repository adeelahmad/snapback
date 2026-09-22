// Package logging turns the user-facing log settings (level, format, file)
// into the values the slog handlers are built from.
package logging

import "log/slog"

// Format is the encoding a log record is written in.
type Format string

// The formats Snapback can write.
const (
	FormatText Format = "text"
	FormatJSON Format = "json"
)

// Options are the log settings exactly as the user spelled them, from flags or
// from the config file. The zero value means "the defaults".
type Options struct {
	Level  string
	Format string
	File   string
}

// Resolved is a set of Options that has been validated and given defaults.
type Resolved struct {
	Level  slog.Level
	Format Format
	File   string
}

// Parse maps a level name to its slog.Level, ignoring case.
func Parse(s string) (slog.Level, error) {
	return 0, nil // SUB-AGENT-TODO
}

// ParseFormat maps a format name to its Format, ignoring case.
func ParseFormat(s string) (Format, error) {
	return "", nil // SUB-AGENT-TODO
}

// Resolve validates o and fills in the defaults for the fields it leaves empty.
func (o Options) Resolve() (Resolved, error) {
	return Resolved{}, nil // SUB-AGENT-TODO
}
