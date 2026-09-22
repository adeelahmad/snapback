package config

import (
	"path/filepath"
	"strings"

	"github.com/adeelahmad/snapback/internal/logging"
)

// Logging is the `logging:` section: the log settings exactly as the user
// spelled them. Every field is optional; the zero value means "the defaults",
// which internal/logging resolves.
type Logging struct {
	Level  string `yaml:"level,omitempty"`
	Format string `yaml:"format,omitempty"`
	File   string `yaml:"file,omitempty"`
}

// checkLogging reports the errors in the logging section: an unknown level on
// logging.level, an unknown format on logging.format and a file that is not an
// absolute, clean path on logging.file. An absent section is valid.
func checkLogging(c *Config) []FieldError {
	var errs []FieldError
	if c.Logging.Level != "" {
		if _, err := logging.Parse(c.Logging.Level); err != nil {
			errs = append(errs, FieldError{Path: "logging.level", Msg: "must be one of " + strings.Join(logging.LevelNames(), ", ")})
		}
	}
	if c.Logging.Format != "" {
		if _, err := logging.ParseFormat(c.Logging.Format); err != nil {
			errs = append(errs, FieldError{Path: "logging.format", Msg: "must be one of " + strings.Join(logging.FormatNames(), ", ")})
		}
	}
	if f := c.Logging.File; f != "" && (!filepath.IsAbs(f) || filepath.Clean(f) != f) {
		errs = append(errs, FieldError{Path: "logging.file", Msg: "must be an absolute, clean path"})
	}
	return errs
}
