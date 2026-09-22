// agentic:shim
package web

import "context"

// setOpener stands in for the Options.Opener field, which lives in server.go.
// The scaffolder replaces it with the field.
func (o *Options) setOpener(fn func(ctx context.Context, dir string) error) {}
