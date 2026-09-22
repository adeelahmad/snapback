package links

import "sync"

// Result reports the outcome of ensuring a directory's link.
type Result struct {
	Key     string
	Created bool
	Path    string
}

// Engine creates and removes managed links, serialising work per key.
type Engine struct {
	reg   *Registry
	pol   Policy
	mu    sync.Mutex
	locks map[string]*sync.Mutex
}

// NewEngine returns an Engine backed by reg and governed by pol.
func NewEngine(reg *Registry, pol Policy) *Engine {
	return &Engine{reg: reg, pol: pol, locks: map[string]*sync.Mutex{}}
}

// lock acquires the mutex for key and returns its unlock function.
func (e *Engine) lock(key string) func() {
	e.mu.Lock()
	l, ok := e.locks[key]
	if !ok {
		l = &sync.Mutex{}
		e.locks[key] = l
	}
	e.mu.Unlock()
	l.Lock()
	return l.Unlock
}
