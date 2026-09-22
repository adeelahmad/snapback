// Package logging turns the user-facing log settings (level, format, file)
// into the values the slog handlers are built from.
package logging

import (
	"fmt"
	"io"
	"log/slog"
	"strings"
)

// Format is the encoding a log record is written in.
type Format string

// The formats Snapback can write.
const (
	FormatText Format = "text"
	FormatJSON Format = "json"
)

// levels are the level names Snapback accepts, with the slog.Level each one
// means, in the order they are listed to the user.
var levels = []struct {
	name  string
	level slog.Level
}{
	{"debug", slog.LevelDebug},
	{"info", slog.LevelInfo},
	{"warn", slog.LevelWarn},
	{"error", slog.LevelError},
}

// formats are the formats Snapback accepts, in the order they are listed to
// the user.
var formats = []Format{FormatText, FormatJSON}

// LevelNames returns the accepted log level names, in the order they are
// listed to the user. Callers that name them in help or error text join this
// list instead of spelling the words again.
func LevelNames() []string {
	names := make([]string, len(levels))
	for i, l := range levels {
		names[i] = l.name
	}
	return names
}

// FormatNames returns the accepted log format names, in the order they are
// listed to the user.
func FormatNames() []string {
	names := make([]string, len(formats))
	for i, f := range formats {
		names[i] = string(f)
	}
	return names
}

// Options are the log settings exactly as the user spelled them, from flags or
// from the config file. The zero value means "the defaults".
type Options struct {
	Level  string
	Format string
	File   string

	// Writer is where records are written. A nil Writer means the caller's
	// fallback writer.
	Writer io.Writer
}

// Resolved is a set of Options that has been validated and given defaults.
type Resolved struct {
	Level  slog.Level
	Format Format
	File   string
}

// Parse maps a level name to its slog.Level, ignoring case.
func Parse(s string) (slog.Level, error) {
	want := strings.ToLower(s)
	for _, l := range levels {
		if l.name == want {
			return l.level, nil
		}
	}
	return 0, fmt.Errorf("unknown log level %q: want one of %s", s, strings.Join(LevelNames(), ", "))
}

// ParseFormat maps a format name to its Format, ignoring case.
func ParseFormat(s string) (Format, error) {
	want := strings.ToLower(s)
	for _, f := range formats {
		if string(f) == want {
			return f, nil
		}
	}
	return "", fmt.Errorf("unknown log format %q: want one of %s", s, strings.Join(FormatNames(), ", "))
}

// Resolve validates o and fills in the defaults for the fields it leaves empty.
func (o Options) Resolve() (Resolved, error) {
	level := slog.LevelInfo
	if o.Level != "" {
		parsed, err := Parse(o.Level)
		if err != nil {
			return Resolved{}, err
		}
		level = parsed
	}

	format := FormatText
	if o.Format != "" {
		parsed, err := ParseFormat(o.Format)
		if err != nil {
			return Resolved{}, err
		}
		format = parsed
	}

	return Resolved{Level: level, Format: format, File: o.File}, nil
}
