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

// Repair finishes pending records and rewrites owned links whose expected
// target changed.
func (e *Engine) Repair(ctx context.Context) (RepairReport, error) {
	panic("SUB-AGENT-TODO: honour ctx; for each registry record: finish pending records per the plan decision (Completed); for owned records whose expected target changed, unlink + symlink only after record+exact-target proof (Repaired); a replaced entry is Preserved with link_conflict and never touched")
}
