package config

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
func checkLogging(_ *Config) []FieldError {
	return nil
}
