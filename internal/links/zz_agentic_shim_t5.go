// agentic:shim
package links

import (
	"context"

	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/rawpath"
)

// RepairEntry is one directory a Repair or RemoveManaged pass acted on.
type RepairEntry struct {
	Key  string
	Dir  rawpath.Path
	Code errcode.Code
}

// RepairReport lists what a Repair or RemoveManaged pass did.
type RepairReport struct {
	Completed, Repaired, Removed, Preserved []RepairEntry
}

// Repair is a compile shim with a deliberately wrong body.
func (e *Engine) Repair(_ context.Context) (RepairReport, error) {
	return RepairReport{}, nil
}

// Remove is a compile shim with a deliberately wrong body.
func (e *Engine) Remove(_ context.Context, _ string) error {
	return nil
}

// RemoveManaged is a compile shim with a deliberately wrong body.
func (e *Engine) RemoveManaged(_ context.Context) (RepairReport, error) {
	return RepairReport{}, nil
}

// List is a compile shim with a deliberately wrong body.
func (e *Engine) List() ([]Record, error) {
	return nil, nil
}
