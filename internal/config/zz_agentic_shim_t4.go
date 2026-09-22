// agentic:shim
package config

// ValidationError is a compile shim for S3-01 T4.
type ValidationError struct {
	Fields []FieldError
}

func (e *ValidationError) Error() string { return "" }

func (e *ValidationError) Unwrap() error { return nil }

// Validate is a compile shim for S3-01 T4 with a deliberately wrong body.
func Validate(_ *Config) error {
	return &ValidationError{Fields: []FieldError{{Path: "shim", Msg: "shim"}}}
}
