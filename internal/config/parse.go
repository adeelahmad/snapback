package config

import (
	"bytes"
	"errors"
	"strings"

	"go.yaml.in/yaml/v3"

	"github.com/adeelahmad/snapback/internal/errcode"
)

// Parse strictly decodes YAML config data, applies defaults and validates it.
func Parse(data []byte) (*Config, error) {
	c := defaults()
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(c); err != nil {
		if strings.Contains(err.Error(), "field password not found") {
			err = errors.New(err.Error() + " (use password_file)")
		}
		return nil, errcode.New(errcode.InvalidConfig, "config.parse", err)
	}
	applyDefaults(c)
	if err := Validate(c); err != nil {
		return nil, err
	}
	return c, nil
}

// Marshal encodes c as YAML that Parse reads back to an equal Config.
func Marshal(c *Config) ([]byte, error) {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(c); err != nil {
		return nil, err
	}
	if err := enc.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
