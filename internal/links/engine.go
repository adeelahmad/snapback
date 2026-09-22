package links

// Result reports the outcome of ensuring a directory's link.
type Result struct {
	Key     string
	Created bool
	Path    string
}

// Engine creates and removes managed links, serialising work per key.
type Engine struct{}

// NewEngine returns an Engine backed by reg and governed by pol.
func NewEngine(reg *Registry, pol Policy) *Engine {
	panic("SUB-AGENT-TODO: T4 — build Engine holding reg, pol and a per-key mutex map (sync.Mutex guarding map[string]*sync.Mutex); no I/O")
}
