package config

import "github.com/adeelahmad/snapback/internal/errcode"

// FieldError is one invalid field. Path is the YAML path, such as
// repositories[0].password_file. Code is empty for ordinary errors and names
// a specific errcode for the v0.1 rejections.
type FieldError struct {
	Path string
	Msg  string
	Code errcode.Code
}

// ValidationError collects every invalid field found by Validate.
type ValidationError struct {
	Fields []FieldError
}

func (e *ValidationError) Error() string {
	panic("SUB-AGENT-TODO: T4 — join every field as \"<path>: <msg>\" with \"; \"")
}

func (e *ValidationError) Unwrap() error {
	panic("SUB-AGENT-TODO: T4 — return errcode.New(errcode.InvalidConfig, \"config.validate\", nil)")
}

// Validate checks every field rule plus the topology and credential rules and
// returns a *ValidationError listing all failures, or nil.
func Validate(c *Config) error {
	panic("SUB-AGENT-TODO: T4 — apply the tasks.md field rules (IDs, uniqueness, repository refs, absolute clean paths, relative seed/exclude paths, enums incl. UnsupportedServiceManager/OnAccessUnavailable codes, link_name, loopback web.listen, numeric bounds, environment keys), append checkTopology(c) and checkCredentials(c), return &ValidationError{Fields} when any fail, else nil")
}
