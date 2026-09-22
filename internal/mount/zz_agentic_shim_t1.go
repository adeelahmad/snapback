// agentic:shim
package mount

// KindFile is a compile shim with a deliberately wrong value.
const KindFile Kind = 0

// OpRead is a compile shim; String does not name it yet.
const OpRead Op = 4

// Gate is a compile shim for the reader-policy seam.
type Gate interface {
	Allow(ev Event) bool
}

// Publisher is a compile shim for the generation-swap seam.
type Publisher interface {
	Publish(cat Catalog)
}
