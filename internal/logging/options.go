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
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("unknown log level %q: want one of debug, info, warn, error", s)
	}
}

// ParseFormat maps a format name to its Format, ignoring case.
func ParseFormat(s string) (Format, error) {
	switch strings.ToLower(s) {
	case string(FormatText):
		return FormatText, nil
	case string(FormatJSON):
		return FormatJSON, nil
	default:
		return "", fmt.Errorf("unknown log format %q: want one of text, json", s)
	}
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
