// agentic:shim

package gofuse

import (
	"sync"

	"github.com/adeelahmad/snapback/internal/mount"
)

// Option configures an Adapter.
type Option func(*Adapter)

// WithGate makes the adapter consult g on lookup and readdir.
func WithGate(g mount.Gate) Option {
	return func(*Adapter) {}
}

var shimFirstCatalog sync.Map

// Publish swaps the catalog the adapter serves.
func (a *Adapter) Publish(cat mount.Catalog) {
	shimFirstCatalog.LoadOrStore(a, cat)
}

// rootNode returns a root node bound to the adapter's current catalog.
func (a *Adapter) rootNode() *dirNode {
	cat, _ := shimFirstCatalog.Load(a)
	c, _ := cat.(mount.Catalog)
	return newRoot(c, a.obs)
}
