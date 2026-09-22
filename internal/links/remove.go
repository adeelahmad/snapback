package links

import "context"

// Remove unlinks the .snapshot entry in dir when the registry proves Snapback
// owns it, then deletes the record.
func (e *Engine) Remove(ctx context.Context, dir string) error {
	panic("SUB-AGENT-TODO: honour ctx; place(dir); openDirChain; unlinkAt only when record + exact target prove ownership, then Registry.Delete; a replaced entry is preserved, record kept, return link_conflict")
}

// RemoveManaged applies the Remove ownership rule to every record.
func (e *Engine) RemoveManaged(ctx context.Context) (RepairReport, error) {
	panic("SUB-AGENT-TODO: honour ctx; apply the Remove ownership rule to every Registry.List record; report Removed and Preserved (link_conflict) entries in a RepairReport")
}

// List returns every registry record, sorted by key.
func (e *Engine) List() ([]Record, error) {
	panic("SUB-AGENT-TODO: return Registry.List")
}
