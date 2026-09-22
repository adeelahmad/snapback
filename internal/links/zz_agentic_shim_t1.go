// agentic:shim

package links

import "errors"

// State is the lifecycle state of a registry record.
type State uint8

// Record states.
const (
	StatePending State = iota + 1
	StateOwned
)

// Record is one registry entry. Rel and Dir become rawpath.Path once S3-03
// merges; the tests assign []byte values so they compile against either.
type Record struct {
	Key    string
	RootID string
	Rel    []byte
	Dir    []byte
	Target string
	State  State
}

// ErrKeyCollision reports a key already mapped to a different root and path.
var ErrKeyCollision = errors.New("links: shim key collision")

// Registry is the bbolt-backed link registry.
type Registry struct{}

// OpenRegistry opens or creates the registry at path.
func OpenRegistry(path string) (*Registry, error) { return &Registry{}, nil }

// Put stores rec under rec.Key.
func (r *Registry) Put(rec Record) error { return nil }

// Get returns the record for key and whether it exists.
func (r *Registry) Get(key string) (Record, bool, error) { return Record{}, false, nil }

// Delete removes the record for key.
func (r *Registry) Delete(key string) error { return nil }

// List returns every record sorted by key.
func (r *Registry) List() ([]Record, error) { return nil, nil }

// Close releases the registry file.
func (r *Registry) Close() error { return nil }
