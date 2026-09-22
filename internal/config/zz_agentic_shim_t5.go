// agentic:shim

package config

import "errors"

// Parse is a RED compile shim with a deliberately wrong body.
func Parse(data []byte) (*Config, error) { return nil, errors.New("shim: parse") }

// DefaultPath is a RED compile shim with a deliberately wrong body.
func DefaultPath() (string, error) { return "", errors.New("shim: default path") }

// Marshal is a RED compile shim with a deliberately wrong body.
func Marshal(c *Config) ([]byte, error) { return nil, errors.New("shim: marshal") }

// Redact is a RED compile shim with a deliberately wrong body.
func Redact(c *Config) *Config { return &Config{} }
