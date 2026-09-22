// agentic:shim

package links

import "context"

// Result is a compile shim for S3-07 T4.
type Result struct {
	Key     string
	Created bool
	Path    string
}

// Engine is a compile shim for S3-07 T4.
type Engine struct{}

// NewEngine is a compile shim for S3-07 T4.
func NewEngine(_ *Registry, _ Policy) *Engine { return &Engine{} }

// Ensure is a compile shim for S3-07 T4 with a deliberately wrong body.
func (e *Engine) Ensure(_ context.Context, _ string) (Result, error) {
	return Result{Key: "shim"}, nil
}
